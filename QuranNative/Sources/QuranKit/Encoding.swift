import Foundation

public let CP_TO_CHAR_TYPE: [UInt32: CharacterType] = [
    0x0621: .hamza,
    0x0622: .alifWithMadda,
    0x0623: .alifWithHamzaAbove,
    0x0624: .wawWithHamza,
    0x0625: .alifWithHamzaBelow,
    0x0626: .yaWithHamza,
    0x0627: .alif,
    0x0628: .ba,
    0x0629: .taMarbuta,
    0x062A: .ta,
    0x062B: .tha,
    0x062C: .jeem,
    0x062D: .hha,
    0x062E: .kha,
    0x062F: .dal,
    0x0630: .thal,
    0x0631: .ra,
    0x0632: .zai,
    0x0633: .seen,
    0x0634: .sheen,
    0x0635: .sad,
    0x0636: .dad,
    0x0637: .tta,
    0x0638: .zza,
    0x0639: .ain,
    0x063A: .ghain,
    0x0641: .fa,
    0x0642: .qaf,
    0x0643: .kaf,
    0x0644: .lam,
    0x0645: .meem,
    0x0646: .noon,
    0x0647: .ha,
    0x0648: .waw,
    0x0649: .alifMaksura,
    0x064A: .ya,
    0x0671: .alifWasla,
    0x0640: .tatweel,
    0x0670: .daggerAlif,
    0x06D6: .smallHighSadLamAlef,
    0x06D7: .smallHighQafLamAlef,
    0x06D8: .smallHighMeemInit,
    0x06D9: .smallHighLamAlef,
    0x06DA: .smallHighJeem,
    0x06DB: .smallHighThreeDots,
    0x06DC: .smallHighSeen,
    0x06DF: .smallHighRoundedZero,
    0x06E0: .smallHighRectZero,
    0x06E2: .smallHighMeem,
    0x06E3: .smallLowSeen,
    0x06E5: .smallWaw,
    0x06E6: .smallYeh,
    0x06E7: .smallHighYeh,
    0x06E8: .smallHighNoon,
    0x06EA: .emptyCentreLowStop,
    0x06EB: .emptyCentreHighStop,
    0x06EC: .roundedHighStop,
    0x06ED: .smallLowMeem,
]

public let CP_TO_DIACRITIC: [UInt32: DiacriticTypes] = [
    0x064E: .fatha,
    0x064F: .damma,
    0x0650: .kasra,
    0x0651: .shadda,
    0x0652: .sukun,
    0x064B: .tanweenFatha,
    0x064C: .tanweenDamma,
    0x064D: .tanweenKasra,
    0x0653: .maddahAbove,
    0x0654: .hamzaAbove,
]

public let CHAR_TYPE_TO_BUCKWALTER: [CharacterType: Character] = [
    .hamza: "'", .alifWithMadda: "|", .alifWithHamzaAbove: ">",
    .wawWithHamza: "&", .alifWithHamzaBelow: "<", .yaWithHamza: "}",
    .alif: "A", .ba: "b", .taMarbuta: "p", .ta: "t",
    .tha: "v", .jeem: "j", .hha: "H", .kha: "x",
    .dal: "d", .thal: "*", .ra: "r", .zai: "z",
    .seen: "s", .sheen: "$", .sad: "S", .dad: "D",
    .tta: "T", .zza: "Z", .ain: "E", .ghain: "g",
    .fa: "f", .qaf: "q", .kaf: "k", .lam: "l",
    .meem: "m", .noon: "n", .ha: "h", .waw: "w",
    .alifMaksura: "Y", .ya: "y", .alifWasla: "{",
    .tatweel: "_", .daggerAlif: "`",
]

public let CHAR_TYPE_TO_CP: [CharacterType: UnicodeScalar] = [
    .hamza: "\u{0621}", .alifWithMadda: "\u{0622}", .alifWithHamzaAbove: "\u{0623}",
    .wawWithHamza: "\u{0624}", .alifWithHamzaBelow: "\u{0625}", .yaWithHamza: "\u{0626}",
    .alif: "\u{0627}", .ba: "\u{0628}", .taMarbuta: "\u{0629}", .ta: "\u{062A}",
    .tha: "\u{062B}", .jeem: "\u{062C}", .hha: "\u{062D}", .kha: "\u{062E}",
    .dal: "\u{062F}", .thal: "\u{0630}", .ra: "\u{0631}", .zai: "\u{0632}",
    .seen: "\u{0633}", .sheen: "\u{0634}", .sad: "\u{0635}", .dad: "\u{0636}",
    .tta: "\u{0637}", .zza: "\u{0638}", .ain: "\u{0639}", .ghain: "\u{063A}",
    .fa: "\u{0641}", .qaf: "\u{0642}", .kaf: "\u{0643}", .lam: "\u{0644}",
    .meem: "\u{0645}", .noon: "\u{0646}", .ha: "\u{0647}", .waw: "\u{0648}",
    .alifMaksura: "\u{0649}", .ya: "\u{064A}", .alifWasla: "\u{0671}",
    .tatweel: "\u{0640}", .daggerAlif: "\u{0670}",
    .smallHighSadLamAlef: "\u{06D6}", .smallHighQafLamAlef: "\u{06D7}",
    .smallHighMeemInit: "\u{06D8}", .smallHighLamAlef: "\u{06D9}",
    .smallHighJeem: "\u{06DA}", .smallHighThreeDots: "\u{06DB}",
    .smallHighSeen: "\u{06DC}", .smallHighRoundedZero: "\u{06DF}",
    .smallHighRectZero: "\u{06E0}", .smallHighMeem: "\u{06E2}",
    .smallLowSeen: "\u{06E3}", .smallWaw: "\u{06E5}",
    .smallYeh: "\u{06E6}", .smallHighYeh: "\u{06E7}",
    .smallHighNoon: "\u{06E8}", .emptyCentreLowStop: "\u{06EA}",
    .emptyCentreHighStop: "\u{06EB}", .roundedHighStop: "\u{06EC}",
    .smallLowMeem: "\u{06ED}",
]

