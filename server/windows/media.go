//go:build windows

package windows

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// ──────────────────────────────────────────────────────────
// Session isolation & media key injection
// ──────────────────────────────────────────────────────────

var (
	kernel32              = syscall.NewLazyDLL("kernel32.dll")
	pProcessIdToSessionId = kernel32.NewProc("ProcessIdToSessionId")
)

// runInUserSession executes an executable in the active interactive user session.
// If running in Session 0 (as a Windows service), it registers a temporary
// scheduled task to bypass Session 0 isolation. Otherwise, it runs the command directly.
func runInUserSession(exePath string, args ...string) error {
	var sessionID uint32
	pid := uint32(os.Getpid())
	ret, _, _ := pProcessIdToSessionId.Call(uintptr(pid), uintptr(unsafe.Pointer(&sessionID)))

	// If we are already running in an interactive user session (Session ID > 0),
	// we do not need to use schtasks to bypass Session 0 isolation.
	if ret != 0 && sessionID > 0 {
		slog.Info("Running command directly (interactive session)", "session_id", sessionID, "exePath", exePath, "args", args)
		cmd := exec.Command(exePath, args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("direct command execution failed: %w — %s", err, string(out))
		}
		return nil
	}

	slog.Info("Running command via schtasks (Session 0 service)", "exePath", exePath, "args", args)
	// Use a unique task name with timestamp to avoid conflicts
	taskName := fmt.Sprintf("PCRemoteTask_%d_%d", syscall.Getpid(), time.Now().UnixMilli())

	// Format /tr command line
	// Since we use short 8.3 path (no spaces), we don't need nested quoting!
	taskCmdLine := exePath
	for _, arg := range args {
		taskCmdLine += " " + arg
	}

	register := exec.Command("schtasks", "/create", "/tn", taskName,
		"/tr", taskCmdLine,
		"/sc", "ONCE",
		"/st", "00:00",
		"/f",
		"/ru", "INTERACTIVE",
	)
	register.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if out, err := register.CombinedOutput(); err != nil {
		return fmt.Errorf("schtasks /create failed: %w — %s", err, string(out))
	}

	// Run the task immediately
	run := exec.Command("schtasks", "/run", "/tn", taskName)
	run.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := run.CombinedOutput(); err != nil {
		// Try to clean up even if run failed
		delCmd := exec.Command("schtasks", "/delete", "/tn", taskName, "/f")
		delCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		delCmd.Run()
		return fmt.Errorf("schtasks /run failed: %w — %s", err, string(out))
	}

	// Delete the task in a background goroutine after a short sleep to allow the process to finish
	go func(tn string) {
		time.Sleep(1 * time.Second)
		delCmd := exec.Command("schtasks", "/delete", "/tn", tn, "/f")
		delCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		delCmd.Run()
	}(taskName)
	return nil
}

// runInUserSessionStart is like runInUserSession but uses Start() instead of
// CombinedOutput(). It only cares that the process was launched, not its exit code.
// This is essential for commands like explorer.exe <url> which may exit non-zero
// even when the URL is successfully opened.
func runInUserSessionStart(exePath string, args ...string) error {
	var sessionID uint32
	pid := uint32(os.Getpid())
	ret, _, _ := pProcessIdToSessionId.Call(uintptr(pid), uintptr(unsafe.Pointer(&sessionID)))

	if ret != 0 && sessionID > 0 {
		slog.Info("Running command directly (interactive session)", "session_id", sessionID, "exePath", exePath, "args", args)
		cmd := exec.Command(exePath, args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("cmd.Start failed: %w", err)
		}
		cmd.Process.Release()
		return nil
	}

	slog.Info("Running command via schtasks (Session 0 service)", "exePath", exePath, "args", args)
	taskName := fmt.Sprintf("PCRemoteTask_%d_%d", syscall.Getpid(), time.Now().UnixMilli())

	taskCmdLine := exePath
	for _, arg := range args {
		taskCmdLine += " " + arg
	}

	register := exec.Command("schtasks", "/create", "/tn", taskName,
		"/tr", taskCmdLine,
		"/sc", "ONCE",
		"/st", "00:00",
		"/f",
		"/ru", "INTERACTIVE",
	)
	register.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := register.CombinedOutput(); err != nil {
		return fmt.Errorf("schtasks /create failed: %w — %s", err, string(out))
	}

	run := exec.Command("schtasks", "/run", "/tn", taskName)
	run.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := run.CombinedOutput(); err != nil {
		delCmd := exec.Command("schtasks", "/delete", "/tn", taskName, "/f")
		delCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		delCmd.Run()
		return fmt.Errorf("schtasks /run failed: %w — %s", err, string(out))
	}

	go func(tn string) {
		time.Sleep(1 * time.Second)
		delCmd := exec.Command("schtasks", "/delete", "/tn", tn, "/f")
		delCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		delCmd.Run()
	}(taskName)
	return nil
}

