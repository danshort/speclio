package openspec

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// golden_test.go pins the domain layer's observable behavior against a committed
// fixture corpus, so the Go loader and the planned Swift port (OpenSpecKit)
// reproduce identical output. It covers every public entry point — not only the
// parsed Project, but ParseTasks, ExtractRequirement, parseWorktreeList,
// ConfigToMarkdown, validation, and the byte-exact task-toggle write path.
//
// Regenerate goldens with:  go test ./internal/openspec/ -run TestGolden -update
// See testdata/corpus/README.md for the serialization contract.

var update = flag.Bool("update", false, "regenerate golden files")

func corpusDir(t *testing.T) string {
	t.Helper()
	d, err := filepath.Abs(filepath.Join("..", "..", "testdata", "corpus"))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// ── canonical serialization ──────────────────────────────────────────────────

// canonicalJSON emits the shared contract: sorted keys, 2-space indent, no HTML
// escaping, trailing newline. It round-trips through `any` so every object's
// keys sort alphabetically (matching Swift's JSONEncoder .sortedKeys), and uses
// json.Number so integers survive without float reformatting.
func canonicalJSON(t *testing.T, v any) []byte {
	t.Helper()
	var raw bytes.Buffer
	e1 := json.NewEncoder(&raw)
	e1.SetEscapeHTML(false)
	if err := e1.Encode(v); err != nil {
		t.Fatalf("encode: %v", err)
	}
	dec := json.NewDecoder(&raw)
	dec.UseNumber()
	var generic any
	if err := dec.Decode(&generic); err != nil {
		t.Fatalf("decode: %v", err)
	}
	var out bytes.Buffer
	e2 := json.NewEncoder(&out)
	e2.SetEscapeHTML(false)
	e2.SetIndent("", "  ")
	if err := e2.Encode(generic); err != nil {
		t.Fatalf("re-encode: %v", err)
	}
	return out.Bytes()
}

func checkGolden(t *testing.T, name string, actual []byte) {
	t.Helper()
	goldenPath := filepath.Join(corpusDir(t), "golden", name)
	if *update {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, actual, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with -update to generate)", name, err)
	}
	if !bytes.Equal(want, actual) {
		t.Errorf("golden mismatch for %s\n--- want ---\n%s\n--- got ---\n%s", name, want, actual)
	}
}

func checkGoldenJSON(t *testing.T, name string, v any) {
	t.Helper()
	checkGolden(t, name, canonicalJSON(t, v))
}

// ── path / read-error normalization (the machine- and language-portable view) ─

func relPath(abs, root string) string {
	r, err := filepath.Rel(root, abs)
	if err != nil {
		r = abs
	}
	return filepath.ToSlash(r)
}

// normContent makes artifact content portable: a readable artifact's content is
// returned as-is; an unreadable artifact's placeholder is reduced to the prefix
// plus the relative path, dropping the OS/locale-specific error tail.
func normContent(content string, readErr error, root string) string {
	if readErr == nil {
		return content
	}
	// A missing-spec.md placeholder is "<prefix><dir>" with no error tail.
	if rest, ok := strings.CutPrefix(content, missingSpecPrefix); ok {
		return missingSpecPrefix + relPath(rest, root)
	}
	rest := strings.TrimPrefix(content, unreadablePrefix)
	if i := strings.Index(rest, ": "); i >= 0 {
		rest = rest[:i]
	}
	return unreadablePrefix + relPath(rest, root)
}

// Golden DTOs: read errors become a boolean; paths are relative. See README.
type gArtifact struct {
	Content   string `json:"content"`
	Present   bool   `json:"present"`
	ReadError bool   `json:"read_error"`
}

type gNamedSpec struct {
	Name      string `json:"name"`
	Content   string `json:"content"`
	ReadError bool   `json:"read_error"`
}

type gChange struct {
	Name        string       `json:"name"`
	Path        string       `json:"path"`
	Created     string       `json:"created"`
	DisplayDate string       `json:"display_date"`
	Proposal    gArtifact    `json:"proposal"`
	Design      gArtifact    `json:"design"`
	Tasks       gArtifact    `json:"tasks"`
	Specs       gArtifact    `json:"specs"`
	SpecFiles   []gNamedSpec `json:"spec_files"`
}

type gProject struct {
	Name    string    `json:"name"`
	Changes []gChange `json:"changes"`
}

