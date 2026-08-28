package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"pcremote-server/auth"
	"pcremote-server/config"
	"pcremote-server/handlers"
	"pcremote-server/middleware"
	tlsutil "pcremote-server/tls"
	"pcremote-server/tray"
)

func main() {
	// 0. Ensure working directory is the executable's directory.
	if exePath, err := os.Executable(); err == nil {
		_ = os.Chdir(filepath.Dir(exePath))
	}

	// 1. Load config (auto-migrates plaintext PIN to bcrypt)
	config.Init()

	// 2. Setup structured logging
	logPath := filepath.Join("logs", "server.log")
	_ = os.MkdirAll("logs", 0755)

	var logWriter io.Writer = os.Stdout
	if logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		defer logFile.Close()
		logWriter = io.MultiWriter(os.Stdout, logFile)
	} else if fbFile, err := os.OpenFile(getFallbackLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		defer fbFile.Close()
		logWriter = io.MultiWriter(os.Stdout, fbFile)
	}

	logger := slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// 3. Setup TLS certificates (auto-generate self-signed if needed)
	certFile, keyFile, fingerprint, err := tlsutil.EnsureCertificates(config.App.TLSCertDir)
	if err != nil {
		slog.Error("Failed to setup TLS certificates", "error", err)
		log.Fatalf("TLS setup failed: %v", err)
	}

	// 4. Initialize session manager (max 1 device, persistent)
	sessions := auth.NewSessionManager()
	defer sessions.Stop()

	// 5. Initialize pairing manager with server details
	localIP := tlsutil.GetLocalIP()
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "PC Remote"
	}
	allLANs := tlsutil.GetAllLANInterfaces()
	var alternateHosts []string
	for _, iface := range allLANs {
		if iface.IP != localIP {
			alternateHosts = append(alternateHosts, iface.IP)
		}
	}
	pairing := auth.NewPairingManager(localIP, config.App.Port, fingerprint, hostname, alternateHosts)

	// 6. Initialize rate limiters
	authLimiter := middleware.NewAuthRateLimiter() // 5 req/min for auth endpoints
	generalLimiter := middleware.NewRateLimiter()  // 60 req/min for API endpoints

	// 7. Auth handlers
	authH := &auth.AuthHandlers{
		Sessions: sessions,
		Pairing:  pairing,
		Limiter:  authLimiter,
	}

	// 8. QR page handler (internal, localhost-only)
	qrPage := &tray.QRPageHandler{
		Pairing:  pairing,
		Sessions: sessions,
	}

	// 9. Setup router
	mux := http.NewServeMux()

	// ── Public endpoints (no auth) ───────────────────────
	mux.HandleFunc("/health", handlers.HealthHandler)

	// ── Auth endpoints (rate-limited, no Bearer required) ─
	authMux := http.NewServeMux()
	authMux.HandleFunc("/auth/login", authH.LoginHandler)
	authMux.HandleFunc("/auth/pair", authH.PairHandler)
	mux.Handle("/auth/login", authLimiter.Middleware(authMux))
	mux.Handle("/auth/pair", authLimiter.Middleware(authMux))

	// ── Internal endpoints (localhost only) ──────────────
	internalMux := http.NewServeMux()
	internalMux.HandleFunc("/internal/qr", qrPage.ServeQRPage)
	internalMux.HandleFunc("/internal/qr/image", qrPage.ServeQRImage)
	internalMux.HandleFunc("/internal/status", qrPage.ServeStatus)
	internalMux.HandleFunc("/internal/unpair", authH.UnpairHandler)
	internalMux.HandleFunc("/internal/verify-pin", authH.VerifyPinHandler)
	internalMux.HandleFunc("/auth/sessions", authH.SessionsHandler)
	mux.Handle("/internal/", middleware.LocalhostOnly(internalMux))
	mux.Handle("/auth/sessions", middleware.LocalhostOnly(http.HandlerFunc(authH.SessionsHandler)))

	// ── Protected API endpoints ──────────────────────────
	protectedMux := http.NewServeMux()

	protectedMux.HandleFunc("/auth/verify", authH.VerifyHandler)
	protectedMux.HandleFunc("/auth/logout", authH.LogoutHandler)

	protectedMux.HandleFunc("/audio/volume", handlers.AudioVolumeHandler)
	protectedMux.HandleFunc("/audio/mute", handlers.AudioMuteHandler)
	protectedMux.HandleFunc("/audio/status", handlers.AudioStatusHandler)
	protectedMux.HandleFunc("/audio/channels", handlers.AudioChannelsHandler)
	protectedMux.HandleFunc("/audio/channel/volume", handlers.AudioChannelVolumeHandler)
	protectedMux.HandleFunc("/audio/channel/mute", handlers.AudioChannelMuteHandler)
	protectedMux.HandleFunc("/audio/devices", handlers.AudioDevicesHandler)
	protectedMux.HandleFunc("/audio/device/volume", handlers.AudioDeviceVolumeHandler)
	protectedMux.HandleFunc("/audio/device/mute", handlers.AudioDeviceMuteHandler)

	protectedMux.HandleFunc("/browser/open", handlers.BrowserOpenHandler)

	protectedMux.HandleFunc("/media/play", handlers.MediaPlayHandler)
	protectedMux.HandleFunc("/media/next", handlers.MediaNextHandler)
	protectedMux.HandleFunc("/media/prev", handlers.MediaPrevHandler)
	protectedMux.HandleFunc("/media/status", handlers.MediaStatusHandler)

	protectedMux.HandleFunc("/system/lock", handlers.SystemLockHandler)
	protectedMux.HandleFunc("/system/shutdown", handlers.SystemShutdownHandler)
	protectedMux.HandleFunc("/system/shutdown/cancel", handlers.SystemShutdownCancelHandler)
	protectedMux.HandleFunc("/system/sleep", handlers.SystemSleepHandler)
	protectedMux.HandleFunc("/system/restart", handlers.SystemRestartHandler)
	protectedMux.HandleFunc("/system/display/off", handlers.SystemDisplayOffHandler)
	protectedMux.HandleFunc("/system/pin", handlers.HandleChangePIN)

	// Apply auth middleware to protected endpoints
	authMiddleware := middleware.WithAuth(sessions, authLimiter)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/internal/qr", http.StatusFound)
			return
		}
		authMiddleware(protectedMux).ServeHTTP(w, r)
	})

	// 10. Build handler chain: SecurityHeaders → CORS → RateLimit → Logging → Router
	var handler http.Handler = mux
	handler = middleware.WithLogging(handler)
	handler = generalLimiter.Middleware(handler)
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.WithSecurityHeaders(handler)

	server := &http.Server{
		Addr:         ":" + config.App.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 11. Warn if PIN is not configured
	if config.App.PIN == "" {
		slog.Warn("No PIN configured — all authenticated endpoints will return 403. Set PIN or APP_PIN in .env")
	}

	// 12. Configure the QR page URL for the system tray
	qrPageURL := fmt.Sprintf("https://localhost:%s/internal/qr", config.App.Port)

	// 13. Create system tray application
	trayApp := tray.New(sessions, pairing, config.App.Port, qrPageURL)

	// 14. Start HTTPS server immediately in a background goroutine
	go func() {
		slog.Info("PCRemote Server listening",
			"addr", ":"+config.App.Port,
			"protocol", "HTTPS",
			"tls_cert", certFile,
			"local_ip", localIP,
		)
		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			slog.Error("Server listen failed", "error", err)
			os.Exit(1)
		}
	}()

	// 15. Graceful shutdown coordination
	shutdownCh := make(chan struct{})

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		slog.Info("OS signal received, shutting down...")
		select {
		case <-shutdownCh:
		default:
			close(shutdownCh)
		}
	}()

	// 16. Run system tray
	// If tray runs on desktop, user can quit via tray menu.
	// If tray fails to initialize (headless/service), onExit without explicitQuit won't close shutdownCh.
	trayApp.Run(
		nil,
		func() {
			slog.Info("Quit requested from system tray")
			select {
			case <-shutdownCh:
			default:
				close(shutdownCh)
			}
		},
	)

	// 17. Wait indefinitely for explicit shutdown signal (Ctrl+C, taskkill, or Tray Quit)
	<-shutdownCh
	slog.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}
	slog.Info("Server exited cleanly")
}

func getFallbackLogPath() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		dir := filepath.Join(localAppData, "PCRemote", "logs")
		if err := os.MkdirAll(dir, 0755); err == nil {
			return filepath.Join(dir, "server.log")
		}
	}
	return "server.log"
}
