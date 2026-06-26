import Foundation

// Tasks.swift mirrors Go internal/openspec/tasks.go.

private let rxSection = Rx("^## (.+)$")
private let rxPending = Rx("^- \\[ \\] (.+)$")
private let rxDone = Rx("^- \\[x\\] (.+)$")

public func parseTasks(_ content: String) -> [TaskItem] {
    var items: [TaskItem] = []
    for (i, line) in splitLines(content).enumerated() {
        if let m = rxSection.firstCapture(line) {
            items.append(TaskItem(kind: .section, text: m, done: false, lineNum: i))
        } else if let m = rxPending.firstCapture(line) {
            items.append(TaskItem(kind: .task, text: m, done: false, lineNum: i))
        } else if let m = rxDone.firstCapture(line) {
            items.append(TaskItem(kind: .task, text: m, done: true, lineNum: i))
        }
    }
    return items
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
