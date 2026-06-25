## Why

The TUI describes its layout in two places that must agree by hand: `view.go` assembles the chrome rows, while `model.go:contentHeight()` independently hand-sums chrome constants to size the viewport. Mouse hit-testing re-derives the same geometry a third way — tab x-ranges in `handleMouseClick` mirror `renderTabBar`, and `indexItemAtContentLine` replays `renderIndexContent`'s line increments. Any layout tweak must be mirrored across these copies or the viewport clips/gaps and clicks land on the wrong target — and nothing (compiler or test) catches the drift. This is latent bug debt (issue #35).

## What Changes

- Make the rendered layout the single source of truth: the renderer emits the layout facts (chrome row list, tab x-ranges, index line→item map) and both `View()` and the mouse handlers consume them, instead of re-deriving geometry independently.
- Derive `contentHeight()` (viewport height) from the chrome-row list rather than a parallel hand-sum; the scattered row-index literals (`Y==1`, `Y==2`, `indexViewportContentStart`) fall out of the same source.
- `renderTabBar` emits tab x-ranges; the click handler looks them up. This also removes the latent `len(label)` vs `lipgloss.Width(label)` mismatch.
- `renderIndexContent` emits a line→item map; `indexItemAtContentLine` becomes a bounds-checked lookup instead of a hand-synchronized second walk.
- Add a test seatbelt: a render⇄hit-test round-trip property and a `lipgloss.Height(View()) == m.height` invariant across modes and widths.

## Capabilities

### New Capabilities
<!-- None: this hardens existing behavior; no new user-facing capability. -->

### Modified Capabilities
- `tui-viewer`: add a requirement that the viewport height is derived from the single rendered chrome-row list (not a parallel constant sum), with a total-height invariant.
- `mouse-navigation`: add a requirement that tab and index hit-testing derive from layout data emitted by the renderer, so a click always maps to the item actually rendered at that position.

## Non-goals

- No change to user-facing behavior: the same rows render, the same clicks select the same targets, the same viewport size. This is a pure internal refactor plus tests.
- Not redesigning the index/spec rendering, styling, or navigation model.
- Not introducing a general layout engine; only enough structure to make the three couplings share one source.

## Impact

- Code: `internal/ui/view.go` (row assembly + tab x-range emission), `internal/ui/model.go` (`contentHeight` derivation, layout storage), `internal/ui/mouse.go` (tab/Y/index hit-testing via layout), `internal/ui/index.go` (`renderIndexContent` emits line→item map; `indexItemAtContentLine` becomes a lookup).
- Tests: new round-trip and height-invariant tests in `internal/ui`.
- No dependency, CLI, or data-format changes. Delivered as a 4-PR stack (tests → vertical → tabs → index) to isolate the riskiest step.
