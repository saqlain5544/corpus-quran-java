import SwiftUI
import QuranKit

enum AppSection: String, CaseIterable {
    case reader = "Reader"
    case search = "Search"
    case roots = "Roots"
    case divineNames = "99 Names"
    case stats = "Statistics"
}

struct SidebarView: View {
    @Environment(QuranStore.self) private var store
    @Binding var selectedSection: AppSection

    var body: some View {
        VStack(spacing: 0) {
            VStack(spacing: 4) {
                ZStack {
                    RoundedRectangle(cornerRadius: 10)
                        .fill(LinearGradient(colors: [Color.accentColor.opacity(0.9), Color.accentColor.opacity(0.5)], startPoint: .topLeading, endPoint: .bottomTrailing))
                        .frame(width: 38, height: 38)
                    Image(systemName: "book.fill")
                        .font(.system(size: 18, weight: .medium)).foregroundStyle(.white)
                }
                Text("Quran Corpus").font(.body.weight(.semibold))
                Text("Study Tool").font(.caption).foregroundStyle(.secondary)
            }
            .frame(maxWidth: .infinity).padding(.vertical, 14)

            Divider()

            List(selection: $selectedSection) {
                Section("Browse") {
                    SidebarRow(label: "Reader", icon: "book.fill", tag: .reader, shortcut: "1")
                    SidebarRow(label: "Search", icon: "magnifyingglass", tag: .search, shortcut: "2")
                    SidebarRow(label: "Roots", icon: "leaf.fill", tag: .roots, shortcut: "3")
                }
                Section("Reference") {
                    SidebarRow(label: "99 Names", icon: "sparkles", tag: .divineNames, shortcut: "4")
                    SidebarRow(label: "Statistics", icon: "chart.bar.fill", tag: .stats, shortcut: "5")
                }

                if !store.bookmarkedVerses.isEmpty {
                    Section("Bookmarks") {
                        ForEach(Array(store.bookmarkedVerses).sorted(), id: \.self) { key in
                            let parts = key.split(separator: ":").compactMap { Int($0) }
                            if parts.count == 2 {
                                Button {
                                    Task { await store.navigateTo(sura: parts[0], ayah: parts[1]) }
                                } label: {
                                    Label("[\(parts[0]):\(parts[1])]", systemImage: "bookmark.fill")
                                }
                                .buttonStyle(.plain)
                            }
                        }
                    }
                }
            }
            .listStyle(.sidebar)

            Divider()

            VStack(spacing: 0) {
                HStack {
                    Text("Surahs").font(.caption.weight(.semibold)).foregroundStyle(.secondary)
                    Spacer()
                    Text("\(store.selectedSura) of 114")
                        .font(.caption2.monospaced()).foregroundStyle(.tertiary)
                }
                .padding(.horizontal, 16).padding(.top, 8).padding(.bottom, 4)

                SurahFilterField()
                SuraListView()
            }
        }
    }
}

private struct SidebarRow: View {
    let label: String; let icon: String; let tag: AppSection; let shortcut: String

    var body: some View {
        Label { Text(label) } icon: {
            Image(systemName: icon).symbolRenderingMode(.hierarchical)
        }
        .tag(tag)
        .keyboardShortcut(KeyEquivalent(Character(shortcut)), modifiers: .command)
    }
}

// MARK: - Surah search filter

struct SurahFilterField: View {
    @State private var filterText: String = ""

    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: "magnifyingglass").foregroundStyle(.tertiary).font(.system(size: 10))
            TextField("Filter surahs...", text: $filterText)
                .textFieldStyle(.plain).font(.caption)
            if !filterText.isEmpty {
                Button { filterText = "" } label: {
                    Image(systemName: "xmark.circle.fill").font(.system(size: 10))
                }
                .buttonStyle(.borderless).foregroundStyle(.secondary)
            }
        }
        .padding(.horizontal, 12).padding(.vertical, 4)
        .background(.quinary, in: RoundedRectangle(cornerRadius: 6))
        .padding(.horizontal, 10).padding(.bottom, 4)
        .environment(\.surahFilter, filterText)
    }
}

