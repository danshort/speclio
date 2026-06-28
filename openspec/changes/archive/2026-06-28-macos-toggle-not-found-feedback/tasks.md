## 1. Engine

- [x] 1.1 `toggleTaskByText` (`OpenSpecKit/Tasks.swift`): throw `TaskEditError.fileChanged` when no task matches `text`, instead of returning the unchanged list
- [x] 1.2 Update `ToggleTests` "unknown task" case to expect the thrown `.fileChanged` and an unchanged file

## 2. UI

- [x] 2.1 `toggle` (`ContentView.swift`): catch `TaskEditError.fileChanged` → transient notice ("Couldn't find that task — it may have changed on disk; refreshed.") + `refreshFromDisk()`; keep the generic write-error message

## 3. Verification

- [x] 3.1 `swift test` (OpenSpecKit) + golden green; `swift build` (LecternApp) clean
- [ ] 3.2 Manual: externally delete a task line, then click its (stale) checkbox — confirm the notice appears and the list refreshes
