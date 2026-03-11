import type { Metadata } from "next";

import "./globals.css";

export const metadata: Metadata = {
  title: "Write Me",
  description: "Self-hosted 자기소개서 보조 서비스"
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="ko">
      <body>{children}</body>
    </html>
  );
}
