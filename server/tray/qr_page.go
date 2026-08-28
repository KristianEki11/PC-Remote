package tray

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pcremote-server/auth"
	tlsutil "pcremote-server/tls"
)

// QRPageHandler serves the HTML page that displays the QR code for device pairing.
// This endpoint is protected by the LocalhostOnly middleware.
type QRPageHandler struct {
	Pairing  *auth.PairingManager
	Sessions *auth.SessionManager
}

// ServeQRPage handles GET /internal/qr — renders the full modern HTML page with embedded QR code.
func (h *QRPageHandler) ServeQRPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Get all usable network interfaces on the host
	interfaces := tlsutil.GetAllLANInterfaces()
	
	// Add USB Cable / Localhost option for ADB reverse connection
	interfaces = append(interfaces, tlsutil.InterfaceInfo{
		Name:      "USB Cable / Localhost",
		IP:        "127.0.0.1",
		IsPrimary: false,
	})

	if len(interfaces) > 0 {
		h.Pairing.UpdateHost(interfaces[0].IP)
		var alts []string
		for _, iface := range interfaces[1:] {
			alts = append(alts, iface.IP)
		}
		h.Pairing.UpdateAlternateHosts(alts)
	}

	// 2. Generate a fresh pairing payload (new one-time token)
	payload, err := h.Pairing.GetQRPayload()
	if err != nil {
		slog.Error("Failed to generate QR payload", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	payloadJSON, _ := json.Marshal(payload)
	interfacesJSON, _ := json.Marshal(interfaces)
	expiresAt := h.Pairing.ExpiresAt()
	expiresIn := int(time.Until(expiresAt).Seconds())

	// 3. Check connection status
	connected, deviceName := h.Sessions.IsDeviceConnected()
	statusText := "Menunggu Sambungan HP"
	statusClass := "disconnected"
	if connected {
		statusText = fmt.Sprintf("Terhubung: %s", deviceName)
		statusClass = "connected"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	// Render template safely without printf format collision
	page := qrPageHTML
	page = strings.ReplaceAll(page, "{{HOST}}", payload.Host)
	page = strings.ReplaceAll(page, "{{PORT}}", payload.Port)
	page = strings.ReplaceAll(page, "{{STATUS_TEXT}}", statusText)
	page = strings.ReplaceAll(page, "{{STATUS_CLASS}}", statusClass)
	page = strings.ReplaceAll(page, "{{PAYLOAD_JSON}}", string(payloadJSON))
	page = strings.ReplaceAll(page, "{{FINGERPRINT}}", payload.Fingerprint)
	page = strings.ReplaceAll(page, "{{EXPIRES_IN}}", strconv.Itoa(expiresIn))
	page = strings.ReplaceAll(page, "{{INTERFACES_JSON}}", string(interfacesJSON))
	page = strings.ReplaceAll(page, "{{SERVER_NAME}}", payload.ServerName)

	w.Write([]byte(page))
}

// ServeQRImage handles GET /internal/qr/image — returns the QR code as a PNG image.
func (h *QRPageHandler) ServeQRImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pngData, err := h.Pairing.GetQRImage(300)
	if err != nil {
		slog.Error("Failed to generate QR image", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(pngData)
}

// ServeStatus handles GET /internal/status — returns connection status as JSON.
func (h *QRPageHandler) ServeStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	connected, deviceName := h.Sessions.IsDeviceConnected()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"connected":   connected,
		"device_name": deviceName,
	})
}

