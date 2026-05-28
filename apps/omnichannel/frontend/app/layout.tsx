import type { Metadata } from "next";
import Script from "next/script";
import "./globals.css";

export const metadata: Metadata = {
  title: "Omnichannel Notifications",
  description: "Operational dashboard for omnichannel notification delivery",
};

const themeInitScript = `
(() => {
  try {
    const mode = localStorage.getItem('console:theme-mode');
    const root = document.documentElement;
    if (mode === 'light' || mode === 'dark') {
      root.setAttribute('data-theme', mode);
      return;
    }
    root.removeAttribute('data-theme');
  } catch {}
})();
`;

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="antialiased">
        <Script id="theme-init" strategy="beforeInteractive">
          {themeInitScript}
        </Script>
        {children}
      </body>
    </html>
  );
}
