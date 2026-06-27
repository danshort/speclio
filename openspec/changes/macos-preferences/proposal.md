## Why

The rendered content can feel small, and there's no way to adjust it (#63). Rather than a one-off control, the app should gain a standard macOS **Settings window (⌘,)** — the native home for preferences — and ship it with the first real preference: **content font size**. The Settings scene is near-free in SwiftUI and gives future preferences an obvious place to land.

## What Changes

- Add a standard **Settings window** (SwiftUI `Settings` scene → auto-wires ⌘, and the "Settings…" menu item). A single "General" pane for now, structured to grow into tabs later.
- Add an **adjustable content font size**: a user multiplier applied **only to the rendered content** (the `MarkdownView` reading pane), on top of Dynamic Type — `effective = scaledBodySize × userScale`. Adjustable from Settings **and** via **⌘+ / ⌘− / ⌘0** (increase / decrease / reset). Persisted across launches.

## Non-goals

- Scaling the sidebar or app chrome — **content-only** (they use semantic fonts that already follow Dynamic Type).
- Any other preference (appearance override, line spacing, "open in" editor, etc.) — the Settings scene is the scaffold; further prefs land as their own changes when wanted.
- Replacing Dynamic Type — the user scale multiplies on top of it, not instead of it.

## Capabilities

### Modified Capabilities

- `macos-app`: adds a Settings window and an adjustable content font size.

## Impact

- `macos/LecternApp` — `App.swift` (Settings scene + ⌘±/⌘0 commands), a new General settings view, `@AppStorage("contentFontScale")` as the single source of truth, `MarkdownView` reads it. No `OpenSpecKit` / `internal/openspec` / corpus changes.
