## Context

The render cache (`m.renderCache map[Tab]string`) stores glamour-rendered ANSI per artifact tab. On editor return (`editorReturnMsg`, `update.go`), the rooted-change branch reloads the change from disk via `mergeReloadedChange`, which deletes `renderCache[TabProposal|TabDesign|TabSpecs]` **only when** the reloaded content differs from what is held in memory.

After `tea.ExecProcess` completes, two messages race to the event loop:
- `editorReturnMsg` (from the exec callback), and
- `WindowSizeMsg` (from terminal re-entry / `checkResize`).

When `editorReturnMsg` wins and the user exited without saving, the cache is a hit on content that was rendered at the *previous* viewport width. If the terminal was resized while the editor was open, the cached ANSI has stale line breaks → garbled output.

## Decision

In the rooted-change branch of `editorReturnMsg`, unconditionally `delete(m.renderCache, m.viewer.tab)` after the reload/merge, then fall through to `loadViewport()`, which re-renders the current tab at the live width.

- Use `m.viewer.tab` (the active artifact tab in this fork's `viewerState`), matching the adjacent worktree-change path and `mergeReloadedChange`. The top-level `m.tab` field is vestigial here and must not be used.
- Scope the invalidation to the current tab only; other tabs keep `mergeReloadedChange`'s content-diff policy, so an instant return to an unchanged, unresized tab stays cache-friendly.

## Risks / Trade-offs

- One extra glamour render of the current tab on every editor return. Negligible: it is a single tab, already the common path, and only when actually viewing.

## Migration

None — internal behavior fix, no spec-visible API or key-binding change.
