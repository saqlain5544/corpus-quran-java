import Foundation

public actor Reader {
    private var verseData: [Int: [Int: String]] = [:]
    private var bismillahData: [Int: String] = [:]
    private let xmlPath: String
    public private(set) var isLoaded = false

    public init(xmlPath: String) {
        self.xmlPath = xmlPath
    }

    public func load() async throws {
        let url = URL(fileURLWithPath: xmlPath)
        // Try the uthmani-waqf.txt first (proper Madinah Mushaf marks
        // from quran.com).  Fall back to the XML if the .txt is missing.
        let txtPath = xmlPath.replacingOccurrences(of: "uthmani-full.xml", with: "uthmani-waqf.txt")
        if FileManager.default.fileExists(atPath: txtPath) {
            try loadTextFile(path: txtPath)
        } else {
            let rawReader = RawXMLReader()
            let (verses, bismillah) = try rawReader.load(url: url)
            verseData = verses
            bismillahData = bismillah
        }
        isLoaded = true
    }

    private func loadTextFile(path: String) throws {
        let url = URL(fileURLWithPath: path)
        let content = try String(contentsOf: url, encoding: .utf8)
        for line in content.split(separator: "\n", omittingEmptySubsequences: true) {
            let parts = line.split(separator: "|", maxSplits: 2, omittingEmptySubsequences: false)
            guard parts.count == 3,
                  let sura = Int(parts[0]),
                  let ayah = Int(parts[1]) else { continue }
            let text = String(parts[2])
            if verseData[sura] == nil { verseData[sura] = [:] }
            verseData[sura, default: [:]][ayah] = text
        }
    }

    public func verse(sura: Int, aya: Int) -> String {
        verseData[sura]?[aya] ?? ""
    }

    // Bulk-fetch verse texts for a list of (sura, ayah) pairs.
    // Avoids 6,236+ individual actor calls.
    public func verses(forPairs pairs: [(sura: Int, ayah: Int)]) -> [String: String] {
        var result: [String: String] = [:]
        result.reserveCapacity(pairs.count)
        for (s, a) in pairs {
            if let text = verseData[s]?[a] {
                result["\(s):\(a)"] = text
            }
        }
        return result
    }

    public func bismillah(sura: Int) -> String? {
        bismillahData[sura]
    }

    public var suraNumbers: [Int] { Array(verseData.keys).sorted() }

    public func verseCount(_ sura: Int) -> Int {
        verseData[sura]?.count ?? 0
    }
}

