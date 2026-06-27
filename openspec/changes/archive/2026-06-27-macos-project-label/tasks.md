# Tasks

## 1. Sidebar project header
- [x] 1.1 Add a prominent, non-selectable project header (name + `books.vertical` icon) pinned at the top of the sidebar `List`, shown only when a project is open
- [x] 1.2 Keep it visible across all modes and selections (independent of `sidebarNodes`)

## 2. Window title + subtitle
- [x] 2.1 Set the detail pane `.navigationTitle` to the project name
- [x] 2.2 Add a `locationTrail` helper that derives a `·`-separated trail (mode · change/spec · artifact) from `model.mode` and `model.selection`
- [x] 2.3 Set `.navigationSubtitle(locationTrail)`, replacing the previous per-view artifact `.navigationTitle`

## 3. Verify
- [x] 3.1 `swift build` the LecternApp package cleanly
- [x] 3.2 Manual check: project name stays visible across Active/Archived/Specs/Worktrees and with the sidebar collapsed; subtitle reflects the location
