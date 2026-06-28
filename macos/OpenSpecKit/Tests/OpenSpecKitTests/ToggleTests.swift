import XCTest
@testable import OpenSpecKit

final class ToggleTests: XCTestCase {
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

    func testToggleLFPreservesEndings() throws {
        let path = try writeTemp("## S\n\n- [ ] 1.1 alpha\n- [x] 1.2 beta\n")
        try toggleTaskByText(path, "1.1 alpha")
        XCTAssertEqual(try read(path), "## S\n\n- [x] 1.1 alpha\n- [x] 1.2 beta\n")
    }

    func testToggleCRLFPreservesEndings() throws {
        let path = try writeTemp("## S\r\n\r\n- [ ] 1.1 alpha\r\n- [x] 1.2 beta\r\n")
        try toggleTaskByText(path, "1.2 beta")
        XCTAssertEqual(try read(path), "## S\r\n\r\n- [ ] 1.1 alpha\r\n- [ ] 1.2 beta\r\n")
    }

    // The intended task must flip even if the file was edited (line numbers
    // shifted) after it was last rendered — toggleTaskByText re-reads + re-parses
    // and matches by text, never a stale index.
    func testToggleAfterExternalModificationHitsRightTask() throws {
        let path = try writeTemp("- [ ] 1.1 alpha\n- [ ] 1.2 beta\n")
        try Data("## New Section\n\n- [ ] 1.1 alpha\n- [ ] 1.2 beta\n".utf8)
            .write(to: URL(fileURLWithPath: path))
        try toggleTaskByText(path, "1.2 beta")
        XCTAssertEqual(try read(path), "## New Section\n\n- [ ] 1.1 alpha\n- [x] 1.2 beta\n")
    }

    func testToggleUnknownTaskThrowsAndLeavesFileUnchanged() throws {
        // #101: a vanished task is a conflict (so the UI can notify), not a
        // silent no-op — but the file must still be left untouched.
        let path = try writeTemp("- [ ] 1.1 alpha\n")
        XCTAssertThrowsError(try toggleTaskByText(path, "nonexistent")) { err in
            XCTAssertEqual(err as? TaskEditError, .fileChanged)
        }
        XCTAssertEqual(try read(path), "- [ ] 1.1 alpha\n")
    }
}
