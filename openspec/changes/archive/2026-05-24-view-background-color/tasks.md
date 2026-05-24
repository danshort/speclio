## 1. Data model

- [x] 1.1 Define `Theme` struct with `ViewBg lipgloss.TerminalColor` field in `internal/ui/model.go`
- [x] 1.2 Add `theme Theme` field to the `Model` struct
- [x] 1.3 Set a default hardcoded background color on the theme during `New()` (e.g., `lipgloss.Color("234")` or `lipgloss.NoColor` for no-op)

## 2. Background render pipeline

- [x] 2.1 Implement `func (m *Model) renderWithBackground(content string) string` in `internal/ui/view.go`
  - Wraps content with `lipgloss.NewStyle().Background(m.theme.ViewBg).Width(m.width).Height(m.height).Render(content)`
  - Post-processes: `strings.ReplaceAll(wrapped, "\033[0m", fmt.Sprintf("\033[0;48;5;%dm", bgValue))`
  - Skips when `m.theme.ViewBg` is nil or `lipgloss.NoColor`
- [x] 2.2 Add a helper to extract the ANSI color number from `lipgloss.TerminalColor` to build the SGR restore sequence

## 3. Wire into views

- [x] 3.1 Call `renderWithBackground` from `viewIndex()` wrapping the `strings.Join(rows, "\n")` result
- [x] 3.2 Call `renderWithBackground` from `viewConfig()` wrapping the `strings.Join(rows, "\n")` result
- [x] 3.3 Call `renderWithBackground` from the main `View()` method wrapping the `strings.Join(rows, "\n")` result in the `ModeNormal` path
- [x] 3.4 Verify `emptyView()` and spec viewer rendering also pass through the pipeline (or are handled by `View()`)

## 4. Verification

- [x] 4.1 Run existing tests with `go test ./...` to ensure no regressions
- [x] 4.2 Manual visual check: launch dossier with and without the background color to verify solid fill and no gaps
- [x] 4.3 Manual check: ensure cursor-highlighted index items still show their dark blue background on top of the view background
