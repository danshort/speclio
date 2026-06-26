import Foundation
import OpenSpecKit

// OpenSpecKitGolden runs every domain entry point against the shared
// testdata/corpus/ and byte-compares to the committed golden files produced by
// the Go golden test. Shared by the `oskgolden` executable (CLT-runnable) and
// the XCTest target. See ../../../../testdata/corpus/README.md for the contract.

public struct GoldenFailure: Error, CustomStringConvertible {
    public let name: String
    public let detail: String
    public var description: String { "\(name): \(detail)" }
}

// ── corpus location (relative to this source file) ────────────────────────────

public func corpusDir(file: StaticString = #filePath) -> String {
    // Walk up from this file until a dir containing testdata/corpus is found.
    var dir = URL(fileURLWithPath: "\(file)").deletingLastPathComponent()
    for _ in 0..<8 {
        let candidate = dir.appendingPathComponent("testdata/corpus")
        if FileManager.default.fileExists(atPath: candidate.path) {
            return candidate.path
        }
        dir = dir.deletingLastPathComponent()
    }
    fatalError("could not locate testdata/corpus from \(file)")
}

// ── canonical JSON: must match Go's encoder byte-for-byte ─────────────────────
// Go uses json.Encoder with SetIndent("", "  ") + SetEscapeHTML(false): 2-space
// indent, `": "` separators, inline `[]`/`{}` for empties, sorted keys, slashes
// and most unicode unescaped, trailing "\n". Swift's JSONEncoder.prettyPrinted
// differs (it emits `" : "` and expands empty collections), so we decode into a
// neutral tree and emit with a printer that matches Go exactly.

func canonicalJSON<T: Encodable>(_ value: T) throws -> Data {
    let compact = try JSONEncoder().encode(value) // intermediate; escaping here is irrelevant
    let tree = try JSONDecoder().decode(JSONValue.self, from: compact)
    return Data((emitCanonical(tree, indent: 0) + "\n").utf8)
}

// A neutral JSON tree that preserves value types (avoids NSNumber bool/int
// ambiguity from JSONSerialization). Keys are re-sorted at emit time.
indirect enum JSONValue: Decodable {
    case object([String: JSONValue])
    case array([JSONValue])
    case string(String)
    case int(Int)
    case double(Double)
    case bool(Bool)
    case null

    private struct AnyKey: CodingKey {
        var stringValue: String
        var intValue: Int?
        init?(stringValue s: String) { stringValue = s; intValue = nil }
        init?(intValue: Int) { return nil }
    }

    init(from decoder: Decoder) throws {
        let single = try? decoder.singleValueContainer()
        if let c = single, c.decodeNil() { self = .null; return }
        if let c = single, let b = try? c.decode(Bool.self) { self = .bool(b); return }
        if let c = single, let i = try? c.decode(Int.self) { self = .int(i); return }
        if let c = single, let d = try? c.decode(Double.self) { self = .double(d); return }
        if let c = single, let s = try? c.decode(String.self) { self = .string(s); return }
        if let kc = try? decoder.container(keyedBy: AnyKey.self) {
            var obj: [String: JSONValue] = [:]
            for k in kc.allKeys { obj[k.stringValue] = try kc.decode(JSONValue.self, forKey: k) }
            self = .object(obj)
            return
        }
        if var uc = try? decoder.unkeyedContainer() {
            var arr: [JSONValue] = []
            while !uc.isAtEnd { arr.append(try uc.decode(JSONValue.self)) }
            self = .array(arr)
            return
        }
        throw DecodingError.dataCorrupted(.init(codingPath: decoder.codingPath, debugDescription: "unknown JSON value"))
    }
}

