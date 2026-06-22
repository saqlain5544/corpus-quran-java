import { PrismaClient } from "@/generated/prisma";

const prisma = new PrismaClient();

/** Get root glosses for quick tooltip lookups */
export async function getRootGlosses(): Promise<Record<string, string>> {
  const meanings = await prisma.rootMeaning.findMany({
    select: { root: true, generalMeaning: true },
  });
  const map: Record<string, string> = {};
  for (const m of meanings) {
    map[m.root] = m.generalMeaning.split(".")[0] || m.generalMeaning.substring(0, 80);
  }
  return map;
}

/** Get full root meaning entry.
 *  Accepts either an Arabic-script root (e.g. "قول") or a Buckwalter
 *  transliteration (e.g. "qwl"). Tries Arabic first, then Buckwalter. */
export async function getRootMeaning(root: string) {
  // Try Arabic (rootArabic) first — this is what the UI sends after
  // buckwalterToArabic conversion. Fall back to Buckwalter for any
  // callers still passing transliteration directly.
  const isArabic = /^[\u0600-\u06FF]+$/.test(root);
  let rm = isArabic
    ? await prisma.rootMeaning.findFirst({ where: { rootArabic: root } })
    : await prisma.rootMeaning.findUnique({ where: { root } });

  // Last-resort fallback: try the other field
  if (!rm) {
    rm = isArabic
      ? await prisma.rootMeaning.findUnique({ where: { root } })
      : await prisma.rootMeaning.findFirst({ where: { rootArabic: root } });
  }

  if (!rm) return null;
  return {
    root: rm.root,
    rootArabic: rm.rootArabic,
    generalMeaning: rm.generalMeaning,
    shadesOfMeaning: safeJson(rm.shadesOfMeaning),
    hadithExamples: safeJson(rm.hadithExamples),
    classicalSources: safeJson(rm.classicalSources),
    idiomsCustoms: safeJson(rm.idiomsCustoms),
    wordByWordShades: safeJson(rm.wordByWordShades),
    pos: rm.pos,
  };
}

function safeJson(s: string): unknown[] {
  try { return JSON.parse(s); } catch { return []; }
}
