package config

import (
	"bufio"
	"bytes"
	"crypto/subtle"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// Config holds all server configuration loaded from .env and environment variables.
type Config struct {
	Port       string
	PIN        string // bcrypt hash (or plaintext for legacy, auto-migrated on startup)
	TLSCertDir string
}

// App is the global configuration instance.
var App Config

// Init loads configuration from .env files and environment variables.
// If a plaintext PIN is detected, it is automatically hashed with bcrypt
// and the .env file is updated in-place.
// Init loads configuration from .env files and environment variables.
// If a plaintext PIN is detected, it is automatically hashed with bcrypt
// and the .env file is updated in-place.
func Init() {
	// 1. First load via godotenv for standard env vars
	exePath, err := os.Executable()
	var envPath string
	if err == nil {
		envPath = filepath.Join(filepath.Dir(exePath), ".env")
		_ = godotenv.Load(envPath)
	}
	_ = godotenv.Load()

	// 2. Direct literal parser to prevent godotenv $ variable expansion from corrupting $2a$ bcrypt hashes
	if envPath != "" {
		loadLiteralEnv(envPath)
	}
	loadLiteralEnv(".env")

	// ── Port ─────────────────────────────────────────────
	if App.Port == "" {
		App.Port = os.Getenv("PORT")
	}
	if App.Port == "" {
		App.Port = os.Getenv("APP_PORT")
	}
	if App.Port == "" {
		App.Port = "8000"
	}

	// ── PIN ──────────────────────────────────────────────
	if App.PIN == "" {
		rawPIN := os.Getenv("PIN")
		if rawPIN == "" {
			rawPIN = os.Getenv("APP_PIN")
		}
		App.PIN = strings.Trim(rawPIN, "'\"")
	}

	// Auto-migrate: if PIN is plaintext (not a bcrypt hash), hash it now
	// and rewrite the .env file so the plaintext is never stored at rest.
	if App.PIN != "" && !isBcryptHash(App.PIN) {
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(App.PIN), bcrypt.DefaultCost)
		if hashErr == nil {
			App.PIN = string(hash)
			slog.Info("PIN auto-migrated to bcrypt hash")
			autoHashPINInEnv(string(hash))
		} else {
			slog.Error("Failed to hash PIN", "error", hashErr)
		}
	}

	// ── TLS Certificate Directory ────────────────────────
	if App.TLSCertDir == "" {
		App.TLSCertDir = os.Getenv("TLS_CERT_DIR")
	}
	if App.TLSCertDir == "" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			App.TLSCertDir = filepath.Join(localAppData, "PCRemote", "tls")
		} else {
			App.TLSCertDir = "tls"
		}
	}
}

func loadLiteralEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "'\"")

		if key == "PIN" || key == "APP_PIN" {
			if val != "" {
				App.PIN = val
			}
		} else if key == "PORT" || key == "APP_PORT" {
			if val != "" {
				App.Port = val
			}
		} else if key == "TLS_CERT_DIR" {
			if val != "" {
				App.TLSCertDir = val
			}
		}
	}
}

// ValidatePIN checks the provided PIN against the stored hash.
// Supports both bcrypt hashes (preferred) and plaintext (legacy fallback).
func ValidatePIN(pin string) bool {
	if App.PIN == "" {
		return false
	}
	if isBcryptHash(App.PIN) {
		return bcrypt.CompareHashAndPassword([]byte(App.PIN), []byte(pin)) == nil
	}
	// Legacy fallback: constant-time compare for plaintext
	return subtle.ConstantTimeCompare([]byte(pin), []byte(App.PIN)) == 1
}

// HashPIN generates a bcrypt hash of the given PIN string.
func HashPIN(pin string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	return string(hash), err
}

// SetPIN updates the in-memory PIN hash.
func SetPIN(hash string) {
	App.PIN = hash
}

// isBcryptHash returns true if the string looks like a bcrypt hash.
func isBcryptHash(s string) bool {
	return strings.HasPrefix(s, "$2a$") ||
		strings.HasPrefix(s, "$2b$") ||
		strings.HasPrefix(s, "$2y$")
}

// autoHashPINInEnv rewrites the .env file to replace the plaintext PIN with the bcrypt hash.
// This is a one-time migration that runs automatically on first startup after upgrade.
func autoHashPINInEnv(hash string) {
	envPath := ".env"
	exePath, err := os.Executable()
	if err == nil {
		targetPath := filepath.Join(filepath.Dir(exePath), ".env")
		if _, statErr := os.Stat(targetPath); statErr == nil {
			envPath = targetPath
		}
	}

	file, err := os.Open(envPath)
	if err != nil {
		slog.Warn("Could not open .env for PIN auto-migration", "error", err)
		return
	}

	var newContent bytes.Buffer
	scanner := bufio.NewScanner(file)
	keyFound := false
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		var pinKeyName string
		if strings.HasPrefix(trimmed, "PIN=") {
			pinKeyName = "PIN"
		} else if strings.HasPrefix(trimmed, "APP_PIN=") {
			pinKeyName = "APP_PIN"
		}

		if pinKeyName != "" {
			newContent.WriteString(pinKeyName + "='" + hash + "'\n")
			keyFound = true
		} else {
			newContent.WriteString(line + "\n")
		}
	}
	file.Close()

	if scanner.Err() != nil {
		slog.Warn("Error reading .env during PIN auto-migration", "error", scanner.Err())
		return
	}

	if !keyFound {
		newContent.WriteString("APP_PIN='" + hash + "'\n")
	}

	// Atomic write: temp file → rename
	tmpPath := envPath + ".tmp"
	if err := os.WriteFile(tmpPath, newContent.Bytes(), 0600); err != nil {
		slog.Warn("Failed to write temp .env during PIN auto-migration", "error", err)
		os.Remove(tmpPath)
		return
	}
	if err := os.Rename(tmpPath, envPath); err != nil {
		slog.Warn("Failed to rename temp .env during PIN auto-migration", "error", err)
		os.Remove(tmpPath)
		return
	}

	slog.Info("PIN successfully hashed in .env file", "path", envPath)
}
