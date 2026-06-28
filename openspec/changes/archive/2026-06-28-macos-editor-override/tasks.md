## 1. Preference + picker (Settings)

- [x] 1.1 Add an `EditorPref` (storageKey `editorAppPath`, "" = system default) in `Preferences.swift`
- [x] 1.2 Add an "Opening artifacts" section to `GeneralSettingsView`: show the current choice (System default / app display name), a "Choose App…" button (`NSOpenPanel`, `allowedContentTypes = [.application]`, default `/Applications`) storing `url.path`, and a "Use Default" button (clears it, disabled when already default)

## 2. Open action honors the override

- [x] 2.1 Add `AppModel.openCurrentArtifactExternally()`: read `editorAppPath`; if set and the path exists, `NSWorkspace.open([fileURL], withApplicationAt: appURL, configuration:)`; else `NSWorkspace.open(fileURL)`
- [x] 2.2 `ContentView`: flyout item calls `model.openCurrentArtifactExternally()`, relabeled "Open in Editor"; remove the local `openInEditor`

## 3. Keyboard shortcut + menu

- [x] 3.1 `App.swift` `ProjectCommands`: add a File-menu "Open in Editor" button bound to `⌘E` via the focused `AppModel`, disabled when `currentFilePath() == nil`

## 4. Verification

- [x] 4.1 `swift build` (LecternApp) clean; OpenSpecKit tests + golden green (unaffected)
- [ ] 4.2 Manual: with no app chosen, `⌘E`/flyout opens the default app; choose an app in Settings → opens in it; remove that app → falls back to default; "Use Default" resets
