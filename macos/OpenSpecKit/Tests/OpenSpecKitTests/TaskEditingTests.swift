import XCTest
@testable import OpenSpecKit

final class TaskEditingTests: XCTestCase {
    private func writeTemp(_ content: String) throws -> String {
        let dir = (NSTemporaryDirectory() as NSString).appendingPathComponent("osk-\(UUID().uuidString)")
        try FileManager.default.createDirectory(atPath: dir, withIntermediateDirectories: true)
        let path = (dir as NSString).appendingPathComponent("tasks.md")
        try Data(content.utf8).write(to: URL(fileURLWithPath: path))
        return path
    }

    private func read(_ path: String) throws -> String {
        String(decoding: try Data(contentsOf: URL(fileURLWithPath: path)), as: UTF8.self)
    }

    // ── Parsing: number / identity / strikethrough / letter prefix (1.3, 1.4) ──

    func testParseExtractsNumberPrefixOrdinalAndIdentity() {
        let items = parseTasks("## 1. Setup\n\n- [ ] 1.1 Create module\n- [x] 1.2 Wire it\n")
        let tasks = items.filter { $0.kind == .task }
        XCTAssertEqual(tasks[0].sectionPrefix, "1")
        XCTAssertEqual(tasks[0].ordinal, 1)
        XCTAssertEqual(tasks[0].taskDescription, "Create module")
        XCTAssertEqual(tasks[1].ordinal, 2)
        XCTAssertTrue(tasks[1].done)
    }

    func testParseLetterSuffixSectionPrefix() {
        let items = parseTasks("## 3b. Recents\n\n- [x] 3b.1 Maintain list\n- [x] 3b.2 Add menu\n")
        let tasks = items.filter { $0.kind == .task }
        XCTAssertEqual(tasks[0].sectionPrefix, "3b")
        XCTAssertEqual(tasks[0].ordinal, 1)
        XCTAssertEqual(tasks[0].taskDescription, "Maintain list")
    }

    func testIdentityStripsStrikethroughAndNumber() {
        XCTAssertEqual(taskIdentity("~~6.1 Drop the cache~~ (skipped)"), "Drop the cache (skipped)")
        XCTAssertEqual(taskIdentity("1.3 Add frontend component"), "Add frontend component")
    }

    // ── Add (after selected) renumbers the tail (2.x, 3.1) ─────────────────────

    func testAddInsertsAfterSelectedAndRenumbers() throws {
        let path = try writeTemp("## 1. S\n\n- [ ] 1.1 a\n- [ ] 1.2 b\n- [ ] 1.3 c\n")
        try addTask(path, afterIdentity: "b", inSection: "1", description: "new")
        XCTAssertEqual(try read(path),
            "## 1. S\n\n- [ ] 1.1 a\n- [ ] 1.2 b\n- [ ] 1.3 new\n- [ ] 1.4 c\n")
    }

    // ── Delete renumbers the tail; surrounding sections untouched (2.4, 3.2) ───

    func testDeleteRenumbersTail() throws {
        let path = try writeTemp("## 1. S\n\n- [ ] 1.1 a\n- [ ] 1.2 b\n- [ ] 1.3 c\n")
        try deleteTask(path, identity: "b", inSection: "1")
        XCTAssertEqual(try read(path), "## 1. S\n\n- [ ] 1.1 a\n- [ ] 1.2 c\n")
    }

    func testDeleteLeavesOtherSectionsByteIdentical() throws {
        let src = "## 1. One\n\n- [ ] 1.1 a\n- [ ] 1.2 b\n\n## 2. Two\n\n- [x] 2.1 keep\n- [ ] 2.2 keep2\n"
        let path = try writeTemp(src)
        try deleteTask(path, identity: "a", inSection: "1")
        XCTAssertEqual(try read(path),
            "## 1. One\n\n- [ ] 1.1 b\n\n## 2. Two\n\n- [x] 2.1 keep\n- [ ] 2.2 keep2\n")
    }

    // ── Inline reorder within a section (2.4, 3.3) ─────────────────────────────

    func testReorderWithinSection() throws {
        let path = try writeTemp("## 1. S\n\n- [ ] 1.1 a\n- [ ] 1.2 b\n- [ ] 1.3 c\n")
        // Move "c" to the top (index 0).
        try moveTask(path, identity: "c", fromSection: "1", toSection: "1", toIndex: 0)
        XCTAssertEqual(try read(path),
            "## 1. S\n\n- [ ] 1.1 c\n- [ ] 1.2 a\n- [ ] 1.3 b\n")
    }

    // ── Cross-section move adopts destination prefix, renumbers both (3.4) ─────

    func testMoveAcrossSectionsAdoptsPrefixAndRenumbers() throws {
        let src = "## 1. One\n\n- [ ] 1.1 a\n- [ ] 1.2 b\n\n## 3. Three\n\n- [ ] 3.1 x\n- [ ] 3.2 y\n"
        let path = try writeTemp(src)
        // Move "b" to end of section 3.
        try moveTask(path, identity: "b", fromSection: "1", toSection: "3", toIndex: 99)
        XCTAssertEqual(try read(path),
            "## 1. One\n\n- [ ] 1.1 a\n\n## 3. Three\n\n- [ ] 3.1 x\n- [ ] 3.2 y\n- [ ] 3.3 b\n")
    }

    func testMoveIntoLetterPrefixSectionPreservesPrefix() throws {
        let src = "## 1. One\n\n- [ ] 1.1 a\n- [ ] 1.2 b\n\n## 3b. Bee\n\n- [ ] 3b.1 x\n"
        let path = try writeTemp(src)
        try moveTask(path, identity: "b", fromSection: "1", toSection: "3b", toIndex: 99)
        XCTAssertEqual(try read(path),
            "## 1. One\n\n- [ ] 1.1 a\n\n## 3b. Bee\n\n- [ ] 3b.1 x\n- [ ] 3b.2 b\n")
    }

    // ── Inline text edit preserves checkbox + number (3.5) ─────────────────────

    func testEditTextPreservesStateAndNumber() throws {
        let path = try writeTemp("## 2. S\n\n- [x] 2.1 Implment API\n")
        try editTaskText(path, identity: "Implment API", inSection: "2", newDescription: "Implement API")
        XCTAssertEqual(try read(path), "## 2. S\n\n- [x] 2.1 Implement API\n")
    }

    // ── Conflict: target no longer present → abort, no write (3.6, D5) ─────────

    func testConflictThrowsAndDoesNotWrite() throws {
        let src = "## 1. S\n\n- [ ] 1.1 a\n"
        let path = try writeTemp(src)
        XCTAssertThrowsError(try deleteTask(path, identity: "gone", inSection: "1")) { err in
            XCTAssertEqual(err as? TaskEditError, .fileChanged)
        }
        XCTAssertEqual(try read(path), src) // unchanged
    }

    // ── CRLF round-trips for a structural edit ─────────────────────────────────

    func testDeletePreservesCRLF() throws {
        let path = try writeTemp("## 1. S\r\n\r\n- [ ] 1.1 a\r\n- [ ] 1.2 b\r\n")
        try deleteTask(path, identity: "a", inSection: "1")
        XCTAssertEqual(try read(path), "## 1. S\r\n\r\n- [ ] 1.1 b\r\n")
    }
}
