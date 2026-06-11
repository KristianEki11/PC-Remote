//go:build windows

package windows

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// ──────────────────────────────────────────────────────────
// RealAPI — System control (lock, sleep, shutdown, restart, display)
// ──────────────────────────────────────────────────────────

func (RealAPI) LockWorkstation() error {
	return runInUserSession("rundll32.exe", "user32.dll,LockWorkStation")
}

func (RealAPI) ScheduleShutdown(delaySeconds int) error {
	delaySec := strconv.Itoa(delaySeconds)
	cmd := exec.Command("shutdown", "/s", "/t", delaySec)
	return cmd.Run()
}

func (RealAPI) CancelShutdown() error {
	cmd := exec.Command("shutdown", "/a")
	return cmd.Run()
}

var (
	powrprof               = syscall.NewLazyDLL("powrprof.dll")
	pSetSuspendState       = powrprof.NewProc("SetSuspendState")
	advapi32               = syscall.NewLazyDLL("advapi32.dll")
	pOpenProcessToken      = advapi32.NewProc("OpenProcessToken")
	pLookupPrivilegeValue  = advapi32.NewProc("LookupPrivilegeValueW")
	pAdjustTokenPrivileges = advapi32.NewProc("AdjustTokenPrivileges")
)

func enableShutdownPrivilege() error {
	const (
		TOKEN_ADJUST_PRIVILEGES = 0x0020
		TOKEN_QUERY             = 0x0008
		SE_PRIVILEGE_ENABLED    = 0x00000002
	)

	type LUID struct {
		LowPart  uint32
		HighPart int32
	}

	type LUID_AND_ATTRIBUTES struct {
		Luid       LUID
		Attributes uint32
	}

	type TOKEN_PRIVILEGES struct {
		PrivilegeCount uint32
		Privileges     [1]LUID_AND_ATTRIBUTES
	}

	currentProcess := syscall.Handle(^uintptr(0))
	var token syscall.Handle
	ret, _, err := pOpenProcessToken.Call(
		uintptr(currentProcess),
		TOKEN_ADJUST_PRIVILEGES|TOKEN_QUERY,
		uintptr(unsafe.Pointer(&token)),
	)
	if ret == 0 {
		return fmt.Errorf("OpenProcessToken failed: %w", err)
	}
	defer syscall.CloseHandle(token)

	var luid LUID
	privilegeName := syscall.StringToUTF16Ptr("SeShutdownPrivilege")
	ret, _, err = pLookupPrivilegeValue.Call(
		0,
		uintptr(unsafe.Pointer(privilegeName)),
		uintptr(unsafe.Pointer(&luid)),
	)
	if ret == 0 {
		return fmt.Errorf("LookupPrivilegeValue failed: %w", err)
	}

	tp := TOKEN_PRIVILEGES{
		PrivilegeCount: 1,
		Privileges: [1]LUID_AND_ATTRIBUTES{
			{
				Luid:       luid,
				Attributes: SE_PRIVILEGE_ENABLED,
			},
		},
	}

	ret, _, err = pAdjustTokenPrivileges.Call(
		uintptr(token),
		0,
		uintptr(unsafe.Pointer(&tp)),
		0,
		0,
		0,
	)
	if ret == 0 {
		return fmt.Errorf("AdjustTokenPrivileges failed: %w", err)
	}

	return nil
}

func (RealAPI) Sleep() error {
	// 1. Try native SetSuspendState (works for traditional S3 sleep systems)
	if err := enableShutdownPrivilege(); err != nil {
		slog.Error("Failed to enable shutdown privilege", "error", err)
	}
	nativeRet, _, nativeErr := pSetSuspendState.Call(0, 0, 0)
	slog.Info("Native SetSuspendState call completed", "ret", nativeRet, "err", nativeErr)

	// 2. Support for Modern Standby (S0) systems by turning off the display.
	// We write a temporary, self-deleting PowerShell script in C:\Users\Public
	// and run it via a temporary scheduled task in the user session.
	psPath := `C:\Users\Public\sleep_temp.ps1`
	psContent := `$code = '[DllImport("user32.dll")] public static extern int SendMessage(int h, int m, int w, int l);'
$type = Add-Type -MemberDefinition $code -Name Win32 -PassThru
$type::SendMessage(-1, 0x0112, 0xF170, 2)
Remove-Item $PSCommandPath -Force
`
	if err := os.WriteFile(psPath, []byte(psContent), 0666); err == nil {
		if runErr := runInUserSession("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", psPath); runErr != nil {
			slog.Error("Failed to run Modern Standby sleep helper in user session", "error", runErr)
			// Clean up file if task scheduling failed
			if _, statErr := os.Stat(psPath); statErr == nil {
				os.Remove(psPath)
			}
		} else {
			slog.Info("Modern Standby sleep helper executed successfully in user session")
		}
	} else {
		slog.Error("Failed to write temporary sleep helper script to C:\\Users\\Public", "error", err)
	}

	return nil
}

