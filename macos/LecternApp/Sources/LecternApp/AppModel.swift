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
    // A change (and one of its artifacts) belonging to another worktree —
    // rendered read-only. Carries the worktree path so it resolves unambiguously
    // even when two worktrees have a change of the same name.
    case worktreeArtifact(worktreePath: String, changeName: String, kind: ArtifactKind)
}

extension ProjectConfig {
    var hasContent: Bool { !context.isEmpty || (rules?.isEmpty == false) }
}

// A node in the sidebar tree. Rendered by an OutlineGroup, which manages
// disclosure, selection, and triangles natively (leaves have nil children and
// get no triangle) — avoiding the hand-rolled DisclosureGroup glitches.
struct ChangeProgress: Hashable {
    let done: Int
    let total: Int
}

struct SidebarNode: Identifiable, Hashable {
    let id: String
    let title: String
    let subtitle: String?
    let icon: String
    let prominent: Bool
    var progress: ChangeProgress?   // shown as a small bar (worktree change rows)
    var children: [SidebarNode]?

    init(id: String, title: String, subtitle: String?, icon: String, prominent: Bool,
         progress: ChangeProgress? = nil, children: [SidebarNode]? = nil) {
        self.id = id
        self.title = title
        self.subtitle = subtitle
        self.icon = icon
        self.prominent = prominent
        self.progress = progress
        self.children = children
    }
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
    @Published var worktreeChanges: [String: [Change]] = [:]   // worktree path → its active changes
    private var worktreesWithProject: Set<String> = []          // worktrees that have an openspec/ project

    @Published var rootPath: String?
    @Published var loadError: String?

