## Context

`App.swift` is a SwiftUI `App` with a `WindowGroup` + a `.commands` block. There's no `Settings` scene yet, and font sizing is a hardcoded `@ScaledMetric bodySize = 15` in `MarkdownView` (already Dynamic-Type-aware). No `@AppStorage` preferences exist (only the bookmark in UserDefaults).

## Decisions

- **`Settings { }` scene** — SwiftUI auto-wires ⌘, and the "Settings…" menu item; no AppKit plumbing. One pane (`GeneralSettingsView`) now; wrap in a `TabView` only when a second category exists.
- **Single source of truth: `@AppStorage("contentFontScale")` (Double, default 1.0).** Declared wherever read/written (Settings view, the commands, `MarkdownView`); UserDefaults keeps them in sync and persists across launches. **Clamped to 0.8…2.0** so the UI can't reach unreadable extremes.
- **Multiplier, not absolute size.** `MarkdownView` keeps `@ScaledMetric bodySize = 15` (system Dynamic Type) and renders at `bodySize × contentFontScale`. So accessibility settings and the personal preference compose: `effective = DynamicType(15) × userScale`. Headings/code derive from the same base.
- **Content-only.** Only `MarkdownView` consumes the scale. Sidebar rows, toolbar, worktree detail fields, etc. keep their semantic fonts (already Dynamic-Type-aware). This matches the "rendered docs feel small" feedback and avoids a lopsided sidebar.
- **⌘+ / ⌘− / ⌘0 accelerators** in a `CommandGroup` (View menu): increase/decrease by a step (e.g. 0.1) and reset to 1.0, writing the same `@AppStorage` key. Complementary to the Settings slider, not a duplicate.

## Risks / Trade-offs

- **[Low] Two scaling inputs.** Dynamic Type × user scale could compound to large text; the 0.8–2.0 clamp and a sensible step bound it.
- **[Low] Discoverability.** The Settings slider + menu items with shortcuts cover both discovery and speed.
- **[Low] Scope creep.** The Settings scene invites piling on prefs; this change ships only the scaffold + font size — further prefs are separate changes.
