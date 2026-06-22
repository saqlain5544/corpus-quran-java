import { NextRequest } from "next/server";
import { PrismaClient } from "@/generated/prisma";

const prisma = new PrismaClient();

interface RouteParams {
  params: Promise<{ id: string }>;
}

/**
 * GET /api/word/[id]
 * Returns the full morphology of a single word (all segments + word-level
 * metadata). Used by WordAnalysisModal when the user clicks a word.
 *
 * The word id comes from the `data-word-id` attribute on the SSR-rendered
 * word span — no re-querying required to navigate from a word click to its
 * detailed analysis.
 */
export async function GET(_request: NextRequest, { params }: RouteParams) {
  const { id } = await params;
  const wordId = parseInt(id, 10);
  if (isNaN(wordId) || wordId <= 0) {
    return Response.json({ error: "Invalid word id" }, { status: 400 });
  }

  const word = await prisma.word.findUnique({
    where: { id: wordId },
    include: {
      segments: { orderBy: { number: "asc" } },
    },
  });

  if (!word) {
    return Response.json({ error: `Word ${wordId} not found` }, { status: 404 });
  }

  return Response.json({
    id: word.id,
    wordNumber: word.number,
    surahId: word.surahId,
    verseId: word.verseId,
    token: word.token,
    withoutDiacritics: word.withoutDiacritics,
    translation: word.translation,
    punctuationMark: word.punctuationMark,
    segments: word.segments.map((s: typeof word.segments[number]) => ({
      number: s.number,
      text: s.text,
      partOfSpeech: s.partOfSpeech,
      morphType: s.morphType,
      lemma: s.lemma,
      root: s.root,
      gender: s.gender,
      case: s.caseField,
      syntacticRole: s.syntacticRole,
      gloss: s.gloss,
      possessiveConstruct: s.possessiveConstruct,
      caseMoodMarker: s.caseMoodMarker,
      invariableDeclinable: s.invariableDeclinable,
      phrase: s.phrase,
      phrasalFunction: s.phrasalFunction,
      punctuationMark: s.punctuationMark,
    })),
  });
}