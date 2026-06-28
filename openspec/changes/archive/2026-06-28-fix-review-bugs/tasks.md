## 1. B1 — same-section downward reorder

- [x] 1.1 In `moveTask` (`TaskEditing.swift`), when `fromPrefix == toPrefix` and the source's pre-removal position precedes `toIndex`, use `toIndex - 1` for the destination slot
- [x] 1.2 Add a downward same-section reorder test (the `a b c d` → `b a c d` case) plus keep the existing upward/cross-section/end-of-section tests green

## 2. B2 — reap the detached editor process

- [x] 2.1 In `openInEditor` (`internal/ui/viewer.go`), after a successful `cmd.Start()` in the detached branch, `go func() { _ = cmd.Wait() }()`

## 3. B3 — Esc reliably cancels

- [x] 3.1 Add a `cancellingEdit` state flag; set it in `onExitCommand` (Esc) and reset it in `beginEdit`; the blur (`onChange(of: editorFocused)`) commits only when not cancelling

## 4. B4 — engine newline sanitization

- [x] 4.1 In `editTaskText` and `addTask` (`TaskEditing.swift`), collapse `\r`/`\n` → space in the description before splicing
- [x] 4.2 Add a test that an embedded newline is flattened (no stray non-task line)

## 5. Polish

- [x] 5.1 `doToggle` (`internal/ui/tasks.go`): `filepath.Join(ch.Path, openspec.FileTasks)` and `clearErrAfter()`
- [x] 5.2 `ResolveOpener` (`internal/config/opener.go`): `TrimSpace` the value before the switch; add a `" system "` test
- [x] 5.3 Correct comments: `taskIdentity` strips all `~~`; `#nosec` note re Windows `cmd /c start`; `addTask` doc fallback claim

## 6. Verification

- [x] 6.1 `swift test` (OpenSpecKit) green; `swift build` (LecternApp) clean
- [x] 6.2 `go test -race ./...` green; `go vet` + `gofmt` clean
- [ ] 6.3 Manual: drag a task downward within a section in the macOS app and confirm it lands at the drop line; press Esc mid-edit and confirm no save
