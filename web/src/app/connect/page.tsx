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
  Camera
} from "lucide-react";
import { Html5QrcodeScanner } from "html5-qrcode";

export default function WebController() {
  const [ipAddress, setIpAddress] = useState("");
  const [pin, setPin] = useState("");
  const [isConnected, setIsConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [volume, setVolume] = useState(50);
  const [isPlaying, setIsPlaying] = useState(false);
  const [showQrScanner, setShowQrScanner] = useState(false);

  // Authentication Handler
  const handleConnect = (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    setIsConnecting(true);
    
    // Simulate connection delay
    setTimeout(() => {
      setIsConnecting(false);
      setIsConnected(true);
    }, 1500);
  };

  // QR Scanner Initialization
  useEffect(() => {
    let scanner: Html5QrcodeScanner | null = null;
    
    if (showQrScanner) {
      scanner = new Html5QrcodeScanner(
        "qr-reader",
        { fps: 10, qrbox: { width: 250, height: 250 } },
        /* verbose= */ false
      );

      scanner.render(
        (decodedText) => {
          console.log("QR Scanned Payload:", decodedText);
          try {
            // Attempt JSON parse if it's raw JSON or handle encrypted string
            if (decodedText.startsWith("{")) {
              const data = JSON.parse(decodedText);
              if (data.host) setIpAddress(data.host);
              if (data.pair_token) setPin(data.pair_token);
            } else if (decodedText.includes(":")) {
              // Standard IP:PORT or URL
              setIpAddress(decodedText);
            } else {
              setIpAddress(decodedText);
            }
          } catch {
            setIpAddress(decodedText);
          }
          scanner?.clear();
          setShowQrScanner(false);
        },
        () => {
          // Scanner error or scanning...
        }
      );
    }

    return () => {
      if (scanner) {
        scanner.clear().catch(console.error);
      }
    };
  }, [showQrScanner]);

  // REST API Mock Calls
  const callBackend = async (endpoint: string, method = "POST", data = {}) => {
    console.log(`[Mock API Call] ${method} http://${ipAddress}:8000/api/${endpoint}`, data);
  };

  const handlePowerAction = (action: string) => {
    callBackend(`power/${action}`);
  };

  const handleMediaAction = (action: string) => {
    callBackend(`media/${action}`);
    if (action === 'play-pause') setIsPlaying(!isPlaying);
  };

  const handleVolumeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newVol = parseInt(e.target.value);
    setVolume(newVol);
    callBackend('media/volume', 'POST', { volume: newVol });
  };

  return (
    <div className="min-h-[100dvh] bg-background text-foreground font-sans overflow-hidden selection:bg-blue-500/30">
      <AnimatePresence mode="wait">
        {!isConnected ? (
          <motion.div 
            key="login"
            className="flex flex-col h-[100dvh] p-6 justify-center max-w-md mx-auto relative"
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, y: -20, filter: "blur(10px)" }}
            transition={{ duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
          >
            <div className="mb-8 text-center">
              <div className="mx-auto w-16 h-16 bg-blue-500/10 dark:bg-blue-500/20 text-blue-500 rounded-3xl flex items-center justify-center mb-4 shadow-sm">
                <Server className="w-8 h-8" />
              </div>
              <h1 className="text-3xl font-bold tracking-tight mb-2">Connect to PC</h1>
              <p className="text-foreground/50 text-sm">Scan QR Code from PC Dashboard or enter credentials manually.</p>
            </div>

            {/* QR Scan Button */}
            <button
              onClick={() => setShowQrScanner(true)}
              className="mb-6 w-full bg-foreground/5 hover:bg-foreground/10 text-foreground border border-foreground/10 rounded-2xl py-3 px-4 flex items-center justify-center space-x-3 transition-all active:scale-[0.98]"
            >
              <QrCode className="w-5 h-5 text-blue-500" />
              <span className="font-semibold text-sm">Scan QR Code from PC</span>
            </button>

            <div className="flex items-center my-4">
              <div className="flex-1 border-t border-foreground/10"></div>
              <span className="px-3 text-xs text-foreground/40 font-medium uppercase">Or Manual Input</span>
              <div className="flex-1 border-t border-foreground/10"></div>
            </div>

            <form onSubmit={handleConnect} className="space-y-4">
              <div className="space-y-3 bg-foreground/[0.03] dark:bg-white/[0.03] p-4 rounded-3xl border border-foreground/[0.05] dark:border-white/[0.05]">
                <div className="relative">
                  <div className="absolute inset-y-0 left-4 flex items-center pointer-events-none text-foreground/40">
                    <Wifi className="w-5 h-5" />
                  </div>
                  <input 
                    type="text" 
                    value={ipAddress}
                    onChange={(e) => setIpAddress(e.target.value)}
                    placeholder="PC IP Address / Tunnel URL"
                    className="w-full bg-background border border-foreground/10 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 rounded-2xl py-3.5 pl-12 pr-4 text-base outline-none transition-all"
                    required
                  />
                </div>
                
                <div className="relative">
                  <div className="absolute inset-y-0 left-4 flex items-center pointer-events-none text-foreground/40">
                    <Lock className="w-5 h-5" />
                  </div>
                  <input 
                    type="password" 
                    value={pin}
                    onChange={(e) => setPin(e.target.value)}
                    placeholder="PIN Code"
                    className="w-full bg-background border border-foreground/10 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 rounded-2xl py-3.5 pl-12 pr-4 text-base outline-none transition-all font-mono"
                    required
                  />
                </div>
              </div>

              <button 
                type="submit" 
                disabled={isConnecting || !ipAddress || !pin}
                className="w-full bg-blue-500 text-white rounded-2xl py-4 font-semibold text-lg flex items-center justify-center space-x-2 hover:bg-blue-600 active:scale-[0.98] transition-all disabled:opacity-50 disabled:active:scale-100 mt-4"
              >
                {isConnecting ? (
                  <div className="w-6 h-6 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                ) : (
                  <>
                    <span>Connect</span>
                    <ChevronRight className="w-5 h-5" />
                  </>
                )}
              </button>
            </form>

            {/* QR Scanner Modal */}
            <AnimatePresence>
              {showQrScanner && (
                <motion.div 
                  className="fixed inset-0 bg-black/80 backdrop-blur-md z-50 flex flex-col justify-center items-center p-6"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                >
                  <div className="bg-background rounded-3xl p-6 w-full max-w-sm text-center relative border border-foreground/10">
                    <button 
                      onClick={() => setShowQrScanner(false)}
                      className="absolute top-4 right-4 p-2 text-foreground/60 hover:text-foreground rounded-full"
                    >
                      <X className="w-6 h-6" />
                    </button>
                    
                    <div className="flex items-center justify-center space-x-2 mb-4">
                      <Camera className="w-5 h-5 text-blue-500" />
                      <h3 className="font-bold text-lg">Scan PC QR Code</h3>
                    </div>
                    
                    <div id="qr-reader" className="w-full overflow-hidden rounded-2xl border border-foreground/10"></div>
                    <p className="text-xs text-foreground/50 mt-4">Point camera at QR Code shown on your PC dashboard</p>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </motion.div>
        ) : (
          <motion.div 
            key="dashboard"
            className="flex flex-col h-[100dvh] bg-gray-100 dark:bg-black"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
          >
            {/* Dashboard Header */}
            <div className="pt-14 pb-6 px-6 bg-white dark:bg-[#1c1c1e] rounded-b-[2.5rem] shadow-sm border-b border-gray-200 dark:border-white/5 z-10 relative">
              <div className="flex justify-between items-center mb-6">
                <div>
                  <h2 className="text-2xl font-bold tracking-tight">Windows PC</h2>
                  <p className="text-sm text-green-500 font-medium flex items-center mt-1">
                    <span className="w-2 h-2 bg-green-500 rounded-full mr-2"></span>
                    Connected to {ipAddress}
                  </p>
                </div>
                <button 
                  onClick={() => setIsConnected(false)}
                  className="bg-gray-100 dark:bg-white/10 text-foreground p-3 rounded-full hover:bg-gray-200 dark:hover:bg-white/20 transition-colors"
                >
                  <Power className="w-5 h-5" />
                </button>
              </div>
            </div>

            {/* Controls */}
            <div className="flex-1 overflow-y-auto p-6 space-y-6 pb-12">
              
              {/* Media Controls */}
              <div className="bg-white dark:bg-[#1c1c1e] p-6 rounded-[2rem] shadow-sm border border-gray-100 dark:border-white/5">
                <h3 className="text-lg font-semibold mb-6 flex items-center text-foreground/80">
                  <Play className="w-5 h-5 mr-2" />
                  Media Playback
                </h3>
                
                <div className="flex justify-center items-center space-x-6 mb-8">
                  <button 
                    onClick={() => handleMediaAction('prev')}
                    className="p-4 rounded-full bg-gray-100 dark:bg-white/5 hover:bg-gray-200 dark:hover:bg-white/10 transition-colors active:scale-95"
                  >
                    <SkipBack className="w-6 h-6 fill-current" />
                  </button>
                  <button 
                    onClick={() => handleMediaAction('play-pause')}
                    className="p-6 rounded-full bg-foreground text-background hover:scale-105 active:scale-95 transition-all shadow-lg"
                  >
                    {isPlaying ? <Pause className="w-8 h-8 fill-current" /> : <Play className="w-8 h-8 fill-current translate-x-0.5" />}
                  </button>
                  <button 
                    onClick={() => handleMediaAction('next')}
                    className="p-4 rounded-full bg-gray-100 dark:bg-white/5 hover:bg-gray-200 dark:hover:bg-white/10 transition-colors active:scale-95"
                  >
                    <SkipForward className="w-6 h-6 fill-current" />
                  </button>
                </div>

                {/* Volume Slider */}
                <div className="flex items-center space-x-4">
                  <VolumeX className="w-5 h-5 text-foreground/50" />
                  <input 
                    type="range" 
                    min="0" 
                    max="100" 
                    value={volume}
                    onChange={handleVolumeChange}
                    className="flex-1 h-2 bg-gray-200 dark:bg-white/10 rounded-full appearance-none [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-6 [&::-webkit-slider-thumb]:h-6 [&::-webkit-slider-thumb]:bg-foreground [&::-webkit-slider-thumb]:rounded-full shadow-sm"
                  />
                  <Volume2 className="w-5 h-5 text-foreground/50" />
                </div>
              </div>

              {/* Power Actions */}
              <div className="bg-white dark:bg-[#1c1c1e] p-6 rounded-[2rem] shadow-sm border border-gray-100 dark:border-white/5">
                <h3 className="text-lg font-semibold mb-6 flex items-center text-foreground/80">
                  <Power className="w-5 h-5 mr-2" />
                  Power Actions
                </h3>
                
                <div className="grid grid-cols-2 gap-4">
                  <button 
                    onClick={() => handlePowerAction('sleep')}
                    className="flex flex-col items-center justify-center p-4 rounded-2xl bg-blue-50 dark:bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-100 dark:border-blue-500/20 hover:bg-blue-100 dark:hover:bg-blue-500/20 transition-colors active:scale-95"
                  >
                    <Moon className="w-6 h-6 mb-2" />
                    <span className="font-medium">Sleep</span>
                  </button>
                  <button 
                    onClick={() => handlePowerAction('lock')}
                    className="flex flex-col items-center justify-center p-4 rounded-2xl bg-amber-50 dark:bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-100 dark:border-amber-500/20 hover:bg-amber-100 dark:hover:bg-amber-500/20 transition-colors active:scale-95"
                  >
                    <Lock className="w-6 h-6 mb-2" />
                    <span className="font-medium">Lock PC</span>
                  </button>
                  <button 
                    onClick={() => handlePowerAction('shutdown')}
                    className="col-span-2 flex flex-col items-center justify-center p-4 rounded-2xl bg-red-50 dark:bg-red-500/10 text-red-600 dark:text-red-400 border border-red-100 dark:border-red-500/20 hover:bg-red-100 dark:hover:bg-red-500/20 transition-colors active:scale-95 mt-2"
                  >
                    <Power className="w-6 h-6 mb-2" />
                    <span className="font-medium">Shutdown</span>
                  </button>
                </div>
              </div>

            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
