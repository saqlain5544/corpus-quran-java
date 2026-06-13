import AppKit
import SwiftUI

@MainActor
final class AppDelegate: NSObject, NSApplicationDelegate {
    let store = QuranStore()
    var window: NSWindow?

    func applicationDidFinishLaunching(_ notification: Notification) {
        // The content view, so we can manage fullscreen / split-view via collectionBehavior
        let content = ContentView().environment(store)
        let hosting = NSHostingController(rootView: content)
        hosting.view.translatesAutoresizingMaskIntoConstraints = true
        hosting.view.autoresizingMask = [.width, .height]

        let window = NSWindow(contentViewController: hosting)
        window.setContentSize(NSSize(width: 1200, height: 800))
        window.title = "Quran Corpus"
        window.setFrameAutosaveName("QuranMainWindow")
        window.titlebarAppearsTransparent = true
        window.titleVisibility = .visible
        window.isMovableByWindowBackground = false
        window.minSize = NSSize(width: 900, height: 600)

        // Enable fullscreen (green button) and macOS window-tiling snap zones
        // (left/right dock from Window menu and ⌃⌥ ←/→).  This collection
        // behavior makes the window participate in Mission Control spaces
        // and expose the full-screen toggle.
        window.collectionBehavior = [
            .fullScreenPrimary,
            .managed,
            .participatesInCycle
        ]
        window.tabbingMode = .disallowed
        // Show the green fullscreen button in the traffic-light row
        window.standardWindowButton(.zoomButton)?.isEnabled = true
        window.standardWindowButton(.zoomButton)?.isHidden = false

        // Use the unified system title bar look
        window.toolbarStyle = .unified
        window.styleMask.insert([.fullSizeContentView, .titled, .closable, .miniaturizable, .resizable])

        // Window controllers (file / new window) — but we start with just this one
        window.center()
        window.makeKeyAndOrderFront(nil)

        // Activate the app so its menu bar shows
        NSApp.activate(ignoringOtherApps: true)

        self.window = window

        Task { await store.load() }
    }

    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool {
        true
    }
}

let app = NSApplication.shared
let delegate = AppDelegate()
app.delegate = delegate
app.run()
