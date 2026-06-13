import Foundation

public enum MaddRule: Sendable {
    case laazimKalami
    case laazimHarfi
    case muttasil
    case munfasil
}

public enum MaddLetterClass: Sendable {
    case alif
    case waw
    case ya
    case alifMaksura
}

public struct MaddMatch: Sendable {
    public let rule: MaddRule
    public let letterClass: MaddLetterClass

    public init(rule: MaddRule, letterClass: MaddLetterClass) {
        self.rule = rule
        self.letterClass = letterClass
    }

    public var tracking: CGFloat {
        switch letterClass {
        case .alif, .waw:     return 0.6
        case .ya, .alifMaksura: return 0.4
        }
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

    private static let hamzahSet: Set<UInt32> = [0x0621, 0x0622, 0x0623, 0x0625, 0x0626]

    // Unicode 17.0 ArabicShaping.txt — Joining_Type R (Right_Joining)
    private static let rightOnlySet: Set<UInt32> = [
        0x0622, 0x0623, 0x0624, 0x0625, 0x0627, 0x0629,
        0x062F, 0x0630, 0x0631, 0x0632, 0x0648,
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
        let isRightOnly: Bool
        var maddLetterClass: MaddLetterClass? {
            if isAlif { return .alif }
            if isWaw { return .waw }
            if isYa { return .ya }
            if isAlifMaksura { return .alifMaksura }
            return nil
        }
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
                isAlifMaksura: cp == cpAlifMaksura,
                isRightOnly: rightOnlySet.contains(cp)
            ))
        }
        return tokens
    }

    private enum State {
        case idle
        case vowelFatha
        case vowelDamma
        case vowelKasra
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
                    if t.hasFatha { state = .vowelFatha }
                    else if t.hasDamma { state = .vowelDamma }
                    else if t.hasKasra { state = .vowelKasra }
                }

            case .vowelFatha:
                if t.isAlif || t.isAlifMaksura {
                    if let m = classify(maddToken: t, next: next, nextNext: nextNext) {
                        return m
                    }
                    state = .idle
                    continue
                }
                transition(from: &state, token: t)

            case .vowelDamma:
                if t.isWaw {
                    if let m = classify(maddToken: t, next: next, nextNext: nextNext) {
                        return m
                    }
                    state = .idle
                    continue
                }
                transition(from: &state, token: t)

            case .vowelKasra:
                if t.isYa, !t.hasShaddah {
                    if let m = classify(maddToken: t, next: next, nextNext: nextNext) {
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
        let v = vowel(on: t)
        if v == nil || t.hasSukun || t.hasShaddah { state = .idle; return }
        state = v!
    }

    private static func vowel(on t: Token) -> State? {
        if t.hasFatha { return .vowelFatha }
        if t.hasDamma { return .vowelDamma }
        if t.hasKasra { return .vowelKasra }
        return nil
    }

    private static func classify(maddToken: Token, next: Token?, nextNext: Token?) -> MaddMatch? {
        guard let letterClass = maddToken.maddLetterClass else { return nil }
        guard let n = next else { return nil }
        if n.hasShaddah { return MaddMatch(rule: .laazimKalami, letterClass: letterClass) }
        if n.hasMaddah, let nn = nextNext, nn.hasShaddah {
            return MaddMatch(rule: .laazimKalami, letterClass: letterClass)
        }
        if n.isHamzah { return MaddMatch(rule: .muttasil, letterClass: letterClass) }
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

    public static func detectMuqatta(inWord word: String) -> MaddMatch? {
        let tokens = tokenize(word)
        guard tokens.count == 1 else { return nil }
        let t = tokens[0]
        guard !t.hasSukun, !t.hasShaddah else { return nil }
        if let letterClass = t.maddLetterClass, t.hasFatha {
            return MaddMatch(rule: .laazimHarfi, letterClass: letterClass)
        }
        return nil
    }
}