const qrPageHTML = `<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>PC Remote Connect — Pairing Dashboard</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
<style>
  :root {
    --bg-base: #090B11;
    --bg-surface: rgba(18, 22, 34, 0.78);
    --bg-surface-elevated: rgba(26, 32, 50, 0.65);
    --border-subtle: rgba(255, 255, 255, 0.08);
    --border-hover: rgba(99, 102, 241, 0.4);
    --primary-gradient: linear-gradient(135deg, #6366F1 0%, #8B5CF6 50%, #D946EF 100%);
    --primary-glow: rgba(99, 102, 241, 0.35);
    --text-main: #FFFFFF;
    --text-muted: #94A3B8;
    --text-dim: #64748B;
    --emerald-glow: rgba(16, 185, 129, 0.2);
    --emerald-border: rgba(16, 185, 129, 0.4);
    --emerald-text: #34D399;
    --ruby-text: #F87171;
    --amber-text: #FBBF24;
    --radius-xl: 28px;
    --radius-lg: 20px;
    --radius-md: 14px;
    --radius-sm: 10px;
  }

  * { margin: 0; padding: 0; box-sizing: border-box; }

  body {
    font-family: 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    background-color: var(--bg-base);
    color: var(--text-main);
    min-height: 100vh;
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 32px 20px;
    overflow-x: hidden;
    position: relative;
    -webkit-font-smoothing: antialiased;
  }

  /* Ambient Cosmic Glow Background Blobs */
  .glow-blob {
    position: fixed;
    border-radius: 50%;
    filter: blur(120px);
    z-index: 0;
    pointer-events: none;
    opacity: 0.45;
  }
  .glow-1 {
    width: 500px;
    height: 500px;
    background: radial-gradient(circle, #4F46E5 0%, rgba(79, 70, 229, 0) 70%);
    top: -150px;
    left: -100px;
  }
  .glow-2 {
    width: 600px;
    height: 600px;
    background: radial-gradient(circle, #7C3AED 0%, rgba(124, 58, 237, 0) 70%);
    bottom: -200px;
    right: -150px;
  }

  /* Main Card Container */
  .container {
    position: relative;
    z-index: 1;
    width: 100%;
    max-width: 540px;
    background: var(--bg-surface);
    backdrop-filter: blur(28px);
    -webkit-backdrop-filter: blur(28px);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-xl);
    box-shadow: 
      0 24px 70px rgba(0, 0, 0, 0.65),
      inset 0 1px 1px rgba(255, 255, 255, 0.12);
    padding: 36px 32px;
    text-align: center;
  }

  /* Header Section */
  .brand-header {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    margin-bottom: 24px;
  }

  .brand-badge {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 6px 14px;
    background: rgba(99, 102, 241, 0.1);
    border: 1px solid rgba(99, 102, 241, 0.25);
    border-radius: 999px;
    font-size: 0.78rem;
    font-weight: 700;
    color: #A5B4FC;
    letter-spacing: 0.5px;
    text-transform: uppercase;
  }

  .brand-title {
    font-size: 1.85rem;
    font-weight: 800;
    letter-spacing: -0.6px;
    color: #FFFFFF;
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .brand-title span.accent {
    background: var(--primary-gradient);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .brand-subtitle {
    font-size: 0.9rem;
    color: var(--text-muted);
    max-width: 380px;
    line-height: 1.45;
  }

  /* Status Indicator Pill */
  .status-pill {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    padding: 8px 18px;
    border-radius: 999px;
    font-size: 0.86rem;
    font-weight: 600;
    margin-bottom: 24px;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .status-pill.connected {
    background: var(--emerald-glow);
    border: 1px solid var(--emerald-border);
    color: var(--emerald-text);
  }

  .status-pill.disconnected {
    background: rgba(148, 163, 184, 0.08);
    border: 1px solid rgba(148, 163, 184, 0.16);
    color: var(--text-muted);
  }

  .pulse-dot {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    background-color: currentColor;
    position: relative;
  }

  .status-pill.connected .pulse-dot::after {
    content: '';
    position: absolute;
    top: -3px; left: -3px; right: -3px; bottom: -3px;
    border-radius: 50%;
    background-color: var(--emerald-text);
    animation: pulse 1.8s infinite cubic-bezier(0.4, 0, 0.2, 1);
    opacity: 0.7;
  }

  @keyframes pulse {
    0% { transform: scale(1); opacity: 0.8; }
    50% { transform: scale(2.2); opacity: 0; }
    100% { transform: scale(1); opacity: 0; }
  }

  /* Network Selector Section */
  .network-selector-box {
    background: var(--bg-surface-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 12px 16px;
    margin-bottom: 22px;
    text-align: left;
  }

  .network-label-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }

  .network-title {
    font-size: 0.78rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-dim);
  }

  .network-hint {
    font-size: 0.75rem;
    color: #818CF8;
    font-weight: 600;
  }

  .custom-select-wrapper {
    position: relative;
    width: 100%;
  }

  .network-select {
    width: 100%;
    appearance: none;
    -webkit-appearance: none;
    background: rgba(10, 14, 24, 0.85);
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: var(--radius-sm);
    color: #F1F5F9;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.88rem;
    font-weight: 500;
    padding: 10px 38px 10px 14px;
    outline: none;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .network-select:focus, .network-select:hover {
    border-color: #6366F1;
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.2);
  }

  .select-chevron {
    position: absolute;
    right: 14px;
    top: 50%;
    transform: translateY(-50%);
    pointer-events: none;
    color: var(--text-dim);
  }

  /* QR Frame Card */
  .qr-card {
    position: relative;
    display: inline-block;
    padding: 20px;
    background: #FFFFFF;
    border-radius: var(--radius-lg);
    box-shadow: 
      0 12px 36px rgba(0, 0, 0, 0.4),
      0 0 40px var(--primary-glow);
    margin-bottom: 20px;
    transition: transform 0.3s ease;
  }

  .qr-card:hover {
    transform: scale(1.015);
  }

  .qr-card canvas {
    display: block;
    border-radius: 8px;
  }

  /* Connection Info Matrix */
  .info-matrix {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
    margin-bottom: 22px;
    text-align: left;
  }

  .info-tile {
    background: var(--bg-surface-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 12px 14px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    position: relative;
    transition: border-color 0.2s ease;
  }

  .info-tile:hover {
    border-color: var(--border-hover);
  }

  .tile-label {
    font-size: 0.74rem;
    font-weight: 600;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.4px;
  }

  .tile-value {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.85rem;
    color: #F8FAFC;
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .info-tile-copy {
    cursor: pointer;
  }

  /* Expiry Countdown Bar */
  .timer-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    margin-bottom: 24px;
  }

  .timer-text {
    font-size: 0.85rem;
    color: var(--text-muted);
  }

  .timer-counter {
    font-family: 'JetBrains Mono', monospace;
    font-weight: 700;
    color: var(--amber-text);
  }

  .progress-track {
    width: 100%;
    height: 4px;
    background: rgba(255, 255, 255, 0.08);
    border-radius: 999px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    width: 100%;
    background: var(--primary-gradient);
    border-radius: 999px;
    transition: width 1s linear;
  }

  /* Action Buttons */
  .btn-primary {
    width: 100%;
    padding: 14px 20px;
    background: var(--primary-gradient);
    border: none;
    border-radius: var(--radius-md);
    color: #FFFFFF;
    font-family: inherit;
    font-size: 0.95rem;
    font-weight: 700;
    letter-spacing: 0.3px;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    box-shadow: 0 10px 25px var(--primary-glow);
    transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .btn-primary:hover {
    transform: translateY(-2px);
    box-shadow: 0 14px 30px rgba(99, 102, 241, 0.5);
  }

  .btn-primary:active {
    transform: translateY(0);
  }

  .btn-primary svg {
    transition: transform 0.4s ease;
  }

  .btn-primary:hover svg {
    transform: rotate(180deg);
  }

  /* 3-Step Guide Drawer */
  .guide-card {
    margin-top: 24px;
    background: rgba(10, 14, 24, 0.5);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 18px 20px;
    text-align: left;
  }

  .guide-header {
    font-size: 0.84rem;
    font-weight: 700;
    color: #E2E8F0;
    margin-bottom: 12px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .guide-steps {
    list-style: none;
    display: flex;
    flex-direction: column;
    gap: 8px;
    font-size: 0.82rem;
    color: var(--text-muted);
  }

  .guide-steps li {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .step-num {
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: rgba(99, 102, 241, 0.2);
    color: #A5B4FC;
    font-size: 0.72rem;
    font-weight: 700;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  /* Toast Notification */
  .toast {
    position: fixed;
    bottom: 24px;
    left: 50%;
    transform: translateX(-50%) translateY(100px);
    background: #1E293B;
    color: #F8FAFC;
    border: 1px solid rgba(255, 255, 255, 0.1);
    padding: 10px 20px;
    border-radius: 999px;
    font-size: 0.84rem;
    font-weight: 600;
    box-shadow: 0 10px 30px rgba(0,0,0,0.5);
    opacity: 0;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    z-index: 100;
    pointer-events: none;
  }

  .toast.show {
    transform: translateX(-50%) translateY(0);
    opacity: 1;
  }
</style>
<!-- QRCode.js (MIT License) -->
<script src="https://cdn.jsdelivr.net/npm/qrcode-generator@1.4.4/qrcode.min.js"></script>
</head>
<body>

<div class="glow-blob glow-1"></div>
<div class="glow-blob glow-2"></div>

<div class="container">
  <!-- Brand Header -->
  <div class="brand-header">
    <div class="brand-badge">
      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
      TLS 1.3 SECURED
    </div>
    <h1 class="brand-title">
      PC Remote <span class="accent">Connect</span>
    </h1>
    <p class="brand-subtitle">Pindai QR Code di bawah dengan aplikasi HP untuk menghubungkan remote controller.</p>
  </div>

  <!-- Live Status Pill -->
  <div class="status-pill {{STATUS_CLASS}}" id="statusPill">
    <div class="pulse-dot"></div>
    <span id="statusLabel">{{STATUS_TEXT}}</span>
  </div>

  <!-- Network Adapter Selector -->
  <div class="network-selector-box">
    <div class="network-label-row">
      <span class="network-title">Pilih Jaringan / IP Adapter</span>
      <span class="network-hint" id="adapterCount"></span>
    </div>
    <div class="custom-select-wrapper">
      <select class="network-select" id="networkSelect" onchange="onNetworkChange()">
        <!-- Populated via JavaScript -->
      </select>
      <svg class="select-chevron" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"></polyline></svg>
    </div>
  </div>

  <!-- QR Code Showcase Card -->
  <div class="qr-card">
    <canvas id="qr-canvas"></canvas>
  </div>

  <!-- Expiry Countdown & Progress -->
  <div class="timer-section">
    <div class="timer-text">
      Token berlaku hingga <span class="timer-counter" id="countdown">--:--</span>
    </div>
    <div class="progress-track">
      <div class="progress-fill" id="progressFill"></div>
    </div>
  </div>

  <!-- Connection Details Matrix -->
  <div class="info-matrix">
    <div class="info-tile info-tile-copy" onclick="copyText(currentSelectedHost + ':{{PORT}}', 'IP & Port disalin!')">
      <span class="tile-label">Alamat Server</span>
      <span class="tile-value" id="tileServer">{{HOST}}:{{PORT}}</span>
    </div>
    <div class="info-tile">
      <span class="tile-label">Protokol Keamanan</span>
      <span class="tile-value">HTTPS (TLS)</span>
    </div>
    <div class="info-tile info-tile-copy" style="grid-column: span 2;" onclick="copyText('{{FINGERPRINT}}', 'Fingerprint disalin!')">
      <span class="tile-label">TLS Cert Fingerprint (SHA-256)</span>
      <span class="tile-value" id="tileFp">{{FINGERPRINT}}</span>
    </div>
  </div>

  <!-- Action Button -->
  <button class="btn-primary" onclick="location.reload()">
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M21.5 2v6h-6M21.34 15.57a10 10 0 1 1-.57-8.38l5.67-5.67"/></svg>
    Buat Token Pairing Baru
  </button>

  <!-- 3-Step Guide -->
  <div class="guide-card">
    <div class="guide-header">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#818CF8" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>
      Langkah Mudah Pairing:
    </div>
    <ul class="guide-steps">
      <li><span class="step-num">1</span> <strong>Nyalakan Wi-Fi di HP</strong> & hubungkan ke Wi-Fi yang sama dengan PC (bukan kuota 4G).</li>
      <li><span class="step-num">2</span> Buka aplikasi PC Remote di HP & tekan tombol <strong>"Pindai QR Code di PC"</strong></li>
      <li><span class="step-num">3</span> Arahkan kamera HP ke QR Code di layar ini.</li>
      <li><span class="step-num">⚡</span> <em style="color:#A5B4FC;">Mode USB: Jika HP tersambung kabel USB, pilih adapter "USB Cable / Localhost (127.0.0.1)" di atas.</em></li>
    </ul>
  </div>
</div>

<div class="toast" id="toast">Teks disalin!</div>

<script>
  // 1. Data Injection from Server
  const initialPayload = {{PAYLOAD_JSON}};
  const networkInterfaces = {{INTERFACES_JSON}};
  const totalSeconds = {{EXPIRES_IN}};
  let remainingSeconds = totalSeconds;
  let currentSelectedHost = initialPayload.host;

  // 2. Populate Network Interfaces Dropdown
  const networkSelect = document.getElementById('networkSelect');
  const adapterCount = document.getElementById('adapterCount');

  if (networkInterfaces && networkInterfaces.length > 0) {
    adapterCount.textContent = networkInterfaces.length + ' adapter terdeteksi';
    networkSelect.innerHTML = '';
    networkInterfaces.forEach(iface => {
      const opt = document.createElement('option');
      opt.value = iface.ip;
      const isPri = iface.is_primary ? ' ★ Disarankan' : '';
      opt.textContent = iface.name + ' (' + iface.ip + ')' + isPri;
      if (iface.ip === currentSelectedHost) {
        opt.selected = true;
      }
      networkSelect.appendChild(opt);
    });
  } else {
    adapterCount.textContent = '1 adapter';
    networkSelect.innerHTML = '<option value="' + initialPayload.host + '">' + initialPayload.host + '</option>';
  }

  // 3. Render QR Code on Canvas
  function renderQRCode(host) {
    const payload = Object.assign({}, initialPayload, { host: host });
    
    // Include alternate hosts
    if (networkInterfaces && networkInterfaces.length > 1) {
      payload.alternate_hosts = networkInterfaces
        .map(i => i.ip)
        .filter(ip => ip !== host);
    }

    const qr = qrcode(0, 'M');
    qr.addData(JSON.stringify(payload));
    qr.make();

    const canvas = document.getElementById('qr-canvas');
    const size = 260;
    const count = qr.getModuleCount();
    const cellSize = size / count;
    
    canvas.width = size;
    canvas.height = size;
    const ctx = canvas.getContext('2d');
    
    // Clean background
    ctx.fillStyle = '#FFFFFF';
    ctx.fillRect(0, 0, size, size);

    // Render cells
    ctx.fillStyle = '#0F172A';
    for (let r = 0; r < count; r++) {
      for (let c = 0; c < count; c++) {
        if (qr.isDark(r, c)) {
          ctx.fillRect(c * cellSize, r * cellSize, cellSize + 0.4, cellSize + 0.4);
        }
      }
    }

    // Update UI labels
    document.getElementById('tileServer').textContent = host + ':{{PORT}}';
  }

  function onNetworkChange() {
    currentSelectedHost = networkSelect.value;
    renderQRCode(currentSelectedHost);
    showToast('QR Code diperbarui untuk ' + currentSelectedHost);
  }

  // Initial QR render
  renderQRCode(currentSelectedHost);

  // 4. Countdown Timer & Progress Bar
  const countdownEl = document.getElementById('countdown');
  const progressFill = document.getElementById('progressFill');

  function tickCountdown() {
    if (remainingSeconds <= 0) {
      countdownEl.textContent = 'KEDALUWARSA';
      countdownEl.style.color = 'var(--ruby-text)';
      progressFill.style.width = '0%';
      setTimeout(() => location.reload(), 1500);
      return;
    }

    const min = Math.floor(remainingSeconds / 60);
    const sec = remainingSeconds % 60;
    countdownEl.textContent = (min < 10 ? '0' : '') + min + ':' + (sec < 10 ? '0' : '') + sec;
    
    const pct = (remainingSeconds / totalSeconds) * 100;
    progressFill.style.width = pct + '%';
    
    remainingSeconds--;
    setTimeout(tickCountdown, 1000);
  }
  tickCountdown();

  // 5. Real-time Live Connection Polling
  setInterval(async () => {
    try {
      const resp = await fetch('/internal/status');
      const data = await resp.json();
      const pill = document.getElementById('statusPill');
      const label = document.getElementById('statusLabel');

      if (data.connected) {
        pill.className = 'status-pill connected';
        label.textContent = 'Terhubung: ' + data.device_name;
      } else {
        pill.className = 'status-pill disconnected';
        label.textContent = 'Menunggu Sambungan HP';
      }
    } catch(e) {}
  }, 4000);

  // 6. Toast Notification Helper
  function copyText(text, msg) {
    navigator.clipboard.writeText(text).then(() => {
      showToast(msg);
    });
  }

  function showToast(msg) {
    const t = document.getElementById('toast');
    t.textContent = msg;
    t.classList.add('show');
    setTimeout(() => t.classList.remove('show'), 2200);
  }
</script>
</body>
</html>`
