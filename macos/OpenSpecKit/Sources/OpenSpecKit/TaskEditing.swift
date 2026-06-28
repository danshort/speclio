import Foundation

// TaskEditing.swift — structured edits to tasks.md for the macOS app: add,
// delete, reorder, cross-section move, and inline text-edit. Every op re-reads
// the file, derives the change against current content, rewrites only the
// affected section line-span(s), and preserves existing line endings. macOS
// only; the Go TUI is unaffected. See openspec/changes/macos-task-editing/.

/// Raised when the pre-write re-read shows the file changed underneath the edit
/// (target task no longer found, or the section structure shifted). The caller
/// refreshes from disk and surfaces a visible notice; no write occurs.
public enum TaskEditError: Error, Equatable {
    case fileChanged
}

// ── Raw-line file model (line-ending preserving) ──────────────────────────────
//
// We operate on the file split by raw "\n" (NOT splitLines, which strips "\r")
// so CRLF files round-trip byte-for-byte — the same discipline as toggleTask.
// Each task/section is a single line (the OpenSpec format is flat single-line).

private let rxRawSection = Rx("^## (.+)$")
private let rxRawTask = Rx("^- \\[[ x]\\] (.+)$")
// Rewrites the number token in a task line, preserving the checkbox, an optional
// leading `~~`, and everything after the number (incl. trailing `~~ (skipped)`).
private let rxRenumber = Rx("^(- \\[[ x]\\] (?:~~)?)([0-9]+[A-Za-z]*\\.[0-9]+)(.*)$")

private struct RawSection {
    var prefix: String          // verbatim from the heading ("1", "3b"); "" if none
    var taskLineIdxs: [Int]     // indices into `lines` of this section's task lines
}

/// A parsed view of the raw lines: the line array (endings preserved) plus the
/// ordered sections with their task-line indices.
private struct RawDoc {
    var lines: [String]
    var sections: [RawSection]
    var usesCRLF: Bool

    /// Index of the section that owns task line `lineIdx`, or nil.
    func sectionIndex(owningLine lineIdx: Int) -> Int? {
        sections.firstIndex { $0.taskLineIdxs.contains(lineIdx) }
    }
}

private func parseRaw(_ content: String) -> RawDoc {
    let lines = content.components(separatedBy: "\n")
    var sections: [RawSection] = []
    var usesCRLF = false
    for (i, raw) in lines.enumerated() {
        let line = raw.hasSuffix("\r") ? String(raw.dropLast()) : raw
        if raw.hasSuffix("\r") { usesCRLF = true }
        if let h = rxRawSection.firstCapture(line) {
            sections.append(RawSection(prefix: sectionPrefix(fromHeading: h), taskLineIdxs: []))
        } else if rxRawTask.matches(line), !sections.isEmpty {
            sections[sections.count - 1].taskLineIdxs.append(i)
        } else if rxRawTask.matches(line) {
            // A task before any heading: park it in a synthetic prefix-less section.
            sections.append(RawSection(prefix: "", taskLineIdxs: [i]))
        }
    }
    return RawDoc(lines: lines, sections: sections, usesCRLF: usesCRLF)
}

/// Rewrites the number on a single task line (content without its trailing CR),
/// setting it to `<prefix>.<ordinal>` and preserving checkbox / `~~` / tail.
private func setNumber(onLine content: String, prefix: String, ordinal: Int) -> String {
    guard let head = rxRenumber.firstCapture(content, 1),
          let tail = rxRenumber.firstCapture(content, 3) else { return content }
    return "\(head)\(prefix).\(ordinal)\(tail)"
}

/// Renumbers the ordinals of a section's task lines sequentially (1..n), keeping
/// each line's section prefix set to `prefix`. Mutates `lines` in place.
private func renumber(_ lines: inout [String], section: RawSection, prefix: String) {
    for (ordinalMinus1, lineIdx) in section.taskLineIdxs.enumerated() {
        let raw = lines[lineIdx]
        let hadCR = raw.hasSuffix("\r")
        let content = hadCR ? String(raw.dropLast()) : raw
        let renumbered = setNumber(onLine: content, prefix: prefix, ordinal: ordinalMinus1 + 1)
        lines[lineIdx] = hadCR ? renumbered + "\r" : renumbered
    }
}

/// Builds a fresh pending task line `- [ ] <prefix>.<ordinal> <description>`
/// with the document's prevailing line ending. Ordinal is a placeholder fixed
/// by the subsequent renumber pass.
private func newTaskLine(prefix: String, description: String, crlf: Bool) -> String {
    let content = "- [ ] \(prefix).0 \(description)"
    return crlf ? content + "\r" : content
}

/// Identity of a raw line if it is a task, else nil (CR-tolerant).
private func lineIdentity(_ raw: String) -> String? {
    let content = raw.hasSuffix("\r") ? String(raw.dropLast()) : raw
    guard let cap = rxRawTask.firstCapture(content) else { return nil }
    return taskIdentity(cap)
}

/// Locates the raw-line index of the task with `identity` inside the section
/// whose prefix is `prefix`. nil when not found (→ conflict).
private func findTaskLine(_ doc: RawDoc, identity: String, prefix: String) -> Int? {
    for section in doc.sections where section.prefix == prefix {
        for idx in section.taskLineIdxs where lineIdentity(doc.lines[idx]) == identity {
            return idx
        }
    }
    return nil
}

/// Renumbers every section whose prefix is in `prefixes`, in place.
private func renumberSections(_ lines: inout [String], _ doc: RawDoc, prefixes: Set<String>) {
    for section in doc.sections where prefixes.contains(section.prefix) {
        renumber(&lines, section: section, prefix: section.prefix)
    }
}

