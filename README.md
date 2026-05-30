# 📱 PC Remote Controller (v2.1.0)

A secure, high-performance, and lightweight remote control suite that allows you to manage your Windows PC directly from your Android device over a local WiFi network. 

This repository contains the complete codebase for both the **Go-based backend server** (run as a Windows service) and the **Flutter-based Android application**.

---

## 🏗️ System Architecture

The project consists of three main components:

1. **Go Server (`/server`)**: A lightweight, concurrency-safe Windows service that listens for authenticated commands from the mobile app and executes them using Windows API integrations and COM interfaces.
2. **Flutter App (`/app`)**: A modern, responsive Android application featuring a clean dark-mode dashboard to control system volume, media playback, browser links, and power states.
3. **NSIS Installer (`/installer`)**: An installer script that packs the server executable, configures it as a Windows Service via NSSM, prompts for firewall rules, and guides the user to set up a secure Private network profile.

```
┌─────────────────┐       Local WiFi (HTTP)       ┌──────────────────┐
│   Android App   ├──────────────────────────────>│    Go Backend    │
│ (Flutter Client)│   PIN Auth & Bearer Token     │ (Windows Service)│
└─────────────────┘                               └────────┬─────────┘
                                                           │ Win32 API / COM
                                                           ▼
                                                  ┌──────────────────┐
                                                  │    Windows OS    │
                                                  │ (Volume, Media,  │
                                                  │  Power, Browser) │
                                                  └──────────────────┘
```

---

## ✨ Features

- 🔐 **PIN Authentication**: Secure authorization using custom PINs (4-8 digits). Supports atomic `.env` configuration updates directly from the Android app without restarting the server.
- 🔊 **Advanced Audio Control**: Retrieve and modify system master volume and mute state. Supports deep COM interfaces to query application sessions (including support for SteelSeries Sonar virtual channels).
- 🎵 **Media Injection**: Simulate system-wide media keys (Play/Pause, Next, Previous) to control media players (Spotify, YouTube, VLC, etc.) remotely.
- 🌐 **Remote Browser Launch**: Open any URL instantly in Microsoft Edge or the system's default browser.
- ⚡ **Power & Lifecycle Management**: Sleep, lock, restart, and shutdown commands. Includes cancellation endpoints to recover from accidental trigger commands.
- 📦 **Seamless Windows Service**: Runs quietly in the background. The installer automatically manages service states (`nssm start/stop`) during upgrades.
- 📶 **Installer Private Network Wizard**: Guides the user to configure their WiFi profile as "Private", avoiding typical connection blocks caused by Windows Defender Firewall.

---

## 📂 Repository Structure

```
.
├── app/                        # Android Flutter application
│   ├── lib/                    # Dart source code (screens, widgets, models)
│   └── pubspec.yaml            # Flutter project configuration
├── server/                     # Go server backend
│   ├── cmd/                    # Entrypoints & auxiliary tools
│   ├── config/                 # Environment & config handlers
│   ├── handlers/               # HTTP API handler logic & unit tests
│   ├── middleware/             # Auth & Rate limiters
│   ├── windows/                # Win32 & COM API wrappers
│   ├── dist/                   # Directory containing compiled server binary
│   └── go.mod                  # Go module definition
├── installer/                  # NSIS Installer project files
│   ├── PCRemoteSetup.nsi       # NSIS installation script
│   └── license.txt             # License terms displayed in installer
├── BRS_PRD_PCRemote.md         # Original product requirements document
└── MIGRATION_NOTES.md          # Reference notes for Python -> Go migration
```

---

## 🚀 Getting Started

### 1. Server Setup (Go)

#### Prerequisites
- Go 1.21 or newer (on Windows)
- Cgo is not strictly required, as we interact with COM and Win32 APIs dynamically.

#### Development Run
Navigate to the server directory, set up your local environment file, and run:
```powershell
cd server
copy .env.example .env
# Edit .env with your preferred APP_PIN and APP_PORT
go run main.go
```

#### Build
To build a production-ready, windowless Windows executable:
```powershell
go build -ldflags="-H windowsgui" -o dist/pcremote-server.exe main.go
```

---

### 2. Android App Setup (Flutter)

#### Prerequisites
- Flutter SDK (latest stable channel)
- Android SDK & Android device or emulator on the same network subnet as the PC.

#### Build & Run
```powershell
cd app
flutter pub get
flutter run --release
```

#### Build APK
```powershell
flutter build apk --release
```
The output will be generated at `app/build/app/outputs/flutter-apk/app-release.apk`.

---

### 3. Creating the Installer (NSIS)

#### Prerequisites
- [NSIS (Nullsoft Scriptable Install System)](https://nsis.sourceforge.io/) installed.
- NSSM (Non-Sucking Service Manager) binary placed under `installer/tools/nssm.exe` (or adjusted path).

#### Compilation
Compile `installer/PCRemoteSetup.nsi` using the NSIS Compiler interface or command line:
```powershell
makensis installer/PCRemoteSetup.nsi
```
This produces `installer/PCRemoteSetup.exe`.

---

## 🔒 Security Practices

- **Local-Only Communication**: The server rejects requests originating outside local network subnets (`192.168.x.x`, `10.x.x.x`, `172.16.x.x-172.31.x.x`).
- **Constant-Time Verification**: Uses Go's `crypto/subtle.ConstantTimeCompare` to avoid timing-based side-channel attacks during PIN validation.
- **Token Security**: Tokens are generated cryptographically (`crypto/rand`) and expire after 24 hours. They reside strictly in volatile memory.
- **Zero PIN Logs**: The PIN is never logged or exposed. 

---

## 📄 API Specification

All protected endpoints require the header `Authorization: Bearer <session_token>`.

| Endpoint | Method | Description |
|---|---|---|
| `/health` | `GET` | Retrieve server status and version details. No auth required. |
| `/auth` | `POST` | Authenticate using PIN. Returns a Bearer token. |
| `/system/pin` | `POST` | Update the auth PIN atomically. |
| `/audio/volume` | `GET` / `POST` | Get or set the master volume scalar (0-100). |
| `/audio/mute` | `POST` | Toggle volume mute state. |
| `/media/playpause` | `POST` | Inject Play/Pause key event. |
| `/media/next` | `POST` | Inject Skip Next key event. |
| `/media/prev` | `POST` | Inject Skip Previous key event. |
| `/browser/open` | `POST` | Opens the given URL in the default browser. |
| `/system/lock` | `POST` | Lock the active Windows session. |
| `/system/sleep` | `POST` | Suspend the system to S3 Sleep. |
| `/system/restart` | `POST` | Restart the system (5s timeout). |
| `/system/shutdown` | `POST` | Shut down the system (5s timeout). |

---

## 🛠️ Testing

Unit tests cover the server API routing, configuration updating, and authentication logic.
```powershell
cd server/handlers
go test -v
```

---

## 📝 License
This project is proprietary and for personal use. See [installer/license.txt](installer/license.txt) for more details.
