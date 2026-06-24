## 1. Validator (pure, dependency-free)

- [x] 1.1 Add `internal/openspec/validate.go` with `ValidateSpec(content string) []string`: requires `## Purpose` and `## Requirements`; every `### Requirement:` block has at least one `#### Scenario:`
- [x] 1.2 Add `ValidateChange(ch Change) []string`: proposal present; every delta spec has a delta header; every `### Requirement:` in an `ADDED`/`MODIFIED` section has a scenario
- [x] 1.3 Add table-driven tests in `internal/openspec/validate_test.go` (valid spec, missing Purpose, missing Requirements, requirement without scenario; valid change, missing proposal, delta without header, ADDED requirement without scenario)

## 2. Surface in the index

- [x] 2.1 In `renderIndexContent`, compute validity per project spec and active change and append a red `errStyle` marker to invalid items (archived changes are not marked — frozen history)
- [x] 2.2 Confirm the trailing marker does not shift the click hit-testing line math (`indexItemAtContentLine`) — markers add no lines
- [x] 2.3 Add a UI test asserting an invalid spec renders the marker and an all-valid index does not

## 3. Surface in the spec view

- [x] 3.1 When `ModeViewingSpec` shows an invalid spec, prepend a styled "Validation errors:" block listing `ValidateSpec` messages above the rendered content

## 4. Verification

- [x] 4.1 `gofmt`, `go vet`, `go test -race ./...` pass
- [x] 4.2 Confirm against this repo: project specs and the active change validate clean; archived history is not marked (audited 58 archived changes — markers intentionally suppressed there)