/// Joins raw lines back with "\n" (line endings already embedded) and parses the
/// written content into TaskItems for the caller to display.
private func writeAndParse(_ path: String, _ lines: [String], fs: FileSystem) throws -> [TaskItem] {
    let joined = lines.joined(separator: "\n")
    try fs.writeFile(path, Data(joined.utf8))
    return parseTasks(joined)
}

// ── Edit operations ───────────────────────────────────────────────────────────
//
// Each re-reads the file, re-derives against current content, validates the
// target still exists (else throws .fileChanged), splices, renumbers the
// affected section(s), and writes. Returns the freshly parsed items.

/// Inserts a new pending task immediately after the task identified by
/// `afterIdentity` within section `prefix`. Falls back to end-of-section when
/// the anchor is absent only if `prefix` has no such task — otherwise conflict.
@discardableResult
public func addTask(_ path: String, afterIdentity: String, inSection prefix: String,
                    description: String, fs: FileSystem = OSFileSystem()) throws -> [TaskItem] {
    let content = String(decoding: try fs.readFile(path), as: UTF8.self)
    let doc = parseRaw(content)
    guard let anchor = findTaskLine(doc, identity: afterIdentity, prefix: prefix) else {
        throw TaskEditError.fileChanged
    }
    var lines = doc.lines
    lines.insert(newTaskLine(prefix: prefix, description: description, crlf: doc.usesCRLF),
                 at: anchor + 1)
    let doc2 = parseRaw(lines.joined(separator: "\n"))
    renumberSections(&lines, doc2, prefixes: [prefix])
    return try writeAndParse(path, lines, fs: fs)
}

/// Deletes the task identified by `identity` within section `prefix`, then
/// renumbers that section.
@discardableResult
public func deleteTask(_ path: String, identity: String, inSection prefix: String,
                       fs: FileSystem = OSFileSystem()) throws -> [TaskItem] {
    let content = String(decoding: try fs.readFile(path), as: UTF8.self)
    let doc = parseRaw(content)
    guard let idx = findTaskLine(doc, identity: identity, prefix: prefix) else {
        throw TaskEditError.fileChanged
    }
    var lines = doc.lines
    lines.remove(at: idx)
    let doc2 = parseRaw(lines.joined(separator: "\n"))
    renumberSections(&lines, doc2, prefixes: [prefix])
    return try writeAndParse(path, lines, fs: fs)
}

/// Replaces only the description text of the task identified by `identity`
/// within section `prefix`, preserving its checkbox state and number.
@discardableResult
public func editTaskText(_ path: String, identity: String, inSection prefix: String,
                         newDescription: String, fs: FileSystem = OSFileSystem()) throws -> [TaskItem] {
    let content = String(decoding: try fs.readFile(path), as: UTF8.self)
    let doc = parseRaw(content)
    guard let idx = findTaskLine(doc, identity: identity, prefix: prefix) else {
        throw TaskEditError.fileChanged
    }
    var lines = doc.lines
    let raw = lines[idx]
    let hadCR = raw.hasSuffix("\r")
    let lineContent = hadCR ? String(raw.dropLast()) : raw
    guard let head = rxRenumber.firstCapture(lineContent, 1),
          let number = rxRenumber.firstCapture(lineContent, 2) else {
        throw TaskEditError.fileChanged
    }
    let rebuilt = "\(head)\(number) \(newDescription)"
    lines[idx] = hadCR ? rebuilt + "\r" : rebuilt
    return try writeAndParse(path, lines, fs: fs)
}

/// Moves the task identified by `identity` (currently in section `fromPrefix`)
/// to logical position `toIndex` within the section `toPrefix`, renumbering both
/// sections. `toPrefix == fromPrefix` performs an in-section reorder. The moved
/// task adopts the destination prefix via the renumber pass.
@discardableResult
public func moveTask(_ path: String, identity: String, fromSection fromPrefix: String,
                     toSection toPrefix: String, toIndex: Int,
                     fs: FileSystem = OSFileSystem()) throws -> [TaskItem] {
    let content = String(decoding: try fs.readFile(path), as: UTF8.self)
    let doc = parseRaw(content)
    guard let srcIdx = findTaskLine(doc, identity: identity, prefix: fromPrefix) else {
        throw TaskEditError.fileChanged
    }
    var lines = doc.lines
    let moved = lines.remove(at: srcIdx)

    // Re-parse after removal to get the destination section's current slots.
    let mid = parseRaw(lines.joined(separator: "\n"))
    guard let destSection = mid.sections.first(where: { $0.prefix == toPrefix }) else {
        // Destination section vanished underneath the move.
        throw TaskEditError.fileChanged
    }
    let slots = destSection.taskLineIdxs
    let insertAt: Int
    if slots.isEmpty {
        // Empty destination section: insert right after its heading line.
        // Find the heading by scanning for the section in the re-parsed doc.
        insertAt = headingLineIndex(mid, prefix: toPrefix).map { $0 + 1 } ?? lines.count
    } else if toIndex >= slots.count {
        insertAt = slots[slots.count - 1] + 1
    } else {
        insertAt = slots[max(0, toIndex)]
    }
    lines.insert(moved, at: insertAt)

    let doc2 = parseRaw(lines.joined(separator: "\n"))
    renumberSections(&lines, doc2, prefixes: [fromPrefix, toPrefix])
    return try writeAndParse(path, lines, fs: fs)
}

/// Raw-line index of the heading for the first section with `prefix`.
private func headingLineIndex(_ doc: RawDoc, prefix: String) -> Int? {
    for (i, raw) in doc.lines.enumerated() {
        let line = raw.hasSuffix("\r") ? String(raw.dropLast()) : raw
        if let h = rxRawSection.firstCapture(line), sectionPrefix(fromHeading: h) == prefix {
            return i
        }
    }
    return nil
}
