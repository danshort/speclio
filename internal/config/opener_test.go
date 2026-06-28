package config

import (
	"runtime"
	"testing"
)

func TestResolveOpener(t *testing.T) {
	t.Run("default uses $EDITOR split into fields", func(t *testing.T) {
		t.Setenv("EDITOR", "code --wait")
		op := ResolveOpener("")
		if op.Mode != OpenTerminal || op.Name != "code" || len(op.Args) != 1 || op.Args[0] != "--wait" {
			t.Errorf("got %+v", op)
		}
	})

	t.Run("default falls back to vi when $EDITOR unset", func(t *testing.T) {
		t.Setenv("EDITOR", "")
		op := ResolveOpener("")
		if op.Mode != OpenTerminal || op.Name != "vi" || len(op.Args) != 0 {
			t.Errorf("got %+v", op)
		}
	})

	t.Run("literal $EDITOR behaves like unset", func(t *testing.T) {
		t.Setenv("EDITOR", "nano")
		op := ResolveOpener("$EDITOR")
		if op.Mode != OpenTerminal || op.Name != "nano" {
			t.Errorf("got %+v", op)
		}
	})

	t.Run("system resolves to the OS default handler, detached", func(t *testing.T) {
		op := ResolveOpener("system")
		if op.Mode != OpenDetached {
			t.Fatalf("expected detached, got %v", op.Mode)
		}
		want := map[string]string{"darwin": "open", "windows": "cmd"}
		expected := want[runtime.GOOS]
		if expected == "" {
			expected = "xdg-open"
		}
		if op.Name != expected {
			t.Errorf("GOOS %s: got handler %q, want %q", runtime.GOOS, op.Name, expected)
		}
	})

	t.Run("custom command resolves to a terminal editor", func(t *testing.T) {
		op := ResolveOpener("nvim")
		if op.Mode != OpenTerminal || op.Name != "nvim" {
			t.Errorf("got %+v", op)
		}
	})

	t.Run("whitespace-only command falls back to vi", func(t *testing.T) {
		op := ResolveOpener("   ")
		if op.Mode != OpenTerminal || op.Name != "vi" {
			t.Errorf("got %+v", op)
		}
	})
}
