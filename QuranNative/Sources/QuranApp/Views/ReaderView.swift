import SwiftUI
import QuranKit

func loadQuranFont(size: CGFloat = 40) -> Font {
    if let ct = FontManager.shared.quranFont(size: size) {
        return Font(ct as NSFont)
    }
    return .system(size: size)
}

struct ReaderView: View {
    @Environment(QuranStore.self) private var store
    @State private var highlightAyah: Int? = nil
    @State private var letterSpacing: CGFloat = 0
    @State private var tatweelCount: Int = 0
    @State private var fontSize: CGFloat = 42
    @State private var showTypographyPanel: Bool = false
    @State private var showGloss: Bool = false
    @State private var highlightTask: Task<Void, Never>?
    @State private var maddLaazimCount: Int = 5
    @State private var maddMuttasilCount: Int = 3
    @State private var maddMunfasilSpacing: CGFloat = 10

    var body: some View {
        VStack(spacing: 0) {
            ReaderToolbar(showTypography: $showTypographyPanel, showGloss: $showGloss)
            Divider()
            if !FontManager.shared.loadErrors.isEmpty {
                FontStatusBar()
                Divider()
            }
            SurahHeader()
            if showTypographyPanel {
                TypographyControlBar(
                    letterSpacing: $letterSpacing,
                    tatweelCount: $tatweelCount,
                    fontSize: $fontSize,
                    maddLaazimCount: $maddLaazimCount,
                    maddMuttasilCount: $maddMuttasilCount,
                    maddMunfasilSpacing: $maddMunfasilSpacing
                )
                Divider()
            }
            verseScroll
        }
        .onAppear { loadPrefs() }
        .onChange(of: fontSize) { savePrefs() }
        .onChange(of: letterSpacing) { savePrefs() }
        .onChange(of: tatweelCount) { savePrefs() }
        .onChange(of: maddLaazimCount) { savePrefs() }
        .onChange(of: maddMuttasilCount) { savePrefs() }
        .onChange(of: maddMunfasilSpacing) { savePrefs() }
        .onChange(of: store.selectedSura) { savePrefs() }
    }

    private func loadPrefs() {
        let d = UserDefaults.standard
        if d.object(forKey: "fontSize") != nil { fontSize = d.double(forKey: "fontSize") }
        if d.object(forKey: "letterSpacing") != nil { letterSpacing = d.double(forKey: "letterSpacing") }
        if d.object(forKey: "tatweelCount") != nil { tatweelCount = d.integer(forKey: "tatweelCount") }
        if d.object(forKey: "maddLaazimCount") != nil { maddLaazimCount = d.integer(forKey: "maddLaazimCount") }
        if d.object(forKey: "maddMuttasilCount") != nil { maddMuttasilCount = d.integer(forKey: "maddMuttasilCount") }
        if d.object(forKey: "maddMunfasilSpacing") != nil { maddMunfasilSpacing = d.double(forKey: "maddMunfasilSpacing") }
    }

    private func savePrefs() {
        let d = UserDefaults.standard
        d.set(fontSize, forKey: "fontSize")
        d.set(letterSpacing, forKey: "letterSpacing")
        d.set(tatweelCount, forKey: "tatweelCount")
        d.set(maddLaazimCount, forKey: "maddLaazimCount")
        d.set(maddMuttasilCount, forKey: "maddMuttasilCount")
        d.set(maddMunfasilSpacing, forKey: "maddMunfasilSpacing")
    }

