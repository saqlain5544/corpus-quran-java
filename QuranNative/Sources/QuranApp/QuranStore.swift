import SwiftUI
import QuranKit

@MainActor
@Observable
public final class QuranStore {
    var morphology: MorphologyData?
    var reader: Reader?
    var quran: Quran?
    var translationReader: TranslationReader?

    var selectedSura: Int = UserDefaults.standard.integer(forKey: "lastSura").nonZero ?? 2
    var lastAyah: Int = UserDefaults.standard.integer(forKey: "lastAyah").nonZero ?? 1
    var bismillahText: String = ""
    var verseEntries: [VerseEntry] = []
    var selectedVerse: Int? = nil
    var selectedWordIndex: Int? = nil

    // Monotonically increasing counter — ReaderView watches this and
    // scrolls the matching verse into view. Incrementing it (rather
    // than mutating an Int verse) ensures onChange fires even when
    // navigating to the same verse twice in a row.
    var scrollToAyah: Int = 0
    var pendingScrollVerse: Int? = nil

    // Reader section — which main tab is active when the user is
    // navigating. Tapping a concordance ayah switches to the Reader
    // and jumps to that ayah.
    var readerSection: AppSection = .reader

    var isLoaded = false
    var loadError: String?
    var isLoading = false
    var loadingProgress: Double = 0.0

    var searchQuery: String = ""
    var searchResults: [ConcordanceLine] = []
    var isSearching = false
    var searchMode: SearchMode = .word
    var searchDurationMs: Int = 0   // for "X results in Yms" display

    private struct CacheKey: Hashable { let query: String; let mode: SearchMode }
    private static let cacheLimit = 64
    private var resultCache: [CacheKey: [ConcordanceLine]] = [:]
    private var cacheMRU: [CacheKey] = []

    var bookmarkedVerses: Set<String> = []  // "sura:ayah" keys
    var readingHistory: [(sura: Int, ayah: Int, date: Date)] = []

    enum SearchMode: String, CaseIterable {
        case word = "Word"
        case gloss = "Gloss"
        case root = "Root"
    }

    struct VerseEntry: Identifiable, Sendable {
        let id: Int
        let number: Int
        let text: String
        let translation: String
        let wordAnalyses: [WordAnalysis]
    }

    func load() async {
        guard !isLoaded else { return }
        isLoading = true
        defer { isLoading = false }

        FontManager.shared.registerAll(bundle: .module)

        guard let dataDir = Bundle.module.resourcePath.map({ "\($0)/data" })
                ?? Bundle.main.resourcePath.map({ "\($0)/data" }) else {
            loadError = "Cannot find data directory"
            return
        }

        let csvPath = "\(dataDir)/morphology/MASAQ.csv"
        let rootsPath = "\(dataDir)/morphology/corpus-roots.json"
        let xmlPath = "\(dataDir)/quran-text/uthmani-full.xml"

        guard FileManager.default.fileExists(atPath: csvPath) else {
            loadError = "MASAQ.csv not found at: \(csvPath)"
            return
        }

        do {
            loadingProgress = 0.1
            let morphData = try await MorphologyCache.loadOrBuild(csvPath: csvPath, rootsPath: rootsPath)
            loadingProgress = 0.5
            morphology = morphData

            let r = Reader(xmlPath: xmlPath)
            if FileManager.default.fileExists(atPath: xmlPath) {
                try await r.load()
            }
            loadingProgress = 0.85
            reader = r

            quran = Quran(morphology: morphData, reader: r)
            loadingProgress = 1.0
            isLoaded = true

            let translationPath = "\(dataDir)/quran-text/en.sahih.csv"
            if FileManager.default.fileExists(atPath: translationPath) {
                let tr = TranslationReader()
                try? await tr.load(path: translationPath)
                translationReader = tr
            }

            try? await Task.sleep(for: .milliseconds(100))
            loadBookmarks()
            restoreReadingPosition()
            await loadSurah()
        } catch {
            loadError = error.localizedDescription
        }
    }

