import Foundation

// Tasks.swift mirrors Go internal/openspec/tasks.go.

private let rxSection = Rx("^## (.+)$")
private let rxPending = Rx("^- \\[ \\] (.+)$")
private let rxDone = Rx("^- \\[x\\] (.+)$")

// Section prefix: the token before the first ". " in a heading, e.g. "1" from
// "1. Setup" or "3b" from "3b. Foo". Digits then optional letters.
private let rxSectionPrefix = Rx("^([0-9]+[A-Za-z]*)\\.")
// Task number at the start of a (de-struck) description: <prefix>.<ordinal>.
private let rxTaskNumber = Rx("^([0-9]+[A-Za-z]*)\\.([0-9]+)\\s+(.*)$")

/// sectionPrefix extracts the verbatim prefix from a section heading's captured
/// text (e.g. "1. Setup" -> "1", "3b. Foo" -> "3b"). Empty if not numbered.
func sectionPrefix(fromHeading heading: String) -> String {
    rxSectionPrefix.firstCapture(heading) ?? ""
}

/// taskIdentity is the stable key for safe writes: the description with any
/// wrapping `~~…~~` strikethrough markers removed and the leading
/// `<prefix>.<ordinal>` number stripped. Renumbering changes the number but not
/// the identity, so re-read-before-write still finds the right line.
public func taskIdentity(_ taskText: String) -> String {
    let unstruck = taskText.replacingOccurrences(of: "~~", with: "")
    if let m = rxTaskNumber.firstCapture(unstruck, 3) {
        return trimSpace(m)
    }
    return trimSpace(unstruck)
}

/// taskNumber parses the `<prefix>.<ordinal>` from a (possibly struck-through)
/// task description. Returns nil when the task is unnumbered.
func taskNumber(_ taskText: String) -> (prefix: String, ordinal: Int)? {
    let unstruck = taskText.replacingOccurrences(of: "~~", with: "")
    guard let p = rxTaskNumber.firstCapture(unstruck, 1),
          let oStr = rxTaskNumber.firstCapture(unstruck, 2),
          let o = Int(oStr) else { return nil }
    return (p, o)
}

public func parseTasks(_ content: String) -> [TaskItem] {
    var items: [TaskItem] = []
    var currentPrefix = ""
    for (i, line) in splitLines(content).enumerated() {
        if let m = rxSection.firstCapture(line) {
            currentPrefix = sectionPrefix(fromHeading: m)
            items.append(TaskItem(kind: .section, text: m, done: false, lineNum: i,
                                  sectionPrefix: currentPrefix))
        } else if let m = rxPending.firstCapture(line) {
            items.append(makeTask(text: m, done: false, lineNum: i, prefix: currentPrefix))
        } else if let m = rxDone.firstCapture(line) {
            items.append(makeTask(text: m, done: true, lineNum: i, prefix: currentPrefix))
        }
    }
    return items
}

/// makeTask builds a task item, deriving its ordinal (from its own number) and
/// stable identity. The task adopts the current section's prefix for membership.
private func makeTask(text: String, done: Bool, lineNum: Int, prefix: String) -> TaskItem {
    let ordinal = taskNumber(text)?.ordinal ?? 0
    return TaskItem(kind: .task, text: text, done: done, lineNum: lineNum,
                    sectionPrefix: prefix, ordinal: ordinal, taskDescription: taskIdentity(text))
}

/// findCursorByText mirrors Go: the first task with the given text, else the
/// first task, else 0.
public func findCursorByText(_ items: [TaskItem], _ text: String) -> Int {
    var first = -1
    for (i, item) in items.enumerated() where item.kind == .task {
        if first == -1 { first = i }
        if item.text == text { return i }
    }
    return first == -1 ? 0 : first
}

/// toggleTaskByText re-reads and re-parses the file, finds the first task whose
/// text matches, and toggles it on disk — never trusting a stale line index, so
/// an external edit between render and toggle can't flip the wrong line. Returns
/// the updated items (unchanged if no task matches). This is what a live-
/// reloading GUI should call.
@discardableResult
public func toggleTaskByText(_ path: String, _ text: String,
                             fs: FileSystem = OSFileSystem()) throws -> [TaskItem] {
    let data = try fs.readFile(path)
    var items = parseTasks(String(decoding: data, as: UTF8.self))
    let idx = findCursorByText(items, text)
    guard idx < items.count, items[idx].kind == .task, items[idx].text == text else {
        return items
    }
    try toggleTask(path, &items, idx, fs: fs)
    return items
}

/// toggleTask flips items[idx] in memory (inout) and on disk. The write path
/// splits on raw "\n" WITHOUT stripping "\r", so CRLF files keep CRLF endings —
/// the asymmetry with splitLines is deliberate. It re-reads the file first so a
/// stale line index cannot toggle the wrong line in a live-reloading GUI.
public func toggleTask(_ path: String, _ items: inout [TaskItem], _ idx: Int,
                       fs: FileSystem = OSFileSystem()) throws {
    let data = try fs.readFile(path)
    var lines = String(decoding: data, as: UTF8.self).components(separatedBy: "\n")
    guard idx < items.count, items[idx].lineNum < lines.count else { return }
    let ln = items[idx].lineNum
    if items[idx].done {
        lines[ln] = replaceFirst(lines[ln], "- [x] ", "- [ ] ")
        items[idx].done = false
    } else {
        lines[ln] = replaceFirst(lines[ln], "- [ ] ", "- [x] ")
        items[idx].done = true
    }
    try fs.writeFile(path, Data(lines.joined(separator: "\n").utf8))
}
