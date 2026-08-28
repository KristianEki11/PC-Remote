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

// QRPageHandler serves the HTML dashboard for pairing and managing connected devices.
// This endpoint is restricted to localhost.
type QRPageHandler struct {
	Pairing  *auth.PairingManager
	Sessions *auth.SessionManager
}

// ServeQRPage handles GET /internal/qr.
func (h *QRPageHandler) ServeQRPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Get all usable network interfaces on the host
	interfaces := tlsutil.GetAllLANInterfaces()
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

	// 2. Check if a device is already paired
	session := h.Sessions.GetActiveSession()
	isPaired := session != nil
	pairedDeviceName := ""
	pairedAddr := ""
	pairedSince := ""
	isOnline := false

	if isPaired {
		pairedDeviceName = session.DeviceName
		pairedAddr = session.RemoteAddr
		pairedSince = session.CreatedAt.Format("02 Jan 2006, 15:04 WIB")
		isOnline = session.IsOnline()
	}

	// 3. Generate encrypted pairing payload (AES-256 encrypted for PC Remote only)
	encryptedPayload, err := h.Pairing.GetEncryptedPayload()
	if err != nil {
		slog.Error("Failed to generate encrypted QR payload", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	rawPayload, _ := h.Pairing.GetQRPayload()
	interfacesJSON, _ := json.Marshal(interfaces)
	expiresAt := h.Pairing.ExpiresAt()
	expiresIn := int(time.Until(expiresAt).Seconds())

	statusText := "Siap Pairing (Belum Ada HP Terhubung)"
	statusClass := "disconnected"
	if isPaired {
		if isOnline {
			statusText = fmt.Sprintf("🟢 Terhubung: %s", pairedDeviceName)
			statusClass = "connected"
		} else {
			statusText = fmt.Sprintf("⚪ Tersimpan (Standby): %s", pairedDeviceName)
			statusClass = "standby"
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	page := qrPageHTML
	page = strings.ReplaceAll(page, "{{HOST}}", rawPayload.Host)
	page = strings.ReplaceAll(page, "{{PORT}}", rawPayload.Port)
	page = strings.ReplaceAll(page, "{{STATUS_TEXT}}", statusText)
	page = strings.ReplaceAll(page, "{{STATUS_CLASS}}", statusClass)
	page = strings.ReplaceAll(page, "{{ENCRYPTED_PAYLOAD}}", encryptedPayload)
	page = strings.ReplaceAll(page, "{{FINGERPRINT}}", rawPayload.Fingerprint)
	page = strings.ReplaceAll(page, "{{EXPIRES_IN}}", strconv.Itoa(expiresIn))
	page = strings.ReplaceAll(page, "{{INTERFACES_JSON}}", string(interfacesJSON))
	page = strings.ReplaceAll(page, "{{SERVER_NAME}}", rawPayload.ServerName)

	if isPaired {
		page = strings.ReplaceAll(page, "{{IS_PAIRED}}", "true")
	} else {
		page = strings.ReplaceAll(page, "{{IS_PAIRED}}", "false")
	}
	page = strings.ReplaceAll(page, "{{PAIRED_DEVICE_NAME}}", pairedDeviceName)
	page = strings.ReplaceAll(page, "{{PAIRED_ADDR}}", pairedAddr)
	page = strings.ReplaceAll(page, "{{PAIRED_SINCE}}", pairedSince)
	if isOnline {
		page = strings.ReplaceAll(page, "{{IS_ONLINE}}", "true")
	} else {
		page = strings.ReplaceAll(page, "{{IS_ONLINE}}", "false")
	}

	w.Write([]byte(page))
}

// ServeQRImage handles GET /internal/qr/image.
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

// ServeStatus handles GET /internal/status.
func (h *QRPageHandler) ServeStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := h.Sessions.GetActiveSession()
	connected, deviceName := h.Sessions.IsDeviceConnected()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"has_session": session != nil,
		"connected":   connected,
		"device_name": deviceName,
	})
}

