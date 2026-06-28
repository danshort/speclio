import Foundation

// Models.swift mirrors the Go domain types in internal/openspec. CodingKeys
// document the cross-language serialization contract (snake_case field names);
// see ../../../testdata/corpus/README.md. The golden harness encodes normalized
// DTOs (read errors as a bool, paths relativized) — these domain types carry the
// in-memory shape the app consumes.

public struct Artifact: Codable, Equatable {
    public var content: String
    public var present: Bool
    /// True when the file exists but could not be read (a non-not-found error).
    /// The artifact is still `present`, with placeholder `content`.
    public var readError: Bool

    public init(content: String = "", present: Bool = false, readError: Bool = false) {
        self.content = content
        self.present = present
        self.readError = readError
    }

    enum CodingKeys: String, CodingKey {
        case content, present
        case readError = "read_error"
    }
}

public struct NamedSpec: Codable, Equatable {
    public var name: String
    public var content: String
    public var readError: Bool

    public init(name: String, content: String, readError: Bool = false) {
        self.name = name
        self.content = content
        self.readError = readError
    }

    enum CodingKeys: String, CodingKey {
        case name, content
        case readError = "read_error"
    }
}

public struct Change: Codable, Equatable {
    public var name: String
    public var path: String
    public var created: String
    public var displayDate: String
    public var proposal: Artifact
    public var design: Artifact
    public var tasks: Artifact
    public var specs: Artifact
    public var specFiles: [NamedSpec]

    public init(name: String, path: String, created: String = "", displayDate: String = "",
                proposal: Artifact = Artifact(), design: Artifact = Artifact(),
                tasks: Artifact = Artifact(), specs: Artifact = Artifact(),
                specFiles: [NamedSpec] = []) {
        self.name = name
        self.path = path
        self.created = created
        self.displayDate = displayDate
        self.proposal = proposal
        self.design = design
        self.tasks = tasks
        self.specs = specs
        self.specFiles = specFiles
    }

    enum CodingKeys: String, CodingKey {
        case name, path, created, proposal, design, tasks, specs
        case displayDate = "display_date"
        case specFiles = "spec_files"
    }
}

public struct Project: Codable, Equatable {
    public var name: String
    public var changes: [Change]

    public init(name: String, changes: [Change] = []) {
        self.name = name
        self.changes = changes
    }
}

public struct ProjectSpec: Codable, Equatable {
    public var name: String
    public var requirementCount: Int
    public var requirementNames: [String]
    public var content: String
    public var readError: Bool

    public init(name: String, requirementCount: Int = 0, requirementNames: [String] = [],
                content: String = "", readError: Bool = false) {
        self.name = name
        self.requirementCount = requirementCount
        self.requirementNames = requirementNames
        self.content = content
        self.readError = readError
    }

    enum CodingKeys: String, CodingKey {
        case name, content
        case requirementCount = "requirement_count"
        case requirementNames = "requirement_names"
        case readError = "read_error"
    }
}

public struct ProjectConfig: Encodable, Equatable {
    public var context: String
    public var rules: [String: [String]]?

    public init(context: String = "", rules: [String: [String]]? = nil) {
        self.context = context
        self.rules = rules
    }

    enum CodingKeys: String, CodingKey {
        case context, rules
    }

    // Explicit encode so a nil `rules` serializes as `null` (absent) and an
    // empty map as `{}` — matching Go's nil-vs-empty-map distinction. The
    // synthesized Codable would `encodeIfPresent` and omit the key when nil.
    public func encode(to encoder: Encoder) throws {
        var c = encoder.container(keyedBy: CodingKeys.self)
        try c.encode(context, forKey: .context)
        try c.encode(rules, forKey: .rules)
    }
}

public enum ItemKind: String, Codable, Equatable {
    case section
    case task
}

public struct TaskItem: Codable, Equatable {
    public var kind: ItemKind
    public var text: String
    public var done: Bool
    /// Index into the raw `\n`-split lines of the source file.
    public var lineNum: Int

    // ── In-memory structured fields (macOS task editing) ──────────────────────
    // NOT part of the cross-language serialization contract: kind/text/done/
    // line_num are byte-compared against the Go-produced golden
    // (testdata/corpus/golden/tasks.json), so the fields below are excluded from
    // CodingKeys and carry defaults. They are derived from `text` by parseTasks
    // and consumed only by the editing layer.

    /// Section prefix carried verbatim from the owning `## <prefix>.` heading
    /// (e.g. "1", "3b"). Empty for section items or unnumbered tasks.
    public var sectionPrefix: String = ""
    /// 1-based ordinal within the section (the `M` in `<prefix>.M`). 0 when the
    /// task is unnumbered or for a section item.
    public var ordinal: Int = 0
    /// Description with the leading `<prefix>.<ordinal>` number and any wrapping
    /// `~~…~~` removed — the stable identity used for safe writes.
    public var taskDescription: String = ""

    public init(kind: ItemKind, text: String, done: Bool, lineNum: Int,
                sectionPrefix: String = "", ordinal: Int = 0, taskDescription: String = "") {
        self.kind = kind
        self.text = text
        self.done = done
        self.lineNum = lineNum
        self.sectionPrefix = sectionPrefix
        self.ordinal = ordinal
        self.taskDescription = taskDescription
    }

    enum CodingKeys: String, CodingKey {
        case kind, text, done
        case lineNum = "line_num"
    }
}

public struct Worktree: Codable, Equatable {
    public var path: String
    public var branch: String
    public var head: String
    public var isCurrent: Bool
    public var detached: Bool
    public var bare: Bool
    public var locked: Bool
    public var prunable: Bool

    public init(path: String, branch: String = "", head: String = "", isCurrent: Bool = false,
                detached: Bool = false, bare: Bool = false, locked: Bool = false, prunable: Bool = false) {
        self.path = path
        self.branch = branch
        self.head = head
        self.isCurrent = isCurrent
        self.detached = detached
        self.bare = bare
        self.locked = locked
        self.prunable = prunable
    }

    enum CodingKeys: String, CodingKey {
        case path, branch, head, detached, bare, locked, prunable
        case isCurrent = "is_current"
    }
}
