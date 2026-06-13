import SwiftUI
import QuranKit

struct InteractiveArabicText: View {
    let text: String
    let wordAnalyses: [WordAnalysis]
    let fontSize: CGFloat

    var letterSpacing: CGFloat = 0
    var tatweelCount: Int = 0

    struct WordEntry {
        let text: String
        let index: Int
        let madd: MaddMatch?
        var needsMunfasilSpacing: Bool { madd?.type == .munfasil }
    }

    @State private var cachedWords: [WordEntry] = []
    @State private var analysisByWord: [Int: WordAnalysis] = [:]

    init(text: String, wordAnalyses: [WordAnalysis], fontSize: CGFloat = 40,
         letterSpacing: CGFloat = 0, tatweelCount: Int = 0) {
        self.text = text
        self.wordAnalyses = wordAnalyses
        self.fontSize = fontSize
        self.letterSpacing = letterSpacing
        self.tatweelCount = tatweelCount
    }

    var body: some View {
        FlowLayoutRTL(spacing: 0) {
            ForEach(Array(cachedWords.enumerated()), id: \.offset) { _, word in
                wordView(entry: word)
            }
        }
        .frame(maxWidth: .infinity, alignment: .trailing)
        .onAppear { buildCache() }
        .onChange(of: text) { _, _ in buildCache() }
    }

    private func buildCache() {
        var dict: [Int: WordAnalysis] = [:]
        dict.reserveCapacity(wordAnalyses.count)
        for a in wordAnalyses { dict[a.wordIndex] = a }
        analysisByWord = dict

        let tokens = text.split(separator: " ", omittingEmptySubsequences: true)
        let tokenStrings = tokens.map { String($0) }
        var entries: [WordEntry] = []
        entries.reserveCapacity(tokenStrings.count)
        for (i, token) in tokenStrings.enumerated() {
            var match = MaddDetector.detect(inWord: token)
            if match == nil, i + 1 < tokenStrings.count {
                if MaddDetector.isMunfasil(endOfWord: token, startOfNextWord: tokenStrings[i + 1]) {
                    match = MaddMatch(type: .munfasil, scalarIndex: token.unicodeScalars.count - 1)
                }
            }
            entries.append(WordEntry(text: token, index: i + 1, madd: match))
        }
        cachedWords = entries
    }

    @ViewBuilder
    private func wordView(entry: WordEntry) -> some View {
        let baseText = tatweelCount > 0 ? applyTatweel(to: entry.text, count: tatweelCount) : entry.text
        let displayText = maddElongatedText(base: baseText, entry: entry)
        let extraTrailing = entry.needsMunfasilSpacing ? CGFloat(10) : CGFloat(0)
        if let analysis = analysisByWord[entry.index], !analysis.segments.isEmpty {
            WordTokenView(
                token: displayText,
                analysis: analysis,
                fontSize: fontSize,
                letterSpacing: letterSpacing
            )
            .padding(.trailing, extraTrailing)
        } else {
            Text(displayText)
                .font(loadQuranFont(size: fontSize))
                .kerning(letterSpacing)
                .padding(.trailing, extraTrailing)
        }
    }

    private func maddElongatedText(base: String, entry: WordEntry) -> String {
        guard let m = entry.madd, m.type != .munfasil else { return base }
        return MaddDetector.applyElongation(to: base, match: m)
    }

    private func applyTatweel(to word: String, count: Int) -> String {
        guard count > 0 else { return word }
        let tatweel = String(repeating: "\u{0640}", count: count)
        let madd: Set<UInt32> = [0x0627, 0x0648, 0x064A, 0x0649]
        var result = ""
        for ch in word {
            result.append(ch)
            if let cp = ch.unicodeScalars.first?.value, madd.contains(cp) {
                result.append(tatweel)
            }
        }
        return result
    }
}

struct WordTokenView: View {
    let token: String
    let analysis: WordAnalysis
    let fontSize: CGFloat
    var letterSpacing: CGFloat = 0
    @State private var isHovering = false

    var body: some View {
        Text(token)
            .font(loadQuranFont(size: fontSize))
            .kerning(letterSpacing)
            .padding(.horizontal, 5)
            .background(isHovering ? Color.accentColor.opacity(0.18) : Color.clear, in: RoundedRectangle(cornerRadius: 4))
            .scaleEffect(isHovering ? 1.04 : 1.0)
            .animation(.easeInOut(duration: 0.10), value: isHovering)
            .onHover { isHovering = $0 }
            .popover(isPresented: $isHovering, arrowEdge: .top) {
                WordMorphologyTooltip(analysis: analysis).padding(14)
            }
            .help(tooltip)
    }

