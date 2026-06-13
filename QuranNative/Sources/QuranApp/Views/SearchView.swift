import SwiftUI
import QuranKit

struct SearchView: View {
    @Environment(QuranStore.self) private var store
    @State private var queryText: String = ""
    @State private var searchMode: QuranStore.SearchMode = .word
    @State private var debounceTask: Task<Void, Never>? = nil
    @FocusState private var queryFocused: Bool

    var body: some View {
        VStack(spacing: 0) {
            searchBar
                .padding(.horizontal, 20)
                .padding(.top, 16)
                .padding(.bottom, 12)
                .background(.bar)

            Divider().opacity(0.5)

            resultsContent
        }
        .onAppear {
            queryFocused = true
        }
        .onKeyPress(.escape) {
            queryFocused = false
            return .handled
        }
        .onChange(of: queryText) { _, new in
            debounceAndSearch(query: new)
        }
        .onChange(of: searchMode) { _, _ in
            if !queryText.isEmpty {
                debounceAndSearch(query: queryText, debounceMs: 0)
            }
        }
    }

    // MARK: - Search bar

    private var searchBar: some View {
        VStack(spacing: 10) {
            HStack(spacing: 10) {
                // Search field — use a real macOS text field style so it
                // participates in the responder chain properly.
                HStack(spacing: 8) {
                    Image(systemName: "magnifyingglass")
                        .foregroundStyle(.secondary)
                        .font(.system(size: 14, weight: .medium))
                    TextField("Search the Quran…", text: $queryText)
                        .textFieldStyle(.roundedBorder)
                        .focused($queryFocused)
                        .font(.callout)
                        .disableAutocorrection(true)
                        .onSubmit { performSearchNow() }
                    if !queryText.isEmpty {
                        Button {
                            queryText = ""
                        } label: {
                            Image(systemName: "xmark.circle.fill")
                                .foregroundStyle(.secondary)
                                .font(.system(size: 13))
                        }
                        .buttonStyle(.borderless)
                        .help("Clear")
                    }
                }

                // Mode picker
                Picker("Mode", selection: $searchMode) {
                    ForEach(QuranStore.SearchMode.allCases, id: \.self) { mode in
                        Text(mode.rawValue).tag(mode)
                    }
                }
                .pickerStyle(.segmented)
                .frame(width: 220)
                .controlSize(.regular)
                .labelsHidden()

                // Search button
                Button {
                    performSearchNow()
                } label: {
                    Label("Search", systemImage: "magnifyingglass")
                }
                .buttonStyle(.borderedProminent)
                .controlSize(.regular)
                .keyboardShortcut(.return, modifiers: .command)
                .disabled(queryText.isEmpty)
            }
        }
    }

    // MARK: - Results

    @ViewBuilder
    private var resultsContent: some View {
        ZStack {
            if store.isSearching {
                loadingState
            } else if queryText.isEmpty {
                emptyHintState
            } else if store.searchResults.isEmpty {
                noResultsState
            } else {
                resultsList
            }
        }
    }

