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
            VStack(alignment: .leading, spacing: 1) {
                Text(node.title)
                    .font(node.prominent ? .headline : .body)
                    .lineLimit(1)
                if let subtitle = node.subtitle {
                    Text(subtitle).font(.caption).foregroundStyle(.secondary)
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
        if case .specFile(let name) = ref.kind {
            SpecContentView(artifact: artifact, issues: artifact.readError ? [] : validateChange(change))
                .id("\(change.name)/spec/\(name)")
        } else {
            ScrollableContent { ArtifactBody(artifact: artifact) }
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