private func emitCanonical(_ v: JSONValue, indent: Int) -> String {
    let pad = String(repeating: "  ", count: indent)
    let pad1 = String(repeating: "  ", count: indent + 1)
    switch v {
    case .null: return "null"
    case .bool(let b): return b ? "true" : "false"
    case .int(let i): return String(i)
    case .double(let d): return d == d.rounded() ? String(Int(d)) : String(d)
    case .string(let s): return "\"" + escapeJSON(s) + "\""
    case .array(let a):
        if a.isEmpty { return "[]" }
        return "[\n" + a.map { pad1 + emitCanonical($0, indent: indent + 1) }.joined(separator: ",\n") + "\n" + pad + "]"
    case .object(let o):
        if o.isEmpty { return "{}" }
        let items = o.keys.sorted().map { key in
            pad1 + "\"" + escapeJSON(key) + "\": " + emitCanonical(o[key]!, indent: indent + 1)
        }
        return "{\n" + items.joined(separator: ",\n") + "\n" + pad + "}"
    }
}

// Mirrors Go's string escaping with SetEscapeHTML(false): escapes ", \\, the
// short control escapes, U+2028/U+2029, and other control chars as \u00xx;
// leaves /, <, >, & and all other unicode literal.
private func escapeJSON(_ s: String) -> String {
    var out = ""
    out.reserveCapacity(s.count + 2)
    for scalar in s.unicodeScalars {
        switch scalar {
        case "\"": out += "\\\""
        case "\\": out += "\\\\"
        case "\n": out += "\\n"
        case "\r": out += "\\r"
        case "\t": out += "\\t"
        case "\u{2028}": out += "\\u2028"
        case "\u{2029}": out += "\\u2029"
        default:
            if scalar.value < 0x20 {
                out += String(format: "\\u%04x", scalar.value)
            } else {
                out.unicodeScalars.append(scalar)
            }
        }
    }
    return out
}

// ── golden DTOs (machine- and language-portable view) ─────────────────────────

struct GArtifact: Encodable {
    let content: String
    let present: Bool
    let readError: Bool
    enum CodingKeys: String, CodingKey { case content, present; case readError = "read_error" }
}

struct GNamedSpec: Encodable {
    let name: String
    let content: String
    let readError: Bool
    enum CodingKeys: String, CodingKey { case name, content; case readError = "read_error" }
}

struct GChange: Encodable {
    let name: String
    let path: String
    let created: String
    let displayDate: String
    let proposal: GArtifact
    let design: GArtifact
    let tasks: GArtifact
    let specs: GArtifact
    let specFiles: [GNamedSpec]
    enum CodingKeys: String, CodingKey {
        case name, path, created, proposal, design, tasks, specs
        case displayDate = "display_date"
        case specFiles = "spec_files"
    }
}

struct GProject: Encodable {
    let name: String
    let changes: [GChange]
}

func relPath(_ abs: String, _ root: String) -> String {
    if abs == root { return "." }
    if abs.hasPrefix(root + "/") { return String(abs.dropFirst(root.count + 1)) }
    return abs
}

func normContent(_ content: String, readError: Bool, root: String) -> String {
    if !readError { return content }
    var rest = content
    if rest.hasPrefix(Layout_unreadablePrefix) {
        rest = String(rest.dropFirst(Layout_unreadablePrefix.count))
    }
    if let r = rest.range(of: ": ") {
        rest = String(rest[rest.startIndex..<r.lowerBound])
    }
    return Layout_unreadablePrefix + relPath(rest, root)
}

// The unreadable-artifact placeholder prefix (Layout is internal to OpenSpecKit).
let Layout_unreadablePrefix = "⚠ couldn't read "

func normArtifact(_ a: Artifact, _ root: String) -> GArtifact {
    GArtifact(content: normContent(a.content, readError: a.readError, root: root),
              present: a.present, readError: a.readError)
}

func normChange(_ c: Change, _ root: String) -> GChange {
    GChange(
        name: c.name,
        path: relPath(c.path, root),
        created: c.created,
        displayDate: c.displayDate,
        proposal: normArtifact(c.proposal, root),
        design: normArtifact(c.design, root),
        tasks: normArtifact(c.tasks, root),
        specs: normArtifact(c.specs, root),
        specFiles: c.specFiles.map {
            GNamedSpec(name: $0.name, content: normContent($0.content, readError: $0.readError, root: root), readError: $0.readError)
        }
    )
}

func normProject(_ p: Project, _ root: String) -> GProject {
    GProject(name: p.name, changes: p.changes.map { normChange($0, root) })
}

