package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/danshort/lectern/internal/config"
	"github.com/danshort/lectern/internal/openspec"
	"github.com/danshort/lectern/internal/ui"
)

var version string

func main() {
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.BoolVar(&showVersion, "v", false, "print version and exit (shorthand)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: lectern [flags] [path]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "A keyboard-driven TUI for navigating OpenSpec project artifacts.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  path  Optional path to a change directory (single-change mode)")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if showVersion {
		fmt.Println("lectern", version)
		return
	}

	if flag.NArg() > 1 {
		fmt.Fprintln(os.Stderr, "error: too many arguments; expected at most one path")
		flag.Usage()
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot determine working directory:", err)
		os.Exit(1)
	}

	cfg, err := openspec.LoadConfigFrom(cwd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: loading config:", err)
		os.Exit(1)
	}

	// User config is optional: a malformed file warns and falls back to defaults
	// rather than blocking launch over a cosmetic setting. The warning is shown
	// both on stderr (visible after exit) and in the TUI status line (visible
	// during the session, since the alt-screen hides stderr).
	userCfg, cfgErr := config.Load()
	configWarn := ""
	if cfgErr != nil {
		configWarn = "ignoring invalid config (using defaults): " + cfgErr.Error()
		fmt.Fprintln(os.Stderr, "warning:", configWarn)
	}

	loader := openspec.NewLoader(openspec.OSFS{})

	var (
		project *openspec.Project
		model   ui.Model
	)
	if path := flag.Arg(0); path != "" {
		project, err = openspec.LoadFromPath(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		model = ui.NewSinglePath(project, cfg, path, loader, userCfg.Editor.OpenWith, configWarn)
	} else {
		project, err = openspec.LoadFrom(cwd)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		model = ui.New(project, cfg, cwd, loader, userCfg.Editor.OpenWith, configWarn)
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
