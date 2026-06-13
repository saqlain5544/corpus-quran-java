import Foundation

public enum MaddRule: Sendable {
    case laazimKalami
    case laazimHarfi
    case muttasil
    case munfasil
    case aarid
    case leen
}

public struct MaddMatch: Sendable {
    public let rule: MaddRule
    public let insertionIndex: Int

    public init(rule: MaddRule, insertionIndex: Int) {
        self.rule = rule
        self.insertionIndex = insertionIndex
    }
}

public enum MaddDetector {
    private static let cpAlif: UInt32 = 0x0627
    private static let cpWaw: UInt32 = 0x0648
    private static let cpYa: UInt32 = 0x064A
    private static let cpAlifMaksura: UInt32 = 0x0649
    private static let cpFatha: UInt32 = 0x064E
    private static let cpDamma: UInt32 = 0x064F
    private static let cpKasra: UInt32 = 0x0650
    private static let cpSukun: UInt32 = 0x0652
    private static let cpShaddah: UInt32 = 0x0651
    private static let cpMaddah: UInt32 = 0x0653

    private static let hamzahSet: Set<UInt32> = [0x0621, 0x0622, 0x0623, 0x0625, 0x0626]
    private static let diacriticRange: ClosedRange<UInt32> = 0x064B...0x0654
    private static let maddLetterSet: Set<UInt32> = [cpAlif, cpWaw, cpYa]

    private struct Token {
        let scalar: UnicodeScalar
        let scalarIndex: Int
        let hasFatha: Bool
        let hasDamma: Bool
        let hasKasra: Bool
        let hasSukun: Bool
        let hasShaddah: Bool
        let hasMaddah: Bool
        let isHamzah: Bool
        let isAlif: Bool
        let isWaw: Bool
        let isYa: Bool
        let isAlifMaksura: Bool
    }

    private static func tokenize(_ word: String) -> [Token] {
        let scalars = Array(word.decomposedStringWithCanonicalMapping.unicodeScalars)
        var tokens: [Token] = []
        tokens.reserveCapacity(scalars.count / 2)
        var i = 0
        while i < scalars.count {
            let base = scalars[i]
            let baseIdx = i
            i += 1
            var f = false, d = false, k = false, su = false, sh = false, ma = false
            while i < scalars.count, diacriticRange.contains(scalars[i].value) {
                switch scalars[i].value {
                case cpFatha: f = true
                case cpDamma: d = true
                case cpKasra: k = true
                case cpSukun: su = true
                case cpShaddah: sh = true
                case cpMaddah: ma = true
                default: break
                }
                i += 1
            }
            tokens.append(Token(
                scalar: base, scalarIndex: baseIdx,
                hasFatha: f, hasDamma: d, hasKasra: k,
                hasSukun: su, hasShaddah: sh, hasMaddah: ma,
                isHamzah: hamzahSet.contains(base.value),
                isAlif: base.value == cpAlif,
                isWaw: base.value == cpWaw,
                isYa: base.value == cpYa,
                isAlifMaksura: base.value == cpAlifMaksura
            ))
        }
        return tokens
    }

    private enum SimpleState: Sendable {
        case idle
        case vowelFatha(Int)
        case vowelDamma(Int)
        case vowelKasra(Int)
    }

    private static func findMatch(in tokens: [Token]) -> MaddMatch? {
        var state: SimpleState = .idle

        for ti in 0..<tokens.count {
            let t = tokens[ti]
            let next = ti + 1 < tokens.count ? tokens[ti + 1] : nil
            let nextNext = ti + 2 < tokens.count ? tokens[ti + 2] : nil

            switch state {
            case .idle:
                if !t.hasSukun, !t.hasShaddah {
                    if t.hasFatha { state = .vowelFatha(t.scalarIndex) }
                    else if t.hasDamma { state = .vowelDamma(t.scalarIndex) }
                    else if t.hasKasra { state = .vowelKasra(t.scalarIndex) }
                }

            case .vowelFatha(let baseIdx):
                if t.isAlif || t.isAlifMaksura {
                    return classifyMaddLetter(t, next, nextNext, baseIdx)
                }
                if t.isWaw, t.hasSukun {
                    return MaddMatch(rule: .leen, insertionIndex: baseIdx)
                }
                if t.hasFatha, !t.hasSukun, !t.hasShaddah { state = .vowelFatha(t.scalarIndex) }
                else if t.hasDamma, !t.hasSukun, !t.hasShaddah { state = .vowelDamma(t.scalarIndex) }
                else if t.hasKasra, !t.hasSukun, !t.hasShaddah { state = .vowelKasra(t.scalarIndex) }
                else { state = .idle }

            case .vowelDamma(let baseIdx):
                if t.isWaw {
                    return classifyMaddLetter(t, next, nextNext, baseIdx)
                }
                if t.hasFatha, !t.hasSukun, !t.hasShaddah { state = .vowelFatha(t.scalarIndex) }
                else if t.hasDamma, !t.hasSukun, !t.hasShaddah { state = .vowelDamma(t.scalarIndex) }
                else if t.hasKasra, !t.hasSukun, !t.hasShaddah { state = .vowelKasra(t.scalarIndex) }
                else { state = .idle }

            case .vowelKasra(let baseIdx):
                if t.isYa, !t.hasShaddah {
                    return classifyMaddLetter(t, next, nextNext, baseIdx)
                }
                if t.isYa, t.hasShaddah {
                    if t.hasFatha, !t.hasSukun { state = .vowelFatha(t.scalarIndex) }
                    else { state = .idle }
                    continue
                }
                if t.hasFatha, !t.hasSukun, !t.hasShaddah { state = .vowelFatha(t.scalarIndex) }
                else if t.hasDamma, !t.hasSukun, !t.hasShaddah { state = .vowelDamma(t.scalarIndex) }
                else if t.hasKasra, !t.hasSukun, !t.hasShaddah { state = .vowelKasra(t.scalarIndex) }
                else { state = .idle }
            }
        }

        if case .idle = state { return nil }

        return nil
    }

