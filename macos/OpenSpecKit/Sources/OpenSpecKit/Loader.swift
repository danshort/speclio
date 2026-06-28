import Foundation
import Yams

// Loader mirrors Go internal/openspec/loader.go. It is constructed with a
// FileSystem (default OSFileSystem) so tests can inject faults.

public enum LoaderError: Error, Equatable {
    case noOpenspecDir(String)
    case notFound(String)
    case notAValidChangeDir(String)
}

public struct Loader {
    let fs: FileSystem

    public init(fs: FileSystem = OSFileSystem()) {
        self.fs = fs
    }

    // ── *From(root) variants ──────────────────────────────────────────────────

    public func loadFrom(_ root: String) throws -> Project {
        let openspecDir = joinPath(root, Layout.dirOpenspec)
        do {
            _ = try fs.stat(openspecDir)
        } catch FSError.notFound {
            throw LoaderError.noOpenspecDir(root)
        }

        var project = Project(name: baseName(root))
        let changesDir = joinPath(openspecDir, Layout.dirChanges)
        for name in try listDirs(changesDir, excludeArchive: true) {
            project.changes.append(loadChangeFromDir(joinPath(changesDir, name), name: name, displayDate: ""))
        }
        project.changes = stableSortedChanges(project.changes)
        return project
    }

    public func loadConfigFrom(_ root: String) throws -> ProjectConfig {
        let path = joinPath(root, Layout.dirOpenspec, Layout.fileConfig)
        let data: Data
        do {
            data = try fs.readFile(path)
        } catch FSError.notFound {
            return ProjectConfig()
        }
        // config.yaml PROPAGATES unmarshal errors (unlike .openspec.yaml).
        let raw = try YAMLDecoder().decode(ConfigYAML.self, from: String(decoding: data, as: UTF8.self))
        return ProjectConfig(context: trimSpace(raw.context ?? ""), rules: raw.rules)
    }

    public func loadProjectSpecsFrom(_ root: String) throws -> [ProjectSpec] {
        let specsDir = joinPath(root, Layout.dirOpenspec, Layout.dirSpecs)
        let names = try listDirs(specsDir, excludeArchive: false) // propagates errors (unlike loadSpecs)
        var specs: [ProjectSpec] = []
        for name in names {
            var ps = ProjectSpec(name: name)
            let specPath = joinPath(specsDir, name, Layout.fileSpec)
            do {
                let content = String(decoding: try fs.readFile(specPath), as: UTF8.self)
                ps.content = content
                for line in splitLines(content) where line.hasPrefix(Layout.reqPrefix) {
                    ps.requirementCount += 1
                    ps.requirementNames.append(trimSpace(String(line.dropFirst(Layout.reqPrefix.count))))
                }
            } catch FSError.notFound {
                // spec dir without a spec.md — surface it with a ⚠ rather than
                // listing it as silently empty (#96).
                ps.readError = true
                ps.content = Layout.missingSpecPrefix + name
            } catch {
                ps.readError = true
                ps.content = Layout.unreadablePrefix + specPath + ": " + error.localizedDescription
            }
            specs.append(ps)
        }
        specs.sort { $0.name < $1.name }
        return specs
    }

    public func listArchiveChangesFrom(_ root: String) throws -> [Change] {
        let archiveDir = joinPath(root, Layout.dirOpenspec, Layout.dirChanges, Layout.dirArchive)
        var dirs = try listDirs(archiveDir, excludeArchive: false)
        dirs.sort(by: >) // reverse, matching Go sort.Reverse(StringSlice)
        var changes: [Change] = []
        for dir in dirs {
            let (clean, disp) = parseArchiveName(dir)
            changes.append(loadChangeFromDir(joinPath(archiveDir, dir), name: clean, displayDate: disp))
        }
        return changes
    }

    // ── Path-based loader ──────────────────────────────────────────────────────

    public func loadFromPath(_ path: String) throws -> Project {
        do {
            _ = try fs.stat(path)
        } catch FSError.notFound {
            throw LoaderError.notFound(path)
        }
        do {
            _ = try fs.stat(joinPath(path, Layout.fileMeta))
        } catch FSError.notFound {
            throw LoaderError.notAValidChangeDir(path)
        }
        let ch = loadChangeFromDir(path, name: baseName(path), displayDate: "")
        return Project(name: baseName(dirName(path)), changes: [ch])
    }

    public func reloadChange(_ ch: Change) -> Change {
        var out = ch
        out.proposal = loadFile(joinPath(ch.path, Layout.fileProposal))
        out.design = loadFile(joinPath(ch.path, Layout.fileDesign))
        out.tasks = loadFile(joinPath(ch.path, Layout.fileTasks))
        let (specs, files) = loadSpecs(joinPath(ch.path, Layout.dirSpecs))
        out.specs = specs
        out.specFiles = files
        return out
    }

    // ── helpers ──────────────────────────────────────────────────────────────

    func loadChangeFromDir(_ dir: String, name: String, displayDate: String) -> Change {
        var ch = Change(name: name, path: dir, displayDate: displayDate)
        if let raw = try? fs.readFile(joinPath(dir, Layout.fileMeta)) {
            ch.created = parseMetaCreated(raw)
        }
        ch.proposal = loadFile(joinPath(dir, Layout.fileProposal))
        ch.design = loadFile(joinPath(dir, Layout.fileDesign))
        ch.tasks = loadFile(joinPath(dir, Layout.fileTasks))
        let (specs, files) = loadSpecs(joinPath(dir, Layout.dirSpecs))
        ch.specs = specs
        ch.specFiles = files
        return ch
    }

