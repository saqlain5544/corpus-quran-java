import type { Metadata } from "next";
import localFont from "next/font/local";
import { ThemeInitScript, ThemeProvider } from "@/hooks/useTheme";
import { ThemeToggle } from "@/components/ThemeToggle";
import { BookmarkList } from "@/components/BookmarkButton";
import "./globals.css";

/**
 * Hafs Quranic font — loaded via next/font/local for automatic
 * preloading, self-hosting, and zero-CLS application.
 *
 * Exposed as a CSS variable (`--font-hafs`) so the rest of the
 * stylesheet can use the same value via the `--font-arabic` chain.
 */
const hafs = localFont({
  src: "../../public/fonts/hafs.woff2",
  display: "swap",
  variable: "--font-hafs",
  weight: "400",
  style: "normal",
  preload: true,
  fallback: [
    "KFGQPC HAFS Uthmanic Script",
    "KFGQPC Uthmanic Script HAFS",
    "Traditional Arabic",
    "Scheherazade New",
    "serif",
  ],
});

export const metadata: Metadata = {
  title: "Al-Quran Al-Kareem — Read, Study & Reflect",
  description:
    "A full-featured Quran reader with word-by-word morphology analysis, root lookup, and multi-mode search.",
  keywords: [
    "Quran",
    "Koran",
    "Arabic",
    "morphology",
    "tafsir",
    "word analysis",
    "roots",
  ],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" data-scroll-behavior="smooth" className={`${hafs.variable} h-full`} suppressHydrationWarning>
      <head>
        <ThemeInitScript />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <meta name="theme-color" content="#f8f6f1" />
      </head>
      <body className="min-h-full flex flex-col antialiased">
        <ThemeProvider>
          {/* Global top bar — compact editorial masthead */}
          <div className="sticky top-0 z-50 glass-strong">
            <div className="max-w-5xl mx-auto px-4 h-9 flex items-center justify-between">
              <a
                href="/"
                className="flex items-center gap-2 group"
              >
                <span className="arabic text-base text-foreground group-hover:text-accent transition-colors">
                  ٱ
                </span>
                <span className="text-[11px] font-serif font-medium text-foreground group-hover:text-accent transition-colors tracking-wide">
                  Quran Reader
                </span>
              </a>
              <div className="flex items-center gap-0.5">
                <BookmarkList />
                <ThemeToggle />
              </div>
            </div>
          </div>
          <main className="flex-1">{children}</main>
        </ThemeProvider>
      </body>
    </html>
  );
}
