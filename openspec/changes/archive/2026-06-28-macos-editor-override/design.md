## Context

`ContentView.openInEditor()` calls `NSWorkspace.shared.open(url)` (system default), reachable only from a toolbar flyout — no app choice, no shortcut. The app is non-sandboxed (uses security-scoped bookmarks for project folders), so it can launch an arbitrary chosen app via `NSWorkspace` without extra entitlements. Settings (`GeneralSettingsView`) currently holds only the font-scale preference; File-menu commands (`ProjectCommands`) reach the focused window through `@FocusedObject AppModel`.

## Goals / Non-Goals

**Goals:** pick a specific editor app (persisted), open the current artifact in it (else default), and a `⌘E` shortcut + menu item. Graceful fallback when unset or the app is gone.

**Non-Goals:** broader Settings expansion (#100); a shell-command editor (TUI's model); sandboxing.

## Decisions

### D1 — Store the chosen app's path in `@AppStorage`
`EditorPref.storageKey = "editorAppPath"`, `""` meaning "system default". The Settings picker uses `NSOpenPanel` (`allowedContentTypes = [.application]`, defaulting to `/Applications`) and stores the picked `url.path`; "Use Default" sets it to `""`.
- *Why a path, not a bundle id:* simplest for a non-sandboxed app; resolve-and-check existence at open time. A missing path falls back to default, so a moved/deleted app degrades gracefully rather than erroring.

### D2 — Open logic on the model, shared by toolbar + menu
Move opening into `AppModel.openCurrentArtifactExternally()`: read `editorAppPath`; if set and the path exists, `NSWorkspace.shared.open([fileURL], withApplicationAt: appURL, configuration:)`; else `NSWorkspace.shared.open(fileURL)`. The toolbar flyout (relabeled "Open in Editor") and a new File-menu `⌘E` command (`ProjectCommands`, via the focused `AppModel`) both call it; both disable when `currentFilePath() == nil`.
- *Why on the model:* the menu command and the toolbar need one implementation, and the model already owns `currentFilePath()`.

### D3 — Settings section
Add an "Opening artifacts" section to `GeneralSettingsView`: shows the current choice (System default, or the chosen app's display name), a "Choose App…" button (the picker), and a "Use Default" button (disabled when already default). Reuses the existing `Settings`/`⌘,` window.

## Risks / Trade-offs

- **Chosen app missing/moved** → existence check + fallback to default (D1); never errors.
- **`NSOpenPanel` is AppKit/modal** → run from the Settings button via `runModal()`; acceptable in a settings dialog.
- **Not headlessly testable** → `swift build` + manual QA; the override resolution is a small, self-contained branch.

## Migration Plan

Additive, macOS-only. With no app chosen, behavior is identical to today (default-app open) plus a new `⌘E`. Rollback is reverting. No data/contract change.

## Open Questions

- None.
