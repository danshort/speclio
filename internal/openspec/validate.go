package openspec

import (
	"regexp"
	"strings"
)

// validate.go implements a lightweight, dependency-free subset of the OpenSpec
// structural validation rules — enough to flag malformed artifacts in the
// reader without shelling out to the external `openspec` CLI. All functions are
// pure (operate on in-memory content) so callers can revalidate freely.

var deltaHeaderRe = regexp.MustCompile(`^## (ADDED|MODIFIED|REMOVED|RENAMED) Requirements\s*$`)

func hasHeader(lines []string, header string) bool {
	for _, l := range lines {
		if strings.HasPrefix(l, header) {
			return true
		}
	}
	return false
}

// ValidateSpec checks a main spec's structure: it must contain a "## Purpose"
// section and a "## Requirements" section, and every "### Requirement:" must
// contain at least one "#### Scenario:". It returns one message per violation;
// an empty result means the spec is valid.
func ValidateSpec(content string) []string {
	lines := splitLines(content)
	var errs []string
	if !hasHeader(lines, "## Purpose") {
		errs = append(errs, `missing "## Purpose" section`)
	}
	if !hasHeader(lines, "## Requirements") {
		errs = append(errs, `missing "## Requirements" section`)
	}
	errs = append(errs, requirementsMissingScenarios(lines, "")...)
	return errs
}

// ValidateChange checks a change: it must have a proposal, and each delta spec
// must contain at least one delta header, with every requirement in an ADDED or
// MODIFIED section containing at least one scenario.
func ValidateChange(ch Change) []string {
	var errs []string
	if !ch.Proposal.Present {
		errs = append(errs, "missing proposal.md")
	}
	for _, sf := range ch.SpecFiles {
		if sf.ReadErr != nil {
			// Unreadable delta spec: a read failure, not a structural one —
			// surfaced elsewhere (⚠), never validated as malformed.
			continue
		}
		errs = append(errs, validateDeltaSpec(sf.Name, sf.Content)...)
	}
	return errs
}

// requirementsMissingScenarios returns an error for each "### Requirement:"
// block that lacks a "#### Scenario:". A new "## " section or "### Requirement:"
// ends the current block. prefix is prepended to each message for context.
func requirementsMissingScenarios(lines []string, prefix string) []string {
	var errs []string
	curReq := ""
	hasScenario := false
	flush := func() {
		if curReq != "" && !hasScenario {
			errs = append(errs, prefix+`requirement "`+curReq+`" has no "#### Scenario:"`)
		}
	}
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, reqPrefix):
			flush()
			curReq = strings.TrimSpace(strings.TrimPrefix(l, reqPrefix))
			hasScenario = false
		case strings.HasPrefix(l, scenarioPrefix):
			if curReq != "" {
				hasScenario = true
			}
		case strings.HasPrefix(l, "## "):
			flush()
			curReq = ""
			hasScenario = false
		}
	}
	flush()
	return errs
}

// validateDeltaSpec validates a change's delta spec: it must have at least one
// delta header, and every requirement under an ADDED/MODIFIED section must have
// a scenario (REMOVED/RENAMED sections are exempt).
func validateDeltaSpec(name, content string) []string {
	lines := splitLines(content)

	hasDelta := false
	for _, l := range lines {
		if deltaHeaderRe.MatchString(l) {
			hasDelta = true
			break
		}
	}
	if !hasDelta {
		return []string{`delta spec "` + name + `" has no delta header (## ADDED/MODIFIED/REMOVED/RENAMED Requirements)`}
	}

	var errs []string
	section := ""
	curReq := ""
	hasScenario := false
	flush := func() {
		if curReq != "" && (section == "ADDED" || section == "MODIFIED") && !hasScenario {
			errs = append(errs, `delta spec "`+name+`": requirement "`+curReq+`" in `+section+` section has no scenario`)
		}
	}
	for _, l := range lines {
		if mm := deltaHeaderRe.FindStringSubmatch(l); mm != nil {
			flush()
			section = mm[1]
			curReq = ""
			hasScenario = false
			continue
		}
		switch {
		case strings.HasPrefix(l, reqPrefix):
			flush()
			curReq = strings.TrimSpace(strings.TrimPrefix(l, reqPrefix))
			hasScenario = false
		case strings.HasPrefix(l, scenarioPrefix):
			if curReq != "" {
				hasScenario = true
			}
		}
	}
	flush()
	return errs
}
