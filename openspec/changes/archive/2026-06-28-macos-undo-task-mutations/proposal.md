# Add Undo/Redo for task mutations (macOS)

## Why

Task changes in the macOS app are written to `tasks.md` immediately, with no
Undo. On a native Mac app, `⌘Z` is a baseline expectation — a user who toggles
the wrong checkbox, mis-edits a task, deletes the wrong one, or drops a drag in
the wrong place has no way back except to redo the work by hand. This is one of
the "it doesn't feel native" gaps from the product review (#103).

Rather than hand-derive a semantic inverse for each of the five mutating
operations (toggle, inline-edit, add, delete, reorder/move) — where delete and
move are position- and number-sensitive and easy to get subtly wrong — the app
takes a **file-snapshot** approach: capture `tasks.md` before a mutation and
register an undo that restores those exact bytes. One mechanism covers every
operation with perfect fidelity (numbers, position, line endings), and undo is
guarded so it never clobbers an external edit made in the meantime.

## What Changes

- Every task mutation in the interactive Tasks view (toggle, inline-edit, add,
  delete, reorder, move-across-sections) registers an undo with the window's
  `UndoManager`. `⌘Z` restores the pre-mutation `tasks.md`; `⌘⇧Z` re-applies it.
  The Edit menu names the action ("Undo Toggle Task", "Undo Edit Task", etc.).
- Undo/redo are implemented as byte-exact file snapshots, so they restore the
  file precisely — including task numbers, ordering, and LF/CRLF endings.
- Undo/redo are **conflict-guarded**: before restoring, the app checks that
  `tasks.md` still matches what the mutation produced. If another process
  changed the file in between, the app skips the restore (no clobber), shows a
  transient notice, and refreshes from disk.
- Scope: the interactive Tasks view in the main project window. Read-only
  worktree task views remain non-interactive and register no undo.

## Out of scope

- Undo for changes outside the Tasks view (config edits, etc.).
- Cross-window or cross-launch undo — each window owns its own `UndoManager`;
  reloading or reopening a project starts a fresh undo context.
- Coalescing an "add then inline-rename" into a single undo step — these remain
  two steps (the add, then the text edit), which is conventional.