    private var verseScroll: some View {
        ScrollViewReader { proxy in
            ScrollView {
                LazyVStack(alignment: .trailing, spacing: 12) {
                    Color.clear.frame(height: 0).id("top")

                    if !store.bismillahText.isEmpty {
                        VerseBlock(
                            entry: QuranStore.VerseEntry(id: 0, number: 0, text: store.bismillahText, translation: "", wordAnalyses: []),
                            isSelected: false, isHighlighted: false,
                            selectedWordIndex: nil, onSelectWord: { _ in },
                            letterSpacing: letterSpacing, tatweelCount: tatweelCount,
                            fontSize: fontSize, showGloss: showGloss,
                            maddLaazimCount: maddLaazimCount, maddMuttasilCount: maddMuttasilCount,
                            maddMunfasilSpacing: maddMunfasilSpacing
                        ).id("bismillah")
                    }

                    ForEach(store.verseEntries) { entry in
                        VerseBlock(
                            entry: entry,
                            isSelected: store.selectedVerse == entry.number,
                            isHighlighted: highlightAyah == entry.number,
                            selectedWordIndex: store.selectedWordIndex,
                            onSelectWord: { wordIdx in store.selectWord(wordIdx, in: entry.number) },
                            letterSpacing: letterSpacing, tatweelCount: tatweelCount,
                            fontSize: fontSize, showGloss: showGloss,
                            maddLaazimCount: maddLaazimCount, maddMuttasilCount: maddMuttasilCount,
                            maddMunfasilSpacing: maddMunfasilSpacing
                        ).id(entry.number)
                    }

                    Color.clear.frame(height: 80)
                }
                .padding(.horizontal, 24)
                .padding(.top, 12)
            }
            .onChange(of: store.selectedSura) { _, _ in
                withAnimation(.easeInOut(duration: 0.3)) { proxy.scrollTo("top", anchor: .top) }
            }
            .onChange(of: store.scrollToAyah) { _, _ in
                guard let v = store.pendingScrollVerse else { return }
                highlightTask?.cancel()
                highlightTask = Task { @MainActor in
                    try? await Task.sleep(for: .milliseconds(50))
                    if Task.isCancelled { return }
                    withAnimation(.easeInOut(duration: 0.5)) { proxy.scrollTo(v, anchor: .center) }
                    highlightAyah = v
                    try? await Task.sleep(for: .milliseconds(1500))
                    if Task.isCancelled { return }
                    withAnimation(.easeOut(duration: 0.4)) { highlightAyah = nil }
                }
            }
        }
    }
}

// MARK: - Toolbar

struct ReaderToolbar: View {
    @Environment(QuranStore.self) private var store
    @Binding var showTypography: Bool
    @Binding var showGloss: Bool

    var body: some View {
        HStack(spacing: 12) {
            Menu {
                ForEach(1...114, id: \.self) { s in
                    Button("\(s). \(surahName(s))") {
                        Task { await store.navigateTo(sura: s) }
                    }
                }
            } label: {
                HStack(spacing: 8) {
                    Image(systemName: "book.closed.fill")
                        .font(.system(size: 13, weight: .medium))
                        .foregroundStyle(Color.accentColor)
                    Text(surahName(store.selectedSura))
                        .font(.callout.weight(.medium))
                        .lineLimit(1)
                    Text("#\(store.selectedSura)")
                        .font(.caption.monospaced().weight(.semibold))
                        .foregroundStyle(.secondary)
                        .padding(.horizontal, 6).padding(.vertical, 2)
                        .background(.quaternary, in: Capsule())
                    Image(systemName: "chevron.up.chevron.down")
                        .font(.system(size: 8, weight: .bold))
                        .foregroundStyle(.tertiary)
                }
                .padding(.horizontal, 10).padding(.vertical, 6)
                .background(.quinary, in: RoundedRectangle(cornerRadius: 8))
            }
            .menuStyle(.borderlessButton).buttonStyle(.plain)
            .help("Jump to surah")
            .keyboardShortcut("j", modifiers: .command)

            Text("\(verseCount(store.selectedSura)) verses")
                .font(.caption).foregroundStyle(.tertiary).padding(.leading, 4)

            Spacer()

            ToolbarButton(systemName: showGloss ? "text.bubble.fill" : "text.bubble", label: "Toggle translation") {
                withAnimation(.easeInOut(duration: 0.15)) { showGloss.toggle() }
            }
            .foregroundColor(showGloss ? Color.accentColor : .secondary)

            ToolbarButton(systemName: showTypography ? "textformat.size.fill" : "textformat.size", label: "Typography") {
                withAnimation(.easeInOut(duration: 0.15)) { showTypography.toggle() }
            }
            .foregroundColor(showTypography ? Color.accentColor : .secondary)

            ToolbarButton(systemName: "chevron.left", label: "Previous surah",
                          disabled: store.selectedSura == 1) {
                Task { await store.navigateTo(sura: store.selectedSura - 1) }
            }
            .keyboardShortcut(.leftArrow, modifiers: .command)

            ToolbarButton(systemName: "chevron.right", label: "Next surah",
                          disabled: store.selectedSura == 114) {
                Task { await store.navigateTo(sura: store.selectedSura + 1) }
            }
            .keyboardShortcut(.rightArrow, modifiers: .command)
        }
        .padding(.horizontal, 16).padding(.vertical, 8)
    }
}

