import Foundation

public struct Location: Hashable, Codable, Sendable {
    public let sura: Int
    public let verse: Int
    public let token: Int

    public init(sura: Int, verse: Int, token: Int) {
        self.sura = sura
        self.verse = verse
        self.token = token
    }
}

public enum IRab: Int, CaseIterable, Sendable {
    case none = 0
    case invariable = 1
    case nominative = 2
    case accusative = 3
    case genitive = 4
    case jussive = 5
}

public enum Declinability: Int, CaseIterable, Sendable {
    case none = 0
    case invar = 1
    case decln = 2
    case jointPron = 3
    case defArt = 4
}

public struct Segment: Identifiable, Hashable, Codable, Sendable {
    public let id: Int
    public let sura: Int
    public let verse: Int
    public let word: Int
    public let segment: Int
    public let wordText: String
    public let withoutDiacritics: String
    public let segmentedText: String
    public let pos: String
    public let morphType: String
    public let syntacticRole: String
    public let possessive: String
    public let caseMood: String
    public let caseMarker: String
    public let phrase: String
    public let phraseFunction: String
    public let gloss: String
    public let punctuation: String
    public let declinability: String

    public init(
        id: Int, sura: Int, verse: Int, word: Int, segment: Int,
        wordText: String, withoutDiacritics: String, segmentedText: String,
        pos: String, morphType: String, syntacticRole: String,
        possessive: String, caseMood: String, caseMarker: String,
        phrase: String, phraseFunction: String, gloss: String,
        punctuation: String = "", declinability: String = ""
    ) {
        self.id = id
        self.sura = sura
        self.verse = verse
        self.word = word
        self.segment = segment
        self.wordText = wordText
        self.withoutDiacritics = withoutDiacritics
        self.segmentedText = segmentedText
        self.pos = pos
        self.morphType = morphType
        self.syntacticRole = syntacticRole
        self.possessive = possessive
        self.caseMood = caseMood
        self.caseMarker = caseMarker
        self.phrase = phrase
        self.phraseFunction = phraseFunction
        self.gloss = gloss
        self.punctuation = punctuation
        self.declinability = declinability
    }

    public var isPrefix: Bool { morphType == "Prefix" }
    public var isStem: Bool { morphType == "Stem" }
    public var isSuffix: Bool { morphType == "Suffix" }

    public var irab: IRab {
        switch caseMood {
        case "INVARIABLE": return .invariable
        case "NOMINATIVE": return .nominative
        case "ACCUSATIVE": return .accusative
        case "GENITIVE": return .genitive
        case "JUSSIVE": return .jussive
        default: return .none
        }
    }

    public var declType: Declinability {
        switch declinability {
        case "INVAR": return .invar
        case "DECLN": return .decln
        case "JONT_PRON": return .jointPron
        case "DEF_ART": return .defArt
        default: return .none
        }
    }
}

private let IRAB_MAP: [String: IRab] = [
    "INVARIABLE": .invariable,
    "NOMINATIVE": .nominative,
    "ACCUSATIVE": .accusative,
    "GENITIVE": .genitive,
    "JUSSIVE": .jussive,
]

private let DECLINABILITY_MAP: [String: Declinability] = [
    "INVAR": .invar,
    "DECLN": .decln,
    "JONT_PRON": .jointPron,
    "DEF_ART": .defArt,
]
