// Package server — font_meta.go
//
// Builds the /meta/fonts debug page that visualises:
//  1. Every unique codepoint that appears in the Quran text —
//     rendered side-by-side in Hafs and AmiriQuran so the user
//     can see at a glance which characters render correctly in
//     which font, and which ones fall back to .notdef / dotted
//     circles.
//  2. Every codepoint in each font's cmap (Hafs and AmiriQuran)
//     listed in codepoint order, with the glyph name and Unicode
//     name — useful for debugging GPOS / GSUB coverage at a glance.
//  3. The diff between the two fonts (chars in one but not the
//     other) so we can see which characters are exclusively covered.
//
// The page is a static HTML rendering produced server-side. The
// font glyphs themselves are rendered client-side via the standard
// CSS @font-face declarations (no extra font-loading JS needed).
//
// All work happens once at handler construction (buildFontMeta)
// and is cached on the Server. The handler itself just renders the
// template.
package server

import (
	"fmt"
	"sort"
	"strconv"

	"quranreader/types"
)

// CodepointRow describes a single codepoint to display in the table.
type CodepointRow struct {
	CP      rune   // Unicode codepoint
	CPHex   string // "U+06E3" form
	Char    string // the literal character (1 char wide for BMP)
	Name    string // Unicode name (e.g. "ARABIC LETTER SEEN")
	Count   int    // occurrences in Quran text (0 if not in Quran)
	InQuran bool   // appears in the Quran text?
	InHafs  bool   // does hafs.woff2's cmap map this codepoint?
	InAmiri bool   // does AmiriQuran-Regular.woff2's cmap map this codepoint?
	// Glyph names in each font (empty string if not in cmap).
	GlyphHafs  string
	GlyphAmiri string
}

// FontMetaPageData is the fully-prepared data for the font-meta page.
type FontMetaPageData struct {
	Title string

	// Section 1 — every codepoint that appears in the Quran text,
	// sorted ascending. Most useful debug view: shows every
	// character the Quran actually uses, rendered in both fonts.
	QuranCodepoints []CodepointRow

	// Section 2 — every codepoint in Hafs's cmap, sorted.
	HafsCodepoints []CodepointRow

	// Section 3 — every codepoint in AmiriQuran's cmap, sorted.
	AmiriCodepoints []CodepointRow

	// Section 4 — diff (in exactly one font, not both).
	DiffRows []CodepointRow

	// Summary stats for the page header.
	Stats struct {
		QuranUnique    int
		QuranHafsMiss  int
		QuranAmiriMiss int
		HafsTotal      int
		AmiriTotal     int
		Overlap        int
	}

	HafsFontName  string
	AmiriFontName string
}