    private var loadingState: some View {
        VStack(spacing: 14) {
            ProgressView()
                .controlSize(.large)
            Text("Searching…")
                .font(.callout)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private var emptyHintState: some View {
        VStack(spacing: 16) {
            Image(systemName: "text.magnifyingglass")
                .font(.system(size: 56, weight: .ultraLight))
                .foregroundStyle(.tertiary)
            VStack(spacing: 6) {
                Text("Search the Quran")
                    .font(.title3.weight(.medium))
                Text("Type an Arabic word, an English gloss, or a root like kwn")
                    .font(.callout)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
            }
        }
        .padding(40)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private var noResultsState: some View {
        ContentUnavailableView {
            Label("No matches", systemImage: "questionmark.text.page")
        } description: {
            Text("Nothing found for \"\(queryText)\" in \(searchMode.rawValue.lowercased()) mode.")
        } actions: {
            Button("Try a different mode") {
                cycleMode()
            }
            .buttonStyle(.bordered)
        }
    }

    private var resultsList: some View {
        // Header bar with result count + timing
        VStack(spacing: 0) {
            HStack {
                Text("\(store.searchResults.count) result\(store.searchResults.count == 1 ? "" : "s")")
                    .font(.caption.weight(.medium))
                if store.searchDurationMs > 0 {
                    Text("in \(store.searchDurationMs) ms")
                        .font(.caption.monospaced()).foregroundStyle(.secondary)
                } else {
                    Text("(cached)").font(.caption).foregroundStyle(.tertiary)
                }
                Spacer()
                Button {
                    let text = store.searchResults.prefix(500).map { "[\($0.sura):\($0.verse):\($0.word)] \($0.text)" }.joined(separator: "\n")
                    NSPasteboard.general.clearContents()
                    NSPasteboard.general.setString(text, forType: .string)
                } label: {
                    Label("Copy results", systemImage: "doc.on.doc").font(.caption2)
                }
                .buttonStyle(.borderless).help("Copy first 500 results to clipboard")
                Text("Click a row to open the ayah").font(.caption2).foregroundStyle(.tertiary)
            }
            .padding(.horizontal, 20)
            .padding(.vertical, 8)
            .background(.bar)

            Divider().opacity(0.5)

            ScrollView {
                let first500 = Array(store.searchResults.prefix(500))
                let lastId = first500.last?.id
                LazyVStack(alignment: .leading, spacing: 0) {
                    ForEach(first500) { line in
                        SearchResultRow(
                            line: line,
                            query: queryText
                        ) {
                            Task {
                                await store.navigateTo(sura: line.sura, ayah: line.verse)
                            }
                        }
                        if line.id != lastId {
                            Divider().opacity(0.3)
                        }
                    }
                    if store.searchResults.count > 500 {
                        HStack {
                            Spacer()
                            Text("Showing first 500 of \(store.searchResults.count)")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                                .padding(.vertical, 8)
                            Spacer()
                        }
                    }
                }
            }
        }
    }

    // MARK: - Actions

    private func debounceAndSearch(query: String, debounceMs: Int = 250) {
        debounceTask?.cancel()
        let q = query
        let mode = searchMode
        debounceTask = Task {
            if debounceMs > 0 {
                try? await Task.sleep(for: .milliseconds(debounceMs))
            }
            if Task.isCancelled { return }
            await store.search(q, mode: mode)
        }
    }

    private func performSearchNow() {
        debounceTask?.cancel()
        Task { await store.search(queryText, mode: searchMode) }
    }

    private func cycleMode() {
        let all = QuranStore.SearchMode.allCases
        let idx = all.firstIndex(of: searchMode) ?? 0
        let next = all[(idx + 1) % all.count]
        searchMode = next
    }
}

// MARK: - Search result row

private struct SearchResultRow: View {
    let line: ConcordanceLine
    let query: String
    let onTap: () -> Void
    @State private var isHovering = false

    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            // Reference pill
            VStack(spacing: 2) {
                Text("\(line.sura):\(line.verse)")
                    .font(.system(size: 12, weight: .semibold, design: .monospaced))
                    .foregroundStyle(.primary)
                Text("w\(line.word)")
                    .font(.system(size: 10, design: .monospaced))
                    .foregroundStyle(.tertiary)
            }
            .frame(width: 64)
            .padding(.vertical, 6)
            .background(
                RoundedRectangle(cornerRadius: 6)
                    .fill(Color.accentColor.opacity(isHovering ? 0.15 : 0.06))
            )

            // Verse context with highlighted target
            VStack(alignment: .leading, spacing: 4) {
                Text(highlightedContext)
                    .font(.system(size: 14))
                    .multilineTextAlignment(.leading)
                    .lineLimit(3)

                HStack(spacing: 12) {
                    Label(line.targetWord, systemImage: "textformat")
                        .font(.caption.monospaced())
                        .foregroundStyle(.blue)
                    if let ref = line.text.first.map({ _ in "\(line.sura):\(line.verse):\(line.word)" }) {
                        Text(ref)
                            .font(.caption2.monospaced())
                            .foregroundStyle(.tertiary)
                    }
                }
            }
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 10)
        .background(
            isHovering ? Color.accentColor.opacity(0.04) : Color.clear,
            in: Rectangle()
        )
        .contentShape(Rectangle())
        .onTapGesture { onTap() }
        .onHover { isHovering = $0 }
        .help("Click to open \(line.sura):\(line.verse)")
    }

    private var highlightedContext: AttributedString {
        let raw = "\(line.leftContext)  ⟨\(line.targetWord)⟩  \(line.rightContext)"
        var attr = AttributedString(raw)
        // Highlight the target word with accent background
        if let range = attr.range(of: line.targetWord) {
            attr[range].backgroundColor = .blue.opacity(0.18)
            attr[range].foregroundColor = .primary
        }
        // Highlight the ⟨ ⟩ brackets
        if let openRange = attr.range(of: "⟨") {
            attr[openRange].foregroundColor = .secondary
        }
        if let closeRange = attr.range(of: "⟩") {
            attr[closeRange].foregroundColor = .secondary
        }
        return attr
    }
}
