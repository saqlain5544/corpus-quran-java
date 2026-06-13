import Testing
import Foundation
import QuranKit

private func dataDir() -> String {
    let file = URL(fileURLWithPath: #file)
    var dir = file.deletingLastPathComponent()
    for _ in 0..<5 {
        let candidate = dir.appendingPathComponent("Sources/QuranApp/Resources/data").path
        if FileManager.default.fileExists(atPath: candidate) { return candidate }
        let altCandidate = dir.appendingPathComponent("QuranNative/Sources/QuranApp/Resources/data").path
        if FileManager.default.fileExists(atPath: altCandidate) { return altCandidate }
        dir = dir.deletingLastPathComponent()
    }
    let fallback = file.deletingLastPathComponent().deletingLastPathComponent().deletingLastPathComponent()
        .appendingPathComponent("Sources/QuranApp/Resources/data").path
    return fallback
}

private func requireCSV() throws -> String {
    let path = "\(dataDir())/morphology/MASAQ.csv"
    try #require(FileManager.default.fileExists(atPath: path), "MASAQ.csv not found at \(path)")
    return path
}

private func requireRootsJSON() throws -> String {
    let path = "\(dataDir())/morphology/corpus-roots.json"
    try #require(FileManager.default.fileExists(atPath: path), "corpus-roots.json not found at \(path)")
    return path
}

private func requireXML() throws -> String {
    let path = "\(dataDir())/quran-text/uthmani-full.xml"
    try #require(FileManager.default.fileExists(atPath: path), "uthmani-full.xml not found at \(path)")
    return path
}

private func requireFontsDir() throws -> String {
    let path = "\(dataDir())/../fonts"
    try #require(FileManager.default.fileExists(atPath: path), "fonts dir not found at \(path)")
    return path
}

@Test func testConstants() {
    #expect(verseCount(1) == 7)
    #expect(verseCount(114) == 6)
    #expect(surahName(1) == "Al-Fatiḥah")
    #expect(isValid(sura: 1, verse: 1) == true)
    #expect(isValid(sura: 999, verse: 1) == false)
}

@Test func testEncodingRoundTrip() {
    let text = "بسم"
    let decoded = decode(text)
    #expect(decoded.count == 3)
    for ch in decoded { #expect(ch.diacritics.rawValue == 0) }
    let encoded = encode(decoded)
    #expect(encoded == text)
}

@Test func testMASAQLoading() async throws {
    let csvPath = try requireCSV()
    let rootsPath = try requireRootsJSON()

    let morph = try await MorphologyCache.loadOrBuild(csvPath: csvPath, rootsPath: rootsPath)

    let segCount = await morph.segmentCount
    #expect(segCount >= 150_000, "Expected at least 150k segments, got \(segCount)")

    let wc = await morph.wordCount
    #expect(wc >= 75_000, "Expected at least 75k words, got \(wc)")

    let segs = await morph.word(sura: 1, verse: 1, wordNum: 1)
    #expect(!segs.isEmpty)
    #expect(segs.first?.pos == "PREP")

    let search = await morph.searchGloss("mercy")
    #expect(!search.isEmpty)
}

@Test func testReaderMarksPreserved() async throws {
    let xmlPath = try requireXML()

    let reader = Reader(xmlPath: xmlPath)
    try await reader.load()

    let text = await reader.verse(sura: 2, aya: 5)
    let hasPauseMarks = text.unicodeScalars.contains { (0x06DC...0x06ED).contains($0.value) }
    #expect(hasPauseMarks, "uthmani-full.xml should contain Quranic pause marks")
}

@Test func testReaderLoadsAllSurahs() async throws {
    let xmlPath = try requireXML()

    let reader = Reader(xmlPath: xmlPath)
    try await reader.load()

    let suras = await reader.suraNumbers
    #expect(suras.count == 114)
    #expect(!(await reader.verse(sura: 1, aya: 1)).isEmpty)
    #expect(!(await reader.verse(sura: 114, aya: 6)).isEmpty)
}

@Test func testMaddaAndPauseMarksInVerse() async throws {
    let xmlPath = try requireXML()

    let reader = Reader(xmlPath: xmlPath)
    try await reader.load()

    let text = await reader.verse(sura: 2, aya: 1)
    let hasMadda = text.unicodeScalars.contains { $0.value == 0x0653 }
    #expect(hasMadda, "Sura 2 v1 should contain madda (U+0653) — got: \(text)")

    let text56 = await reader.verse(sura: 2, aya: 5)
    let hasPause = text56.unicodeScalars.contains { $0.value == 0x06DF }
    #expect(hasPause, "Sura 2 v5 should contain pause mark (U+06DF) — got first 100: \(String(text56.prefix(100)))")
}

@Test func testBismillahSeparation() async throws {
    let xmlPath = try requireXML()

    let reader = Reader(xmlPath: xmlPath)
    try await reader.load()

    let v1 = await reader.verse(sura: 1, aya: 1)
    let bis1 = await reader.bismillah(sura: 1)
    #expect(v1.contains("بِسْمِ"), "Al-Fatiha v1 should start with basmala")
    #expect(bis1 == nil, "Al-Fatiha v1 basmala is part of the verse text, not separate")

    let v2_1 = await reader.verse(sura: 2, aya: 1)
    let bis2 = await reader.bismillah(sura: 2)
    #expect(!v2_1.isEmpty, "Sura 2 verse 1 should have text (the alif-lam-mim)")
    #expect(bis2 != nil, "Sura 2 should have separate bismillah")
    #expect(bis2?.contains("بِسْمِ") == true, "Bismillah should start with basmala")
    #expect(!v2_1.contains("بِسْمِ"), "Sura 2 verse 1 text should NOT contain basmala")

    let v9_1 = await reader.verse(sura: 9, aya: 1)
    let bis9 = await reader.bismillah(sura: 9)
    #expect(!v9_1.isEmpty)
    #expect(bis9 == nil, "Sura 9 (At-Tawba) has no bismillah")
}

@Test func testRenderArabicVerseProducesGlyphs() async throws {
    let xmlPath = try requireXML()

    let reader = Reader(xmlPath: xmlPath)
    try await reader.load()

    let verse = await reader.verse(sura: 2, aya: 1)
    let hasMadda = verse.contains("\u{0653}")
    #expect(hasMadda, "Verse 2:1 should contain madda — got: \(verse)")
}

@Test func testMASAQNullHandling() async throws {
    let csvPath = try requireCSV()
    let rootsPath = try requireRootsJSON()

    let morph = try await MorphologyCache.loadOrBuild(csvPath: csvPath, rootsPath: rootsPath)

    let segs = await morph.word(sura: 2, verse: 21, wordNum: 1)
    let seg = segs.first(where: { $0.segment == 3 })
    #expect(seg != nil, "Should find segment 3 in 2:21:1")
    if let seg = seg {
        #expect(seg.pos.isEmpty, "POS should be empty (was 'None' in CSV), got: '\(seg.pos)'")
        #expect(seg.segmentedText.isEmpty, "Segmented text should be empty, got: '\(seg.segmentedText)'")
        #expect(seg.morphType == "Other_i3rab", "Morph_Type should be preserved, got: '\(seg.morphType)'")
        #expect(seg.gloss == "o", "Gloss should be 'o', got: '\(seg.gloss)'")
    }

    let segs2 = await morph.word(sura: 1, verse: 1, wordNum: 1)
    if let s = segs2.first {
        #expect(s.caseMood != "None", "caseMood should not be literal 'None'")
    }

    let segs3 = await morph.word(sura: 1, verse: 4, wordNum: 2)
    let parenSeg = segs3.first(where: { $0.segment == 2 })
    if let parenSeg = parenSeg {
        #expect(parenSeg.segmentedText != "(null)", "Segmented text should not be literal '(null)', got: '\(parenSeg.segmentedText)'")
        #expect(parenSeg.segmentedText.isEmpty, "(null) should be cleaned to empty, got: '\(parenSeg.segmentedText)'")
    }
}

@Test func testFontRegistration() async throws {
    let fontsDir = try requireFontsDir()

    let amiriQuranURL = "\(fontsDir)/AmiriQuran.ttf"
    let exists = FileManager.default.fileExists(atPath: amiriQuranURL)
    #expect(exists, "AmiriQuran font should exist")
}

@Test func testWordIndicesAfterBismillahFix() async throws {
    let csvPath = try requireCSV()
    let rootsPath = try requireRootsJSON()
    let xmlPath = try requireXML()

    let reader = Reader(xmlPath: xmlPath)
    try await reader.load()
    let morph = try await MorphologyCache.loadOrBuild(csvPath: csvPath, rootsPath: rootsPath)

    let text = await reader.verse(sura: 2, aya: 1)
    let words = text.split(separator: " ")
    #expect(words.count == 1, "Sura 2 v1 should have 1 word (الٓمٓ), got \(words.count): \(text)")

    let segs = await morph.word(sura: 2, verse: 1, wordNum: 1)
    #expect(segs.count == 1, "MASAQ should have 1 segment for 2:1:1, got \(segs.count)")

    let segs22_1 = await morph.word(sura: 2, verse: 2, wordNum: 1)
    #expect(!segs22_1.isEmpty, "2:2:1 should have segments (got 0 — likely word index mismatch)")
    #expect(segs22_1.first?.gloss == "that", "2:2:1 should be 'ذَلِكَ' (that), got '\(segs22_1.first?.gloss ?? "nil")'")
}
