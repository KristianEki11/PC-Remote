"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { 
  Wifi, 
  Lock, 
  ChevronRight, 
  Power, 
  Moon, 
  Play, 
  Pause, 
  SkipForward, 
  SkipBack,
  Volume2,
  VolumeX,
  Server,
  QrCode,
  X,
  Camera,
  EyeOff
} from "lucide-react";
import { Html5QrcodeScanner } from "html5-qrcode";

export default function WebController() {
  const [ipAddress, setIpAddress] = useState("");
  const [pin, setPin] = useState(""); // Can be 4-digit PIN or long Pair Token
  const [isConnected, setIsConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  
  const [volume, setVolume] = useState(50);
  const [isPlaying, setIsPlaying] = useState(false);
  const [isMuted, setIsMuted] = useState(false);
  const [showQrScanner, setShowQrScanner] = useState(false);
  const [lastAction, setLastAction] = useState<string | null>(null);

  // Auto-reconnect if session exists
  useEffect(() => {
    const savedUrl = localStorage.getItem("pc_remote_url");
    const savedToken = localStorage.getItem("pc_remote_token");
    if (savedUrl && savedToken) {
      setIpAddress(savedUrl);
      setIsConnected(true);
      
      // Fetch initial status
      fetch(`${savedUrl}/audio/status`, {
        headers: { "Authorization": `Bearer ${savedToken}` }
      })
      .then(res => res.json())
      .then(data => {
        if (data.level !== undefined) setVolume(Math.round(data.level * 100));
        if (data.muted !== undefined) setIsMuted(data.muted);
      })
      .catch(() => {
        // Token likely expired or server offline
        setIsConnected(false);
        localStorage.removeItem("pc_remote_token");
      });
    }
  }, []);

  const handleConnect = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    setIsConnecting(true);
    
    try {
      let baseUrl = ipAddress.trim();
      if (!baseUrl.startsWith("http")) {
        baseUrl = `https://${baseUrl}`;
        if (!baseUrl.includes(":8000") && !baseUrl.includes("trycloudflare.com")) {
            baseUrl = `${baseUrl}:8000`;
        }
      }
      
      let token = "";
      if (pin.length > 8) {
        // Pair token
        const res = await fetch(`${baseUrl}/auth/pair`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ pair_token: pin })
        });
        if (!res.ok) throw new Error("Invalid pair token");
        const data = await res.json();
        token = data.token;
      } else {
        // 4-Digit PIN
        const res = await fetch(`${baseUrl}/auth/login`, {
          method: "POST",
          headers: { "Content-Type": "application/json", "X-Device-Name": "Web Remote" },
          body: JSON.stringify({ pin: pin })
        });
        if (!res.ok) throw new Error("Invalid PIN");
        const data = await res.json();
        token = data.token;
      }
      
      localStorage.setItem("pc_remote_url", baseUrl);
      localStorage.setItem("pc_remote_token", token);
      
      setIpAddress(baseUrl);
      setIsConnected(true);
      
      // Fetch initial state
      const statusRes = await fetch(`${baseUrl}/audio/status`, {
        headers: { "Authorization": `Bearer ${token}` }
      });
      if (statusRes.ok) {
        const data = await statusRes.json();
        if (data.level !== undefined) setVolume(Math.round(data.level * 100));
        if (data.muted !== undefined) setIsMuted(data.muted);
      }

    } catch (err) {
      alert("Failed to connect: " + (err as Error).message);
    } finally {
      setIsConnecting(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem("pc_remote_token");
    setIsConnected(false);
    setPin("");
  };

  useEffect(() => {
    let scanner: Html5QrcodeScanner | null = null;
    if (showQrScanner) {
      scanner = new Html5QrcodeScanner(
        "qr-reader",
        { fps: 10, qrbox: { width: 250, height: 250 } },
        false
      );

      scanner.render(
        (decodedText) => {
          try {
            if (decodedText.startsWith("{")) {
              const data = JSON.parse(decodedText);
              if (data.host) setIpAddress(data.host);
              if (data.pair_token) {
                setPin(data.pair_token);
                // Cannot trigger handleConnect directly here due to closure state limits,
                // but setting state is enough, user can click Connect
              }
            } else {
              setIpAddress(decodedText);
            }
          } catch {
            setIpAddress(decodedText);
          }
          scanner?.clear();
          setShowQrScanner(false);
        },
        () => {}
      );
    }
    return () => {
      if (scanner) scanner.clear().catch(console.error);
    };
  }, [showQrScanner]);

  const callBackend = async (endpoint: string, payload: any = {}, actionLabel?: string) => {
    if (actionLabel) {
      setLastAction(actionLabel);
      setTimeout(() => setLastAction(null), 2000);
    }
    
    const token = localStorage.getItem("pc_remote_token");
    const baseUrl = localStorage.getItem("pc_remote_url");
    if (!token || !baseUrl) return;
    
    try {
      await fetch(`${baseUrl}${endpoint}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`
        },
        body: JSON.stringify(payload)
      });
    } catch (e) {
      console.error("API Call Failed:", e);
    }
  };

  return (
    <div className="min-h-[100dvh] bg-[#09090b] text-[#f4f4f5] font-sans selection:bg-white/20 flex flex-col">
      <AnimatePresence mode="wait">
        {!isConnected ? (
          <motion.div 
            key="login"
            className="flex-1 flex flex-col p-6 justify-center max-w-md mx-auto w-full relative"
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, y: -20, filter: "blur(10px)" }}
            transition={{ duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
          >
            <div className="mb-10 text-center">
              <div className="mx-auto w-16 h-16 bg-white/5 border border-white/10 text-white rounded-2xl flex items-center justify-center mb-6">
                <Server className="w-8 h-8" />
              </div>
              <h1 className="text-2xl font-bold tracking-tight mb-2 text-white">Connect to PC</h1>
              <p className="text-[#a1a1aa] text-sm">Scan QR Code from dashboard or enter manually.</p>
            </div>

            <button
              type="button"
              onClick={() => setShowQrScanner(true)}
              className="mb-8 w-full bg-white/[0.02] hover:bg-white/[0.06] text-white border border-white/10 rounded-2xl py-4 flex items-center justify-center space-x-3 transition-all active:scale-95"
            >
              <QrCode className="w-5 h-5 text-[#a1a1aa]" />
              <span className="font-medium text-sm">Scan QR Code</span>
            </button>

            <div className="flex items-center mb-8">
              <div className="flex-1 border-t border-white/[0.08]"></div>
              <span className="px-4 text-[10px] font-mono text-[#71717a] uppercase tracking-wider">Manual Input</span>
              <div className="flex-1 border-t border-white/[0.08]"></div>
            </div>

            <form onSubmit={handleConnect} className="space-y-4">
              <div className="relative">
                <div className="absolute inset-y-0 left-4 flex items-center pointer-events-none text-[#71717a]">
                  <Wifi className="w-5 h-5" />
                </div>
                <input 
                  type="text" 
                  value={ipAddress}
                  onChange={(e) => setIpAddress(e.target.value)}
                  placeholder="PC IP / Tunnel URL"
                  className="w-full bg-[#121215] border border-white/10 focus:border-white/30 rounded-xl py-4 pl-12 pr-4 text-sm text-white outline-none transition-all placeholder:text-[#71717a]"
                  required
                />
              </div>
              
              <div className="relative">
                <div className="absolute inset-y-0 left-4 flex items-center pointer-events-none text-[#71717a]">
                  <Lock className="w-5 h-5" />
                </div>
                <input 
                  type="password" 
                  value={pin}
                  onChange={(e) => setPin(e.target.value)}
                  placeholder="4-Digit PIN or Pairing Token"
                  className="w-full bg-[#121215] border border-white/10 focus:border-white/30 rounded-xl py-4 pl-12 pr-4 text-sm text-white outline-none transition-all font-mono tracking-widest placeholder:text-[#71717a] placeholder:tracking-normal"
                  required
                />
              </div>

              <button 
                type="submit" 
                disabled={isConnecting || !ipAddress || !pin}
                className="w-full bg-white text-black rounded-xl py-4 font-semibold text-sm flex items-center justify-center space-x-2 hover:bg-white/90 active:scale-95 transition-all disabled:opacity-50 mt-6"
              >
                {isConnecting ? (
                  <div className="w-5 h-5 border-2 border-black/30 border-t-black rounded-full animate-spin" />
                ) : (
                  <>
                    <span>Connect</span>
                    <ChevronRight className="w-4 h-4" />
                  </>
                )}
              </button>
            </form>
          </motion.div>
        ) : (
          <motion.div 
            key="dashboard"
            className="flex-1 flex flex-col max-w-md mx-auto w-full relative"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.3 }}
          >
            <div className="bg-[#121215] px-6 py-5 border-b border-white/[0.06] flex items-center justify-between sticky top-0 z-10">
              <div>
                <div className="text-sm font-semibold text-white flex items-center space-x-2">
                  <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
                  <span>Windows PC</span>
                </div>
                <div className="text-[10px] font-mono text-[#71717a] mt-1 line-clamp-1">{ipAddress.replace('https://', '')}</div>
              </div>
              <div className="flex items-center space-x-3">
                {lastAction && (
                  <span className="text-[10px] font-mono px-2 py-1 rounded bg-white/10 text-white animate-pulse hidden sm:inline-block">
                    {lastAction}
                  </span>
                )}
                <button 
                  onClick={handleLogout}
                  className="p-2 bg-white/5 border border-white/10 hover:bg-white/10 rounded-lg text-white transition-colors"
                >
                  <Power className="w-4 h-4" />
                </button>
              </div>
            </div>

            <div className="flex-1 overflow-y-auto p-6">
              
              {/* Volume Control */}
              <div className="pb-8 border-b border-white/[0.06]">
                <div className="flex justify-between items-center text-xs mb-5 text-[#a1a1aa]">
                  <span className="font-medium uppercase tracking-wider">Master Audio</span>
                  <span className="font-mono text-white">{isMuted ? "MUTE" : `${volume}%`}</span>
                </div>
                <div className="flex items-center space-x-5">
                  <button 
                    onClick={() => {
                      const newMuted = !isMuted;
                      setIsMuted(newMuted);
                      callBackend("/audio/mute", { muted: newMuted }, newMuted ? "Audio Muted" : "Audio Unmuted");
                    }}
                    className="text-[#a1a1aa] hover:text-white p-2 rounded-lg hover:bg-white/5 transition-colors"
                  >
                    {isMuted ? <VolumeX className="w-6 h-6 text-red-400" /> : <Volume2 className="w-6 h-6" />}
                  </button>
                  <input 
                    type="range"
                    min="0"
                    max="100"
                    value={isMuted ? 0 : volume}
                    disabled={isMuted}
                    onChange={(e) => {
                      const val = parseInt(e.target.value);
                      setVolume(val);
                    }}
                    onMouseUp={(e) => {
                      const val = parseInt((e.target as HTMLInputElement).value);
                      callBackend("/audio/volume", { level: val / 100.0 }, `Volume: ${val}%`);
                    }}
                    onTouchEnd={(e) => {
                      const val = parseInt((e.target as HTMLInputElement).value);
                      callBackend("/audio/volume", { level: val / 100.0 }, `Volume: ${val}%`);
                    }}
                    className="w-full h-1.5 bg-white/10 rounded-full appearance-none cursor-pointer accent-white"
                  />
                </div>
              </div>

              {/* Media Player Controls */}
              <div className="py-8 border-b border-white/[0.06]">
                <div className="text-[10px] font-mono text-[#71717a] uppercase tracking-wider mb-2">Windows Media Transport</div>
                <div className="text-sm font-semibold text-white mb-6">Now Playing Control</div>

                <div className="flex items-center justify-between gap-4">
                  <button 
                    onClick={() => callBackend("/media/prev", {}, "Media: Previous")}
                    className="flex-1 py-4 flex justify-center rounded-xl border border-white/10 bg-white/[0.02] hover:bg-white/[0.06] text-white active:scale-95 transition-all"
                  >
                    <SkipBack className="w-5 h-5 fill-current" />
                  </button>
                  <button 
                    onClick={() => {
                      setIsPlaying(!isPlaying);
                      callBackend("/media/play", {}, isPlaying ? "Media: Paused" : "Media: Play");
                    }}
                    className="flex-1 py-4 flex justify-center rounded-xl bg-white text-black hover:bg-white/90 active:scale-95 transition-all shadow-[0_0_20px_rgba(255,255,255,0.1)]"
                  >
                    {isPlaying ? <Pause className="w-5 h-5 fill-current" /> : <Play className="w-5 h-5 fill-current" />}
                  </button>
                  <button 
                    onClick={() => callBackend("/media/next", {}, "Media: Next")}
                    className="flex-1 py-4 flex justify-center rounded-xl border border-white/10 bg-white/[0.02] hover:bg-white/[0.06] text-white active:scale-95 transition-all"
                  >
                    <SkipForward className="w-5 h-5 fill-current" />
                  </button>
                </div>
              </div>

              {/* Power Actions */}
              <div className="pt-8">
                <div className="text-[10px] font-mono text-[#71717a] uppercase tracking-wider mb-5">Power & System Actions</div>
                <div className="grid grid-cols-2 gap-4">
                  <button 
                    onClick={() => {
                      if (confirm("Turn off PC display?")) callBackend("/system/display/off", {}, "Display Off");
                    }}
                    className="py-4 px-4 rounded-xl border border-white/10 bg-white/[0.02] hover:bg-white/[0.06] text-xs font-medium text-white flex flex-col items-center justify-center space-y-3 active:scale-95 transition-all"
                  >
                    <EyeOff className="w-6 h-6 text-[#a1a1aa]" />
                    <span>Turn Off Display</span>
                  </button>
                  <button 
                    onClick={() => callBackend("/system/lock", {}, "PC Locked")}
                    className="py-4 px-4 rounded-xl border border-white/10 bg-white/[0.02] hover:bg-white/[0.06] text-xs font-medium text-white flex flex-col items-center justify-center space-y-3 active:scale-95 transition-all"
                  >
                    <Lock className="w-6 h-6 text-[#a1a1aa]" />
                    <span>Lock PC</span>
                  </button>
                  <button 
                    onClick={() => {
                      if (confirm("Put PC to sleep?")) callBackend("/system/sleep", {}, "S3 Sleep");
                    }}
                    className="col-span-2 py-4 px-4 rounded-xl border border-white/10 bg-white/[0.02] hover:bg-white/[0.06] text-xs font-medium text-white flex items-center justify-center space-x-3 active:scale-95 transition-all"
                  >
                    <Moon className="w-5 h-5 text-[#a1a1aa]" />
                    <span>Sleep Mode</span>
                  </button>
                </div>
              </div>

            </div>
          </motion.div>
        )}
      </AnimatePresence>
      
      <AnimatePresence>
        {showQrScanner && (
          <motion.div 
            className="fixed inset-0 bg-black/90 backdrop-blur-xl z-50 flex flex-col justify-center items-center p-6"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
          >
            <div className="bg-[#121215] rounded-3xl p-6 w-full max-w-sm text-center relative border border-white/10">
              <button 
                onClick={() => setShowQrScanner(false)}
                className="absolute top-4 right-4 p-2 text-[#71717a] hover:text-white rounded-full transition-colors"
              >
                <X className="w-6 h-6" />
              </button>
              
              <div className="flex items-center justify-center space-x-3 mb-6">
                <Camera className="w-5 h-5 text-white" />
                <h3 className="font-semibold text-white">Scan PC QR Code</h3>
              </div>
              
              <div id="qr-reader" className="w-full overflow-hidden rounded-2xl border border-white/10 bg-black"></div>
              <p className="text-xs text-[#71717a] mt-6">Point camera at QR Code shown on your PC dashboard</p>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
