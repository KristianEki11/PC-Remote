package tray

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/getlantern/systray"

	"pcremote-server/auth"
)

// TrayApp manages the system tray icon, menu, and event handling.
type TrayApp struct {
	Sessions  *auth.SessionManager
	Pairing   *auth.PairingManager
	Port      string
	QRPageURL string

	mStatus *systray.MenuItem
	stopCh  chan struct{}
}

// New creates a new TrayApp instance.
func New(sessions *auth.SessionManager, pairing *auth.PairingManager, port, qrPageURL string) *TrayApp {
	return &TrayApp{
		Sessions:  sessions,
		Pairing:   pairing,
		Port:      port,
		QRPageURL: qrPageURL,
		stopCh:    make(chan struct{}),
	}
}

// Run starts the system tray. This MUST be called on the main goroutine.
// onServerReady is called when the tray is initialized (start your HTTP server here).
// onServerExit is called when the user quits from the tray (shutdown your HTTP server here).
func (t *TrayApp) Run(onServerReady func(), onServerExit func()) {
	systray.Run(func() {
		t.onReady()
		if onServerReady != nil {
			onServerReady()
		}
	}, func() {
		close(t.stopCh)
		if onServerExit != nil {
			onServerExit()
		}
	})
}

func (t *TrayApp) onReady() {
	// Load icon from filesystem (favicon.ico next to executable)
	iconData := loadIcon()
	if iconData != nil {
		systray.SetIcon(iconData)
	}
	systray.SetTitle("PC Remote")
	systray.SetTooltip("PC Remote Server")

	// ── Menu Items ───────────────────────────────────────

	mTitle := systray.AddMenuItem(
		fmt.Sprintf("PC Remote Server — Port %s", t.Port),
		"Server information",
	)
	mTitle.Disable()

	systray.AddSeparator()

	t.mStatus = systray.AddMenuItem("🔴 No Device Connected", "Connection status")
	t.mStatus.Disable()

	systray.AddSeparator()

	mQR := systray.AddMenuItem("📱 Show QR Code", "Show QR code for device pairing")

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Stop the server and exit")

	// ── Event Loop ───────────────────────────────────────
	go func() {
		for {
			select {
			case <-mQR.ClickedCh:
				t.openQRPage()
			case <-mQuit.ClickedCh:
				slog.Info("Quit requested from system tray")
				systray.Quit()
			case <-t.stopCh:
				return
			}
		}
	}()

	// ── Status Polling Loop ──────────────────────────────
	go t.statusLoop()
}

// statusLoop periodically checks the session manager and updates the tray menu text.
func (t *TrayApp) statusLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			connected, deviceName := t.Sessions.IsDeviceConnected()
			if connected {
				t.mStatus.SetTitle(fmt.Sprintf("🟢 Connected: %s", deviceName))
				systray.SetTooltip(fmt.Sprintf("PC Remote — %s", deviceName))
			} else {
				t.mStatus.SetTitle("🔴 No Device Connected")
				systray.SetTooltip("PC Remote Server")
			}
		case <-t.stopCh:
			return
		}
	}
}

// openQRPage opens the QR code display page in the default browser.
func (t *TrayApp) openQRPage() {
	slog.Info("Opening QR code page", "url", t.QRPageURL)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", t.QRPageURL)
	case "darwin":
		cmd = exec.Command("open", t.QRPageURL)
	default:
		cmd = exec.Command("xdg-open", t.QRPageURL)
	}
	if err := cmd.Start(); err != nil {
		slog.Error("Failed to open QR page in browser", "error", err)
	}
}

// loadIcon loads the tray icon from the filesystem.
// Tries the executable directory first, then the current working directory.
func loadIcon() []byte {
	paths := []string{}

	// Try executable directory
	if exePath, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exePath), "favicon.ico"))
	}

	// Try current working directory
	paths = append(paths, "favicon.ico")

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			return data
		}
	}

	slog.Warn("System tray icon not found, using default")
	return nil
}
