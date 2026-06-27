import SwiftUI
import OpenSpecKit

struct ContentView: View {
    @EnvironmentObject var model: AppModel

    var body: some View {
        NavigationSplitView {
            Sidebar()
                .navigationSplitViewColumnWidth(min: 220, ideal: 280)
        } detail: {
            DetailView()
        }
        .toolbar {
            ToolbarItem(placement: .navigation) {
                Button { model.openPanel() } label: { Label("Open", systemImage: "folder") }
            }
            ToolbarItem(placement: .principal) {
                Picker("Mode", selection: $model.mode) {
                    ForEach(Mode.allCases) { Text($0.rawValue).tag($0) }
                }
                .pickerStyle(.segmented)
                .labelsHidden()
                .fixedSize()
                .disabled(model.project == nil)
                .accessibilityLabel("View mode")
            }
            ToolbarItem {
                Menu {
                    Button("Reveal in Finder") { revealInFinder() }
                    Button("Open in Default App") { openInEditor() }
                } label: {
                    Label("File actions", systemImage: "ellipsis.circle")
                }
                .disabled(model.currentFilePath() == nil)
            }
            ToolbarItem {
                Button { model.reload() } label: { Label("Reload", systemImage: "arrow.clockwise") }
                    .disabled(model.project == nil)
            }
        }
    }

    private func revealInFinder() {
        guard let path = model.currentFilePath() else { return }
        NSWorkspace.shared.activateFileViewerSelecting([URL(fileURLWithPath: path)])
    }

    private func openInEditor() {
        guard let path = model.currentFilePath() else { return }
        NSWorkspace.shared.open(URL(fileURLWithPath: path))
    }
}

// MARK: - Sidebar

struct Sidebar: View {
    @EnvironmentObject var model: AppModel

    var body: some View {
        Group {
            if model.project == nil {
                EmptyProjectState()
            } else if model.sidebarNodes.isEmpty {
                emptyMode
            } else {
                List(selection: $model.selectedNodeID) {
                    OutlineGroup(model.sidebarNodes, children: \.children) { node in
                        SidebarRow(node: node)
                    }
                }
            }
        }
        .frame(minWidth: 220)
    }

    private var emptyMode: some View {
        VStack(spacing: 8) {
            Text(model.mode == .worktrees ? (model.worktreesError ?? "No worktrees") : "Nothing here")
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .padding()
    }
}

struct SidebarRow: View {
    let node: SidebarNode

    var body: some View {
        Label {
            VStack(alignment: .leading, spacing: 2) {
                Text(node.title)
                    .font(node.prominent ? .headline : .body)
                    .lineLimit(1)
                if let subtitle = node.subtitle {
                    Text(subtitle).font(.caption).foregroundStyle(.secondary)
                }
                if let p = node.progress {
                    HStack(spacing: 6) {
                        ProgressView(value: Double(p.done), total: Double(max(p.total, 1)))
                            .progressViewStyle(.linear)
                            .frame(width: 70)
                        Text("\(p.done)/\(p.total)").font(.caption2).foregroundStyle(.secondary)
                    }
                    .accessibilityLabel("\(p.done) of \(p.total) tasks complete")
                }
            }
        } icon: {
            Image(systemName: node.icon)
        }
        .help(node.title)
    }
}

struct EmptyProjectState: View {
    @EnvironmentObject var model: AppModel

