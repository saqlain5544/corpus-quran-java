import Foundation

// RawXMLReader — pure passthrough of the Tanzil Uthmani XML.
//
// The source data already contains the correct marks at the correct
// positions (tajwid/recitation marks: ۟ ۥ ۦ ۢ ۜ, etc. plus 6922 total
// marks across 3386 verses).  We do NOT inject any additional marks —
// that would corrupt the source data.  If waqf stop signs (ج معانقة
// صلى etc.) are needed, they must be sourced from a Mushaf-specific
// data file, not invented here.

public final class RawXMLReader: NSObject, XMLParserDelegate, @unchecked Sendable {
    private var verseData: [Int: [Int: String]] = [:]
    private var bismillahData: [Int: String] = [:]
    private var currentSura: Int = 0
    private var currentVerses: [Int: String] = [:]

    public func load(url: URL) throws -> ([Int: [Int: String]], [Int: String]) {
        let parser = XMLParser(contentsOf: url)
        guard let parser else { throw JQuranTreeError.parseError("Cannot open: \(url.path)") }
        parser.delegate = self
        parser.parse()
        if let error = parser.parserError {
            throw JQuranTreeError.parseError(error.localizedDescription)
        }
        return (verseData, bismillahData)
    }

    public func parser(_ parser: XMLParser, didStartElement elementName: String,
                       namespaceURI: String?, qualifiedName qName: String?,
                       attributes: [String: String] = [:]) {
        switch elementName {
        case "sura":
            currentSura = Int(attributes["index"] ?? "0") ?? 0
            currentVerses = [:]
        case "aya":
            let ayaIndex = Int(attributes["index"] ?? "0") ?? 0
            let text = attributes["text"] ?? ""
            let bismillah = attributes["bismillah"] ?? ""
            currentVerses[ayaIndex] = text
            if ayaIndex == 1 && !bismillah.isEmpty {
                bismillahData[currentSura] = bismillah
            }
        default: break
        }
    }

    public func parser(_ parser: XMLParser, didEndElement elementName: String,
                       namespaceURI: String?, qualifiedName qName: String?) {
        if elementName == "sura", currentSura > 0 {
            verseData[currentSura] = currentVerses
        }
    }
}
