import Foundation

public enum MASAQParser {
    private static let fieldNames = [
        "ID", "Sura_No", "Verse_No", "Word_No", "Segment_No",
        "Word", "Without_Diacritics", "Segmented_Word",
        "Morph_Tag", "Morph_Type", "Punctuation_Mark",
        "Invariable_Declinable", "Syntactic_Role",
        "Possessive_Construct", "Case_Mood", "Case_Mood_Marker",
        "Phrase", "Phrasal_Function", "Gloss",
    ]

    public static func parse(url: URL) throws -> [Segment] {
        let content = try String(contentsOf: url, encoding: .utf8)
        var segments: [Segment] = []
        let allLines = content.components(separatedBy: "\n")
        guard !allLines.isEmpty else { return [] }

        let header = allLines[0]
        let lines = allLines.dropFirst()
        let headerCols = header.split(separator: ",", omittingEmptySubsequences: false).map {
            String($0).trimmingCharacters(in: CharacterSet(charactersIn: "\r"))
        }
        let colIndex = Dictionary(headerCols.enumerated().map { ($1, $0) }, uniquingKeysWith: { first, _ in first })

        func col(_ name: String) -> Int { colIndex[name] ?? -1 }

        func field(_ cols: [String], _ name: String, _ default: String = "") -> String {
            let i = col(name)
            guard i >= 0, i < cols.count else { return `default` }
            let raw = cols[i].trimmingCharacters(in: CharacterSet(charactersIn: "\"\r"))
            // MASAQ uses two null-like sentinels: literal "None" and "(null)"
            // Both indicate missing/morphologically-null data
            let lower = raw.lowercased().trimmingCharacters(in: .whitespaces)
            if lower == "none" || lower == "null" || lower == "(null)" { return "" }
            return raw
        }

        func fieldInt(_ cols: [String], _ name: String, _ default: Int = 0) -> Int {
            Int(field(cols, name)) ?? `default`
        }

        for line in lines {
            let trimmed = line.trimmingCharacters(in: .whitespacesAndNewlines)
            if trimmed.isEmpty { continue }
            let cols = parseCSVLine(trimmed)
            guard cols.count >= 10 else { continue }

            let s = Segment(
                id: fieldInt(cols, "ID"),
                sura: fieldInt(cols, "Sura_No"),
                verse: fieldInt(cols, "Verse_No"),
                word: fieldInt(cols, "Word_No"),
                segment: fieldInt(cols, "Segment_No"),
                wordText: field(cols, "Word"),
                withoutDiacritics: field(cols, "Without_Diacritics"),
                segmentedText: field(cols, "Segmented_Word"),
                pos: field(cols, "Morph_Tag"),
                morphType: field(cols, "Morph_Type"),
                syntacticRole: field(cols, "Syntactic_Role"),
                possessive: field(cols, "Possessive_Construct"),
                caseMood: field(cols, "Case_Mood"),
                caseMarker: field(cols, "Case_Mood_Marker"),
                phrase: field(cols, "Phrase"),
                phraseFunction: field(cols, "Phrasal_Function"),
                gloss: field(cols, "Gloss"),
                punctuation: field(cols, "Punctuation_Mark"),
                declinability: field(cols, "Invariable_Declinable")
            )
            segments.append(s)
        }
        return segments
    }

    private static func parseCSVLine(_ line: String) -> [String] {
        var fields: [String] = []
        var current = ""
        var inQuotes = false
        for ch in line {
            if ch == "\"" {
                inQuotes.toggle()
            } else if ch == "," && !inQuotes {
                fields.append(current)
                current = ""
            } else {
                current.append(ch)
            }
        }
        fields.append(current)
        return fields
    }
}
