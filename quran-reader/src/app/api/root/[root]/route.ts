import { NextRequest } from "next/server";
import { getRootMeaning } from "@/lib/data/roots";

interface RouteParams {
  params: Promise<{ root: string }>;
}

/**
 * GET /api/root/[root]
 * Returns the full root meaning entry (general meaning, shades of meaning,
 * hadith examples, classical sources, idioms, word-by-word shades).
 */
export async function GET(_request: NextRequest, { params }: RouteParams) {
  const { root } = await params;

  // Validate root format: should be 1-4 Arabic letters
  if (!root || root.length > 8 || !/^[\u0600-\u06FF]+$/.test(decodeURIComponent(root))) {
    return Response.json({ error: "Invalid root" }, { status: 400 });
  }

  const decoded = decodeURIComponent(root);
  const data = await getRootMeaning(decoded);

  if (!data) {
    return Response.json(
      { error: `No meaning found for root "${decoded}"` },
      { status: 404 }
    );
  }

  return Response.json(data);
}
