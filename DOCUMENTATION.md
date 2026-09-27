# PC Remote - Development Documentation & Journey

## 1. Introduction
This document details the development journey, architectural decisions, obstacles encountered, and solutions implemented for the PC Remote Controller project up to version 5.0.0.

## 2. Evolution and Version History
- **Early Versions (Python-based)**: The initial versions of the PC Remote backend were built using Python and FastAPI. While functional, it faced significant challenges with deployment, performance, and antivirus false positives when packaged with PyInstaller.
- **Version 5.0.0 (Go Rewrite)**: A complete rewrite of the backend in Go. This version introduced Cloudflare Quick Tunnels for WAN access, a much smaller and faster native executable, better Windows COM API handling, and Screen Mirroring capabilities.

## 3. Major Obstacles and Solutions

### A. Deployment and Antivirus Flags (Python -> Go)
**Obstacle**: Distributing the Python backend required PyInstaller. This resulted in huge executables (100MB+) and frequent false-positive flags from Windows Defender and other antivirus software, making installation a poor experience for users.
**Solution**: Rewrote the entire backend in Go. Go compiles to native, statically linked executables that are small (~10MB), incredibly fast, and very rarely flagged by AV heuristics. The transition also improved concurrency and memory usage significantly.

### B. Windows Audio API and COM Threading
**Obstacle**: Controlling volume and muting applications required interacting with Windows Core Audio APIs (`IMMDeviceEnumerator`, `IAudioEndpointVolume`, etc.). These are COM interfaces. In Python, managing Single-Threaded Apartments (STA) was brittle and led to deadlocks.
**Solution**: In Go, we utilized `github.com/go-ole/go-ole`. We carefully locked goroutines to OS threads (`runtime.LockOSThread()`) during COM operations, ensuring stable, reliable communication with Windows Audio endpoints without crashing or deadlocking.

### C. Seamless Public Internet Access (WAN)
**Obstacle**: Users wanted to control their PCs from outside the local network (e.g., on 4G/5G). Traditional solutions required complex router port-forwarding or DDNS, which is too difficult for average users.
**Solution**: Integrated **Cloudflare Quick Tunnels** directly into the Go backend. The server automatically spins up a secure tunnel upon startup, generating a unique `.trycloudflare.com` URL. The Flutter app was upgraded with an "Intelligent Dual-Stack Auto-Failover" that seamlessly switches between the local IP and the Cloudflare Tunnel depending on network availability.

### D. Secure Pairing via QR Code
**Obstacle**: Entering long IP addresses, ports, and PINs manually was tedious.
**Solution**: Implemented an AES-encrypted QR code pairing system. The Windows dashboard generates a QR code containing the local IP, tunnel URL, and cryptographic tokens. The mobile app scans this QR, decrypts it, and instantly pairs with the PC, verifying TLS certificates to prevent MITM attacks.

### E. Screen Mirroring Latency
**Obstacle**: Streaming the PC screen to the mobile app needed to be real-time with very low latency, which HTTP polling or standard HLS couldn't provide.
**Solution**: Implemented a custom WebSocket + H.264/MJPEG streaming protocol. We use a 13-byte binary header to pack frame metadata efficiently. On LAN, this achieves ~80-150ms latency. The Flutter app uses a custom canvas renderer to display the frames and map touch inputs back to the PC.

## 4. Web Client for iOS (iPhone)
**Obstacle**: Apple's App Store policies and the requirement for a Mac to compile iOS apps meant we couldn't easily distribute a native iPhone app.
**Solution**: We developed a Progressive Web App (PWA) client. iPhone users can add the web dashboard to their home screen via Safari. The web app is currently being rebuilt (Next.js) to include a dedicated landing page and an isolated `/connect` route for the actual remote control interface.

## 5. Copyright and Licensing
**Copyright (c) 2026 kidev. All Rights Reserved.**
This project uses a strict custom license:
- Free for personal use.
- **NO** commercial use allowed.
- **NO** redistribution or incorporation into other projects, whether in whole or in part.
This ensures the software remains available for individuals while protecting the intellectual property from unauthorized commercial exploitation.