func (RealAPI) Restart() error {
	cmd := exec.Command("shutdown", "/r", "/t", "5")
	return cmd.Run()
}

var (
	pSetThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")
	displayMutex             sync.Mutex
	isMonitoring             bool
)

const (
	ES_SYSTEM_REQUIRED = 0x00000001
	ES_CONTINUOUS      = 0x80000000
)

func setKeepAwake(awake bool) {
	if awake {
		slog.Info("Setting keep awake execution state (ES_SYSTEM_REQUIRED)")
		pSetThreadExecutionState.Call(uintptr(ES_CONTINUOUS | ES_SYSTEM_REQUIRED))
	} else {
		slog.Info("Clearing keep awake execution state")
		pSetThreadExecutionState.Call(uintptr(ES_CONTINUOUS))
	}
}

func monitorDisplayState() {
	defer func() {
		displayMutex.Lock()
		isMonitoring = false
		displayMutex.Unlock()
	}()

	// Wait 5 seconds to let the screen actually turn off and WMI update its state
	time.Sleep(5 * time.Second)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Check if screen has been turned back on by user
		cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command",
			"Get-CimInstance -Namespace root\\wmi -ClassName WmiMonitorBasicDisplayParams | Select-Object -ExpandProperty Active")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		out, err := cmd.Output()
		if err != nil {
			// If WMI query fails (e.g. system is temporarily unresponsive), keep polling
			continue
		}

		status := strings.ToLower(string(out))
		if strings.Contains(status, "true") {
			slog.Info("Monitor has been turned back on by user. Releasing keep-awake state.")
			setKeepAwake(false)
			return
		}
	}
}

// TurnOffDisplay turns off the monitor without locking or sleeping the PC.
// It uses the Win32 SendMessage(HWND_BROADCAST, WM_SYSCOMMAND, SC_MONITORPOWER, 2) API,
// executed inside the active user session (via WTS) so the display actually turns off
// even when the server runs as a Windows Service / SYSTEM account.
// The PC remains fully awake — only the backlight is cut.
func (RealAPI) TurnOffDisplay() error {
	displayMutex.Lock()
	defer displayMutex.Unlock()

	psPath := `C:\Users\Public\display_off_temp.ps1`
	// SC_MONITORPOWER = 0xF170, value 2 = power off
	// We broadcast to HWND_BROADCAST (-1) which reaches the desktop window manager.
	psContent := `$code = '[DllImport("user32.dll")] public static extern int SendMessage(int h, int m, int w, int l);'
$type = Add-Type -MemberDefinition $code -Name WinUser -PassThru
$type::SendMessage(-1, 0x0112, 0xF170, 2)
Remove-Item $PSCommandPath -Force
`
	if err := os.WriteFile(psPath, []byte(psContent), 0666); err != nil {
		return fmt.Errorf("failed to write display-off helper script: %w", err)
	}

	// Set system keep-awake state before turning off display
	setKeepAwake(true)

	// Use runInUserSessionStart to execute the script asynchronously so the API returns immediately
	if runErr := runInUserSessionStart("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", psPath); runErr != nil {
		slog.Error("Failed to run display-off helper in user session", "error", runErr)
		setKeepAwake(false) // Clean up keep-awake on failure
		if _, statErr := os.Stat(psPath); statErr == nil {
			os.Remove(psPath)
		}
		return runErr
	}

	slog.Info("Display off triggered asynchronously")

	// Start background monitoring if not already monitoring
	if !isMonitoring {
		isMonitoring = true
		go monitorDisplayState()
	}

	return nil
}

// ──────────────────────────────────────────────────────────
// RealAPI — Browser
// ──────────────────────────────────────────────────────────

func (RealAPI) OpenBrowser(url string) error {
	if !hasHTTPPrefix(url) {
		return errors.New("url must start with http:// or https://")
	}
	// Use Start() instead of CombinedOutput() — we only care that the process
	// was launched successfully, not its exit code. explorer.exe may exit with
	// non-zero even when it successfully opens the URL.
	err := runInUserSessionStart("explorer.exe", url)
	if err == nil {
		return nil
	}
	slog.Warn("Failed to open browser via explorer.exe, falling back to rundll32", "error", err)
	// Fallback to standard FileProtocolHandler
	return runInUserSessionStart("rundll32", "url.dll,FileProtocolHandler", url)
}

func hasHTTPPrefix(url string) bool {
	return len(url) > 7 && (url[:7] == "http://" || (len(url) > 8 && url[:8] == "https://"))
}
