import SwiftUI
import AppKit

// LecternApp is the native macOS reader for OpenSpec artifacts. It reuses the
// OpenSpecKit domain layer (the Swift port of Go internal/openspec) and renders
// with swift-markdown. Per the App Sandbox decision (Option C), it is
// non-sandboxed but accesses each project through a security-scoped bookmark, so
// a future sandbox flip changes nothing here.
//
// Multiple projects can be open at once (#69): the scene is a value-based
// WindowGroup keyed by ProjectRef, so each window/tab owns its own AppModel and
// project, and SwiftUI restores the open set across launches.

final class AppDelegate: NSObject, NSApplicationDelegate {
    // Set at the start of quit, before windows tear down, so a window's teardown
    // can tell "user closed me" (drop from reopen-all) from "app is quitting"
    // (keep me) — see AppModel.teardown / ProjectStore (#69).
    static var isTerminating = false

    func applicationDidFinishLaunching(_ notification: Notification) {
        // A SwiftPM executable launches as an accessory by default; promote it
        // to a regular app so it gets a Dock icon, menu bar, and key focus.
        NSApp.setActivationPolicy(.regular)
        NSApp.activate(ignoringOtherApps: true)
    }

    func applicationShouldTerminate(_ sender: NSApplication) -> NSApplication.TerminateReply {
        AppDelegate.isTerminating = true
        return .terminateNow
    }

    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool { true }
}

// One-shot guard so only the first (launch) window performs reopen-all.
@MainActor enum LaunchRestore { static var done = false }

@main
struct LecternApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) private var delegate
    @AppStorage(ContentFont.storageKey) private var contentFontScale = ContentFont.defaultScale

    var body: some Scene {
        WindowGroup(for: ProjectRef.self) { $ref in
            RootView(ref: ref)
                .frame(minWidth: 820, minHeight: 520)
        }
        .commands {
            ProjectCommands()
            // Content text-size shortcuts (View menu), sharing the same stored
            // scale as the Settings slider. Global across all windows.
            CommandGroup(after: .toolbar) {
                Button("Increase Text Size") {
                    contentFontScale = ContentFont.clamp(contentFontScale + ContentFont.step)
                }
                .keyboardShortcut("+", modifiers: .command)
                // Also accept ⌘= (no Shift) — the key most people press to zoom
                // in; "+" alone only matches ⌘⇧=.
                Button("Increase Text Size") {
                    contentFontScale = ContentFont.clamp(contentFontScale + ContentFont.step)
                }
                .keyboardShortcut("=", modifiers: .command)

                Button("Decrease Text Size") {
                    contentFontScale = ContentFont.clamp(contentFontScale - ContentFont.step)
                }
                .keyboardShortcut("-", modifiers: .command)

                Button("Actual Size") { contentFontScale = ContentFont.defaultScale }
                    .keyboardShortcut("0", modifiers: .command)
                Divider()
            }
        }

        Settings {
            GeneralSettingsView()
        }
    }
}

// The per-window root: owns one AppModel, loads the window's project, and
// publishes its model to menu commands via the focus environment (#69).
struct RootView: View {
    let ref: ProjectRef?
    @StateObject private var model = AppModel()
    @Environment(\.openWindow) private var openWindow
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        ContentView()
            .environmentObject(model)
            .focusedSceneObject(model)
            .onAppear {
                if let ref {
                    model.load(path: ref.path)
                } else {
                    reopenProjectsOnLaunch()
                }
            }
            .onChange(of: ref) { newRef in if let newRef { model.load(path: newRef.path) } }
            .onDisappear { model.teardown() }
    }

    // The launch window has no ref. If projects were open last time, reopen one
    // window per project (single-instance dedups any the OS already restored) and
    // close this empty launch window (#69).
    private func reopenProjectsOnLaunch() {
        guard !LaunchRestore.done else { return }
        LaunchRestore.done = true
        let paths = ProjectStore.openPaths()
        guard !paths.isEmpty else { return }
        for path in paths { openWindow(value: ProjectRef(path: path)) }
        dismiss()
    }
}

// File-menu project commands, routed to the focused window where relevant (#69).
// Open reuses the focused empty window, otherwise opens a new window; Open
// Recent (and any new window) goes through openWindow(value:), which focuses an
// already-open project instead of duplicating it.
struct ProjectCommands: Commands {
    @Environment(\.openWindow) private var openWindow
    @FocusedObject private var model: AppModel?
    @ObservedObject private var recents = RecentsStore.shared

    var body: some Commands {
        CommandGroup(replacing: .newItem) {
            Button("Open Project…") { openProject() }
                .keyboardShortcut("o", modifiers: .command)

            Menu("Open Recent") {
                ForEach(recents.paths, id: \.self) { path in
                    Button((path as NSString).lastPathComponent) {
                        openWindow(value: ProjectRef(path: path))
                    }
                    .help(path)
                }
                if !recents.paths.isEmpty {
                    Divider()
                    Button("Clear Menu") { recents.clear() }
                }
            }
            .disabled(recents.paths.isEmpty)
        }
    }

    private func openProject() {
        guard let ref = ProjectStore.chooseProject() else { return }
        if let model, model.project == nil {
            model.load(path: ref.path)   // reuse the focused empty window
        } else {
            openWindow(value: ref)        // new window, or focus if already open
        }
    }
}
