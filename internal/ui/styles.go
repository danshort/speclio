package ui

import "charm.land/lipgloss/v2"

// Named palette. Naming the ANSI indices keeps the semantic intent in one place
// and avoids a bare literal like "8" silently meaning three unrelated things.
var (
	headerColor   = lipgloss.Color("12") // bright blue: header/title accent
	brightFg      = lipgloss.Color("15") // bright white: text on active highlight
	activeBg      = lipgloss.Color("4")  // blue: active tab / index selection background
	faintColor    = lipgloss.Color("8")  // bright black: disabled tabs, borders, empty progress
	sectionColor  = lipgloss.Color("11") // yellow: task section headers
	primaryFg     = lipgloss.Color("7")  // white: primary/pending text
	errorColor    = lipgloss.Color("9")  // red: error banner
	progressColor = lipgloss.Color("6")  // cyan: in-progress bar fill
	completeColor = lipgloss.Color("2")  // green: 100%-complete bar

	// dimColor is the "minimized"/secondary text color. ANSI 8 (bright black)
	// renders too dark to read on most terminals; 245 is a mid-gray that stays
	// clearly secondary to primary text (7/15) while remaining legible.
	dimColor = lipgloss.Color("245")
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(headerColor)

	tabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(brightFg).
			Background(activeBg).
			Padding(0, 1)

	indexActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(brightFg).
				Background(activeBg)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(brightFg).
				Padding(0, 1)

	tabDisabledStyle = lipgloss.NewStyle().
				Foreground(faintColor).
				Padding(0, 1)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(sectionColor)

	taskCursorMarkStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(brightFg)

	taskDoneStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	taskPendingStyle = lipgloss.NewStyle().
				Foreground(primaryFg)

	helpStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	errStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	warnStyle = lipgloss.NewStyle().
			Foreground(sectionColor).
			Bold(true)

	progressDoneStyle = lipgloss.NewStyle().
				Foreground(progressColor)

	progressCompleteStyle = lipgloss.NewStyle().
				Foreground(completeColor)

	progressEmptyStyle = lipgloss.NewStyle().
				Foreground(faintColor)

	separatorStyle = lipgloss.NewStyle().
			Foreground(faintColor)
)
