package openspec

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// OpenSpec directory and artifact names, centralized so the layout lives in one
// place (consumed by this package and by internal/ui path-building).
const (
	DirOpenspec = "openspec"
	DirChanges  = "changes"
	DirSpecs    = "specs"
	DirArchive  = "archive"

	FileProposal = "proposal.md"
	FileDesign   = "design.md"
	FileTasks    = "tasks.md"
	FileSpec     = "spec.md"
	FileConfig   = "config.yaml"
	fileMeta     = ".openspec.yaml"

	// reqPrefix / scenarioPrefix mark requirement and scenario headers. They omit
	// the trailing space; matching uses HasPrefix + TrimSpace (so a line like
	// "### Requirement: Name" yields "Name").
	reqPrefix      = "### Requirement:"
	scenarioPrefix = "#### Scenario:"

	// unreadablePrefix begins the placeholder content of an artifact that exists
	// but could not be read (its ReadErr is also set).
	unreadablePrefix = "⚠ couldn't read "
)

// JSON tags on the domain types document the cross-language serialization
// contract shared with the Swift port (snake_case field names). The error
// fields are tagged `json:"-"`: a Go error has no byte-stable, cross-language
// representation, so the golden harness records read failures as a derived
// boolean plus normalized placeholder content instead. See
// testdata/corpus/README.md for the full contract.
type Artifact struct {
	Content string `json:"content"`
	Present bool   `json:"present"`
	// ReadErr is set when the file exists but could not be read (a non-not-found
	// error). The artifact is still Present, with placeholder Content; callers
	// surface it (a ⚠ marker, the placeholder on open) instead of treating it as
	// absent or as a validation failure.
	ReadErr error `json:"-"`
}

type NamedSpec struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	ReadErr error  `json:"-"`
}

type Change struct {
	Name        string      `json:"name"`
	Path        string      `json:"path"`
	Created     string      `json:"created"`
	DisplayDate string      `json:"display_date"`
	Proposal    Artifact    `json:"proposal"`
	Design      Artifact    `json:"design"`
	Tasks       Artifact    `json:"tasks"`
	Specs       Artifact    `json:"specs"`
	SpecFiles   []NamedSpec `json:"spec_files"`
}

type Project struct {
	Name    string   `json:"name"`
	Changes []Change `json:"changes"`
}

type ProjectSpec struct {
	Name             string   `json:"name"`
	RequirementCount int      `json:"requirement_count"`
	RequirementNames []string `json:"requirement_names"`
	Content          string   `json:"content"`
	ReadErr          error    `json:"-"`
}

type openspecMeta struct {
	Schema  string `yaml:"schema"`
	Created string `yaml:"created"`
}

type ProjectConfig struct {
	Context string              `json:"context"`
	Rules   map[string][]string `json:"rules"`
}

type projectConfigYAML struct {
	Context string              `yaml:"context"`
	Rules   map[string][]string `yaml:"rules"`
}

type Loader struct {
	fs fileSystem
}

func NewLoader(fs fileSystem) *Loader {
	return &Loader{fs: fs}
}

var defaultLoader = NewLoader(OSFS{})

// ── *From(root) variants ──────────────────────────────────────────────────────

func (l *Loader) LoadFrom(root string) (*Project, error) {
	openspecDir := filepath.Join(root, DirOpenspec)
	if _, err := l.fs.Stat(openspecDir); errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("no openspec/ directory found in %s", root)
	}

	project := &Project{Name: filepath.Base(root)}

	changesDir := filepath.Join(openspecDir, DirChanges)
	names, err := l.listDirs(changesDir, true)
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		ch := l.loadChangeFromDir(filepath.Join(changesDir, name), name, "")
		project.Changes = append(project.Changes, ch)
	}

	sort.SliceStable(project.Changes, func(i, j int) bool {
		a, b := project.Changes[i].Created, project.Changes[j].Created
		switch {
		case a == "" && b == "":
			return project.Changes[i].Name < project.Changes[j].Name
		case a == "":
			return false
		case b == "":
			return true
		default:
			return a > b
		}
	})

	return project, nil
}

func (l *Loader) LoadConfigFrom(root string) (ProjectConfig, error) {
	data, err := l.fs.ReadFile(filepath.Join(root, DirOpenspec, FileConfig))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ProjectConfig{}, nil
		}
		return ProjectConfig{}, err
	}
	var raw projectConfigYAML
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return ProjectConfig{}, fmt.Errorf("openspec/config.yaml: %w", err)
	}
	return ProjectConfig{Context: strings.TrimSpace(raw.Context), Rules: raw.Rules}, nil
}