    func loadFile(_ path: String) -> Artifact {
        do {
            let data = try fs.readFile(path)
            return Artifact(content: String(decoding: data, as: UTF8.self), present: true)
        } catch FSError.notFound {
            return Artifact() // absent
        } catch {
            // Exists but unreadable: surface it rather than vanishing.
            return Artifact(content: Layout.unreadablePrefix + path + ": " + error.localizedDescription,
                            present: true, readError: true)
        }
    }

    func loadSpecs(_ dir: String) -> (Artifact, [NamedSpec]) {
        let names: [String]
        do {
            names = try listDirs(dir, excludeArchive: false)
        } catch {
            return (Artifact(), [])
        }
        var parts: [String] = []
        var files: [NamedSpec] = []
        for name in names {
            let specPath = joinPath(dir, name, Layout.fileSpec)
            do {
                let content = String(decoding: try fs.readFile(specPath), as: UTF8.self)
                files.append(NamedSpec(name: name, content: content))
                parts.append("# " + name + "\n\n" + content)
            } catch FSError.notFound {
                // spec dir without a spec.md — surface it with a ⚠ rather than
                // dropping it silently (#96). The placeholder names the
                // capability (not a path) so combined content stays portable.
                let ph = Layout.missingSpecPrefix + name
                files.append(NamedSpec(name: name, content: ph, readError: true))
                parts.append("# " + name + "\n\n" + ph)
            } catch {
                let ph = Layout.unreadablePrefix + specPath + ": " + error.localizedDescription
                files.append(NamedSpec(name: name, content: ph, readError: true))
                parts.append("# " + name + "\n\n" + ph)
            }
        }
        if parts.isEmpty { return (Artifact(), []) }
        return (Artifact(content: parts.joined(separator: "\n\n---\n\n"), present: true), files)
    }

    func listDirs(_ path: String, excludeArchive: Bool) throws -> [String] {
        let entries: [DirEntry]
        do {
            entries = try fs.readDir(path)
        } catch FSError.notFound {
            return []
        }
        return entries
            .filter { $0.isDir && !(excludeArchive && $0.name == Layout.dirArchive) }
            .map { $0.name }
    }

    private func parseMetaCreated(_ data: Data) -> String {
        // .openspec.yaml is optional metadata: swallow all parse errors.
        guard let meta = try? YAMLDecoder().decode(Meta.self, from: String(decoding: data, as: UTF8.self)) else {
            return ""
        }
        return meta.created ?? ""
    }
}

// stableSortedChanges replicates Go's sort.SliceStable: Created descending,
// empty Created last, ties by Name ascending ONLY when both Created are empty;
// equal non-empty Created keep input order (stability via an index tiebreak).
func stableSortedChanges(_ changes: [Change]) -> [Change] {
    changes.enumerated().sorted { lhs, rhs in
        let a = lhs.element.created, b = rhs.element.created
        if a.isEmpty && b.isEmpty {
            if lhs.element.name != rhs.element.name { return lhs.element.name < rhs.element.name }
            return lhs.offset < rhs.offset
        }
        if a.isEmpty { return false }
        if b.isEmpty { return true }
        if a != b { return a > b }
        return lhs.offset < rhs.offset
    }.map { $0.element }
}

private let archiveRx = Rx("^([0-9]{4}-[0-9]{2}-[0-9]{2})-(.+)$")

/// parseArchiveName mirrors Go: the dir must match the date-prefix regex AND the
/// prefix must be a real calendar date, else (dir, "").
func parseArchiveName(_ dir: String) -> (name: String, date: String) {
    let range = NSRange(dir.startIndex..<dir.endIndex, in: dir)
    guard let m = archiveRx.re.firstMatch(in: dir, range: range),
          let dr = Range(m.range(at: 1), in: dir),
          let nr = Range(m.range(at: 2), in: dir) else {
        return (dir, "")
    }
    let datePart = String(dir[dr])
    if !isValidISODate(datePart) { return (dir, "") }
    return (String(dir[nr]), datePart)
}

private func isValidISODate(_ s: String) -> Bool {
    let parts = s.split(separator: "-")
    guard parts.count == 3, let y = Int(parts[0]), let mo = Int(parts[1]), let d = Int(parts[2]) else {
        return false
    }
    var cal = Calendar(identifier: .gregorian)
    cal.timeZone = TimeZone(identifier: "UTC")!
    return DateComponents(calendar: cal, year: y, month: mo, day: d).isValidDate
}

/// extractRequirement mirrors Go: the block from the matching `### Requirement:`
/// (exact trimmed name) to the next `### Requirement:`.
public func extractRequirement(_ raw: String, _ name: String) -> String {
    let lines = splitLines(raw)
    var start = -1
    for (i, l) in lines.enumerated()
    where l.hasPrefix(Layout.reqPrefix) && trimSpace(String(l.dropFirst(Layout.reqPrefix.count))) == name {
        start = i
        break
    }
    if start < 0 { return "" }
    var block = [lines[start]]
    for l in lines[(start + 1)...] {
        if l.hasPrefix(Layout.reqPrefix) { break }
        block.append(l)
    }
    return block.joined(separator: "\n")
}

public func configToMarkdown(_ cfg: ProjectConfig) -> String {
    var sb = ""
    if !cfg.context.isEmpty {
        sb += "## Context\n\n"
        sb += cfg.context
        sb += "\n"
    }
    if let rules = cfg.rules, !rules.isEmpty {
        if !cfg.context.isEmpty { sb += "\n" }
        sb += "## Rules\n"
        for k in rules.keys.sorted() {
            sb += "\n### "
            sb += k
            sb += "\n\n"
            for item in rules[k]! {
                sb += "- "
                sb += item
                sb += "\n"
            }
        }
    }
    return sb
}

private struct Meta: Decodable {
    var created: String?
}

private struct ConfigYAML: Decodable {
    var context: String?
    var rules: [String: [String]]?
}
