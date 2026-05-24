## Why

Currently, dossier's views inherit the terminal emulator's default background color, which varies across terminals. Users cannot control the visual backdrop of any view. This is particularly noticeable in the index view, where whitespace areas between elements and below the box frame expose the terminal background rather than a cohesive application background. As dossier moves toward a future theme system, establishing a configurable per-view background color is the foundational step.

## What Changes

- Define a per-view background color that fills the entire terminal area (no gaps, no terminal-default-bg "holes")
- The background applies consistently across: box borders, content area, inter-element whitespace, and empty area below the box
- Index view is the first view to support this; the rendering pipeline is built to be reusable by all views
- When no background color is configured, behavior is unchanged (terminal default background)
- The color is hardcoded for now, with the value and its location on the Model struct anticipating future theme configuration

## Capabilities

### New Capabilities
- `view-background`: Configurable per-view background color that fills the entire terminal viewport with a solid color, including whitespace areas between styled elements and empty vertical space. Applies to all views via a shared rendering pipeline.

### Modified Capabilities
- `change-index`: The index view now renders with a configurable background color instead of relying on terminal default.
- `tui-viewer`: The main View() method and view-specific rendering functions (viewIndex, viewConfig, etc.) gain a shared background-fill rendering pipeline.

## Impact

- `internal/ui/view.go` — `viewIndex()`, `viewConfig()`, and the main `View()` method adopt the background pipeline
- `internal/ui/styles.go` — introduces a `Theme` struct and a `viewBackgroundStyle` helper (or keeps styles as-is and adds background at the pipeline level)
- `internal/ui/model.go` — gains a `theme` field (or `viewBg` field) on the Model struct
- No external API changes
- No new dependencies
