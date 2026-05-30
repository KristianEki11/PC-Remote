package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	PIN  string
}

var App Config

func Init() {
	// Load .env relative to the executable's directory so the service
	// finds it regardless of what working directory NSSM sets.
	exePath, err := os.Executable()
	if err == nil {
		envPath := filepath.Join(filepath.Dir(exePath), ".env")
		_ = godotenv.Load(envPath)
	}
	// Fallback: also try loading from current working directory
	_ = godotenv.Load()

	App.Port = os.Getenv("PORT")
	if App.Port == "" {
		App.Port = os.Getenv("APP_PORT")
	}
	if App.Port == "" {
		App.Port = "8000"
	}

	App.PIN = os.Getenv("PIN")
	if App.PIN == "" {
		// Try to read APP_PIN as fallback based on MIGRATION_NOTES
		App.PIN = os.Getenv("APP_PIN")
	}
}

