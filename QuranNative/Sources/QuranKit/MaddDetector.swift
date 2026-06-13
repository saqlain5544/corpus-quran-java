import Foundation

public enum MaddType: Sendable {
    case laazim
    case muttasil
    case munfasil
}

public struct MaddMatch: Sendable {
    public let type: MaddType
    public let scalarIndex: Int

    public init(type: MaddType, scalarIndex: Int) {
        self.type = type
        self.scalarIndex = scalarIndex
    }
}

public enum MaddDetector {
    private static let maddLetters: Set<UInt32> = [0x0627, 0x0648, 0x064A]
    private static let hamzaScalars: Set<UInt32> = [0x0621, 0x0623, 0x0625, 0x0626]

    public static func detect(inWord word: String) -> MaddMatch? {
        let scalars = Array(word.precomposedStringWithCanonicalMapping.unicodeScalars)
        for i in 0..<scalars.count {
            guard maddLetters.contains(scalars[i].value) else { continue }

            let next = i + 1
            guard next < scalars.count else { continue }

            if scalars[next].value == 0x0651 {
                return MaddMatch(type: .laazim, scalarIndex: i)
            }
            if scalars[next].value == 0x0653 {
                let afterMaddah = next + 1
                if afterMaddah < scalars.count, scalars[afterMaddah].value == 0x0651 {
                    return MaddMatch(type: .laazim, scalarIndex: i)
                }
            }

            for j in (i + 1)..<scalars.count {
                let cp = scalars[j].value
                if cp == 0x0651 { return MaddMatch(type: .laazim, scalarIndex: i) }
                if hamzaScalars.contains(cp) { return MaddMatch(type: .muttasil, scalarIndex: i) }
                if cp >= 0x064B { break }
            }
        }
        return nil
    }

    public static func isMunfasil(endOfWord: String, startOfNextWord: String) -> Bool {
        guard let last = endOfWord.precomposedStringWithCanonicalMapping.unicodeScalars.last else {
            return false
        }
        let bareLast: UInt32
        switch last.value {
        case 0x0649: bareLast = 0x064A
        case 0x06E5, 0x06E6, 0x06E7, 0x06E8, 0x06ED: return true
        default: bareLast = last.value
        }
        guard maddLetters.contains(bareLast) else { return false }

        guard let first = startOfNextWord.precomposedStringWithCanonicalMapping.unicodeScalars.first else {
            return false
        }
        return hamzaScalars.contains(first.value)
    }

    public static func applyElongation(to word: String, match: MaddMatch, count: Int) -> String {
        let scalars = Array(word.decomposedStringWithCanonicalMapping.unicodeScalars)
        guard count > 0 else { return word }
        let tatweel = String(repeating: "\u{0640}", count: count)
        var result = ""
        for (i, scalar) in scalars.enumerated() {
            result.append(String(scalar))
            if i == match.scalarIndex {
                result.append(tatweel)
            }
        }
        return result.precomposedStringWithCanonicalMapping
    }
}