func normArtifact(a Artifact, root string) gArtifact {
	return gArtifact{
		Content:   normContent(a.Content, a.ReadErr, root),
		Present:   a.Present,
		ReadError: a.ReadErr != nil,
	}
}

func normChange(c Change, root string) gChange {
	gc := gChange{
		Name:        c.Name,
		Path:        relPath(c.Path, root),
		Created:     c.Created,
		DisplayDate: c.DisplayDate,
		Proposal:    normArtifact(c.Proposal, root),
		Design:      normArtifact(c.Design, root),
		Tasks:       normArtifact(c.Tasks, root),
		Specs:       normArtifact(c.Specs, root),
		SpecFiles:   []gNamedSpec{},
	}
	for _, sf := range c.SpecFiles {
		gc.SpecFiles = append(gc.SpecFiles, gNamedSpec{
			Name:      sf.Name,
			Content:   normContent(sf.Content, sf.ReadErr, root),
			ReadError: sf.ReadErr != nil,
		})
	}
	return gc
}

func normChanges(changes []Change, root string) []gChange {
	out := []gChange{}
	for _, c := range changes {
		out = append(out, normChange(c, root))
	}
	return out
}

func normProject(p *Project, root string) gProject {
	return gProject{Name: p.Name, Changes: normChanges(p.Changes, root)}
}

type gProjectSpec struct {
	Name             string   `json:"name"`
	RequirementCount int      `json:"requirement_count"`
	RequirementNames []string `json:"requirement_names"`
	Content          string   `json:"content"`
	ReadError        bool     `json:"read_error"`
}

func normProjectSpecs(specs []ProjectSpec, root string) []gProjectSpec {
	out := []gProjectSpec{}
	for _, ps := range specs {
		names := ps.RequirementNames
		if names == nil {
			names = []string{}
		}
		out = append(out, gProjectSpec{
			Name:             ps.Name,
			RequirementCount: ps.RequirementCount,
			RequirementNames: names,
			Content:          normContent(ps.Content, ps.ReadErr, root),
			ReadError:        ps.ReadErr != nil,
		})
	}
	return out
}

// faultFS injects a synthetic (non-not-found) read error for a single path, so
// the present-but-unreadable branch is exercised deterministically — a real
// unreadable file cannot survive a git checkout.
type faultFS struct {
	OSFS
	failPath string
}

func (f faultFS) ReadFile(name string) ([]byte, error) {
	if name == f.failPath {
		return nil, errors.New("permission denied")
	}
	return f.OSFS.ReadFile(name)
}

// ── the golden tests ─────────────────────────────────────────────────────────

