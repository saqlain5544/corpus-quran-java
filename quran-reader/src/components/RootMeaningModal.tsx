"use client";

import { useEffect, useState } from "react";
import { lockBodyScroll, unlockBodyScroll } from "@/lib/scroll-lock";

interface RootMeaningModalProps {
  root: string;
  visible: boolean;
  onClose: () => void;
}

export function RootMeaningModal({ root, visible, onClose }: RootMeaningModalProps) {
  const [data, setData] = useState<RootEntry | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [activeTab, setActiveTab] = useState<string>("general");

  useEffect(() => {
    if (!visible || !root) return;
    let cancelled = false;
    setLoading(true);
    setError("");

    fetch(`/api/root/${encodeURIComponent(root)}`)
      .then(async (res) => {
        if (!res.ok) {
          const body = await res.json().catch(() => ({}));
          throw new Error(body.error || `Request failed (${res.status})`);
        }
        return res.json();
      })
      .then((d: RootEntry) => {
        if (!cancelled) {
          setData(d);
          setLoading(false);
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Failed to load root meaning");
          setLoading(false);
        }
      });

    return () => { cancelled = true; };
  }, [root, visible]);

  useEffect(() => {
    if (!visible) return;
    lockBodyScroll();
    return () => unlockBodyScroll();
  }, [visible]);

  useEffect(() => {
    if (!visible) return;
    const onKey = (e: KeyboardEvent) => { if (e.key === "Escape") onClose(); };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [visible, onClose]);

  if (!visible) return null;

  const sections: Array<{ id: string; label: string; count?: number; color: string }> = [
    { id: "general", label: "General", color: "emerald" },
    { id: "shades", label: "Shades", count: data?.shadesOfMeaning.length, color: "purple" },
    { id: "morph", label: "Morphology", count: data?.wordByWordShades.length, color: "teal" },
    { id: "sources", label: "Classical", count: data?.classicalSources.length, color: "amber" },
    { id: "hadith", label: "Hadith", count: data?.hadithExamples.length, color: "rose" },
    { id: "idioms", label: "Idioms", count: data?.idiomsCustoms.length, color: "pink" },
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/20 backdrop-blur-sm modal-enter" onClick={onClose}>
      <div
        className="bg-card border border-border rounded-lg shadow-xl w-[95vw] max-w-4xl max-h-[90vh] flex flex-col overflow-hidden font-serif"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex-shrink-0 px-5 py-4 border-b border-border flex items-center justify-between">
          <div className="flex items-baseline gap-3">
            <span className="text-[10px] uppercase tracking-[0.18em] text-muted">Root</span>
            <span className="arabic text-2xl text-accent font-medium">{root}</span>
            {data?.rootArabic && (
              <span className="arabic text-sm text-muted" dir="rtl">
                {data.rootArabic}
              </span>
            )}
          </div>
          <button onClick={onClose} className="w-7 h-7 rounded hover:bg-card-hover flex items-center justify-center text-muted hover:text-foreground transition-colors" aria-label="Close">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Tabs */}
        <div className="flex-shrink-0 px-5 border-b border-border">
          <div className="flex items-center gap-4 overflow-x-auto">
            {sections.map((s) => (
              <button
                key={s.id}
                onClick={() => setActiveTab(s.id)}
                className={`py-2 text-xs uppercase tracking-wider transition-colors border-b whitespace-nowrap ${
                  activeTab === s.id
                    ? "text-foreground border-foreground"
                    : "text-muted border-transparent hover:text-foreground"
                }`}
              >
                {s.label}
                {s.count !== undefined && s.count > 0 && (
                  <span className="ml-1.5 text-[10px] text-muted-light tabular-nums">{s.count}</span>
                )}
              </button>
            ))}
          </div>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-auto p-5">
          {loading && (
            <div className="flex flex-col items-center justify-center py-12 gap-3 text-muted">
              <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
              <span className="text-sm">Loading root details…</span>
            </div>
          )}

          {error && (
            <div className="text-center py-8 text-red-500 text-sm">{error}</div>
          )}

          {data && (
            <>
              {activeTab === "general" && data.generalMeaning && (
                <Section title="General Meaning" color="emerald">
                  <div className="text-sm text-foreground leading-relaxed whitespace-pre-line">
                    {data.generalMeaning}
                  </div>
                </Section>
              )}

              {activeTab === "shades" && data.shadesOfMeaning.length > 0 && (
                <Section title="Shades of Meaning" color="purple" count={data.shadesOfMeaning.length}>
                  <div className="grid gap-3">
                    {data.shadesOfMeaning.map((s, i) => (
                      <ShadeCard
                        key={i}
                        shade={s.shade}
                        reference={s.reference}
                        explanation={s.explanation}
                        arabicWord={s.arabic_word}
                        color="purple"
                      />
                    ))}
                  </div>
                </Section>
              )}

              {activeTab === "morph" && data.wordByWordShades.length > 0 && (
                <Section title="Morphological Shades" color="teal" count={data.wordByWordShades.length}>
                  <div className="grid gap-2">
                    {data.wordByWordShades.map((s, i) => (
                      <div key={i} className="bg-card rounded-lg p-3 border border-border/60 flex items-start gap-3">
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium text-foreground">{s.shade}</p>
                          <p className="text-xs text-muted mt-0.5">{s.explanation}</p>
                        </div>
                        <div className="text-right flex-shrink-0 space-y-0.5">
                          {s.arabic_word && <p className="arabic text-base text-quran-text">{s.arabic_word}</p>}
                          {s.reference && <p className="text-[10px] text-muted font-mono">{s.reference}</p>}
                        </div>
                      </div>
                    ))}
                  </div>
                </Section>
              )}

              {activeTab === "sources" && data.classicalSources.length > 0 && (
                <Section title="Classical Sources" color="amber" count={data.classicalSources.length}>
                  <div className="space-y-3">
                    {data.classicalSources.map((s, i) => (
                      <ClassicalCard key={i} arabic={s.arabic} translation={s.translation} />
                    ))}
                  </div>
                </Section>
              )}

              {activeTab === "hadith" && data.hadithExamples.length > 0 && (
                <Section title="Hadith Examples" color="rose" count={data.hadithExamples.length}>
                  <div className="space-y-3">
                    {data.hadithExamples.map((s, i) => (
                      <ClassicalCard key={i} arabic={s.arabic} translation={s.translation} />
                    ))}
                  </div>
                </Section>
              )}

              {activeTab === "idioms" && data.idiomsCustoms.length > 0 && (
                <Section title="Idioms & Customs" color="pink" count={data.idiomsCustoms.length}>
                  <div className="space-y-3">
                    {data.idiomsCustoms.map((s, i) => (
                      <ClassicalCard key={i} arabic={s.arabic} translation={s.translation} />
                    ))}
                  </div>
                </Section>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}

// ── Helper components ──────────────────────────────────────────

function Section({ title, color, count, children }: { title: string; color: string; count?: number; children: React.ReactNode }) {
  const colors: Record<string, string> = {
    emerald: "bg-emerald-500",
    purple: "bg-purple-500",
    teal: "bg-teal-500",
    amber: "bg-amber-500",
    rose: "bg-rose-500",
    pink: "bg-pink-500",
  };
  return (
    <div>
      <div className="flex items-center gap-2 mb-4">
        <span className={`w-1.5 h-1.5 rounded-full ${colors[color] || "bg-accent"}`} />
        <h4 className="text-sm font-semibold text-foreground">{title}</h4>
        {count !== undefined && (
          <span className="text-[10px] text-muted-light">({count} entries)</span>
        )}
      </div>
      {children}
    </div>
  );
}

function ShadeCard({ shade, reference, explanation, arabicWord, color }: {
  shade: string;
  reference: string;
  explanation: string;
  arabicWord?: string;
  color: string;
}) {
  const colors: Record<string, string> = {
    purple: "bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-400",
  };
  return (
    <div className="bg-card-hover/50 rounded-xl p-4 border border-border/60">
      <div className="flex items-center gap-2 mb-2 flex-wrap">
        <span className={`text-xs font-semibold px-2 py-0.5 rounded ${colors[color] || ""}`}>{shade}</span>
        {reference && <span className="text-[10px] px-1.5 py-0.5 rounded bg-card-hover text-muted font-mono">{reference}</span>}
        {arabicWord && <span className="arabic text-base text-quran-text ml-auto">{arabicWord}</span>}
      </div>
      <p className="text-xs text-muted leading-relaxed">{explanation}</p>
    </div>
  );
}

function ClassicalCard({ arabic, translation }: { arabic: string; translation: string }) {
  return (
    <div className="bg-quran-bg rounded-xl p-4 border border-border/60">
      <p className="arabic text-base text-quran-text leading-loose text-right mb-3" dir="rtl">
        {arabic}
      </p>
      <div className="border-t border-border/50 pt-3 text-xs text-muted leading-relaxed">
        {translation}
      </div>
    </div>
  );
}

interface RootEntry {
  rootArabic: string;
  generalMeaning: string;
  shadesOfMeaning: Array<{
    shade: string;
    reference: string;
    explanation: string;
    arabic_word?: string;
  }>;
  hadithExamples: Array<{ arabic: string; translation: string }>;
  classicalSources: Array<{ arabic: string; translation: string }>;
  idiomsCustoms: Array<{ arabic: string; translation: string }>;
  wordByWordShades: Array<{
    shade: string;
    reference: string;
    explanation: string;
    arabic_word?: string;
  }>;
  pos: string;
}