struct ToolbarButton: View {
    let systemName: String; let label: String
    var disabled: Bool = false; let action: () -> Void
    @State private var isHovering = false

    var body: some View {
        Button(action: action) {
            Image(systemName: systemName)
                .font(.system(size: 12, weight: .medium))
                .foregroundStyle(disabled ? .tertiary : (isHovering ? .primary : .secondary))
                .frame(width: 28, height: 28)
                .background(.quaternary.opacity(isHovering && !disabled ? 1 : 0), in: RoundedRectangle(cornerRadius: 6))
                .contentShape(RoundedRectangle(cornerRadius: 6))
        }
        .buttonStyle(.plain).disabled(disabled).onHover { isHovering = $0 }.help(label)
    }
}

// MARK: - Typography

struct TypographyControlBar: View {
    @Binding var letterSpacing: CGFloat
    @Binding var tatweelCount: Int
    @Binding var fontSize: CGFloat
    @Binding var maddLaazimCount: Int
    @Binding var maddMuttasilCount: Int
    @Binding var maddMunfasilSpacing: CGFloat

    var body: some View {
        HStack(spacing: 14) {
            HStack(spacing: 6) {
                Image(systemName: "textformat.size").foregroundStyle(.secondary).font(.caption)
                Text("Font").font(.caption).foregroundStyle(.secondary).frame(width: 30, alignment: .leading)
                Slider(value: $fontSize, in: 20...72, step: 2).frame(width: 100).tint(.accentColor)
                Text(String(format: "%.0f", fontSize))
                    .font(.caption.monospaced()).foregroundStyle(.primary).frame(width: 24, alignment: .trailing)
            }
            Divider().frame(height: 18)
            HStack(spacing: 6) {
                Image(systemName: "arrow.left.and.right").foregroundStyle(.secondary).font(.caption)
                Text("Spacing").font(.caption).foregroundStyle(.secondary)
                Slider(value: $letterSpacing, in: 0...6, step: 0.5).frame(width: 80).tint(.accentColor)
                Text(String(format: "%.1f", letterSpacing))
                    .font(.caption.monospaced()).foregroundStyle(.primary).frame(width: 24, alignment: .trailing)
            }
            Divider().frame(height: 18)
            HStack(spacing: 6) {
                Image(systemName: "arrow.left.and.right.text.vertical").foregroundStyle(.secondary).font(.caption)
                Text("Tatweel").font(.caption).foregroundStyle(.secondary)
                Stepper("", value: $tatweelCount, in: 0...5).labelsHidden()
                Text(tatweelCount == 0 ? "off" : "\(tatweelCount)×")
                    .font(.caption.monospaced()).foregroundStyle(.primary).frame(width: 26, alignment: .leading)
            }
            Divider().frame(height: 18)
            HStack(spacing: 6) {
                Image(systemName: "text.alignleft").foregroundStyle(.secondary).font(.caption)
                Text("Lzm").font(.caption).foregroundStyle(.secondary)
                Stepper("", value: $maddLaazimCount, in: 0...10).labelsHidden()
                Text(maddLaazimCount == 0 ? "off" : "\(maddLaazimCount)×")
                    .font(.caption.monospaced()).foregroundStyle(.primary).frame(width: 26, alignment: .leading)
            }
            HStack(spacing: 6) {
                Image(systemName: "text.aligncenter").foregroundStyle(.secondary).font(.caption)
                Text("Mts").font(.caption).foregroundStyle(.secondary)
                Stepper("", value: $maddMuttasilCount, in: 0...10).labelsHidden()
                Text(maddMuttasilCount == 0 ? "off" : "\(maddMuttasilCount)×")
                    .font(.caption.monospaced()).foregroundStyle(.primary).frame(width: 26, alignment: .leading)
            }
            HStack(spacing: 6) {
                Image(systemName: "text.justify").foregroundStyle(.secondary).font(.caption)
                Text("Mfs").font(.caption).foregroundStyle(.secondary)
                Slider(value: $maddMunfasilSpacing, in: 0...30, step: 1).frame(width: 60).tint(.accentColor)
                Text(String(format: "%.0f", maddMunfasilSpacing))
                    .font(.caption.monospaced()).foregroundStyle(.primary).frame(width: 22, alignment: .trailing)
            }
            Divider().frame(height: 18)
            Button {
                withAnimation {
                    letterSpacing = 0; tatweelCount = 0; fontSize = 42
                    maddLaazimCount = 5; maddMuttasilCount = 3; maddMunfasilSpacing = 10
                }
            } label: { Image(systemName: "arrow.counterclockwise").font(.caption) }
                .buttonStyle(.borderless).help("Reset to defaults")
                .disabled(letterSpacing == 0 && tatweelCount == 0 && fontSize == 42
                          && maddLaazimCount == 5 && maddMuttasilCount == 3 && maddMunfasilSpacing == 10)
            Spacer()
        }
        .padding(.horizontal, 16).padding(.vertical, 8).background(Color.accentColor.opacity(0.04))
    }
}