    var body: some View {
        VStack(spacing: 12) {
            Image(systemName: "books.vertical")
                .font(.system(size: 40))
                .foregroundStyle(.secondary)
            Text("No project open").font(.headline)
            Button("Open Project…") { model.openPanel() }
            if let err = model.loadError {
                Text(err)
                    .font(.callout)
                    .foregroundStyle(.red)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .padding()
    }
}

// MARK: - Detail

struct DetailView: View {
    @EnvironmentObject var model: AppModel

    var body: some View {
        VStack(spacing: 0) {
            if let progress = model.currentChangeProgress() {
                ChangeProgressBar(progress: progress)
            }
            detailContent
        }
    }

    @ViewBuilder
    private var detailContent: some View {
        switch model.selection {
        case .artifact(let ref):
            if let change = model.change(named: ref.changeName) {
                ArtifactDetail(change: change, ref: ref)
                    .navigationTitle(change.name)
            } else {
                Placeholder()
            }
        case .projectSpec(let name):
            if let spec = model.projectSpec(named: name) {
                SpecContentView(
                    artifact: Artifact(content: spec.content, present: true, readError: spec.readError),
                    issues: spec.readError ? [] : validateSpec(spec.content)
                )
                .id("spec:\(name)")
                .navigationTitle(name)
            } else {
                Placeholder()
            }
        case .worktree(let path):
            if let wt = model.worktree(path: path) {
                WorktreeDetail(worktree: wt).navigationTitle("Worktrees")
            } else {
                Placeholder()
            }
        case .config:
            if let cfg = model.projectConfig {
                ScrollableContent { MarkdownView(configToMarkdown(cfg)) }
                    .navigationTitle("Project Config")
            } else {
                Placeholder()
            }
        case .worktreeArtifact(let wtPath, let name, let kind):
            if let change = model.worktreeChange(worktreePath: wtPath, changeName: name) {
                WorktreeArtifactView(change: change, kind: kind)
                    .id("\(wtPath)#\(name)#\(kind)")
                    .navigationTitle(name)
                    .navigationSubtitle("Read-only · \(model.worktree(path: wtPath).map(model.worktreeTitle) ?? "worktree")")
            } else {
                Placeholder()
            }
        case .none:
            Placeholder()
        }
    }
}

// A foreign worktree's change artifact, rendered read-only: specs reuse the
// (non-writing) SpecContentView; Tasks reuse TasksView in read-only mode
// (progress bar + check state, but not toggleable), since cross-worktree writes
// are out of scope.
struct WorktreeArtifactView: View {
    @EnvironmentObject var model: AppModel
    let change: Change
    let kind: ArtifactKind

    var body: some View {
        let artifact = model.artifact(for: ArtifactRef(changeName: change.name, kind: kind), in: change)
        switch kind {
        case .specFile:
            SpecContentView(artifact: artifact, issues: artifact.readError ? [] : validateChange(change))
        case .tasks:
            // Read-only checklist: progress bar + check state, but not toggleable.
            TasksView(changePath: change.path, content: artifact.content, readOnly: true)
        default:
            ScrollableContent { ArtifactBody(artifact: artifact) }
        }
    }
}

// A thin progress bar pinned at the top of the detail pane (below the toolbar),
// showing the current change's overall task completion regardless of which
// artifact is open (#65).
struct ChangeProgressBar: View {
    let progress: ChangeProgress

    var body: some View {
        VStack(spacing: 0) {
            HStack(spacing: 8) {
                ProgressView(value: Double(progress.done), total: Double(max(progress.total, 1)))
                    .progressViewStyle(.linear)
                Text("\(progress.done)/\(progress.total)")
                    .font(.caption).monospacedDigit().foregroundStyle(.secondary)
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 6)
            Divider()
        }
        .background(.bar)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("Change progress: \(progress.done) of \(progress.total) tasks complete")
    }
}

struct ArtifactDetail: View {
    @EnvironmentObject var model: AppModel
    let change: Change
    let ref: ArtifactRef

    var body: some View {
        switch ref.kind {
        case .specFile(let name):
            let artifact = model.artifact(for: ref, in: change)
            SpecContentView(artifact: artifact, issues: artifact.readError ? [] : validateChange(change))
                .id("\(change.name)/spec/\(name)")
        case .tasks:
            TasksView(changePath: change.path, content: change.tasks.content)
                .id("\(change.path)/tasks")
        default:
            ScrollableContent { ArtifactBody(artifact: model.artifact(for: ref, in: change)) }
        }
    }
}

// Interactive Tasks view: renders tasks.md as toggleable checkboxes. Toggling
// re-reads and re-parses the file, finds the intended task by text, then writes
// via OpenSpecKit.toggleTask (CRLF-safe), so an external edit between render and
// click can't flip the wrong line (task 5.1).
struct TasksView: View {
    let changePath: String
    let content: String   // from the model; changes when the file is reloaded (incl. external edits)
    var readOnly: Bool = false   // worktree changes render checkboxes + progress but don't toggle

    @State private var items: [TaskItem] = []
    @State private var errorText: String?

    private var tasksPath: String { (changePath as NSString).appendingPathComponent("tasks.md") }

    var body: some View {
        ScrollableContent {
            if let errorText {
                Label(errorText, systemImage: "exclamationmark.triangle.fill")
                    .foregroundStyle(.orange).font(.callout)
            }
            // Overall change progress lives in the persistent bar at the top of
            // the detail pane (#65); the Tasks view shows only per-section bars.
            ForEach(Array(items.enumerated()), id: \.offset) { index, item in
                if item.kind == .section {
                    HStack(alignment: .firstTextBaseline) {
                        Text(item.text).font(.title3).bold()
                        Spacer(minLength: 12)
                        if let sp = sectionProgress(startingAfter: index) {
                            HStack(spacing: 6) {
                                ProgressView(value: Double(sp.done), total: Double(max(sp.total, 1)))
                                    .progressViewStyle(.linear)
                                    .frame(width: 70)
                                Text("\(sp.done)/\(sp.total)").font(.caption2).monospacedDigit().foregroundStyle(.secondary)
                            }
                        }
                    }
                    .padding(.top, index == 0 ? 0 : 12)
                } else if readOnly {
                    taskRow(item)
                        .accessibilityLabel(item.text)
                        .accessibilityValue(item.done ? "completed" : "not completed")
                } else {
                    Button { toggle(item) } label: { taskRow(item) }
                        .buttonStyle(.plain)
                        .accessibilityLabel(item.text)
                        .accessibilityValue(item.done ? "completed" : "not completed")
                        .accessibilityHint("Toggles this task")
                }
            }
        }
        .onAppear { items = parseTasks(content) }
        .onChange(of: content) { newContent in items = parseTasks(newContent) }
    }

    // Completed/total of the tasks belonging to the section at `index` (the
    // tasks until the next section), or nil if the section has no tasks.
    private func sectionProgress(startingAfter index: Int) -> ChangeProgress? {
        var done = 0, total = 0
        var i = index + 1
        while i < items.count, items[i].kind != .section {
            if items[i].kind == .task {
                total += 1
                if items[i].done { done += 1 }
            }
            i += 1
        }
        return total > 0 ? ChangeProgress(done: done, total: total) : nil
    }

    private func taskRow(_ item: TaskItem) -> some View {
        HStack(alignment: .firstTextBaseline, spacing: 8) {
            Image(systemName: item.done ? "checkmark.square.fill" : "square")
                .foregroundStyle(item.done ? Color.accentColor : .secondary)
            Text(item.text)
                .strikethrough(item.done, color: .secondary)
                .foregroundStyle(item.done ? .secondary : .primary)
        }
    }

    private func toggle(_ item: TaskItem) {
        do {
            items = try toggleTaskByText(tasksPath, item.text)
            errorText = nil
        } catch {
            errorText = "Couldn't write tasks.md: \(error.localizedDescription)"
        }
    }
}

// Renders a spec with a requirement-focus picker: "All requirements" shows the
// full spec (with its validation banner); picking one shows just that
// requirement's extracted block (task 4.9, via OpenSpecKit.extractRequirement).
struct SpecContentView: View {
    let artifact: Artifact
    let issues: [String]
    @State private var focused: String?

    var body: some View {
        ScrollableContent {
            if artifact.readError {
                Label(artifact.content, systemImage: "exclamationmark.triangle.fill")
                    .foregroundStyle(.orange)
                    .font(.callout)
            } else if !artifact.present {
                Text("This artifact is not present.").foregroundStyle(.secondary)
            } else {
                let names = requirementNames(in: artifact.content)
                if !names.isEmpty {
                    focusPicker(names)
                }
                if focused == nil {
                    ValidationBanner(issues: issues)
                }
                MarkdownView(focused.map { extractRequirement(artifact.content, $0) } ?? artifact.content)
            }
        }
    }

    private func focusPicker(_ names: [String]) -> some View {
        HStack(spacing: 6) {
            Image(systemName: "scope").foregroundStyle(.secondary)
            Picker("Focus requirement", selection: $focused) {
                Text("All requirements").tag(String?.none)
                ForEach(names, id: \.self) { name in
                    Text(name).tag(String?.some(name))
                }
            }
            .labelsHidden()
            .fixedSize()
        }
    }
}

// Requirement names for the focus picker — presentation helper mirroring
// extractRequirement's matching (`### Requirement:` prefix, trimmed, non-empty).
func requirementNames(in content: String) -> [String] {
    content.components(separatedBy: "\n").compactMap { raw in
        let line = raw.hasSuffix("\r") ? String(raw.dropLast()) : raw
        guard line.hasPrefix("### Requirement:") else { return nil }
        let name = line.dropFirst("### Requirement:".count).trimmingCharacters(in: .whitespaces)
        return name.isEmpty ? nil : name
    }
}

struct ArtifactBody: View {
    let artifact: Artifact

    var body: some View {
        if artifact.readError {
            Label(artifact.content, systemImage: "exclamationmark.triangle.fill")
                .foregroundStyle(.orange)
                .font(.callout)
        } else if !artifact.present {
            Text("This artifact is not present.").foregroundStyle(.secondary)
        } else {
            MarkdownView(artifact.content)
        }
    }
}

struct WorktreeDetail: View {
    let worktree: Worktree

    var body: some View {
        ScrollableContent {
            VStack(alignment: .leading, spacing: 10) {
                field("Path", worktree.path)
                field("Branch", worktree.branch.isEmpty ? "—" : worktree.branch)
                field("HEAD", worktree.head.isEmpty ? "—" : worktree.head)
                let flags = [
                    worktree.isCurrent ? "current" : nil,
                    worktree.detached ? "detached" : nil,
                    worktree.bare ? "bare" : nil,
                    worktree.locked ? "locked" : nil,
                    worktree.prunable ? "prunable" : nil,
                ].compactMap { $0 }
                field("Flags", flags.isEmpty ? "—" : flags.joined(separator: ", "))
            }
        }
    }

    private func field(_ label: String, _ value: String) -> some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(label).font(.caption).foregroundStyle(.secondary)
            Text(value).font(.system(.body, design: .monospaced)).textSelection(.enabled)
        }
    }
}

struct ValidationBanner: View {
    let issues: [String]

    var body: some View {
        if !issues.isEmpty {
            VStack(alignment: .leading, spacing: 4) {
                Label("Validation issues", systemImage: "exclamationmark.triangle.fill")
                    .font(.callout.bold())
                    .foregroundStyle(.orange)
                ForEach(issues, id: \.self) { issue in
                    Text("• \(issue)").font(.callout)
                }
            }
            .padding(12)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(Color.orange.opacity(0.12))
            .clipShape(RoundedRectangle(cornerRadius: 8))
        }
    }
}

// Shared scroll container for detail content.
struct ScrollableContent<Content: View>: View {
    @ViewBuilder var content: Content

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 14) {
                content
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(24)
        }
    }
}

struct Placeholder: View {
    var body: some View {
        VStack(spacing: 8) {
            Image(systemName: "doc.text.magnifyingglass")
                .font(.system(size: 36))
                .foregroundStyle(.secondary)
            Text("Select an item").foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}
