## Why

When a change has more than one spec file, the `specs` tab shows a chip row listing them, but the only way to switch specs is to press `3` repeatedly (forward-only, no way back) or click the already-active `specs` tab. This is undiscoverable and one-directional, so users get stuck on the first spec of a multi-spec change — in both active changes (`ModeNormal`) and archived changes (`ModeViewingArchive`). Reported as issue #51.

## What Changes

- Establish a general two-level navigation **convention** for the viewer:
  - **Primary** artifact tabs (`proposal`/`design`/`specs`/`tasks`): `1`–`4`, `Tab`/`Shift+Tab`, `←`/`→` (unchanged).
  - **Secondary** sub-navigation rows: `[` (previous) and `]` (next), wrapping around. The specs chip row is the only sub-nav today; any future sub-nav inherits the same keys.
- Make the specs chip row navigable with `[` / `]` and selectable by left-clicking an individual chip. Applies identically in active and archived changes.
- **BREAKING** (key behavior): remove the `3` key's forward-cycle. `3` becomes a plain primary-tab selector like `1`/`2`/`4` — it selects the `specs` tab and no longer cycles specs.
- `←`/`→` continue to mirror `Tab`/`Shift+Tab` and still do **not** cycle specs (existing rule preserved; the new `[`/`]` keys own the secondary level).

## Capabilities

### New Capabilities
<!-- None: this extends existing navigation capabilities. -->

### Modified Capabilities
- `specs-subnav`: translate the spec to English; replace the "cycle specs with `3`" requirement with `[`/`]` + left-click chip navigation, wrapping around and applying in both active and archived changes.
- `tui-viewer`: add the secondary sub-navigation convention (`[`/`]` move between secondary sub-tabs wherever a sub-nav row is present); update the `3`-key requirement to remove its cycle behavior so it is a plain primary-tab selector; update the keyboard help-bar requirement to advertise `[`/`]` on the specs tab and drop the `3`-cycle.
- `mouse-navigation`: add a "spec chip selection via left-click" requirement, reusing the existing tab-click coordinate mapping (`X=1` past the `│` border, label width including `Padding(0,1)` plus the one-space join).

## Impact

- Code: `internal/ui/viewer.go` (key handling), `internal/ui/mouse.go` (click hit-testing + a new `specRanges()` helper mirroring `tabRanges()`), `internal/ui/view.go` (`renderSpecSubnav`).
- Tests: add seatbelts mirroring `TestTabHitTestRoundTrip` / `TestTabRangesMatchRenderedWidths` for the spec chip row; update any tests asserting the `3`-cycle behavior.
- Docs/help: update the in-app help bar (`renderHelpBar`) and the `?` help overlay (`help.go` Change/Archive viewer groups, plus `help_test.go`) for the `specs` tab; update the key table in `README.md` (remove the `3`-cycle note, add `[`/`]` and chip-click).

## Non-goals

- No modal/focus state and no new `Model` fields — `[`/`]` act directly whenever the `specs` tab is active with more than one spec.
- No change to primary tab navigation keys or to `←`/`→` behavior on other tabs.
- No new sub-nav rows are introduced; this only defines the convention and applies it to the existing specs chip row.
- No visual redesign of the chip row beyond what selection already shows.
