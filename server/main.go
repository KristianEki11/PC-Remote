package main

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pcremote-server/config"
	"pcremote-server/handlers"
	"pcremote-server/middleware"
)

func main() {
	// 1. Load config
	config.Init()

	// 2. Setup structured logging
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	logFile, err := os.OpenFile("logs/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	multiWriter := io.MultiWriter(os.Stdout, logFile)

	logger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// 3. Setup router (using stdlib ServeMux)
	mux := http.NewServeMux()

	// 3a. Health endpoint (no auth)
	mux.HandleFunc("/health", handlers.HealthHandler)

	// 3b. Protected endpoints sub-router
	protectedMux := http.NewServeMux()

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

	protectedMux.HandleFunc("/system/lock", handlers.SystemLockHandler)
	protectedMux.HandleFunc("/system/shutdown", handlers.SystemShutdownHandler)
	protectedMux.HandleFunc("/system/shutdown/cancel", handlers.SystemShutdownCancelHandler)
	protectedMux.HandleFunc("/system/sleep", handlers.SystemSleepHandler)
	protectedMux.HandleFunc("/system/restart", handlers.SystemRestartHandler)
	protectedMux.HandleFunc("/system/pin", handlers.HandleChangePIN)

	// 3c. Apply auth middleware ONLY to the protected endpoints
	mux.Handle("/", middleware.WithAuth(protectedMux))

	// 4. Create the main server handler with logging middleware wrapping EVERYTHING
	var handler http.Handler = middleware.WithLogging(mux)

	server := &http.Server{
		Addr:    ":" + config.App.Port,
		Handler: handler,
	}

	// 5. Warn if PIN is not configured
	if config.App.PIN == "" {
		slog.Warn("No PIN configured — all authenticated endpoints will return 403. Set PIN or APP_PIN in .env")
	}

	// 6. Setup quit channel for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 7. Start server in a goroutine
	go func() {
		slog.Info("PCRemote Server listening on :" + config.App.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err)
			// Signal quit channel instead of os.Exit to allow graceful cleanup.
			// This prevents orphaned processes when NSSM restarts the service.
			quit <- syscall.SIGTERM
		}
	}()

	// 8. Wait for interrupt signal to gracefully shut down the server
	<-quit

	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exiting")
}
