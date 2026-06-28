## Why

Today the macOS Lectern app is read-only except for toggling task checkboxes; every other edit forces a round-trip out to an external editor (open flyout → switch app → edit → save → switch back). OpenSpec changes are authored by agents and *reviewed and lightly tweaked* by humans — and the most common tweaks (prune an over-generated task, fix its wording, reorder, add a missing one) are exactly the structured operations that are slow and error-prone by hand in a text editor but trivial as in-app controls. This is the structured editing a dedicated Markdown app fundamentally cannot do, because it does not understand the OpenSpec task model.

## What Changes

- Add **structured task editing** to the macOS tasks view, alongside the existing toggle:
  - **Add** a task — inserts after the currently selected task and gets the next ordinal.
  - **Delete** a task — with a confirm step (no Undo until #103).
  - **Drag-reorder** a task — free movement, including across sections.
  - **Inline-edit** a task's text — double-click to an editable field, save in place.
- Task numbers are **positional (Model A)**: on any structural edit, recompute the task **ordinal** sequentially and **preserve the section prefix verbatim** (e.g. `3b`). A cross-section drag adopts the **destination** section's prefix and renumbers the ordinals of both the source and destination sections.
- All writes are **surgical line-span splices** with the minimal blast radius (only the affected section(s)), **re-reading the file before every write** so a concurrent agent edit can't corrupt the file. If the structure moved underneath the edit, **abort and show a visible notice** rather than write a stale change.
- Task **identity** for the safe re-read matches on the description text with the leading number **stripped**, so renumbering never breaks the match. Matching must tolerate `~~strikethrough~~` (skipped) tasks.

## Capabilities

### New Capabilities
- `macos-task-editing`: Structured editing of `tasks.md` from the macOS app's tasks view — add, delete, drag-reorder (within and across sections), and inline text-edit of tasks, with positional renumbering, surgical re-read-before-write persistence, and conflict-abort semantics.

### Modified Capabilities
<!-- None. The existing `tasks-toggle` capability is TUI-specific (Bubble Tea); this change adds a new macOS capability rather than altering toggle's requirements. -->

## Impact

- **Code:** `macos/OpenSpecKit/Sources/OpenSpecKit/Tasks.swift` (generalize the existing surgical toggle into add/delete/reorder/edit + a positional renumber routine; refine identity to strip the number and tolerate strikethrough), `macos/OpenSpecKit/Sources/OpenSpecKit/Models.swift` (richer `TaskItem`: section prefix + ordinal as structured fields), and the macOS tasks view in `macos/LecternApp/Sources/LecternApp/ContentView.swift` (controls, drag, inline field, confirm dialog, conflict notice).
- **Tests:** `macos/OpenSpecKit/Tests/OpenSpecKitTests/` — extend the golden/toggle tests to cover renumbering, cross-section moves, strikethrough, and conflict-abort.
- **Behavior:** Cross-worktree (foreign) changes remain **read-only** — editing controls are disabled there, as toggle already is.
- **No change** to the Go TUI; this is macOS-only.

## Non-goals

- **Prose/paragraph editing** of any artifact (proposal/design/spec body, or multi-line task notes) — that is handled by the separate "open artifact" change (#105), which opens the file in the user's default Markdown app with live reload.
- **Create / archive / validate** a change — tracked in #104.
- **Undo** for edits — tracked in #103; until then, delete is guarded by a confirm step.
- **Nested / indented sub-tasks** and **unnumbered checklists** — both are against the OpenSpec `tasks.md` format (confirmed: the canonical format is flat, single-line, `## N.` sections with `- [ ] N.M` tasks; an empirical scan of all 87 `tasks.md` in this repo found zero of either). Editing operates on flat, numbered tasks only.
- **Section-level operations** (add/delete/reorder/renumber sections, including the `3b` insertion convention) — out of scope; section prefixes are read and preserved, never generated or reordered.
