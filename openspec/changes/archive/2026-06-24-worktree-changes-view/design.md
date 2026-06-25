## Context

`lectern` is single-root today. `main.go` resolves `cwd`, calls `openspec.LoadFrom(cwd)` once, and builds a `ui.Model` holding exactly one `root string`, one `*openspec.Project`, and one `loader`. The 500 ms tick (`handleTick`) reloads that single root.

Two existing pieces make a cross-worktree view largely a matter of reuse:

- **The loader is already root-agnostic.** `Loader.LoadFrom(root)` loads the active changes of any directory; only `main.go` and the zero-arg wrappers are bound to `cwd`. Loading another worktree is just another `LoadFrom(path)`.
- **A read-only "foreign change" viewer already exists.** Archived changes live in their own slice (`indexState.ArchiveChanges`), separate from `project.Changes`, and are viewed via `ModeViewingArchive`. Every mutation is gated on `m.mode == ModeNormal` (`viewer.go` `space`/`j`/`k`, `viewport.go` task rendering, `tasks.go` `doToggle`), so that mode is inherently read-only. A worktree's change is structurally identical: an externally-loaded `openspec.Change` rendered read-only.

The progress bar (`taskCounts` → `renderActiveItem` in `index.go`) is already factored and reusable for any `Change`.

"Worktrees attached to the current project" has an authoritative definition: `git worktree list --porcelain` enumerates every worktree sharing this repository's git dir, from any worktree, with each entry's path, HEAD, and branch (or `detached`/`bare`/`locked`).

## Goals / Non-Goals

**Goals:**
- A separate, on-demand `ModeWorktrees` opened with `w` from the index.
- Discover the repo's git worktrees and list each one's active changes with task progress.
- `Enter` opens a foreign change read-only, reusing the `ModeViewingArchive` path; `e` opens in `$EDITOR`; `esc` returns.
- Poll only while the view (and an open foreign change) is active.

**Non-Goals:**
- Integrating cross-worktree changes into the index, or showing live cross-worktree counts there.
- Editing foreign worktrees in place (toggling tasks / writing artifacts).
- Watching arbitrary non-worktree directories.

## Decisions

### Decision: Separate `ModeWorktrees` view, not index integration
Add a new `Mode` and a dedicated render/key path rather than folding cross-worktree changes into the index. The index's flat `[]indexItem` model with hand-counted mouse hit-testing (`indexItemAtContentLine`), filtering, and sorting is fragile; a separate view isolates all new rendering and keeps the index untouched apart from one helpbar entry.

- *Alternative considered:* an "Other Worktrees" section inside the index (the issue's "all on the index" ideal). Rejected for v1 — higher blast radius (two-level nesting, hit-testing, filter/sort across a fourth group) for a feature that reads better as an explicit lens.

### Decision: Foreign changes are read-only, reusing `ModeViewingArchive`
`Enter` on a foreign change loads it (if not already loaded) and enters the existing read-only archive viewing path with the appropriate tab. No new read-only plumbing is needed because mutation is already gated on `ModeNormal`.

Beyond simplicity, this is a safety property: the worktrees being inspected are actively written by agents. Toggling their `tasks.md` from lectern would create a second concurrent writer and risk lost writes. The invariant is: **lectern only mutates the worktree it is rooted in.** `e` (open in `$EDITOR`) remains available as a user-driven, single-file escape hatch.

- *Alternative considered:* allow toggling foreign tasks. Rejected — two-writer hazard against live agent sessions.
- *Future, out of scope:* an `o` action that re-roots lectern onto the selected worktree for a full editable `ModeNormal` session.

### Decision: Polling only while the view is active
On entering `ModeWorktrees`, shell out to `git worktree list --porcelain` once and `LoadFrom` each worktree to build the snapshot. While the view (or a foreign change opened from it) is the active mode, the existing tick reloads change content so progress bars track agents live. On returning to the index, polling stops entirely; the normal index/viewer path is unchanged and pays nothing. The worktree *list* itself is captured once on entry (its membership changes rarely) and is not re-shelled on every tick.

### Decision: Static `w` affordance on the index, no dynamic count footer
The index advertises the view with a static `w` helpbar entry only. A dynamic footer (e.g. "3 worktrees · 4 active changes") would require polling worktrees while sitting on the index, which contradicts the polling decision above. The trade-off — no live cross-worktree counts until you press `w` — is accepted deliberately.

### Decision: Identity and edge-case rendering
- Label each worktree by **branch name**; fall back to a short HEAD SHA when detached.
- List the current worktree first, badged `(current)`; it is shown for completeness but its changes are the normal index, so it is not openable here.
- Skip `bare` worktrees; annotate `locked`/`prunable`.
- A worktree with no `openspec/` directory or no active changes renders as present-but-empty (`(no active changes)`), never an error.
- If `git` is unavailable or the project is not inside a git working tree, the view renders a single explanatory line instead of failing.

### Decision: Discovery lives in `internal/openspec`
Add a small helper (e.g. `ListWorktrees`) that runs `git worktree list --porcelain` and parses path/branch/HEAD/flags into a struct. Keeping it beside the loader lets the UI layer stay declarative and makes the parser unit-testable against canned porcelain output.

## Risks / Trade-offs

- [`git` subprocess cost] → Shell out once per view entry, not per tick; cache the parsed list. Content reloads reuse the existing loader.
- [Foreign change loaded from a path not in `project.Changes`] → Store worktree-loaded changes in their own slice (mirroring `ArchiveChanges`) and view via `ModeViewingArchive`, so no assumptions about `project.Changes` indices are violated.
- [Stale worktree list within a session] → Re-shell `git worktree list` on each entry to the view; mid-session membership drift is acceptable and self-corrects on re-entry.
- [No live counts on the index] → Accepted consequence of the polling decision; revisit only if an index summary proves worth a lightweight cached poll.

## Open Questions

None blocking. The `o` re-root flow and any index summary footer are explicitly deferred.
