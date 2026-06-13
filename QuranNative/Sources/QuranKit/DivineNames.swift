import Foundation

public struct DivineName: Identifiable, Hashable, Sendable {
    public var id: Int { number }
    public let number: Int
    public let arabic: String
    public let transliteration: String
    public let meaning: String

    public init(number: Int, arabic: String, transliteration: String, meaning: String) {
        self.number = number; self.arabic = arabic; self.transliteration = transliteration; self.meaning = meaning
    }
}

public let DIVINE_NAMES: [DivineName] = [
    DivineName(number: 1, arabic: "الرَّحْمَنُ", transliteration: "ar-Raḥmān", meaning: "The Most Compassionate"),
    DivineName(number: 2, arabic: "الرَّحِيمُ", transliteration: "ar-Raḥīm", meaning: "The Most Merciful"),
    DivineName(number: 3, arabic: "الْمَلِكُ", transliteration: "al-Malik", meaning: "The Sovereign"),
    DivineName(number: 4, arabic: "الْقُدُّوسُ", transliteration: "al-Quddūs", meaning: "The Most Holy"),
    DivineName(number: 5, arabic: "السَّلَامُ", transliteration: "as-Salām", meaning: "The Source of Peace"),
    DivineName(number: 6, arabic: "الْمُؤْمِنُ", transliteration: "al-Muʾmin", meaning: "The Guardian of Faith"),
    DivineName(number: 7, arabic: "الْمُهَيْمِنُ", transliteration: "al-Muhaymin", meaning: "The Protector"),
    DivineName(number: 8, arabic: "الْعَزِيزُ", transliteration: "al-ʿAzīz", meaning: "The Almighty"),
    DivineName(number: 9, arabic: "الْجَبَّارُ", transliteration: "al-Jabbār", meaning: "The Compeller"),
    DivineName(number: 10, arabic: "الْمُتَكَبِّرُ", transliteration: "al-Mutakabbir", meaning: "The Supreme"),
    DivineName(number: 11, arabic: "الْخَالِقُ", transliteration: "al-Khāliq", meaning: "The Creator"),
    DivineName(number: 12, arabic: "الْبَارِئُ", transliteration: "al-Bāriʾ", meaning: "The Evolver"),
    DivineName(number: 13, arabic: "الْمُصَوِّرُ", transliteration: "al-Muṣawwir", meaning: "The Fashioner"),
    DivineName(number: 14, arabic: "الْغَفَّارُ", transliteration: "al-Ghaffār", meaning: "The Constant Forgiver"),
    DivineName(number: 15, arabic: "الْقَهَّارُ", transliteration: "al-Qahhār", meaning: "The All-Subduer"),
    DivineName(number: 16, arabic: "الْوَهَّابُ", transliteration: "al-Wahhāb", meaning: "The Bestower"),
    DivineName(number: 17, arabic: "الرَّزَّاقُ", transliteration: "ar-Razzāq", meaning: "The Provider"),
    DivineName(number: 18, arabic: "الْفَتَّاحُ", transliteration: "al-Fattāḥ", meaning: "The Opener"),
    DivineName(number: 19, arabic: "الْعَلِيمُ", transliteration: "al-ʿAlīm", meaning: "The All-Knowing"),
    DivineName(number: 20, arabic: "الْقَابِضُ", transliteration: "al-Qābiḍ", meaning: "The Constrictor"),
    DivineName(number: 21, arabic: "الْبَاسِطُ", transliteration: "al-Bāsiṭ", meaning: "The Expander"),
    DivineName(number: 22, arabic: "الْخَافِضُ", transliteration: "al-Khāfiḍ", meaning: "The Abaser"),
    DivineName(number: 23, arabic: "الرَّافِعُ", transliteration: "ar-Rāfiʿ", meaning: "The Exalter"),
    DivineName(number: 24, arabic: "الْمُعِزُّ", transliteration: "al-Muʿizz", meaning: "The Bestower of Honor"),
    DivineName(number: 25, arabic: "الْمُذِلُّ", transliteration: "al-Mudhill", meaning: "The Humiliator"),
    DivineName(number: 26, arabic: "السَّمِيعُ", transliteration: "as-Samīʿ", meaning: "The All-Hearing"),
    DivineName(number: 27, arabic: "الْبَصِيرُ", transliteration: "al-Baṣīr", meaning: "The All-Seeing"),
    DivineName(number: 28, arabic: "الْحَكَمُ", transliteration: "al-Ḥakam", meaning: "The Judge"),
    DivineName(number: 29, arabic: "الْعَدْلُ", transliteration: "al-ʿAdl", meaning: "The Just"),
    DivineName(number: 30, arabic: "اللَّطِيفُ", transliteration: "al-Laṭīf", meaning: "The Subtle"),
    DivineName(number: 31, arabic: "الْخَبِيرُ", transliteration: "al-Khabīr", meaning: "The All-Aware"),
    DivineName(number: 32, arabic: "الْحَلِيمُ", transliteration: "al-Ḥalīm", meaning: "The Forbearing"),
    DivineName(number: 33, arabic: "الْعَظِيمُ", transliteration: "al-ʿAẓīm", meaning: "The Magnificent"),
    DivineName(number: 34, arabic: "الْغَفُورُ", transliteration: "al-Ghafūr", meaning: "The Forgiving"),
    DivineName(number: 35, arabic: "الشَّكُورُ", transliteration: "ash-Shakūr", meaning: "The Appreciative"),
    DivineName(number: 36, arabic: "الْعَلِيُّ", transliteration: "al-ʿAlī", meaning: "The Most High"),
    DivineName(number: 37, arabic: "الْكَبِيرُ", transliteration: "al-Kabīr", meaning: "The Greatest"),
    DivineName(number: 38, arabic: "الْحَفِيظُ", transliteration: "al-Ḥafīẓ", meaning: "The Preserver"),
    DivineName(number: 39, arabic: "الْمُقِيتُ", transliteration: "al-Muqīt", meaning: "The Sustainer"),
    DivineName(number: 40, arabic: "الْحَسِيبُ", transliteration: "al-Ḥasīb", meaning: "The Reckoner"),
    DivineName(number: 41, arabic: "الْجَلِيلُ", transliteration: "al-Jalīl", meaning: "The Majestic"),
    DivineName(number: 42, arabic: "الْكَرِيمُ", transliteration: "al-Karīm", meaning: "The Generous"),
    DivineName(number: 43, arabic: "الرَّقِيبُ", transliteration: "ar-Raqīb", meaning: "The Watchful"),
    DivineName(number: 44, arabic: "الْمُجِيبُ", transliteration: "al-Mujīb", meaning: "The Responsive"),
    DivineName(number: 45, arabic: "الْوَاسِعُ", transliteration: "al-Wāsiʿ", meaning: "The All-Encompassing"),
    DivineName(number: 46, arabic: "الْحَكِيمُ", transliteration: "al-Ḥakīm", meaning: "The All-Wise"),
    DivineName(number: 47, arabic: "الْوَدُودُ", transliteration: "al-Wadūd", meaning: "The Loving"),
    DivineName(number: 48, arabic: "الْمَجِيدُ", transliteration: "al-Majīd", meaning: "The Glorious"),
    DivineName(number: 49, arabic: "الْبَاعِثُ", transliteration: "al-Bāʿith", meaning: "The Resurrector"),
    DivineName(number: 50, arabic: "الشَّهِيدُ", transliteration: "ash-Shahīd", meaning: "The Witness"),
    DivineName(number: 51, arabic: "الْحَقُّ", transliteration: "al-Ḥaqq", meaning: "The Truth"),
    DivineName(number: 52, arabic: "الْوَكِيلُ", transliteration: "al-Wakīl", meaning: "The Trustee"),
    DivineName(number: 53, arabic: "الْقَوِيُّ", transliteration: "al-Qawī", meaning: "The All-Strong"),
    DivineName(number: 54, arabic: "الْمَتِينُ", transliteration: "al-Matīn", meaning: "The Firm"),
    DivineName(number: 55, arabic: "الْوَلِيُّ", transliteration: "al-Walī", meaning: "The Protecting Friend"),
    DivineName(number: 56, arabic: "الْحَمِيدُ", transliteration: "al-Ḥamīd", meaning: "The Praiseworthy"),
    DivineName(number: 57, arabic: "الْمُحْصِي", transliteration: "al-Muḥṣī", meaning: "The Reckoner"),
    DivineName(number: 58, arabic: "الْمُبْدِئُ", transliteration: "al-Mubdiʾ", meaning: "The Originator"),
    DivineName(number: 59, arabic: "الْمُعِيدُ", transliteration: "al-Muʿīd", meaning: "The Restorer"),
    DivineName(number: 60, arabic: "الْمُحْيِي", transliteration: "al-Muḥyī", meaning: "The Giver of Life"),
    DivineName(number: 61, arabic: "الْمُمِيتُ", transliteration: "al-Mumīt", meaning: "The Bringer of Death"),
    DivineName(number: 62, arabic: "الْحَيُّ", transliteration: "al-Ḥayy", meaning: "The Ever-Living"),
    DivineName(number: 63, arabic: "الْقَيُّومُ", transliteration: "al-Qayyūm", meaning: "The Self-Sustaining"),
    DivineName(number: 64, arabic: "الْوَاجِدُ", transliteration: "al-Wājid", meaning: "The Finder"),
    DivineName(number: 65, arabic: "الْمَاجِدُ", transliteration: "al-Mājid", meaning: "The Illustrious"),
    DivineName(number: 66, arabic: "الْوَاحِدُ", transliteration: "al-Wāḥid", meaning: "The One"),
    DivineName(number: 67, arabic: "الْأَحَدُ", transliteration: "al-Aḥad", meaning: "The Unique"),
    DivineName(number: 68, arabic: "الصَّمَدُ", transliteration: "aṣ-Ṣamad", meaning: "The Self-Sufficient"),
    DivineName(number: 69, arabic: "الْقَادِرُ", transliteration: "al-Qādir", meaning: "The Able"),
    DivineName(number: 70, arabic: "الْمُقْتَدِرُ", transliteration: "al-Muqtadir", meaning: "The Powerful"),
    DivineName(number: 71, arabic: "الْمُقَدِّمُ", transliteration: "al-Muqaddim", meaning: "The Expediter"),
    DivineName(number: 72, arabic: "الْمُؤَخِّرُ", transliteration: "al-Muʾakhkhir", meaning: "The Delayer"),
    DivineName(number: 73, arabic: "الْأَوَّلُ", transliteration: "al-Awwal", meaning: "The First"),
    DivineName(number: 74, arabic: "الْآخِرُ", transliteration: "al-Ākhir", meaning: "The Last"),
    DivineName(number: 75, arabic: "الظَّاهِرُ", transliteration: "aẓ-Ẓāhir", meaning: "The Manifest"),
    DivineName(number: 76, arabic: "الْبَاطِنُ", transliteration: "al-Bāṭin", meaning: "The Hidden"),
    DivineName(number: 77, arabic: "الْوَالِي", transliteration: "al-Wālī", meaning: "The Governor"),
    DivineName(number: 78, arabic: "الْمُتَعَالِي", transliteration: "al-Mutaʿālī", meaning: "The Most Exalted"),
    DivineName(number: 79, arabic: "الْبَرُّ", transliteration: "al-Barr", meaning: "The Source of Goodness"),
    DivineName(number: 80, arabic: "التَّوَّابُ", transliteration: "at-Tawwāb", meaning: "The Acceptor of Repentance"),
    DivineName(number: 81, arabic: "الْمُنْتَقِمُ", transliteration: "al-Muntaqim", meaning: "The Avenger"),
    DivineName(number: 82, arabic: "الْعَفُوُّ", transliteration: "al-ʿAfuww", meaning: "The Pardoner"),
    DivineName(number: 83, arabic: "الرَّؤُوفُ", transliteration: "ar-Raʾūf", meaning: "The Kind"),
    DivineName(number: 84, arabic: "مَالِكُ الْمُلْكِ", transliteration: "Mālik al-Mulk", meaning: "Master of the Kingdom"),
    DivineName(number: 85, arabic: "ذُو الْجَلَالِ وَالْإِكْرَامِ", transliteration: "Dhū al-Jalāl wa-l-Ikrām", meaning: "Possessor of Glory and Honor"),
    DivineName(number: 86, arabic: "الْمُقْسِطُ", transliteration: "al-Muqsiṭ", meaning: "The Equitable"),
    DivineName(number: 87, arabic: "الْجَامِعُ", transliteration: "al-Jāmiʿ", meaning: "The Gatherer"),
    DivineName(number: 88, arabic: "الْغَنِيُّ", transliteration: "al-Ghanī", meaning: "The Self-Sufficient"),
    DivineName(number: 89, arabic: "الْمُغْنِي", transliteration: "al-Mughnī", meaning: "The Enricher"),
    DivineName(number: 90, arabic: "الْمَانِعُ", transliteration: "al-Māniʿ", meaning: "The Preventer"),
    DivineName(number: 91, arabic: "الضَّارُّ", transliteration: "aḍ-Ḍārr", meaning: "The Inflictor of Harm"),
    DivineName(number: 92, arabic: "النَّافِعُ", transliteration: "an-Nāfiʿ", meaning: "The Benefactor"),
    DivineName(number: 93, arabic: "النُّورُ", transliteration: "an-Nūr", meaning: "The Light"),
    DivineName(number: 94, arabic: "الْهَادِي", transliteration: "al-Hādī", meaning: "The Guide"),
    DivineName(number: 95, arabic: "الْبَدِيعُ", transliteration: "al-Badīʿ", meaning: "The Originator"),
    DivineName(number: 96, arabic: "الْبَاقِي", transliteration: "al-Bāqī", meaning: "The Everlasting"),
    DivineName(number: 97, arabic: "الْوَارِثُ", transliteration: "al-Wārith", meaning: "The Inheritor"),
    DivineName(number: 98, arabic: "الرَّشِيدُ", transliteration: "ar-Rashīd", meaning: "The Guide to the Right Path"),
    DivineName(number: 99, arabic: "الصَّبُورُ", transliteration: "aṣ-Ṣabūr", meaning: "The Patient"),
]

