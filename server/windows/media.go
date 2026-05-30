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

// runInUserSession executes a command in the active interactive user session.
// If running in Session 0 (as a Windows service), it registers a temporary
// scheduled task to bypass Session 0 isolation. Otherwise, it runs the command directly.
func runInUserSession(command string) error {
	var sessionID uint32
	pid := uint32(os.Getpid())
	ret, _, _ := pProcessIdToSessionId.Call(uintptr(pid), uintptr(unsafe.Pointer(&sessionID)))

	// If we are already running in an interactive user session (Session ID > 0),
	// we do not need to use schtasks to bypass Session 0 isolation.
	if ret != 0 && sessionID > 0 {
		slog.Info("Running command directly (interactive session)", "session_id", sessionID, "command", command)
		cmd := exec.Command("cmd.exe", "/c", command)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("direct command execution failed: %w — %s", err, string(out))
		}
		return nil
	}

	slog.Info("Running command via schtasks (Session 0 service)", "command", command)
	// Use a unique task name with timestamp to avoid conflicts
	taskName := fmt.Sprintf("PCRemoteTask_%d_%d", syscall.Getpid(), time.Now().UnixMilli())

	// Register the task to run once immediately (interactive = all logged on users)
	register := exec.Command("schtasks", "/create", "/tn", taskName,
		"/tr", command,
		"/sc", "ONCE",
		"/st", "00:00",
		"/f",
		"/ru", "INTERACTIVE",
	)
	if out, err := register.CombinedOutput(); err != nil {
		return fmt.Errorf("schtasks /create failed: %w — %s", err, string(out))
	}

	// Run the task immediately
	run := exec.Command("schtasks", "/run", "/tn", taskName)
	if out, err := run.CombinedOutput(); err != nil {
		// Try to clean up even if run failed
		exec.Command("schtasks", "/delete", "/tn", taskName, "/f").Run() //nolint
		return fmt.Errorf("schtasks /run failed: %w — %s", err, string(out))
	}

	// Delete the task in a background goroutine after a short sleep to allow the process to finish
	go func(tn string) {
		time.Sleep(1 * time.Second)
		exec.Command("schtasks", "/delete", "/tn", tn, "/f").Run()
	}(taskName)
	return nil
}

// ──────────────────────────────────────────────────────────
// RealAPI — Media methods
// ──────────────────────────────────────────────────────────

func (RealAPI) SendMediaKey(action string) error {
	action = strings.ToLower(action)
	if action != "play_pause" && action != "next" && action != "prev" {
		return fmt.Errorf("unknown media action: %q", action)
	}

	// Resolve sendkey.exe path: check install dir first, then exe-relative
	exePath := filepath.Join(os.Getenv("PROGRAMFILES"), "PCRemote", "sendkey.exe")
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		// Fallback: same directory as the server executable
		if serverExe, exeErr := os.Executable(); exeErr == nil {
			exePath = filepath.Join(filepath.Dir(serverExe), "sendkey.exe")
		}
	}

	cmd := fmt.Sprintf(`"%s" %s`, exePath, action)
	return runInUserSession(cmd)
}

const mediaStatusScript = `
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

try {
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
`

var (
	mediaCacheVal  map[string]any
	mediaCacheTime time.Time
)

func (RealAPI) GetMediaStatus() (map[string]any, error) {
	if time.Since(mediaCacheTime) < 500*time.Millisecond && mediaCacheVal != nil {
		return mediaCacheVal, nil
	}

	// Run PowerShell script dynamically using stdin to avoid disk writes
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", "-")
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// Write script to stdin asynchronously
	go func() {
		defer stdin.Close()
		_, _ = stdin.Write([]byte(mediaStatusScript))
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run powershell media status: %w - %s", err, string(out))
	}

	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to decode JSON from powershell media status: %w - raw: %s", err, string(out))
	}

	mediaCacheVal = result
	mediaCacheTime = time.Now()
	return result, nil
}