public let DIACRITIC_TO_BUCKWALTER: [DiacriticTypes: Character] = [
    .fatha: "a", .damma: "u", .kasra: "i",
    .shadda: "~", .sukun: "o",
    .tanweenFatha: "F", .tanweenDamma: "N", .tanweenKasra: "K",
]

public let BUCKWALTER_TO_CP: [Character: UnicodeScalar] = {
    var result: [Character: UnicodeScalar] = [:]
    for (ct, bw) in CHAR_TYPE_TO_BUCKWALTER {
        if let cp = CHAR_TYPE_TO_CP[ct] {
            result[bw] = cp
        }
    }
    result["a"] = "\u{064E}"
    result["u"] = "\u{064F}"
    result["i"] = "\u{0650}"
    result["o"] = "\u{0652}"
    result["~"] = "\u{0651}"
    result["F"] = "\u{064B}"
    result["N"] = "\u{064C}"
    result["K"] = "\u{064D}"
    return result
}()

private let DIACRITIC_CP_ORDER: [UInt32] = [
    0x0651, 0x064E, 0x064F, 0x0650,
    0x064B, 0x064C, 0x064D, 0x0652,
    0x0653, 0x0654,
]

public func decode(_ text: String) -> [ArabicCharacter] {
    var result: [ArabicCharacter] = []
    for scalar in text.precomposedStringWithCanonicalMapping.unicodeScalars {
        let cp = scalar.value
        if let charType = CP_TO_CHAR_TYPE[cp] {
            result.append(ArabicCharacter(charType: charType))
        } else if let diacritic = CP_TO_DIACRITIC[cp] {
            if var last = result.last {
                last.diacritics.insert(diacritic)
                result[result.count - 1] = last
            }
        }
    }
    return result
}

public func encode(_ characters: [ArabicCharacter]) -> String {
    var scalars = String.UnicodeScalarView()
    for ch in characters {
        if let cp = CHAR_TYPE_TO_CP[ch.charType] {
            scalars.append(cp)
        }
        for cpOrder in DIACRITIC_CP_ORDER {
            switch cpOrder {
            case 0x0651: if ch.diacritics.contains(.shadda) { scalars.append(UnicodeScalar(cpOrder)!) }
            case 0x064E: if ch.diacritics.contains(.fatha) { scalars.append(UnicodeScalar(cpOrder)!) }
            case 0x064F: if ch.diacritics.contains(.damma) { scalars.append(UnicodeScalar(cpOrder)!) }
            case 0x0650: if ch.diacritics.contains(.kasra) { scalars.append(UnicodeScalar(cpOrder)!) }
            case 0x064B: if ch.diacritics.contains(.tanweenFatha) { scalars.append(UnicodeScalar(cpOrder)!) }
            case 0x064C: if ch.diacritics.contains(.tanweenDamma) { scalars.append(UnicodeScalar(cpOrder)!) }
            case 0x064D: if ch.diacritics.contains(.tanweenKasra) { scalars.append(UnicodeScalar(cpOrder)!) }
            case 0x0652: if ch.diacritics.contains(.sukun) { scalars.append(UnicodeScalar(cpOrder)!) }
            case 0x0653: if ch.diacritics.contains(.maddahAbove) { scalars.append(UnicodeScalar(cpOrder)!) }
            case 0x0654: if ch.diacritics.contains(.hamzaAbove) { scalars.append(UnicodeScalar(cpOrder)!) }
            default: break
            }
        }
    }
    return String(scalars)
}

public func decodeBuckwalter(_ text: String) -> [ArabicCharacter] {
    var result: [ArabicCharacter] = []
    var i = text.startIndex
    while i < text.endIndex {
        let ch = text[i]
        if let cp = BUCKWALTER_TO_CP[ch] {
            if let charType = CP_TO_CHAR_TYPE[cp.value] {
                result.append(ArabicCharacter(charType: charType))
            } else if let diacritic = CP_TO_DIACRITIC[cp.value] {
                if var last = result.last {
                    last.diacritics.insert(diacritic)
                    result[result.count - 1] = last
                }
            }
        }
        i = text.index(after: i)
    }
    return result
}
