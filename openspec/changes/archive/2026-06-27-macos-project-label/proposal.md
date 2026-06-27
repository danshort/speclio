## Why

When you navigate away from the Active tab, nothing tells you which project you're in (#70). Tracing the "where am I" context shows that **mode** is in the toolbar picker and the **artifact** name is in the window title (`.navigationTitle`), but the **project name** has no persistent home anywhere — `WindowGroup("lectern")` would show it, but `.navigationTitle(artifact)` overrides it. The one missing piece is exactly the project identity.

## What Changes

- Add a **project header at the top of the left sidebar** — the project name (folder basename, e.g. `lectern`) shown prominently above the navigation list, always visible while a project is open. This is the primary, durable answer to "which project am I in".
- Move the project name into the **window title bar**: set `.navigationTitle(project.name)` and put the current location (mode · change · artifact) into `.navigationSubtitle(...)`. This keeps project identity visible even when the sidebar is collapsed, and gives a breadcrumb-style trail for free without a separate strip.

## Non-goals

- A dedicated breadcrumb strip below the toolbar — the mode picker + title bar already carry that context; a separate strip would be redundant chrome.
- A project switcher / recents menu, or showing the full project path — out of scope; the basename is the identity.
- Changing the sidebar's tree, selection, or progress behavior.

## Capabilities

### Modified Capabilities

- `macos-app`: adds a persistent project label (sidebar header + window title/subtitle).

## Impact

- `macos/LecternApp` — `ContentView.swift` (sidebar project header; detail-pane `.navigationTitle`/`.navigationSubtitle` now derive from the project + current location). No `OpenSpecKit` / `internal/openspec` / corpus changes.
