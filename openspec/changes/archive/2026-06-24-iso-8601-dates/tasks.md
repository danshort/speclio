## 1. Implementation

- [x] 1.1 In `internal/openspec/loader.go`, change the `parseArchiveName` output layout from `02/01/2006` to `2006-01-02` (ISO 8601).

## 2. Tests

- [x] 2.1 Add a `parseArchiveName` unit test in `internal/openspec/loader_test.go` asserting that `2026-05-02-specs-subnav` yields name `specs-subnav` and date `2026-05-02`.
- [x] 2.2 Update `DisplayDate` fixtures in `internal/ui/view_test.go` (lines ~122 and ~702) from `DD/MM/YYYY` to ISO `YYYY-MM-DD` form.
- [x] 2.3 Run `go test ./...` and confirm all tests pass.

## 3. Verification

- [x] 3.1 Run `openspec validate iso-8601-dates` and confirm the change is valid.
- [x] 3.2 Build and run the TUI against a project with archived changes; confirm the index renders archived dates as `YYYY-MM-DD`.
