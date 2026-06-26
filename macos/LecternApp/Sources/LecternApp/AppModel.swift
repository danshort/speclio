import SwiftUI
import AppKit
import OpenSpecKit

// Top-level navigation modes, shown in the toolbar segmented switcher.
enum Mode: String, CaseIterable, Identifiable {
    case activeChanges = "Active"
    case archivedChanges = "Archived"
    case specs = "Specs"
    case worktrees = "Worktrees"

    var id: String { rawValue }
}

// Identifies a single artifact within a change (active or archived).
enum ArtifactKind: Hashable {
    case proposal
    case design
    case tasks
    case specFile(String)
}

struct ArtifactRef: Hashable {
    let changeName: String
    let kind: ArtifactKind
}

// A unified sidebar selection across the heterogeneous modes.
enum Selection: Hashable {
    case artifact(ArtifactRef)
    case projectSpec(String)
    case worktree(String)
}

@MainActor
final class AppModel: ObservableObject {
    @Published var mode: Mode = .activeChanges {
        didSet { if mode != oldValue { selection = defaultSelection(for: mode) } }
    }
    @Published var selection: Selection?

    @Published var project: Project?
    @Published var archivedChanges: [Change] = []
    @Published var projectSpecs: [ProjectSpec] = []
    @Published var worktrees: [Worktree] = []
    @Published var worktreesError: String?

    @Published var rootPath: String?
    @Published var loadError: String?

    private let bookmarkKey = "projectBookmark"
    private var accessedURL: URL?

    init() {
        restoreBookmark()
    }

    // MARK: - Opening / restoring the project (security-scoped bookmark)

    func openPanel() {
        let panel = NSOpenPanel()
        panel.canChooseDirectories = true
        panel.canChooseFiles = false
        panel.allowsMultipleSelection = false
        panel.message = "Choose a project folder containing an openspec/ directory"
        panel.prompt = "Open"
        guard panel.runModal() == .OK, let url = panel.url else { return }
        persistBookmark(for: url)
        load(url)
    }

    func reload() {
        if let path = rootPath { load(URL(fileURLWithPath: path)) }
    }

    private func persistBookmark(for url: URL) {
        if let data = try? url.bookmarkData(options: .withSecurityScope,
                                            includingResourceValuesForKeys: nil, relativeTo: nil) {
            UserDefaults.standard.set(data, forKey: bookmarkKey)
        }
    }

    private func restoreBookmark() {
        guard let data = UserDefaults.standard.data(forKey: bookmarkKey) else { return }
        var stale = false
        guard let url = try? URL(resolvingBookmarkData: data, options: .withSecurityScope,
                                 relativeTo: nil, bookmarkDataIsStale: &stale) else { return }
        load(url)
        if stale { persistBookmark(for: url) }
    }

    private func load(_ url: URL) {
        accessedURL?.stopAccessingSecurityScopedResource()
        _ = url.startAccessingSecurityScopedResource()
        accessedURL = url
        rootPath = url.path

        let loader = Loader()
        do {
            project = try loader.loadFrom(url.path)
            loadError = nil
        } catch {
            project = nil
            loadError = describe(error)
        }
        archivedChanges = (try? loader.listArchiveChangesFrom(url.path)) ?? []
        projectSpecs = (try? loader.loadProjectSpecsFrom(url.path)) ?? []
        loadWorktrees(url.path)

        selection = defaultSelection(for: mode)
    }

    private func loadWorktrees(_ path: String) {
        do {
            worktrees = try ProcessGitService().listWorktrees(root: path)
            worktreesError = nil
        } catch {
            worktrees = []
            worktreesError = "Worktrees unavailable (git not found or not a working tree)."
        }
    }

    private func describe(_ error: Error) -> String {
        switch error {
        case LoaderError.noOpenspecDir(let root): return "No openspec/ directory found in \(root)"
        default: return "\(error)"
        }
    }

    // MARK: - Per-mode data + default selection

    func changes(for mode: Mode) -> [Change] {
        switch mode {
        case .activeChanges: return project?.changes ?? []
        case .archivedChanges: return archivedChanges
        default: return []
        }
    }

    private func defaultSelection(for mode: Mode) -> Selection? {
        switch mode {
        case .activeChanges, .archivedChanges:
            guard let first = changes(for: mode).first else { return nil }
            return .artifact(ArtifactRef(changeName: first.name, kind: firstArtifactKind(first)))
        case .specs:
            guard let first = projectSpecs.first else { return nil }
            return .projectSpec(first.name)
        case .worktrees:
            guard let first = worktrees.first else { return nil }
            return .worktree(first.path)
        }
    }

    private func firstArtifactKind(_ change: Change) -> ArtifactKind {
        if change.proposal.present { return .proposal }
        if change.design.present { return .design }
        if let sf = change.specFiles.first { return .specFile(sf.name) }
        return .tasks
    }

    // MARK: - Resolving selections

    func change(named name: String) -> Change? {
        changes(for: mode).first { $0.name == name }
    }

    func projectSpec(named name: String) -> ProjectSpec? {
        projectSpecs.first { $0.name == name }
    }

    func worktree(path: String) -> Worktree? {
        worktrees.first { $0.path == path }
    }

    func artifact(for ref: ArtifactRef, in change: Change) -> Artifact {
        switch ref.kind {
        case .proposal: return change.proposal
        case .design: return change.design
        case .tasks: return change.tasks
        case .specFile(let name):
            if let sf = change.specFiles.first(where: { $0.name == name }) {
                return Artifact(content: sf.content, present: true, readError: sf.readError)
            }
            return Artifact()
        }
    }
}
