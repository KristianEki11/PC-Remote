"use client";

import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { 
  MonitorSmartphone, 
  Shield, 
  Zap, 
  Globe, 
  Download, 
  ChevronRight, 
  Layers, 
  Play, 
  Pause, 
  SkipForward, 
  SkipBack, 
  Volume2, 
  VolumeX, 
  Moon, 
  Lock, 
  EyeOff, 
  Radio, 
  Sparkles, 
  CheckCircle2, 
  QrCode, 
  Cpu, 
  Terminal, 
  ExternalLink,
  Wifi,
  Sliders,
  Maximize2
} from "lucide-react";
import Link from "next/link";

export default function LandingPage() {
  // Interactive Hero Mockup States
  const [mockVolume, setMockVolume] = useState(68);
  const [mockPlaying, setMockPlaying] = useState(true);
  const [mockTrackIndex, setMockTrackIndex] = useState(0);
  const [actionNotice, setActionNotice] = useState<string | null>(null);

  const mockTracks = [
    { title: "Starboy", artist: "The Weeknd, Daft Punk", album: "Starboy" },
    { title: "Cornfield Chase", artist: "Hans Zimmer", album: "Interstellar OST" },
    { title: "Blinding Lights", artist: "The Weeknd", album: "After Hours" }
  ];

  const triggerAction = (msg: string) => {
    setActionNotice(msg);
    setTimeout(() => setActionNotice(null), 2200);
  };

  return (
    <div className="min-h-screen flex flex-col bg-[#050508] text-[#f5f5f7] overflow-x-hidden font-sans selection:bg-blue-500/30 selection:text-white">
      
      {/* Dynamic Ambient Background Glows */}
      <div className="fixed inset-0 pointer-events-none z-0 overflow-hidden">
        <div className="absolute top-[-10%] left-[20%] w-[650px] h-[650px] rounded-full bg-gradient-to-tr from-blue-600/20 to-indigo-600/10 blur-[130px] opacity-70" />
        <div className="absolute top-[15%] right-[10%] w-[550px] h-[550px] rounded-full bg-gradient-to-bl from-purple-600/15 to-pink-600/10 blur-[140px] opacity-60" />
        <div className="absolute top-[60%] left-[-5%] w-[600px] h-[600px] rounded-full bg-gradient-to-r from-cyan-600/15 to-blue-600/10 blur-[150px] opacity-50" />
      </div>

      {/* Navigation */}
      <header className="fixed top-0 w-full z-50 glass border-b border-white/[0.08] backdrop-blur-2xl">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <Link href="/" className="flex items-center space-x-3 group">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center shadow-lg shadow-blue-500/25 group-hover:scale-105 transition-transform duration-300">
                <MonitorSmartphone className="w-5 h-5 text-white" />
              </div>
              <div className="flex items-center space-x-2">
                <span className="font-semibold text-lg tracking-tight text-white">PC Remote</span>
                <span className="text-[11px] font-mono tracking-wider px-2 py-0.5 rounded-full bg-white/[0.08] border border-white/10 text-blue-400 font-medium">v5.0 PRO</span>
              </div>
            </Link>

            <nav className="hidden md:flex items-center space-x-8 text-sm font-medium text-[#a1a1a6]">
              <a href="#features" className="hover:text-white transition-colors duration-200">Features</a>
              <a href="#screen-mirror" className="hover:text-white transition-colors duration-200">Screen Mirror</a>
              <a href="#download" className="hover:text-white transition-colors duration-200">Download</a>
              <a href="#quickstart" className="hover:text-white transition-colors duration-200">Quickstart</a>
              <a href="#documentation" className="hover:text-white transition-colors duration-200">Docs</a>
            </nav>

            <div className="flex items-center space-x-3">
              <Link 
                href="/connect" 
                className="relative group overflow-hidden px-4 py-2 rounded-full text-xs font-semibold bg-white text-black hover:bg-white/90 active:scale-95 transition-all shadow-md flex items-center space-x-1.5"
              >
                <span>Launch Web Remote</span>
                <ExternalLink className="w-3.5 h-3.5" />
              </Link>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-grow pt-28 relative z-10">

        {/* Hero Section */}
        <section className="pt-12 pb-24 px-4 sm:px-6 lg:px-8 text-center max-w-7xl mx-auto">
          
          {/* Release Badge */}
          <motion.div 
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            className="inline-flex items-center space-x-2.5 px-4 py-1.5 rounded-full bg-white/[0.05] border border-white/[0.1] backdrop-blur-xl mb-8 hover:border-white/20 transition-colors"
          >
            <span className="flex h-2 w-2 relative">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span>
            </span>
            <span className="text-xs font-medium tracking-wide text-white/90">
              Next-Gen v5.0.0 Release — Native Go & Low-Latency Mirroring
            </span>
            <ChevronRight className="w-3.5 h-3.5 text-white/50" />
          </motion.div>

          {/* Hero Typography */}
          <motion.h1 
            className="text-5xl sm:text-6xl md:text-8xl font-black tracking-[-0.04em] text-white leading-[1.05] max-w-5xl mx-auto"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
          >
            Your entire PC. <br />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-blue-400 via-indigo-400 to-purple-500">
              In the palm of your hand.
            </span>
          </motion.h1>

          <motion.p 
            className="mt-8 text-lg sm:text-xl md:text-2xl text-[#86868b] max-w-3xl mx-auto font-normal leading-relaxed tracking-tight"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.1, ease: [0.16, 1, 0.3, 1] }}
          >
            The ultimate Windows remote control companion. Live desktop streaming, granular Core Audio mixing, media synchronization, and secure power management — across your local WiFi or through zero-config Cloudflare tunnels from anywhere in the world.
          </motion.p>

          {/* Action Buttons */}
          <motion.div 
            className="mt-10 flex flex-col sm:flex-row justify-center items-center gap-4 max-w-md mx-auto sm:max-w-none"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.2, ease: [0.16, 1, 0.3, 1] }}
          >
            <a 
              href="#download" 
              className="w-full sm:w-auto px-8 py-4 rounded-2xl bg-blue-600 hover:bg-blue-500 text-white font-semibold text-base shadow-xl shadow-blue-600/30 active:scale-95 transition-all flex items-center justify-center space-x-2.5 group"
            >
              <Download className="w-5 h-5 group-hover:-translate-y-0.5 transition-transform" />
              <span>Download Windows Setup</span>
            </a>

            <Link 
              href="/connect" 
              className="w-full sm:w-auto px-8 py-4 rounded-2xl bg-white/[0.08] hover:bg-white/[0.14] border border-white/10 text-white font-semibold text-base backdrop-blur-xl active:scale-95 transition-all flex items-center justify-center space-x-2"
            >
              <Radio className="w-5 h-5 text-blue-400" />
              <span>Open Web Remote (iOS)</span>
              <ChevronRight className="w-4 h-4 text-white/50" />
            </Link>
          </motion.div>

          {/* Trust Specs Strip */}
          <motion.div 
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.4, duration: 1 }}
            className="mt-14 pt-8 border-t border-white/[0.06] grid grid-cols-2 sm:grid-cols-4 gap-6 max-w-4xl mx-auto text-left"
          >
            <div className="flex items-center space-x-3">
              <Zap className="w-5 h-5 text-amber-400 shrink-0" />
              <div>
                <div className="text-sm font-semibold text-white">&lt;100ms Latency</div>
                <div className="text-xs text-[#86868b]">WebSocket + H.264</div>
              </div>
            </div>
            <div className="flex items-center space-x-3">
              <Globe className="w-5 h-5 text-blue-400 shrink-0" />
              <div>
                <div className="text-sm font-semibold text-white">Zero Port-Forward</div>
                <div className="text-xs text-[#86868b]">Cloudflare Quick Tunnel</div>
              </div>
            </div>
            <div className="flex items-center space-x-3">
              <Shield className="w-5 h-5 text-emerald-400 shrink-0" />
              <div>
                <div className="text-sm font-semibold text-white">AES-256 Crypto</div>
                <div className="text-xs text-[#86868b]">QR One-Time Tokens</div>
              </div>
            </div>
            <div className="flex items-center space-x-3">
              <Cpu className="w-5 h-5 text-purple-400 shrink-0" />
              <div>
                <div className="text-sm font-semibold text-white">10MB Go Binary</div>
                <div className="text-xs text-[#86868b]">0.1% CPU Idle Usage</div>
              </div>
            </div>
          </motion.div>

          {/* Interactive Floating Hardware/Remote Device Showcase */}
          <motion.div 
            className="mt-16 relative max-w-4xl mx-auto"
            initial={{ opacity: 0, y: 40 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 1, delay: 0.3, ease: [0.16, 1, 0.3, 1] }}
          >
            {/* Specular Ambient Glow Frame */}
            <div className="absolute -inset-1.5 bg-gradient-to-r from-blue-500/30 via-purple-500/20 to-cyan-500/30 rounded-[2.5rem] blur-2xl opacity-60" />

            <div className="relative rounded-[2.2rem] bg-[#0c0d12]/90 border border-white/[0.12] p-6 sm:p-8 backdrop-blur-2xl shadow-2xl overflow-hidden text-left">
              
              {/* Window Header */}
              <div className="flex flex-wrap items-center justify-between gap-4 pb-6 border-b border-white/[0.08]">
                <div className="flex items-center space-x-3">
                  <div className="flex space-x-2">
                    <div className="w-3 h-3 rounded-full bg-red-500/80" />
                    <div className="w-3 h-3 rounded-full bg-yellow-500/80" />
                    <div className="w-3 h-3 rounded-full bg-green-500/80" />
                  </div>
                  <span className="text-xs font-mono text-white/60 ml-2">Living Room PC • Windows 11 Enterprise</span>
                </div>

                <div className="flex items-center space-x-3">
                  <div className="flex items-center space-x-2 px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-medium">
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
                    <span>Tunnel Connected (28ms)</span>
                  </div>
                  <span className="hidden sm:inline text-xs font-mono text-white/40">LAN: 192.168.1.100:8000</span>
                </div>
              </div>

              {/* Toast Feedback */}
              <AnimatePresence>
                {actionNotice && (
                  <motion.div 
                    initial={{ opacity: 0, y: -10 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -10 }}
                    className="absolute top-20 right-8 z-30 px-4 py-2 rounded-xl bg-blue-600/90 text-white text-xs font-semibold shadow-lg backdrop-blur-lg flex items-center space-x-2"
                  >
                    <CheckCircle2 className="w-4 h-4 text-white" />
                    <span>{actionNotice}</span>
                  </motion.div>
                )}
              </AnimatePresence>

              {/* Interactive Widget Deck */}
              <div className="mt-8 grid grid-cols-1 lg:grid-cols-12 gap-6">
                
                {/* Media Player Card (Col 7) */}
                <div className="lg:col-span-7 rounded-2xl bg-white/[0.03] border border-white/[0.08] p-5 sm:p-6 flex flex-col justify-between">
                  <div className="flex justify-between items-start mb-4">
                    <div>
                      <span className="text-[11px] font-mono uppercase tracking-wider text-blue-400 font-semibold">System Media Transport Controls</span>
                      <h4 className="text-xl font-bold text-white mt-1">{mockTracks[mockTrackIndex].title}</h4>
                      <p className="text-sm text-[#86868b]">{mockTracks[mockTrackIndex].artist}</p>
                    </div>
                    <div className="flex items-center space-x-1.5 h-6">
                      {[40, 70, 90, 60, 85, 45].map((h, i) => (
                        <span 
                          key={i} 
                          className="w-1 bg-blue-400 rounded-full transition-all duration-300"
                          style={{ 
                            height: mockPlaying ? `${h}%` : '20%',
                            opacity: mockPlaying ? 0.9 : 0.3
                          }} 
                        />
                      ))}
                    </div>
                  </div>

                  {/* Playback Progress */}
                  <div className="w-full bg-white/10 h-1.5 rounded-full overflow-hidden my-4">
                    <div className="bg-gradient-to-r from-blue-500 to-indigo-500 h-full w-[54%] rounded-full"></div>
                  </div>

                  {/* Interactive Controls */}
                  <div className="flex items-center justify-between pt-2">
                    <button 
                      onClick={() => {
                        setMockTrackIndex((prev) => (prev > 0 ? prev - 1 : mockTracks.length - 1));
                        triggerAction("Media: Previous Track");
                      }}
                      className="p-3 rounded-full hover:bg-white/10 active:scale-90 text-white/80 hover:text-white transition-all"
                      title="Previous Track"
                    >
                      <SkipBack className="w-5 h-5 fill-current" />
                    </button>

                    <button 
                      onClick={() => {
                        setMockPlaying(!mockPlaying);
                        triggerAction(mockPlaying ? "Media: Paused" : "Media: Playing");
                      }}
                      className="w-14 h-14 rounded-full bg-white text-black hover:scale-105 active:scale-95 flex items-center justify-center shadow-lg shadow-white/15 transition-all"
                      title="Play/Pause"
                    >
                      {mockPlaying ? <Pause className="w-6 h-6 fill-current" /> : <Play className="w-6 h-6 fill-current ml-0.5" />}
                    </button>

                    <button 
                      onClick={() => {
                        setMockTrackIndex((prev) => (prev + 1) % mockTracks.length);
                        triggerAction("Media: Next Track");
                      }}
                      className="p-3 rounded-full hover:bg-white/10 active:scale-90 text-white/80 hover:text-white transition-all"
                      title="Next Track"
                    >
                      <SkipForward className="w-5 h-5 fill-current" />
                    </button>
                  </div>
                </div>

                {/* System Volume & Quick Actions (Col 5) */}
                <div className="lg:col-span-5 flex flex-col justify-between space-y-4">
                  
                  {/* Master Volume Slider Widget */}
                  <div className="rounded-2xl bg-white/[0.03] border border-white/[0.08] p-5">
                    <div className="flex justify-between items-center mb-3">
                      <span className="text-xs font-semibold uppercase tracking-wider text-white/70">Master Volume</span>
                      <span className="text-sm font-mono text-blue-400 font-bold">{mockVolume}%</span>
                    </div>

                    <div className="flex items-center space-x-3">
                      <button 
                        onClick={() => {
                          setMockVolume(mockVolume === 0 ? 50 : 0);
                          triggerAction(mockVolume === 0 ? "Audio: Unmuted" : "Audio: Muted");
                        }}
                        className="text-white/60 hover:text-white"
                      >
                        {mockVolume === 0 ? <VolumeX className="w-5 h-5 text-red-400" /> : <Volume2 className="w-5 h-5" />}
                      </button>
                      <input 
                        type="range"
                        min="0"
                        max="100"
                        value={mockVolume}
                        onChange={(e) => setMockVolume(parseInt(e.target.value))}
                        className="w-full h-2 bg-white/10 rounded-full appearance-none cursor-pointer accent-blue-500"
                      />
                    </div>
                  </div>

                  {/* Power & Security Control Center */}
                  <div className="grid grid-cols-3 gap-2.5">
                    <button 
                      onClick={() => triggerAction("System: Monitor Off (Backlight Suspended)")}
                      className="p-3.5 rounded-xl bg-white/[0.03] hover:bg-white/[0.08] border border-white/[0.08] active:scale-95 transition-all flex flex-col items-center justify-center text-center group"
                    >
                      <EyeOff className="w-5 h-5 text-indigo-400 mb-1.5 group-hover:scale-110 transition-transform" />
                      <span className="text-[11px] font-medium text-white/80">Display Off</span>
                    </button>

                    <button 
                      onClick={() => triggerAction("System: Lock WorkStation Triggered")}
                      className="p-3.5 rounded-xl bg-white/[0.03] hover:bg-white/[0.08] border border-white/[0.08] active:scale-95 transition-all flex flex-col items-center justify-center text-center group"
                    >
                      <Lock className="w-5 h-5 text-amber-400 mb-1.5 group-hover:scale-110 transition-transform" />
                      <span className="text-[11px] font-medium text-white/80">Lock PC</span>
                    </button>

                    <button 
                      onClick={() => triggerAction("System: Suspend to S3 Sleep")}
                      className="p-3.5 rounded-xl bg-white/[0.03] hover:bg-white/[0.08] border border-white/[0.08] active:scale-95 transition-all flex flex-col items-center justify-center text-center group"
                    >
                      <Moon className="w-5 h-5 text-blue-400 mb-1.5 group-hover:scale-110 transition-transform" />
                      <span className="text-[11px] font-medium text-white/80">Sleep</span>
                    </button>
                  </div>

                  {/* Mirror Stream Live Pill */}
                  <div className="flex items-center justify-between p-3 rounded-xl bg-purple-500/10 border border-purple-500/20 text-purple-300 text-xs">
                    <div className="flex items-center space-x-2">
                      <Maximize2 className="w-4 h-4 text-purple-400" />
                      <span className="font-semibold">Screen Mirror Active</span>
                    </div>
                    <span className="font-mono text-[10px] px-2 py-0.5 rounded bg-purple-500/20">60 FPS • H.264</span>
                  </div>

                </div>

              </div>

              {/* Bottom Interactive Prompt */}
              <div className="mt-6 pt-4 border-t border-white/[0.06] flex items-center justify-between text-xs text-[#86868b]">
                <div className="flex items-center space-x-2">
                  <Sparkles className="w-3.5 h-3.5 text-blue-400" />
                  <span>Try clicking the controls above! This simulates real live PC feedback.</span>
                </div>
                <Link href="/connect" className="text-blue-400 hover:text-blue-300 font-medium flex items-center space-x-1">
                  <span>Open Full Controller</span>
                  <ChevronRight className="w-3 h-3" />
                </Link>
              </div>

            </div>
          </motion.div>

        </section>

        {/* Feature Bento Grid Section (Apple HIG Masterpiece) */}
        <section id="features" className="py-24 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-20">
            <span className="text-xs font-mono uppercase tracking-widest text-blue-400 font-bold">Architecture & Features</span>
            <h2 className="text-3xl sm:text-5xl font-black tracking-tight text-white mt-3">
              Engineered for pure speed and elegance.
            </h2>
            <p className="text-lg text-[#86868b] max-w-2xl mx-auto mt-4">
              Built from scratch using native Go and Win32 COM APIs to deliver uncompromising performance and instant responsiveness.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">

            {/* Bento 1: Real-Time Screen Mirroring (Span 2) */}
            <div className="md:col-span-2 rounded-[2rem] bg-gradient-to-b from-white/[0.05] to-white/[0.02] border border-white/[0.08] p-8 sm:p-10 flex flex-col justify-between relative overflow-hidden group hover:border-white/20 transition-all duration-300">
              <div className="relative z-10 max-w-lg">
                <div className="w-12 h-12 rounded-2xl bg-purple-500/20 border border-purple-500/30 flex items-center justify-center text-purple-400 mb-6">
                  <MonitorSmartphone className="w-6 h-6" />
                </div>
                <h3 className="text-2xl font-bold text-white mb-3">Ultra-Low Latency Desktop Mirroring</h3>
                <p className="text-[#86868b] leading-relaxed text-sm sm:text-base">
                  Stream your Windows desktop in real-time over WebSockets with hardware-accelerated H.264/MJPEG encoding. Packed with a custom 13-byte binary protocol achieving ~80-150ms latency on local networks.
                </p>
                <div className="flex flex-wrap gap-2 mt-6">
                  <span className="text-xs font-mono px-3 py-1 rounded-full bg-white/[0.06] border border-white/10 text-white/80">60 FPS Smooth</span>
                  <span className="text-xs font-mono px-3 py-1 rounded-full bg-white/[0.06] border border-white/10 text-white/80">Direct Touch Gestures</span>
                  <span className="text-xs font-mono px-3 py-1 rounded-full bg-white/[0.06] border border-white/10 text-white/80">Liquid Glass Toolbar</span>
                </div>
              </div>

              {/* Graphic Visual Representation */}
              <div className="mt-8 pt-6 border-t border-white/[0.08] flex items-center justify-between">
                <div className="flex items-center space-x-3 text-xs text-white/70">
                  <div className="w-2 h-2 rounded-full bg-green-400 animate-ping"></div>
                  <span>LAN: 85ms • WAN (Tunnel): 135ms</span>
                </div>
                <span className="text-xs font-mono text-purple-400">WebSocket /screen/stream</span>
              </div>
            </div>

            {/* Bento 2: Cloudflare Zero-Config Tunnel (Span 1) */}
            <div className="rounded-[2rem] bg-gradient-to-b from-white/[0.05] to-white/[0.02] border border-white/[0.08] p-8 flex flex-col justify-between hover:border-white/20 transition-all duration-300">
              <div>
                <div className="w-12 h-12 rounded-2xl bg-blue-500/20 border border-blue-500/30 flex items-center justify-center text-blue-400 mb-6">
                  <Globe className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-white mb-3">Zero-Config Cloudflare Tunnel</h3>
                <p className="text-[#86868b] text-sm leading-relaxed">
                  Automatic, out-of-the-box public WAN access. Control your PC from cellular data (4G/5G) or across router AP isolation without touching port forwarding.
                </p>
              </div>
              <div className="mt-6 p-3 rounded-xl bg-black/40 border border-white/10 text-xs font-mono text-blue-400">
                *.trycloudflare.com
              </div>
            </div>

            {/* Bento 3: AES-256 QR Instant Pairing (Span 1) */}
            <div className="rounded-[2rem] bg-gradient-to-b from-white/[0.05] to-white/[0.02] border border-white/[0.08] p-8 flex flex-col justify-between hover:border-white/20 transition-all duration-300">
              <div>
                <div className="w-12 h-12 rounded-2xl bg-emerald-500/20 border border-emerald-500/30 flex items-center justify-center text-emerald-400 mb-6">
                  <QrCode className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-white mb-3">AES-256 QR Pairing</h3>
                <p className="text-[#86868b] text-sm leading-relaxed">
                  Instant camera scan pairing with cryptographic one-time tokens and TLS certificate pinning. Never type manual IP addresses or PINs again.
                </p>
              </div>
              <div className="mt-6 flex items-center space-x-2 text-xs text-emerald-400 font-mono">
                <CheckCircle2 className="w-4 h-4 shrink-0" />
                <span>Single-use 5min TTL token</span>
              </div>
            </div>

            {/* Bento 4: Core Audio & GSMTC Sync (Span 2) */}
            <div className="md:col-span-2 rounded-[2rem] bg-gradient-to-b from-white/[0.05] to-white/[0.02] border border-white/[0.08] p-8 sm:p-10 flex flex-col justify-between hover:border-white/20 transition-all duration-300">
              <div>
                <div className="w-12 h-12 rounded-2xl bg-amber-500/20 border border-amber-500/30 flex items-center justify-center text-amber-400 mb-6">
                  <Sliders className="w-6 h-6" />
                </div>
                <h3 className="text-2xl font-bold text-white mb-3">Deep Windows Core Audio & GSMTC</h3>
                <p className="text-[#86868b] leading-relaxed text-sm sm:text-base max-w-xl">
                  Leveraging native Windows COM interfaces (`IMMDeviceEnumerator`, `IAudioEndpointVolume`). Inspect and adjust individual session volumes, including SteelSeries Sonar virtual channels and live song metadata sync from Spotify, Chrome, Edge, and VLC.
                </p>
              </div>

              <div className="mt-8 pt-6 border-t border-white/[0.08] grid grid-cols-2 sm:grid-cols-3 gap-4 text-xs font-mono text-white/70">
                <div className="p-3 rounded-xl bg-white/[0.02] border border-white/[0.06]">
                  <span className="text-amber-400 font-bold block mb-1">COM STA Worker</span>
                  Locked OS thread
                </div>
                <div className="p-3 rounded-xl bg-white/[0.02] border border-white/[0.06]">
                  <span className="text-amber-400 font-bold block mb-1">GSMTC Sync</span>
                  Real-time metadata
                </div>
                <div className="p-3 rounded-xl bg-white/[0.02] border border-white/[0.06] col-span-2 sm:col-span-1">
                  <span className="text-amber-400 font-bold block mb-1">3-Pass Volume</span>
                  Per-session fallback
                </div>
              </div>
            </div>

            {/* Bento 5: Dual-Stack Auto-Failover (Span 3) */}
            <div className="md:col-span-3 rounded-[2rem] bg-gradient-to-r from-blue-900/20 via-purple-900/20 to-indigo-900/20 border border-white/[0.08] p-8 sm:p-10 flex flex-col md:flex-row items-center justify-between gap-8 hover:border-white/20 transition-all duration-300">
              <div className="max-w-xl text-left">
                <div className="inline-flex items-center space-x-2 px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-400 text-xs font-medium mb-4">
                  <Layers className="w-3.5 h-3.5" />
                  <span>Intelligent Connectivity</span>
                </div>
                <h3 className="text-2xl font-bold text-white mb-2">Dual-Stack Auto-Failover</h3>
                <p className="text-[#86868b] text-sm sm:text-base leading-relaxed">
                  The client app continuously monitors network latency. When you step out of your home WiFi, it transparently migrates your active session to the Cloudflare WAN tunnel within 200ms without interrupting media playback or screen streaming.
                </p>
              </div>

              <div className="flex flex-col sm:flex-row gap-4 w-full md:w-auto shrink-0">
                <div className="px-5 py-4 rounded-2xl bg-black/60 border border-white/10 text-center">
                  <div className="text-xs text-white/50 mb-1 font-mono">PRIMARY</div>
                  <div className="text-sm font-bold text-emerald-400 flex items-center justify-center space-x-1.5">
                    <Wifi className="w-4 h-4" />
                    <span>LAN HTTPS</span>
                  </div>
                </div>
                <div className="px-5 py-4 rounded-2xl bg-black/60 border border-white/10 text-center">
                  <div className="text-xs text-white/50 mb-1 font-mono">FALLBACK</div>
                  <div className="text-sm font-bold text-blue-400 flex items-center justify-center space-x-1.5">
                    <Globe className="w-4 h-4" />
                    <span>WAN Tunnel</span>
                  </div>
                </div>
              </div>
            </div>

          </div>
        </section>

        {/* Screen Mirroring Deep Dive Showcase */}
        <section id="screen-mirror" className="py-24 border-t border-white/[0.08] bg-gradient-to-b from-transparent via-purple-950/10 to-transparent">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 items-center">
              
              <div className="lg:col-span-6 text-left">
                <div className="inline-flex items-center space-x-2 px-3 py-1 rounded-full bg-purple-500/10 border border-purple-500/20 text-purple-300 text-xs font-medium mb-4">
                  <Sparkles className="w-3.5 h-3.5 text-purple-400" />
                  <span>Flagship Feature in v5.0</span>
                </div>
                <h2 className="text-3xl sm:text-5xl font-black tracking-tight text-white mb-6">
                  Desktop Mirroring. <br />
                  <span className="text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-pink-500">
                    With full touch injection.
                  </span>
                </h2>
                <p className="text-lg text-[#86868b] leading-relaxed mb-8">
                  Watch your monitor on your mobile screen in fluid 60 frames per second. Tap, drag, right-click, and scroll as naturally as if you were using a wireless trackpad.
                </p>

                <div className="space-y-4">
                  <div className="flex items-start space-x-3.5">
                    <div className="w-6 h-6 rounded-full bg-purple-500/20 flex items-center justify-center shrink-0 mt-0.5">
                      <CheckCircle2 className="w-4 h-4 text-purple-400" />
                    </div>
                    <div>
                      <h4 className="text-base font-semibold text-white">Touch-to-Click & Coordinate Normalization</h4>
                      <p className="text-sm text-[#86868b]">Maps touch coordinates (0.0 to 1.0) directly to high-DPI Windows displays with SendInput injection.</p>
                    </div>
                  </div>

                  <div className="flex items-start space-x-3.5">
                    <div className="w-6 h-6 rounded-full bg-purple-500/20 flex items-center justify-center shrink-0 mt-0.5">
                      <CheckCircle2 className="w-4 h-4 text-purple-400" />
                    </div>
                    <div>
                      <h4 className="text-base font-semibold text-white">Liquid Glass Floating Toolbar</h4>
                      <p className="text-sm text-[#86868b]">Collapsible on-screen toolbar to switch stream quality, toggle fullscreen, or trigger on-screen keyboards.</p>
                    </div>
                  </div>

                  <div className="flex items-start space-x-3.5">
                    <div className="w-6 h-6 rounded-full bg-purple-500/20 flex items-center justify-center shrink-0 mt-0.5">
                      <CheckCircle2 className="w-4 h-4 text-purple-400" />
                    </div>
                    <div>
                      <h4 className="text-base font-semibold text-white">Adaptive Quality Compression</h4>
                      <p className="text-sm text-[#86868b]">Dynamically scales bitrates from 720p 60fps on local WiFi down to optimized frames over mobile data.</p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Visual Mockup */}
              <div className="lg:col-span-6">
                <div className="relative rounded-3xl p-4 bg-white/[0.04] border border-white/[0.1] shadow-2xl backdrop-blur-xl">
                  <div className="rounded-2xl bg-black/80 aspect-[16/10] overflow-hidden relative border border-white/[0.08] flex items-center justify-center">
                    
                    {/* Simulated Screen Mirror Desktop */}
                    <div className="absolute inset-0 bg-gradient-to-tr from-slate-900 via-indigo-950 to-slate-900 opacity-90 p-4 flex flex-col justify-between">
                      <div className="flex justify-between items-center text-[10px] font-mono text-white/50">
                        <span>DISPLAY 1 (2560 x 1440)</span>
                        <span className="px-2 py-0.5 rounded bg-green-500/20 text-green-400 font-bold">LIVE STREAMING</span>
                      </div>

                      {/* Mock Desktop Icons / Cursor */}
                      <div className="relative w-full h-full flex items-center justify-center">
                        <div className="w-4 h-4 rounded-full bg-blue-400/80 animate-ping absolute"></div>
                        <div className="w-3 h-3 rounded-full bg-white shadow-lg shadow-blue-500 border border-blue-400 z-10"></div>
                        <span className="ml-16 mt-2 text-xs font-mono text-white/60 bg-black/60 px-2 py-0.5 rounded backdrop-blur">Touch Injected (x: 0.54, y: 0.48)</span>
                      </div>

                      {/* Floating Glass Toolbar Simulation */}
                      <div className="mx-auto px-4 py-2 rounded-full glass border border-white/20 flex items-center space-x-4 shadow-xl">
                        <span className="text-[11px] font-medium text-white/90">Quality: 1080p</span>
                        <div className="h-3 w-px bg-white/20"></div>
                        <span className="text-[11px] font-mono text-purple-300">60 FPS</span>
                        <div className="h-3 w-px bg-white/20"></div>
                        <span className="text-[11px] text-white/70">Keyboard</span>
                      </div>
                    </div>

                  </div>
                </div>
              </div>

            </div>
          </div>
        </section>

        {/* Download Hub */}
        <section id="download" className="py-24 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <span className="text-xs font-mono uppercase tracking-widest text-blue-400 font-bold">Release Downloads</span>
          <h2 className="text-3xl sm:text-5xl font-black tracking-tight text-white mt-3 mb-6">
            Get PC Remote for your devices.
          </h2>
          <p className="text-lg text-[#86868b] max-w-2xl mx-auto mb-16">
            Download pre-built production packages. Zero compiling or developer toolchains required.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-5xl mx-auto">
            
            {/* Windows Server Setup Card */}
            <div className="rounded-[2.5rem] bg-gradient-to-b from-blue-600/15 via-white/[0.03] to-white/[0.02] border border-blue-500/30 p-8 flex flex-col justify-between text-left hover:scale-[1.02] transition-transform duration-300 shadow-xl">
              <div>
                <div className="flex justify-between items-center mb-6">
                  <div className="w-14 h-14 rounded-2xl bg-blue-600/20 text-blue-400 flex items-center justify-center border border-blue-500/30">
                    <Download className="w-7 h-7" />
                  </div>
                  <span className="text-xs font-mono px-3 py-1 rounded-full bg-blue-500/20 text-blue-300 border border-blue-500/30 font-medium">Windows 10 / 11</span>
                </div>
                <h3 className="text-2xl font-bold text-white mb-2">Windows Setup</h3>
                <p className="text-sm text-[#86868b] mb-6">
                  Installer wizard. Packs Go backend, system tray app, and autostart registration.
                </p>
                <div className="text-xs font-mono text-white/50 space-y-1 mb-8">
                  <div>Package: PCRemoteSetup.exe</div>
                  <div>Size: ~12 MB • UAC Safe Autostart</div>
                </div>
              </div>

              <a 
                href="https://github.com/KristianEki11/PC-Remote/releases/latest" 
                target="_blank" 
                rel="noreferrer"
                className="w-full py-4 rounded-2xl bg-blue-600 hover:bg-blue-500 text-white font-semibold text-center shadow-lg shadow-blue-600/25 active:scale-95 transition-all flex items-center justify-center space-x-2"
              >
                <span>Download .exe Installer</span>
                <ChevronRight className="w-4 h-4" />
              </a>
            </div>

            {/* Android APK Card */}
            <div className="rounded-[2.5rem] bg-gradient-to-b from-emerald-600/15 via-white/[0.03] to-white/[0.02] border border-emerald-500/30 p-8 flex flex-col justify-between text-left hover:scale-[1.02] transition-transform duration-300 shadow-xl">
              <div>
                <div className="flex justify-between items-center mb-6">
                  <div className="w-14 h-14 rounded-2xl bg-emerald-600/20 text-emerald-400 flex items-center justify-center border border-emerald-500/30">
                    <MonitorSmartphone className="w-7 h-7" />
                  </div>
                  <span className="text-xs font-mono px-3 py-1 rounded-full bg-emerald-500/20 text-emerald-300 border border-emerald-500/30 font-medium">Android 8.0+</span>
                </div>
                <h3 className="text-2xl font-bold text-white mb-2">Android App</h3>
                <p className="text-sm text-[#86868b] mb-6">
                  Native Flutter client with full dark mode, QR scanner camera, and touch screen mirror.
                </p>
                <div className="text-xs font-mono text-white/50 space-y-1 mb-8">
                  <div>Package: PCRemoteApp.apk</div>
                  <div>Direct APK install • No Google Play needed</div>
                </div>
              </div>

              <a 
                href="https://github.com/KristianEki11/PC-Remote/releases/latest" 
                target="_blank" 
                rel="noreferrer"
                className="w-full py-4 rounded-2xl bg-emerald-600 hover:bg-emerald-500 text-white font-semibold text-center shadow-lg shadow-emerald-600/25 active:scale-95 transition-all flex items-center justify-center space-x-2"
              >
                <span>Download .apk Package</span>
                <ChevronRight className="w-4 h-4" />
              </a>
            </div>

            {/* iOS Web Remote PWA Card */}
            <div className="rounded-[2.5rem] bg-gradient-to-b from-purple-600/15 via-white/[0.03] to-white/[0.02] border border-purple-500/30 p-8 flex flex-col justify-between text-left hover:scale-[1.02] transition-transform duration-300 shadow-xl">
              <div>
                <div className="flex justify-between items-center mb-6">
                  <div className="w-14 h-14 rounded-2xl bg-purple-600/20 text-purple-400 flex items-center justify-center border border-purple-500/30">
                    <Radio className="w-7 h-7" />
                  </div>
                  <span className="text-xs font-mono px-3 py-1 rounded-full bg-purple-500/20 text-purple-300 border border-purple-500/30 font-medium">iOS / Safari PWA</span>
                </div>
                <h3 className="text-2xl font-bold text-white mb-2">iOS Web Remote</h3>
                <p className="text-sm text-[#86868b] mb-6">
                  PWA controller for iPhone. Built with camera QR scanner and responsive mobile viewport.
                </p>
                <div className="text-xs font-mono text-white/50 space-y-1 mb-8">
                  <div>Route: remote.redlinevis.site/connect</div>
                  <div>Add to Home Screen in Safari</div>
                </div>
              </div>

              <Link 
                href="/connect" 
                className="w-full py-4 rounded-2xl bg-purple-600 hover:bg-purple-500 text-white font-semibold text-center shadow-lg shadow-purple-600/25 active:scale-95 transition-all flex items-center justify-center space-x-2"
              >
                <span>Launch /connect PWA</span>
                <ExternalLink className="w-4 h-4" />
              </Link>
            </div>

          </div>
        </section>

        {/* Quickstart Guide */}
        <section id="quickstart" className="py-24 border-t border-white/[0.08] max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <span className="text-xs font-mono uppercase tracking-widest text-blue-400 font-bold">Getting Started</span>
            <h2 className="text-3xl sm:text-4xl font-bold text-white mt-3">Up and running in 3 minutes.</h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div className="p-8 rounded-3xl bg-white/[0.03] border border-white/[0.08] text-left">
              <div className="w-10 h-10 rounded-full bg-blue-500/20 text-blue-400 flex items-center justify-center font-bold text-lg mb-6 border border-blue-500/30">
                1
              </div>
              <h3 className="text-xl font-bold text-white mb-3">Install on Windows</h3>
              <p className="text-[#86868b] text-sm leading-relaxed">
                Run `PCRemoteSetup.exe` on your PC. Choose your 4-digit PIN (e.g. 1234). The installer automatically configures background autostart.
              </p>
            </div>

            <div className="p-8 rounded-3xl bg-white/[0.03] border border-white/[0.08] text-left">
              <div className="w-10 h-10 rounded-full bg-indigo-500/20 text-indigo-400 flex items-center justify-center font-bold text-lg mb-6 border border-indigo-500/30">
                2
              </div>
              <h3 className="text-xl font-bold text-white mb-3">Scan QR or Connect</h3>
              <p className="text-[#86868b] text-sm leading-relaxed">
                Launch the Android app or open `/connect` on iPhone. Tap the QR scanner button to scan the encrypted code from your PC Desktop dashboard.
              </p>
            </div>

            <div className="p-8 rounded-3xl bg-white/[0.03] border border-white/[0.08] text-left">
              <div className="w-10 h-10 rounded-full bg-purple-500/20 text-purple-400 flex items-center justify-center font-bold text-lg mb-6 border border-purple-500/30">
                3
              </div>
              <h3 className="text-xl font-bold text-white mb-3">Control Seamlessly</h3>
              <p className="text-[#86868b] text-sm leading-relaxed">
                You are securely paired. Enjoy instant media manipulation, master volume adjustments, display management, and live screen mirroring!
              </p>
            </div>
          </div>
        </section>

        {/* Documentation & Security Policy */}
        <section id="documentation" className="py-20 max-w-5xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="p-8 sm:p-12 rounded-[2.5rem] bg-white/[0.02] border border-white/[0.08] text-left">
            <div className="flex items-center space-x-3 mb-6">
              <Terminal className="w-6 h-6 text-blue-400" />
              <h3 className="text-2xl font-bold text-white">Security & Zero-Telemetry Contract</h3>
            </div>
            <div className="space-y-4 text-sm text-[#86868b] leading-relaxed">
              <p>
                PC Remote is designed with security as a non-negotiable cornerstone. All PIN codes are compared in constant-time (`crypto/subtle.ConstantTimeCompare`) to resist side-channel timing attacks. Session tokens are generated cryptographically and maintained strictly in volatile RAM.
              </p>
              <p>
                Zero personal data, metrics, or telemetry are ever harvested or sent to external third parties. All communication over the public internet leverages authenticated Cloudflare Quick Tunnels with TLS encryption.
              </p>
            </div>
          </div>
        </section>

      </main>

      {/* Footer with Copyright */}
      <footer className="py-12 border-t border-white/[0.08] relative z-10 bg-black/40 text-center text-sm text-[#86868b]">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-3">
          <div className="flex items-center justify-center space-x-2">
            <span className="font-semibold text-white">PC Remote Controller</span>
            <span>•</span>
            <span className="font-mono text-xs">remote.redlinevis.site</span>
          </div>
          <p className="text-xs text-[#86868b] max-w-xl mx-auto">
            &copy; 2026 kidev. All Rights Reserved. Provided for personal, non-commercial use only.
            Redistribution, resale, or incorporation into other software products is strictly prohibited.
          </p>
        </div>
      </footer>

    </div>
  );
}
