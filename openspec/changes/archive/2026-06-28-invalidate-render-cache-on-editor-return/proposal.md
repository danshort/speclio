## Why

Returning from the external editor on a change artifact can show a stale or garbled render. The `editorReturnMsg` handler in `update.go` reloads the change via `mergeReloadedChange`, which only drops a tab's `renderCache` entry when the artifact content actually changed. When the user exits the editor **without saving**, the content is unchanged, so the cache is kept — but `tea.ExecProcess` re-entry races a `WindowSizeMsg`, and if the terminal was resized while the editor was open, the cached ANSI was wrapped at the old viewport width. The result is mis-wrapped / garbled output until the user forces a re-render.

The sibling worktree-change path already guards against this (`delete(m.renderCache, m.viewer.tab)` on return); the rooted-change path does not. This is the gap.

## What Changes

- On editor return for the rooted change, always invalidate the current tab's render cache (`delete(m.renderCache, m.viewer.tab)`) after reloading, regardless of whether `mergeReloadedChange` reported a content change. The subsequent `loadViewport()` then re-renders at the current width.

## Capabilities

### New Capabilities
<!-- None: no new user-facing capability. -->

### Modified Capabilities
- `editor-launch`: tighten the "Immediate reload after closing the editor" requirement so the current tab's render cache is dropped on every return, covering the unsaved-exit-then-resize case.

## Non-goals

- No change to `mergeReloadedChange`'s content-diff cache policy for the other tabs.
- No change to the worktree-change return path (already correct).
- No change to editor launch, `$EDITOR` resolution, key bindings, or spec-view return behavior.

## Impact

- **Code**: `internal/ui/update.go` (one line + comment in the `editorReturnMsg` rooted-change branch).
- **Tests**: `internal/ui/` suite stays green; add a regression test pinning cache invalidation on return.
- **Specs**: `openspec/specs/editor-launch/spec.md` — one modified requirement.
- **APIs / dependencies**: none.
