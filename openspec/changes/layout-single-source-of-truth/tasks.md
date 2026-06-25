## 1. Test seatbelt (PR1, on main)

- [x] 1.1 Add a `lipgloss.Height(View().Content) == m.height` invariant test across the viewport-backed modes (normal-with-changes, archive, index, spec, config, worktrees) at several widths (incl. narrow), with the help overlay closed, asserted only for heights >= the mode's chrome-row count. Add a boundary case at the threshold and one just below it documenting the `contentHeight` clamp (`h < 1 → 1`, which intentionally breaks the equality). Add a separate assertion that the empty-project welcome view renders fixed content (the invariant does not apply there).
- [x] 1.2 Add a render⇄hit-test round-trip test for the index that drives state through the production path (`buildIndexItems` → `applyFilter` → `refreshIndexViewport`) **before** hit-testing, so the test body is identical pre- and post-PR4: for every visible item rendered at content line L, assert `indexItemAtContentLine(L)` returns that item, across empty / active-only / active+specs+archives / expanded-specs / expanded-archives / active-filter / no-match / sorted-by-suffix (`SortBySuffix`) states. Also bring the existing `TestClickArchivedArtifact` onto this render-first footing — it currently hit-tests without rendering and would break under PR4.
- [x] 1.3 Add a tab round-trip test: for every available tab, a click at its rendered x-range resolves to that tab.
- [x] 1.4 Run `gofmt`, `go vet`, full `go test ./...`; confirm the new tests pass against current behavior (they characterize the status quo).

## 2. Vertical SSOT — chrome rows drive height (PR2, on PR1)

- [x] 2.1 Express the chrome rows as one descriptor keyed by chrome *shape* (with-tab-bar + optional spec subnav for normal/archive; the without-tab-bar shape that backs all four `viewContentWithChrome` modes — config/index/spec/worktrees; and the fixed empty-project view), rendered by `View()`. Note one descriptor backs the four viewport-with-chrome modes — they differ only in header/help text, not row count.
- [x] 2.2 Derive `contentHeight()` from that chrome-row list (terminal height minus chrome rows, incl. the optional spec subnav) instead of the hand-summed `chrome*` constants.
- [x] 2.3 Replace the row-index literals `Y==1`, `Y==2` (mouse.go) and `indexViewportContentStart` (index.go) with values derived from the same chrome-row source.
- [x] 2.4 Confirm the height-invariant test (1.1) still passes; `gofmt`/`vet`/`go test ./...` clean.

## 3. Tab x-ranges emitted by the renderer (PR3, on PR2)

- [ ] 3.1 Have `renderTabBar` compute each tab's x-range (start/end) over the `tabCount` label segments only (excluding the trailing progress bar), using `lipgloss.Width`, and store the ranges on `Model`. Because ranges depend only on the constant `tabLabels` and the terminal width — not on `m.tab`, active/disabled styling, or archive vs normal — recomputing on resize (the `WindowSizeMsg` handler) is sufficient; a tab switch does not move them, so no per-state-change refresh is needed.
- [ ] 3.2 Replace the re-derived geometry in `handleMouseClick` (`len(label)+2`, `+1`, `x=1`) with a lookup over the emitted ranges; preserve disabled-tab and spec-subnav-cycle behavior.
- [ ] 3.3 Confirm the tab round-trip test (1.3) still passes. Current labels are ASCII, so `lipgloss.Width == len` today and there is no behavior change; add a synthetic wide-label unit test for the range math to guard future non-ASCII labels. `gofmt`/`vet`/`go test ./...` clean.

## 4. Index line→item map (PR4, on PR3)

- [ ] 4.1 Have `renderIndexContent` emit a line→item map (capturing the `line` it already tracks per item) and store it on `Model` via the existing `refreshIndexViewport` path.
- [ ] 4.2 Rewrite `indexItemAtContentLine` as a bounds-checked lookup into the emitted map; delete the hand-synchronized second walk.
- [ ] 4.3 Confirm the round-trip test (1.2) passes with no edits to its body — it already renders via `refreshIndexViewport` before hit-testing, so PR4 introduces the map without touching test code (preserving the seatbelt). Verify YOffset composition and stale/empty-map safety (lookup returns not-found, never panics); `gofmt`/`vet`/`go test ./...` clean.

## 5. Verify and validate

- [ ] 5.1 Manual click sweep: launch the TUI at 2–3 terminal sizes, click each tab and several index rows (including expanded/filtered), resize, and confirm no clipping or mis-targeting.
- [ ] 5.2 Run `openspec validate layout-single-source-of-truth` and resolve any issues.
