"use client";

import { motion } from "framer-motion";
import { MonitorSmartphone, Shield, Zap, Globe, Download, ChevronRight, BookOpen, Layers } from "lucide-react";
import Link from "next/link";

export default function LandingPage() {
  const features = [
    {
      title: "WAN Support via Cloudflare",
      description: "Connect to your PC from anywhere in the world securely through Cloudflare tunnels.",
      icon: <Globe className="w-6 h-6 text-blue-500" />
    },
    {
      title: "Zero-config",
      description: "No port forwarding or complicated network setup required. It just works.",
      icon: <Zap className="w-6 h-6 text-yellow-500" />
    },
    {
      title: "Screen Mirroring",
      description: "View and control your PC screen directly from your mobile device with ultra-low latency.",
      icon: <MonitorSmartphone className="w-6 h-6 text-purple-500" />
    },
    {
      title: "Dual-Stack Auto-Failover",
      description: "Seamlessly switches between LAN and WAN for the most optimal connection.",
      icon: <Layers className="w-6 h-6 text-indigo-500" />
    },
    {
      title: "AES QR Pairing",
      description: "Military-grade AES encryption ensures your connection is secure from the moment you pair.",
      icon: <Shield className="w-6 h-6 text-green-500" />
    }
  ];

  return (
    <div className="min-h-screen flex flex-col bg-background text-foreground overflow-hidden font-sans">
      {/* Navigation */}
      <nav className="fixed top-0 w-full z-50 glass">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-2">
              <MonitorSmartphone className="w-8 h-8" />
              <span className="font-semibold text-xl tracking-tight">PC Remote</span>
            </div>
            <div className="hidden md:flex space-x-8">
              <a href="#features" className="text-sm text-foreground/80 hover:text-foreground transition-colors">Features</a>
              <a href="#download" className="text-sm text-foreground/80 hover:text-foreground transition-colors">Download</a>
              <a href="#docs" className="text-sm text-foreground/80 hover:text-foreground transition-colors">Docs</a>
            </div>
            <Link href="/connect" className="bg-foreground text-background px-4 py-2 rounded-full text-sm font-medium hover:opacity-90 transition-opacity">
              Web Remote
            </Link>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <main className="flex-grow pt-32 pb-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <motion.h1 
            className="text-5xl md:text-7xl font-bold tracking-tight mb-6"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
          >
            Your PC, <br />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-blue-500 to-purple-600">
              in the palm of your hand.
            </span>
          </motion.h1>
          <motion.p 
            className="text-xl md:text-2xl text-foreground/70 max-w-2xl mx-auto mb-10"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.1, ease: [0.16, 1, 0.3, 1] }}
          >
            Control your Windows PC from anywhere. Secure, fast, and remarkably simple.
          </motion.p>
          
          <motion.div 
            className="flex flex-col sm:flex-row justify-center items-center space-y-4 sm:space-y-0 sm:space-x-6"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.2, ease: [0.16, 1, 0.3, 1] }}
          >
            <a href="#download" className="w-full sm:w-auto flex items-center justify-center space-x-2 bg-foreground text-background px-8 py-4 rounded-full text-lg font-medium hover:scale-105 transition-transform">
              <Download className="w-5 h-5" />
              <span>Get Started</span>
            </a>
            <Link href="/connect" className="w-full sm:w-auto flex items-center justify-center space-x-2 bg-foreground/5 dark:bg-white/10 text-foreground px-8 py-4 rounded-full text-lg font-medium hover:bg-foreground/10 dark:hover:bg-white/20 transition-colors">
              <span>Open Web Remote</span>
              <ChevronRight className="w-5 h-5" />
            </Link>
          </motion.div>
        </div>

        {/* Features Section */}
        <div id="features" className="mt-32 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">Powerful features. Simple design.</h2>
            <p className="text-lg text-foreground/70 max-w-2xl mx-auto">Everything you need to manage your desktop remotely, built with cutting-edge technology.</p>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {features.map((feature, index) => (
              <motion.div 
                key={index}
                className="p-8 rounded-3xl bg-foreground/[0.03] dark:bg-white/[0.03] border border-foreground/[0.05] dark:border-white/[0.05]"
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: index * 0.1 }}
              >
                <div className="mb-4 bg-background w-12 h-12 rounded-2xl flex items-center justify-center shadow-sm">
                  {feature.icon}
                </div>
                <h3 className="text-xl font-semibold mb-2">{feature.title}</h3>
                <p className="text-foreground/70 leading-relaxed">{feature.description}</p>
              </motion.div>
            ))}
          </div>
        </div>

        {/* Download Section */}
        <div id="download" className="mt-32 max-w-5xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="p-10 md:p-16 rounded-[3rem] bg-gradient-to-br from-blue-500/10 to-purple-600/10 border border-blue-500/20 text-center">
            <h2 className="text-3xl md:text-5xl font-bold mb-6">Ready to take control?</h2>
            <p className="text-xl text-foreground/70 mb-10 max-w-2xl mx-auto">Download the server for your Windows PC and the client app for your Android device.</p>
            
            <div className="flex flex-col md:flex-row justify-center items-center gap-6">
              <a href="#" className="w-full md:w-64 flex flex-col items-center p-6 rounded-3xl bg-background border border-foreground/10 hover:shadow-xl transition-all group">
                <Download className="w-8 h-8 mb-3 text-blue-500 group-hover:scale-110 transition-transform" />
                <span className="font-semibold text-lg">Windows Setup</span>
                <span className="text-sm text-foreground/50 mt-1">Windows 10 / 11</span>
              </a>
              <a href="#" className="w-full md:w-64 flex flex-col items-center p-6 rounded-3xl bg-background border border-foreground/10 hover:shadow-xl transition-all group">
                <MonitorSmartphone className="w-8 h-8 mb-3 text-green-500 group-hover:scale-110 transition-transform" />
                <span className="font-semibold text-lg">Android APK</span>
                <span className="text-sm text-foreground/50 mt-1">Android 8.0+</span>
              </a>
            </div>
            <p className="mt-8 text-sm text-foreground/50">
              iPhone user? Use the <Link href="/connect" className="text-blue-500 hover:underline">Web Remote</Link> directly in Safari.
            </p>
          </div>
        </div>

        {/* Documentation Section */}
        <div id="docs" className="mt-32 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-3xl font-bold mb-8">Documentation</h2>
          <div className="inline-flex items-center justify-center p-8 rounded-3xl bg-foreground/[0.03] dark:bg-white/[0.03] border border-foreground/[0.05] dark:border-white/[0.05]">
            <BookOpen className="w-10 h-10 mr-6 text-foreground/70" />
            <div className="text-left">
              <h3 className="text-xl font-semibold">How to connect</h3>
              <p className="text-foreground/70 mt-2 max-w-md">
                1. Install the Windows Server and run it.<br />
                2. Scan the QR code using the Android App or enter the IP/PIN in the Web Remote.<br />
                3. You're connected securely.
              </p>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="py-8 border-t border-foreground/10 text-center text-sm text-foreground/50">
        <p>&copy; 2026 kidev. All Rights Reserved. Free for personal use only. No commercial use.</p>
        <p className="mt-2">remote.redlinevis.site</p>
      </footer>
    </div>
  );
}
