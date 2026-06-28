# Tasks: Undo/Redo for task mutations (macOS)

## 1. Model: snapshot undo engine

- [x] 1.1 Add `AppModel.mutateTasks(changePath:undoManager:actionName:_:)` that snapshots `tasks.md` (as `Data`) before running the passed mutation closure, snapshots after, and — only if the bytes changed — registers a byte-exact undo via the restore helper. Returns the closure's items.
- [x] 1.2 Add private `registerTasksRestore(path:to:expecting:undoManager:actionName:)` and `applyTasksRestore(...)`: the restore writes `target` only when the on-disk file still equals `expecting` (conflict guard), re-registers its mirror (redo), sets the action name, and calls `refreshData(resetSelection:false)`. On conflict, set `taskMutationNotice` and refresh without writing.
- [x] 1.3 Add `@Published var taskMutationNotice: String?` to carry the undo-time conflict message to the view.

## 2. View: route every mutation through the engine

- [x] 2.1 In `TasksView`, add `@Environment(\.undoManager)` and `@EnvironmentObject var model: AppModel`.
- [x] 2.2 Route `toggle(_:)` through `model.mutateTasks(actionName: "Toggle Task")`, preserving its `.fileChanged` notice/refresh UX.
- [x] 2.3 Give the shared `run(_:)` helper an `actionName` and route it through `model.mutateTasks`; pass "Edit Task" / "Add Task" / "Delete Task" / "Move Task" at the call sites (commitEdit, performAdd, performDelete, performMove, performMoveToEnd).
- [x] 2.4 Surface `model.taskMutationNotice` in the Tasks view's notice slot (cleared at the start of the next `mutateTasks`, not on content reload — the conflict path refreshes content right after setting it); keep the read-only worktree path registering no undo.

## 3. Verify

- [x] 3.1 `swift build` (LecternApp) compiles; OpenSpecKit tests + golden corpus stay green; Go untouched.
- [x] 3.2 `openspec validate macos-undo-task-mutations --strict` passes.
- [x] 3.3 Manual QA (user): for each of toggle / inline-edit / add / delete / reorder / move-across-sections — perform it, `⌘Z` reverts on disk and in the view, `⌘⇧Z` re-applies; Edit menu shows the right "Undo <Action>" label; a delete+undo restores number and position exactly; performing a mutation, editing `tasks.md` externally, then `⌘Z` is a guarded no-op with a notice (no clobber).
