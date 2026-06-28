## 1. Swift — ambiguity guard (OpenSpecKit)

- [x] 1.1 Add `TaskEditError.ambiguous`; change `findTaskLine` to return the single match, throw `.ambiguous` on more than one, nil on none
- [x] 1.2 Confirm `editTaskText`/`deleteTask`/`moveTask` propagate it (they already `try findTaskLine`); document the uniqueness precondition near the identity helpers
- [x] 1.3 Test: an edit targeting a duplicated description in a section throws `.ambiguous` and does not write

## 2. Swift — UI (LecternApp)

- [x] 2.1 `performAdd`: generate a unique placeholder (`New task`, `New task 2`, …) among the section's existing descriptions; select + edit it
- [x] 2.2 Surface `TaskEditError.ambiguous` in `run` with a "rename to disambiguate" notice
- [x] 2.3 Key the tasks `ForEach` on a kind-tagged stable id (task: prefix+description; section: text) instead of array offset

## 3. Go — toggle guard (TUI)

- [x] 3.1 Add `ErrAmbiguousTask` in `internal/openspec/tasks.go`; `ToggleTaskByText` returns it when the cursor's text matches more than one task (no write). Leave `FindCursorByText` first-match (cursor restore)
- [x] 3.2 Document the uniqueness precondition; `doToggle` surfaces the error via `m.errMsg`
- [x] 3.3 Test: `ToggleTaskByText` on a file with two identical task texts returns `ErrAmbiguousTask` and leaves the file unchanged

## 4. Verification

- [x] 4.1 `swift test` (OpenSpecKit) + golden green; `swift build` (LecternApp) clean
- [x] 4.2 `go test -race ./...` green; `go vet` + `gofmt` clean
- [ ] 4.3 Manual: add two tasks in a section (distinct placeholders); start editing one, reorder, confirm focus stays on the right task
