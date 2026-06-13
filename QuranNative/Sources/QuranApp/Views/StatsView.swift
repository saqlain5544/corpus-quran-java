import SwiftUI
import QuranKit

struct StatsView: View {
    @Environment(QuranStore.self) private var store
    @State private var segmentCount: Int = 0
    @State private var rootCount: Int = 0
    @State private var partStats: (n: Int, v: Int, p: Int, pr: Int, m: Int) = (0,0,0,0,0)
    @State private var topRoots: [(name: String, count: Int)] = []

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                HStack {
                    VStack(alignment: .leading, spacing: 2) {
                        Text("Quran Statistics").font(.largeTitle.weight(.bold))
                        Text("Live corpus metrics").font(.callout).foregroundStyle(.secondary)
                    }
                    Spacer()
                }

                LazyVGrid(columns: [GridItem(.adaptive(minimum: 360), spacing: 20)], spacing: 20) {
                    StatsCard(title: "Overview", systemImage: "doc.text") {
                        StatRow(icon: "book.closed", label: "Surahs", value: "114", color: .blue)
                        StatRow(icon: "list.number", label: "Verses", value: "6,236", color: .indigo)
                        StatRow(icon: "textformat.abc", label: "Words", value: "77,430", color: .purple)
                        StatRow(icon: "square.grid.3x3", label: "Morph Segments", value: formatNum(segmentCount), color: .pink)
                    }

                    StatsCard(title: "Revelation", systemImage: "moon.stars") {
                        StatRow(icon: "sun.max", label: "Makki Surahs", value: "\(makkiSuras.count)", color: .orange)
                        StatRow(icon: "building.columns", label: "Madani Surahs", value: "\(madaniSuras.count)", color: .green)
                        StatRow(icon: "sparkles", label: "First revelation", value: "Al-ʿAlaq (96:1–5)", color: .yellow)
                        StatRow(icon: "moon", label: "Last revelation", value: "Al-Māʾidah (5:3)", color: .teal)
                    }

                    StatsCard(title: "Divisions", systemImage: "rectangle.split.3x1") {
                        StatRow(icon: "1.square", label: "Juzʾ (Parts)", value: "30", color: .red)
                        StatRow(icon: "2.square", label: "Hizb (Groups)", value: "60", color: .orange)
                        StatRow(icon: "4.square", label: "Rubʿ (Quarters)", value: "240", color: .yellow)
                        StatRow(icon: "figure.flexibility", label: "Sajda Verses", value: "\(sajdaCount)", color: .green)
                        StatRow(icon: "scroll", label: "Waqf Marks", value: formatNum(partStats.m), color: .teal)
                    }

                    StatsCard(title: "Triliteral Roots", systemImage: "leaf") {
                        StatRow(icon: "circle.grid.cross", label: "Unique Roots", value: formatNum(rootCount), color: .blue)
                        StatRow(icon: "chart.bar", label: "Most Frequent", value: topRoots.first?.name ?? "—", color: .orange)
                        StatRow(icon: "chart.line.downtrend.xyaxis", label: "Occurrences", value: formatNum(topRoots.first?.count), color: .purple)
                        StatRow(icon: "character.book.closed", label: "Divine Names", value: "\(DIVINE_NAMES.count)", color: .pink)
                        StatRow(icon: "text.word.spacing", label: "Distinct Forms", value: formatNum(topRoots.count), color: .indigo)
                    }

                    if segmentCount > 0 {
                        StatsCard(title: "Word Classes", systemImage: "tag") {
                            StatRow(icon: "character", label: "Nouns", value: formatNum(partStats.n), color: .blue)
                            StatRow(icon: "arrow.triangle.branch", label: "Verbs", value: formatNum(partStats.v), color: .green)
                            StatRow(icon: "circle.hexagongrid", label: "Particles", value: formatNum(partStats.p), color: .orange)
                            StatRow(icon: "person", label: "Pronouns", value: formatNum(partStats.pr), color: .purple)
                        }
                    }
                }

                Color.clear.frame(height: 40)
            }
            .padding(.horizontal, 30).padding(.top, 24)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .background(Color(nsColor: .windowBackgroundColor))
        .task { await loadStats() }
    }

    private func formatNum(_ n: Int?) -> String {
        guard let n else { return "—" }
        if n >= 1000 {
            let k = Double(n) / 1000.0
            return String(format: "%.1fk", k)
        }
        return "\(n)"
    }

    private func loadStats() async {
        guard let m = store.morphology else { return }
        segmentCount = await m.segmentCount
        rootCount = await m.rootCount
        let stats = await m.allRootStats()
        topRoots = stats.map { ($0.root, $0.totalOccurrences) }.sorted { $0.1 > $1.1 }
        partStats = await computeParts(m)
    }

    private func computeParts(_ m: MorphologyData) async -> (n: Int, v: Int, p: Int, pr: Int, m: Int) {
        let segs = await m.segments
        var noun = 0, verb = 0, particle = 0, pronoun = 0, marks = 0
        for s in segs {
            switch categorize(s.pos) {
            case .noun: noun += 1
            case .verb: verb += 1
            case .particle: particle += 1
            case .pronoun: pronoun += 1
            case .other: break
            }
            if !s.punctuation.isEmpty { marks += 1 }
        }
        return (noun, verb, particle, pronoun, marks)
    }
}

// MARK: - Stats card

struct StatsCard<Content: View>: View {
    let title: String; let systemImage: String
    @ViewBuilder let content: Content

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            HStack(spacing: 10) {
                Image(systemName: systemImage)
                    .font(.system(size: 18, weight: .medium)).foregroundStyle(Color.accentColor)
                    .frame(width: 32, height: 32)
                    .background(Color.accentColor.opacity(0.12), in: RoundedRectangle(cornerRadius: 8))
                Text(title).font(.title3.weight(.semibold))
                Spacer()
            }
            .padding(.horizontal, 18).padding(.top, 18).padding(.bottom, 12)
            VStack(spacing: 0) { content }.padding(.vertical, 4)
        }
        .background(
            RoundedRectangle(cornerRadius: 14)
                .fill(Color(nsColor: .controlBackgroundColor))
                .shadow(color: .black.opacity(0.06), radius: 8, y: 2)
        )
        .overlay(RoundedRectangle(cornerRadius: 14).strokeBorder(Color.primary.opacity(0.05), lineWidth: 1))
    }
}

struct StatRow: View {
    let icon: String; let label: String; let value: String; let color: Color

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: icon).font(.system(size: 14, weight: .medium)).foregroundStyle(color)
                .frame(width: 28, height: 28).background(color.opacity(0.12), in: RoundedRectangle(cornerRadius: 6))
            Text(label).font(.callout)
            Spacer()
            Text(value).font(.callout.monospaced().weight(.medium)).foregroundStyle(.primary)
        }
        .padding(.horizontal, 18).padding(.vertical, 10)
    }
}