    private static func classifyMaddLetter(_ maddToken: Token, _ next: Token?, _ nextNext: Token?,
                                            _ baseIdx: Int) -> MaddMatch? {
        if let n = next {
            if n.hasShaddah || (n.hasMaddah && nextNext?.hasShaddah == true) {
                return MaddMatch(rule: .laazimKalami, insertionIndex: baseIdx)
            }
            if n.isHamzah || n.isAlifMaksura {
                return MaddMatch(rule: .muttasil, insertionIndex: baseIdx)
            }
            for fwd in [n] {
                if fwd.hasShaddah {
                    return MaddMatch(rule: .laazimKalami, insertionIndex: baseIdx)
                }
                if fwd.isHamzah || fwd.isAlifMaksura {
                    return MaddMatch(rule: .muttasil, insertionIndex: baseIdx)
                }
            }
        }
        if next?.hasSukun == true || next == nil {
            return MaddMatch(rule: .aarid, insertionIndex: baseIdx)
        }
        return nil
    }

    public static func detect(inWord word: String) -> MaddMatch? {
        let tokens = tokenize(word)
        return findMatch(in: tokens)
    }

    public static func isMunfasil(endOfWord: String, startOfNextWord: String) -> Bool {
        let endTokens = tokenize(endOfWord)
        guard let last = endTokens.last else { return false }
        let isMaddEnd: Bool = (last.isAlif || last.isWaw || last.isYa || last.isAlifMaksura)
            && (last.hasFatha || last.hasDamma || last.hasKasra)
        if !isMaddEnd {
            if last.isAlif || last.isWaw || last.isYa || last.isAlifMaksura { return true }
            if last.hasSukun == false && (endTokens.count >= 2) {
                let prev = endTokens[endTokens.count - 2]
                if (prev.hasFatha && (last.isAlif || last.isAlifMaksura))
                    || (prev.hasDamma && last.isWaw)
                    || (prev.hasKasra && last.isYa && !last.hasShaddah) {
                } else { return false }
            } else { return false }
        }
        let nextTokens = tokenize(startOfNextWord)
        guard let first = nextTokens.first else { return false }
        return first.isHamzah
    }

    public static func munfasilInsertionIndex(_ word: String) -> Int? {
        let tokens = tokenize(word)
        for i in stride(from: tokens.count - 2, through: 0, by: -1) {
            let t = tokens[i]
            let next = tokens[i + 1]
            let isMadd: Bool = (next.isAlif && t.hasFatha)
                || (next.isWaw && t.hasDamma)
                || (next.isYa && t.hasKasra && !next.hasShaddah)
            if isMadd { return t.scalarIndex }
        }
        return nil
    }

    private static let muqattaLetter: Set<UInt32> = [0x0643, 0x0647, 0x064A, 0x0639, 0x0635, 0x062D, 0x0645, 0x0642, 0x0646, 0x0637, 0x0633, 0x0627, 0x0644, 0x0631]

    public static func detectMuqatta(inWord word: String) -> MaddMatch? {
        let tokens = tokenize(word)
        guard tokens.count == 1, muqattaLetter.contains(tokens[0].scalar.value) else { return nil }
        if tokens[0].hasFatha { return MaddMatch(rule: .laazimHarfi, insertionIndex: tokens[0].scalarIndex) }
        return nil
    }

    public static func applyElongation(to word: String, match: MaddMatch, count: Int) -> String {
        let scalars = Array(word.decomposedStringWithCanonicalMapping.unicodeScalars)
        guard count > 0 else { return word }
        let tatweel = String(repeating: "\u{0640}", count: count)
        var result = ""
        for (i, scalar) in scalars.enumerated() {
            result.append(String(scalar))
            if i == match.insertionIndex {
                result.append(tatweel)
            }
        }
        return result.precomposedStringWithCanonicalMapping
    }
}
