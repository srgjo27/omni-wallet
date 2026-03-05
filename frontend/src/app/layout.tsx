import type { Metadata } from "next";
import "@/styles/globals.css";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "OmniWallet",
  description: "Platform dompet digital yang lengkap",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="id">
      <body className="antialiased">{children}</body>
    </html>
  );
}
