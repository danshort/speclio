package openspec

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// errFS delegates to the real filesystem but returns a chosen non-not-found
// error when ReadFile is called on failPath — to exercise the "present but
// unreadable" path portably (no chmod, works as root / on CI).
type errFS struct {
	OSFS
	failPath string
	failErr  error
}

func (f errFS) ReadFile(name string) ([]byte, error) {
	if name == f.failPath {
		return nil, f.failErr
	}
	return f.OSFS.ReadFile(name)
}

func TestLoadFileUnreadableVsMissing(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "proposal.md")
	if err := os.WriteFile(good, []byte("# hi"), 0644); err != nil {
		t.Fatal(err)
	}

	l := NewLoader(errFS{failPath: good, failErr: errors.New("permission denied")})

	t.Run("unreadable file is present with placeholder and ReadErr", func(t *testing.T) {
		a := l.loadFile(good)
		if !a.Present || a.ReadErr == nil {
			t.Fatalf("expected present+ReadErr, got present=%v err=%v", a.Present, a.ReadErr)
		}
		if !strings.HasPrefix(a.Content, unreadablePrefix) {
			t.Errorf("expected placeholder content, got %q", a.Content)
		}
	})

	t.Run("missing file stays absent", func(t *testing.T) {
		a := l.loadFile(filepath.Join(dir, "nope.md"))
		if a.Present || a.ReadErr != nil {
			t.Errorf("expected absent (no ReadErr), got present=%v err=%v", a.Present, a.ReadErr)
		}
	})
}

func TestLoadProjectSpecsResilient(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, DirOpenspec, DirSpecs)
	for _, name := range []string{"auth", "export"} {
		if err := os.MkdirAll(filepath.Join(specsDir, name), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(specsDir, name, FileSpec), []byte("## Purpose\nP\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	badSpec := filepath.Join(specsDir, "auth", FileSpec)
	l := NewLoader(errFS{failPath: badSpec, failErr: errors.New("EIO")})

	specs, err := l.LoadProjectSpecsFrom(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(specs) != 2 {
		t.Fatalf("expected 2 specs (one unreadable does not sink the rest), got %d", len(specs))
	}
	byName := map[string]ProjectSpec{}
	for _, s := range specs {
		byName[s.Name] = s
	}
	if byName["auth"].ReadErr == nil || !strings.HasPrefix(byName["auth"].Content, unreadablePrefix) {
		t.Errorf("auth should be unreadable with placeholder, got %+v", byName["auth"])
	}
	if byName["export"].ReadErr != nil {
		t.Errorf("export should load normally, got ReadErr=%v", byName["export"].ReadErr)
	}
}

func TestLoadSpecsResilient(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a", "b"} {
		if err := os.MkdirAll(filepath.Join(dir, name), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name, FileSpec), []byte("# "+name), 0644); err != nil {
			t.Fatal(err)
		}
	}
	bad := filepath.Join(dir, "b", FileSpec)
	l := NewLoader(errFS{failPath: bad, failErr: errors.New("EIO")})

	agg, files := l.loadSpecs(dir)
	if len(files) != 2 {
		t.Fatalf("expected 2 NamedSpecs (one unreadable kept), got %d", len(files))
	}
	byName := map[string]NamedSpec{}
	for _, f := range files {
		byName[f.Name] = f
	}
	if byName["b"].ReadErr == nil || !strings.HasPrefix(byName["b"].Content, unreadablePrefix) {
		t.Errorf("b should be unreadable with placeholder, got %+v", byName["b"])
	}
	if byName["a"].ReadErr != nil {
		t.Errorf("a should load normally, got ReadErr=%v", byName["a"].ReadErr)
	}
	if !agg.Present || !strings.Contains(agg.Content, unreadablePrefix) {
		t.Error("aggregate Specs artifact should be present and embed the placeholder")
	}
}

func TestValidateChangeSkipsUnreadable(t *testing.T) {
	boom := errors.New("permission denied")

	t.Run("unreadable proposal is not reported missing", func(t *testing.T) {
		ch := Change{Proposal: Artifact{Present: true, ReadErr: boom, Content: unreadablePrefix + "proposal.md: x"}}
		for _, e := range ValidateChange(ch) {
			if strings.Contains(e, "missing proposal.md") {
				t.Errorf("unreadable proposal should not be reported missing; got %q", e)
			}
		}
	})

	t.Run("genuinely missing proposal is still reported", func(t *testing.T) {
		errs := ValidateChange(Change{})
		found := false
		for _, e := range errs {
			if strings.Contains(e, "missing proposal.md") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected 'missing proposal.md', got %v", errs)
		}
	})

	t.Run("unreadable delta spec is not validated as malformed", func(t *testing.T) {
		ch := Change{
			Proposal:  Artifact{Present: true},
			SpecFiles: []NamedSpec{{Name: "auth", Content: unreadablePrefix + "spec.md: x", ReadErr: boom}},
		}
		for _, e := range ValidateChange(ch) {
			if strings.Contains(e, "delta spec") {
				t.Errorf("unreadable delta spec should be skipped, got %q", e)
			}
		}
	})
}