    private var tooltip: String {
        var parts: [String] = [token]
        if let r = analysis.root { parts.append("root: \(r)") }
        if !analysis.gloss.isEmpty { parts.append(analysis.gloss) }
        return parts.joined(separator: " · ")
    }
}

struct WordMorphologyTooltip: View {
    let analysis: WordAnalysis

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(alignment: .firstTextBaseline) {
                Text(analysis.text)
                    .font(.system(size: 26, design: .serif))
                Spacer(minLength: 12)
                if let root = analysis.root {
                    VStack(alignment: .trailing) {
                        Text("Root")
                            .font(.caption2)
                            .foregroundStyle(.secondary)
                        Text(root)
                            .font(.system(size: 18, design: .monospaced))
                            .foregroundStyle(.blue)
                    }
                }
            }

            Divider()

            if let lemma = analysis.lemma {
                LabeledRow(label: "Lemma", value: lemma, mono: true)
            }
            LabeledRow(label: "Gloss", value: analysis.gloss.isEmpty ? "—" : analysis.gloss)

            if !analysis.segments.isEmpty {
                Divider()
                Text("Morphology")
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(.secondary)
                ForEach(Array(analysis.segments.enumerated()), id: \.offset) { _, seg in
                    HStack(alignment: .top, spacing: 8) {
                        Text(seg.segmentedText)
                            .font(.system(size: 14, design: .serif))
                            .frame(minWidth: 60, alignment: .leading)
                        VStack(alignment: .leading, spacing: 1) {
                            if !seg.pos.isEmpty {
                                TagPill(text: seg.pos, color: .blue)
                            }
                            if !seg.morphType.isEmpty {
                                TagPill(text: "type: \(seg.morphType)", color: .gray)
                            }
                            if !seg.syntacticRole.isEmpty {
                                TagPill(text: "role: \(seg.syntacticRole)", color: .green)
                            }
                            if !seg.caseMood.isEmpty {
                                TagPill(text: "case: \(seg.caseMood)", color: .orange)
                            }
                            if !seg.phrase.isEmpty {
                                TagPill(text: "phrase: \(seg.phrase)", color: .purple)
                            }
                        }
                    }
                }
            }
        }
        .frame(maxWidth: 380)
    }
}

// MARK: - Layout

struct FlowLayoutRTL: Layout {
    struct Cache {
        var sizes: [CGSize] = []
    }

    var spacing: CGFloat = 0

    func makeCache(subviews: Subviews) -> Cache {
        Cache(sizes: subviews.map { $0.sizeThatFits(.unspecified) })
    }

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout Cache) -> CGSize {
        let maxWidth = proposal.width ?? .infinity
        var x: CGFloat = 0
        var y: CGFloat = 0
        var rowHeight: CGFloat = 0
        let s = Swift.max(0, spacing)
        for size in cache.sizes {
            if x > 0 && x + size.width > maxWidth {
                x = 0
                y += rowHeight + s
                rowHeight = 0
            }
            x += size.width + s
            rowHeight = Swift.max(rowHeight, size.height)
        }
        return CGSize(width: maxWidth.isFinite ? maxWidth : x, height: y + rowHeight)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout Cache) {
        let s = Swift.max(0, spacing)
        var x: CGFloat = bounds.maxX
        var y: CGFloat = bounds.minY
        var rowWidth: CGFloat = 0
        var rowHeight: CGFloat = 0
        for (i, subview) in subviews.enumerated() {
            let size = cache.sizes[i]
            if rowWidth > 0 && x - size.width < bounds.minX {
                y += rowHeight + s
                x = bounds.maxX
                rowWidth = 0
                rowHeight = 0
            }
            x -= size.width + s
            rowWidth += size.width + s
            subview.place(at: CGPoint(x: x, y: y), proposal: ProposedViewSize(size))
            rowHeight = Swift.max(rowHeight, size.height)
        }
    }
}

// MARK: - Shared small components

struct LabeledRow: View {
    let label: String
    let value: String
    var mono: Bool = false

    var body: some View {
        HStack(alignment: .firstTextBaseline) {
            Text(label)
                .font(.caption.weight(.semibold))
                .foregroundStyle(.secondary)
                .frame(width: 50, alignment: .leading)
            Text(value)
                .font(mono ? .system(size: 13, design: .monospaced) : .callout)
        }
    }
}

struct TagPill: View {
    let text: String
    let color: Color

    var body: some View {
        Text(text)
            .font(.caption.monospaced())
            .padding(.horizontal, 5)
            .padding(.vertical, 1)
            .background(color.opacity(0.15), in: Capsule())
            .foregroundStyle(color)
    }
}
