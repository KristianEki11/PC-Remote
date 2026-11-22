"use client";

import { useState } from "react";
import {
  Monitor,
  Smartphone,
  Globe,
  Download,
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
  BellRing,
  Sliders,
} from "lucide-react";

const RELEASES_URL =
  "https://github.com/KristianEki11/PC-Remote/releases/latest";

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
            <span className="font-semibold text-base tracking-tight text-white">
              PC Remote
            </span>
            <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-white/5 border border-white/10 text-[#a1a1aa]">
              v5.0.0
            </span>
          </div>

          <nav className="hidden sm:flex items-center space-x-6 text-sm text-[#a1a1aa]">
            <a href="#download" className="hover:text-white transition-colors">
              Download
            </a>
            <a href="#features" className="hover:text-white transition-colors">
              Features
            </a>
            <a
              href="#quickstart"
              className="hover:text-white transition-colors"
            >
              Quickstart
            </a>
            <a href="#specs" className="hover:text-white transition-colors">
              Specs
            </a>
          </nav>

          <a
            href="#download"
            className="text-xs font-semibold px-4 py-2 rounded-lg bg-white text-black hover:bg-white/90 active:scale-95 transition-all flex items-center space-x-1.5"
          >
            <Download className="w-3.5 h-3.5" />
            <span>Download v5.0</span>
          </a>
        </div>
      </header>

      {/* Hero Section */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 pt-16 lg:pt-24 pb-24">
        <div className="flex flex-col lg:flex-row items-center lg:items-start gap-12 lg:gap-20">
          {/* Left Column: Headline & Description & Downloads */}
          <div className="flex-1 w-full max-w-2xl lg:pt-8">
            <div className="inline-flex items-center space-x-2 text-xs font-mono px-3 py-1 rounded-full bg-white/5 border border-white/10 text-[#a1a1aa] mb-6">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
              <span>Rust Native Win32 • Tailscale • 24/7 Android Companion</span>
            </div>

            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold tracking-tight text-white leading-[1.1]">
              Control your Windows PC from the palm of your hand.
            </h1>

            <p className="mt-6 text-lg text-[#a1a1aa] leading-relaxed">
              Rebuilt in Rust with a native Win32 desktop app, per-user
              Tailscale zero re-pairing, and a 24/7 Android companion
              notification. Manage system audio, media playback, hybrid OLED
              brightness, and power states over Wi-Fi or anywhere on your
              tailnet.
            </p>

            {/* Quick Action / Download Hub Direct */}
            <div
              id="download"
              className="mt-10 pt-8 border-t border-white/[0.08] grid grid-cols-1 sm:grid-cols-2 gap-4"
            >
              {/* Windows Download */}
              <a
                href={RELEASES_URL}
                target="_blank"
                rel="noreferrer"
                className="p-5 rounded-xl bg-white/[0.02] hover:bg-white/[0.06] border border-white/10 hover:border-white/20 transition-all text-left group flex flex-col h-full"
              >
                <div className="flex items-center justify-between mb-3">
                  <Monitor className="w-5 h-5 text-blue-400" />
                  <Download className="w-4 h-4 text-[#71717a] group-hover:text-white transition-colors" />
                </div>
                <div className="font-semibold text-white text-base">
                  Windows Setup
                </div>
                <div className="text-xs text-[#a1a1aa] mt-1">
                  PCRemoteRustSetup.exe • Native Win32
                </div>
                <div className="text-[11px] font-mono text-[#71717a] mt-auto pt-4">
                  Windows 10 / 11 (64-bit)
                </div>
              </a>

              {/* Android Download */}
              <a
                href={RELEASES_URL}
                target="_blank"
                rel="noreferrer"
                className="p-5 rounded-xl bg-white/[0.02] hover:bg-white/[0.06] border border-white/10 hover:border-white/20 transition-all text-left group flex flex-col h-full"
              >
                <div className="flex items-center justify-between mb-3">
                  <Smartphone className="w-5 h-5 text-emerald-400" />
                  <Download className="w-4 h-4 text-[#71717a] group-hover:text-white transition-colors" />
                </div>
                <div className="font-semibold text-white text-base">
                  Android Companion APK
                </div>
                <div className="text-xs text-[#a1a1aa] mt-1">
                  PCRemoteApp.apk • 24/7 Service
                </div>
                <div className="text-[11px] font-mono text-[#71717a] mt-auto pt-4">
                  Android 8.0+ (TLS Pinned)
                </div>
              </a>

              {/* Direct Download Tailscale HP (Android) */}
              <a
                href="https://tailscale.com/download/android"
                target="_blank"
                rel="noreferrer"
                className="sm:col-span-2 p-5 rounded-xl bg-white/[0.02] hover:bg-white/[0.06] border border-white/10 hover:border-white/20 transition-all text-left group flex items-center justify-between"
              >
                <div className="flex items-center space-x-4">
                  <div className="w-10 h-10 rounded-full bg-blue-500/10 flex items-center justify-center">
                    <Globe className="w-5 h-5 text-blue-400" />
                  </div>
                  <div>
                    <div className="font-semibold text-white text-base">
                      Download Tailscale HP (Android)
                    </div>
                    <div className="text-xs text-[#a1a1aa] mt-1">
                      Direct official Tailscale Android APK / Play Store for away-from-home access
                    </div>
                  </div>
                </div>
                <Download className="w-4 h-4 text-[#71717a] group-hover:text-white transition-colors shrink-0 ml-4" />
              </a>
            </div>
          </div>

          {/* Right Column: Authentic Remote Control Interface Preview */}
          <div className="flex-1 w-full max-w-[400px] lg:max-w-md relative">
            {/* Ambient Glow */}
            <div className="absolute -inset-1 bg-gradient-to-tr from-blue-500/10 via-emerald-500/10 to-purple-500/10 rounded-3xl blur-2xl opacity-50"></div>

            <div className="relative rounded-3xl border border-white/10 bg-[#09090b] overflow-hidden shadow-2xl">
              {/* App Header */}
              <div className="bg-[#121215] px-6 py-5 border-b border-white/[0.06] flex items-center justify-between">
                <div>
                  <div className="text-sm font-semibold text-white flex items-center space-x-2">
                    <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
                    <span>Companion Preview</span>
                  </div>
                  <div className="text-[10px] font-mono text-[#71717a] mt-1">
                    100.x.y.z:8443 • TLS Pinned
                  </div>
                </div>
                {lastAction && (
                  <span className="text-[10px] font-mono px-2 py-1 rounded bg-white/10 text-white animate-pulse">
                    Action: {lastAction}
                  </span>
                )}
              </div>

              <div className="p-6">
                {/* Volume Control */}
                <div className="pb-6 border-b border-white/[0.06]">
                  <div className="flex justify-between items-center text-xs mb-4 text-[#a1a1aa]">
                    <span className="font-medium uppercase tracking-wider">
                      Master Audio
                    </span>
                    <span className="font-mono text-white">
                      {isMuted ? "MUTE" : `${volume}%`}
                    </span>
                  </div>
                  <div className="flex items-center space-x-4">
                    <button
                      type="button"
                      onClick={() => {
                        setIsMuted(!isMuted);
                        handleAction(isMuted ? "Audio Unmute" : "Audio Mute");
                      }}
                      className="text-[#a1a1aa] hover:text-white p-2 rounded-lg hover:bg-white/5 transition-colors"
                      title="Toggle Mute"
                    >
                      {isMuted ? (
                        <VolumeX className="w-5 h-5 text-red-400" />
                      ) : (
                        <Volume2 className="w-5 h-5" />
                      )}
                    </button>
                    <input
                      type="range"
                      min="0"
                      max="100"
                      value={isMuted ? 0 : volume}
                      disabled={isMuted}
                      onChange={(e) => {
                        setVolume(parseInt(e.target.value, 10));
                        handleAction(`Volume: ${e.target.value}%`);
                      }}
                      className="w-full h-1.5 bg-white/10 rounded-full appearance-none cursor-pointer accent-white"
                    />
                  </div>
                </div>

                {/* Media Player Controls */}
                <div className="py-6 border-b border-white/[0.06]">
                  <div className="text-[10px] font-mono text-[#71717a] uppercase tracking-wider mb-1">
                    Windows Media Transport
                  </div>
                  <div className="text-sm font-semibold text-white mb-4">
                    Spotify / YouTube / Media
                  </div>

                  <div className="flex items-center justify-between gap-3">
                    <button
                      type="button"
                      onClick={() => handleAction("Media: Previous")}
                      className="flex-1 py-3 flex justify-center rounded-xl border border-white/10 bg-white/[0.02] hover:bg-white/[0.06] text-white active:scale-95 transition-all"
                    >
                      <SkipBack className="w-5 h-5 fill-current" />
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        setIsPlaying(!isPlaying);
                        handleAction(
                          isPlaying ? "Media: Paused" : "Media: Play"
                        );
                      }}
                      className="flex-1 py-3 flex justify-center rounded-xl bg-white text-black hover:bg-white/90 active:scale-95 transition-all shadow-[0_0_20px_rgba(255,255,255,0.1)]"
                    >
                      {isPlaying ? (
                        <Pause className="w-5 h-5 fill-current" />
                      ) : (
                        <Play className="w-5 h-5 fill-current" />
                      )}
                    </button>
                    <button
                      type="button"
                      onClick={() => handleAction("Media: Next")}
                      className="flex-1 py-3 flex justify-center rounded-xl border border-white/10 bg-white/[0.02] hover:bg-white/[0.06] text-white active:scale-95 transition-all"
                    >
                      <SkipForward className="w-5 h-5 fill-current" />
                    </button>
                  </div>
                </div>

                {/* Power Actions */}
                <div className="pt-6">
                  <div className="text-[10px] font-mono text-[#71717a] uppercase tracking-wider mb-4">
                    Power &amp; System Actions
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <button
                      type="button"
                      onClick={() => handleAction("Turn Off Display")}
                      className="py-3 px-3 rounded-xl border border-white/10 bg-white/[0.02] hover:bg-white/[0.06] text-xs font-medium text-white flex flex-col items-center justify-center space-y-2 active:scale-95 transition-all"
                    >
                      <EyeOff className="w-5 h-5 text-[#a1a1aa]" />
                      <span>Turn Off Display</span>
                    </button>
                    <button
                      type="button"
                      onClick={() => handleAction("Lock PC")}
                      className="py-3 px-3 rounded-xl border border-white/10 bg-white/[0.02] hover:bg-white/[0.06] text-xs font-medium text-white flex flex-col items-center justify-center space-y-2 active:scale-95 transition-all"
                    >
                      <Lock className="w-5 h-5 text-[#a1a1aa]" />
                      <span>Lock PC</span>
                    </button>
                    <button
                      type="button"
                      onClick={() => handleAction("S3 Sleep")}
                      className="col-span-2 py-3 px-3 rounded-xl border border-white/10 bg-white/[0.02] hover:bg-white/[0.06] text-xs font-medium text-white flex items-center justify-center space-x-2 active:scale-95 transition-all"
                    >
                      <Moon className="w-4 h-4 text-[#a1a1aa]" />
                      <span>Sleep Mode</span>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Feature Highlights Section */}
        <div id="features" className="mt-20 pt-16 border-t border-white/[0.08]">
          <h2 className="text-2xl font-bold text-white tracking-tight mb-2">
            Core Capabilities in v5.0
          </h2>
          <p className="text-sm text-[#a1a1aa] mb-10 max-w-2xl">
            Native performance on Windows and Android without third-party relay
            servers or browser dashboards.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="p-6 rounded-xl border border-white/10 bg-white/[0.02]">
              <div className="flex items-center space-x-3 mb-3">
                <Monitor className="w-5 h-5 text-blue-400" />
                <h3 className="font-semibold text-white text-base">
                  Rust Native Win32 GUI &amp; Tray
                </h3>
              </div>
              <p className="text-sm text-[#a1a1aa] leading-relaxed">
                Single lightweight Rust binary with a native Win32/GDI dark
                window, live QR pairing, 3-step Tailscale onboarding wizard, and
                System Tray close-to-background.
              </p>
            </div>

            <div className="p-6 rounded-xl border border-white/10 bg-white/[0.02]">
              <div className="flex items-center space-x-3 mb-3">
                <Globe className="w-5 h-5 text-emerald-400" />
                <h3 className="font-semibold text-white text-base">
                  Per-User Tailscale Zero Re-Pairing
                </h3>
              </div>
              <p className="text-sm text-[#a1a1aa] leading-relaxed">
                Connect over your own private tailnet. Fast boot polling,
                automatic endpoint sync, and parallel LAN-to-Tailscale failover
                eliminate QR re-scanning after sleep or network switches.
              </p>
            </div>

            <div className="p-6 rounded-xl border border-white/10 bg-white/[0.02]">
              <div className="flex items-center space-x-3 mb-3">
                <BellRing className="w-5 h-5 text-purple-400" />
                <h3 className="font-semibold text-white text-base">
                  24/7 Android Companion Media Notification
                </h3>
              </div>
              <p className="text-sm text-[#a1a1aa] leading-relaxed">
                A native Android Foreground Service keeps connection status live
                and puts interactive Prev, Play/Pause, and Next controls right
                in your notification shade and lock screen.
              </p>
            </div>

            <div className="p-6 rounded-xl border border-white/10 bg-white/[0.02]">
              <div className="flex items-center space-x-3 mb-3">
                <ShieldCheck className="w-5 h-5 text-amber-400" />
                <h3 className="font-semibold text-white text-base">
                  COM / WinRT Control &amp; DPAPI Security
                </h3>
              </div>
              <p className="text-sm text-[#a1a1aa] leading-relaxed">
                Direct Windows COM audio (Master, devices, and SteelSeries
                Sonar), WinRT media transport, hybrid WMI + OLED brightness, and
                DPAPI-protected SHA-256 TLS certificate pinning.
              </p>
            </div>
          </div>
        </div>

        {/* 3-Step Quickstart */}
        <div
          id="quickstart"
          className="mt-20 pt-16 border-t border-white/[0.08]"
        >
          <h2 className="text-2xl font-bold text-white tracking-tight mb-2">
            How to Connect
          </h2>
          <p className="text-sm text-[#a1a1aa] mb-10">
            Three quick steps with zero terminal commands required.
          </p>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
            <div className="p-6 rounded-xl border border-white/10 bg-white/[0.02]">
              <div className="text-xs font-mono font-bold text-white/50 mb-3">
                STEP 01
              </div>
              <h3 className="font-semibold text-white mb-2">
                Install on Windows
              </h3>
              <p className="text-xs text-[#a1a1aa] leading-relaxed">
                Run `PCRemoteRustSetup.exe` and open PC Remote. The native
                window displays your pairing QR code and 3-step Tailscale setup
                wizard.
              </p>
            </div>

            <div className="p-6 rounded-xl border border-white/10 bg-white/[0.02]">
              <div className="text-xs font-mono font-bold text-white/50 mb-3">
                STEP 02
              </div>
              <h3 className="font-semibold text-white mb-2">
                Install Android Companion
              </h3>
              <p className="text-xs text-[#a1a1aa] leading-relaxed">
                Install `PCRemoteApp.apk` on your Android phone. For outside-home
                access, sign in to Tailscale on both your PC and phone.
              </p>
            </div>

            <div className="p-6 rounded-xl border border-white/10 bg-white/[0.02]">
              <div className="text-xs font-mono font-bold text-white/50 mb-3">
                STEP 03
              </div>
              <h3 className="font-semibold text-white mb-2">
                Pair Once &amp; Roam Freely
              </h3>
              <p className="text-xs text-[#a1a1aa] leading-relaxed">
                Scan the QR code once to pin your PC&apos;s TLS certificate.
                Smart roaming and the 24/7 companion notification handle every
                reconnect automatically.
              </p>
            </div>
          </div>
        </div>

        {/* Specifications & License Table */}
        <div id="specs" className="mt-20 pt-16 border-t border-white/[0.08]">
          <h2 className="text-2xl font-bold text-white tracking-tight mb-6 flex items-center space-x-2">
            <Sliders className="w-5 h-5 text-[#a1a1aa]" />
            <span>Technical Specs &amp; License</span>
          </h2>

          <div className="rounded-xl border border-white/10 overflow-hidden">
            <table className="w-full text-left text-sm">
              <tbody className="divide-y divide-white/[0.06]">
                <tr className="bg-white/[0.02]">
                  <td className="py-3 px-4 font-mono text-xs text-[#71717a] w-48">
                    PC Server &amp; GUI
                  </td>
                  <td className="py-3 px-4 text-white">
                    Rust 2024, Axum HTTPS (:8443), Native Win32 / GDI / COM /
                    WinRT
                  </td>
                </tr>
                <tr>
                  <td className="py-3 px-4 font-mono text-xs text-[#71717a]">
                    Android Companion
                  </td>
                  <td className="py-3 px-4 text-white">
                    Flutter 3.x + Kotlin Foreground Service (24/7 Media
                    Notification)
                  </td>
                </tr>
                <tr className="bg-white/[0.02]">
                  <td className="py-3 px-4 font-mono text-xs text-[#71717a]">
                    Network Transport
                  </td>
                  <td className="py-3 px-4 text-white">
                    Direct LAN &amp; Per-User Tailscale (IPv4 + MagicDNS),
                    WebSocket `/events`
                  </td>
                </tr>
                <tr>
                  <td className="py-3 px-4 font-mono text-xs text-[#71717a]">
                    Security
                  </td>
                  <td className="py-3 px-4 text-white">
                    Windows DPAPI Identity, SHA-256 TLS Pinning, One-Time QR
                    Pairing
                  </td>
                </tr>
                <tr className="bg-white/[0.02]">
                  <td className="py-3 px-4 font-mono text-xs text-[#71717a]">
                    Copyright &amp; License
                  </td>
                  <td className="py-3 px-4 text-[#a1a1aa]">
                    <span className="text-white font-medium">
                      &copy; 2026 kidev.
                    </span>{" "}
                    All rights reserved. Free for personal use. Commercial sale
                    or integration into other products/projects without
                    permission is strictly prohibited.
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
            <span className="text-white font-medium">PC Remote v5.0</span> •
            remote.redlinevis.site
          </div>
          <div>&copy; 2026 kidev. All Rights Reserved.</div>
        </div>
      </footer>
    </div>
  );
}
