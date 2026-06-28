package config

import (
	"os"
	"runtime"
	"strings"
)

// OpenMode is how the TUI should launch the resolved opener.
type OpenMode int

const (
	// OpenTerminal runs an in-terminal editor: the TUI yields the terminal
	// (tea.ExecProcess) and resumes when the editor exits.
	OpenTerminal OpenMode = iota
	// OpenDetached launches the OS default handler (a GUI app) without yielding
	// the terminal; the saved file is picked up by the normal reload path.
	OpenDetached
)

// Opener is a resolved, OS-appropriate way to open a file. The artifact path is
// appended to Args by the caller at open time.
type Opener struct {
	Mode OpenMode
	Name string   // executable to run
	Args []string // arguments preceding the file path
}

// ResolveOpener maps an editor.open_with value to a concrete Opener for the
// current OS. The launch mode is implied by the value:
//   - "" or "$EDITOR" → $EDITOR (fields), else vi — terminal
//   - "system"        → OS default handler — detached
//   - anything else   → that command (fields) — terminal
func ResolveOpener(openWith string) Opener {
	switch strings.TrimSpace(openWith) {
	case "system":
		name, args := systemHandler()
		return Opener{Mode: OpenDetached, Name: name, Args: args}
	case "", "$EDITOR":
		return terminalOpener(os.Getenv("EDITOR"))
	default:
		return terminalOpener(openWith)
	}
}

// terminalOpener field-splits a command, falling back to vi when empty.
func terminalOpener(command string) Opener {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		fields = []string{"vi"}
	}
	return Opener{Mode: OpenTerminal, Name: fields[0], Args: fields[1:]}
}

// systemHandler returns the OS default-handler command and any leading args.
func systemHandler() (string, []string) {
	switch runtime.GOOS {
	case "darwin":
		return "open", nil
	case "windows":
		// `start` is a cmd builtin; the empty "" is the (ignored) window title.
		return "cmd", []string{"/c", "start", ""}
	default:
		return "xdg-open", nil
	}
}
