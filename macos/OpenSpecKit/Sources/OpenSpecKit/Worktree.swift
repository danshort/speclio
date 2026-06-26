import Foundation

// Worktree.swift mirrors Go internal/openspec/worktree.go. Per the App Sandbox
// decision (Option C), git access lives behind a GitService protocol with the
// pure porcelain PARSER separated from the Process invocation, so a future
// sandbox flip swaps only the invocation. The parser is golden-tested; the
// Process-based service is integration-tested in a later phase.

public protocol GitService {
    func listWorktrees(root: String) throws -> [Worktree]
}

/// parseWorktreeList parses `git worktree list --porcelain`. Each worktree is a
/// block of "key value" lines terminated by a blank line; valueless keys
/// (detached/bare/locked/prunable) appear without a value.
public func parseWorktreeList(_ out: Data) -> [Worktree] {
    var wts: [Worktree] = []
    var cur: Worktree?
    func flush() {
        if let c = cur { wts.append(c); cur = nil }
    }
    for rawLine in String(decoding: out, as: UTF8.self).components(separatedBy: "\n") {
        var line = rawLine
        while line.hasSuffix("\r") { line = String(line.dropLast()) }
        if line.isEmpty {
            flush()
            continue
        }
        let (key, val) = cut(line, " ")
        switch key {
        case "worktree":
            flush()
            cur = Worktree(path: val)
        case "HEAD":
            cur?.head = val
        case "branch":
            cur?.branch = stripPrefix(val, "refs/heads/")
        case "detached":
            cur?.detached = true
        case "bare":
            cur?.bare = true
        case "locked":
            cur?.locked = true
        case "prunable":
            cur?.prunable = true
        default:
            break
        }
    }
    flush()
    return wts
}

func markCurrentWorktree(_ wts: inout [Worktree], toplevel: String) {
    if toplevel.isEmpty { return }
    let target = normalizePath(toplevel)
    for i in wts.indices where normalizePath(wts[i].path) == target {
        wts[i].isCurrent = true
        return
    }
}

/// normalizePath resolves symlinks (like Go filepath.EvalSymlinks) with a
/// lexical fallback when the path does not exist (Go falls back to
/// filepath.Clean). NOT URL.resolvingSymlinksInPath, which behaves differently.
func normalizePath(_ p: String) -> String {
    var buf = [CChar](repeating: 0, count: Int(PATH_MAX))
    if realpath(p, &buf) != nil {
        return String(cString: buf)
    }
    return (p as NSString).standardizingPath
}

/// ProcessGitService shells out to `git` with a timeout, then delegates to the
/// pure parser above. Integration-tested in a later phase (needs a real repo).
public struct ProcessGitService: GitService {
    public init() {}

    public func listWorktrees(root: String) throws -> [Worktree] {
        let out = try runGit(dir: root, args: ["worktree", "list", "--porcelain"])
        var wts = parseWorktreeList(out)
        if let top = try? runGit(dir: root, args: ["rev-parse", "--show-toplevel"]) {
            markCurrentWorktree(&wts, toplevel: trimSpace(String(decoding: top, as: UTF8.self)))
        }
        return wts
    }

    private func runGit(dir: String, args: [String], timeout: TimeInterval = 5) throws -> Data {
        let proc = Process()
        proc.executableURL = URL(fileURLWithPath: "/usr/bin/env")
        proc.arguments = ["git", "-C", dir] + args
        let pipe = Pipe()
        proc.standardOutput = pipe
        proc.standardError = Pipe()
        try proc.run()

        let deadline = DispatchTime.now() + timeout
        let group = DispatchGroup()
        group.enter()
        DispatchQueue.global().async {
            proc.waitUntilExit()
            group.leave()
        }
        if group.wait(timeout: deadline) == .timedOut {
            proc.terminate()
            throw LoaderError.notFound("git timed out")
        }
        return pipe.fileHandleForReading.readDataToEndOfFile()
    }
}