const qrPageHTML = `<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>PC Remote — Device Pairing & Dashboard</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
<style>
  :root {
    --bg-base: #080A0F;
    --bg-surface: rgba(16, 20, 32, 0.85);
    --bg-surface-elevated: rgba(24, 30, 48, 0.7);
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
    --ruby-glow: rgba(244, 63, 94, 0.2);
    --ruby-border: rgba(244, 63, 94, 0.4);
    --ruby-text: #F87171;
    --radius-xl: 24px;
    --radius-lg: 16px;
    --radius-md: 12px;
  }

  * { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, sans-serif;
    background-color: var(--bg-base);
    color: var(--text-main);
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    background-image: 
      radial-gradient(ellipse 80% 50% at 50% -20%, rgba(99, 102, 241, 0.2), transparent),
      radial-gradient(ellipse 60% 40% at 80% 100%, rgba(217, 70, 239, 0.12), transparent);
    background-attachment: fixed;
  }

  .container {
    width: 100%;
    max-width: 900px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  /* Header Section */
  .header {
    background: var(--bg-surface);
    backdrop-filter: blur(20px);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-xl);
    padding: 24px 32px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    box-shadow: 0 20px 40px -15px rgba(0, 0, 0, 0.5);
  }

  .brand { display: flex; align-items: center; gap: 16px; }
  .logo-badge {
    width: 48px;
    height: 48px;
    border-radius: 14px;
    background: var(--primary-gradient);
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 8px 20px var(--primary-glow);
  }
  .logo-badge svg { width: 26px; height: 26px; fill: #fff; }
  .title-group h1 { font-size: 22px; font-weight: 800; letter-spacing: -0.5px; }
  .title-group p { font-size: 13px; color: var(--text-muted); margin-top: 2px; }

  /* Status Pill */
  .status-pill {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 16px;
    border-radius: 9999px;
    font-size: 13px;
    font-weight: 600;
    border: 1px solid var(--border-subtle);
    background: rgba(255, 255, 255, 0.03);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }
  .status-pill.connected {
    background: var(--emerald-glow);
    border-color: var(--emerald-border);
    color: var(--emerald-text);
  }
  .status-pill.standby {
    background: rgba(251, 191, 36, 0.15);
    border-color: rgba(251, 191, 36, 0.35);
    color: #FBBF24;
  }
  .status-pill.disconnected {
    background: rgba(255, 255, 255, 0.05);
    border-color: var(--border-subtle);
    color: var(--text-muted);
  }
  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: currentColor;
    box-shadow: 0 0 10px currentColor;
    animation: pulse 2s infinite ease-in-out;
  }
  @keyframes pulse { 0%, 100% { opacity: 1; transform: scale(1); } 50% { opacity: 0.4; transform: scale(0.85); } }

  /* Main Grid */
  .main-grid {
    display: grid;
    grid-template-columns: 380px 1fr;
    gap: 20px;
  }
  @media (max-width: 840px) {
    .main-grid { grid-template-columns: 1fr; }
  }

  .card {
    background: var(--bg-surface);
    backdrop-filter: blur(20px);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-xl);
    padding: 28px;
    display: flex;
    flex-direction: column;
    box-shadow: 0 20px 40px -15px rgba(0, 0, 0, 0.5);
  }

  /* QR Box & States */
  .qr-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 18px;
    height: 100%;
  }

  .qr-frame {
    position: relative;
    padding: 16px;
    background: #FFFFFF;
    border-radius: var(--radius-lg);
    box-shadow: 0 12px 30px rgba(0, 0, 0, 0.4);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: transform 0.3s ease;
  }
  .qr-frame:hover { transform: translateY(-2px); }
  .qr-frame canvas { display: block; width: 220px; height: 220px; }

  /* Paired Device Card */
  .paired-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 16px;
    padding: 24px 16px;
    background: var(--bg-surface-elevated);
    border: 1px solid var(--emerald-border);
    border-radius: var(--radius-lg);
    width: 100%;
  }
  .paired-icon {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    background: var(--emerald-glow);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--emerald-text);
  }
  .paired-device-name {
    font-size: 18px;
    font-weight: 700;
    color: #FFFFFF;
  }
  .paired-meta {
    font-size: 12px;
    color: var(--text-muted);
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  /* Buttons */
  .btn-unpair {
    width: 100%;
    padding: 12px 20px;
    border-radius: var(--radius-md);
    background: rgba(244, 63, 94, 0.15);
    border: 1px solid var(--ruby-border);
    color: var(--ruby-text);
    font-size: 13px;
    font-weight: 700;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    transition: all 0.2s ease;
  }
  .btn-unpair:hover {
    background: rgba(244, 63, 94, 0.25);
    transform: translateY(-1px);
  }

  /* Adapter Selector */
  .selector-group {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .selector-label {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  .custom-select {
    width: 100%;
    padding: 10px 14px;
    background: var(--bg-surface-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-main);
    font-size: 13px;
    font-family: inherit;
    outline: none;
    cursor: pointer;
    transition: border-color 0.2s ease;
  }
  .custom-select:focus { border-color: #6366F1; }

  /* Info Matrix */
  .info-matrix {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-top: auto;
  }
  .info-tile {
    background: var(--bg-surface-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 14px 18px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }
  .tile-left { display: flex; flex-direction: column; gap: 2px; }
  .tile-title { font-size: 11px; font-weight: 600; color: var(--text-dim); text-transform: uppercase; letter-spacing: 0.5px; }
  .tile-val { font-family: 'JetBrains Mono', monospace; font-size: 13px; color: var(--text-main); }
  .tile-val.masked { letter-spacing: 2px; color: var(--text-dim); }

  .btn-unlock {
    padding: 6px 12px;
    background: rgba(99, 102, 241, 0.15);
    border: 1px solid rgba(99, 102, 241, 0.35);
    color: #A5B4FC;
    border-radius: 8px;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 6px;
    transition: all 0.2s ease;
  }
  .btn-unlock:hover {
    background: rgba(99, 102, 241, 0.3);
  }

  /* Guide Steps */
  .guide-list {
    display: flex;
    flex-direction: column;
    gap: 14px;
    margin: 16px 0;
  }
  .guide-step {
    display: flex;
    gap: 14px;
    align-items: flex-start;
  }
  .step-num {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: var(--primary-gradient);
    font-size: 13px;
    font-weight: 700;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }
  .step-content h4 { font-size: 14px; font-weight: 700; margin-bottom: 2px; }
  .step-content p { font-size: 12px; color: var(--text-muted); line-height: 1.5; }

  /* PIN Modal */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.75);
    backdrop-filter: blur(10px);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.25s ease;
    z-index: 1000;
  }
  .modal-overlay.active {
    opacity: 1;
    pointer-events: auto;
  }
  .modal-box {
    background: #121622;
    border: 1px solid var(--border-hover);
    border-radius: var(--radius-xl);
    padding: 28px;
    width: 100%;
    max-width: 360px;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.8);
    display: flex;
    flex-direction: column;
    gap: 18px;
    transform: translateY(20px);
    transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  }
  .modal-overlay.active .modal-box { transform: translateY(0); }
  .modal-box h3 { font-size: 18px; font-weight: 800; }
  .modal-box p { font-size: 13px; color: var(--text-muted); }
  .pin-input {
    width: 100%;
    padding: 12px 16px;
    font-size: 18px;
    text-align: center;
    letter-spacing: 4px;
    border-radius: var(--radius-md);
    background: var(--bg-surface-elevated);
    border: 1px solid var(--border-subtle);
    color: #FFFFFF;
    outline: none;
  }
  .pin-input:focus { border-color: #6366F1; }
  .modal-actions { display: flex; gap: 10px; }
  .btn-primary {
    flex: 1;
    padding: 12px;
    border-radius: var(--radius-md);
    background: var(--primary-gradient);
    color: #FFFFFF;
    font-weight: 700;
    border: none;
    cursor: pointer;
  }
  .btn-secondary {
    padding: 12px 18px;
    border-radius: var(--radius-md);
    background: transparent;
    border: 1px solid var(--border-subtle);
    color: var(--text-muted);
    font-weight: 600;
    cursor: pointer;
  }
  .error-text { color: var(--ruby-text); font-size: 12px; text-align: center; display: none; }
</style>
<script src="https://cdn.jsdelivr.net/npm/qrcode-generator@1.4.4/qrcode.min.js"></script>
</head>
<body>

<div class="container">
  <!-- Header -->
  <header class="header">
    <div class="brand">
      <div class="logo-badge">
        <svg viewBox="0 0 24 24"><path d="M4 6h16v10H4z M2 18h20v2H2z"/></svg>
      </div>
      <div class="title-group">
        <h1>PC Remote Dashboard</h1>
        <p>{{SERVER_NAME}} • v3.0.0</p>
      </div>
    </div>
    <div class="status-pill {{STATUS_CLASS}}" id="statusPill">
      <span class="status-dot"></span>
      <span id="statusText">{{STATUS_TEXT}}</span>
    </div>
  </header>

  <!-- Main Content -->
  <div class="main-grid">
    <!-- Left: QR / Paired Card -->
    <div class="card">
      <div class="qr-container">
        <!-- If already paired -->
        <div id="pairedView" style="display: {{IS_PAIRED}} ? 'flex' : 'none'; width: 100%; flex-direction: column; gap: 16px;">
          <div class="paired-card">
            <div class="paired-icon">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="5" y="2" width="14" height="20" rx="2" ry="2"/><line x1="12" y1="18" x2="12.01" y2="18"/></svg>
            </div>
            <div>
              <div class="paired-device-name" id="pairedDeviceName">{{PAIRED_DEVICE_NAME}}</div>
              <div class="paired-meta">
                <span>Alamat: <strong style="color: #fff;" id="pairedAddr">{{PAIRED_ADDR}}</strong></span>
                <span>Terhubung Sejak: <span id="pairedSince">{{PAIRED_SINCE}}</span></span>
              </div>
            </div>
            <div style="font-size: 12px; color: var(--emerald-text); background: var(--emerald-glow); padding: 4px 12px; border-radius: 999px; border: 1px solid var(--emerald-border);">
              🔒 Sesi Aktif & Terdaftar Permanen
            </div>
          </div>
          <button class="btn-unpair" onclick="unpairDevice()">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18.36 6.64a9 9 0 1 1-12.73 0"/><line x1="12" y1="2" x2="12" y2="12"/></svg>
            Putuskan Perangkat (Unpair)
          </button>
        </div>

        <!-- If not paired: Show QR -->
        <div id="unpairedView" style="display: {{IS_PAIRED}} ? 'none' : 'flex'; flex-direction: column; align-items: center; gap: 16px; width: 100%;">
          <div class="selector-group">
            <label class="selector-label">Pilih Jalur Jaringan (Adapter)</label>
            <select class="custom-select" id="networkSelect" onchange="switchAdapter(this.value)">
              <!-- Options populated by JS -->
            </select>
          </div>

          <div class="qr-frame">
            <div id="qrCanvas"></div>
          </div>

          <div style="font-size: 12px; color: var(--text-dim); text-align: center;">
            🔐 Payload QR dienkripsi (AES-256) khusus untuk PC Remote App
          </div>
        </div>
      </div>
    </div>

    <!-- Right: Info Matrix & Steps -->
    <div class="card" style="justify-content: space-between;">
      <div>
        <h3 style="font-size: 16px; font-weight: 700; margin-bottom: 4px;">Panduan Koneksi Cepat</h3>
        <p style="font-size: 13px; color: var(--text-muted);">Hanya perlu 1x scan QR untuk menghubungkan HP Anda selamanya.</p>

        <div class="guide-list">
          <div class="guide-step">
            <div class="step-num">1</div>
            <div class="step-content">
              <h4>Buka PC Remote di HP</h4>
              <p>Pastikan HP terhubung ke Wi-Fi yang sama atau tersambung kabel USB.</p>
            </div>
          </div>
          <div class="guide-step">
            <div class="step-num">2</div>
            <div class="step-content">
              <h4>Pindai QR Code di Kiri</h4>
              <p>Tekan tombol "Pindai QR Code" pada aplikasi HP dan arahkan ke layar.</p>
            </div>
          </div>
          <div class="guide-step">
            <div class="step-num">3</div>
            <div class="step-content">
              <h4>Selesai & Otomatis Tersimpan</h4>
              <p>Sesi HP akan tersimpan permanen tanpa perlu scan ulang saat buka app.</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Security Info Matrix -->
      <div class="info-matrix">
        <div class="info-tile">
          <div class="tile-left">
            <span class="tile-title">Host Server Lokal</span>
            <span class="tile-val" id="tileHost">{{HOST}}:{{PORT}}</span>
          </div>
        </div>

        <div class="info-tile">
          <div class="tile-left">
            <span class="tile-title">TLS SHA-256 Fingerprint</span>
            <span class="tile-val masked" id="fingerprintVal">••••••••••••••••••••••••••••••••••••••••</span>
          </div>
          <button class="btn-unlock" id="btnUnlockFp" onclick="openPinModal()">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
            Buka Kunci (PIN)
          </button>
        </div>
      </div>
    </div>
  </div>
</div>

<!-- Master PIN Modal -->
<div class="modal-overlay" id="pinModal">
  <div class="modal-box">
    <h3>Verifikasi PIN Server</h3>
    <p>Masukkan PIN PC Remote untuk membuka detail sidik jari sertifikat TLS.</p>
    <input type="password" id="pinInput" class="pin-input" placeholder="••••" maxlength="8" autofocus onkeydown="if(event.key==='Enter') submitPinVerification()">
    <div class="error-text" id="pinError">PIN yang Anda masukkan salah.</div>
    <div class="modal-actions">
      <button class="btn-secondary" onclick="closePinModal()">Batal</button>
      <button class="btn-primary" onclick="submitPinVerification()">Buka Kunci</button>
    </div>
  </div>
</div>

<script>
  const isPairedInit = {{IS_PAIRED}};
  const realFingerprint = "{{FINGERPRINT}}";
  const encryptedPayload = "{{ENCRYPTED_PAYLOAD}}";
  const networkInterfaces = {{INTERFACES_JSON}};

  // Init Adapter selector & QR
  function initPage() {
    const isPaired = isPairedInit;
    document.getElementById('pairedView').style.display = isPaired ? 'flex' : 'none';
    document.getElementById('unpairedView').style.display = isPaired ? 'none' : 'flex';

    if (!isPaired) {
      const select = document.getElementById('networkSelect');
      select.innerHTML = '';
      networkInterfaces.forEach(iface => {
        const opt = document.createElement('option');
        opt.value = iface.ip;
        opt.textContent = iface.name + ' (' + iface.ip + ')' + (iface.is_primary ? ' — Rekomendasi' : '');
        select.appendChild(opt);
      });
      renderQR(encryptedPayload);
    }
  }

  function renderQR(text) {
    try {
      const qr = qrcode(0, 'M');
      qr.addData(text);
      qr.make();
      document.getElementById('qrCanvas').innerHTML = qr.createSvgTag({ scalable: true, margin: 0 });
      const svg = document.querySelector('#qrCanvas svg');
      if (svg) {
        svg.style.width = '220px';
        svg.style.height = '220px';
      }
    } catch (e) {
      console.error('QR render error:', e);
    }
  }

  function switchAdapter(ip) {
    document.getElementById('tileHost').textContent = ip + ':{{PORT}}';
    // Encrypted payload contains all alternate hosts, so it works across adapters
  }

  // Unpair Action
  async function unpairDevice() {
    if (!confirm('Apakah Anda yakin ingin memutuskan perangkat ini? HP harus scan QR ulang untuk tersambung kembali.')) return;
    try {
      const res = await fetch('/internal/unpair', { method: 'POST' });
      if (res.ok) {
        window.location.reload();
      }
    } catch (e) {
      alert('Gagal memutuskan perangkat: ' + e);
    }
  }

  // PIN Unlock Modal
  function openPinModal() {
    document.getElementById('pinModal').classList.add('active');
    document.getElementById('pinInput').value = '';
    document.getElementById('pinError').style.display = 'none';
    document.getElementById('pinInput').focus();
  }

  function closePinModal() {
    document.getElementById('pinModal').classList.remove('active');
  }

  async function submitPinVerification() {
    const pin = document.getElementById('pinInput').value.trim();
    if (!pin) return;

    try {
      const res = await fetch('/internal/verify-pin', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ pin: pin })
      });

      if (res.ok) {
        // Unlock fingerprint display
        const fpEl = document.getElementById('fingerprintVal');
        fpEl.textContent = realFingerprint;
        fpEl.classList.remove('masked');
        document.getElementById('btnUnlockFp').style.display = 'none';
        closePinModal();
      } else {
        document.getElementById('pinError').style.display = 'block';
      }
    } catch (e) {
      document.getElementById('pinError').textContent = 'Error: ' + e;
      document.getElementById('pinError').style.display = 'block';
    }
  }

  // Poll status every 4 seconds
  setInterval(async () => {
    try {
      const res = await fetch('/internal/status');
      if (res.ok) {
        const data = await res.json();
        // If state changed from unpaired to paired, reload to switch view
        if (data.has_session !== isPairedInit) {
          window.location.reload();
        }
      }
    } catch (_) {}
  }, 4000);

  initPage();
</script>

</body>
</html>`
