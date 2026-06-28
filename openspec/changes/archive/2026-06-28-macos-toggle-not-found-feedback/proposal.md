## Why

When the user toggles a macOS task checkbox whose text no longer exists in `tasks.md` (it was edited/removed by an agent or in an editor since render), `toggleTaskByText` silently returns the unchanged list — the click does nothing and the user gets no explanation (#101). The structured edit ops already treat a missing target as a `TaskEditError.fileChanged` conflict; the toggle should behave the same so the UI can surface a notice and refresh.

## What Changes

- `toggleTaskByText` (`OpenSpecKit/Tasks.swift`) throws `TaskEditError.fileChanged` when no task matches the given text, instead of returning the list unchanged. (Found-and-toggled and write-error behavior are unchanged.)
- The macOS toggle handler surfaces a transient notice ("Couldn't find that task — it may have changed on disk; refreshed.") and re-syncs from disk, reusing the existing conflict-notice path.
- The `ToggleTests` "unknown task" case is updated to expect the thrown conflict (and an unchanged file).

## Capabilities

### Modified Capabilities
- `macos-app`: the "Task toggling that preserves line endings" requirement gains a not-found-is-a-visible-conflict rule (no silent no-op).

## Impact

- **Code:** `macos/OpenSpecKit/Sources/OpenSpecKit/Tasks.swift` (`toggleTaskByText` throws on not-found), `macos/LecternApp/Sources/LecternApp/ContentView.swift` (`toggle` catches `.fileChanged` → notice + refresh).
- **Tests:** `OpenSpecKitTests/ToggleTests` updated; the rest stay green.
- No new dependency; no cross-language contract change (Go's toggle already returns unchanged on not-found, which the TUI surfaces differently; this is macOS-only).

## Non-goals

- The TUI toggle's not-found behavior — the Go path is separate and out of scope here.
- Toggle ambiguity (duplicate text) — handled in #115.
