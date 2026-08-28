package tunnel

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"syscall"
	"time"
)

var (
	urlRegex = regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)
)

type TunnelStatus string

const (
	StatusDisabled TunnelStatus = "disabled"
	StatusStarting TunnelStatus = "starting"
	StatusActive   TunnelStatus = "active"
	StatusError    TunnelStatus = "error"
)

// TunnelManager manages a Cloudflare Quick Tunnel for zero-config public internet access.
type TunnelManager struct {
	mu           sync.RWMutex
	port         string
	publicURL    string
	status       TunnelStatus
	lastError    string
	cmd          *exec.Cmd
	cancel       context.CancelFunc
	onURLChanged func(publicURL string)
	enabled      bool
}

// New creates a new TunnelManager.
func New(port string, onURLChanged func(publicURL string)) *TunnelManager {
	return &TunnelManager{
		port:         port,
		status:       StatusDisabled,
		onURLChanged: onURLChanged,
		enabled:      true, // Auto-enable by default for anywhere access
	}
}

// Start launches the Cloudflare tunnel in the background.
func (tm *TunnelManager) Start() error {
	tm.mu.Lock()
	if tm.status == StatusActive || tm.status == StatusStarting {
		tm.mu.Unlock()
		return nil
	}
	tm.status = StatusStarting
	tm.mu.Unlock()

	go tm.runLoop()
	return nil
}

// Stop cleanly terminates the active tunnel.
func (tm *TunnelManager) Stop() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.enabled = false
	if tm.cancel != nil {
		tm.cancel()
	}
	if tm.cmd != nil && tm.cmd.Process != nil {
		_ = tm.cmd.Process.Kill()
	}
	tm.status = StatusDisabled
	tm.publicURL = ""
	if tm.onURLChanged != nil {
		tm.onURLChanged("")
	}
}

// GetPublicURL returns the current public HTTPS tunnel URL (if active).
func (tm *TunnelManager) GetPublicURL() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.publicURL
}

// GetStatus returns the current tunnel status.
func (tm *TunnelManager) GetStatus() (TunnelStatus, string, string) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.status, tm.publicURL, tm.lastError
}

func (tm *TunnelManager) runLoop() {
	for {
		tm.mu.RLock()
		enabled := tm.enabled
		tm.mu.RUnlock()
		if !enabled {
			return
		}

		binPath, err := tm.findOrDownloadBinary()
		if err != nil {
			slog.Warn("Cloudflare tunnel binary unavailable", "error", err)
			tm.mu.Lock()
			tm.status = StatusError
			tm.lastError = err.Error()
			tm.mu.Unlock()
			time.Sleep(30 * time.Second)
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		tm.mu.Lock()
		tm.cancel = cancel
		tm.mu.Unlock()

		targetLocalURL := fmt.Sprintf("https://localhost:%s", tm.port)
		cmd := exec.CommandContext(ctx, binPath, "tunnel", "--url", targetLocalURL, "--no-tls-verify")
		if runtime.GOOS == "windows" {
			cmd.SysProcAttr = &syscall.SysProcAttr{
				HideWindow: true,
			}
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			slog.Error("Failed to get tunnel stderr pipe", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			slog.Error("Failed to get tunnel stdout pipe", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if err := cmd.Start(); err != nil {
			slog.Error("Failed to start cloudflared process", "error", err)
			tm.mu.Lock()
			tm.status = StatusError
			tm.lastError = err.Error()
			tm.mu.Unlock()
			time.Sleep(10 * time.Second)
			continue
		}

		tm.mu.Lock()
		tm.cmd = cmd
		tm.mu.Unlock()

		// Cloudflare tunnel writes connection logs and assigned URLs to stderr
		go io.Copy(io.Discard, stdout)
		scanner := bufio.NewScanner(stderr)

		for scanner.Scan() {
			line := scanner.Text()
			if match := urlRegex.FindString(line); match != "" {
				slog.Info("Cloudflare Public Tunnel active", "public_url", match)
				tm.mu.Lock()
				tm.publicURL = match
				tm.status = StatusActive
				tm.lastError = ""
				tm.mu.Unlock()

				if tm.onURLChanged != nil {
					tm.onURLChanged(match)
				}
			}
		}

		_ = cmd.Wait()
		slog.Warn("Cloudflare tunnel disconnected, reconnecting in 5s...")

		tm.mu.Lock()
		if tm.status == StatusActive {
			tm.status = StatusStarting
		}
		tm.publicURL = ""
		tm.mu.Unlock()
		if tm.onURLChanged != nil {
			tm.onURLChanged("")
		}

		time.Sleep(5 * time.Second)
	}
}

func (tm *TunnelManager) findOrDownloadBinary() (string, error) {
	// 1. Check local program directory & PATH
	candidates := []string{
		"cloudflared.exe",
		"cloudflared",
		filepath.Join("dist", "cloudflared.exe"),
		`C:\Program Files\PCRemote\cloudflared.exe`,
	}

	if exePath, err := os.Executable(); err == nil {
		candidates = append(candidates,
			filepath.Join(filepath.Dir(exePath), "cloudflared.exe"),
			filepath.Join(filepath.Dir(exePath), "cloudflared"),
			filepath.Join(filepath.Dir(exePath), "dist", "cloudflared.exe"),
		)
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		candidates = append(candidates,
			filepath.Join(localAppData, "PCRemote", "bin", "cloudflared.exe"),
			filepath.Join(localAppData, "PCRemote", "cloudflared.exe"),
		)
	}

	for _, path := range candidates {
		if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
			if abs, err := filepath.Abs(path); err == nil {
				return abs, nil
			}
			return path, nil
		}
	}

	if path, err := exec.LookPath("cloudflared"); err == nil {
		return path, nil
	}

	// 2. If not found, download official binary on demand
	targetDir := filepath.Join(localAppData, "PCRemote", "bin")
	_ = os.MkdirAll(targetDir, 0755)
	targetFile := filepath.Join(targetDir, "cloudflared.exe")

	slog.Info("Downloading official Cloudflare tunnel binary...", "destination", targetFile)

	downloadURL := "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-windows-amd64.exe"
	if runtime.GOOS == "darwin" {
		downloadURL = "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-darwin-amd64"
	} else if runtime.GOOS == "linux" {
		downloadURL = "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64"
	}

	client := &http.Client{
		Timeout: 3 * time.Minute,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	resp, err := client.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download cloudflared: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad HTTP status downloading cloudflared: %s", resp.Status)
	}

	out, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create binary file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", fmt.Errorf("failed to write binary: %w", err)
	}

	return targetFile, nil
}
