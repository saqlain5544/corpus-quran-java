import Foundation

public struct SajdaVerse: Identifiable, Sendable {
    public var id: String { "\(sura):\(verse)" }
    public let sura: Int
    public let verse: Int
    public let name: String
    public let note: String

    public init(sura: Int, verse: Int, name: String, note: String = "") {
        self.sura = sura; self.verse = verse; self.name = name; self.note = note
    }
}

public let SAJDA_VERSES: [SajdaVerse] = [
    SajdaVerse(sura: 7, verse: 206, name: "Al-Aʿrāf", note: "recommended"),
    SajdaVerse(sura: 13, verse: 15, name: "Ar-Raʿd", note: "recommended"),
    SajdaVerse(sura: 16, verse: 50, name: "An-Naḥl", note: "recommended"),
    SajdaVerse(sura: 17, verse: 109, name: "Al-Isrāʾ", note: "recommended"),
    SajdaVerse(sura: 19, verse: 58, name: "Maryam", note: "recommended"),
    SajdaVerse(sura: 22, verse: 18, name: "Al-Ḥajj (1)", note: "recommended"),
    SajdaVerse(sura: 22, verse: 77, name: "Al-Ḥajj (2)", note: "recommended"),
    SajdaVerse(sura: 25, verse: 60, name: "Al-Furqān", note: "recommended"),
    SajdaVerse(sura: 27, verse: 26, name: "An-Naml", note: "recommended"),
    SajdaVerse(sura: 32, verse: 15, name: "As-Sajdah", note: "obligatory (Shāfiʿī)"),
    SajdaVerse(sura: 38, verse: 24, name: "Ṣād", note: "recommended"),
    SajdaVerse(sura: 41, verse: 38, name: "Fuṣṣilat", note: "obligatory (Ḥanafī)"),
    SajdaVerse(sura: 53, verse: 62, name: "An-Najm", note: "obligatory"),
    SajdaVerse(sura: 84, verse: 21, name: "Al-Inshiqāq", note: "recommended"),
    SajdaVerse(sura: 96, verse: 19, name: "Al-ʿAlaq", note: "obligatory"),
]

public let SAJDA_SURAS: Set<Int> = Set(SAJDA_VERSES.map(\.sura))

public func isSajdaVerse(sura: Int, verse: Int) -> Bool {
    SAJDA_VERSES.contains { $0.sura == sura && $0.verse == verse }
}

public func getSajda(sura: Int, verse: Int) -> SajdaVerse? {
    SAJDA_VERSES.first { $0.sura == sura && $0.verse == verse }
}

public func allSajdaVerses() -> [SajdaVerse] { SAJDA_VERSES }
public var sajdaCount: Int { SAJDA_VERSES.count }
