## Context

`OpenSpecKit.parseTasks` already yields done/total; the worktree-survey change added a `ChangeProgress` type, sidebar progress bars on worktree-change rows, and the Tasks-top progress bar. This change generalizes that progress display across the app. No new domain surface — pure UI reuse, so the golden corpus is untouched.

## Decisions

- **Compute via `parseTasks(change.tasks.content)`** → reuse `ChangeProgress(done, total)`. One helper for "the change behind the current selection."
- **Resolve the current change from the selection:** `.artifact` → `change(named:)`; `.worktreeArtifact` → `worktreeChange(path,name)`; project spec / config / worktree-metadata / none → no change → no persistent bar. This is the only new model logic.
- **Persistent bar placement:** a thin bar at the very top of the detail pane, above the scrolling content, so it stays put while the artifact scrolls. Rendered by `DetailView` when a current change exists and has tasks; otherwise omitted (no empty bar for task-less changes or non-change views).
- **Per-section progress:** in `TasksView`, walk the parsed items; for each `section` item, sum the `task` items until the next section → that section's done/total, rendered to the right of the heading. Works for both the interactive and read-only (`readOnly`) paths since it's display-only.
- **Sidebar everywhere:** active/archived change nodes get `progress:` set (same `ChangeProgress`), so `SidebarRow`'s existing bar renders for them too — unifying with worktree-change rows.

## Risks / Trade-offs

- **[Low] Cost** — `parseTasks` per change on sidebar build / selection; cheap and already done for worktree changes.
- **[Low] Redundancy** — progress appears in three places (sidebar, persistent bar, Tasks). Intentional per #65; the persistent bar is the always-visible one.
- **[Low] Task-less changes** — show no bar (rather than 0/0); the helper returns nil when there are no tasks.
