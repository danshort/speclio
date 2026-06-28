## Context

The macOS app renders `tasks.md` read-only and supports exactly one write: toggling a checkbox. That write is already done safely — `toggleTaskByText` (in `macos/OpenSpecKit/Sources/OpenSpecKit/Tasks.swift`) **re-reads the file**, finds the task **by its text**, and does a **surgical single-line replace** (`- [ ] ` ↔ `- [x] `) so a concurrent edit can't flip the wrong line. The parser (`parseTasks`) is line-regex based (not the swift-markdown AST that rendering uses): it captures `## (.+)` as a section and `- \[ \]/\[x\] (.+)` as a task, leaving the `N.M` number **inside** `TaskItem.text`.

This change generalizes that one safe write into a small family of structured edits (add, delete, reorder, cross-section move, inline text-edit), while keeping the same safety guarantees. The driving insight from exploration: humans don't author OpenSpec changes (agents do) — they review and lightly tweak generated changes, and the tweaks worth doing in-app are *structural*, the things a generic Markdown editor can't do.

Grounding facts established during exploration:
- OpenSpec's canonical `tasks.md` is **flat, single-line, numbered**: `## N. Section` headings with `- [ ] N.M text` tasks (per OpenSpec `AGENTS.md`). An empirical scan of all 87 `tasks.md` in this repo found **0** indented checkboxes and **0** genuinely unnumbered tasks.
- The scan surfaced two real variants the format does contain: **strikethrough** skipped tasks (`- [ ] ~~6.1 …~~ (skipped)`) and **letter-suffixed section prefixes** (`## 3b.`, tasks `3b.1`).
- OpenSpec's apply skill refers to tasks **positionally** ("task 3/7", "N/M complete"), never as durable IDs — and agents re-read files they're working on, so renumbering won't desync them.

## Goals / Non-Goals

**Goals:**
- Add, delete, drag-reorder (within and across sections), and inline text-edit tasks from the macOS tasks view.
- Keep every write surgical, re-read-before-write, and minimal-blast-radius.
- Make renumbering predictable and faithful to OpenSpec convention.
- Fail safe and visibly when the file changed underneath an edit.

**Non-Goals:**
- Prose editing (→ "open artifact" handoff, #105); create/archive/validate (#104); Undo (#103).
- Nested/indented sub-tasks and unnumbered checklists (against the format).
- Section-level operations (add/delete/reorder/renumber sections). Section prefixes are read and preserved, never generated.

## Decisions

### D1 — Numbers are positional (Model A), recomputed per section; prefixes preserved verbatim
A task number is modeled as `<section-prefix>.<ordinal>`. On any structural edit, recompute `ordinal` sequentially (1..n) **only for the affected section(s)**; copy `section-prefix` verbatim from the `##` heading (so `3` stays `3` and `3b` stays `3b`). A cross-section move makes the task adopt the **destination** heading's prefix.
- *Why:* matches what humans expect when dragging a list, and mirrors OpenSpec's own positional framing. Preserving the prefix respects the `3b` insertion convention without us owning section numbering.
- *Alternative considered:* labels-are-identity (Model B) — preserve numbers, renumber minimally. Rejected: produces gaps / out-of-order numbers, and the only motivation (don't confuse agents) was disproved (agents re-read and think positionally).

### D2 — Identity = description with the number stripped, strikethrough-tolerant
Extend the matcher so a task's fingerprint is its description **after** removing a leading `<prefix>.<ordinal>` token and any wrapping `~~…~~`. This is the keystone that makes D1 safe: renumbering changes the visible number but not the fingerprint, so re-read-before-write still finds the right line.
- *Why:* without this, renumbering breaks `findCursorByText`.
- *Trade-off:* two tasks with identical descriptions in the same section are ambiguous — disambiguate by section + nearest-position. Rare in practice.

### D3 — Structured `TaskItem`, still line-based (no AST switch)
Add structured fields to `TaskItem` (`sectionPrefix`, `ordinal`, `done`, `lineNum`, and the bare `description`) parsed out of the line, rather than leaving the number embedded in `text`. Keep the line-regex parser; do **not** move the task subsystem to swift-markdown ranges.
- *Why:* renumbering is fundamentally about line spans and number prefixes — the line model is closer to the problem than an AST. Rendering can keep using swift-markdown independently.

### D4 — Surgical, minimal-blast-radius writes via re-derive-on-current-content
Every edit: re-read `tasks.md` → re-parse → locate the target by D2 fingerprint → compute the new lines for the affected section(s) only → splice those line spans → write (preserving existing line endings, per the current toggle's deliberate CRLF handling). Everything outside the touched section(s), including interspersed prose, is untouched.
- *Why:* smallest possible diff and smallest collision window against a writing agent.

### D5 — Conflict = abort + visible notice
If the pre-write re-read can't locate the target task (D2) or the affected section's structure changed since the edit began, do not write: refresh the view from the current file and surface a visible notice ("file changed on disk"). No attempt to replay the edit.
- *Why:* a reorder is a multi-line transaction; replaying a stale move onto shifted content risks silent corruption. Aborting is the safe, comprehensible behavior. Reuses the feedback pattern from #101.

### D6 — Delete is confirmed; cross-worktree stays read-only
Delete routes through a confirmation step (Undo is #103, not yet available). Editing controls are disabled for foreign-worktree changes, exactly as toggle already is.

### D7 — UI lives in the existing tasks view
Add controls (`+`, `−`, drag handle, double-click-to-edit field) to the macOS tasks view in `ContentView.swift`, driven by new `OpenSpecKit` functions. "Add after selected" inserts relative to the focused task; if nothing is selected (e.g. focus on a heading), fall back to the end of that section.

## Risks / Trade-offs

- **Multi-line renumber diffs are noisier than a toggle** → keep blast radius to the affected section(s) only (D4); the diff is still bounded and local.
- **Free cross-section drag can re-scope a task by accident** (the user accepted free drag) → make drop targets and the destination section visually explicit during the drag so the section change is obvious before release.
- **Ambiguous identical descriptions** (D2) → disambiguate by section + nearest position; accept residual ambiguity for true duplicates.
- **Strikethrough / letter-prefix parsing edge cases** → cover both explicitly in tests (golden + unit) so the number/identity logic is pinned.
- **Concurrent agent writes during a drag** → re-read-before-write + abort (D4/D5); never trust the line index captured at drag start.

## Migration Plan

Additive and macOS-only; no data migration. The existing toggle write path is generalized, not replaced — its re-read-before-write behavior is preserved and reused. Rollback is reverting the change; `tasks.md` files are unaffected when the feature is unused. No impact on the Go TUI.

## Open Questions

- Exact drag affordance for cross-section drops (insertion line vs. highlighted destination section) — a UI detail to settle during implementation; functionally specified by the cross-section requirement.
- Whether inline-edit should also allow toggling done-state from the field, or keep that to the existing checkbox (lean: keep separate).
