public enum CharacterType: UInt8, Sendable, CaseIterable {
    case hamza              = 0
    case alifWithMadda      = 1
    case alifWithHamzaAbove = 2
    case wawWithHamza       = 3
    case alifWithHamzaBelow = 4
    case yaWithHamza        = 5
    case alif               = 6
    case ba                 = 7
    case taMarbuta          = 8
    case ta                 = 9
    case tha                = 10
    case jeem               = 11
    case hha                = 12
    case kha                = 13
    case dal                = 14
    case thal               = 15
    case ra                 = 16
    case zai                = 17
    case seen               = 18
    case sheen              = 19
    case sad                = 20
    case dad                = 21
    case tta                = 22
    case zza                = 23
    case ain                = 24
    case ghain              = 25
    case fa                 = 26
    case qaf                = 27
    case kaf                = 28
    case lam                = 29
    case meem               = 30
    case noon               = 31
    case ha                 = 32
    case waw                = 33
    case alifMaksura        = 34
    case ya                 = 35
    case alifWasla          = 36
    case tatweel            = 37
    case daggerAlif         = 38
    case smallHighSadLamAlef   = 39
    case smallHighQafLamAlef   = 40
    case smallHighMeemInit     = 41
    case smallHighLamAlef      = 42
    case smallHighJeem         = 43
    case smallHighThreeDots    = 44
    case smallHighSeen         = 45
    case smallHighRoundedZero  = 46
    case smallHighRectZero     = 47
    case smallHighMeem         = 48
    case smallLowSeen          = 49
    case smallWaw              = 50
    case smallYeh              = 51
    case smallHighYeh          = 52
    case smallHighNoon         = 53
    case emptyCentreLowStop    = 54
    case emptyCentreHighStop   = 55
    case roundedHighStop       = 56
    case smallLowMeem          = 57
}

public struct DiacriticTypes: OptionSet, Hashable, Sendable {
    public let rawValue: UInt16
    public init(rawValue: UInt16) { self.rawValue = rawValue }

    public static let fatha       = DiacriticTypes(rawValue: 1 << 0)
    public static let damma       = DiacriticTypes(rawValue: 1 << 1)
    public static let kasra       = DiacriticTypes(rawValue: 1 << 2)
    public static let shadda      = DiacriticTypes(rawValue: 1 << 3)
    public static let sukun       = DiacriticTypes(rawValue: 1 << 4)
    public static let tanweenFatha = DiacriticTypes(rawValue: 1 << 5)
    public static let tanweenDamma = DiacriticTypes(rawValue: 1 << 6)
    public static let tanweenKasra = DiacriticTypes(rawValue: 1 << 7)
    public static let maddahAbove  = DiacriticTypes(rawValue: 1 << 8)
    public static let hamzaAbove   = DiacriticTypes(rawValue: 1 << 9)
}

public struct ArabicCharacter: Hashable, Sendable {
    public let charType: CharacterType
    public var diacritics: DiacriticTypes

    public init(charType: CharacterType, diacritics: DiacriticTypes = []) {
        self.charType = charType
        self.diacritics = diacritics
    }
}
