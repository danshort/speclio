## Why

Developers using agentic coders often run several git worktrees of the same project at once, each on its own branch with its own in-flight OpenSpec change. Running `lectern` from one worktree shows only that worktree's active changes — work happening in sibling worktrees is invisible, so there is no single place to see what every agent is working on or how far along each change is. Issue #25 asks to surface the active changes (and task progress) of every worktree attached to the current repository.

## What Changes

- Add a new read-only **worktrees view** (`ModeWorktrees`), opened with `w` from `ModeIndex`.
- On entry, discover every git worktree of the current repository via `git worktree list` and load each worktree's active changes through the existing loader.
- Render the worktrees grouped: each worktree shows its branch (or short SHA when detached) and its nested active changes, each with the existing done/total progress bar. The current worktree is listed first and badged `(current)`; worktrees with no active changes render as empty.
- `j`/`k` navigate; `Enter` on a foreign change opens it **read-only**, reusing the `ModeViewingArchive` viewing path (no task toggling or in-place edits). `e` opens the artifact in `$EDITOR`. `esc` returns to the index.
- Poll only while the worktrees view (and an open foreign change reached from it) is active: discover the worktree list once on entry and refresh change content on the existing 500 ms tick so progress tracks agents live. Outside the view, nothing is polled.
- Add a static `w` entry to the index helpbar.

## Capabilities

### New Capabilities
- `worktrees-view`: A read-only view listing the current repository's git worktrees, their active changes, and per-change task progress, with navigation into a read-only artifact viewer for a foreign change.

### Modified Capabilities
- `change-index`: Add a requirement for opening the worktrees view from the index with `w`, including the helpbar affordance.

## Non-goals

- No integration of cross-worktree changes into the index itself (no extra index section, no dynamic count footer). The worktrees view is a separate, on-demand lens.
- No editing of foreign worktrees from this view: task checkboxes are not toggleable and artifacts are not written in place. Re-rooting lectern onto another worktree for a full editable session is a possible future addition, not part of this change.
- No background polling while outside the worktrees view; the index does not show live cross-worktree counts.
- No support for watching arbitrary configured project directories — scope is git worktrees of the current repository only.
- No changes to archived-change or specification handling.

## Impact

- Affected code: a new `internal/ui/worktrees.go` (discovery, view state, render, key handling); `internal/ui/model.go` (new `ModeWorktrees`, worktrees view state); `internal/ui/update.go` (mode dispatch, tick polling gated to the view); `internal/ui/viewer.go`/index helpbar text (the `w` affordance); `internal/openspec` (a small `git worktree list` helper).
- Affected specs: new `openspec/specs/worktrees-view/spec.md`; modified `openspec/specs/change-index/spec.md`.
- Tests: new UI tests for discovery parsing, rendering, navigation, and read-only enforcement.
- External dependency: requires the `git` binary on `PATH` and the project to be inside a git working tree; absence is handled gracefully (the view reports that worktrees are unavailable rather than erroring).
- No data-format or external API changes.
