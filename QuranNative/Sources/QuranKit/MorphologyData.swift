import Foundation

public struct WordKey: Hashable, Sendable {
    public let sura: Int; public let verse: Int; public let word: Int
    public init(sura: Int, verse: Int, word: Int) { self.sura = sura; self.verse = verse; self.word = word }
}

public struct VerseKey: Hashable, Sendable {
    public let sura: Int; public let verse: Int
    public init(sura: Int, verse: Int) { self.sura = sura; self.verse = verse }
}

public actor MorphologyData {
    public private(set) var segments: [Segment] = []

    var posIdx: [String: [Segment]] = [:]
    var wordIdx: [WordKey: [Segment]] = [:]
    var verseIdx: [VerseKey: [Segment]] = [:]
    var suraIdx: [Int: [Segment]] = [:]
    var roleIdx: [String: [Segment]] = [:]
    var phraseIdx: [String: [Segment]] = [:]
    var phraseFnIdx: [String: [Segment]] = [:]
    var caseMarkerIdx: [String: [Segment]] = [:]
    var constructIdx: [String: [Segment]] = [:]

    var rootIndex: [String: [String]] = [:]
    var refToRoot: [String: String] = [:]

    var parseLengthSec: Double = 0
    var glossIdx: [String: [Segment]] = [:]
    var wordTextIdx: [String: [Segment]] = [:]

    public var segmentCount: Int { segments.count }
    public var wordCount: Int { wordIdx.count }
    public var rootCount: Int { rootIndex.count }

    public init() {}

    public func load(csvPath: String, rootsPath: String) async throws {
        let csvURL = URL(fileURLWithPath: csvPath)
        segments = try MASAQParser.parse(url: csvURL)
        buildAllIndexes()

        let rootsURL = URL(fileURLWithPath: rootsPath)
        if FileManager.default.fileExists(atPath: rootsPath) {
            let data = try Data(contentsOf: rootsURL)
            if let dict = try JSONSerialization.jsonObject(with: data) as? [String: [String]] {
                rootIndex = dict
                var inv: [String: String] = [:]
                for (root, refs) in dict {
                    for ref in refs { inv[ref] = root }
                }
                refToRoot = inv
            }
        }
    }

    private func buildAllIndexes() {
        for seg in segments {
            posIdx[seg.pos, default: []].append(seg)
            wordIdx[WordKey(sura: seg.sura, verse: seg.verse, word: seg.word), default: []].append(seg)
            verseIdx[VerseKey(sura: seg.sura, verse: seg.verse), default: []].append(seg)
            suraIdx[seg.sura, default: []].append(seg)
            if !seg.syntacticRole.isEmpty { roleIdx[seg.syntacticRole, default: []].append(seg) }
            if !seg.phrase.isEmpty { phraseIdx[seg.phrase, default: []].append(seg) }
            if !seg.phraseFunction.isEmpty { phraseFnIdx[seg.phraseFunction, default: []].append(seg) }
            if !seg.caseMarker.isEmpty { caseMarkerIdx[seg.caseMarker, default: []].append(seg) }
            if !seg.possessive.isEmpty { constructIdx[seg.possessive, default: []].append(seg) }
            if !seg.withoutDiacritics.isEmpty { wordTextIdx[seg.withoutDiacritics, default: []].append(seg) }
            if !seg.gloss.isEmpty {
                for word in seg.gloss.lowercased().split(separator: " ") where !word.isEmpty {
                    let key = String(word)
                    glossIdx[key, default: []].append(seg)
                }
            }
        }
    }

    public func word(sura: Int, verse: Int, wordNum: Int) -> [Segment] {
        wordIdx[WordKey(sura: sura, verse: verse, word: wordNum)] ?? []
    }

    public func allWordSegments(sura: Int, verse aya: Int, wordCount: Int) -> [[Segment]] {
        var result: [[Segment]] = []
        result.reserveCapacity(wordCount)
        for w in 1...wordCount {
            result.append(wordIdx[WordKey(sura: sura, verse: aya, word: w)] ?? [])
        }
        return result
    }

    public func verseSegments(sura: Int, verse aya: Int) -> [Segment] {
        verseIdx[VerseKey(sura: sura, verse: aya)] ?? []
    }

    public func surahSegments(_ n: Int) -> [Segment] {
        suraIdx[n] ?? []
    }

    public func byPOS(_ pos: String) -> [Segment] {
        posIdx[pos] ?? []
    }

    public func byRole(_ role: String) -> [Segment] {
        roleIdx[role] ?? []
    }

    public func byPhraseType(_ phrase: String) -> [Segment] {
        phraseIdx[phrase] ?? []
    }

    public func byPhraseFunction(_ fn: String) -> [Segment] {
        phraseFnIdx[fn] ?? []
    }

    public func byCaseMarker(_ marker: String) -> [Segment] {
        caseMarkerIdx[marker] ?? []
    }

    public func byConstructState(_ state: String) -> [Segment] {
        constructIdx[state] ?? []
    }

    public func searchGloss(_ term: String) -> [Segment] {
        let lt = term.lowercased()
        var result: [Segment] = []
        for (key, segs) in glossIdx where key.lowercased().contains(lt) {
            result.append(contentsOf: segs)
        }
        return result
    }

    public func searchWord(_ term: String) -> [Segment] {
        if let exact = wordTextIdx[term] { return exact }
        return segments.filter { $0.withoutDiacritics == term }
    }

    public func rootFor(sura: Int, verse: Int, word: Int) -> String {
        refToRoot["\(sura):\(verse):\(word)"] ?? ""
    }

    public func versesWithRoot(_ root: String) -> [Location] {
        guard let refs = rootIndex[root] else { return [] }
        return refs.compactMap { ref in
            let parts = ref.split(separator: ":")
            guard parts.count == 3, let s = Int(parts[0]), let v = Int(parts[1]), let w = Int(parts[2]) else { return nil }
            return Location(sura: s, verse: v, token: w)
        }
    }

    // All roots, sorted alphabetically (using Arabic-aware ordering).
    public func allRoots() -> [String] {
        Array(rootIndex.keys).sorted()
    }

    // ── Fast bulk loader used by the Roots landing page ──
    //
    // Builds (synchronously, inside the actor) an index of every root
    // with its unique-verse-count and total-occurrence-count.  This
    // is a single O(N) sweep over `rootIndex` instead of 2 × 1642
    // async round-trips.  Returns immediately.
    public struct RootStats: Sendable {
        public let root: String
        public let ayahCount: Int
        public let totalOccurrences: Int
    }
    public func allRootStats() -> [RootStats] {
        var result: [RootStats] = []
        result.reserveCapacity(rootIndex.count)
        for (root, refs) in rootIndex {
            var uniqueVerses = Set<String>()
            uniqueVerses.reserveCapacity(refs.count)
            for ref in refs {
                let parts = ref.split(separator: ":")
                if parts.count >= 2 {
                    uniqueVerses.insert("\(parts[0]):\(parts[1])")
                }
            }
            result.append(RootStats(
                root: root,
                ayahCount: uniqueVerses.count,
                totalOccurrences: refs.count
            ))
        }
        return result
    }

    // For a root, get the unique (sura, ayah) pairs (deduplicated across
    // words that share the same root in the same verse).
    public func versesContainingRoot(_ root: String) -> [(sura: Int, ayah: Int)] {
        let refs = rootIndex[root] ?? []
        var seen: Set<String> = []
        var result: [(sura: Int, ayah: Int)] = []
        for ref in refs {
            let parts = ref.split(separator: ":")
            guard parts.count == 3,
                  let s = Int(parts[0]), let v = Int(parts[1]) else { continue }
            let key = "\(s):\(v)"
            if !seen.contains(key) {
                seen.insert(key)
                result.append((s, v))
            }
        }
        return result.sorted { lhs, rhs in
            if lhs.sura != rhs.sura { return lhs.sura < rhs.sura }
            return lhs.ayah < rhs.ayah
        }
    }

    // Bulk fetch: root + lemma for every word in a verse, in ONE call.
    // Used by Quran.words() to avoid N round-trips per word.
    public struct WordInfo: Sendable {
        public let wordNum: Int
        public let root: String
        public let lemma: String
    }
    public func wordsInfo(sura: Int, verse aya: Int, wordCount: Int) -> [WordInfo] {
        var result: [WordInfo] = []
        result.reserveCapacity(wordCount)
        for w in 1...wordCount {
            let r = refToRoot["\(sura):\(aya):\(w)"] ?? ""
            let segs = word(sura: sura, verse: aya, wordNum: w)
            let stem = segs.first(where: { $0.isStem })
            let lemma = stem?.withoutDiacritics ?? segs.first?.withoutDiacritics ?? ""
            result.append(WordInfo(wordNum: w, root: r, lemma: lemma))
        }
        return result
    }

    // All word occurrences of a root, with the Arabic surface form.
    public func wordsForRoot(_ root: String) -> [(sura: Int, ayah: Int, word: Int, surface: String)] {
        let refs = rootIndex[root] ?? []
        var result: [(sura: Int, ayah: Int, word: Int, surface: String)] = []
        for ref in refs {
            let parts = ref.split(separator: ":")
            guard parts.count == 3,
                  let s = Int(parts[0]), let v = Int(parts[1]), let w = Int(parts[2]) else { continue }
            let segs = word(sura: s, verse: v, wordNum: w)
            let stem = segs.first(where: { $0.isStem })
            let surface = stem?.wordText ?? segs.first?.wordText ?? ""
            result.append((s, v, w, surface))
        }
        return result
    }

    // Return up to `limit` distinct surface forms that share this root.
    // Used to show word-form variety in the Roots view.
    public func distinctSurfaceForms(forRoot root: String, limit: Int = 12) -> [String] {
        let refs = rootIndex[root] ?? []
        var seen: Set<String> = []
        var result: [String] = []
        for ref in refs {
            let parts = ref.split(separator: ":")
            guard parts.count == 3,
                  let s = Int(parts[0]), let v = Int(parts[1]), let w = Int(parts[2]) else { continue }
            let segs = word(sura: s, verse: v, wordNum: w)
            let stem = segs.first(where: { $0.isStem })
            let surface = stem?.wordText ?? segs.first?.wordText ?? ""
            if !surface.isEmpty, !seen.contains(surface) {
                seen.insert(surface)
                result.append(surface)
                if result.count >= limit { break }
            }
        }
        return result
    }

    public func roleAt(sura: Int, verse aya: Int, wordIdx w: Int, role: String) -> Segment? {
        word(sura: sura, verse: aya, wordNum: w).first { $0.syntacticRole == role }
    }

    public func subjectsInVerse(sura: Int, verse aya: Int) -> [Segment] {
        verseSegments(sura: sura, verse: aya).filter { $0.syntacticRole == "SUBJ" }
    }

    public func objectsInVerse(sura: Int, verse aya: Int) -> [Segment] {
        verseSegments(sura: sura, verse: aya).filter { $0.syntacticRole == "OBJ" }
    }

    public func agentsInVerse(sura: Int, verse aya: Int) -> [Segment] {
        verseSegments(sura: sura, verse: aya).filter { $0.syntacticRole == "AGNT" }
    }

    private static let predicateRoles: Set<String> = [
        "PRED", "PREP_OBJ", "GEN_CONS", "ADJ",
        "SUBJ_COP_PART", "SUBJ_COP_V", "PART_COP_PRED", "V_COP_PRED",
    ]

    public func predicatesInVerse(sura: Int, verse aya: Int) -> [Segment] {
        verseSegments(sura: sura, verse: aya).filter { Self.predicateRoles.contains($0.syntacticRole) }
    }

    public func sentenceType(sura: Int, verse aya: Int) -> String {
        let segs = verseSegments(sura: sura, verse: aya)
        let phrases = segs.map(\.phrase).filter { !$0.isEmpty }
        guard !phrases.isEmpty else { return "unknown" }
        var counts: [String: Int] = [:]
        for p in phrases { counts[p, default: 0] += 1 }
        let dominant = counts.max(by: { $0.value < $1.value })?.key ?? "unknown"
        switch dominant {
        case "VERB_SNT": return "verbal"
        case "NOM_SNT": return "nominal"
        case "COND_SNT": return "conditional"
        default: return "phrasal"
        }
    }

    public func lemmaFor(sura: Int, verse aya: Int, wordIdx w: Int) -> String {
        let segs = word(sura: sura, verse: aya, wordNum: w)
        return segs.first(where: \.isStem)?.withoutDiacritics ?? segs.first?.withoutDiacritics ?? ""
    }

    public func glossFor(sura: Int, verse aya: Int, wordIdx w: Int) -> String {
        let segs = word(sura: sura, verse: aya, wordNum: w)
        let stems = segs.filter(\.isStem)
        if !stems.isEmpty { return stems.map(\.gloss).joined(separator: " + ") }
        return segs.first?.gloss ?? ""
    }

    public func allSyntacticRoles() -> [String] { Array(roleIdx.keys).sorted() }
    public func allPhraseTypes() -> [String] { Array(phraseIdx.keys).sorted() }
    public func allCaseMarkers() -> [String] { Array(caseMarkerIdx.keys).sorted() }

    public func setSegmentsForCache(_ segs: [Segment], rootIndex: [String: [String]], refToRoot: [String: String]) {
        segments = segs
        buildAllIndexes()
        self.rootIndex = rootIndex
        self.refToRoot = refToRoot
    }
}