func (l *Loader) LoadProjectSpecsFrom(root string) ([]ProjectSpec, error) {
	specsDir := filepath.Join(root, DirOpenspec, DirSpecs)
	names, err := l.listDirs(specsDir, false)
	if err != nil {
		return nil, err
	}

	specs := make([]ProjectSpec, 0, len(names))
	for _, name := range names {
		ps := ProjectSpec{Name: name}
		specPath := filepath.Join(specsDir, name, FileSpec)
		data, readErr := l.fs.ReadFile(specPath)
		switch {
		case readErr == nil:
			ps.Content = string(data)
			for _, line := range splitLines(ps.Content) {
				if strings.HasPrefix(line, reqPrefix) {
					ps.RequirementCount++
					ps.RequirementNames = append(ps.RequirementNames, strings.TrimSpace(strings.TrimPrefix(line, reqPrefix)))
				}
			}
		case errors.Is(readErr, fs.ErrNotExist):
			// spec dir without a spec.md — leave empty (unchanged behavior).
		default:
			ps.ReadErr = readErr
			ps.Content = unreadablePrefix + specPath + ": " + readErr.Error()
		}
		specs = append(specs, ps)
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].Name < specs[j].Name })
	return specs, nil
}

func (l *Loader) ListChangeNamesFrom(root string) ([]string, error) {
	return l.listDirs(filepath.Join(root, DirOpenspec, DirChanges), true)
}

func (l *Loader) ListArchiveChangesFrom(root string) ([]Change, error) {
	archiveDir := filepath.Join(root, DirOpenspec, DirChanges, DirArchive)
	dirs, err := l.listDirs(archiveDir, false)
	if err != nil {
		return nil, err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dirs)))

	var changes []Change
	for _, dir := range dirs {
		cleanName, dispDate := parseArchiveName(dir)
		ch := l.loadChangeFromDir(filepath.Join(archiveDir, dir), cleanName, dispDate)
		changes = append(changes, ch)
	}
	return changes, nil
}

func (l *Loader) ListArchiveNamesFrom(root string) ([]string, error) {
	names, err := l.listDirs(filepath.Join(root, DirOpenspec, DirChanges, DirArchive), false)
	if err != nil {
		return nil, err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	return names, nil
}

func (l *Loader) ListSpecNamesFrom(root string) ([]string, error) {
	names, err := l.listDirs(filepath.Join(root, DirOpenspec, DirSpecs), false)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

// listDirs returns the names of the immediate subdirectories of path, optionally
// excluding the archive dir. A not-found path yields (nil, nil); other ReadDir
// errors propagate. The single source for the package's directory enumeration.
func (l *Loader) listDirs(path string, excludeArchive bool) ([]string, error) {
	entries, err := l.fs.ReadDir(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() || (excludeArchive && e.Name() == DirArchive) {
			continue
		}
		names = append(names, e.Name())
	}
	return names, nil
}

// ── Path-based loader ─────────────────────────────────────────────────────────

func (l *Loader) LoadFromPath(path string) (*Project, error) {
	if _, err := l.fs.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("path not found: %s", path)
	}
	if _, err := l.fs.Stat(filepath.Join(path, fileMeta)); errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("not a valid change directory (missing .openspec.yaml): %s", path)
	}

	ch := l.loadChangeFromDir(path, filepath.Base(path), "")

	project := &Project{
		Name:    filepath.Base(filepath.Dir(path)),
		Changes: []Change{ch},
	}
	return project, nil
}

func (l *Loader) ReloadChange(ch Change) Change {
	ch.Proposal = l.loadFile(filepath.Join(ch.Path, FileProposal))
	ch.Design = l.loadFile(filepath.Join(ch.Path, FileDesign))
	ch.Tasks = l.loadFile(filepath.Join(ch.Path, FileTasks))
	ch.Specs, ch.SpecFiles = l.loadSpecs(filepath.Join(ch.Path, DirSpecs))
	return ch
}

// FileSig is a cheap change-detection signature for a file: present-ness plus
// modification time (UnixNano) and size. A not-found stat yields the zero value
// (Present=false). It is comparable with ==, so callers can gate expensive
// re-reads on whether the signature changed.
type FileSig struct {
	Present     bool
	ModUnixNano int64
	Size        int64
}

// Signature returns a FileSig for path via a single stat — no file read. A stat
// error (e.g. the file does not exist) yields the zero value.
func (l *Loader) Signature(path string) FileSig {
	info, err := l.fs.Stat(path)
	if err != nil {
		return FileSig{}
	}
	return FileSig{Present: true, ModUnixNano: info.ModTime().UnixNano(), Size: info.Size()}
}

// ── Helpers ────────────────────────────────────────────────────────────────────

func (l *Loader) loadChangeFromDir(dir, name, displayDate string) Change {
	ch := Change{Name: name, Path: dir, DisplayDate: displayDate}
	if raw, err := l.fs.ReadFile(filepath.Join(dir, fileMeta)); err == nil {
		var m openspecMeta
		// Ignore unmarshal errors: .openspec.yaml is optional metadata,
		// missing or malformed fields are non-fatal.
		_ = yaml.Unmarshal(raw, &m)
		ch.Created = m.Created
	}
	ch.Proposal = l.loadFile(filepath.Join(dir, FileProposal))
	ch.Design = l.loadFile(filepath.Join(dir, FileDesign))
	ch.Tasks = l.loadFile(filepath.Join(dir, FileTasks))
	ch.Specs, ch.SpecFiles = l.loadSpecs(filepath.Join(dir, DirSpecs))
	return ch
}

