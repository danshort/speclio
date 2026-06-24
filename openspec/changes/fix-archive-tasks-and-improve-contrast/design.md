## Context

`updateViewer` in `internal/ui/viewer.go` handles key input for both `ModeNormal` and `ModeViewingArchive`. For the Tasks tab it routed `j`/`k` to `moveCursorDown`/`moveCursorUp` + `refreshTasksViewport`, and `Space` to `doToggle`. But `loadViewport` only uses the interactive task renderer (`loadViewportForTasks`) in `ModeNormal`; in archive mode the Tasks tab is rendered as read-only markdown via `loadViewportForArtifact`. So in archive mode `refreshTasksViewport` re-rendered from `m.tasks.Items` — which is never populated for an archived change — producing an empty view. `mouse.go` already guarded the cursor with `m.mode == ModeNormal`; the keyboard path did not. The `archive-viewer` spec already requires `j`/`k` to scroll and `Space` to be ignored in archive mode, so this is a regression against an existing requirement rather than a new behavior.

Separately, the "minimized" text style used ANSI color `8` (bright black), which many terminal themes render at very low contrast against the background.

## Goals / Non-Goals

**Goals:**
- Archive-mode Tasks tab scrolls with `j`/`k` and ignores `Space`, consistent with `archive-viewer` and with `mouse.go`.
- Secondary text is legible while remaining visually subordinate to primary text.

**Non-Goals:**
- No configurable theming.
- No change to normal-mode task interaction.
- No restyling of decorative chrome (borders, disabled tabs, empty progress segments).

## Decisions

- **Gate interactive task actions on `ModeNormal`.** Add `&& m.mode == ModeNormal` to the Tasks-tab branches for `j`/`k`/`Space` in `updateViewer`. The existing `else` branch already scrolls the viewport, so archive mode falls through to scrolling with no new code. This mirrors the guard already present in `mouse.go`.
- **Introduce a single `dimColor` constant (256-color `245`).** One source of truth for "minimized" text. ANSI `8` → `245` lifts contrast to a readable mid-gray that is still dimmer than primary text (`7`/`15`), preserving hierarchy (done `245` < pending `7`).
- **Limit the color change to text.** Borders (`separatorStyle`), disabled tabs (`tabDisabledStyle`), and empty progress segments (`progressEmptyStyle`) keep ANSI `8` — their muting is intentional and they are not text to read.

## Risks / Trade-offs

- **[Low] Fixed 256-color value ignores terminal theme.** `245` is chosen to read on both dark and most light backgrounds; a full theming system is out of scope. If a user's theme makes it uncomfortable, that is a follow-up.
- **[Low] Behavior change for the archive Tasks tab.** Users who relied on cursor movement in archive mode now scroll instead — but that cursor movement was the bug that blanked the view, so this is strictly an improvement and aligns with the spec.