// MARK: - Font status

struct FontStatusBar: View {
    var body: some View {
        if !FontManager.shared.loadErrors.isEmpty {
            HStack(spacing: 6) {
                Image(systemName: "exclamationmark.triangle.fill").foregroundStyle(.orange).font(.system(size: 10))
                Text(FontManager.shared.loadErrors.joined(separator: "; "))
                    .font(.caption2).foregroundStyle(.secondary)
            }.padding(.horizontal, 16).padding(.vertical, 3)
        }
    }
}

// MARK: - Surah header

struct SurahHeader: View {
    @Environment(QuranStore.self) private var store

    var body: some View {
        HStack(spacing: 12) {
            ZStack {
                RoundedRectangle(cornerRadius: 8)
                    .fill(LinearGradient(colors: [Color.accentColor.opacity(0.85), Color.accentColor.opacity(0.55)], startPoint: .topLeading, endPoint: .bottomTrailing))
                    .frame(width: 34, height: 34)
                Text("\(store.selectedSura)")
                    .font(.system(size: 14, weight: .bold, design: .serif)).foregroundStyle(.white)
            }
            VStack(alignment: .leading, spacing: 1) {
                Text(surahName(store.selectedSura)).font(.body.weight(.semibold))
                Text("Surah \(store.selectedSura) of 114 · \(verseCount(store.selectedSura)) verses")
                    .font(.caption).foregroundStyle(.secondary)
            }
            Spacer()
            if let selected = store.selectedVerse {
                HStack(spacing: 4) {
                    Text("Ayah \(selected)")
                        .font(.caption.monospaced().weight(.medium)).foregroundStyle(.secondary)
                        .padding(.horizontal, 8).padding(.vertical, 3)
                        .background(.quaternary, in: Capsule())
                    if store.isBookmarked(sura: store.selectedSura, ayah: selected) {
                        Image(systemName: "bookmark.fill").font(.caption2).foregroundStyle(.blue)
                    }
                }
            }
        }
        .padding(.horizontal, 16).padding(.vertical, 8)
    }
}

// MARK: - Verse block

struct VerseBlock: View {
    let entry: QuranStore.VerseEntry
    let isSelected: Bool
    let isHighlighted: Bool
    let selectedWordIndex: Int?
    let onSelectWord: (Int) -> Void
    var letterSpacing: CGFloat = 0
    var tatweelCount: Int = 0
    var fontSize: CGFloat = 42
    var showGloss: Bool = false
    var maddLaazimCount: Int = 5
    var maddMuttasilCount: Int = 3
    var maddMunfasilSpacing: CGFloat = 10
    @Environment(QuranStore.self) private var store
    @State private var showActions = false
    @State private var copied = false

