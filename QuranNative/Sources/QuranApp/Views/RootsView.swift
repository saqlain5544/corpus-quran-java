import SwiftUI
import QuranKit

struct RootsView: View {
    @Environment(QuranStore.self) private var store

    @State private var allRoots: [RootEntry] = []
    @State private var filtered: [RootEntry] = []
    @State private var searchText: String = ""
    @State private var debouncedSearch: String = ""
    @State private var debounceTask: Task<Void, Never>?

    @State private var sortMode: SortMode = .frequency
    @State private var minOccurrences: Double = 1
    @State private var maxOccurrences: Double = 3000
    @State private var selectedRoot: String?
    @State private var isLoading = true
    @State private var loadError: String?

    struct RootEntry: Identifiable {
        let root: String
        let rootArabic: String
        let ayahCount: Int
        let totalOccurrences: Int
        var id: String { root }
    }

    enum SortMode: String, CaseIterable, Identifiable {
        case frequency = "Frequency"
        case alphabetical = "A → Z"
        case arabicAlphabetical = "عربية"
        var id: String { rawValue }
    }

    var body: some View {
        // Use a horizontal split that fills all available space.
        // NavigationSplitView's sidebar detail layout expands to fill
        // the parent container; HSplitView was being constrained by
        // its children's ideal widths.
        HStack(spacing: 0) {
            rootsList
                .frame(minWidth: 280, idealWidth: 340, maxWidth: 420)
                .background(Color(nsColor: .controlBackgroundColor))
            Divider()
            detail
                .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .onAppear { loadRoots() }
        // Debounce search input by 150ms so we don't filter on every keystroke
        .onChange(of: searchText) { _, new in
            debounceTask?.cancel()
            debounceTask = Task {
                try? await Task.sleep(for: .milliseconds(150))
                if Task.isCancelled { return }
                if searchText == new {
                    await MainActor.run {
                        debouncedSearch = new
                        applyFilter()
                    }
                }
            }
        }
    }

    @ViewBuilder
    private var detail: some View {
        if isLoading {
            VStack(spacing: 12) {
                ProgressView()
                Text("Loading 1,642 roots…")
                    .foregroundStyle(.secondary)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        } else if let err = loadError {
            ContentUnavailableView(
                "Failed to load roots",
                systemImage: "exclamationmark.triangle",
                description: Text(err)
            )
        } else if let root = selectedRoot {
            RootConcordanceView(root: root)
                .id(root)
        } else {
            ContentUnavailableView(
                "Roots",
                systemImage: "leaf.fill",
                description: Text("Pick a root from the list to see every ayah where it appears")
            )
        }
    }

    private var rootsList: some View {
        VStack(spacing: 0) {
            // ── Header: search + sort + range filter ──
            VStack(spacing: 8) {
                HStack {
                    Image(systemName: "magnifyingglass")
                        .foregroundStyle(.secondary)
                    TextField("Search roots (kwn, علم, k…)", text: $searchText)
                        .textFieldStyle(.plain)
                    if !searchText.isEmpty {
                        Button { searchText = "" } label: {
                            Image(systemName: "xmark.circle.fill")
                        }
                        .buttonStyle(.borderless)
                        .foregroundStyle(.secondary)
                    }
                }
                .padding(8)
                .background(.quaternary, in: RoundedRectangle(cornerRadius: 6))

                HStack {
                    Picker("Sort", selection: $sortMode) {
                        ForEach(SortMode.allCases) { mode in
                            Text(mode.rawValue).tag(mode)
                        }
                    }
                    .pickerStyle(.segmented)
                    .frame(maxWidth: 240)
                    .onChange(of: sortMode) { _, _ in applyFilter() }
                    Spacer()
                    Text("\(filtered.count) of \(allRoots.count)")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }

                HStack {
                    Text("\(Int(minOccurrences))–\(Int(maxOccurrences))")
                        .font(.caption.monospaced())
                        .foregroundStyle(.secondary)
                        .frame(width: 80, alignment: .leading)
                    RangeSlider(minValue: 1, maxValue: Double(Swift.max(1, maxAllowedOccurrences)),
                                lowValue: $minOccurrences, highValue: $maxOccurrences)
                        .onChange(of: minOccurrences) { _, _ in applyFilter() }
                        .onChange(of: maxOccurrences) { _, _ in applyFilter() }
                }
            }
            .padding(10)

            Divider()

            List(selection: $selectedRoot) {
                ForEach(filtered) { entry in
                    RootRow(entry: entry).tag(entry.root)
                }
            }
            .listStyle(.plain)
        }
    }

    // The maximum occurrence count seen in the data, used as the
    // upper bound of the range slider.  Updated when loading.
    @State private var maxAllowedOccurrences: Int = 3000

    private func loadRoots() {
        guard let morph = store.morphology else {
            loadError = "Morphology data not loaded"
            isLoading = false
            return
        }
        Task {
            // Single actor call → all root stats at once
            let stats = await morph.allRootStats()
            // Sort by frequency to compute the maxAllowedOccurrences
            // once (no per-keystroke recomputation).
            let maxOcc = stats.map(\.totalOccurrences).max() ?? 3000
            await MainActor.run {
                self.maxAllowedOccurrences = maxOcc
                self.maxOccurrences = Double(maxOcc)
                self.allRoots = stats.map { s in
                    RootEntry(
                        root: s.root,
                        rootArabic: BuckwalterArabic.convert(s.root),
                        ayahCount: s.ayahCount,
                        totalOccurrences: s.totalOccurrences
                    )
                }
                self.isLoading = false
                self.applyFilter()
            }
        }
    }

    private func applyFilter() {
        let q = debouncedSearch.trimmingCharacters(in: .whitespaces).lowercased()
        let qArabic = debouncedSearch.trimmingCharacters(in: .whitespaces)
        let lo = Int(minOccurrences), hi = Int(maxOccurrences)

        var items = allRoots

        if !q.isEmpty {
            items = items.filter {
                $0.root.lowercased().contains(q) ||
                $0.rootArabic.contains(qArabic)
            }
        }

        items = items.filter { $0.totalOccurrences >= lo && $0.totalOccurrences <= hi }

        switch sortMode {
        case .frequency:
            items.sort { $0.totalOccurrences > $1.totalOccurrences }
        case .alphabetical:
            items.sort { $0.root.lowercased() < $1.root.lowercased() }
        case .arabicAlphabetical:
            items.sort { $0.rootArabic < $1.rootArabic }
        }
        filtered = items
    }
}

private struct RootRow: View {
    let entry: RootsView.RootEntry

    var body: some View {
        HStack(alignment: .center, spacing: 8) {
            VStack(alignment: .leading, spacing: 2) {
                Text(entry.rootArabic)
                    .font(.system(size: 22, design: .serif))
                    .foregroundStyle(.primary)
                Text(entry.root)
                    .font(.system(size: 10, design: .monospaced))
                    .foregroundStyle(.secondary)
            }
            Spacer()
            VStack(alignment: .trailing, spacing: 1) {
                Text("\(entry.ayahCount)")
                    .font(.system(size: 14, weight: .semibold, design: .monospaced))
                Text("ayahs")
                    .font(.caption2)
                    .foregroundStyle(.secondary)
                Text("\(entry.totalOccurrences)×")
                    .font(.caption2.monospaced())
                    .foregroundStyle(.blue)
            }
        }
    }
}

// MARK: - Root Concordance

struct RootConcordanceView: View {
    @Environment(QuranStore.self) private var store
    let root: String

    @State private var verses: [VerseRef] = []
    @State private var verseTexts: [String: String] = [:]
    @State private var words: [String: [WordOccurrence]] = [:]
    @State private var surfaceForms: [String] = []
    @State private var isLoading = true

    struct VerseRef: Identifiable, Hashable {
        let sura: Int
        let ayah: Int
        var id: String { "\(sura):\(ayah)" }
    }

    struct WordOccurrence: Identifiable {
        let word: Int
        let surface: String
        var id: String { "\(word)" }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            HStack(alignment: .firstTextBaseline, spacing: 12) {
                VStack(alignment: .leading, spacing: 2) {
                    Text(BuckwalterArabic.convert(root))
                        .font(.system(size: 36, design: .serif))
                        .foregroundStyle(.blue)
                    Text(root)
                        .font(.system(size: 14, design: .monospaced))
                        .foregroundStyle(.secondary)
                }
                Spacer()
                VStack(alignment: .trailing, spacing: 2) {
                    Text("\(verses.count) ayahs")
                        .font(.callout)
                    Text("\(totalOccurrences) occurrences")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
                Button {
                    let lines = verses.map { "\($0.sura):\($0.ayah)" }.joined(separator: "\n")
                    let pb = NSPasteboard.general
                    pb.clearContents()
                    pb.setString(lines, forType: .string)
                } label: {
                    Image(systemName: "doc.on.doc")
                }
                .help("Copy ayah list to clipboard")
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 12)
            .background(.quaternary)

            if !surfaceForms.isEmpty {
                Divider()
                HStack(alignment: .firstTextBaseline, spacing: 8) {
                    Text("Forms")
                        .font(.caption.weight(.semibold))
                        .foregroundStyle(.secondary)
                        .frame(width: 50, alignment: .leading)
                    ScrollView(.horizontal, showsIndicators: false) {
                        HStack(spacing: 6) {
                            ForEach(surfaceForms, id: \.self) { form in
                                Text(form)
                                    .font(.system(size: 20, design: .serif))
                                    .padding(.horizontal, 8)
                                    .padding(.vertical, 3)
                                    .background(.quaternary, in: Capsule())
                            }
                        }
                    }
                }
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
            }

            Divider()

            if isLoading {
                ProgressView()
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else if verses.isEmpty {
                ContentUnavailableView(
                    "No occurrences",
                    systemImage: "leaf",
                    description: Text("This root has no occurrences in the data")
                )
            } else {
                List(verses) { verse in
                    ConcordanceRow(
                        sura: verse.sura,
                        ayah: verse.ayah,
                        text: verseTexts[verse.id] ?? "",
                        occurrences: words[verse.id] ?? []
                    )
                    .contentShape(Rectangle())
                    .onTapGesture {
                        Task { await store.navigateTo(sura: verse.sura, ayah: verse.ayah) }
                    }
                }
                .listStyle(.plain)
            }
        }
        .onAppear { loadData() }
        .onChange(of: root) { _, newRoot in loadData() }
    }

    private var totalOccurrences: Int {
        words.values.reduce(0) { $0 + $1.count }
    }

    private func loadData() {
        guard let morph = store.morphology, let reader = store.reader else { return }
        isLoading = true
        verses = []
        verseTexts = [:]
        words = [:]
        surfaceForms = []
        Task {
            // ── Three parallel calls: verses, verse-texts, surface-forms ──
            async let vsTask = morph.versesContainingRoot(root)
            async let occsTask = morph.wordsForRoot(root)
            async let formsTask = morph.distinctSurfaceForms(forRoot: root, limit: 20)
            let (vs, occs, forms) = await (vsTask, occsTask, formsTask)

            // Group occurrences by verse-key
            var occsByKey: [String: [WordOccurrence]] = [:]
            occsByKey.reserveCapacity(vs.count)
            for o in occs {
                let key = "\(o.sura):\(o.ayah)"
                if occsByKey[key] == nil { occsByKey[key] = [] }
                occsByKey[key]!.append(WordOccurrence(word: o.word, surface: o.surface))
            }

            // Bulk-fetch verse texts in ONE call instead of N calls
            let versePairs = vs.map { (sura: $0.sura, ayah: $0.ayah) }
            let texts = await reader.verses(forPairs: versePairs)

            await MainActor.run {
                self.verses = vs.map { VerseRef(sura: $0.sura, ayah: $0.ayah) }
                self.verseTexts = texts
                self.words = occsByKey
                self.surfaceForms = forms
                self.isLoading = false
            }
        }
    }
}

private struct ConcordanceRow: View {
    let sura: Int
    let ayah: Int
    let text: String
    let occurrences: [RootConcordanceView.WordOccurrence]

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack {
                Text("\(sura):\(ayah)")
                    .font(.caption.monospaced())
                    .foregroundStyle(.secondary)
                Spacer()
                Text("\(occurrences.count)×")
                    .font(.caption.monospaced())
                    .foregroundStyle(.blue)
            }
            HStack(alignment: .center, spacing: 8) {
                Text(highlightedText)
                    .font(.system(size: 16, design: .serif))
                    .multilineTextAlignment(.trailing)
                    .environment(\.layoutDirection, .rightToLeft)
                Spacer()
            }
        }
        .padding(.vertical, 4)
    }

    private var highlightedText: AttributedString {
        var attr = AttributedString(text)
        for occ in occurrences {
            let surface = occ.surface
            if let range = attr.range(of: surface) {
                attr[range].backgroundColor = .blue.opacity(0.25)
                attr[range].foregroundColor = .primary
            }
        }
        return attr
    }
}

// MARK: - Range slider

struct RangeSlider: View {
    let minValue: Double
    let maxValue: Double
    @Binding var lowValue: Double
    @Binding var highValue: Double

