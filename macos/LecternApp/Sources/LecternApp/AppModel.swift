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

// What a selected sidebar node resolves to in the detail pane.
enum Selection: Hashable {
    case artifact(ArtifactRef)
    case projectSpec(String)
    case worktree(String)
    case config
}

extension ProjectConfig {
    var hasContent: Bool { !context.isEmpty || (rules?.isEmpty == false) }
}

// A node in the sidebar tree. Rendered by an OutlineGroup, which manages
// disclosure, selection, and triangles natively (leaves have nil children and
// get no triangle) — avoiding the hand-rolled DisclosureGroup glitches.
struct SidebarNode: Identifiable, Hashable {
    let id: String
    let title: String
    let subtitle: String?
    let icon: String
    let prominent: Bool
    var children: [SidebarNode]?
}

@MainActor
final class AppModel: ObservableObject {
    @Published var mode: Mode = .activeChanges {
        didSet {
            if mode != oldValue {
                rebuildSidebar()
                selectDefault()
            }
        }
    }

    // List selection binds to the node id; `selection` is derived from it.
    @Published var selectedNodeID: String? {
        didSet { selection = selectedNodeID.flatMap { idToSelection[$0] } }
    }
    @Published private(set) var selection: Selection?

    @Published private(set) var sidebarNodes: [SidebarNode] = []
    private var idToSelection: [String: Selection] = [:]

    @Published var project: Project?
    @Published var archivedChanges: [Change] = []
    @Published var projectSpecs: [ProjectSpec] = []
    @Published var projectConfig: ProjectConfig?
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
        projectConfig = try? loader.loadConfigFrom(url.path)
        loadWorktrees(url.path)

        rebuildSidebar()
        selectDefault()
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

    // MARK: - Sidebar tree

    func changes(for mode: Mode) -> [Change] {
        switch mode {
        case .activeChanges: return project?.changes ?? []
        case .archivedChanges: return archivedChanges
        default: return []
        }
    }

    private func rebuildSidebar() {
        var map: [String: Selection] = [:]
        var nodes: [SidebarNode]
        switch mode {
        case .activeChanges, .archivedChanges:
            nodes = changes(for: mode).map { changeNode($0, into: &map) }
        case .specs:
            nodes = []
            if let cfg = projectConfig, cfg.hasContent {
                map["config"] = .config
                nodes.append(SidebarNode(id: "config", title: "Project Config", subtitle: nil,
                                         icon: "gearshape", prominent: false, children: nil))
            }
            nodes.append(contentsOf: projectSpecs.map { spec in
                let id = "spec:\(spec.name)"
                map[id] = .projectSpec(spec.name)
                let count = spec.requirementCount
                return SidebarNode(id: id, title: spec.name,
                                   subtitle: count > 0 ? "\(count) requirement\(count == 1 ? "" : "s")" : nil,
                                   icon: "doc.plaintext", prominent: false, children: nil)
            })
        case .worktrees:
            nodes = worktrees.map { wt in
                let id = "wt:\(wt.path)"
                map[id] = .worktree(wt.path)
                return SidebarNode(id: id, title: worktreeTitle(wt),
                                   subtitle: wt.isCurrent ? "current" : (wt.branch.isEmpty ? nil : wt.branch),
                                   icon: wt.isCurrent ? "checkmark.circle.fill" : "externaldrive",
                                   prominent: false, children: nil)
            }
        }
        sidebarNodes = nodes
        idToSelection = map
    }

    private func changeNode(_ c: Change, into map: inout [String: Selection]) -> SidebarNode {
        var kids: [SidebarNode] = []

        func leaf(_ suffix: String, _ title: String, _ icon: String, _ kind: ArtifactKind) {
            let id = "\(c.path)#\(suffix)"
            map[id] = .artifact(ArtifactRef(changeName: c.name, kind: kind))
            kids.append(SidebarNode(id: id, title: title, subtitle: nil, icon: icon, prominent: false, children: nil))
        }

        if c.proposal.present { leaf("proposal", "Proposal", "doc.text", .proposal) }
        if c.design.present { leaf("design", "Design", "pencil.and.outline", .design) }
        if !c.specFiles.isEmpty {
            var specKids: [SidebarNode] = []
            for sf in c.specFiles {
                let id = "\(c.path)#spec:\(sf.name)"
                map[id] = .artifact(ArtifactRef(changeName: c.name, kind: .specFile(sf.name)))
                specKids.append(SidebarNode(id: id, title: sf.name, subtitle: nil,
                                            icon: "doc.plaintext", prominent: false, children: nil))
            }
            let specsID = "\(c.path)#specs"
            if let first = c.specFiles.first {
                map[specsID] = .artifact(ArtifactRef(changeName: c.name, kind: .specFile(first.name)))
            }
            kids.append(SidebarNode(id: specsID, title: "Specs", subtitle: nil,
                                    icon: "folder", prominent: false, children: specKids))
        }
        if c.tasks.present { leaf("tasks", "Tasks", "checklist", .tasks) }

        // Selecting the change row shows its proposal (or design).
        if c.proposal.present {
            map[c.path] = .artifact(ArtifactRef(changeName: c.name, kind: .proposal))
        } else if c.design.present {
            map[c.path] = .artifact(ArtifactRef(changeName: c.name, kind: .design))
        }

        return SidebarNode(id: c.path, title: c.name,
                           subtitle: c.displayDate.isEmpty ? nil : c.displayDate,
                           icon: "shippingbox", prominent: true,
                           children: kids.isEmpty ? nil : kids)
    }

    private func selectDefault() {
        selectedNodeID = sidebarNodes.first?.id
    }

    func worktreeTitle(_ wt: Worktree) -> String {
        if wt.bare { return "(bare)" }
        if wt.detached { return "(detached)" }
        return wt.branch.isEmpty ? (wt.path as NSString).lastPathComponent : wt.branch
    }

    // MARK: - Resolving selections for the detail pane

    func change(named name: String) -> Change? {
        changes(for: mode).first { $0.name == name }
    }

    func projectSpec(named name: String) -> ProjectSpec? {
        projectSpecs.first { $0.name == name }
    }

    func worktree(path: String) -> Worktree? {
        worktrees.first { $0.path == path }
    }

    // The on-disk file (or directory) backing the current selection, for
    // reveal-in-Finder / open-in-editor.
    func currentFilePath() -> String? {
        func join(_ parts: String...) -> String {
            parts.dropFirst().reduce(parts.first ?? "") { ($0 as NSString).appendingPathComponent($1) }
        }
        switch selection {
        case .artifact(let ref):
            guard let c = change(named: ref.changeName) else { return nil }
            switch ref.kind {
            case .proposal: return join(c.path, "proposal.md")
            case .design: return join(c.path, "design.md")
            case .tasks: return join(c.path, "tasks.md")
            case .specFile(let name): return join(c.path, "specs", name, "spec.md")
            }
        case .projectSpec(let name):
            guard let root = rootPath else { return nil }
            return join(root, "openspec", "specs", name, "spec.md")
        case .config:
            guard let root = rootPath else { return nil }
            return join(root, "openspec", "config.yaml")
        case .worktree(let path):
            return path
        case .none:
            return nil
        }
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
