import Foundation

public actor TranslationReader {
    private var data: [Int: [Int: String]] = [:]
    public private(set) var isLoaded = false

    public init() {}

    public func load(path: String) throws {
        let url = URL(fileURLWithPath: path)
        let content = try String(contentsOf: url, encoding: .utf8)
        for line in content.split(separator: "\n", omittingEmptySubsequences: true) {
            let parts = line.split(separator: "|", maxSplits: 2, omittingEmptySubsequences: false)
            guard parts.count == 3,
                  let sura = Int(parts[0]),
                  let ayah = Int(parts[1]) else { continue }
            if data[sura] == nil { data[sura] = [:] }
            data[sura, default: [:]][ayah] = String(parts[2])
        }
        isLoaded = true
    }

    public func translation(sura: Int, ayah: Int) -> String? {
        data[sura]?[ayah]
    }

    public func translations(forPairs pairs: [(sura: Int, ayah: Int)]) -> [String: String] {
        var result: [String: String] = [:]
        result.reserveCapacity(pairs.count)
        for (s, a) in pairs {
            if let text = data[s]?[a] {
                result["\(s):\(a)"] = text
            }
        }
        return result
    }
}
