## Context

The `specs` tab renders a chip row (the "specs sub-nav") listing a change's spec files. Today the only way to switch specs is pressing `3` repeatedly (forward-only) or clicking the already-active `specs` tab — both undiscoverable and one-directional (issue #51). The viewer already has a clean, tested pattern for tab-bar hit-testing: `tabRanges()` derives each tab's screen-X span from the same styled widths `renderTabBar` lays out, and `TestTabHitTestRoundTrip` / `TestTabRangesMatchRenderedWidths` pin it. The chip row (`renderSpecSubnav`) renders with the identical `Padding(0, 1)` + single-space-join layout but has no hit-test counterpart.

Both active (`ModeNormal`) and archived (`ModeViewingArchive`) changes flow through `current()` and `updateViewer` / `handleMouseClick`, so a single implementation covers both.

## Goals / Non-Goals

**Goals:**
- Bidirectional, discoverable spec navigation that works in active and archived changes.
- A reusable, viewer-wide convention: primary tabs vs. secondary sub-navs, so any future sub-nav inherits the behavior.
- Mouse-clickable spec chips, reusing the proven tab hit-test pattern.

**Non-Goals:**
- No modal/focus state and no new `Model` fields.
- No change to primary tab keys or to `←`/`→` behavior.
- No new sub-nav rows; only the convention plus its application to the existing chip row.

## Decisions

### Modeless `[` / `]` for the secondary level (not modal, not arrows)
A two-level convention: primary artifact tabs use `1`–`4` / `Tab`·`Shift+Tab` / `←`·`→`; secondary sub-navs use `[` / `]`. `[`/`]` act directly whenever the `specs` tab is active with >1 spec.

- **Why not modal (Enter to focus, then arrows):** a focus mode needs new `Model` state, which directly aggravates the cohesion debt tracked in issue #37, and it remaps `h`/`l`/`Tab` while focused — surprising. Two keystrokes to do one thing.
- **Why not reuse `←`/`→`:** `tui-viewer` already specs that the arrows mirror `Tab` and "SHALL NOT cycle spec files." Reusing arrows would overturn a deliberate rule and couple two levels onto one key pair. `[`/`]` leave that rule intact and cleanly separate the levels — and generalize to any future sub-nav.

### Remove the `3`-key forward-cycle
With `[`/`]` and clicks providing real navigation, `3`'s dual behavior is a special case (only `3` cycled; `1`/`2`/`4` did not). Collapse `3` to a plain primary-tab selector. This removes an asymmetry and is the only behavioral subtraction.

### `specRanges()` mirrors `tabRanges()`
Add `specRanges()` computing each chip's inclusive screen-X span from the same styled width `renderSpecSubnav` lays out (`lipgloss.Width`, `Padding(0,1)`, single-space join, start X=1 past the `│`). Hit-testing on the sub-nav row (`chromeRowIndex(rowSubnav)`, which is only present when `hasSpecSubnav()`) maps a click to a spec index. This keeps render and hit-test from drifting, exactly as the tab bar does. A round-trip test (mirroring `TestTabHitTestRoundTrip`) is the seatbelt.

### Convention lives in `tui-viewer`; application in `specs-subnav`
The general "secondary sub-navigation" requirement goes in `tui-viewer` (the overarching viewer spec); the spec-specific behavior (`[`/`]` + click switch specs, wrap-around, archive parity) goes in `specs-subnav`. `mouse-navigation` gets the chip-click hit-test requirement next to the existing tab-click one. The `specs-subnav` spec is translated Spanish→English in this change via `RENAMED` (header) + `MODIFIED` (English body) per OpenSpec's apply order (RENAMED → REMOVED → MODIFIED → ADDED).

## Risks / Trade-offs

- **Discoverability of `[`/`]`** → mitigate with a help-bar hint on the `specs` tab; chips are clickable as the obvious fallback.
- **`←`/`→` mean tab-nav on the specs tab, not spec-nav** (a user may expect arrows to move the visible chip row) → accepted: it preserves the existing `tui-viewer` rule and the level separation; the help bar names `[`/`]`.
- **Removing `3`-cycle is a (minor) breaking key change** → documented as BREAKING in the proposal and via `REMOVED` (Reason + Migration) in the `specs-subnav` delta; muscle-memory cost is low for a personal tool and `3` still reaches the specs tab.
- **`specs-subnav` Purpose line still mentions "cycling … with the 3 key"** → the delta merge only recomposes the Requirements section, so the Purpose paragraph must be corrected by hand when the main spec is updated at archive time (captured as a task).
