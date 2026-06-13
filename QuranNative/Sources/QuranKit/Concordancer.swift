import Foundation

public struct ConcordanceLine: Identifiable, Hashable, Sendable {
    public var id: String { "\(sura):\(verse):\(word)" }
    public let sura: Int
    public let verse: Int
    public let word: Int
    public let text: String
    public let targetWord: String
    public let position: Int
    public let leftContext: String
    public let rightContext: String

    public init(sura: Int, verse: Int, word: Int, text: String, targetWord: String,
                position: Int, leftContext: String, rightContext: String) {
        self.sura = sura; self.verse = verse; self.word = word
        self.text = text; self.targetWord = targetWord
        self.position = position; self.leftContext = leftContext; self.rightContext = rightContext
    }
}

public actor Concordancer {
    private let morphology: MorphologyData

    public init(morphology: MorphologyData) {
        self.morphology = morphology
    }

    public func searchWord(_ word: String, context: Int = 3) async -> [ConcordanceLine] {
        let segments = await morphology.searchWord(word)
        return await buildLines(from: segments, targetWord: word, context: context)
    }

    public func searchGloss(_ term: String, context: Int = 3) async -> [ConcordanceLine] {
        let segments = await morphology.searchGloss(term)
        return await buildLines(from: segments, targetWord: term, context: context)
    }

    public func searchRoot(_ root: String, context: Int = 3) async -> [ConcordanceLine] {
        let refs = await morphology.versesWithRoot(root)
        var lines: [ConcordanceLine] = []
        var wordsCache: [VerseKey: [String]] = [:]
        for loc in refs {
            let key = VerseKey(sura: loc.sura, verse: loc.verse)
            let words: [String]
            if let cached = wordsCache[key] {
                words = cached
            } else {
                let segs = await morphology.verseSegments(sura: loc.sura, verse: loc.verse)
                var seen = Set<Int>()
                words = segs
                    .filter { seen.insert($0.word).inserted }
                    .sorted { $0.word < $1.word }
                    .map { $0.wordText }
                wordsCache[key] = words
            }
            let wordSegs = await morphology.word(sura: loc.sura, verse: loc.verse, wordNum: loc.token)
            guard let first = wordSegs.first else { continue }
            let verseText = words.joined(separator: " ")
            let target = first.wordText
            let wIdx = loc.token - 1
            let leftStart = max(0, wIdx - context)
            let leftEnd = wIdx
            let rightStart = wIdx + 1
            let rightEnd = min(words.count, wIdx + 1 + context)
            lines.append(ConcordanceLine(
                sura: loc.sura, verse: loc.verse, word: loc.token,
                text: verseText, targetWord: target,
                position: loc.token,
                leftContext: words[leftStart..<leftEnd].joined(separator: " "),
                rightContext: rightStart < rightEnd ? words[rightStart..<rightEnd].joined(separator: " ") : ""
            ))
        }
        return lines
    }

    public func searchSura(_ sura: Int) async -> [ConcordanceLine] {
        let segs = await morphology.surahSegments(sura)
        var lines: [ConcordanceLine] = []
        var byWord: [String: [Segment]] = [:]
        for s in segs {
            let key = "\(s.sura):\(s.verse):\(s.word)"
            byWord[key, default: []].append(s)
        }
        let sortedEntries = byWord.map { (key, segs) -> (key: String, segs: [Segment], s: Int, v: Int, w: Int) in
            let p = key.split(separator: ":").compactMap { Int($0) }
            return (key, segs, p[0], p[1], p.count > 2 ? p[2] : 0)
        }.sorted { lhs, rhs in
            if lhs.s != rhs.s { return lhs.s < rhs.s }
            if lhs.v != rhs.v { return lhs.v < rhs.v }
            return lhs.w < rhs.w
        }
        for (_, wordSegs, sNum, vNum, wNum) in sortedEntries {
            let text = wordSegs.first?.wordText ?? ""
            let target = wordSegs.first?.withoutDiacritics ?? ""
            lines.append(ConcordanceLine(
                sura: sNum, verse: vNum, word: wNum,
                text: text, targetWord: target,
                position: 0, leftContext: "", rightContext: ""
            ))
        }
        return lines
    }

    private func buildLines(from segments: [Segment], targetWord: String, context: Int) async -> [ConcordanceLine] {
        var lines: [ConcordanceLine] = []
        lines.reserveCapacity(segments.count)
        var currentKey: VerseKey?
        var currentWords: [String] = []
        for seg in segments {
            let key = VerseKey(sura: seg.sura, verse: seg.verse)
            if key != currentKey {
                currentKey = key
                let segs = await morphology.verseSegments(sura: seg.sura, verse: seg.verse)
                var seen = Set<Int>()
                currentWords = segs
                    .filter { seen.insert($0.word).inserted }
                    .sorted { $0.word < $1.word }
                    .map { $0.wordText }
            }
            let verseText = currentWords.joined(separator: " ")
            let wIdx = seg.word - 1
            let leftStart = max(0, wIdx - context)
            let leftEnd = wIdx
            let rightStart = min(currentWords.count, wIdx + 1)
            let rightEnd = min(currentWords.count, wIdx + 1 + context)
            lines.append(ConcordanceLine(
                sura: seg.sura, verse: seg.verse, word: seg.word,
                text: verseText, targetWord: seg.withoutDiacritics,
                position: seg.word,
                leftContext: currentWords[leftStart..<leftEnd].joined(separator: " "),
                rightContext: rightStart < rightEnd ? currentWords[rightStart..<rightEnd].joined(separator: " ") : ""
            ))
        }
        return lines
    }
}