struct GProjectSpec: Encodable {
    let name: String
    let requirementCount: Int
    let requirementNames: [String]
    let content: String
    let readError: Bool
    enum CodingKeys: String, CodingKey {
        case name, content
        case requirementCount = "requirement_count"
        case requirementNames = "requirement_names"
        case readError = "read_error"
    }
}

func normProjectSpecs(_ specs: [ProjectSpec], _ root: String) -> [GProjectSpec] {
    specs.map {
        GProjectSpec(name: $0.name, requirementCount: $0.requirementCount,
                     requirementNames: $0.requirementNames,
                     content: normContent($0.content, readError: $0.readError, root: root),
                     readError: $0.readError)
    }
}

// ── fault injection (deterministic present-but-unreadable) ────────────────────

struct FaultFS: FileSystem {
    let base = OSFileSystem()
    let failPath: String
    func readFile(_ path: String) throws -> Data {
        if path == failPath {
            throw NSError(domain: "OpenSpecKitGolden", code: 1,
                          userInfo: [NSLocalizedDescriptionKey: "permission denied"])
        }
        return try base.readFile(path)
    }
    func writeFile(_ path: String, _ data: Data) throws { try base.writeFile(path, data) }
    func readDir(_ path: String) throws -> [DirEntry] { try base.readDir(path) }
    func stat(_ path: String) throws -> FileInfo { try base.stat(path) }
}

// ── the checks ────────────────────────────────────────────────────────────────

public func runAllGoldenChecks(corpus: String? = nil) -> [GoldenFailure] {
    let root = corpus ?? corpusDir()
    var failures: [GoldenFailure] = []

    func compare(_ name: String, _ actual: Data) {
        let goldenPath = (root as NSString).appendingPathComponent("golden/\(name)")
        guard let want = try? Data(contentsOf: URL(fileURLWithPath: goldenPath)) else {
            failures.append(GoldenFailure(name: name, detail: "missing golden \(goldenPath)"))
            return
        }
        if want != actual {
            failures.append(GoldenFailure(name: name, detail: diff(want: want, got: actual)))
        }
    }
    func compareJSON<T: Encodable>(_ name: String, _ value: T) {
        do { compare(name, try canonicalJSON(value)) }
        catch { failures.append(GoldenFailure(name: name, detail: "encode error: \(error)")) }
    }
    func fixture(_ rel: String) -> String { (root as NSString).appendingPathComponent(rel) }
    func read(_ rel: String) -> Data { (try? Data(contentsOf: URL(fileURLWithPath: fixture(rel)))) ?? Data() }

    // project / basic-project
    if let p = try? Loader().loadFrom(fixture("basic-project")) {
        compareJSON("basic-project.project.json", normProject(p, fixture("basic-project")))
    } else { failures.append(GoldenFailure(name: "basic-project.project.json", detail: "loadFrom threw")) }

    // project / malformed-meta
    if let p = try? Loader().loadFrom(fixture("malformed-meta")) {
        compareJSON("malformed-meta.project.json", normProject(p, fixture("malformed-meta")))
    } else { failures.append(GoldenFailure(name: "malformed-meta.project.json", detail: "loadFrom threw")) }

    // project / unreadable-artifact (inject a fault on proposal.md)
    do {
        let r = fixture("unreadable-artifact")
        let failPath = joinPathG(r, "openspec", "changes", "has-unreadable", "proposal.md")
        let p = try Loader(fs: FaultFS(failPath: failPath)).loadFrom(r)
        compareJSON("unreadable-artifact.project.json", normProject(p, r))
    } catch { failures.append(GoldenFailure(name: "unreadable-artifact.project.json", detail: "\(error)")) }

    // archive / malformed-archive-name
    if let changes = try? Loader().listArchiveChangesFrom(fixture("malformed-archive-name")) {
        compareJSON("malformed-archive-name.archive.json", changes.map { normChange($0, fixture("malformed-archive-name")) })
    } else { failures.append(GoldenFailure(name: "malformed-archive-name.archive.json", detail: "listArchive threw")) }

    // tasks/parse (LF) + CRLF parity
    let lfItems = parseTasks(String(decoding: read("lf-tasks/tasks.md"), as: UTF8.self))
    let crlfItems = parseTasks(String(decoding: read("crlf-tasks/tasks.md"), as: UTF8.self))
    compareJSON("tasks.json", lfItems)
    if (try? canonicalJSON(lfItems)) != (try? canonicalJSON(crlfItems)) {
        failures.append(GoldenFailure(name: "tasks.json", detail: "CRLF and LF parsed differently"))
    }

    // project-specs/basic-project
    if let specs = try? Loader().loadProjectSpecsFrom(fixture("basic-project")) {
        compareJSON("basic-project.project-specs.json", normProjectSpecs(specs, fixture("basic-project")))
    } else {
        failures.append(GoldenFailure(name: "basic-project.project-specs.json", detail: "loadProjectSpecsFrom threw"))
    }

    // requirements/extract
    let authSpec = String(decoding: read("basic-project/openspec/changes/alpha-change/specs/auth/spec.md"), as: UTF8.self)
    compareJSON("requirements.json", [
        "Login": extractRequirement(authSpec, "Login"),
        "Logout": extractRequirement(authSpec, "Logout"),
        "Missing": extractRequirement(authSpec, "Missing"),
    ])

    // config/variants
    var configs: [String: ProjectConfig] = [:]
    for v in ["absent-rules", "empty-rules", "multiline"] {
        if let cfg = try? Loader().loadConfigFrom(fixture("config-variants/\(v)")) {
            configs[v] = cfg
            compare("config-\(v).md", Data(configToMarkdown(cfg).utf8))
        } else {
            failures.append(GoldenFailure(name: "config-\(v).md", detail: "loadConfig threw"))
        }
    }
    compareJSON("config.json", configs)

    // worktrees/parse
    for f in ["three-worktrees", "locked-prunable"] {
        compareJSON("worktrees-\(f).json", parseWorktreeList(read("worktree-porcelain/\(f).txt")))
    }

    // validation
    var validation: [String: [String]] = [:]
    for name in ["valid-spec", "missing-sections", "req-without-scenario"] {
        let content = String(decoding: read("delta-specs/specs/\(name)/spec.md"), as: UTF8.self)
        validation["spec/\(name)"] = validateSpec(content)
    }
    if let ch = try? Loader().loadFromPath(fixture("delta-specs/changes/needs-work")) {
        validation["change/needs-work"] = validateChange(ch.changes[0])
    } else {
        // needs-work has no .openspec.yaml; load it directly as a change dir.
        validation["change/needs-work"] = validateChange(loadChangeDirect(fixture("delta-specs/changes/needs-work")))
    }
    compareJSON("validation.json", validation)

    return failures
}