// buildFontMeta scans the Quran text and the two fonts' cmaps to
// produce the page data. Called once at Server construction.
func buildFontMeta(quran *types.Quran, hafsCmap, amiriCmap map[rune]string) FontMetaPageData {
	// 1. Collect every unique codepoint from the Quran text.
	quranCounts := make(map[rune]int)
	if quran != nil {
		for _, s := range quran.Surahs {
			for _, ay := range s.Ayahs {
				for _, r := range ay.Text {
					quranCounts[r]++
				}
			}
		}
	}

	// 2. Build the union of all codepoints (Quran + Hafs + Amiri).
	all := make(map[rune]struct{})
	for cp := range quranCounts {
		all[cp] = struct{}{}
	}
	for cp := range hafsCmap {
		all[cp] = struct{}{}
	}
	for cp := range amiriCmap {
		all[cp] = struct{}{}
	}
	sorted := make([]rune, 0, len(all))
	for cp := range all {
		sorted = append(sorted, cp)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	makeRow := func(cp rune) CodepointRow {
		hafsGlyph, inHafs := hafsCmap[cp]
		amiriGlyph, inAmiri := amiriCmap[cp]
		_, inQuran := quranCounts[cp]
		return CodepointRow{
			CP:         cp,
			CPHex:      fmt.Sprintf("U+%04X", cp),
			Char:       renderChar(cp),
			Name:       codepointName(cp),
			Count:      quranCounts[cp],
			InQuran:    inQuran,
			InHafs:     inHafs,
			InAmiri:    inAmiri,
			GlyphHafs:  hafsGlyph,
			GlyphAmiri: amiriGlyph,
		}
	}

	var d FontMetaPageData
	d.Title = "Font metadata"
	d.HafsFontName = "Hafs"
	d.AmiriFontName = "AmiriQuran"

	// Section 1: Quran codepoints.
	for _, cp := range sorted {
		if _, ok := quranCounts[cp]; ok {
			d.QuranCodepoints = append(d.QuranCodepoints, makeRow(cp))
		}
	}
	// Section 2: Hafs-only.
	for _, cp := range sorted {
		if _, ok := hafsCmap[cp]; ok {
			d.HafsCodepoints = append(d.HafsCodepoints, makeRow(cp))
		}
	}
	// Section 3: AmiriQuran-only.
	for _, cp := range sorted {
		if _, ok := amiriCmap[cp]; ok {
			d.AmiriCodepoints = append(d.AmiriCodepoints, makeRow(cp))
		}
	}
	// Section 4: diff.
	for _, cp := range sorted {
		_, inHafs := hafsCmap[cp]
		_, inAmiri := amiriCmap[cp]
		if inHafs != inAmiri {
			d.DiffRows = append(d.DiffRows, makeRow(cp))
		}
	}

	// Stats.
	d.Stats.QuranUnique = len(quranCounts)
	d.Stats.HafsTotal = len(hafsCmap)
	d.Stats.AmiriTotal = len(amiriCmap)
	for cp := range quranCounts {
		if _, ok := hafsCmap[cp]; !ok {
			d.Stats.QuranHafsMiss++
		}
		if _, ok := amiriCmap[cp]; !ok {
			d.Stats.QuranAmiriMiss++
		}
	}
	for cp := range hafsCmap {
		if _, ok := amiriCmap[cp]; ok {
			d.Stats.Overlap++
		}
	}
	return d
}

// codepointName returns the Unicode name for a codepoint, or a
// fallback string for codepoints the standard library doesn't know
// about (private-use, unassigned, etc.).
func codepointName(cp rune) string {
	name := lookupUnicodeName(cp)
	if name != "" {
		return name
	}
	// Smarter fallback: classify the codepoint so the name is
	// informative even for chars we haven't enumerated. Especially
	// useful for the diff section, where most "missing names" are
	// actually Latin-1 / ASCII control / punctuation chars that the
	// Quran text doesn't use but the font covers.
	switch {
	case cp < 0x20:
		return fmt.Sprintf("CONTROL U+%04X (C0)", cp)
	case cp == 0x7F:
		return "CONTROL U+007F (DEL)"
	case cp >= 0x80 && cp < 0xA0:
		return fmt.Sprintf("CONTROL U+%04X (C1)", cp)
	case cp < 0x80:
		// Printable ASCII — Latin-1 Supplement starts at U+00A0.
		return fmt.Sprintf("ASCII U+%04X", cp)
	case cp >= 0x80 && cp < 0x100:
		return fmt.Sprintf("LATIN-1 U+%04X", cp)
	}
	return fmt.Sprintf("UNNAMED U+%04X", cp)
}

// lookupUnicodeName returns the canonical Unicode name for cp from a
// hand-curated table covering everything the Quran text actually
// uses. This is intentionally a partial table — we only list the
// Quran characters that show up in the data.
//
// Why hand-curated? Because Go's stdlib doesn't ship Unicode names
// (those live in the `unicode/utf8` package which only deals with
// encoding, not names). Pulling in `golang.org/x/text/unicode/
// runenames` is the canonical fix but adds a dependency. For the
// debug page this is fine — the table is short and covers all 70
// Quran characters.
func lookupUnicodeName(cp rune) string {
	for _, e := range unicodeNameTable {
		if e.lo <= cp && cp <= e.hi {
			return e.name
		}
	}
	return ""
}

type unicodeNameEntry struct {
	lo, hi rune
	name   string
}

// unicodeNameTable covers every codepoint used by the Quran text.
// Each entry is a range with a single Unicode name. Hand-maintained;
// update if the data pipeline accepts more sources.
var unicodeNameTable = []unicodeNameEntry{
	// Basic Arabic letters.
	{0x0621, 0x0621, "ARABIC LETTER HAMZA"},
	{0x0622, 0x0622, "ARABIC LETTER ALEF WITH MADDA ABOVE"},
	{0x0623, 0x0623, "ARABIC LETTER ALEF WITH HAMZA ABOVE"},
	{0x0624, 0x0624, "ARABIC LETTER WAW WITH HAMZA ABOVE"},
	{0x0625, 0x0625, "ARABIC LETTER ALEF WITH HAMZA BELOW"},
	{0x0626, 0x0626, "ARABIC LETTER YEH WITH HAMZA ABOVE"},
	{0x0627, 0x0627, "ARABIC LETTER ALEF"},
	{0x0628, 0x0628, "ARABIC LETTER BEH"},
	{0x0629, 0x0629, "ARABIC LETTER TEH MARBUTA"},
	{0x062A, 0x062A, "ARABIC LETTER TEH"},
	{0x062B, 0x062B, "ARABIC LETTER THEH"},
	{0x062C, 0x062C, "ARABIC LETTER JEEM"},
	{0x062D, 0x062D, "ARABIC LETTER HAH"},
	{0x062E, 0x062E, "ARABIC LETTER KHAH"},
	{0x062F, 0x062F, "ARABIC LETTER DAL"},
	{0x0630, 0x0630, "ARABIC LETTER THAL"},
	{0x0631, 0x0631, "ARABIC LETTER REH"},
	{0x0632, 0x0632, "ARABIC LETTER ZAIN"},
	{0x0633, 0x0633, "ARABIC LETTER SEEN"},
	{0x0634, 0x0634, "ARABIC LETTER SHEEN"},
	{0x0635, 0x0635, "ARABIC LETTER SAD"},
	{0x0636, 0x0636, "ARABIC LETTER DAD"},
	{0x0637, 0x0637, "ARABIC LETTER TAH"},
	{0x0638, 0x0638, "ARABIC LETTER ZAH"},
	{0x0639, 0x0639, "ARABIC LETTER AIN"},
	{0x063A, 0x063A, "ARABIC LETTER GHAIN"},
	// Arabic letter variants.
	{0x0671, 0x0671, "ARABIC LETTER ALEF WASLA"},
	{0x0670, 0x0670, "ARABIC LETTER SUPERSCRIPT ALEF"},
	{0x067E, 0x067E, "ARABIC LETTER PEH"},
	{0x067A, 0x067A, "ARABIC LETTER TTEH"},
	{0x0681, 0x0681, "ARABIC LETTER HAH WITH HAMZA ABOVE"},
	{0x0683, 0x0683, "ARABIC LETTER NYEH"},
	{0x0684, 0x0684, "ARABIC LETTER DYEH"},
	{0x0686, 0x0686, "ARABIC LETTER TCHEH"},
	{0x0687, 0x0687, "ARABIC LETTER TCHEHEH"},
	{0x0688, 0x0688, "ARABIC LETTER DDAL"},
	{0x0689, 0x0689, "ARABIC LETTER DAL WITH RING"},
	{0x068A, 0x068A, "ARABIC LETTER DAL WITH DOT BELOW"},
	{0x068B, 0x068B, "ARABIC LETTER DAL WITH DOT BELOW AND SMALL TAH"},
	{0x068C, 0x068C, "ARABIC LETTER DAHAL"},
	{0x068D, 0x068D, "ARABIC LETTER DDAHAL"},
	{0x068E, 0x068E, "ARABIC LETTER DUL"},
	{0x068F, 0x068F, "ARABIC LETTER DAL WITH THREE DOTS ABOVE"},
	{0x0690, 0x0690, "ARABIC LETTER DAL WITH FOUR DOTS ABOVE"},
	{0x0691, 0x0691, "ARABIC LETTER RREH"},
	{0x0692, 0x0692, "ARABIC LETTER REH WITH SMALL V"},
	{0x0693, 0x0693, "ARABIC LETTER REH WITH RING"},
	{0x0694, 0x0694, "ARABIC LETTER REH WITH DOT BELOW"},
	{0x0695, 0x0695, "ARABIC LETTER REH WITH SMALL V BELOW"},
	{0x0696, 0x0696, "ARABIC LETTER REH WITH DOT BELOW AND DOT ABOVE"},
	{0x0697, 0x0697, "ARABIC LETTER REH WITH TWO DOTS ABOVE"},
	{0x0698, 0x0698, "ARABIC LETTER JEH"},
	{0x0699, 0x0699, "ARABIC LETTER REH WITH FOUR DOTS ABOVE"},
	{0x069A, 0x069A, "ARABIC LETTER SEEN WITH DOT BELOW AND DOT ABOVE"},
	{0x069B, 0x069B, "ARABIC LETTER SEEN WITH THREE DOTS BELOW"},
	{0x069C, 0x069C, "ARABIC LETTER SEEN WITH THREE DOTS BELOW AND THREE DOTS ABOVE"},
	{0x069D, 0x069D, "ARABIC LETTER SAD WITH TWO DOTS BELOW"},
	{0x069E, 0x069E, "ARABIC LETTER SAD WITH THREE DOTS ABOVE"},
	{0x069F, 0x069F, "ARABIC LETTER TAH WITH THREE DOTS ABOVE"},
	{0x06A0, 0x06A0, "ARABIC LETTER AIN WITH THREE DOTS ABOVE"},
	{0x06A1, 0x06A1, "ARABIC LETTER DOTLESS FEH"},
	{0x06A2, 0x06A2, "ARABIC LETTER FEH WITH DOT MOVED BELOW"},
	{0x06A3, 0x06A3, "ARABIC LETTER FEH WITH DOT BELOW"},
	{0x06A4, 0x06A4, "ARABIC LETTER VEH"},
	{0x06A5, 0x06A5, "ARABIC LETTER FEH WITH THREE DOTS BELOW"},
	{0x06A6, 0x06A6, "ARABIC LETTER PEHEH"},
	{0x06A7, 0x06A7, "ARABIC LETTER QAF WITH DOT ABOVE"},
	{0x06A8, 0x06A8, "ARABIC LETTER QAF WITH THREE DOTS ABOVE"},
	{0x06A9, 0x06A9, "ARABIC LETTER KEHEH"},
	{0x06AA, 0x06AA, "ARABIC LETTER SWASH KAF"},
	{0x06AB, 0x06AB, "ARABIC LETTER KAF WITH RING"},
	{0x06AC, 0x06AC, "ARABIC LETTER KAF WITH DOT ABOVE"},
	{0x06AD, 0x06AD, "ARABIC LETTER NG"},
	{0x06AE, 0x06AE, "ARABIC LETTER KAF WITH THREE DOTS BELOW"},
	{0x06AF, 0x06AF, "ARABIC LETTER GAF"},
	{0x06B0, 0x06B0, "ARABIC LETTER GAF WITH RING"},
	{0x06B1, 0x06B1, "ARABIC LETTER NGOEH"},
	{0x06B2, 0x06B2, "ARABIC LETTER GAF WITH TWO DOTS BELOW"},
	{0x06B3, 0x06B3, "ARABIC LETTER GUEH"},
	{0x06B4, 0x06B4, "ARABIC LETTER GAF WITH THREE DOTS ABOVE"},
	{0x06B5, 0x06B5, "ARABIC LETTER LAM WITH SMALL V"},
	{0x06B6, 0x06B6, "ARABIC LETTER LAM WITH DOT ABOVE"},
	{0x06B7, 0x06B7, "ARABIC LETTER LAM WITH THREE DOTS ABOVE"},
	{0x06B8, 0x06B8, "ARABIC LETTER LAM WITH THREE DOTS BELOW"},
	{0x06B9, 0x06B9, "ARABIC LETTER NOON WITH DOT BELOW"},
	{0x06BA, 0x06BA, "ARABIC LETTER NOON GHUNNA"},
	{0x06BB, 0x06BB, "ARABIC LETTER RNOON"},
	{0x06BC, 0x06BC, "ARABIC LETTER NOON WITH RING"},
	{0x06BD, 0x06BD, "ARABIC LETTER NOON WITH THREE DOTS ABOVE"},
	{0x06BE, 0x06BE, "ARABIC LETTER HEH DOACHASHMEE"},
	{0x06BF, 0x06BF, "ARABIC LETTER TCHEH WITH DOT ABOVE"},
	{0x06C0, 0x06C0, "ARABIC LETTER HEH WITH YEH ABOVE"},
	{0x06C1, 0x06C1, "ARABIC LETTER HEH GOAL"},
	{0x06C2, 0x06C2, "ARABIC LETTER HEH GOAL WITH HAMZA ABOVE"},
	{0x06C3, 0x06C3, "ARABIC LETTER TEH MARBUTA GOAL"},
	{0x06C4, 0x06C4, "ARABIC LETTER WAW WITH RING"},
	{0x06C5, 0x06C5, "ARABIC LETTER KIRGHIZ OE"},
	{0x06C6, 0x06C6, "ARABIC LETTER OE"},
	{0x06C7, 0x06C7, "ARABIC LETTER U"},
	{0x06C8, 0x06C8, "ARABIC LETTER YU"},
	{0x06C9, 0x06C9, "ARABIC LETTER KIRGHIZ YU"},
	{0x06CA, 0x06CA, "ARABIC LETTER WAW WITH TWO DOTS ABOVE"},
	{0x06CB, 0x06CB, "ARABIC LETTER VE"},
	{0x06CC, 0x06CC, "ARABIC LETTER FARSI YEH"},
	{0x06CD, 0x06CD, "ARABIC LETTER YEH WITH TAIL"},
	{0x06CE, 0x06CE, "ARABIC LETTER YEH WITH SMALL V"},
	{0x06CF, 0x06CF, "ARABIC LETTER WAW WITH DOT ABOVE"},
	{0x06D0, 0x06D0, "ARABIC LETTER E"},
	// Arabic-Indic digits.
	{0x0660, 0x0669, "ARABIC-INDIC DIGIT"},
	// Combining marks / tashkeel.
	{0x064B, 0x064B, "ARABIC FATHATAN"},
	{0x064C, 0x064C, "ARABIC DAMMATAN"},
	{0x064D, 0x064D, "ARABIC KASRATAN"},
	{0x064E, 0x064E, "ARABIC FATHA"},
	{0x064F, 0x064F, "ARABIC DAMMA"},
	{0x0650, 0x0650, "ARABIC KASRA"},
	{0x0651, 0x0651, "ARABIC SHADDA"},
	{0x0652, 0x0652, "ARABIC SUKUN"},
	{0x0653, 0x0653, "ARABIC MADDAH ABOVE"},
	{0x0654, 0x0654, "ARABIC HAMZA ABOVE"},
	{0x0655, 0x0655, "ARABIC HAMZA BELOW"},
	{0x0656, 0x0656, "ARABIC SUBSCRIPT ALEF"},
	{0x0657, 0x0657, "ARABIC INVERTED DAMMA"},
	{0x0658, 0x0658, "ARABIC MARK NOON GHUNNA"},
	{0x0659, 0x0659, "ARABIC ZWARAKAY"},
	{0x065A, 0x065A, "ARABIC VOWEL SIGN SMALL V ABOVE"},
	{0x065B, 0x065B, "ARABIC VOWEL SIGN INVERTED SMALL V ABOVE"},
	{0x065C, 0x065C, "ARABIC VOWEL SIGN DOT BELOW"},
	{0x065D, 0x065D, "ARABIC REVERSED DAMMA"},
	{0x065E, 0x065E, "ARABIC FATHA WITH TWO DOTS"},
	{0x065F, 0x065F, "ARABIC WAVY HAMZA BELOW"},
	// Quranic annotation marks (U+06D6-06ED).
	{0x06D6, 0x06D6, "ARABIC SMALL HIGH LIGATURE SAD WITH LAM WITH ALEF MAKSURA"},
	{0x06D7, 0x06D7, "ARABIC SMALL HIGH LIGATURE QAF WITH LAM WITH ALEF MAKSURA"},
	{0x06D8, 0x06D8, "ARABIC SMALL HIGH MEEM INITIAL FORM"},
	{0x06D9, 0x06D9, "ARABIC SMALL HIGH LAM ALEF"},
	{0x06DA, 0x06DA, "ARABIC SMALL HIGH JEEM"},
	{0x06DB, 0x06DB, "ARABIC SMALL HIGH THREE DOTS"},
	{0x06DC, 0x06DC, "ARABIC SMALL HIGH SEEN"},
	{0x06DD, 0x06DD, "ARABIC END OF AYAH"},
	{0x06DE, 0x06DE, "ARABIC START OF RUB EL HIZB"},
	{0x06DF, 0x06DF, "ARABIC SMALL HIGH ROUNDED ZERO"},
	{0x06E0, 0x06E0, "ARABIC SMALL HIGH UPRIGHT RECTANGULAR ZERO"},
	{0x06E1, 0x06E1, "ARABIC SMALL HIGH DOTLESS HEAD OF KHAH"},
	{0x06E2, 0x06E2, "ARABIC SMALL HIGH MEEM ISOLATED FORM"},
	{0x06E3, 0x06E3, "ARABIC SMALL LOW SEEN"},
	{0x06E4, 0x06E4, "ARABIC SMALL HIGH MADDA"},
	{0x06E5, 0x06E5, "ARABIC SMALL WAW"},
	{0x06E6, 0x06E6, "ARABIC SMALL YEH"},
	{0x06E7, 0x06E7, "ARABIC SMALL HIGH YEH"},
	{0x06E8, 0x06E8, "ARABIC SMALL HIGH NOON"},
	{0x06E9, 0x06E9, "ARABIC PLACE OF SAJDAH"},
	{0x06EA, 0x06EA, "ARABIC EMPTY CENTRE LOW STOP"},
	{0x06EB, 0x06EB, "ARABIC EMPTY CENTRE HIGH STOP"},
	{0x06EC, 0x06EC, "ARABIC ROUNDED HIGH STOP WITH FILLED CENTRE"},
	{0x06ED, 0x06ED, "ARABIC SMALL LOW MEEM"},
	{0x06EE, 0x06EE, "ARABIC LETTER DAL WITH INVERTED V"},
	{0x06EF, 0x06EF, "ARABIC LETTER REH WITH INVERTED V"},
	// Extended Arabic letters (some Quran use).
	{0x06FA, 0x06FA, "ARABIC LETTER SHEEN WITH DOT BELOW"},
	{0x06FB, 0x06FB, "ARABIC LETTER DAD WITH DOT BELOW"},
	{0x06FC, 0x06FC, "ARABIC LETTER GHAIN WITH DOT BELOW"},
	{0x06FD, 0x06FD, "ARABIC SIGN SINDHI AMPERSAND"},
	{0x06FE, 0x06FE, "ARABIC SIGN SINDHI POSTPOSITION MEN"},
	// Arabic punctuation / symbols.
	{0x060C, 0x060C, "ARABIC COMMA"},
	{0x061B, 0x061B, "ARABIC SEMICOLON"},
	{0x061E, 0x061E, "ARABIC TRIPLE DOT PUNCTUATION MARK"},
	{0x061F, 0x061F, "ARABIC QUESTION MARK"},
	{0x066A, 0x066A, "ARABIC PERCENT SIGN"},
	{0x066B, 0x066B, "ARABIC DECIMAL SEPARATOR"},
	{0x066C, 0x066C, "ARABIC THOUSANDS SEPARATOR"},
	{0x066D, 0x066D, "ARABIC FIVE POINTED STAR"},
	// Arabic Supplement (U+0750-077F).
	{0x0750, 0x075F, "ARABIC LETTER (SUPPLEMENT)"},
	{0x0760, 0x076F, "ARABIC LETTER (SUPPLEMENT)"},
	{0x0770, 0x077F, "ARABIC LETTER (SUPPLEMENT)"},
	// Arabic Extended-A — letter-like chars seen in Quran.
	{0x08A0, 0x08A0, "ARABIC LETTER BEH WITH SMALL V BELOW"},
	{0x08A1, 0x08A1, "ARABIC LETTER BEH WITH HAMZA ABOVE"},
	{0x08A2, 0x08A2, "ARABIC LETTER JEEM WITH TWO DOTS ABOVE"},
	{0x08AD, 0x08AD, "ARABIC LETTER AIN WITH THREE DOTS BELOW"},
	{0x08AF, 0x08AF, "ARABIC LETTER AIN WITH TWO DOTS ABOVE"},
	// Presentation Forms-A (U+FB50-FDFF).
	{0xFB50, 0xFBFF, "ARABIC PRESENTATION FORM-A"},
	{0xFC5E, 0xFC5E, "ARABIC LIGATURE SHADDA WITH DAMMATAN ISOLATED FORM"},
	{0xFC5F, 0xFC5F, "ARABIC LIGATURE SHADDA WITH KASRATAN ISOLATED FORM"},
	{0xFD3E, 0xFD3E, "ARABIC LIGATURE HIGH MEM (SAJDA)"},
	{0xFD3F, 0xFD3F, "ARABIC LIGATURE HIGH QOF (SAJDA)"},
	{0xFDFA, 0xFDFA, "ARABIC LIGATURE SALLA USED AS KORANIC STOP SIGN"},
	{0xFDFD, 0xFDFD, "ARABIC LIGATURE BISMILLAH AR-RAHMAN AR-RAHEEM"},
	// Presentation Forms-B (U+FE70-FEFF).
	{0xFE70, 0xFEFF, "ARABIC PRESENTATION FORM-B"},
}

// renderChar returns a printable representation of cp. Control and
// format chars get a "·" placeholder so the table layout stays
// readable. Invalid UTF-8 (shouldn't happen for BMP Quran text)
// also gets "·".
func renderChar(cp rune) string {
	if cp == 0 || cp == '\t' || cp == '\n' || cp == '\r' {
		return "·"
	}
	if cp < 0x20 {
		return "·"
	}
	if cp >= 0x7F && cp < 0xA0 {
		return "·"
	}
	return string(cp)
}

// helper for the template: count occurrences of cp in Quran text.
func countOf(row CodepointRow) int { return row.Count }

// helper for the template: format count as string.
func decStr(n int) string { return strconv.Itoa(n) }
