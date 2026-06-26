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
            }
            ToolbarItem {
                Button { model.reload() } label: { Label("Reload", systemImage: "arrow.clockwise") }
                    .disabled(model.project == nil)
            }
        }
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
                List(selection: $model.selection) {
                    switch model.mode {
                    case .activeChanges, .archivedChanges:
                        changeList(model.changes(for: model.mode))
                    case .specs:
                        specList
                    case .worktrees:
                        worktreeList
                    }
                }
            }
        }
        .frame(minWidth: 220)
    }

    @ViewBuilder
    private func changeList(_ changes: [Change]) -> some View {
        if changes.isEmpty {
            Text("Nothing here").foregroundStyle(.secondary)
        } else {
            ForEach(changes, id: \.path) { change in
                DisclosureGroup(isExpanded: model.changeExpansion(change)) {
                    ArtifactRows(change: change)
                } label: {
                    ChangeLabel(change: change)
                }
            }
        }
    }

    @ViewBuilder
    private var specList: some View {
        if model.projectSpecs.isEmpty {
            Text("No project specs").foregroundStyle(.secondary)
        } else {
            ForEach(model.projectSpecs, id: \.name) { spec in
                Label(spec.name, systemImage: "doc.plaintext")
                    .badge(spec.requirementCount)
                    .tag(Selection.projectSpec(spec.name))
            }
        }
    }

    @ViewBuilder
    private var worktreeList: some View {
        if model.worktrees.isEmpty {
            Text(model.worktreesError ?? "No worktrees").foregroundStyle(.secondary)
        } else {
            ForEach(model.worktrees, id: \.path) { wt in
                Label(worktreeTitle(wt), systemImage: wt.isCurrent ? "checkmark.circle.fill" : "point.3.connected.trianglepath.dotted")
                    .tag(Selection.worktree(wt.path))
            }
        }
    }

    private func worktreeTitle(_ wt: Worktree) -> String {
        if wt.bare { return "(bare)" }
        if wt.detached { return "(detached)" }
        return wt.branch.isEmpty ? (wt.path as NSString).lastPathComponent : wt.branch
    }
}

struct ChangeLabel: View {
    let change: Change

    var body: some View {
        VStack(alignment: .leading, spacing: 1) {
            Text(change.name)
                .font(.headline)
                .lineLimit(1)
            if !change.displayDate.isEmpty {
                Text(change.displayDate)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
        .help(change.name)
    }
}

struct ArtifactRows: View {
    let change: Change

    var body: some View {
        if change.proposal.present {
            row(.proposal, "Proposal", "doc.text")
        }
        if change.design.present {
            row(.design, "Design", "pencil.and.outline")
        }
        if !change.specFiles.isEmpty {
            Label("Specs", systemImage: "folder")
                .font(.subheadline.weight(.semibold))
                .foregroundStyle(.secondary)
            ForEach(change.specFiles, id: \.name) { sf in
                Label(sf.name, systemImage: "doc.plaintext")
                    .tag(Selection.artifact(ArtifactRef(changeName: change.name, kind: .specFile(sf.name))))
                    .padding(.leading, 14)
            }
        }
        if change.tasks.present {
            row(.tasks, "Tasks", "checklist")
        }
    }

    private func row(_ kind: ArtifactKind, _ title: String, _ icon: String) -> some View {
        Label(title, systemImage: icon)
            .tag(Selection.artifact(ArtifactRef(changeName: change.name, kind: kind)))
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
                ProjectSpecDetail(spec: spec).navigationTitle(name)
            } else {
                Placeholder()
            }
        case .worktree(let path):
            if let wt = model.worktree(path: path) {
                WorktreeDetail(worktree: wt).navigationTitle("Worktrees")
            } else {
                Placeholder()
            }
        case .none:
            Placeholder()
        }
    }
}

struct ArtifactDetail: View {
    @EnvironmentObject var model: AppModel
    let change: Change
    let ref: ArtifactRef

    var body: some View {
        let artifact = model.artifact(for: ref, in: change)
        ScrollableContent {
            if isSpec(ref.kind), !artifact.readError {
                ValidationBanner(issues: validateChange(change))
            }
            ArtifactBody(artifact: artifact)
        }
    }

    private func isSpec(_ kind: ArtifactKind) -> Bool {
        if case .specFile = kind { return true }
        return false
    }
}

struct ProjectSpecDetail: View {
    let spec: ProjectSpec

    var body: some View {
        ScrollableContent {
            if !spec.readError {
                ValidationBanner(issues: validateSpec(spec.content))
            }
            ArtifactBody(artifact: Artifact(content: spec.content, present: true, readError: spec.readError))
        }
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
