package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// shortcut is a single key/description pair shown in the help overlay.
type shortcut struct {
	Keys string
	Desc string
}

// helpGroup is a per-screen cluster of shortcuts shown in the help overlay.
type helpGroup struct {
	Title     string
	Shortcuts []shortcut
}

// helpCatalog is the single source of truth for the keyboard-shortcut overlay.
// Each group documents the bindings actually handled by that screen (see
// viewer.go, index.go, spec.go, config.go). Keep this table in sync when a
// binding changes; helpOverlayContains in the tests pins the key entries.
var helpCatalog = []helpGroup{
	{
		Title: "Global",
		Shortcuts: []shortcut{
			{"?", "toggle this help"},
			{"q / Ctrl+C", "quit"},
		},
	},
	{
		Title: "Index",
		Shortcuts: []shortcut{
			{"j/k ↑↓", "navigate"},
			{"Enter", "open"},
			{"Space", "expand/collapse spec"},
			{"/", "filter"},
			{"s", "toggle spec sort"},
			{"w", "worktrees view"},
			{"i", "project config"},
			{"Esc", "clear filter / quit"},
		},
	},
	{
		Title: "Worktrees",
		Shortcuts: []shortcut{
			{"j/k ↑↓", "navigate"},
			{"Enter", "open change (read-only)"},
			{"e", "edit artifact"},
			{"a / Esc", "index"},
		},
	},
	{
		Title: "Change viewer",
		Shortcuts: []shortcut{
			{"h / l", "previous/next change"},
			{"1-4", "proposal/design/specs/tasks"},
			{"Tab/Shift+Tab/←→", "cycle artifacts"},
			{"j/k ↑↓", "navigate tasks / scroll"},
			{"Space", "toggle task"},
			{"e", "edit artifact"},
			{"i", "project config"},
			{"a / Esc", "index"},
		},
	},
	{
		Title: "Archive viewer",
		Shortcuts: []shortcut{
			{"1-4", "proposal/design/specs/tasks"},
			{"Tab/Shift+Tab/←→", "cycle artifacts"},
			{"j/k ↑↓", "scroll"},
			{"e", "edit artifact"},
			{"i", "project config"},
			{"a / Esc", "index"},
		},
	},
	{
		Title: "Spec viewer",
		Shortcuts: []shortcut{
			{"j/k ↑↓", "scroll"},
			{"h / l", "previous/next requirement (focus mode)"},
			{"e", "edit spec"},
			{"Esc", "index"},
		},
	},
	{
		Title: "Config viewer",
		Shortcuts: []shortcut{
			{"j/k ↑↓", "scroll"},
			{"i / Esc", "back"},
		},
	},
}

// renderHelpOverlay renders the keyboard-shortcut catalog into a centered,
// bordered box. The caller is responsible for placing it over the screen.
func (m *Model) renderHelpOverlay() string {
	// Width of the key column = widest key string across the whole catalog,
	// so descriptions line up in a single column.
	keyWidth := 0
	for _, g := range helpCatalog {
		for _, s := range g.Shortcuts {
			if w := lipgloss.Width(s.Keys); w > keyWidth {
				keyWidth = w
			}
		}
	}

	var b strings.Builder
	b.WriteString(headerStyle.Render("Keyboard shortcuts"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("?, Esc or q to close"))
	for _, g := range helpCatalog {
		b.WriteString("\n\n")
		b.WriteString(sectionStyle.Render(g.Title))
		for _, s := range g.Shortcuts {
			pad := keyWidth - lipgloss.Width(s.Keys)
			if pad < 0 {
				pad = 0
			}
			b.WriteString("\n")
			b.WriteString("  " + taskPendingStyle.Render(s.Keys) + strings.Repeat(" ", pad))
			b.WriteString("  " + helpStyle.Render(s.Desc))
		}
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("12")).
		Padding(1, 2)

	return box.Render(b.String())
}
