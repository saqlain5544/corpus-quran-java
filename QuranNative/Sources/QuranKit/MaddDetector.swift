import Foundation

public enum MaddRule: Sendable {
    case laazimKalami
    case laazimHarfi
    case muttasil
    case munfasil
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
    private static let diacriticRange: ClosedRange<UInt32> = 0x064B...0x0654

    private static let hamzahSet: Set<UInt32> = [
        0x0621, 0x0622, 0x0623, 0x0625, 0x0626,
    ]

    private struct Token {
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
        tokens.reserveCapacity(max(1, scalars.count / 2))
        var i = 0
        while i < scalars.count {
            let base = scalars[i]
            let baseIdx = i
            let cp = base.value
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
                scalarIndex: baseIdx,
                hasFatha: f, hasDamma: d, hasKasra: k,
                hasSukun: su, hasShaddah: sh, hasMaddah: ma,
                isHamzah: hamzahSet.contains(cp),
                isAlif: cp == cpAlif,
                isWaw: cp == cpWaw,
                isYa: cp == cpYa,
                isAlifMaksura: cp == cpAlifMaksura
            ))
        }
        return tokens
    }

    private enum State {
        case idle
        case vowelFatha(Int)
        case vowelDamma(Int)
        case vowelKasra(Int)
    }

    private static func findMatch(in tokens: [Token]) -> MaddMatch? {
        var state: State = .idle

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
                    if let m = classify(maddToken: t, next: next, nextNext: nextNext, baseIdx: baseIdx) {
                        return m
                    }
                    state = .idle
                    continue
                }
                transition(from: &state, token: t)

            case .vowelDamma(let baseIdx):
                if t.isWaw {
                    if let m = classify(maddToken: t, next: next, nextNext: nextNext, baseIdx: baseIdx) {
                        return m
                    }
                    state = .idle
                    continue
                }
                transition(from: &state, token: t)

            case .vowelKasra(let baseIdx):
                if t.isYa, !t.hasShaddah {
                    if let m = classify(maddToken: t, next: next, nextNext: nextNext, baseIdx: baseIdx) {
                        return m
                    }
                    state = .idle
                    continue
                }
                transition(from: &state, token: t)
            }
        }
        return nil
    }

    private static func transition(from state: inout State, token t: Token) {
        let vowel = vowel(on: t)
        if vowel == nil || t.hasSukun || t.hasShaddah { state = .idle; return }
        state = vowel!
    }

    private static func vowel(on t: Token) -> State? {
        if t.hasFatha { return .vowelFatha(t.scalarIndex) }
        if t.hasDamma { return .vowelDamma(t.scalarIndex) }
        if t.hasKasra { return .vowelKasra(t.scalarIndex) }
        return nil
    }

    private static func classify(maddToken: Token, next: Token?, nextNext: Token?,
                                  baseIdx: Int) -> MaddMatch? {
        guard let n = next else { return nil }
        if n.hasShaddah { return MaddMatch(rule: .laazimKalami, insertionIndex: baseIdx) }
        if n.hasMaddah, let nn = nextNext, nn.hasShaddah {
            return MaddMatch(rule: .laazimKalami, insertionIndex: baseIdx)
        }
        if n.isHamzah { return MaddMatch(rule: .muttasil, insertionIndex: baseIdx) }
        return nil
    }

    public static func detect(inWord word: String) -> MaddMatch? {
        findMatch(in: tokenize(word))
    }

    public static func isMunfasil(endOfWord: String, startOfNextWord: String) -> Bool {
        let endTokens = tokenize(endOfWord)
        guard endTokens.count >= 2 else { return false }
        let prev = endTokens[endTokens.count - 2]
        let last = endTokens[endTokens.count - 1]
        let isMaddEnd: Bool = (prev.hasFatha && (last.isAlif || last.isAlifMaksura))
            || (prev.hasDamma && last.isWaw)
            || (prev.hasKasra && last.isYa && !last.hasShaddah)
        guard isMaddEnd else { return false }
        let nextTokens = tokenize(startOfNextWord)
        guard let first = nextTokens.first else { return false }
        return first.isHamzah
    }

    public static func munfasilInsertionIndex(_ word: String) -> Int? {
        let tokens = tokenize(word)
        guard tokens.count >= 2 else { return nil }
        let prev = tokens[tokens.count - 2]
        let last = tokens[tokens.count - 1]
        let isMadd: Bool = (prev.hasFatha && (last.isAlif || last.isAlifMaksura))
            || (prev.hasDamma && last.isWaw)
            || (prev.hasKasra && last.isYa && !last.hasShaddah)
        return isMadd ? prev.scalarIndex : nil
    }

    public static func detectMuqatta(inWord word: String) -> MaddMatch? {
        let tokens = tokenize(word)
        guard tokens.count == 1 else { return nil }
        let t = tokens[0]
        guard !t.hasSukun, !t.hasShaddah else { return nil }
        if t.hasFatha && (t.isAlif || t.isWaw || t.isYa || t.isAlifMaksura) {
            return MaddMatch(rule: .laazimHarfi, insertionIndex: t.scalarIndex)
        }
        return nil
    }

    public static func applyElongation(to word: String, match: MaddMatch, count: Int) -> String {
        let scalars = Array(word.decomposedStringWithCanonicalMapping.unicodeScalars)
        guard count > 0, match.insertionIndex < scalars.count else { return word }
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
