## 1. Settings scene

- [x] 1.1 Add a SwiftUI `Settings { GeneralSettingsView() }` scene in `App.swift` (auto-wires ⌘, and the "Settings…" menu item)
- [x] 1.2 `GeneralSettingsView` — a single "General" pane, structured so a `TabView` can wrap it later

## 2. Font-size preference

- [x] 2.1 `@AppStorage("contentFontScale")` (Double, default 1.0) as the single source of truth, clamped to 0.8…2.0
- [x] 2.2 Settings control: a slider/stepper for text size with min/max affordances (e.g. small "A" → large "A") and a reset
- [x] 2.3 `⌘+ / ⌘− / ⌘0` commands (View menu) that step / reset the same `@AppStorage` key

## 3. Apply to content only

- [x] 3.1 `MarkdownView` renders at `bodySize × contentFontScale` (Dynamic Type × user multiplier); headings/code derive from the same base
- [x] 3.2 Confirm the sidebar, toolbar, and other chrome are unaffected (content-only)

## 4. Spec + verification

- [x] 4.1 Delta spec: ADD "Settings window" and "Adjustable content font size" requirements to `macos-app`
- [ ] 4.2 `swift build` green; manual QA: slider + ⌘±/⌘0 change content size, persist across relaunch, sidebar/chrome unchanged
- [x] 4.3 Confirm no domain/golden changes; Go + Swift lanes unaffected
