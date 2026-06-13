import Foundation

private struct CachePayload: Codable {
    let segments: [Segment]
    let rootIndex: [String: [String]]
    let refToRoot: [String: String]
}

public actor MorphologyCache {
    public static func loadOrBuild(csvPath: String, rootsPath: String) async throws -> MorphologyData {
        let cachePath = cacheFilePath()

        if FileManager.default.fileExists(atPath: cachePath) {
            if let cached = try? await loadFromCache(cachePath), await cached.segmentCount > 0 {
                return cached
            }
            try? FileManager.default.removeItem(atPath: cachePath)
        }

        let data = MorphologyData()
        try await data.load(csvPath: csvPath, rootsPath: rootsPath)
        let count = await data.segmentCount
        if count > 0 { try? await writeToCache(data, cachePath: cachePath) }
        return data
    }

    private static func cacheFilePath() -> String {
        guard let appSupport = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask).first else {
            return NSTemporaryDirectory() + "QuranNative/morphology.cache"
        }
        let cacheDir = appSupport.appendingPathComponent("QuranNative")
        try? FileManager.default.createDirectory(at: cacheDir, withIntermediateDirectories: true)
        return cacheDir.appendingPathComponent("morphology.cache").path
    }

    private static func writeToCache(_ data: MorphologyData, cachePath: String) async throws {
        let segments = await data.segments
        guard segments.count > 0 else { return }
        let payload = CachePayload(
            segments: segments,
            rootIndex: await data.rootIndex,
            refToRoot: await data.refToRoot
        )
        let encoder = JSONEncoder()
        let encoded = try encoder.encode(payload)
        try encoded.write(to: URL(fileURLWithPath: cachePath))
    }

    private static func loadFromCache(_ cachePath: String) async throws -> MorphologyData {
        let url = URL(fileURLWithPath: cachePath)
        let fileData = try Data(contentsOf: url)
        let decoder = JSONDecoder()
        let payload = try decoder.decode(CachePayload.self, from: fileData)
        let morphologyData = MorphologyData()
        await morphologyData.setSegmentsForCache(payload.segments,
                                                  rootIndex: payload.rootIndex,
                                                  refToRoot: payload.refToRoot)
        return morphologyData
    }
}
