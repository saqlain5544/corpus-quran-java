import { NextRequest } from "next/server";
import { getSurahList, getAllVersesForSearch, getSearchMorphology } from "@/lib/data/quran";
import { searchQuranText, searchQuranByRoot, searchQuranByGloss } from "@/lib/search";
import { SearchQuerySchema } from "@/lib/data/types";

export async function GET(request: NextRequest) {
  const { searchParams } = request.nextUrl;
  const query = searchParams.get("q") || "";
  const type = (searchParams.get("type") as "text" | "root" | "gloss") || "text";
  const limit = parseInt(searchParams.get("limit") || "20", 10);

  const parseResult = SearchQuerySchema.safeParse({ q: query, type, limit });
  if (!parseResult.success) {
    return Response.json(
      { error: "Invalid query parameters", details: parseResult.error.issues },
      { status: 400 }
    );
  }

  try {
    const [surahList, allVerses, morphData] = await Promise.all([
      getSurahList(),
      getAllVersesForSearch(),
      getSearchMorphology(),
    ]);

    const surahNames: Record<number, string> = {};
    for (const s of surahList) {
      surahNames[s.number] = s.englishName;
    }

    let results;
    if (type === "root") {
      results = searchQuranByRoot(query, morphData, allVerses, surahNames, limit);
    } else if (type === "gloss") {
      results = searchQuranByGloss(query, morphData, allVerses, surahNames, limit);
    } else {
      results = searchQuranText(query, allVerses, surahNames, limit);
    }

    return Response.json({ query, type, results, total: results.length });
  } catch (error) {
    console.error("Search error:", error);
    return Response.json(
      { error: "Search failed. Please try again." },
      { status: 500 }
    );
  }
}
