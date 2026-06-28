package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFile(t *testing.T) {
	write := func(t *testing.T, content string) string {
		t.Helper()
		p := filepath.Join(t.TempDir(), "config.toml")
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		return p
	}

	t.Run("missing file yields defaults, no error", func(t *testing.T) {
		c, err := LoadFile(filepath.Join(t.TempDir(), "nope.toml"))
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if c.Editor.OpenWith != "" {
			t.Errorf("expected default empty OpenWith, got %q", c.Editor.OpenWith)
		}
	})

	t.Run("parses editor.open_with", func(t *testing.T) {
		c, err := LoadFile(write(t, "[editor]\nopen_with = \"system\"\n"))
		if err != nil {
			t.Fatal(err)
		}
		if c.Editor.OpenWith != "system" {
			t.Errorf("got %q", c.Editor.OpenWith)
		}
	})

	t.Run("malformed file returns an error", func(t *testing.T) {
		_, err := LoadFile(write(t, "this is = not valid = toml ["))
		if err == nil {
			t.Error("expected a parse error for malformed TOML")
		}
	})

	t.Run("smart-quoted value is rejected (real QA failure)", func(t *testing.T) {
		// A curly opening quote is invalid TOML; this must error (and be
		// surfaced), not silently fall back to defaults.
		_, err := LoadFile(write(t, "[editor]\nopen_with = “system\"\n"))
		if err == nil {
			t.Error("expected a parse error for a smart-quoted value")
		}
	})

	t.Run("unknown keys are ignored", func(t *testing.T) {
		c, err := LoadFile(write(t, "future_thing = 42\n[editor]\nopen_with = \"nvim\"\nunknown = true\n"))
		if err != nil {
			t.Fatalf("unknown keys should not error, got %v", err)
		}
		if c.Editor.OpenWith != "nvim" {
			t.Errorf("recognized key should still load, got %q", c.Editor.OpenWith)
		}
	})
}

func TestEnsureFile(t *testing.T) {
	t.Run("creates a parseable starter file when missing", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		path, err := EnsureFile()
		if err != nil {
			t.Fatal(err)
		}
		if path != filepath.Join(dir, "lectern", "config.toml") {
			t.Errorf("unexpected path %q", path)
		}
		// The starter template must parse and yield defaults.
		c, err := LoadFile(path)
		if err != nil {
			t.Fatalf("starter file should parse, got %v", err)
		}
		if c.Editor.OpenWith != "" {
			t.Errorf("starter file should leave defaults, got %q", c.Editor.OpenWith)
		}
	})

	t.Run("leaves an existing file untouched", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		path := filepath.Join(dir, "lectern", "config.toml")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		const existing = "[editor]\nopen_with = \"nvim\"\n"
		if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := EnsureFile(); err != nil {
			t.Fatal(err)
		}
		got, _ := os.ReadFile(path)
		if string(got) != existing {
			t.Errorf("existing file should be untouched, got %q", string(got))
		}
	})
}

func TestPath(t *testing.T) {
	t.Run("honors XDG_CONFIG_HOME", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/cfg")
		p, err := Path()
		if err != nil {
			t.Fatal(err)
		}
		if p != filepath.Join("/cfg", "lectern", "config.toml") {
			t.Errorf("got %q", p)
		}
	})

	t.Run("falls back to ~/.config", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("HOME", "/home/tester")
		p, err := Path()
		if err != nil {
			t.Fatal(err)
		}
		if p != filepath.Join("/home/tester", ".config", "lectern", "config.toml") {
			t.Errorf("got %q", p)
		}
	})
}