// ──────────────────────────────────────────────────────────
// RealAPI — Media methods
// ──────────────────────────────────────────────────────────

func (RealAPI) SendMediaKey(action string) error {
	action = strings.ToLower(action)
	var method string
	var vkCode int
	switch action {
	case "play_pause":
		method = "TryTogglePlayPauseAsync"
		vkCode = 0xB3 // VK_MEDIA_PLAY_PAUSE
	case "next":
		method = "TrySkipNextAsync"
		vkCode = 0xB0 // VK_MEDIA_NEXT_TRACK
	case "prev":
		method = "TrySkipPreviousAsync"
		vkCode = 0xB1 // VK_MEDIA_PREV_TRACK
	default:
		return fmt.Errorf("unknown media action: %q", action)
	}

	script := fmt.Sprintf(`
[void][System.Reflection.Assembly]::LoadWithPartialName("System.Runtime.WindowsRuntime")
[void][Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager, Windows.Media.Control, ContentType=WindowsRuntime]

$asTaskMethod = [System.WindowsRuntimeSystemExtensions].GetMethods() | Where-Object { $_.Name -eq 'AsTask' -and $_.IsGenericMethod -and $_.GetParameters().Length -eq 1 } | Select-Object -First 1
$genericMethod = $asTaskMethod.MakeGenericMethod([Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager])
$task = $genericMethod.Invoke($null, @([Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager]::RequestAsync()))
$manager = $task.Result

$session = $manager.GetCurrentSession()
$success = $false
if ($session) {
    # Try calling the SMTC method (async)
    $asyncOp = $session.%s()
    $success = $true
}

if (-not $success) {
    # Fallback: Simulate a global media key press if no SMTC session is active.
    # VK_MEDIA_NEXT_TRACK = 0xB0, VK_MEDIA_PREV_TRACK = 0xB1, VK_MEDIA_PLAY_PAUSE = 0xB3
    $signature = @'
    [DllImport("user32.dll")]
    public static extern void keybd_event(byte bVk, byte bScan, uint dwFlags, uint dwExtraInfo);
'@
    $type = Add-Type -MemberDefinition $signature -Name "Keyboard" -Namespace "Win32" -PassThru
    $type::keybd_event(%d, 0, 0, 0) # Key Down
    $type::keybd_event(%d, 0, 2, 0) # Key Up
}
`, method, vkCode, vkCode)

	appDir := getAppDir()
	scriptFile := filepath.Join(appDir, "send_media_temp.ps1")
	if err := os.WriteFile(scriptFile, []byte(script), 0644); err != nil {
		return fmt.Errorf("failed to write media script: %v", err)
	}
	defer os.Remove(scriptFile)

	_, err := runInUserSessionWithOutput(scriptFile, filepath.Join(appDir, "send_media_out.txt"))
	if err != nil {
		slog.Warn("SMTC script execution", "error", err)
	}
	return nil
}

const mediaStatusScript = `
$OutputFile = "{{OUTPUT_FILE}}"

[void][System.Reflection.Assembly]::LoadWithPartialName("System.Runtime.WindowsRuntime")
[void][Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager, Windows.Media.Control, ContentType=WindowsRuntime]
[void][Windows.Media.Control.GlobalSystemMediaTransportControlsSession, Windows.Media.Control, ContentType=WindowsRuntime]
[void][Windows.Media.Control.GlobalSystemMediaTransportControlsSessionPlaybackInfo, Windows.Media.Control, ContentType=WindowsRuntime]
[void][Windows.Media.Control.GlobalSystemMediaTransportControlsSessionMediaProperties, Windows.Media.Control, ContentType=WindowsRuntime]

function Get-WinRTResult($asyncOp, [Type]$type) {
    $asTaskMethod = [System.WindowsRuntimeSystemExtensions].GetMethods() | 
        Where-Object { $_.Name -eq 'AsTask' -and $_.IsGenericMethod -and $_.GetParameters().Length -eq 1 } | 
        Select-Object -First 1
    $genericMethod = $asTaskMethod.MakeGenericMethod($type)
    $task = $genericMethod.Invoke($null, @($asyncOp))
    return $task.Result
}

$result = try {
    $asyncOp = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager]::RequestAsync()
    $manager = Get-WinRTResult $asyncOp ([Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager])
    $session = $manager.GetCurrentSession()
    if ($session) {
        $playbackInfo = $session.GetPlaybackInfo()
        $propAsync = $session.TryGetMediaPropertiesAsync()
        $mediaProperties = Get-WinRTResult $propAsync ([Windows.Media.Control.GlobalSystemMediaTransportControlsSessionMediaProperties])
        
        $status = $playbackInfo.PlaybackStatus.ToString()
        $title = $mediaProperties.Title
        $artist = $mediaProperties.Artist
        $album = $mediaProperties.AlbumTitle
        $appId = $session.SourceAppId
        
        @{
            success = $true
            status = $status
            title = $title
            artist = $artist
            album = $album
            app_id = $appId
        } | ConvertTo-Json
    } else {
        @{
            success = $true
            status = "Closed"
            title = ""
            artist = ""
            album = ""
            app_id = ""
        } | ConvertTo-Json
    }
} catch {
    @{
        success = $false
        error = $_.Exception.Message
    } | ConvertTo-Json
}

if ($OutputFile) {
    $result | Out-File -FilePath $OutputFile -Encoding utf8 -Force
} else {
    $result
}
`

