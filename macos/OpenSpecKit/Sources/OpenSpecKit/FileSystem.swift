import Foundation

// FileSystem is the injection seam mirroring Go's `fileSystem` interface. It
// also concentrates the cross-language hazards the design flags: readDir MUST
// sort by name (Go's os.ReadDir sorts; Swift's FileManager does not), and the
// not-found vs other-error distinction must be preserved (Go uses
// errors.Is(err, fs.ErrNotExist)).

public enum FSError: Error, Equatable {
    case notFound
}

public struct DirEntry: Equatable {
    public let name: String
    public let isDir: Bool
}

public struct FileInfo: Equatable {
    public let isDir: Bool
}

public protocol FileSystem {
    func readFile(_ path: String) throws -> Data
    func writeFile(_ path: String, _ data: Data) throws
    func readDir(_ path: String) throws -> [DirEntry]
    func stat(_ path: String) throws -> FileInfo
}

public struct OSFileSystem: FileSystem {
    public init() {}

    public func readFile(_ path: String) throws -> Data {
        if !FileManager.default.fileExists(atPath: path) {
            throw FSError.notFound
        }
        return try Data(contentsOf: URL(fileURLWithPath: path))
    }

    public func writeFile(_ path: String, _ data: Data) throws {
        try data.write(to: URL(fileURLWithPath: path))
    }

    public func readDir(_ path: String) throws -> [DirEntry] {
        let fm = FileManager.default
        var isDir: ObjCBool = false
        guard fm.fileExists(atPath: path, isDirectory: &isDir), isDir.boolValue else {
            throw FSError.notFound
        }
        // Sort to match Go's os.ReadDir (filename-sorted); FileManager is not.
        let names = try fm.contentsOfDirectory(atPath: path).sorted()
        return names.map { name in
            var entryIsDir: ObjCBool = false
            _ = fm.fileExists(atPath: joinPath(path, name), isDirectory: &entryIsDir)
            return DirEntry(name: name, isDir: entryIsDir.boolValue)
        }
    }

    public func stat(_ path: String) throws -> FileInfo {
        var isDir: ObjCBool = false
        guard FileManager.default.fileExists(atPath: path, isDirectory: &isDir) else {
            throw FSError.notFound
        }
        return FileInfo(isDir: isDir.boolValue)
    }
}

// ── path + string helpers (mirroring Go filepath / strings semantics) ─────────

func joinPath(_ parts: String...) -> String {
    guard var result = parts.first else { return "" }
    for part in parts.dropFirst() {
        result = (result as NSString).appendingPathComponent(part)
    }
    return result
}

func baseName(_ path: String) -> String {
    (path as NSString).lastPathComponent
}

func dirName(_ path: String) -> String {
    (path as NSString).deletingLastPathComponent
}

/// Splits on "\n" (Go strings.Split semantics: trailing empty element for a
/// trailing newline) and strips a trailing "\r" per line, so CRLF-authored
/// files parse identically to LF ones. NOT for the toggle write path.
func splitLines(_ s: String) -> [String] {
    s.components(separatedBy: "\n").map { $0.hasSuffix("\r") ? String($0.dropLast()) : $0 }
}

func trimSpace(_ s: String) -> String {
    s.trimmingCharacters(in: .whitespacesAndNewlines)
}

func stripPrefix(_ s: String, _ prefix: String) -> String {
    s.hasPrefix(prefix) ? String(s.dropFirst(prefix.count)) : s
}

/// Replaces the first occurrence of `old` (Go strings.Replace with n=1).
func replaceFirst(_ s: String, _ old: String, _ new: String) -> String {
    guard let r = s.range(of: old) else { return s }
    return s.replacingCharacters(in: r, with: new)
}

/// Go strings.Cut(s, sep): returns (before, after); after is "" when sep absent.
func cut(_ s: String, _ sep: String) -> (String, String) {
    guard let r = s.range(of: sep) else { return (s, "") }
    return (String(s[s.startIndex..<r.lowerBound]), String(s[r.upperBound...]))
}

// Thin NSRegularExpression wrapper.
struct Rx {
    let re: NSRegularExpression
    init(_ pattern: String) { re = try! NSRegularExpression(pattern: pattern) }

    func matches(_ s: String) -> Bool {
        re.firstMatch(in: s, range: NSRange(s.startIndex..<s.endIndex, in: s)) != nil
    }

    func firstCapture(_ s: String, _ group: Int = 1) -> String? {
        let range = NSRange(s.startIndex..<s.endIndex, in: s)
        guard let m = re.firstMatch(in: s, range: range),
              m.numberOfRanges > group,
              let gr = Range(m.range(at: group), in: s) else { return nil }
        return String(s[gr])
    }
}