// loadChangeDirect builds a Change from a dir without requiring .openspec.yaml
// (mirrors the Go test's loadChangeFromDir call for the validation fixture).
func loadChangeDirect(_ dir: String) -> Change {
    var ch = Change(name: (dir as NSString).lastPathComponent, path: dir)
    ch.proposal = artifactAt(joinPathG(dir, "proposal.md"))
    let specsDir = joinPathG(dir, "specs")
    if let names = try? FileManager.default.contentsOfDirectory(atPath: specsDir).sorted() {
        for n in names {
            let p = joinPathG(specsDir, n, "spec.md")
            if let data = try? Data(contentsOf: URL(fileURLWithPath: p)) {
                ch.specFiles.append(NamedSpec(name: n, content: String(decoding: data, as: UTF8.self)))
            }
        }
    }
    return ch
}

private func artifactAt(_ path: String) -> Artifact {
    if let data = try? Data(contentsOf: URL(fileURLWithPath: path)) {
        return Artifact(content: String(decoding: data, as: UTF8.self), present: true)
    }
    return Artifact()
}

func joinPathG(_ parts: String...) -> String {
    guard var r = parts.first else { return "" }
    for p in parts.dropFirst() { r = (r as NSString).appendingPathComponent(p) }
    return r
}

private func diff(want: Data, got: Data) -> String {
    let w = String(decoding: want, as: UTF8.self)
    let g = String(decoding: got, as: UTF8.self)
    return "byte mismatch\n--- want (\(want.count)B) ---\n\(w)\n--- got (\(got.count)B) ---\n\(g)"
}
