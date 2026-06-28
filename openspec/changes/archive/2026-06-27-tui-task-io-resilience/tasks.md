## 1. Safe toggle (#91)

- [x] 1.1 Add `(l *Loader) ToggleTaskByText(path, text string) error` in `internal/openspec/tasks.go`: re-read the file, `ParseTasks` it, `FindCursorByText` for `text`, and only if found flip that fresh line (raw `\n` split to preserve CRLF) and write; no-op if not found
- [x] 1.2 Add a package-level `ToggleTaskByText` wrapper mirroring `ToggleTask`
- [x] 1.3 Update `internal/ui/tasks.go` to call `ToggleTaskByText(path, cursorTask.Text)` instead of `ToggleTask(path, items, cursor)`
- [x] 1.4 Tests in `internal/openspec/tasks_test.go`: toggle after the file shifted hits the right task; CRLF preserved; unknown text no-ops; mark/unmark round-trip

## 2. Surface reload/poll errors (#92)

- [x] 2.1 In `pollIndexMode` (`internal/ui/index.go`), set `m.errMsg` on the `ListChangeNamesFrom`/`ListArchiveNamesFrom`/`ListSpecNamesFrom` read errors and on the swallowed `LoadFrom` failure, instead of returning `nil` silently; retain currently displayed data
- [x] 2.2 In `pollNormalModeChanges`, set `m.errMsg` on the `ListChangeNamesFrom` read error and the swallowed `LoadFrom` failure

## 3. Verification

- [x] 3.1 `make test` (or `go test -race ./...`) passes, including the new toggle tests
- [x] 3.2 `go vet ./...` clean; `make lint` if available
- [x] 3.3 Manual: with the TUI open on a change, externally insert lines above a task, then toggle it — confirm the intended checkbox flips
- [ ] 3.4 Manual: induce a reload error (e.g. make a change dir unreadable) and confirm the status line shows it rather than the index silently freezing
