// Layout centralizes the OpenSpec directory/artifact names and the header
// prefixes, mirroring the consts in Go internal/openspec/loader.go.
enum Layout {
    static let dirOpenspec = "openspec"
    static let dirChanges = "changes"
    static let dirSpecs = "specs"
    static let dirArchive = "archive"

    static let fileProposal = "proposal.md"
    static let fileDesign = "design.md"
    static let fileTasks = "tasks.md"
    static let fileSpec = "spec.md"
    static let fileConfig = "config.yaml"
    static let fileMeta = ".openspec.yaml"

    static let reqPrefix = "### Requirement:"
    static let scenarioPrefix = "#### Scenario:"

    static let unreadablePrefix = "⚠ couldn't read "

    // Placeholder prefix for a spec capability directory that exists but has no
    // spec.md — surfaced with a ⚠ rather than dropped/empty (#96). Byte-
    // identical to the Go loader's missingSpecPrefix (cross-language contract).
    static let missingSpecPrefix = "⚠ no spec.md in "
}