func (l *Loader) loadFile(path string) Artifact {
	data, err := l.fs.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Artifact{}
		}
		// Exists but unreadable: surface it rather than vanishing.
		return Artifact{Present: true, ReadErr: err, Content: unreadablePrefix + path + ": " + err.Error()}
	}
	return Artifact{Content: string(data), Present: true}
}

func (l *Loader) loadSpecs(dir string) (Artifact, []NamedSpec) {
	names, err := l.listDirs(dir, false)
	if err != nil {
		return Artifact{}, nil
	}
	var parts []string
	var files []NamedSpec
	for _, name := range names {
		specPath := filepath.Join(dir, name, FileSpec)
		data, readErr := l.fs.ReadFile(specPath)
		switch {
		case readErr == nil:
			content := string(data)
			files = append(files, NamedSpec{Name: name, Content: content})
			parts = append(parts, "# "+name+"\n\n"+content)
		case errors.Is(readErr, fs.ErrNotExist):
			// spec dir without a spec.md — skip (unchanged behavior).
		default:
			ph := unreadablePrefix + specPath + ": " + readErr.Error()
			files = append(files, NamedSpec{Name: name, Content: ph, ReadErr: readErr})
			parts = append(parts, "# "+name+"\n\n"+ph)
		}
	}
	if len(parts) == 0 {
		return Artifact{}, nil
	}
	return Artifact{Content: strings.Join(parts, "\n\n---\n\n"), Present: true}, files
}

// archivePrefixRe matches an archived change dir like "2026-06-24-my-change":
// group 1 is the ISO date prefix, group 2 the remaining change name.
var archivePrefixRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)$`)

func parseArchiveName(dir string) (name, date string) {
	m := archivePrefixRe.FindStringSubmatch(dir)
	if m == nil {
		return dir, ""
	}
	t, err := time.Parse("2006-01-02", m[1])
	if err != nil {
		return dir, ""
	}
	return m[2], t.Format("2006-01-02")
}

func ExtractRequirement(raw, name string) string {
	lines := splitLines(raw)
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, reqPrefix) && strings.TrimSpace(strings.TrimPrefix(l, reqPrefix)) == name {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	block := []string{lines[start]}
	for _, l := range lines[start+1:] {
		if strings.HasPrefix(l, reqPrefix) {
			break
		}
		block = append(block, l)
	}
	return strings.Join(block, "\n")
}

func ConfigToMarkdown(cfg ProjectConfig) string {
	var sb strings.Builder
	if cfg.Context != "" {
		sb.WriteString("## Context\n\n")
		sb.WriteString(cfg.Context)
		sb.WriteString("\n")
	}
	if len(cfg.Rules) > 0 {
		if cfg.Context != "" {
			sb.WriteString("\n")
		}
		sb.WriteString("## Rules\n")
		keys := make([]string, 0, len(cfg.Rules))
		for k := range cfg.Rules {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteString("\n### ")
			sb.WriteString(k)
			sb.WriteString("\n\n")
			for _, item := range cfg.Rules[k] {
				sb.WriteString("- ")
				sb.WriteString(item)
				sb.WriteString("\n")
			}
		}
	}
	return sb.String()
}

// ── Backward-compatible package-level wrappers ─────────────────────────────────

func LoadFrom(root string) (*Project, error) {
	return defaultLoader.LoadFrom(root)
}

func LoadConfigFrom(root string) (ProjectConfig, error) {
	return defaultLoader.LoadConfigFrom(root)
}

func LoadProjectSpecsFrom(root string) ([]ProjectSpec, error) {
	return defaultLoader.LoadProjectSpecsFrom(root)
}

func ListChangeNamesFrom(root string) ([]string, error) {
	return defaultLoader.ListChangeNamesFrom(root)
}

func ListArchiveChangesFrom(root string) ([]Change, error) {
	return defaultLoader.ListArchiveChangesFrom(root)
}

func ListArchiveNamesFrom(root string) ([]string, error) {
	return defaultLoader.ListArchiveNamesFrom(root)
}

func ListSpecNamesFrom(root string) ([]string, error) {
	return defaultLoader.ListSpecNamesFrom(root)
}

func LoadFromPath(path string) (*Project, error) {
	return defaultLoader.LoadFromPath(path)
}

func ReloadChange(ch Change) Change {
	return defaultLoader.ReloadChange(ch)
}

// ── Zero-argument wrappers (delegate to *From with os.Getwd()) ─────────────────

func Load() (*Project, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return LoadFrom(cwd)
}

func LoadConfig() (ProjectConfig, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return ProjectConfig{}, err
	}
	return LoadConfigFrom(cwd)
}

func LoadProjectSpecs() ([]ProjectSpec, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return LoadProjectSpecsFrom(cwd)
}

func ListArchiveChanges() ([]Change, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return ListArchiveChangesFrom(cwd)
}

func ListArchiveNames() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return ListArchiveNamesFrom(cwd)
}

func ListSpecNames() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return ListSpecNamesFrom(cwd)
}

func ListChangeNames() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return ListChangeNamesFrom(cwd)
}
