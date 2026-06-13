import Foundation

public enum JQuranTreeError: LocalizedError, Sendable {
    case invalidSura(sura: Int)
    case invalidSuraVerse(sura: Int, verse: Int)
    case unknownVariant(String)
    case parseError(String)
    case downloadFailed(String)
    case missingData(String)
    case cacheError(String)

    public var errorDescription: String? {
        switch self {
        case .invalidSura(let n):
            "Invalid surah number: \(n)"
        case .invalidSuraVerse(let s, let v):
            "Invalid reference: \(s):\(v)"
        case .unknownVariant(let key):
            "Unknown variant: \(key)"
        case .parseError(let msg):
            "Parse error: \(msg)"
        case .downloadFailed(let msg):
            "Download failed: \(msg)"
        case .missingData(let msg):
            "Missing data: \(msg)"
        case .cacheError(let msg):
            "Cache error: \(msg)"
        }
    }
}
