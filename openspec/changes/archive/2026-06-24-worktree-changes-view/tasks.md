## 1. Worktree discovery (internal/openspec)

- [x] 1.1 Add a `Worktree` struct (path, branch, head SHA, flags: current/detached/bare/locked/prunable) and a `ListWorktrees` helper that runs `git worktree list --porcelain` and parses its output
- [x] 1.2 Detect the current worktree (match against `cwd`/the model root) and mark it
- [x] 1.3 Handle absence gracefully: `git` not on `PATH` or not inside a git working tree returns a sentinel/empty result rather than an error the UI cannot render
- [x] 1.4 Unit-test the parser against canned porcelain output covering branch, detached HEAD, bare, locked, and prunable entries

## 2. Model and state (internal/ui)

- [x] 2.1 Add `ModeWorktrees` to the `Mode` enum in `internal/ui/model.go`
- [x] 2.2 Add worktrees view state (the discovered worktrees, each worktree's loaded `[]openspec.Change`, a cursor, and an availability/error message) — store foreign changes in their own slice, mirroring `ArchiveChanges`
- [x] 2.3 Add an enter helper that shells out once via `ListWorktrees`, calls `LoadFrom` per worktree, builds the flattened navigable item list, and sets `ModeWorktrees`

## 3. Rendering (internal/ui/worktrees.go)

- [x] 3.1 Render worktrees grouped: header per worktree with branch (or short SHA when detached) and `(current)` badge on the current worktree, listed first
- [x] 3.2 Render each active change nested under its worktree with the existing progress bar (reuse `taskCounts`/`renderActiveItem`)
- [x] 3.3 Render empty worktrees as `(no active changes)`; skip bare worktrees; annotate locked/prunable
- [x] 3.4 Render an explanatory single line when worktrees are unavailable (no git / not a working tree)
- [x] 3.5 Highlight the item under the cursor using the existing cursor style

## 4. Navigation and read-only viewing

- [x] 4.1 Add the `w` handler in `updateIndex()` to switch to `ModeWorktrees`
- [x] 4.2 Add `j`/`k` navigation and `esc` (return to index) in a `updateWorktrees()` handler; route it from `dispatchKey`
- [x] 4.3 `Enter` on a foreign change opens it read-only via the `ModeViewingArchive` path (load the change if needed, set the change + first available tab), confirming task toggling and edits remain gated off
- [x] 4.4 Confirm `e` (open in `$EDITOR`) works for a foreign change

## 5. Polling

- [x] 5.1 Gate cross-worktree polling to `ModeWorktrees` (and a foreign change opened from it): reload change content on the existing tick while active, and do nothing when outside the view
- [x] 5.2 Capture the worktree list once on entry (not per tick); re-shell only on re-entry

## 6. Index affordance

- [x] 6.1 Add a static `w` entry to the index helpbar; confirm no dynamic cross-worktree counts are computed on the index

## 7. Tests

- [x] 7.1 Test the view lists worktrees with the current one first and badged, including an empty worktree
- [x] 7.2 Test a foreign change renders its progress bar from its loaded tasks
- [x] 7.3 Test `Enter` opens a foreign change read-only and that toggling/edits are not possible in that mode
- [x] 7.4 Test the unavailable-git path renders the explanatory line without panicking
- [x] 7.5 Run `gofmt`/`go vet` and the full `go test ./...` suite

## 8. Documentation

- [x] 8.1 Update the README to document the worktrees view: the `w` affordance from the index, navigation, read-only foreign-change viewing, and the `git` requirement

## 9. Verification

- [x] 9.1 Run `openspec validate worktree-changes-view --strict` and confirm the change is valid
- [x] 9.2 Manually verify against the live sibling worktrees: run lectern, press `w`, confirm other worktrees' active changes and progress appear and are read-only
