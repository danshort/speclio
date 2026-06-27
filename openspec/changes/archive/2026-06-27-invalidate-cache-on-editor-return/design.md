## Context

The render cache (`m.renderCache[Tab]string`) stores glamour-rendered ANSI for each artifact tab. On editor return (`editorReturnMsg` in `update.go:62-75`), the handler reloads the change from disk via `mergeReloadedChange`, which only deletes the cache entry when content actually changed. If the user exited without saving, the cache is preserved.

After `tea.ExecProcess` completes, two asynchronous messages race to the event loop:
- `editorReturnMsg` (from the exec callback)
- `WindowSizeMsg` (from `RestoreTerminal → checkResize`)

When `editorReturnMsg` wins the race and the cache is hit, the old ANSI (rendered at the previous viewport width) is displayed. If the viewport dimensions changed during terminal re-entry, the cached ANSI has wrong line breaks, producing garbled output.

## Goals / Non-Goals

**Goals:**
- Always re-render the current artifact after returning from the editor
- Force a glamour re-render at the current viewport dimensions
- Minimal, localized change

**Non-Goals:**
- Invalidating the full render cache (other tabs are unaffected)
- Changing how `WindowSizeMsg` or `checkResize` works
- Addressing the raw-markdown flash that occurs during any cache miss

## Decisions

### Always invalidate current tab's cache on editor return

Add `delete(m.renderCache, m.tab)` in the `editorReturnMsg` handler before `m.loadViewport()`.

**Alternatives considered:**

- **Invalidate full cache** (`m.renderCache = make(map[Tab]string)`): Unnecessary — other tabs have no reason to be stale and will re-render lazily when visited.
- **Compare viewport width before/after**: Over-engineered for a one-line fix. Width can drift by 1 column in ways that are hard to detect reliably.
- **Only invalidate on width change**: Brittle — the issue isn't just width changes but the race between messages.

### Only current tab, not full cache

The `delete` built-in is a no-op when the key doesn't exist (if `mergeReloadedChange` already cleared it). It's safe whether or not content changed.

## Risks / Trade-offs

- **Raw-markdown flash**: After editor return, the viewport briefly shows raw markdown while glamour re-renders. This already happens during window resizes, tab switches, and content changes — it's an accepted trade-off in the current architecture.
- **Double glamour render**: If `WindowSizeMsg` arrives first, it clears the full cache and starts a glamour render. Then `editorReturnMsg`'s `delete` is a no-op (already cleared) and starts a second glamour render. Both complete with the same content; the second one calls `GotoTop()` redundantly. Harmless.
