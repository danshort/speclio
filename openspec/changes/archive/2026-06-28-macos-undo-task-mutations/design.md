# Design: Undo/Redo for task mutations via file snapshots

## Context

The interactive Tasks view (`TasksView` in `ContentView.swift`) mutates
`tasks.md` through five operations, all of which already re-read the file and
match tasks by identity (not line index) before writing:

- **toggle** — `toggleTaskByText(path, text)`
- **inline-edit** — `editTaskText(path, identity:, inSection:, newDescription:)`
- **add** — `addTask(path, afterIdentity:, inSection:, description:)`
- **delete** — `deleteTask(path, identity:, inSection:)`
- **move/reorder** — `moveTask(path, identity:, fromSection:, toSection:, toIndex:)`

The view updates its local `items` for immediate feedback; FSEvents
(`AppModel.watcher` on `openspec/`) then fires `refreshData`, which refreshes
the model's `change.tasks.content` and flows back into the view's `content`
prop, re-parsing `items`.

## Decision: one snapshot mechanism, not five semantic inverses

Deriving a per-operation inverse is easy for toggle and edit but fiddly and
error-prone for delete and move, because those are position- and number-
sensitive (re-inserting at the right spot, restoring the original number, etc.).
Instead, **every mutation is wrapped in a byte-exact file snapshot**:

```
before = read(tasks.md)        // Data, byte-exact
op()                           // the existing OpenSpecKit mutation; writes the file
after  = read(tasks.md)
if before != after:
    register undo: restore `before`, expecting current == `after`
```

Restoring the file verbatim reproduces the prior state perfectly — numbers,
ordering, and LF/CRLF endings included — for all five operations with one
implementation. `Data` (not `String`) is used for read/compare/write so the
guard is an exact byte comparison (`String ==` uses Unicode canonical
equivalence, which could falsely match differing bytes).

### Conflict guard — never clobber an external edit

Snapshot undo is coarser than a semantic inverse: it rewrites the whole file. To
preserve the app's existing "don't silently lose edits" contract (#101/#97), the
restore is **guarded**: it only writes if `tasks.md` on disk still equals the
content the mutation produced (`after`). If another process changed the file in
between, the app skips the restore, surfaces a transient notice, and refreshes
from disk — the same conflict UX the forward `run {}` path already uses.

### Redo for free

Restoring is symmetric. The restore handler, when it runs, re-registers its own
mirror (restore `after`, expecting `before`). Registering an undo *inside* an
undo handler is the standard `UndoManager` idiom: during undo the mirror lands
on the redo stack; during redo it lands back on the undo stack. So one recursive
helper yields full undo/redo.

## Implementation

### `AppModel` — the snapshot engine (the durable, class-typed target)

`UndoManager.registerUndo(withTarget:)` needs a class target; `TasksView` is a
struct whose `@State` can't be mutated from an escaping handler. `AppModel`
(a `@MainActor` class the window already owns, and the owner of file/reload
concerns) is the target.

```swift
// Wraps any task mutation: runs it, then registers a byte-exact undo.
@discardableResult
func mutateTasks(changePath: String, undoManager: UndoManager?, actionName: String,
                 _ op: (_ tasksPath: String) throws -> [TaskItem]) rethrows -> [TaskItem] {
    let path = (changePath as NSString).appendingPathComponent("tasks.md")
    let before = try? Data(contentsOf: URL(fileURLWithPath: path))
    let items = try op(path)                                   // throws TaskEditError → no undo registered
    let after = try? Data(contentsOf: URL(fileURLWithPath: path))
    if let before, let after, before != after {
        registerTasksRestore(path: path, to: before, expecting: after,
                             undoManager: undoManager, actionName: actionName)
    }
    return items
}

private func registerTasksRestore(path: String, to target: Data, expecting: Data,
                                  undoManager: UndoManager?, actionName: String) {
    undoManager?.registerUndo(withTarget: self) { model in
        MainActor.assumeIsolated {
            model.applyTasksRestore(path: path, to: target, expecting: expecting,
                                    undoManager: undoManager, actionName: actionName)
        }
    }
    undoManager?.setActionName(actionName)
}

private func applyTasksRestore(path: String, to target: Data, expecting: Data,
                               undoManager: UndoManager?, actionName: String) {
    let current = try? Data(contentsOf: URL(fileURLWithPath: path))
    guard current == expecting else {                          // external edit — don't clobber
        taskMutationNotice = "tasks.md changed on disk — undo skipped to avoid losing edits."
        refreshData(resetSelection: false)
        return
    }
    guard (try? target.write(to: URL(fileURLWithPath: path), options: .atomic)) != nil else { return }
    // Mirror: undoing THIS restore goes from `target` back to `expecting`.
    registerTasksRestore(path: path, to: expecting, expecting: target,
                         undoManager: undoManager, actionName: actionName)
    refreshData(resetSelection: false)                          // flow the restore back to the view
}
```

`refreshData(resetSelection:)` stays `private` — the handler lives inside
`AppModel`. The handler runs on the main thread (the responder chain invokes
`undo()` there), so `MainActor.assumeIsolated` is safe. A new
`@Published var taskMutationNotice: String?` carries the undo-time conflict
message to the view (the forward-path notices stay view-local).

### `TasksView` — route all five mutations through `mutateTasks`

The view gains `@Environment(\.undoManager)` and uses the existing
`@EnvironmentObject var model`. The shared `run {}` helper (already wrapping
edit/add/delete/move with the conflict→notice mapping) is given an
`actionName` and routes through `model.mutateTasks`; `toggle(_:)` does the same
with its own notice text. Each call passes a human action name: "Toggle Task",
"Edit Task", "Add Task", "Delete Task", "Move Task". The read-only worktree
path (`readOnly: true`) never mutates, so it registers nothing.

The forward path keeps its immediate local-`items` update (the op's return
value); the undo path relies on `refreshData` → `content` change →
`.onChange(of: content)` re-parse, the same reload path used everywhere.

## Edge cases

- **External edit between mutation and undo** → conflict guard skips the write,
  notice + refresh. No clobber.
- **Add then inline-rename** → two undo steps (rename, then add). Conventional;
  not coalesced.
- **Read-only worktree tasks** → no mutation, no undo registered.
- **Project switch / window close** → each window has its own `UndoManager` and
  `AppModel`; a reload/reopen starts a fresh undo context, so undo can't reach a
  stale file.

## Testing

`UndoManager` + the SwiftUI environment are AppKit-bound and not headlessly
testable, and the snapshot engine is a thin wrapper over file IO already
exercised by OpenSpecKit's mutation tests. Verification is `swift build` +
manual QA across all five operations (mutate → `⌘Z` reverts on disk and in the
view → `⌘⇧Z` re-applies; Edit-menu labels; external-edit-then-undo is a guarded
no-op).
