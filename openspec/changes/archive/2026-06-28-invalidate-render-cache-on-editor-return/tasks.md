## 1. Fix

- [x] 1.1 In `update.go` `editorReturnMsg` rooted-change branch, `delete(m.renderCache, m.viewer.tab)` after the reload/merge, with a comment explaining the unsaved-exit-then-resize case
- [x] 1.2 `go build ./...`, `go vet ./internal/ui/`, and `gofmt -l` clean

## 2. Tests

- [x] 2.1 Add a regression test in `internal/ui/` asserting that returning from the editor drops the current tab's `renderCache` entry even when the reloaded content is unchanged
- [x] 2.2 `go test ./internal/ui/` green

## 3. Validation

- [x] 3.1 `openspec validate invalidate-render-cache-on-editor-return --strict` passes
