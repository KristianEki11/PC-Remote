package main

import (
	"crypto/tls"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

func main() {
	// 1. Ensure working directory is executable's directory
	if exePath, err := os.Executable(); err == nil {
		_ = os.Chdir(filepath.Dir(exePath))
	}

	url := "https://localhost:8000/internal/qr"

	// 2. Check if server is running
	if !isServerHealthy() {
		// Launch pcremote-server.exe in background as fully detached independent process
		exeName := "pcremote-server.exe"
		if runtime.GOOS != "windows" {
			exeName = "pcremote-server"
		}
		
		cmd := exec.Command(filepath.Join(".", exeName))
		if runtime.GOOS == "windows" {
			// 0x00000008 = DETACHED_PROCESS, 0x00000200 = CREATE_NEW_PROCESS_GROUP
			cmd.SysProcAttr = &syscall.SysProcAttr{
				CreationFlags: 0x00000008 | 0x00000200,
			}
		}
		_ = cmd.Start()

		// Wait up to 3 seconds for server to come online
		for i := 0; i < 30; i++ {
			time.Sleep(100 * time.Millisecond)
			if isServerHealthy() {
				break
			}
		}
	}

	// 3. Open browser to the Steam-style QR pairing dashboard
	openBrowser(url)
}

func isServerHealthy() bool {
	client := &http.Client{
		Timeout: 500 * time.Millisecond,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Get("https://localhost:8000/health")
	if err == nil && resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		return true
	}
	return false
}

func openBrowser(targetURL string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL)
	case "darwin":
		cmd = exec.Command("open", targetURL)
	default:
		cmd = exec.Command("xdg-open", targetURL)
	}
	_ = cmd.Start()
}