private let BY_NUMBER: [Int: DivineName] = Dictionary(DIVINE_NAMES.map { ($0.number, $0) }, uniquingKeysWith: { first, _ in first })
private let BY_ARABIC: [String: DivineName] = Dictionary(DIVINE_NAMES.map { ($0.arabic, $0) }, uniquingKeysWith: { first, _ in first })
private let BY_TRANSLIT: [String: DivineName] = Dictionary(DIVINE_NAMES.map { ($0.transliteration, $0) }, uniquingKeysWith: { first, _ in first })

public func byNumber(_ n: Int) -> DivineName? { BY_NUMBER[n] }
public func byArabic(_ text: String) -> DivineName? { BY_ARABIC[text] }
public func byTransliteration(_ name: String) -> DivineName? { BY_TRANSLIT[name] }
public func allArabicForms() -> [String] { DIVINE_NAMES.map(\.arabic) }

public func findInVerse(_ verseText: String) -> [DivineName] {
    let nfd = verseText.decomposedStringWithCanonicalMapping
    let bare = String(nfd.unicodeScalars.filter { (0x0621...0x064A).contains($0.value) || (0x0670...0x0671).contains($0.value) })
    let recomposed = bare.precomposedStringWithCompatibilityMapping
    var results: [DivineName] = []
    for name in DIVINE_NAMES {
        if recomposed.contains(name.arabic) { results.append(name) }
        else {
            let withoutAl = String(name.arabic.dropFirst(2))
            if !withoutAl.isEmpty && recomposed.contains(withoutAl) { results.append(name) }
        }
    }
    return results
}
