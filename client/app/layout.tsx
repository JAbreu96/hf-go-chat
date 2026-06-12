import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "HF Go Chat",
  description: "RAG + web-search chat powered by Hugging Face and Go",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="h-full antialiased">
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
