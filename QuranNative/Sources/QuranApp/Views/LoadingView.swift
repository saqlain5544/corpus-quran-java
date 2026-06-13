import SwiftUI

struct LoadingView: View {
    @Environment(QuranStore.self) private var store

    var body: some View {
        VStack(spacing: 24) {
            // Decorative emblem
            ZStack {
                Circle()
                    .fill(LinearGradient(
                        colors: [Color.accentColor, Color.accentColor.opacity(0.5)],
                        startPoint: .topLeading, endPoint: .bottomTrailing
                    ))
                    .frame(width: 80, height: 80)
                    .opacity(0.2)
                Circle()
                    .strokeBorder(Color.accentColor.opacity(0.4), lineWidth: 2)
                    .frame(width: 80, height: 80)
                    .scaleEffect(store.isLoading ? 1.3 : 1.0)
                    .opacity(store.isLoading ? 0 : 1)
                    .animation(.easeOut(duration: 1.2).repeatForever(autoreverses: false), value: store.isLoading)
                Image(systemName: "book.fill")
                    .font(.system(size: 32, weight: .light))
                    .foregroundStyle(Color.accentColor)
            }

            VStack(spacing: 8) {
                Text("Loading Quran data")
                    .font(.title2.weight(.semibold))
                Text("Parsing morphology and reading text…")
                    .font(.callout)
                    .foregroundStyle(.secondary)
            }

            ProgressView(value: store.loadingProgress)
                .progressViewStyle(.linear)
                .frame(width: 320)
                .tint(.accentColor)

            Text("\(Int(store.loadingProgress * 100))%")
                .font(.caption.monospaced())
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color(nsColor: .windowBackgroundColor))
    }
}
