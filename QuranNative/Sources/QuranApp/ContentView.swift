import SwiftUI
import QuranKit

struct ContentView: View {
    @Environment(QuranStore.self) private var store
    @State private var selectedSection: AppSection = .reader

    var body: some View {
        NavigationSplitView {
            SidebarView(selectedSection: Binding(
                get: { selectedSection },
                set: { newValue in
                    selectedSection = newValue
                    store.readerSection = newValue
                }
            ))
            .frame(minWidth: 220)
        } detail: {
            if !store.isLoaded {
                LoadingView()
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else if let error = store.loadError {
                ErrorView(message: error) { Task { await store.load() } }
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                contentForSelection
            }
        }
        .navigationSplitViewStyle(.balanced)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .onChange(of: store.readerSection) { _, new in
            if selectedSection != new {
                selectedSection = new
            }
        }
        .onAppear { NSApp.keyWindow?.makeFirstResponder(nil) }
    }

    @ViewBuilder
    private var contentForSelection: some View {
        switch selectedSection {
        case .reader:
            ReaderView()
        case .search:
            SearchView()
        case .roots:
            RootsView()
        case .divineNames:
            DivineNamesView()
        case .stats:
            StatsView()
        }
    }
}
