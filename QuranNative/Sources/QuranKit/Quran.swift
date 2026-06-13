import Foundation

public struct WordAnalysis: Identifiable, Sendable {
    public var id: String { "\(wordIndex)" }
    public let wordIndex: Int
    public let text: String
    public let segments: [Segment]
    public let root: String?
    public let lemma: String?

    public init(wordIndex: Int, text: String, segments: [Segment],
                root: String? = nil, lemma: String? = nil) {
        self.wordIndex = wordIndex
        self.text = text
        self.segments = segments
        self.root = root
        self.lemma = lemma
    }
    public var gloss: String {
        let stems = segments.filter(\.isStem)
        if !stems.isEmpty { return stems.map(\.gloss).joined(separator: " + ") }
        return segments.first?.gloss ?? ""
    }
    public var posTags: [String] { segments.map(\.pos) }
    public var lemmas: [String] { segments.filter(\.isStem).map(\.withoutDiacritics) }
}

public class Quran: @unchecked Sendable {
    private let morphology: MorphologyData
    private let reader: Reader

    public init(morphology: MorphologyData, reader: Reader) {
        self.morphology = morphology
        self.reader = reader
    }

    public func verseText(sura: Int, aya: Int) async -> String {
        await reader.verse(sura: sura, aya: aya)
    }

    public func words(sura: Int, aya: Int) async -> [WordAnalysis] {
        let text = await reader.verse(sura: sura, aya: aya)
        let tokens = text.split(separator: " ")
        let wordCount = tokens.count
        async let wordsInfo = morphology.wordsInfo(sura: sura, verse: aya, wordCount: wordCount)
        async let allSegs = morphology.allWordSegments(sura: sura, verse: aya, wordCount: wordCount)
        let info = await wordsInfo
        let segsByWord = await allSegs
        var analyses: [WordAnalysis] = []
        analyses.reserveCapacity(wordCount)
        for (idx, token) in tokens.enumerated() {
            let infoEntry = info[idx]
            let root = infoEntry.root.isEmpty ? nil : infoEntry.root
            let lemma = infoEntry.lemma.isEmpty ? nil : infoEntry.lemma
            analyses.append(WordAnalysis(
                wordIndex: idx + 1,
                text: String(token),
                segments: segsByWord[idx],
                root: root,
                lemma: lemma
            ))
        }
        return analyses
    }

    public func gloss(sura: Int, aya: Int) async -> [String] {
        await words(sura: sura, aya: aya).map(\.gloss)
    }

    public func lemmas(sura: Int, aya: Int) async -> [String] {
        let ws = await words(sura: sura, aya: aya)
        return ws.map { $0.lemmas.joined(separator: "|") }
    }

    public func roots(sura: Int, aya: Int) async -> [String] {
        let text = await reader.verse(sura: sura, aya: aya)
        let tokens = text.split(separator: " ")
        let wordCount = tokens.count
        let info = await morphology.wordsInfo(sura: sura, verse: aya, wordCount: wordCount)
        return info.map { $0.root.isEmpty ? "" : $0.root }
    }

    public func pos(sura: Int, aya: Int) async -> [String] {
        let ws = await words(sura: sura, aya: aya)
        return ws.map { $0.posTags.joined(separator: "|") }
    }

}
