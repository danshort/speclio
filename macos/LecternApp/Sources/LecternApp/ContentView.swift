import SwiftUI
import OpenSpecKit

struct ContentView: View {
    @EnvironmentObject var model: AppModel
    @Environment(\.openWindow) private var openWindow

    // Open into this window when it's empty, otherwise a new window (or focus an
    // already-open project) — mirrors the File ▸ Open Project… command (#69).
    private func openProject() {
        guard let ref = ProjectStore.chooseProject() else { return }
        if model.project == nil {
            model.load(path: ref.path)
        } else {
            openWindow(value: ref)
        }
    }

    var body: some View {
        NavigationSplitView {
            Sidebar()
                .navigationSplitViewColumnWidth(min: 220, ideal: 280)
        } detail: {
            DetailView()
        }
        .toolbar {
            ToolbarItem(placement: .navigation) {
                Button { openProject() } label: { Label("Open", systemImage: "folder") }
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
            } else {
                // Project header pinned above the list so the project name stays
                // visible across every mode and selection (#70).
                VStack(spacing: 0) {
                    ProjectHeader(name: model.project?.name ?? "")
                    Divider()
                    if model.sidebarNodes.isEmpty {
                        emptyMode
                    } else {
                        List(selection: $model.selectedNodeID) {
                            OutlineGroup(model.sidebarNodes, children: \.children) { node in
                                SidebarRow(node: node)
                            }
                        }
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

// Non-selectable project label at the top of the sidebar — the persistent
// answer to "which project am I in" (#70).
struct ProjectHeader: View {
    let name: String

    var body: some View {
        Label {
            Text(name).font(.headline).lineLimit(1).truncationMode(.middle)
        } icon: {
            Image(systemName: "books.vertical.fill").foregroundStyle(.secondary)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .frame(maxWidth: .infinity, alignment: .leading)
        .help(name)
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
            Button("Open Project…") {
                if let ref = ProjectStore.chooseProject() { model.load(path: ref.path) }
            }
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
        // Project name stays in the window title across every navigation state;
        // the subtitle carries the current location (#70). Empty when no project
        // is open, so no project title is shown.
        .navigationTitle(model.project?.name ?? "")
        .navigationSubtitle(model.project == nil ? "" : model.locationTrail())
    }

    @ViewBuilder
    private var detailContent: some View {
        switch model.selection {
        case .artifact(let ref):
            if let change = model.change(named: ref.changeName) {
                ArtifactDetail(change: change, ref: ref)
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
            } else {
                Placeholder()
            }
        case .worktree(let path):
            if let wt = model.worktree(path: path) {
                WorktreeDetail(worktree: wt)
            } else {
                Placeholder()
            }
        case .config:
            if let cfg = model.projectConfig {
                ScrollableContent { MarkdownView(configToMarkdown(cfg)) }
            } else {
                Placeholder()
            }
        case .worktreeArtifact(let wtPath, let name, let kind):
            if let change = model.worktreeChange(worktreePath: wtPath, changeName: name) {
                WorktreeArtifactView(change: change, kind: kind)
                    .id("\(wtPath)#\(name)#\(kind)")
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
    @State private var notice: String?            // file-changed-on-disk notice (#97 conflict)
    @State private var selectedID: String?        // identity of the selected task (add/delete target)
    @State private var editingID: String?          // identity of the task being inline-edited
    @State private var editingText: String = ""
    @State private var pendingDelete: TaskItem?    // task awaiting delete confirmation
    @State private var hoveredID: String?          // row under the pointer (reveals affordances)
    @State private var dropTargetID: String?       // row a drag is currently over (insertion line)
    @FocusState private var editorFocused: Bool    // drives focus + commit-on-blur for inline edit

    private var tasksPath: String { (changePath as NSString).appendingPathComponent("tasks.md") }

    // Stable per-task identity within the view (section prefix + number-stripped
    // description), used for selection, editing, and drag payloads.
    private func id(_ item: TaskItem) -> String { item.sectionPrefix + "\u{1}" + item.taskDescription }

    var body: some View {
        ScrollableContent {
            if let errorText {
                Label(errorText, systemImage: "exclamationmark.triangle.fill")
                    .foregroundStyle(.orange).font(.callout)
            }
            if let notice {
                Label(notice, systemImage: "arrow.clockwise")
                    .foregroundStyle(.secondary).font(.callout)
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
                    editableTaskRow(item)
                    if isLastTaskOfSection(index) {
                        endOfSectionDropZone(prefix: item.sectionPrefix)
                    }
                }
            }
        }
        .onAppear { items = parseTasks(content) }
        .onChange(of: content) { newContent in
            // External reload (incl. FSEvents): re-parse and drop any stale
            // selection/edit state that no longer points at a live task.
            items = parseTasks(newContent)
            let live = Set(items.filter { $0.kind == .task }.map(id))
            if let s = selectedID, !live.contains(s) { selectedID = nil }
            if let e = editingID, !live.contains(e) { editingID = nil }
        }
        .confirmationDialog("Delete this task?",
                            isPresented: Binding(get: { pendingDelete != nil },
                                                 set: { if !$0 { pendingDelete = nil } }),
                            titleVisibility: .visible) {
            Button("Delete", role: .destructive) {
                if let item = pendingDelete { performDelete(item) }
                pendingDelete = nil
            }
            Button("Cancel", role: .cancel) { pendingDelete = nil }
        } message: {
            if let item = pendingDelete { Text(item.taskDescription) }
        }
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
            errorText = nil; notice = nil
        } catch {
            errorText = "Couldn't write tasks.md: \(error.localizedDescription)"
        }
    }

    // ── Editing UI (#97) ──────────────────────────────────────────────────────

    @ViewBuilder
    private func editableTaskRow(_ item: TaskItem) -> some View {
        let rowID = id(item)
        let hovered = hoveredID == rowID
        let editing = editingID == rowID
        // Affordances appear on hover (mouse) or selection (click/keyboard).
        let showControls = (hovered || selectedID == rowID) && !editing
        HStack(alignment: .firstTextBaseline, spacing: 6) {
            // Drag handle — the ONLY draggable element, so clicks on the row's
            // buttons aren't swallowed by the drag gesture. Reserved width keeps
            // the layout stable; it just fades in on hover to signal draggability.
            Image(systemName: "line.3.horizontal")
                .foregroundStyle(hovered ? .secondary : .tertiary)
                .frame(width: 18, height: 20)
                .contentShape(Rectangle())
                .opacity(hovered ? 1 : 0.4)   // faintly visible at rest so it's discoverable + grabbable
                .draggable(rowID)
                .help("Drag to reorder or move to another section")
                .accessibilityLabel("Reorder handle for \(item.taskDescription)")

            // Checkbox toggles.
            Button { toggle(item) } label: {
                Image(systemName: item.done ? "checkmark.square.fill" : "square")
                    .foregroundStyle(item.done ? Color.accentColor : .secondary)
            }
            .buttonStyle(.plain)
            .accessibilityLabel(item.done ? "Completed" : "Not completed")
            .accessibilityHint("Toggles this task")

            if editing {
                // Three-line wrapping editor. Tasks are single-line on disk, so
                // commitEdit collapses any newline to a space; Return adds a line
                // visually, clicking away (focus loss) saves, Esc cancels.
                TextField("Task", text: $editingText, axis: .vertical)
                    .lineLimit(3, reservesSpace: true)
                    .textFieldStyle(.roundedBorder)
                    .focused($editorFocused)
                    .onAppear { editorFocused = true }
                    .onExitCommand { editingID = nil }   // Esc cancels (no save)
                    .onChange(of: editorFocused) { focused in
                        if !focused, editingID == id(item) { commitEdit(item) }
                    }
                // ⌘-Return saves (Return inserts a line in the multi-line box).
                Button("") { commitEdit(item) }
                    .keyboardShortcut(.return, modifiers: .command)
                    .frame(width: 0, height: 0)
                    .opacity(0)
                    .accessibilityHidden(true)
            } else {
                Text(item.text)
                    .strikethrough(item.done, color: .secondary)
                    .foregroundStyle(item.done ? .secondary : .primary)
                    .contentShape(Rectangle())
                    .onTapGesture(count: 2) { beginEdit(item) }
                    .onTapGesture { selectedID = selectedID == rowID ? nil : rowID }
            }

            Spacer(minLength: 8)

            // Always laid out (reserving width + height) and only faded in on
            // hover — so revealing them never reflows text or shifts row height.
            HStack(spacing: 2) {
                rowControl("pencil", "Edit this task", "Edit \(item.taskDescription)") { beginEdit(item) }
                rowControl("plus", "Add a task after this one", "Add task after \(item.taskDescription)") {
                    performAdd(after: item)
                }
                rowControl("minus", "Delete this task", "Delete \(item.taskDescription)") {
                    pendingDelete = item
                }
            }
            .opacity(showControls ? 1 : 0)
            .allowsHitTesting(showControls)
        }
        .padding(.vertical, 2)
        // Fill the row's full width and make the entire strip hit-testable, so
        // hover (and the affordances it reveals) covers the whole row — not just
        // the text/checkbox glyphs.
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(selectedID == rowID ? Color.accentColor.opacity(0.12) : Color.clear)
        // Insertion line: a drag currently hovering this row will drop above it.
        .overlay(alignment: .top) {
            Rectangle().fill(Color.accentColor).frame(height: 2)
                .opacity(dropTargetID == rowID ? 1 : 0)
        }
        .contentShape(Rectangle())
        .onHover { inside in
            if inside { hoveredID = rowID } else if hoveredID == rowID { hoveredID = nil }
        }
        .accessibilityElement(children: .combine)
        .accessibilityValue(item.done ? "completed" : "not completed")
        .dropDestination(for: String.self) { payload, _ in
            dropTargetID = nil
            guard let dropped = payload.first else { return false }
            performMove(draggedID: dropped, onto: item)
            return true
        } isTargeted: { targeted in
            if targeted { dropTargetID = rowID } else if dropTargetID == rowID { dropTargetID = nil }
        }
    }

    // A small hover-revealed row affordance with a comfortable hit target.
    private func rowControl(_ symbol: String, _ help: String, _ label: String,
                            action: @escaping () -> Void) -> some View {
        Button(action: action) {
            Image(systemName: symbol)
                .foregroundStyle(.primary)        // clearly visible (the row already gates on hover)
                .frame(width: 26, height: 22)
                .contentShape(Rectangle())        // full frame is clickable, not just the glyph
        }
        .buttonStyle(.borderless)
        .help(help)
        .accessibilityLabel(label)
    }

    // True when `index` is a task and the next item is a section or the end of
    // the list — i.e. the last task of its section. (items holds only sections
    // and tasks.)
    private func isLastTaskOfSection(_ index: Int) -> Bool {
        guard items[index].kind == .task else { return false }
        let next = index + 1
        return next >= items.count || items[next].kind == .section
    }

    // A drop target at the bottom of a section so a task can be moved to the very
    // end (the per-row targets only insert *above* a row). Subtle until a drag is
    // over it, when it shows the same insertion line.
    @ViewBuilder
    private func endOfSectionDropZone(prefix: String) -> some View {
        let zoneID = "\u{2}END\u{1}" + prefix
        Color.clear
            .frame(maxWidth: .infinity)
            .frame(height: 12)
            .overlay(alignment: .top) {
                Rectangle().fill(Color.accentColor).frame(height: 2)
                    .opacity(dropTargetID == zoneID ? 1 : 0)
            }
            .contentShape(Rectangle())
            .dropDestination(for: String.self) { payload, _ in
                dropTargetID = nil
                guard let dropped = payload.first else { return false }
                performMoveToEnd(draggedID: dropped, sectionPrefix: prefix)
                return true
            } isTargeted: { targeted in
                if targeted { dropTargetID = zoneID } else if dropTargetID == zoneID { dropTargetID = nil }
            }
            .accessibilityHidden(true)
    }

    private func beginEdit(_ item: TaskItem) {
        editingID = id(item)
        editingText = item.taskDescription
        selectedID = id(item)
    }

    private func commitEdit(_ item: TaskItem) {
        // Collapse any visual newlines to spaces — tasks.md tasks are single-line.
        let newText = trimSpaceLocal(editingText
            .replacingOccurrences(of: "\r", with: " ")
            .replacingOccurrences(of: "\n", with: " "))
        editingID = nil
        guard !newText.isEmpty, newText != item.taskDescription else { return }
        run { try editTaskText(tasksPath, identity: item.taskDescription,
                               inSection: item.sectionPrefix, newDescription: newText) }
    }

    private func performAdd(after item: TaskItem) {
        run { try addTask(tasksPath, afterIdentity: item.taskDescription,
                          inSection: item.sectionPrefix, description: "New task") }
        // Select + edit the freshly added task for immediate typing.
        let newItem = TaskItem(kind: .task, text: "", done: false, lineNum: 0,
                               sectionPrefix: item.sectionPrefix, ordinal: item.ordinal + 1,
                               taskDescription: "New task")
        beginEdit(newItem)
    }

    private func performDelete(_ item: TaskItem) {
        run { try deleteTask(tasksPath, identity: item.taskDescription, inSection: item.sectionPrefix) }
        if selectedID == id(item) { selectedID = nil }
    }

    private func performMove(draggedID: String, onto target: TaskItem) {
        let parts = draggedID.split(separator: "\u{1}", maxSplits: 1, omittingEmptySubsequences: false)
        guard parts.count == 2 else { return }
        let fromPrefix = String(parts[0]), draggedDesc = String(parts[1])
        guard !(fromPrefix == target.sectionPrefix && draggedDesc == target.taskDescription) else { return }
        run { try moveTask(tasksPath, identity: draggedDesc, fromSection: fromPrefix,
                           toSection: target.sectionPrefix, toIndex: max(0, target.ordinal - 1)) }
    }

    // Drop onto a section's end zone: append the dragged task to that section.
    private func performMoveToEnd(draggedID: String, sectionPrefix prefix: String) {
        let parts = draggedID.split(separator: "\u{1}", maxSplits: 1, omittingEmptySubsequences: false)
        guard parts.count == 2 else { return }
        let fromPrefix = String(parts[0]), draggedDesc = String(parts[1])
        run { try moveTask(tasksPath, identity: draggedDesc, fromSection: fromPrefix,
                           toSection: prefix, toIndex: Int.max) }
    }

    // Runs an edit op, mapping a conflict to a visible notice + disk refresh.
    private func run(_ op: () throws -> [TaskItem]) {
        do {
            items = try op(); errorText = nil; notice = nil
        } catch TaskEditError.fileChanged {
            notice = "tasks.md changed on disk — refreshed."
            refreshFromDisk()
        } catch {
            errorText = "Couldn't write tasks.md: \(error.localizedDescription)"
        }
    }

    private func refreshFromDisk() {
        if let data = try? Data(contentsOf: URL(fileURLWithPath: tasksPath)) {
            items = parseTasks(String(decoding: data, as: UTF8.self))
        }
    }

    private func trimSpaceLocal(_ s: String) -> String {
        s.trimmingCharacters(in: .whitespacesAndNewlines)
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
