import Foundation

/// Convert Buckwalter transliteration characters to their Arabic
/// equivalents.  Used to display Quranic roots in Arabic script.
///
/// Example: "kwn" → "كون", "Alh" → "الله", "Elm" → "علم"
enum BuckwalterArabic {
    private static let map: [Character: Character] = [
        "'": "\u{0621}", // ء hamza
        "A": "\u{0627}", // ا alif
        "b": "\u{0628}", // ب ba
        "t": "\u{062A}", // ت ta
        "v": "\u{062B}", // ث tha
        "j": "\u{062C}", // ج jeem
        "H": "\u{062D}", // ح hha
        "x": "\u{062E}", // خ kha
        "d": "\u{062F}", // د dal
        "*": "\u{0630}", // ذ thal
        "r": "\u{0631}", // ر ra
        "z": "\u{0632}", // ز zai
        "s": "\u{0633}", // س seen
        "$": "\u{0634}", // ش sheen
        "S": "\u{0635}", // ص sad
        "D": "\u{0636}", // ḑ dad
        "T": "\u{0637}", // ط tta
        "Z": "\u{0638}", // ظ zza
        "E": "\u{0639}", // ع ain
        "g": "\u{063A}", // غ ghain
        "f": "\u{0641}", // ف fa
        "q": "\u{0642}", // ق qaf
        "k": "\u{0643}", // ك kaf
        "l": "\u{0644}", // ل lam
        "m": "\u{0645}", // م meem
        "n": "\u{0646}", // ن noon
        "h": "\u{0647}", // ه ha
        "w": "\u{0648}", // و waw
        "Y": "\u{0649}", // ى alifMaksura
        "y": "\u{064A}", // ي ya
        "p": "\u{0629}", // ة taMarbuta
    ]

    /// Convert a single Buckwalter string to Arabic.  Unknown
    /// characters are passed through.
    static func convert(_ bw: String) -> String {
        var result = ""
        for ch in bw {
            result.append(map[ch] ?? ch)
        }
        return result
    }

}
