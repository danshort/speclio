# Tasks

## 1. Per-window model
- [x] 1.1 Add a `Codable & Hashable` `ProjectRef` (stable id = folder path)
- [x] 1.2 Convert the scene to a value-based `WindowGroup(for: ProjectRef.self)` with a new `RootView` that owns `@StateObject AppModel` and loads its `ref` on appear
- [x] 1.3 Move project state out of any app-level singleton so each window has an independent `AppModel` (mode, selection, sidebar, watcher, security scope)

## 2. Bookmarks keyed by project
- [x] 2.1 Replace the single `projectBookmark` key with a `path → bookmark data` store in `UserDefaults` (add/resolve helpers)
- [x] 2.2 On open: create+persist the bookmark, then open/route to a window for that `ProjectRef`
- [x] 2.3 On restore: resolve a window's bookmark by path, start the security scope, load; refresh stale bookmarks; show the empty/error state when unresolvable
- [x] 2.4 Balance security scope + watcher teardown on window close

## 3. Open semantics
- [x] 3.1 "Open Project…" opens via `openWindow(value: ProjectRef)` — focuses the existing window when the project is already open (single instance), else creates one
- [x] 3.2 When Open is invoked from an empty (project-less) window, load the project in place into that window (no stray empty window); other paths use `openWindow`
- [x] 3.3 Keep automatic window tabbing enabled so New Tab / the Tab Bar opens empty windows that ⌘O fills (fixes the dead menu item)

## 3b. Open Recent
- [x] 3b.1 Maintain an ordered, de-duplicated, capped recents list of project paths in `UserDefaults`, updated on every successful open
- [x] 3b.2 Add a File ▸ Open Recent submenu whose entries `openWindow(value: ProjectRef)`, plus a Clear Menu action that empties the list without touching open windows

## 4. Command routing
- [x] 4.1 Expose the focused window's `AppModel` via `FocusedValues`, set from `RootView`
- [x] 4.2 Route app-level commands (Open) through `openWindow`; confirm Reload / file actions remain per-window via the toolbar; text-size stays global `@AppStorage`

## 5. Verify
- [x] 5.1 `swift build` the LecternApp package cleanly
- [x] 5.2 Manual: open two projects (separate windows + a tab); independent mode/selection; Reload affects only the focused window; text size applies to both
- [x] 5.3 Manual: re-opening an already-open project focuses its window (no duplicate); Open Recent lists projects, reopens/focuses, and Clear Menu empties the list
- [x] 5.4 Manual: quit with multiple projects open and relaunch → all reopen; a moved/deleted project shows the empty/error state without blocking launch