    private let bookmarkKey = "projectBookmark"
    private var accessedURL: URL?
    private var watcher: DirectoryWatcher?

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
        refreshData(resetSelection: false)
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
        refreshData(resetSelection: true)
        startWatching(url.path)
    }

    // Reloads all data from disk, rebuilds the sidebar, and (unless resetting)
    // preserves the current selection if it still exists — used by manual reload
    // and by the FSEvents watcher (task 5.3).
    private func refreshData(resetSelection: Bool) {
        guard let path = rootPath else { return }
        let loader = Loader()
        do {
            project = try loader.loadFrom(path)
            loadError = nil
        } catch {
            project = nil
            loadError = describe(error)
        }
        archivedChanges = (try? loader.listArchiveChangesFrom(path)) ?? []
        projectSpecs = (try? loader.loadProjectSpecsFrom(path)) ?? []
        projectConfig = try? loader.loadConfigFrom(path)
        loadWorktrees(path)

        let previous = selectedNodeID
        rebuildSidebar()
        if !resetSelection, let previous, idToSelection[previous] != nil {
            selectedNodeID = previous
        } else {
            selectDefault()
        }
    }

    private func startWatching(_ root: String) {
        let watched = (root as NSString).appendingPathComponent("openspec")
        let target = FileManager.default.fileExists(atPath: watched) ? watched : root
        watcher = DirectoryWatcher(path: target) { [weak self] in
            Task { @MainActor in self?.refreshData(resetSelection: false) }
        }
    }

    private func loadWorktrees(_ path: String) {
        do {
            let wts = try ProcessGitService().listWorktrees(root: path)
            // Current worktree first; preserve git's order for the rest (stable
            // partition — Swift's sort isn't stable).
            worktrees = wts.filter(\.isCurrent) + wts.filter { !$0.isCurrent }
            worktreesError = nil
        } catch {
            worktrees = []
            worktreesError = "Worktrees unavailable (git not found or not a working tree)."
        }
        // Survey each non-bare worktree's active changes (task 1.1). Track which
        // worktrees actually have an openspec/ project, to distinguish "project
        // with no active changes" from "no project" (6.2).
        var survey: [String: [Change]] = [:]
        var withProject: Set<String> = []
        let loader = Loader()
        for wt in worktrees where !wt.bare {
            if let project = try? loader.loadFrom(wt.path) {
                survey[wt.path] = project.changes
                withProject.insert(wt.path)
            } else {
                survey[wt.path] = []
            }
        }
        worktreeChanges = survey
        worktreesWithProject = withProject
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
            nodes = changes(for: mode).map { c in
                changeNode(c, subtitle: nil, progress: progressOf(c), into: &map) { .artifact(ArtifactRef(changeName: c.name, kind: $0)) }
            }
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
                // Children = this worktree's active changes (read-only), each a
                // change subtree carrying a task-progress bar.
                var changeKids: [SidebarNode] = (worktreeChanges[wt.path] ?? []).map { c in
                    changeNode(c, subtitle: nil, progress: progressOf(c), into: &map) {
                        .worktreeArtifact(worktreePath: wt.path, changeName: c.name, kind: $0)
                    }
                }
                // A worktree that has a project but no active changes gets an
                // explicit affordance, distinct from a bare/no-project worktree.
                if changeKids.isEmpty && worktreesWithProject.contains(wt.path) {
                    let emptyID = "\(id)#empty"
                    map[emptyID] = .worktree(wt.path)
                    changeKids = [SidebarNode(id: emptyID, title: "No active changes", subtitle: nil,
                                              icon: "tray", prominent: false)]
                }
                return SidebarNode(id: id, title: worktreeTitle(wt),
                                   subtitle: worktreeStateLabel(wt),
                                   icon: wt.isCurrent ? "checkmark.circle.fill" : "externaldrive",
                                   prominent: false,
                                   children: changeKids.isEmpty ? nil : changeKids)
            }
        }
        sidebarNodes = nodes
        idToSelection = map
    }

    // Builds a change subtree (change → artifacts). `selection` maps an artifact
    // kind to the Selection the detail resolves — `.artifact` for the current
    // project, `.worktreeArtifact` for a foreign worktree. `subtitle` overrides
    // the row subtitle (e.g. task progress for worktree changes).
    private func changeNode(_ c: Change, subtitle: String?, progress: ChangeProgress?,
                            into map: inout [String: Selection],
                            selection: (ArtifactKind) -> Selection) -> SidebarNode {
        var kids: [SidebarNode] = []

        func leaf(_ suffix: String, _ title: String, _ icon: String, _ kind: ArtifactKind) {
            let id = "\(c.path)#\(suffix)"
            map[id] = selection(kind)
            kids.append(SidebarNode(id: id, title: title, subtitle: nil, icon: icon, prominent: false, children: nil))
        }

        if c.proposal.present { leaf("proposal", "Proposal", "doc.text", .proposal) }
        if c.design.present { leaf("design", "Design", "pencil.and.outline", .design) }
        if !c.specFiles.isEmpty {
            var specKids: [SidebarNode] = []
            for sf in c.specFiles {
                let id = "\(c.path)#spec:\(sf.name)"
                map[id] = selection(.specFile(sf.name))
                specKids.append(SidebarNode(id: id, title: sf.name, subtitle: nil,
                                            icon: "doc.plaintext", prominent: false, children: nil))
            }
            let specsID = "\(c.path)#specs"
            if let first = c.specFiles.first {
                map[specsID] = selection(.specFile(first.name))
            }
            kids.append(SidebarNode(id: specsID, title: "Specs", subtitle: nil,
                                    icon: "folder", prominent: false, children: specKids))
        }
        if c.tasks.present { leaf("tasks", "Tasks", "checklist", .tasks) }

        // Selecting the change row shows its proposal (or design).
        if c.proposal.present {
            map[c.path] = selection(.proposal)
        } else if c.design.present {
            map[c.path] = selection(.design)
        }

        return SidebarNode(id: c.path, title: c.name,
                           subtitle: subtitle ?? (c.displayDate.isEmpty ? nil : c.displayDate),
                           icon: "shippingbox", prominent: true,
                           progress: progress,
                           children: kids.isEmpty ? nil : kids)
    }

    // Task progress for a change, or nil when it has no tasks.
    private func progressOf(_ c: Change) -> ChangeProgress? {
        let tasks = parseTasks(c.tasks.content).filter { $0.kind == .task }
        guard !tasks.isEmpty else { return nil }
        return ChangeProgress(done: tasks.filter(\.done).count, total: tasks.count)
    }

    // Progress of the change behind the current selection (for the persistent
    // detail-pane bar), or nil for non-change selections / task-less changes.
    func currentChangeProgress() -> ChangeProgress? {
        let change: Change?
        switch selection {
        case .artifact(let ref): change = self.change(named: ref.changeName)
        case .worktreeArtifact(let path, let name, _): change = worktreeChange(worktreePath: path, changeName: name)
        default: change = nil
        }
        return change.flatMap { progressOf($0) }
    }

    // A `·`-separated description of the current location (mode, then the
    // selected change/spec/artifact where applicable), shown as the window
    // subtitle beneath the project name (#70). Both the sidebar header and this
    // trail derive from the same model state, so they never drift.
    func locationTrail() -> String {
        var parts: [String] = [mode.rawValue]
        switch selection {
        case .artifact(let ref):
            parts.append(ref.changeName)
            parts.append(artifactLabel(ref.kind))
        case .projectSpec(let name):
            parts.append(name)
        case .config:
            parts.append("Project Config")
        case .worktree(let path):
            if let wt = worktree(path: path) { parts.append(worktreeTitle(wt)) }
        case .worktreeArtifact(let path, let name, let kind):
            if let wt = worktree(path: path) { parts.append(worktreeTitle(wt)) }
            parts.append(name)
            parts.append(artifactLabel(kind))
        case .none:
            break
        }
        return parts.joined(separator: " · ")
    }

    private func artifactLabel(_ kind: ArtifactKind) -> String {
        switch kind {
        case .proposal: return "Proposal"
        case .design: return "Design"
        case .tasks: return "Tasks"
        case .specFile(let name): return name
        }
    }

    // Worktree state flags for the sidebar subtitle (e.g. "current, locked").
    private func worktreeStateLabel(_ wt: Worktree) -> String? {
        let flags = [
            wt.isCurrent ? "current" : nil,
            wt.locked ? "locked" : nil,
            wt.prunable ? "prunable" : nil,
        ].compactMap { $0 }
        return flags.isEmpty ? nil : flags.joined(separator: ", ")
    }

    private func selectDefault() {
        selectedNodeID = sidebarNodes.first?.id
    }

    func worktreeTitle(_ wt: Worktree) -> String {
        if wt.bare { return "(bare)" }
        if wt.detached { return wt.head.isEmpty ? "(detached)" : "detached @ \(wt.head.prefix(7))" }
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

    func worktreeChange(worktreePath: String, changeName: String) -> Change? {
        worktreeChanges[worktreePath]?.first { $0.name == changeName }
    }

    func artifactPath(for kind: ArtifactKind, in change: Change) -> String {
        func join(_ parts: String...) -> String {
            parts.dropFirst().reduce(parts.first ?? "") { ($0 as NSString).appendingPathComponent($1) }
        }
        switch kind {
        case .proposal: return join(change.path, "proposal.md")
        case .design: return join(change.path, "design.md")
        case .tasks: return join(change.path, "tasks.md")
        case .specFile(let name): return join(change.path, "specs", name, "spec.md")
        }
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
            return artifactPath(for: ref.kind, in: c)
        case .worktreeArtifact(let wtPath, let name, let kind):
            guard let c = worktreeChange(worktreePath: wtPath, changeName: name) else { return nil }
            return artifactPath(for: kind, in: c)
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
