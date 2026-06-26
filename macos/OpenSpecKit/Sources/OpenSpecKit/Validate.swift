import Foundation

// Validate.swift mirrors Go internal/openspec/validate.go. Messages are
// byte-identical to the Go output (pinned by validation.json).

private let deltaHeaderRx = Rx("^## (ADDED|MODIFIED|REMOVED|RENAMED) Requirements\\s*$")

private func hasHeader(_ lines: [String], _ header: String) -> Bool {
    // HasPrefix, not equality: `## Purpose` matches `## Purpose and Scope`.
    lines.contains { $0.hasPrefix(header) }
}

public func validateSpec(_ content: String) -> [String] {
    let lines = splitLines(content)
    var errs: [String] = []
    if !hasHeader(lines, "## Purpose") {
        errs.append("missing \"## Purpose\" section")
    }
    if !hasHeader(lines, "## Requirements") {
        errs.append("missing \"## Requirements\" section")
    }
    errs.append(contentsOf: requirementsMissingScenarios(lines, prefix: ""))
    return errs
}

public func validateChange(_ ch: Change) -> [String] {
    var errs: [String] = []
    if !ch.proposal.present {
        errs.append("missing proposal.md")
    }
    for sf in ch.specFiles {
        if sf.readError { continue } // read failure ≠ structural failure
        errs.append(contentsOf: validateDeltaSpec(sf.name, sf.content))
    }
    return errs
}

private func requirementsMissingScenarios(_ lines: [String], prefix: String) -> [String] {
    var errs: [String] = []
    var curReq = ""
    var hasScenario = false
    func flush() {
        if !curReq.isEmpty && !hasScenario {
            errs.append(prefix + "requirement \"" + curReq + "\" has no \"#### Scenario:\"")
        }
    }
    for l in lines {
        if l.hasPrefix(Layout.reqPrefix) {
            flush()
            curReq = trimSpace(String(l.dropFirst(Layout.reqPrefix.count)))
            hasScenario = false
        } else if l.hasPrefix(Layout.scenarioPrefix) {
            if !curReq.isEmpty { hasScenario = true }
        } else if l.hasPrefix("## ") {
            flush()
            curReq = ""
            hasScenario = false
        }
    }
    flush()
    return errs
}

private func validateDeltaSpec(_ name: String, _ content: String) -> [String] {
    let lines = splitLines(content)

    var hasDelta = false
    for l in lines where deltaHeaderRx.matches(l) {
        hasDelta = true
        break
    }
    if !hasDelta {
        return ["delta spec \"" + name + "\" has no delta header (## ADDED/MODIFIED/REMOVED/RENAMED Requirements)"]
    }

    var errs: [String] = []
    var section = ""
    var curReq = ""
    var hasScenario = false
    func flush() {
        if !curReq.isEmpty && (section == "ADDED" || section == "MODIFIED") && !hasScenario {
            errs.append("delta spec \"" + name + "\": requirement \"" + curReq + "\" in " + section + " section has no scenario")
        }
    }
    for l in lines {
        if let sec = deltaHeaderRx.firstCapture(l) {
            flush()
            section = sec
            curReq = ""
            hasScenario = false
            continue
        }
        if l.hasPrefix(Layout.reqPrefix) {
            flush()
            curReq = trimSpace(String(l.dropFirst(Layout.reqPrefix.count)))
            hasScenario = false
        } else if l.hasPrefix(Layout.scenarioPrefix) {
            if !curReq.isEmpty { hasScenario = true }
        }
    }
    flush()
    return errs
}