    func loadSurah() async {
        guard let q = quran, let r = reader else { return }
        bismillahText = (selectedSura > 1) ? (await r.bismillah(sura: selectedSura) ?? "") : ""

        let verseCount = await r.verseCount(selectedSura)
        let pairs = (1...verseCount).map { (sura: selectedSura, ayah: $0) }
        let texts = await r.verses(forPairs: pairs)
        let translations = await translationReader?.translations(forPairs: pairs) ?? [:]

        var entries: [VerseEntry] = []
        entries.reserveCapacity(verseCount)
        for v in 1...verseCount {
            let text = texts["\(selectedSura):\(v)"] ?? ""
            let analyses = await q.words(sura: selectedSura, aya: v)
            let translation = translations["\(selectedSura):\(v)"] ?? ""
            entries.append(VerseEntry(id: v, number: v, text: text, translation: translation, wordAnalyses: analyses))
        }
        verseEntries = entries
        saveReadingPosition()
    }

    // Navigate to a surah only (scrolls to top)
    func navigateTo(sura: Int) async {
        if sura != selectedSura {
            selectedSura = sura
            selectedVerse = nil
            selectedWordIndex = nil
            await loadSurah()
        }
        scrollToAyah &+= 1
        pendingScrollVerse = 1
    }

    // Navigate to a specific ayah. Switches section to Reader, loads
    // the surah if necessary, then triggers a scroll to the ayah.
    func navigateTo(sura: Int, ayah: Int) async {
        readerSection = .reader
        selectedSura = sura
        selectedVerse = ayah
        selectedWordIndex = nil
        await loadSurah()
        scrollToAyah &+= 1
        pendingScrollVerse = ayah
    }

    func selectVerse(_ v: Int) {
        selectedVerse = v
        selectedWordIndex = nil
    }

    func selectWord(_ idx: Int?, in verse: Int) {
        selectedWordIndex = idx
        selectedVerse = verse
    }

    func search(_ query: String, mode: SearchMode = .word) async {
        let trimmed = query.trimmingCharacters(in: .whitespaces)
        guard let m = morphology, !trimmed.isEmpty else {
            searchResults = []
            searchDurationMs = 0
            return
        }
        let cacheKey = CacheKey(query: trimmed, mode: mode)
        if let cached = resultCache[cacheKey] {
            searchResults = cached
            cacheMRU.removeAll { $0 == cacheKey }
            cacheMRU.insert(cacheKey, at: 0)
            searchDurationMs = 0
            return
        }

        isSearching = true
        defer { isSearching = false }
        let start = Date()

        let conc = Concordancer(morphology: m)
        var results: [ConcordanceLine] = []
        switch mode {
        case .word:
            results = await conc.searchWord(trimmed, context: 2)
        case .gloss:
            results = await conc.searchGloss(trimmed, context: 2)
        case .root:
            results = await conc.searchRoot(trimmed, context: 2)
        }
        let elapsed = Int(Date().timeIntervalSince(start) * 1000)

        resultCache[cacheKey] = results
        cacheMRU.insert(cacheKey, at: 0)
        if cacheMRU.count > Self.cacheLimit {
            let evicted = cacheMRU.removeLast()
            resultCache[evicted] = nil
        }

        searchResults = results
        searchDurationMs = elapsed
    }

    func saveReadingPosition() {
        let defaults = UserDefaults.standard
        defaults.set(selectedSura, forKey: "lastSura")
        defaults.set(selectedVerse ?? 1, forKey: "lastAyah")
    }

    func restoreReadingPosition() {
        let sura = UserDefaults.standard.integer(forKey: "lastSura").nonZero ?? 2
        let ayah = UserDefaults.standard.integer(forKey: "lastAyah").nonZero ?? 1
        selectedSura = sura
        lastAyah = ayah
    }

    func toggleBookmark(sura: Int, ayah: Int) {
        let key = "\(sura):\(ayah)"
        if bookmarkedVerses.contains(key) {
            bookmarkedVerses.remove(key)
        } else {
            bookmarkedVerses.insert(key)
        }
        UserDefaults.standard.set(Array(bookmarkedVerses), forKey: "bookmarks")
    }

    func isBookmarked(sura: Int, ayah: Int) -> Bool {
        bookmarkedVerses.contains("\(sura):\(ayah)")
    }

    func loadBookmarks() {
        if let saved = UserDefaults.standard.stringArray(forKey: "bookmarks") {
            bookmarkedVerses = Set(saved)
        }
    }
}

extension Int {
    var nonZero: Int? { self == 0 ? nil : self }
}

