import Link from "next/link";
import {
  ArrowLeft,
  Download,
  Monitor,
  ShieldCheck,
  Smartphone,
} from "lucide-react";

const RELEASES_URL =
  "https://github.com/KristianEki11/PC-Remote/releases/latest";

export default function ConnectNoticePage() {
  return (
    <div className="min-h-[100dvh] bg-[#09090b] text-[#f4f4f5] font-sans selection:bg-white/20 flex flex-col">
      <header className="border-b border-white/[0.08] bg-[#09090b]/80 backdrop-blur-md">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between">
          <Link
            href="/"
            className="inline-flex items-center space-x-2 text-sm text-[#a1a1aa] hover:text-white transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            <span>Back to PC Remote</span>
          </Link>
          <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-white/5 border border-white/10 text-[#a1a1aa]">
            v5.0.0
          </span>
        </div>
      </header>

      <main className="flex-1 flex items-center justify-center px-4 sm:px-6 py-16">
        <div className="max-w-xl w-full rounded-2xl border border-white/10 bg-white/[0.02] p-8 sm:p-10">
          <div className="w-12 h-12 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center mb-6">
            <ShieldCheck className="w-6 h-6 text-emerald-400" />
          </div>

          <div className="inline-block text-[11px] font-mono uppercase tracking-wider px-2.5 py-1 rounded bg-white/5 border border-white/10 text-[#a1a1aa] mb-4">
            Native-Only Architecture
          </div>

          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-white">
            Browser Web Remote is Disabled in v5.0
          </h1>

          <p className="mt-4 text-sm sm:text-base text-[#a1a1aa] leading-relaxed">
            PC Remote v5.0 connects exclusively through the{" "}
            <span className="text-white font-medium">
              Native Android Companion App
            </span>{" "}
            over your local Wi-Fi or personal{" "}
            <span className="text-white font-medium">Tailscale</span> network
            with direct SHA-256 TLS certificate pinning. Browser access is
            disabled to eliminate cross-origin exposure and third-party quick
            tunnels.
          </p>

          <div className="mt-8 grid grid-cols-1 sm:grid-cols-2 gap-4">
            <a
              href={RELEASES_URL}
              target="_blank"
              rel="noreferrer"
              className="p-4 rounded-xl bg-white/[0.03] hover:bg-white/[0.07] border border-white/10 hover:border-white/20 transition-all group flex flex-col"
            >
              <div className="flex items-center justify-between mb-3">
                <Monitor className="w-5 h-5 text-blue-400" />
                <Download className="w-4 h-4 text-[#71717a] group-hover:text-white transition-colors" />
              </div>
              <div className="font-semibold text-white text-sm">
                Windows Setup
              </div>
              <div className="text-xs text-[#a1a1aa] mt-1">
                PCRemoteRustSetup.exe
              </div>
            </a>

            <a
              href={RELEASES_URL}
              target="_blank"
              rel="noreferrer"
              className="p-4 rounded-xl bg-white/[0.03] hover:bg-white/[0.07] border border-white/10 hover:border-white/20 transition-all group flex flex-col"
            >
              <div className="flex items-center justify-between mb-3">
                <Smartphone className="w-5 h-5 text-emerald-400" />
                <Download className="w-4 h-4 text-[#71717a] group-hover:text-white transition-colors" />
              </div>
              <div className="font-semibold text-white text-sm">
                Android Companion APK
              </div>
              <div className="text-xs text-[#a1a1aa] mt-1">
                PCRemoteApp.apk
              </div>
            </a>

            <a
              href="https://tailscale.com/download/android"
              target="_blank"
              rel="noreferrer"
              className="sm:col-span-2 p-4 rounded-xl bg-white/[0.03] hover:bg-white/[0.07] border border-white/10 hover:border-white/20 transition-all group flex items-center justify-between"
            >
              <div>
                <div className="font-semibold text-white text-sm">
                  Download Tailscale HP (Android)
                </div>
                <div className="text-xs text-[#a1a1aa] mt-1">
                  Official Tailscale Android APK / Play Store for away-from-home access
                </div>
              </div>
              <Download className="w-4 h-4 text-[#71717a] group-hover:text-white transition-colors shrink-0 ml-4" />
            </a>
          </div>
        </div>
      </main>
    </div>
  );
}
