## Context

Dossier renders views by concatenating individually-styled segments (borders via `separatorStyle`, headers via `headerStyle`, index items via `indexActiveStyle`, etc.). Each lipgloss `Render()` call produces `\033[...mCONTENT\033[0m` — wrapping content in style sequences and terminating with a full SGR reset. When segments are joined on the same line, the reset leaves gaps of terminal-default-background between them. Additionally, when viewport content is shorter than the viewport height, the area below the box frame is unfilled terminal background.

Current view-rendering entry points:
- `View()` at `view.go:12` — dispatches to mode-specific methods
- `viewIndex()` at `view.go:43` — builds box frame + viewport + help bar rows
- `viewConfig()` at `view.go:229` — same pattern
- `addBorderSides()` at `view.go:177` — pads lines with plain `strings.Repeat(" ", pad)` (no bg)

## Goals / Non-Goals

**Goals:**
- Fill the entire terminal viewport with a solid background color when configured
- No visual gaps ("holes") between styled segments on the same line
- No unfilled terminal-default-bg areas below the box frame
- Reusable across all views (index, config, normal, spec viewer)
- When no background is configured, behavior is unchanged
- Foundation for future theme system — the bg color lives on a `Theme` struct

**Non-Goals:**
- Configurable foreground colors or full theme switching
- Per-user configuration file or environment variable support
- Dark/light mode detection
- Changing glamour's internal background (that's glamour's concern)

## Decisions

### Decision 1: Post-process pipeline over style inheritance

**Chosen**: Apply background via a three-step pipeline — wrap rendered content with lipgloss `Background().Width().Height()`, then replace all `\033[0m` with `\033[0;48;5;<color>m`.

**Rationale**: Style inheritance (adding `Background()` to every `lipgloss.Style` var) requires replacing every plain string in rendering functions with a background-styled equivalent — every `"  "`, every intermediate space, every un-styled text segment. This is error-prone and touches ~100 lines across `index.go`, `view.go`, and `tasks.go`. The post-process pipeline touches ~10 lines, is centralized, and any view gains background support by calling one shared method.

**Alternative considered**: Inheriting background on all styles. Discarded because it doesn't cover plain strings between styled segments. Would require replacing every bare string literal in rendering functions with `bgSpacerStyle.Render("  ")`.

### Decision 2: Replace `\033[0m` with `\033[0;48;5;Xm` specifically

**Chosen**: String-level post-processing of the final rendered output to convert all SGR full-reset sequences to "reset + re-apply background."

**Rationale**: `\033[0m` is a complete reset (foreground, background, bold, etc.). Prepending `48;5;X` after the `0` creates `\033[0;48;5;Xm` — first resets everything, then immediately re-sets background to color X. This means the background persists across all inner segment boundaries. Styles with explicit backgrounds (e.g., `indexActiveStyle` with `Background(Color("4"))`) are unaffected because their own sequence `\033[1;48;5;4;38;5;15m` sets a different background that overrides ours, and after their `\033[0m`, the post-process re-applies our background. The `\033[0m` sequence only appears standalone in lipgloss output (never embedded as a parameter within another SGR sequence), so string replacement is safe.

**Alternative considered**: Pre-rendering each line with a lipgloss style that already has the background. This would wrap each line in `\033[48;5;Xm...CONTENT...\033[0m`, but inner `\033[0m` resets within CONTENT still break the background for following characters. The post-process step is still needed.

### Decision 3: Lipgloss wrapping handles vertical and horizontal fill

**Chosen**: Wrap the view string with `lipgloss.NewStyle().Background(bgColor).Width(m.width).Height(m.height).Render(viewString)` before post-processing.

**Rationale**:
- `Width(m.width)` ensures each line is padded to the full terminal width. The padding spaces use `teWhitespace` which includes the background color (since `colorWhitespace` defaults to `true`). This fills horizontal whitespace from the border edges to the terminal edge.
- `Height(m.height)` ensures the content fills the full terminal height. Empty lines from vertical padding are processed by horizontal alignment into full-width background-colored lines. This fills the area below the box frame.
- This works because `alignTextHorizontal` (called for width) processes ALL lines including empty ones from vertical height padding.

### Decision 4: Theme struct on Model

**Chosen**: Add a `Theme` struct to `Model` with a `ViewBg` field (`lipgloss.TerminalColor`). For now, the theme is set once at startup with a hardcoded color. Nil `ViewBg` means "terminal default" (current behavior).

```go
type Theme struct {
    ViewBg lipgloss.TerminalColor
}

// On Model:
type Model struct {
    // ...
    theme Theme
}
```

**Rationale**: A `Theme` type (rather than a bare field) signals intent for future expansion (foreground colors, per-view overrides). The nil-check on `ViewBg` preserves current behavior as the default path.

## Risks / Trade-offs

- [String replacement fragility] → The `\033[0m` → `\033[0;48;5;Xm` replacement relies on lipgloss/termenv always using `\033[0m` as the reset sequence. The termenv library hardcodes this; a version change that alters the reset sequence would break the approach. **Mitigation**: The replacement is in a single helper function, easy to update if termenv changes. A test covering the post-process function validates behavior across lipgloss upgrades.

- [ANSI sequence length] → Post-processing adds `48;5;X` bytes to every reset, growing the total view string. For a 40-line view with ~50 resets per line, this adds ~16KB. **Mitigation**: Negligible impact on render performance (terminal I/O, not string size, is the bottleneck).

- [Color code collision] → If a foreground/background color code coincidentally contains `0m` as part of its sequence, the replacement would corrupt it. **Mitigation**: Lipgloss SGR sequences format as `\033[N;...;Nm` where parameters are semicolon-separated. `0` (reset) is never a valid parameter in a combined sequence alongside other parameters. The only occurrence of `\033[0m` is the standalone reset at the end of `te.Styled()`.

## Open Questions

None. The design is self-contained and the implementation path is clear.
