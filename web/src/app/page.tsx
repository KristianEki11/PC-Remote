"use client";

import { useState } from "react";
import { 
  Monitor, 
  Smartphone, 
  Globe, 
  Download, 
  ChevronRight, 
  Play, 
  Pause, 
  SkipForward, 
  SkipBack, 
  Volume2, 
  VolumeX, 
  Moon, 
  Lock, 
  EyeOff, 
  ShieldCheck, 
  QrCode, 
  Terminal, 
  ExternalLink,
  Wifi,
  Sliders,
  CheckCircle2,
  HardDrive
} from "lucide-react";
import Link from "next/link";

export default function LandingPage() {
  const [volume, setVolume] = useState(65);
  const [isPlaying, setIsPlaying] = useState(false);
  const [isMuted, setIsMuted] = useState(false);
  const [lastAction, setLastAction] = useState<string | null>(null);

  const handleAction = (label: string) => {
    setLastAction(label);
    setTimeout(() => setLastAction(null), 2000);
  };

  return (
    <div className="min-h-screen bg-[#09090b] text-[#f4f4f5] font-sans selection:bg-white/20">
      
      {/* Top Header */}
      <header className="border-b border-white/[0.08] bg-[#09090b]/80 backdrop-blur-md sticky top-0 z-50">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center border border-white/10">
              <Monitor className="w-4 h-4 text-white" />
            </div>
            <span className="font-semibold text-base tracking-tight text-white">PC Remote</span>
            <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-white/5 border border-white/10 text-[#a1a1aa]">v5.0.0</span>
          </div>

          <nav className="hidden sm:flex items-center space-x-6 text-sm text-[#a1a1aa]">
            <a href="#download" className="hover:text-white transition-colors">Download</a>
            <a href="#features" className="hover:text-white transition-colors">Fitur</a>
            <a href="#quickstart" className="hover:text-white transition-colors">Panduan</a>
            <a href="#specs" className="hover:text-white transition-colors">Spesifikasi</a>
          </nav>

          <Link 
            href="/connect" 
            className="text-xs font-semibold px-4 py-2 rounded-lg bg-white text-black hover:bg-white/90 active:scale-95 transition-all flex items-center space-x-1.5"
          >
            <span>Buka Web Remote</span>
            <ExternalLink className="w-3.5 h-3.5" />
          </Link>
        </div>
      </header>

      {/* Hero Section */}
      <main className="max-w-6xl mx-auto px-4 sm:px-6 pt-16 pb-24">
        
        {/* Headline & Description */}
        <div className="max-w-3xl">
          <div className="inline-flex items-center space-x-2 px-3 py-1 rounded-full bg-white/5 border border-white/10 text-xs text-[#a1a1aa] mb-6">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
            <span>Go Backend & Cloudflare Quick Tunnel</span>
          </div>

          <h1 className="text-4xl sm:text-6xl font-bold tracking-tight text-white leading-tight">
            Kontrol Windows PC Anda langsung dari genggaman.
          </h1>

          <p className="mt-5 text-lg text-[#a1a1aa] leading-relaxed">
            Aplikasi kendali jarak jauh yang ringan, aman, dan tanpa konfigurasi router. Mengatur volume sistem, kontrol media, matikan layar, hingga screen mirroring secara real-time melalui WiFi lokal maupun jaringan internet (WAN).
          </p>

          {/* Quick Action / Download Hub Direct */}
          <div id="download" className="mt-10 pt-8 border-t border-white/[0.08] grid grid-cols-1 sm:grid-cols-3 gap-4">
            
            {/* Windows Download */}
            <a 
              href="https://github.com/KristianEki11/PC-Remote/releases/latest" 
              target="_blank" 
              rel="noreferrer"
              className="p-5 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/10 hover:border-white/20 transition-all text-left group"
            >
              <div className="flex items-center justify-between mb-3">
                <Monitor className="w-5 h-5 text-blue-400" />
                <Download className="w-4 h-4 text-[#71717a] group-hover:text-white transition-colors" />
              </div>
              <div className="font-semibold text-white text-base">Windows Setup</div>
              <div className="text-xs text-[#a1a1aa] mt-1">PCRemoteSetup.exe • ~12 MB</div>
              <div className="text-[11px] font-mono text-[#71717a] mt-3">Windows 10 / 11 (64-bit)</div>
            </a>

            {/* Android Download */}
            <a 
              href="https://github.com/KristianEki11/PC-Remote/releases/latest" 
              target="_blank" 
              rel="noreferrer"
              className="p-5 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/10 hover:border-white/20 transition-all text-left group"
            >
              <div className="flex items-center justify-between mb-3">
                <Smartphone className="w-5 h-5 text-emerald-400" />
                <Download className="w-4 h-4 text-[#71717a] group-hover:text-white transition-colors" />
              </div>
              <div className="font-semibold text-white text-base">Android APK</div>
              <div className="text-xs text-[#a1a1aa] mt-1">PCRemoteApp.apk • Universal</div>
              <div className="text-[11px] font-mono text-[#71717a] mt-3">Android 8.0 atau lebih baru</div>
            </a>

            {/* iOS Web Remote */}
            <Link 
              href="/connect" 
              className="p-5 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/10 hover:border-white/20 transition-all text-left group"
            >
              <div className="flex items-center justify-between mb-3">
                <Globe className="w-5 h-5 text-purple-400" />
                <ExternalLink className="w-4 h-4 text-[#71717a] group-hover:text-white transition-colors" />
              </div>
              <div className="font-semibold text-white text-base">iPhone Web Remote</div>
              <div className="text-xs text-[#a1a1aa] mt-1">Langsung via Safari (PWA)</div>
              <div className="text-[11px] font-mono text-[#71717a] mt-3">remote.redlinevis.site/connect</div>
            </Link>

          </div>
        </div>

        {/* Authentic Remote Control Interface Preview */}
        <div className="mt-16 pt-12 border-t border-white/[0.08]">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
            <div>
              <h2 className="text-xl font-bold text-white tracking-tight">Antarmuka Pengontrol</h2>
              <p className="text-sm text-[#a1a1aa] mt-1">Tampilan dan kontrol yang sama persis seperti yang ada di aplikasi Anda.</p>
            </div>
            {lastAction && (
              <span className="text-xs font-mono px-3 py-1 rounded bg-white/10 text-white self-start sm:self-auto animate-pulse">
                Aksi: {lastAction}
              </span>
            )}
          </div>

          <div className="rounded-2xl border border-white/10 bg-[#121215] p-6 sm:p-8 max-w-3xl">
            
            {/* Top Status Bar */}
            <div className="flex items-center justify-between pb-6 border-b border-white/[0.06] text-xs">
              <div className="flex items-center space-x-2">
                <span className="w-2 h-2 rounded-full bg-emerald-400"></span>
                <span className="font-medium text-white">Status: Terhubung</span>
                <span className="text-[#71717a]">| 192.168.1.100:8000</span>
              </div>
              <span className="text-[#71717a] font-mono">Cloudflare Tunnel: Aktif</span>
            </div>

            {/* Volume Control */}
            <div className="py-6 border-b border-white/[0.06]">
              <div className="flex justify-between items-center text-xs mb-3 text-[#a1a1aa]">
                <span className="font-medium uppercase tracking-wider">Master Audio</span>
                <span className="font-mono text-white">{isMuted ? "MUTE" : `${volume}%`}</span>
              </div>
              <div className="flex items-center space-x-4">
                <button 
                  onClick={() => {
                    setIsMuted(!isMuted);
                    handleAction(isMuted ? "Audio Unmute" : "Audio Mute");
                  }}
                  className="text-[#a1a1aa] hover:text-white p-2 rounded-lg hover:bg-white/5"
                  title="Toggle Mute"
                >
                  {isMuted ? <VolumeX className="w-5 h-5 text-red-400" /> : <Volume2 className="w-5 h-5" />}
                </button>
                <input 
                  type="range"
                  min="0"
                  max="100"
                  value={isMuted ? 0 : volume}
                  disabled={isMuted}
                  onChange={(e) => {
                    setVolume(parseInt(e.target.value));
                    handleAction(`Volume: ${e.target.value}%`);
                  }}
                  className="w-full h-1.5 bg-white/10 rounded-full appearance-none cursor-pointer accent-white"
                />
              </div>
            </div>

            {/* Media Player Controls */}
            <div className="py-6 border-b border-white/[0.06] flex items-center justify-between">
              <div>
                <div className="text-xs font-mono text-[#71717a] uppercase tracking-wider">Windows Media Transport</div>
                <div className="text-sm font-semibold text-white mt-1">Spotify / YouTube / Media Player</div>
              </div>

              <div className="flex items-center space-x-2">
                <button 
                  onClick={() => handleAction("Media: Previous")}
                  className="p-3 rounded-lg border border-white/10 hover:bg-white/5 text-white active:scale-95 transition-all"
                  title="Previous"
                >
                  <SkipBack className="w-4 h-4 fill-current" />
                </button>
                <button 
                  onClick={() => {
                    setIsPlaying(!isPlaying);
                    handleAction(isPlaying ? "Media: Paused" : "Media: Play");
                  }}
                  className="p-3 rounded-lg bg-white text-black hover:bg-white/90 active:scale-95 transition-all"
                  title="Play/Pause"
                >
                  {isPlaying ? <Pause className="w-4 h-4 fill-current" /> : <Play className="w-4 h-4 fill-current" />}
                </button>
                <button 
                  onClick={() => handleAction("Media: Next")}
                  className="p-3 rounded-lg border border-white/10 hover:bg-white/5 text-white active:scale-95 transition-all"
                  title="Next"
                >
                  <SkipForward className="w-4 h-4 fill-current" />
                </button>
              </div>
            </div>

            {/* Power Actions */}
            <div className="pt-6">
              <div className="text-xs font-mono text-[#71717a] uppercase tracking-wider mb-4">Aksi Daya & Sistem</div>
              <div className="grid grid-cols-3 gap-3">
                <button 
                  onClick={() => handleAction("Matikan Layar (Display Off)")}
                  className="py-3 px-4 rounded-lg border border-white/10 hover:bg-white/5 text-xs font-medium text-white flex items-center justify-center space-x-2 active:scale-95 transition-all"
                >
                  <EyeOff className="w-4 h-4 text-[#a1a1aa]" />
                  <span>Matikan Layar</span>
                </button>
                <button 
                  onClick={() => handleAction("Kunci PC (LockWorkStation)")}
                  className="py-3 px-4 rounded-lg border border-white/10 hover:bg-white/5 text-xs font-medium text-white flex items-center justify-center space-x-2 active:scale-95 transition-all"
                >
                  <Lock className="w-4 h-4 text-[#a1a1aa]" />
                  <span>Kunci PC</span>
                </button>
                <button 
                  onClick={() => handleAction("Tidurkan PC (S3 Sleep)")}
                  className="py-3 px-4 rounded-lg border border-white/10 hover:bg-white/5 text-xs font-medium text-white flex items-center justify-center space-x-2 active:scale-95 transition-all"
                >
                  <Moon className="w-4 h-4 text-[#a1a1aa]" />
                  <span>Sleep</span>
                </button>
              </div>
            </div>

          </div>
        </div>

        {/* Feature Highlights Section */}
        <div id="features" className="mt-20 pt-16 border-t border-white/[0.08]">
          <h2 className="text-2xl font-bold text-white tracking-tight mb-2">Kemampuan Inti</h2>
          <p className="text-sm text-[#a1a1aa] mb-10 max-w-2xl">
            Arsitektur yang dibuat untuk keandalan maksimal tanpa bloatware pihak ketiga.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            
            <div className="p-6 rounded-xl border border-white/10 bg-white/[0.02]">
              <div className="flex items-center space-x-3 mb-3">
                <Globe className="w-5 h-5 text-blue-400" />
                <h3 className="font-semibold text-white text-base">Akses WAN Otomatis (Cloudflare Tunnel)</h3>
              </div>
              <p className="text-sm text-[#a1a1aa] leading-relaxed">
                Terkoneksi dari mana saja di luar rumah (4G/5G) tanpa perlu setting port-forwarding router atau DDNS. Server otomatis membuat tunnel aman saat dinyalakan.
              </p>
            </div>

            <div className="p-6 rounded-xl border border-white/10 bg-white/[0.02]">
              <div className="flex items-center space-x-3 mb-3">
                <Sliders className="w-5 h-5 text-emerald-400" />
                <h3 className="font-semibold text-white text-base">Windows Core Audio COM API</h3>
              </div>
              <p className="text-sm text-[#a1a1aa] leading-relaxed">
                Menggunakan COM Worker thread terisolasi di Go (`IMMDeviceEnumerator`). Mendukung kontrol volume master dan multi-channel SteelSeries Sonar.
              </p>
            </div>

            <div className="p-6 rounded-xl border border-white/10 bg-white/[0.02]">
              <div className="flex items-center space-x-3 mb-3">
                <Monitor className="w-5 h-5 text-purple-400" />
                <h3 className="font-semibold text-white text-base">Screen Mirroring H.264 (v5.0)</h3>
              </div>
              <p className="text-sm text-[#a1a1aa] leading-relaxed">
                Streaming tampilan monitor langsung ke HP via WebSocket dengan latensi ~80-150ms di jaringan lokal dan touch-to-click injection.
              </p>
            </div>

            <div className="p-6 rounded-xl border border-white/10 bg-white/[0.02]">
              <div className="flex items-center space-x-3 mb-3">
                <QrCode className="w-5 h-5 text-amber-400" />
                <h3 className="font-semibold text-white text-base">Pairing QR Code AES-256</h3>
              </div>
              <p className="text-sm text-[#a1a1aa] leading-relaxed">
                Cukup buka dashboard di PC dan scan QR Code menggunakan kamera HP. Token satu kali pakai dengan masa berlaku 5 menit mencegah akses tanpa izin.
              </p>
            </div>

          </div>
        </div>

        {/* 3-Step Quickstart */}
        <div id="quickstart" className="mt-20 pt-16 border-t border-white/[0.08]">
          <h2 className="text-2xl font-bold text-white tracking-tight mb-2">Cara Menghubungkan</h2>
          <p className="text-sm text-[#a1a1aa] mb-10">Tiga langkah cepat tanpa perlu pengaturan teknis rumit.</p>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
            <div className="p-6 rounded-xl border border-white/10 bg-white/[0.02]">
              <div className="text-xs font-mono font-bold text-white/50 mb-3">LANGKAH 01</div>
              <h3 className="font-semibold text-white mb-2">Jalankan Setup di PC</h3>
              <p className="text-xs text-[#a1a1aa] leading-relaxed">
                Unduh dan pasang `PCRemoteSetup.exe`. Tentukan 4-digit PIN saat diminta. Server akan otomatis aktif di latar belakang saat PC menyala.
              </p>
            </div>

            <div className="p-6 rounded-xl border border-white/10 bg-white/[0.02]">
              <div className="text-xs font-mono font-bold text-white/50 mb-3">LANGKAH 02</div>
              <h3 className="font-semibold text-white mb-2">Buka Aplikasi di HP</h3>
              <p className="text-xs text-[#a1a1aa] leading-relaxed">
                Untuk Android, install `PCRemoteApp.apk`. Untuk pengguna iPhone, buka menu Web Remote di `remote.redlinevis.site/connect`.
              </p>
            </div>

            <div className="p-6 rounded-xl border border-white/10 bg-white/[0.02]">
              <div className="text-xs font-mono font-bold text-white/50 mb-3">LANGKAH 03</div>
              <h3 className="font-semibold text-white mb-2">Scan QR & Kontrol</h3>
              <p className="text-xs text-[#a1a1aa] leading-relaxed">
                Tekan tombol Scan QR Code pada aplikasi dan arahkan kamera ke dashboard PC Anda, atau masukkan IP dan PIN secara manual. Selesai!
              </p>
            </div>
          </div>
        </div>

        {/* Specifications & License Table */}
        <div id="specs" className="mt-20 pt-16 border-t border-white/[0.08]">
          <h2 className="text-2xl font-bold text-white tracking-tight mb-6">Spesifikasi Teknis & Lisensi</h2>
          
          <div className="rounded-xl border border-white/10 overflow-hidden">
            <table className="w-full text-left text-sm">
              <tbody className="divide-y divide-white/[0.06]">
                <tr className="bg-white/[0.02]">
                  <td className="py-3 px-4 font-mono text-xs text-[#71717a] w-48">Backend Server</td>
                  <td className="py-3 px-4 text-white">Go (Golang 1.21+), Native Win32 / COM, Single Binary ~12 MB</td>
                </tr>
                <tr>
                  <td className="py-3 px-4 font-mono text-xs text-[#71717a]">Aplikasi Android</td>
                  <td className="py-3 px-4 text-white">Flutter 3.x, Dark Mode Native, Camera QR Scanner</td>
                </tr>
                <tr className="bg-white/[0.02]">
                  <td className="py-3 px-4 font-mono text-xs text-[#71717a]">Web Remote Client</td>
                  <td className="py-3 px-4 text-white">Next.js 16 (Static Export), iOS Safari PWA Compatible</td>
                </tr>
                <tr>
                  <td className="py-3 px-4 font-mono text-xs text-[#71717a]">Protokol Jaringan</td>
                  <td className="py-3 px-4 text-white">HTTPS REST API, WebSocket (H.264 Binary Screen Stream)</td>
                </tr>
                <tr className="bg-white/[0.02]">
                  <td className="py-3 px-4 font-mono text-xs text-[#71717a]">Keamanan</td>
                  <td className="py-3 px-4 text-white">Constant-Time PIN Compare, Token Volatile Memory, AES-256 Pairing</td>
                </tr>
                <tr>
                  <td className="py-3 px-4 font-mono text-xs text-[#71717a]">Hak Cipta & Lisensi</td>
                  <td className="py-3 px-4 text-[#a1a1aa]">
                    <span className="text-white font-medium">&copy; 2026 kidev.</span> Hak cipta dilindungi undang-undang. Gratis untuk penggunaan pribadi. Dilarang diperjualbelikan atau dimasukkan ke dalam produk/proyek lain tanpa izin.
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

      </main>

      {/* Footer */}
      <footer className="border-t border-white/[0.08] py-8 text-center text-xs text-[#71717a]">
        <div className="max-w-6xl mx-auto px-4 flex flex-col sm:flex-row items-center justify-between gap-4">
          <div>
            <span className="text-white font-medium">PC Remote</span> • remote.redlinevis.site
          </div>
          <div>
            &copy; 2026 kidev. All Rights Reserved.
          </div>
        </div>
      </footer>

    </div>
  );
}
