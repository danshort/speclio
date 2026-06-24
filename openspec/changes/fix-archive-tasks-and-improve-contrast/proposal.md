## Why

Two defects in the change viewer hurt usability:

- **Archived tasks vanish on arrow navigation (#7).** On an archived change the Tasks tab is read-only markdown, but the down/up arrows drove the interactive task cursor, re-rendering from an unpopulated item list and blanking the view until the user switched tabs and back. This violates the existing `archive-viewer` requirement that `j`/`k` scroll in archive mode.
- **Secondary text is too dim to read (#4).** Help/footer text, the "No active changes" message, requirement count labels, and completed tasks used ANSI color `8` (bright black), which renders at very low contrast on most terminal themes.

## What Changes

- Restrict the interactive tasks cursor (`j`/`k`/`Space`) to `ModeNormal`. In `ModeViewingArchive` the Tasks tab scrolls like every other archived artifact, matching the existing `archive-viewer` spec and the already-correct mouse-wheel path.
- Replace the dim "minimized" text color (ANSI `8`) with a readable mid-gray (256-color `245`) for help text, completed tasks, and done-task code spans. Leave intentionally-muted decoration (borders, disabled tabs, empty progress segments) unchanged.
- Add `DEVELOPING.md` documenting how to run a dev build alongside a Homebrew-installed `speclio`.

## Non-goals

- No change to active-mode task editing, toggling, or cursor behavior.
- No theming system or user-configurable colors; this is a single fixed-palette adjustment.
- No change to which keys exist in archive mode (already specified).

## Capabilities

### New Capabilities

- `ui-text-contrast`: secondary/minimized text SHALL be legible on standard dark terminals while remaining subordinate to primary text.

### Modified Capabilities

- `archive-viewer`: clarify that on the Tasks tab in archive mode, `j`/`k` scroll the viewport and never move a task cursor.

## Impact

- `internal/ui/viewer.go` — guard the Tasks-tab `j`/`k`/`Space` branches with `m.mode == ModeNormal`
- `internal/ui/styles.go` — introduce `dimColor` (256-color `245`); apply to `helpStyle` and `taskDoneStyle`
- `internal/ui/tasks.go` — `doneCodeStyle` uses `dimColor`
- `internal/ui/view_test.go` — regression test for archive-mode Tasks arrows
- `DEVELOPING.md`, `README.md`, `README.es.md` — documentation
- No API, dependency, or filesystem changes
