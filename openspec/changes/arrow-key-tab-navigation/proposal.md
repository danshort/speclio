## Why

When viewing a change, tabs can be switched with `1`–`4`, `Tab`/`Shift+Tab`, or a mouse click — but not the arrow keys, which is the most reflexive "move sideways" gesture. Left/Right arrows are currently unbound in the change viewer, so the natural motion does nothing.

## What Changes

- In the change viewer, `→` (right) switches to the next available tab and `←` (left) to the previous, mirroring `Tab` / `Shift+Tab` exactly (skip disabled tabs, wrap at the ends).
- `h`/`l` keep their existing meaning (previous/next change) — unchanged.
- Update the in-app footer hint and the README keyboard references (EN + ES).

## Non-goals

- No change to `h`/`l` (change navigation).
- Arrows do not cycle spec files; that stays on the `3` key (consistent with `Tab`, which also does not cycle spec files).
- No new tab or reordering of tabs.

## Capabilities

### Modified Capabilities

- `tui-viewer`: the "Tabs de artifact" requirement gains `←`/`→` as secondary tab navigation alongside `1`–`4`, `Tab`/`Shift+Tab`, and mouse clicks.

## Impact

- `internal/ui/viewer.go` — add `left`/`right` to the `Tab`/`Shift+Tab` handlers
- `internal/ui/view.go` — footer hint shows `1-4/Tab/←→: artifact`
- `internal/ui/view_test.go` — tests for arrow tab switching
- `README.md`, `README.es.md` — keyboard reference rows
- No new dependencies, no filesystem changes
