import { notFound } from "next/navigation";
import type { Metadata } from "next";
import Link from "next/link";
import { QuranHtmlEnhancer } from "@/components/QuranHtmlEnhancer";
import { readSurahHtml, extractTitle, extractMain, SURAH_NAMES_INTERNAL } from "@/lib/html-surah";

export { SURAH_NAMES_INTERNAL };

interface Props {
  params: Promise<{ id: string }>;
}

export async function generateStaticParams() {
  return Object.keys(SURAH_NAMES_INTERNAL).map((id) => ({ id }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params;
  const n = parseInt(id, 10);
  if (isNaN(n) || n < 1 || n > 114) return { title: "Surah Not Found" };
  const html = await readSurahHtml(n);
  const title = html ? extractTitle(html) ?? `Surah ${n}` : `Surah ${n}`;
  return { title };
}

export default async function SurahPage({ params }: Props) {
  const { id } = await params;
  const n = parseInt(id, 10);
  if (isNaN(n) || n < 1 || n > 114) notFound();

  const html = await readSurahHtml(n);
  if (!html) notFound();

  const main = extractMain(html);
  if (!main) notFound();

  return (
    <div className="flex flex-col min-h-screen">
      <SurahNavBar surahNumber={n} />

      <QuranHtmlEnhancer html={main} />

      <BottomNav surahNumber={n} />
    </div>
  );
}

function SurahNavBar({ surahNumber }: { surahNumber: number }) {
  return (
    <header className="sticky top-0 z-30 border-b border-border bg-background/95 backdrop-blur">
      <div className="max-w-5xl mx-auto px-4">
        <div className="flex items-center justify-between h-10">
          <Link href="/" className="text-[11px] text-muted hover:text-foreground font-serif">
            ← Surahs
          </Link>
          <span className="text-[11px] font-mono text-muted tabular-nums">
            {surahNumber} / 114
          </span>
          <div className="flex gap-2">
            {surahNumber > 1 && (
              <Link href={`/surah/${surahNumber - 1}`} className="text-[11px] text-muted hover:text-foreground">
                ‹ Prev
              </Link>
            )}
            {surahNumber < 114 && (
              <Link href={`/surah/${surahNumber + 1}`} className="text-[11px] text-muted hover:text-foreground">
                Next ›
              </Link>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}

function BottomNav({ surahNumber }: { surahNumber: number }) {
  return (
    <footer className="border-t border-border">
      <div className="max-w-3xl mx-auto px-4 py-4 flex justify-between text-[11px] font-serif">
        {surahNumber > 1 ? (
          <Link href={`/surah/${surahNumber - 1}`} className="text-muted hover:text-foreground">
            ← Surah {surahNumber - 1}
          </Link>
        ) : <span />}
        {surahNumber < 114 ? (
          <Link href={`/surah/${surahNumber + 1}`} className="text-muted hover:text-foreground">
            Surah {surahNumber + 1} →
          </Link>
        ) : <span />}
      </div>
    </footer>
  );
}