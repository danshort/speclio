## Context

The layout is described in parallel across three sites that must be kept in sync by hand:

1. **Vertical**: `view.go` (`mainViewContent`/`viewContentWithChrome`/`emptyViewContent`) assembles the chrome rows; `model.go:contentHeight()` independently hand-sums chrome constants (`chromeTop + chromeHeader + …`) to size the viewport. The row-index literals `Y==1`, `Y==2` (mouse.go) and `indexViewportContentStart=3` (index.go) are the same chrome order, hardcoded again.
2. **Tabs**: `mouse.go:handleMouseClick` re-derives tab x-ranges (`x += len(label)+2 +1`, start `x=1`) to mirror `renderTabBar`'s `Padding(0,1)` + `" "` join. Uses `len`, not `lipgloss.Width`.
3. **Index**: `index.go:indexItemAtContentLine` is a ~60-line second walk that replays `renderIndexContent`'s `line++`/`line+=2` increments to map a clicked line back to an item.

Bubble Tea is Elm-style: `View()` returns a string; mouse clicks arrive as separate `Update` messages. So the hit-test cannot read `View()`'s output — the shared truth must be computed when state/size changes and stored on `Model`.

Key observation: `renderIndexContent` **already** computes each item's `line` (it sets `cursorLine = line`). The duplication exists only because that information is discarded and reconstructed.

## Goals / Non-Goals

**Goals:**
- One source of truth for layout geometry, consumed by both rendering and hit-testing.
- Eliminate the second index walk and the re-derived tab geometry; derive viewport height from the chrome-row list.
- A test seatbelt that turns silent drift into a failing test.

**Non-Goals:**
- No user-facing behavior change (same rows, same click targets, same viewport size).
- No general layout-engine abstraction — only enough to share the three couplings.
- No redesign of index rendering, styling, or navigation.

## Decisions

**1. Emit, don't re-derive.** The renderer produces the layout facts as a byproduct and they are consumed for hit-testing:
- `renderTabBar` returns the tab x-ranges it laid out (measured with `lipgloss.Width`).
- `renderIndexContent` captures a line→item map (`map[int]int` from content line to raw item index; absent key = not an item line, via comma-ok) — recording the `line` it already tracks.
- The chrome-row composition is expressed as one ordered list per mode; `View()` renders it and `contentHeight()` derives the viewport height as `terminal height − (chrome rows)`.

*Why over "carefully keep them in sync":* render-position and hit-test become the **same data**, so they cannot disagree by construction. The round-trip property (item rendered at line L ⇒ click at L resolves to that item) holds automatically rather than being a thing tests must police forever.

**2. Store derived layout on `Model`, refreshed in `Update`.** The emitted maps/ranges live on `Model`, produced inside the existing `refresh*Viewport`/render path which already runs on every relevant state change before the next message. Lookups are bounds-checked (return "not found" rather than panic) so a stale/empty map degrades safely.

Note the asymmetry: the index map has a natural `Update`-time producer (`refreshIndexViewport`, called on every index state change), but the tab bar does not — `renderTabBar` runs only inside `View()`. So PR3 must recompute tab x-ranges at an explicit `Update`-time site (resize / `commitStateChange`). This is safe because tab ranges depend only on the constant `tabLabels` and the width, not on per-frame state (the trailing progress bar renders after the tabs and shifts no tab's range).

Empty-project mode is the one view that renders no sized viewport (`emptyViewContent` is fixed multi-line content), so the height invariant explicitly exempts it rather than asserting `Height(View()) == m.height` there.

**3. Stage delivery as a 4-PR stack, riskiest last and isolated.** The index map (coupling 3) is the only change that can silently mis-target clicks, so it ships alone on top of a green test seatbelt.

## Risks / Trade-offs

- [A click-targeting regression is invisible — no panic, no failing test today, just the wrong row selected] → Land the **round-trip + height-invariant tests first (PR1)**, on `main`, so every later stage is guarded. The emit-don't-re-derive design also makes the round-trip hold by construction.
- [Refactor introduces a behavior change under a "no-behavior-change" banner, where review is lighter] → Characterization tests pin current behavior across empty/populated/expanded/filtered/no-match states before any refactor; plus a manual click sweep at 2–3 terminal sizes.
- [Stored layout map goes stale between a state change and the next render] → It is produced on every index state change in `Update` (via `refreshIndexViewport`) before a click can be processed; lookups are bounds-checked and YOffset is composed at click time as today.
- [A chrome row wraps to >1 line and breaks the count-based height] → The `lipgloss.Height(View()) == m.height` invariant test across widths catches it; `emptyViewContent` embeds multi-line text and is handled explicitly.
- [Big blast radius if done at once] → The 4-PR stack keeps each step small, independently revertable, and easy to bisect.

## Migration Plan

Delivered as a stacked set of PRs, merged bottom-up:

1. **PR1 — test seatbelt** (`main`): round-trip render⇄hit-test property + `lipgloss.Height(View())==m.height` across modes/widths. No production change.
2. **PR2 — vertical SSOT**: chrome-row list drives both `View()` and `contentHeight()`; absorbs the `Y==1`/`Y==2`/`indexViewportContentStart` literals.
3. **PR3 — tab x-ranges**: `renderTabBar` emits ranges; click handler looks them up; `len`→`lipgloss.Width`.
4. **PR4 — index line→item map**: `renderIndexContent` emits the map; `indexItemAtContentLine` becomes a lookup. Isolated on top so a regression bisects to one small diff.

Rollback: any PR reverts independently; reverting PR4 restores the manual walk without affecting PR1–3.

**4. Verified: deriving height changes no counts today.** Both `contentHeight` branches already equal the rendered row counts — chrome-only = 6 (`viewContentWithChrome`: top, header, sep, sep, help, bottom), normal = 7 (+1 with spec subnav, matching `mainViewContent`). So PR2 produces the identical viewport height in every mode; this is the load-bearing fact behind "no behavior change," and it holds. Likewise all four tab labels are ASCII (`lipgloss.Width == len`), so PR3's measurement switch is future-proofing, not a behavior change.

**5. Storage shape (resolved).** Separate `Model` fields, not one `layout` struct, so PR2/PR3/PR4 stay independent in the stack: `tabRanges` (PR3) and `indexLineMap` (PR4). The chrome rows need **no** stored field — height is derived inline in `contentHeight` from a row-descriptor function, and `View()` renders from the same function; persisting it would add a refresh site with no consumer.

## Open Questions

- None outstanding.
