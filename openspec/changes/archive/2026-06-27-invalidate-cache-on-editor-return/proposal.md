## Why

When pressing `e` to edit an artifact (proposal, design) and returning without changes, the artifact view sometimes displays garbled content. This happens because the cached glamour-rendered ANSI was produced at a potentially different viewport width than the current one after the terminal re-entry, and the render cache is only invalidated when content actually changes.

## What Changes

- Add a single `delete(m.renderCache, m.tab)` in the `editorReturnMsg` handler to always invalidate the current tab's render cache when returning from the editor, forcing a glamour re-render at the current viewport dimensions
- After returning from editor without changes, the viewport will briefly show raw markdown while glamour re-renders (same as any other cache miss), then display correctly rendered content

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

(none)

## Impact

- **Affected code**: `internal/ui/update.go` — single line addition in `editorReturnMsg` case
- **Dependencies**: none
- **Systems**: none
- The raw-markdown flash on editor return is an acceptable tradeoff (it already happens on window resizes and tab switches)

## Scope Clarifications

- Only the current tab's cache is invalidated, not the full cache (other tabs render lazily)
- The change is only needed for `ModeNormal` and `ModeViewingArchive` (the two modes where `editorReturnMsg` can fire)
