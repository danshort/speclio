import SwiftUI
import AppKit

// Identifies a project window by its folder path. Codable + Hashable so it can
// be a value-based WindowGroup value: SwiftUI gives each distinct ref its own
// window (→ single instance per project) and restores open windows across
// launches (→ reopen-all). The path doubles as the key into ProjectStore (#69).
struct ProjectRef: Codable, Hashable {
    var path: String
}

// Security-scoped bookmarks for every project the user has opened, keyed by
// path, plus the folder chooser. Replaces the old single `projectBookmark`
// key so multiple projects can be resolved independently (#69, Option C).
enum ProjectStore {
    private static let bookmarksKey = "projectBookmarks"   // [path: bookmark Data]
    private static let openKey = "openProjects"            // [path] currently open

    // The set of projects currently open in a window, in open order. Persisted
    // so the app can reopen them on next launch (#69) — self-managed rather than
    // relying on macOS window restoration, which needs a signed .app bundle and
    // a clean quit and so never works under `swift run`. Written eagerly on
    // open, so even a hard kill preserves it; entries are removed only when the
    // user closes a window (not on app termination).
    static func openPaths() -> [String] {
        UserDefaults.standard.stringArray(forKey: openKey) ?? []
    }

    static func markOpen(_ path: String) {
        var all = openPaths().filter { $0 != path }
        all.append(path)
        UserDefaults.standard.set(all, forKey: openKey)
    }

    static func markClosed(_ path: String) {
        UserDefaults.standard.set(openPaths().filter { $0 != path }, forKey: openKey)
    }


    static func bookmarks() -> [String: Data] {
        UserDefaults.standard.dictionary(forKey: bookmarksKey) as? [String: Data] ?? [:]
    }

    static func saveBookmark(_ data: Data, for path: String) {
        var all = bookmarks()
        all[path] = data
        UserDefaults.standard.set(all, forKey: bookmarksKey)
    }

    static func bookmark(for path: String) -> Data? { bookmarks()[path] }

    static func makeBookmark(for url: URL) -> Data? {
        try? url.bookmarkData(options: .withSecurityScope,
                              includingResourceValuesForKeys: nil, relativeTo: nil)
    }

    // Prompts for a project folder, persists its bookmark, records it as recent,
    // and returns its ref — or nil if cancelled. The single entry point for
    // "open a project" so every caller shares the same side effects.
    @MainActor
    static func chooseProject() -> ProjectRef? {
        let panel = NSOpenPanel()
        panel.canChooseDirectories = true
        panel.canChooseFiles = false
        panel.allowsMultipleSelection = false
        panel.message = "Choose a project folder containing an openspec/ directory"
        panel.prompt = "Open"
        guard panel.runModal() == .OK, let url = panel.url else { return nil }
        if let data = makeBookmark(for: url) { saveBookmark(data, for: url.path) }
        RecentsStore.shared.add(url.path)
        return ProjectRef(path: url.path)
    }
}

// Observable so the File ▸ Open Recent menu updates live as projects are opened
// or the list is cleared. Most-recent first, de-duplicated, capped.
@MainActor
final class RecentsStore: ObservableObject {
    static let shared = RecentsStore()
    private let key = "recentProjects"
    private let limit = 10

    @Published private(set) var paths: [String]

    private init() {
        paths = UserDefaults.standard.stringArray(forKey: key) ?? []
    }

    func add(_ path: String) {
        var next = paths.filter { $0 != path }
        next.insert(path, at: 0)
        if next.count > limit { next = Array(next.prefix(limit)) }
        paths = next
        persist()
    }

    func clear() {
        paths = []
        persist()
    }

    private func persist() { UserDefaults.standard.set(paths, forKey: key) }
}