    var body: some View {
        GeometryReader { geo in
            let trackHeight: CGFloat = 4
            let thumbSize: CGFloat = 14
            let availableWidth = max(1, geo.size.width - thumbSize)
            let loFrac = (lowValue - minValue) / Swift.max(0.0001, maxValue - minValue)
            let hiFrac = (highValue - minValue) / Swift.max(0.0001, maxValue - minValue)
            let loX = thumbSize / 2 + CGFloat(loFrac) * availableWidth
            let hiX = thumbSize / 2 + CGFloat(hiFrac) * availableWidth

            ZStack(alignment: .leading) {
                Capsule()
                    .fill(.quaternary)
                    .frame(height: trackHeight)
                    .padding(.horizontal, thumbSize / 2)

                Capsule()
                    .fill(Color.accentColor)
                    .frame(width: Swift.max(0, hiX - loX), height: trackHeight)
                    .offset(x: loX)

                thumb(at: loX, size: thumbSize)
                    .gesture(
                        DragGesture()
                            .onChanged { v in
                                let frac = Swift.max(0, Swift.min(1, (v.location.x - thumbSize / 2) / availableWidth))
                                lowValue = minValue + (maxValue - minValue) * Double(frac)
                                if lowValue > highValue { highValue = lowValue }
                            }
                    )

                thumb(at: hiX, size: thumbSize)
                    .gesture(
                        DragGesture()
                            .onChanged { v in
                                let frac = Swift.max(0, Swift.min(1, (v.location.x - thumbSize / 2) / availableWidth))
                                highValue = minValue + (maxValue - minValue) * Double(frac)
                                if highValue < lowValue { lowValue = highValue }
                            }
                    )
            }
            .frame(height: thumbSize + 4)
        }
        .frame(height: 18)
    }

    private func thumb(at x: CGFloat, size: CGFloat) -> some View {
        Circle()
            .fill(Color.accentColor)
            .frame(width: size, height: size)
            .overlay(Circle().stroke(.background, lineWidth: 2))
            .shadow(radius: 1)
            .offset(x: x - size / 2)
    }
}
