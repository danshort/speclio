## 1. Fix

- [x] 1.1 Add `delete(m.renderCache, m.tab)` in the `editorReturnMsg` handler in `internal/ui/update.go` before `m.loadViewport()`

## 2. Verification

- [x] 2.1 Build and run the application (`make build && ./dossier`)
- [x] 2.2 Open an artifact, press `e` to edit, exit without saving, and verify the viewport is correctly rendered
- [x] 2.3 Repeat with an artifact that has pending changes, save in the editor, and verify the updated content is rendered
