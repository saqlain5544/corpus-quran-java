import Foundation

public struct Division: Identifiable, Sendable {
    public var id: Int { number }
    public let number: Int
    public let startSura: Int
    public let startVerse: Int
    public let endSura: Int
    public let endVerse: Int
    public let nameArabic: String
    public let nameEnglish: String

    public init(number: Int, startSura: Int, startVerse: Int, endSura: Int, endVerse: Int,
                nameArabic: String = "", nameEnglish: String = "") {
        self.number = number; self.startSura = startSura; self.startVerse = startVerse
        self.endSura = endSura; self.endVerse = endVerse
        self.nameArabic = nameArabic; self.nameEnglish = nameEnglish
    }
}

private let JUZ_BOUNDARIES: [(sura: Int, verse: Int)] = [
    (1, 1), (2, 142), (2, 253), (3, 93), (4, 24),
    (4, 148), (5, 82), (6, 111), (7, 88), (8, 41),
    (9, 93), (11, 6), (12, 53), (15, 1), (17, 1),
    (18, 75), (21, 1), (23, 1), (25, 21), (27, 56),
    (29, 46), (33, 31), (36, 28), (39, 32), (41, 47),
    (46, 1), (51, 31), (58, 1), (67, 1), (78, 2),
]

public let ALL_JUZ: [Division] = {
    var results: [Division] = []
    for (i, (s, v)) in JUZ_BOUNDARIES.enumerated() {
        let num = i + 1
        let startSura = s
        let startVerse = v
        let endSura: Int
        let endVerse: Int
        if i + 1 < JUZ_BOUNDARIES.count {
            let next = JUZ_BOUNDARIES[i + 1]
            if next.verse == 1 {
                endSura = next.sura - 1
                endVerse = VERSE_COUNTS[endSura - 1]
            } else {
                endSura = next.sura
                endVerse = next.verse - 1
            }
        } else {
            endSura = 114; endVerse = 6
        }
        results.append(Division(number: num, startSura: startSura, startVerse: startVerse,
                                 endSura: endSura, endVerse: endVerse,
                                 nameArabic: "الجزء \(num)", nameEnglish: "Juz' \(num)"))
    }
    return results
}()

extension Array {
    subscript(safe index: Int) -> Element? {
        indices.contains(index) ? self[index] : nil
    }
}
