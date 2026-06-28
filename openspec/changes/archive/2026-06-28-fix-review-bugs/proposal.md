## Why

A full code review of the recent TUI/macOS work (#90, #91/#92, #94/#95, #97) surfaced four contained bugs (#114). They're small and localized; fixing them keeps the recent work correct and tight. The cross-cutting task-identity limitation found in the same review is tracked and deferred separately (#115).

## What Changes

- **B1 — fix same-section downward reorder (macOS).** `moveTask` removes the source line before indexing into the destination's post-removal slots, so a within-section *downward* drag lands one position too low. When source and destination are the same section and the source precedes the target index, decrement the effective index. Add a downward-reorder test.
- **B2 — reap the detached editor process (TUI).** `openInEditor`'s `"system"` (detached) path calls `cmd.Start()` without `Wait()`, leaking a zombie per open. Reap it in a throwaway goroutine.
- **B3 — make Esc reliably cancel (macOS).** The inline editor's Esc (`onExitCommand`) races commit-on-blur; if focus loss is seen first the edit saves. Add an explicit "cancelling" flag the blur handler honors.
- **B4 — sanitize newlines in the engine (macOS).** `editTaskText`/`addTask` splice the description verbatim; collapse `\r`/`\n` → space so the single-line invariant holds even if a caller doesn't pre-sanitize. Add a test.
- **Polish:** `doToggle` uses `filepath.Join(ch.Path, openspec.FileTasks)` and `clearErrAfter()`; `ResolveOpener` trims its input so `"system "` isn't a custom command; correct three stale comments (`taskIdentity` strips all `~~`; the `#nosec` "no shell" note re Windows `cmd /c start`; the `addTask` end-of-section-fallback comment).

## Capabilities

### Modified Capabilities
- `macos-task-editing`: the "Reorder tasks within a section" requirement gains a downward-drag scenario pinning the corrected insertion position (B1).

## Impact

- **Code:** `macos/OpenSpecKit/Sources/OpenSpecKit/TaskEditing.swift` (B1, B4, comment), `macos/LecternApp/Sources/LecternApp/ContentView.swift` (B3, and pass the corrected index for B1), `internal/ui/viewer.go` (B2), `internal/ui/tasks.go` + `internal/config/opener.go` + `internal/openspec/tasks.go` (polish/comments).
- **Tests:** `OpenSpecKitTests` (downward reorder, newline sanitize), Go suite stays green.
- No new dependency; no cross-language contract change.

## Non-goals

- The duplicate-task-identity design fix and `ForEach(id:)` stability — tracked in #115 (explore → propose).
- Giving `performAdd` a unique placeholder — part of #115.