private struct SurahFilterKey: EnvironmentKey {
    static let defaultValue: String = ""
}

extension EnvironmentValues {
    var surahFilter: String {
        get { self[SurahFilterKey.self] }
        set { self[SurahFilterKey.self] = newValue }
    }
}

// MARK: - Sura list

struct SuraListView: View {
    @Environment(QuranStore.self) private var store
    @Environment(\.surahFilter) private var filterText
    @State private var showJuzView = false

    var body: some View {
        VStack(spacing: 0) {
            Picker("View", selection: $showJuzView) {
                Text("Surahs").tag(false)
                Text("Juz'").tag(true)
            }
            .pickerStyle(.segmented)
            .padding(.horizontal, 10).padding(.bottom, 4)
            .controlSize(.small)

            if showJuzView {
                juzList
            } else {
                surahList
            }
        }
    }

    private var surahList: some View {
        List(selection: Binding(get: { store.selectedSura }, set: { new in
            if let sura = new { Task { await store.navigateTo(sura: sura) } }
        })) {
            ForEach(filteredGroups, id: \.name) { group in
                Section {
                    ForEach(group.surahs, id: \.self) { sura in
                        SuraRow(sura: sura).tag(sura)
                    }
                } header: { Text(group.name).font(.caption.weight(.semibold)) }
            }
        }
        .listStyle(.sidebar)
    }

    private var juzList: some View {
        List {
            ForEach(ALL_JUZ.prefix(30)) { juz in
                Button {
                    Task { await store.navigateTo(sura: juz.startSura, ayah: juz.startVerse) }
                } label: {
                    HStack {
                        Text("Juz' \(juz.number)").font(.body)
                        Spacer()
                        Text("\(juz.startSura):\(juz.startVerse)")
                            .font(.caption.monospaced()).foregroundStyle(.secondary)
                    }
                }
                .buttonStyle(.plain)
            }
        }
        .listStyle(.sidebar)
    }

    private var filteredGroups: [(name: String, surahs: [Int])] {
        let medinan = [2,3,4,5,8,9,13,22,24,33,47,48,49,55,57,58,59,60,61,62,63,64,65,66,76,110]
        let earlyMakki = Array(1...30).filter { !medinan.contains($0) }
        let midMakki = Array(31...60).filter { !medinan.contains($0) }
        let lateMakki = Array(61...114).filter { !medinan.contains($0) }
        let all: [(String, [Int])] = [("Medinan", medinan), ("Early Makki", earlyMakki), ("Mid Makki", midMakki), ("Late Makki", lateMakki)]

        let q = filterText.trimmingCharacters(in: .whitespaces).lowercased()
        guard !q.isEmpty else { return all }

        return all.compactMap { group in
            let filtered = group.1.filter { sura in
                surahName(sura).lowercased().contains(q) || String(sura).contains(q)
            }
            return filtered.isEmpty ? nil : (group.0, filtered)
        }
    }
}

private struct SuraRow: View {
    @Environment(QuranStore.self) private var store
    let sura: Int

    var body: some View {
        HStack(spacing: 6) {
            Text("\(sura)").font(.caption.monospaced()).foregroundStyle(.secondary).frame(width: 22, alignment: .trailing)
            Text(surahName(sura)).font(.body).lineLimit(1)
            if SAJDA_SURAS.contains(sura) {
                Image(systemName: "text.page.badge.magnifyingglass").foregroundStyle(.orange).font(.caption)
            }
            Spacer(minLength: 0)
        }
        .contentShape(Rectangle())
        .contextMenu {
            Button("Jump to verse 1") { Task { await store.navigateTo(sura: sura, ayah: 1) } }
            Button("Open in Reader") { Task { await store.navigateTo(sura: sura) } }
            Divider()
            Text(surahName(sura))
        }
    }
}