    var body: some View {
        VStack(alignment: .trailing, spacing: entry.number > 0 ? 8 : 0) {
            HStack(alignment: .top, spacing: entry.number > 0 ? 14 : 0) {
                if entry.number > 0 {
                    AyahNumberBadge(number: entry.number, isSelected: isSelected)
                        .contextMenu { verseContextMenu }
                }
                VStack(alignment: .trailing, spacing: 0) {
                    InteractiveArabicText(
                        text: entry.text, wordAnalyses: entry.wordAnalyses,
                        fontSize: fontSize, letterSpacing: letterSpacing, tatweelCount: tatweelCount,
                        maddLaazimCount: maddLaazimCount, maddMuttasilCount: maddMuttasilCount,
                        maddMunfasilSpacing: maddMunfasilSpacing
                    )
                    .frame(maxWidth: .infinity, alignment: .trailing)
                    .frame(minHeight: 56)
                    .padding(.vertical, 4)
                    .contentShape(Rectangle())
                    .onTapGesture { onSelectWord(-1) }
                    .contextMenu { verseContextMenu }
                }
            }

            if showGloss && entry.number > 0 && !entry.translation.isEmpty {
                Text(entry.translation)
                    .font(.callout)
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .environment(\.layoutDirection, .leftToRight)
                    .padding(.leading, 50)
            }

            if showActions {
                HStack(spacing: 6) {
                    Button { copyVerseText() } label: {
                        Label(copied ? "Copied" : "Copy", systemImage: copied ? "checkmark" : "doc.on.doc")
                            .font(.caption)
                    }
                    .buttonStyle(.borderless)

                    Button { copyVerseRef() } label: {
                        Label("Ref", systemImage: "link").font(.caption)
                    }
                    .buttonStyle(.borderless)

                    Button {
                        store.toggleBookmark(sura: store.selectedSura, ayah: entry.number)
                    } label: {
                        Label(store.isBookmarked(sura: store.selectedSura, ayah: entry.number) ? "Saved" : "Bookmark",
                              systemImage: store.isBookmarked(sura: store.selectedSura, ayah: entry.number) ? "bookmark.fill" : "bookmark")
                            .font(.caption)
                    }
                    .buttonStyle(.borderless)
                }
                .padding(.top, 4)
            }
        }
        .padding(.horizontal, 14).padding(.vertical, 10)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(isSelected ? Color.accentColor.opacity(0.06) : Color(nsColor: .controlBackgroundColor).opacity(0.5))
                .overlay(
                    RoundedRectangle(cornerRadius: 10)
                        .strokeBorder(isHighlighted ? Color.accentColor : (isSelected ? Color.accentColor.opacity(0.3) : Color.primary.opacity(0.06)), lineWidth: isHighlighted ? 2 : 0.5)
                )
        )
        .onHover { hovering in
            if entry.number > 0 { withAnimation(.easeInOut(duration: 0.15)) { showActions = hovering } }
        }
        .animation(.easeInOut(duration: 0.25), value: isHighlighted)
        .animation(.easeInOut(duration: 0.15), value: isSelected)
    }

    private var verseContextMenu: some View {
        Group {
            Button("Copy Verse") { copyVerseText() }
            Button("Copy Reference (\(store.selectedSura):\(entry.number))") { copyVerseRef() }
            Divider()
            Button(store.isBookmarked(sura: store.selectedSura, ayah: entry.number) ? "Remove Bookmark" : "Bookmark This Verse") {
                store.toggleBookmark(sura: store.selectedSura, ayah: entry.number)
            }
        }
    }

    private func copyVerseText() {
        NSPasteboard.general.clearContents()
        NSPasteboard.general.setString(entry.text, forType: .string)
        copied = true
        Task {
            try? await Task.sleep(for: .seconds(2))
            copied = false
        }
    }

    private func copyVerseRef() {
        NSPasteboard.general.clearContents()
        NSPasteboard.general.setString("[\(store.selectedSura):\(entry.number)] \(entry.text)", forType: .string)
    }
}

// MARK: - Ayah number badge

struct AyahNumberBadge: View {
    let number: Int
    let isSelected: Bool

    var body: some View {
        ZStack {
            Circle()
                .strokeBorder(LinearGradient(colors: [Color.accentColor, Color.accentColor.opacity(0.5)], startPoint: .topLeading, endPoint: .bottomTrailing), lineWidth: 1.5)
                .frame(width: 36, height: 36)
            Text(arabicNumeral(number))
                .font(.system(size: 18, weight: .semibold, design: .serif))
                .foregroundStyle(isSelected ? Color.accentColor : .primary)
        }
        .padding(.top, 6)
    }

    private func arabicNumeral(_ n: Int) -> String {
        let arabicDigits = ["٠","١","٢","٣","٤","٥","٦","٧","٨","٩"]
        var result = ""
        for c in String(n) {
            if let d = c.wholeNumberValue, d < arabicDigits.count {
                result.append(arabicDigits[d])
            } else { result.append(c) }
        }
        return result
    }
}