var (
	mediaCacheVal  map[string]any
	mediaCacheTime time.Time
)

// getAppDir returns the directory of the running server executable, converting "C:\Program Files" to "C:\PROGRA~1" to avoid spaces.
func getAppDir() string {
	if serverExe, err := os.Executable(); err == nil {
		dir := filepath.Dir(serverExe)
		return strings.ReplaceAll(dir, "C:\\Program Files", "C:\\PROGRA~1")
	}
	return "."
}

// runInUserSessionWithOutput runs a powershell script in the interactive session, redirecting its output to a file and reading it.
func runInUserSessionWithOutput(scriptFile string, outputFile string) (string, error) {
	_ = os.Remove(outputFile)

	var sessionID uint32
	pid := uint32(os.Getpid())
	ret, _, _ := pProcessIdToSessionId.Call(uintptr(pid), uintptr(unsafe.Pointer(&sessionID)))

	// If interactive session, run directly
	if ret != 0 && sessionID > 0 {
		slog.Info("Running media status directly (interactive session)", "scriptFile", scriptFile)
		cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", scriptFile)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("direct run failed: %w - %s", err, string(out))
		}
		return string(out), nil
	}

	// Run via scheduled task bypass (Session 0)
	slog.Info("Running media status via schtasks (Session 0 service)", "scriptFile", scriptFile)
	taskName := fmt.Sprintf("PCRemoteTask_Media_%d_%d", syscall.Getpid(), time.Now().UnixMilli())

	// Format /tr command line
	// Since we use 8.3 short paths (no spaces), we don't need nested quoting!
	taskCmdLine := fmt.Sprintf("powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -File %s", scriptFile)

	register := exec.Command("schtasks", "/create", "/tn", taskName,
		"/tr", taskCmdLine,
		"/sc", "ONCE",
		"/st", "00:00",
		"/f",
		"/ru", "INTERACTIVE",
	)
	register.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if out, err := register.CombinedOutput(); err != nil {
		return "", fmt.Errorf("schtasks create failed: %w - %s", err, string(out))
	}

	run := exec.Command("schtasks", "/run", "/tn", taskName)
	run.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := run.CombinedOutput(); err != nil {
		delCmd := exec.Command("schtasks", "/delete", "/tn", taskName, "/f")
		delCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		delCmd.Run()
		return "", fmt.Errorf("schtasks run failed: %w - %s", err, string(out))
	}

	go func(tn string) {
		time.Sleep(1500 * time.Millisecond)
		delCmd := exec.Command("schtasks", "/delete", "/tn", tn, "/f")
		delCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		delCmd.Run()
	}(taskName)

	// Poll file read with 1.5s timeout
	for i := 0; i < 15; i++ {
		time.Sleep(100 * time.Millisecond)
		if info, err := os.Stat(outputFile); err == nil && info.Size() > 0 {
			data, readErr := os.ReadFile(outputFile)
			if readErr == nil {
				_ = os.Remove(outputFile)
				return string(data), nil
			}
		}
	}

	return "", fmt.Errorf("timeout waiting for output file: %s", outputFile)
}

func (RealAPI) GetMediaStatus() (map[string]any, error) {
	if time.Since(mediaCacheTime) < 500*time.Millisecond && mediaCacheVal != nil {
		return mediaCacheVal, nil
	}

	appDir := getAppDir()
	scriptFile := filepath.Join(appDir, "get_media_status_temp.ps1")
	outputFile := filepath.Join(appDir, "media_status_temp.json")

	// Determine if running in interactive session
	var sessionID uint32
	pid := uint32(os.Getpid())
	ret, _, _ := pProcessIdToSessionId.Call(uintptr(pid), uintptr(unsafe.Pointer(&sessionID)))
	isInteractive := ret != 0 && sessionID > 0

	var scriptContent string
	if isInteractive {
		scriptContent = strings.ReplaceAll(mediaStatusScript, "{{OUTPUT_FILE}}", "")
	} else {
		scriptContent = strings.ReplaceAll(mediaStatusScript, "{{OUTPUT_FILE}}", outputFile)
	}

	// Write powershell script to file to avoid cmd escaping bugs
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp script file: %w", err)
	}
	defer os.Remove(scriptFile)

	outStr, err := runInUserSessionWithOutput(scriptFile, outputFile)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(outStr), &result); err != nil {
		return nil, fmt.Errorf("failed to decode JSON from powershell media status: %w - raw: %s", err, outStr)
	}

	mediaCacheVal = result
	mediaCacheTime = time.Now()
	return result, nil
}

