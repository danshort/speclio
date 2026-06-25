## 1. Keyboard navigation (`internal/ui/viewer.go`)

- [x] 1.1 Add `[` and `]` handling in `updateViewer`: when `m.tab == TabSpecs` and the current change has more than one spec, decrement/increment `m.specIdx` with wrap-around, clear the specs render cache, and commit the state change. No-op when there is one spec or zero.
- [x] 1.2 Remove the `3`-key forward-cycle: the `case "3"` branch SHALL only select the `specs` tab like `1`/`2`/`4`, with no cycling when already on `specs`.
- [x] 1.3 Make the selected spec persist within a change: remove the `m.specIdx = 0` reset when entering the specs tab via `3` (and the tab-click path in §2.3). `specIdx` resets only on a change-identity switch (`h`/`l`, worktree open, and the index/archive open path `activateIndexItem`) and the out-of-bounds reload clamp. Returning to `specs` via `Tab`/`3`/click within the same change preserves the last-viewed spec.
- [x] 1.4 Confirm `←`/`→` keep mirroring `Tab`/`Shift+Tab` and do not touch `specIdx`.

## 2. Mouse navigation (`internal/ui/mouse.go`, `internal/ui/view.go`)

- [x] 2.1 Add `specRanges()` (in view.go next to `tabRanges()`): return each chip's inclusive screen-X span derived from the same styled width `renderSpecSubnav` lays out (`lipgloss.Width`, start X=1, +1 per single-space join).
- [x] 2.2 In `handleMouseClick`, when the click Y equals `chromeRowIndex(rowSubnav)` (present only when `hasSpecSubnav()`), map X via `specRanges()` to a spec index, set `m.specIdx`, clear the specs cache, and reload the viewport. Ignore clicks between/outside chips.
- [x] 2.3 Remove the spec-cycle special case from the tab-bar click handler (clicking the active `specs` tab now just reloads, consistent with other tabs, and does not reset `specIdx`).

## 3. Help text & docs

- [x] 3.1 Update the help bar for the `specs` tab (active and archive variants in `renderHelpBar`) to advertise `[`/`]` for spec navigation; drop any reference to `3` cycling.
- [x] 3.2 Update the `?` help overlay in `help.go`: add a `{"[ / ]", "previous/next spec"}` entry to the "Change viewer" and "Archive viewer" groups; update `help_test.go` (`TestHelpOverlayContent`) accordingly.
- [x] 3.3 Update the key table in `README.md`: change the `3` row (remove "press again to cycle"), and add rows for `[`/`]` (previous/next spec) and clicking a spec chip.

## 4. Tests (`internal/ui/*_test.go`)

- [x] 4.1 Add a `specRanges` round-trip test mirroring `TestTabHitTestRoundTrip` / `TestTabRangesMatchRenderedWidths`: a click on a chip's rendered position selects that spec; ranges match rendered widths.
- [x] 4.2 Add key-nav tests: `]` / `[` move `specIdx` with wrap-around; single-spec is a no-op; `←`/`→` do not change `specIdx`; `3` selects the specs tab without cycling; selected spec is preserved when leaving and returning to the specs tab within a change and resets on change switch.
- [x] 4.3 Add coverage that navigation and chip-click work in `ModeViewingArchive` — covering both a true archived change and a foreign-worktree change (`viewingWorktreeChange`), not just `ModeNormal`.
- [x] 4.4 Update/remove any existing tests that assert the old `3`-cycle behavior.

## 5. Spec & docs cleanup

- [x] 5.1 At archive time (or when updating the main spec), fix the `specs-subnav` spec `## Purpose` paragraph, which still says specs are cycled "with the 3 key".
- [x] 5.2 Run `gofmt`, `go vet`, `golangci-lint`, and `go test ./...`; verify `openspec validate spec-subnav-navigation --strict` stays green.
