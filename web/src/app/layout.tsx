import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "PC Remote — Windows Remote Control Suite",
  description: "Secure, ultra-low latency Windows PC remote controller with desktop screen mirroring, Core Audio mixing, and zero-config Cloudflare WAN tunnels.",
  icons: {
    icon: "/favicon.ico",
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark h-full antialiased selection:bg-blue-500/30 selection:text-white">
      <body className="min-h-full flex flex-col bg-[#050508] text-[#f5f5f7]">{children}</body>
    </html>
  );
}