func TestGolden(t *testing.T) {
	corpus := corpusDir(t)

	t.Run("project/basic-project", func(t *testing.T) {
		root := filepath.Join(corpus, "basic-project")
		p, err := NewLoader(OSFS{}).LoadFrom(root)
		if err != nil {
			t.Fatal(err)
		}
		checkGoldenJSON(t, "basic-project.project.json", normProject(p, root))
	})

	t.Run("project/malformed-meta", func(t *testing.T) {
		root := filepath.Join(corpus, "malformed-meta")
		p, err := NewLoader(OSFS{}).LoadFrom(root)
		if err != nil {
			t.Fatal(err)
		}
		checkGoldenJSON(t, "malformed-meta.project.json", normProject(p, root))
	})

	t.Run("project/unreadable-artifact", func(t *testing.T) {
		root := filepath.Join(corpus, "unreadable-artifact")
		fail := filepath.Join(root, DirOpenspec, DirChanges, "has-unreadable", FileProposal)
		p, err := NewLoader(faultFS{failPath: fail}).LoadFrom(root)
		if err != nil {
			t.Fatal(err)
		}
		checkGoldenJSON(t, "unreadable-artifact.project.json", normProject(p, root))
	})

	t.Run("archive/malformed-archive-name", func(t *testing.T) {
		root := filepath.Join(corpus, "malformed-archive-name")
		changes, err := NewLoader(OSFS{}).ListArchiveChangesFrom(root)
		if err != nil {
			t.Fatal(err)
		}
		checkGoldenJSON(t, "malformed-archive-name.archive.json", normChanges(changes, root))
	})

	t.Run("tasks/parse", func(t *testing.T) {
		// LF and CRLF must parse to identical TaskItems (splitLines strips \r).
		lf := readFile(t, filepath.Join(corpus, "lf-tasks", "tasks.md"))
		crlf := readFile(t, filepath.Join(corpus, "crlf-tasks", "tasks.md"))
		lfItems := ParseTasks(string(lf))
		crlfItems := ParseTasks(string(crlf))
		checkGoldenJSON(t, "tasks.json", lfItems)
		if !bytes.Equal(canonicalJSON(t, lfItems), canonicalJSON(t, crlfItems)) {
			t.Errorf("CRLF and LF tasks parsed differently")
		}
	})

	t.Run("project-specs/basic-project", func(t *testing.T) {
		root := filepath.Join(corpus, "basic-project")
		specs, err := NewLoader(OSFS{}).LoadProjectSpecsFrom(root)
		if err != nil {
			t.Fatal(err)
		}
		checkGoldenJSON(t, "basic-project.project-specs.json", normProjectSpecs(specs, root))
	})

	t.Run("requirements/extract", func(t *testing.T) {
		spec := string(readFile(t, filepath.Join(corpus, "basic-project", "openspec", "changes", "alpha-change", "specs", "auth", "spec.md")))
		out := map[string]string{
			"Login":   ExtractRequirement(spec, "Login"),
			"Logout":  ExtractRequirement(spec, "Logout"),
			"Missing": ExtractRequirement(spec, "Missing"),
		}
		checkGoldenJSON(t, "requirements.json", out)
	})

	t.Run("config/variants", func(t *testing.T) {
		configs := map[string]ProjectConfig{}
		for _, v := range []string{"absent-rules", "empty-rules", "multiline"} {
			root := filepath.Join(corpus, "config-variants", v)
			cfg, err := NewLoader(OSFS{}).LoadConfigFrom(root)
			if err != nil {
				t.Fatalf("%s: %v", v, err)
			}
			configs[v] = cfg
			checkGolden(t, "config-"+v+".md", []byte(ConfigToMarkdown(cfg)))
		}
		checkGoldenJSON(t, "config.json", configs)
	})

	t.Run("worktrees/parse", func(t *testing.T) {
		for _, f := range []string{"three-worktrees", "locked-prunable"} {
			out := readFile(t, filepath.Join(corpus, "worktree-porcelain", f+".txt"))
			checkGoldenJSON(t, "worktrees-"+f+".json", parseWorktreeList(out))
		}
	})

	t.Run("validation", func(t *testing.T) {
		result := map[string][]string{}
		specsDir := filepath.Join(corpus, "delta-specs", "specs")
		for _, name := range []string{"valid-spec", "missing-sections", "req-without-scenario"} {
			content := string(readFile(t, filepath.Join(specsDir, name, FileSpec)))
			result["spec/"+name] = orEmpty(ValidateSpec(content))
		}
		changeRoot := filepath.Join(corpus, "delta-specs", "changes", "needs-work")
		ch := NewLoader(OSFS{}).loadChangeFromDir(changeRoot, "needs-work", "")
		result["change/needs-work"] = orEmpty(ValidateChange(ch))
		checkGoldenJSON(t, "validation.json", result)
	})
}

// TestGoldenToggle pins the byte-exact write path: toggling a task preserves the
// file's existing line endings (LF stays LF, CRLF stays CRLF).
func TestGoldenToggle(t *testing.T) {
	corpus := corpusDir(t)
	for _, fixture := range []string{"lf-tasks", "crlf-tasks"} {
		t.Run(fixture, func(t *testing.T) {
			src := readFile(t, filepath.Join(corpus, fixture, "tasks.md"))
			tmp := filepath.Join(t.TempDir(), "tasks.md")
			if err := os.WriteFile(tmp, src, 0o644); err != nil {
				t.Fatal(err)
			}
			items := ParseTasks(string(src))
			idx := FindCursorByText(items, "1.1 alpha")
			if err := ToggleTask(tmp, items, idx); err != nil {
				t.Fatal(err)
			}
			got := readFile(t, tmp)
			checkGolden(t, fixture+".after-toggle.tasks.md", got)
		})
	}
}

// orEmpty coerces a nil slice to an empty one so the golden emits `[]` not
// `null`, per the serialization contract (a Swift validator returns `[]`).
func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
