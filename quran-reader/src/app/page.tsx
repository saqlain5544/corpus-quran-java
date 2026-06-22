import Link from "next/link";
import { getSurahList } from "@/lib/data/quran";
import { SearchBar } from "@/components/SearchBar";
import { toArabicNumeral } from "@/lib/arabic";

export default async function HomePage() {
  const surahs = await getSurahList();

  return (
    <div className="flex flex-col min-h-screen">
      {/* Hero — editorial, compact */}
      <header className="border-b border-border">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 py-8 sm:py-10 text-center">
          <p className="text-[10px] uppercase tracking-[0.22em] text-muted mb-3 font-serif">
            <span className="text-muted-light">· </span>
            The Noble Quran
            <span className="text-muted-light"> ·</span>
          </p>
          <h1 className="arabic text-4xl sm:text-5xl text-quran-text leading-tight mb-2 surah-title-frame inline-block">
            ٱلْقُرْآن ٱلْكَرِيم
          </h1>
          <h2 className="text-xl sm:text-2xl font-medium text-foreground mb-1 font-serif">
            Al-Quran Al-Kareem
          </h2>
          <p className="text-sm text-muted font-serif italic max-w-xl mx-auto">
            Read, study &amp; reflect. Word-by-word morphology and root analysis.
          </p>

          {/* Search */}
          <div className="mt-6 max-w-md mx-auto">
            <SearchBar />
          </div>

          {/* Quick stats — editorial hairlines */}
          <div className="mt-6 flex flex-wrap items-center justify-center gap-x-4 gap-y-1.5 text-[11px] text-muted font-serif tabular-nums">
            <span><span className="font-semibold text-foreground">{surahs.length}</span> surahs</span>
            <span className="text-muted-light">·</span>
            <span><span className="font-semibold text-foreground">114</span> total</span>
            <span className="text-muted-light">·</span>
            <span><span className="font-semibold text-foreground">6,236</span> verses</span>
            <span className="text-muted-light">·</span>
            <span><span className="font-semibold text-foreground">77K+</span> words</span>
            <span className="text-muted-light">·</span>
            <span><span className="font-semibold text-foreground">1.6K+</span> roots</span>
          </div>
        </div>
      </header>

      {/* Surah list — editorial table-style */}
      <section className="flex-1 max-w-3xl mx-auto px-4 sm:px-6 py-6 w-full">
        <div className="flex items-center justify-between mb-3 pb-2 border-b border-border-strong">
          <h2 className="text-sm font-semibold text-foreground font-serif tracking-wide">
            All Surahs
          </h2>
          <div className="flex items-center gap-3 text-[10px] text-muted font-serif">
            <span className="flex items-center gap-1.5">
              <span className="w-1.5 h-1.5 rounded-full bg-accent" />
              Meccan
            </span>
            <span className="flex items-center gap-1.5">
              <span className="w-1.5 h-1.5 rounded-full bg-sky-500" />
              Medinan
            </span>
          </div>
        </div>

        <ul className="divide-y divide-border">
          {surahs.map((surah) => (
            <li key={surah.number}>
              <Link
                href={`/surah/${surah.number}`}
                className="group flex items-center gap-4 py-3 hover:bg-verse-hover transition-colors"
              >
                {/* Surah number — small, serif */}
                <span className="verse-number w-8 flex-shrink-0 text-center tabular-nums">
                  {surah.number}
                </span>

                {/* Arabic name */}
                <span className="arabic text-xl text-quran-text flex-1 min-w-0 truncate" dir="rtl">
                  {surah.name}
                </span>

                {/* English name */}
                <span className="flex-1 min-w-0">
                  <span className="block text-sm font-medium text-foreground group-hover:text-accent transition-colors truncate font-serif">
                    {surah.englishName}
                  </span>
                  <span className="block text-[11px] text-muted truncate font-serif italic">
                    {surah.englishNameTranslation}
                  </span>
                </span>

                {/* Meta */}
                <span className="flex items-center gap-2 flex-shrink-0 text-[10px] text-muted tabular-nums font-serif">
                  <span
                    className={`px-1.5 py-0.5 ${
                      surah.revelationType === "Meccan"
                        ? "text-accent"
                        : "text-sky-600 dark:text-sky-400"
                    }`}
                  >
                    {surah.revelationType === "Meccan" ? "M" : "Md"}
                  </span>
                  <span className="w-7 text-right">{surah.numberOfAyahs}</span>
                </span>
              </Link>
            </li>
          ))}
        </ul>
      </section>

      {/* Footer — compact editorial */}
      <footer className="mt-8 border-t border-border">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 py-4 text-center text-[11px] text-muted font-serif">
          <p className="mb-1">Built with care for Quran learners worldwide.</p>
          <p>
            Text from{" "}
            <a href="https://tanzil.net" className="text-accent hover:underline" target="_blank" rel="noopener noreferrer">
              Tanzil
            </a>
            {" · "}
            Morphology from{" "}
            <a href="https://corpus.quran.com" className="text-accent hover:underline" target="_blank" rel="noopener noreferrer">
              Quranic Arabic Corpus
            </a>
          </p>
        </div>
      </footer>
    </div>
  );
}
