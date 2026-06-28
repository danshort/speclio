package ui

import (
	"image/color"
	"os"
	"path/filepath"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"github.com/danshort/lectern/internal/openspec"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeIndex
	ModeViewingArchive
	ModeViewingSpec
	ModeViewingConfig
	ModeWorktrees
)

type Tab int

const (
	TabProposal Tab = iota
	TabDesign
	TabSpecs
	TabTasks
	tabCount
)

var tabLabels = [tabCount]string{"proposal", "design", "specs", "tasks"}

type errClearMsg struct{}

// editorReturnMsg is posted when a terminal editor exits. err is non-nil when
// the editor failed to launch or run, so the handler can surface it.
type editorReturnMsg struct{ err error }

// renderedMsg carries async glamour output back to the event loop.
type renderedMsg struct {
	tab     Tab
	content string
}

// specRenderedMsg carries async glamour output for ModeViewingSpec.
type specRenderedMsg struct {
	content  string
	jumpLine int // line offset to scroll to after render; 0 = start of document
}

// renderedConfigMsg carries async glamour output for ModeViewingConfig.
type renderedConfigMsg struct {
	content string
}

type tickMsg time.Time

type indexState struct {
	Items            []indexItem
	Cursor           int
	ExpandedSpecs    map[int]bool
	ExpandedArchives map[int]bool
	SortBySuffix     bool
	Order            []int
	ArchiveChanges   []openspec.Change
	ArchiveCursor    int

	FilterText     string
	FilterActive   bool
	FilterIndices  []int
	PrevFilterText string

	// lineMap maps a rendered content line to the raw index item on it,
	// captured by renderIndexContent. indexItemAtContentLine is a lookup into
	// it, so click hit-testing can't drift from what was rendered.
	lineMap map[int]int
}

// worktreeEntry pairs a discovered git worktree with its loaded active changes.
type worktreeEntry struct {
	wt      openspec.Worktree
	changes []openspec.Change
}

// worktreeItem is a single navigable row in the worktrees view: a change at
// changeIdx within entry wtIdx.
type worktreeItem struct {
	wtIdx     int
	changeIdx int
}

type worktreesState struct {
	Entries   []worktreeEntry
	Items     []worktreeItem
	Cursor    int
	Available bool   // false when git is unavailable / not a working tree
	Message   string // explanatory line shown when Available is false
}

type specViewerState struct {
	Cursor     int
	JumpTarget string
	FocusMode  bool
	ReqCursor  int
}

type taskState struct {
	Items  []openspec.TaskItem
	Cursor int
}

type indexItemKind int

const (
	indexKindActive indexItemKind = iota
	indexKindArchived
	indexKindSpec
	indexKindRequirement
	indexKindArchivedArtifact
)

type indexItem struct {
	kind indexItemKind
	idx  int // into project.Changes (active), archiveChanges (archived/archivedArtifact), or projectSpecs (spec/requirement)
	// reqIdx carries an index into projectSpecs[idx].RequirementNames for
	// indexKindRequirement, and the target Tab for indexKindArchivedArtifact.
	reqIdx int
}

type Theme struct {
	ViewBg color.Color
}

// viewerState is the active/archive change-viewing position: which change,
// which artifact tab, and which spec file within the specs tab. Grouped so the
// fields that form invalid mode/tab/cursor combinations live together and
// setMode can clamp them as a unit.
type viewerState struct {
	changeIdx int // into project.Changes (ModeNormal)
	tab       Tab
	specIdx   int // active spec on TabSpecs (ModeNormal + ModeViewingArchive)
}

type Model struct {
	root   string
	loader *openspec.Loader

	project *openspec.Project
	viewer  viewerState

	vp      viewport.Model
	vpReady bool

	tasks taskState

	errMsg     string
	loading    bool
	singlePath bool

	// editorOpenWith is the resolved editor.open_with config value driving how
	// `e` opens artifacts (see internal/config.ResolveOpener). Empty → $EDITOR/vi.
	editorOpenWith string

	width, height int

	renderCache     map[Tab]string
	glamourRenderer *glamour.TermRenderer
	lastRenderWidth int

	mode          Mode
	prevMode      Mode
	helpOpen      bool
	index         indexState
	projectSpecs  []openspec.ProjectSpec
	spec          specViewerState
	projectConfig openspec.ProjectConfig
	theme         Theme

	worktrees worktreesState
	// worktreeViewChange holds a foreign worktree's change while it is open
	// read-only via the ModeViewingArchive path; viewingWorktreeChange gates it.
	worktreeViewChange    openspec.Change
	viewingWorktreeChange bool

	// Per-mode freshness caches: skip the periodic ReloadChange for a change
	// whose tasks.md signature is unchanged (#90). Index and worktrees keep
	// separate caches because they hold independent in-memory copies.
	fresh         *freshness
	worktreeFresh *freshness
}

func New(project *openspec.Project, cfg openspec.ProjectConfig, root string, loader *openspec.Loader, editorOpenWith, startupWarn string) Model {
	m := Model{
		root:           root,
		loader:         loader,
		project:        project,
		renderCache:    make(map[Tab]string),
		projectConfig:  cfg,
		theme:          Theme{},
		editorOpenWith: editorOpenWith,
		errMsg:         startupWarn, // e.g. a rejected user config — visible in the status line
		fresh:          newFreshness(),
		worktreeFresh:  newFreshness(),
	}
	if len(project.Changes) > 0 {
		m.viewer.tab = m.defaultTab()
		m.loadTaskItems()
	} else {
		var archiveErr error
		m.index.ArchiveChanges, archiveErr = loader.ListArchiveChangesFrom(root)
		if archiveErr != nil {
			m.errMsg = "error loading archive changes: " + archiveErr.Error()
		}
		var specErr error
		m.projectSpecs, specErr = loader.LoadProjectSpecsFrom(root)
		if specErr != nil {
			m.errMsg = "error loading project specs: " + specErr.Error()
		}
		m.index.ExpandedSpecs = make(map[int]bool)
		m.index.ExpandedArchives = make(map[int]bool)
		m.buildIndexItems()
		m.setMode(ModeIndex)
	}
	return m
}

func NewSinglePath(project *openspec.Project, cfg openspec.ProjectConfig, root string, loader *openspec.Loader, editorOpenWith, startupWarn string) Model {
	m := New(project, cfg, root, loader, editorOpenWith, startupWarn)
	m.singlePath = true
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) View() tea.View {
	if !m.vpReady {
		return tea.NewView("")
	}

	var content string
	if len(m.project.Changes) == 0 && m.mode == ModeNormal {
		content = m.emptyViewContent()
	} else {
		content = m.viewportLayout()
	}

	if m.helpOpen {
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.renderHelpOverlay())
	}

	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.BackgroundColor = m.theme.ViewBg
	return v
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (m *Model) current() *openspec.Change {
	if m.mode == ModeViewingArchive {
		return m.currentArchive()
	}
	if m.mode == ModeIndex || m.mode == ModeViewingSpec {
		return nil
	}
	if len(m.project.Changes) == 0 {
		return nil
	}
	return &m.project.Changes[m.viewer.changeIdx]
}

// hasMultipleSpecs reports whether the current change has more than one spec
// file — the precondition for spec sub-navigation (`[` / `]` and chip clicks).
func (m *Model) hasMultipleSpecs() bool {
	ch := m.current()
	return ch != nil && len(ch.SpecFiles) > 1
}

func firstAvailableTab(ch openspec.Change) Tab {
	if ch.Proposal.Present {
		return TabProposal
	}
	if ch.Design.Present {
		return TabDesign
	}
	if ch.Specs.Present {
		return TabSpecs
	}
	if ch.Tasks.Present {
		return TabTasks
	}
	return TabProposal
}

// archiveArtifactTabs returns the artifact tabs present on ch, in tab order.
// It mirrors the presence checks used by firstAvailableTab.
func archiveArtifactTabs(ch openspec.Change) []Tab {
	var tabs []Tab
	if ch.Proposal.Present {
		tabs = append(tabs, TabProposal)
	}
	if ch.Design.Present {
		tabs = append(tabs, TabDesign)
	}
	if ch.Specs.Present {
		tabs = append(tabs, TabSpecs)
	}
	if ch.Tasks.Present {
		tabs = append(tabs, TabTasks)
	}
	return tabs
}

func (m *Model) tabAvailable(t Tab) bool {
	ch := m.current()
	if ch == nil {
		return false
	}
	switch t {
	case TabProposal:
		return ch.Proposal.Present
	case TabDesign:
		return ch.Design.Present
	case TabTasks:
		return ch.Tasks.Present
	case TabSpecs:
		return ch.Specs.Present
	}
	return false
}

func (m *Model) defaultTab() Tab {
	for t := Tab(0); t < tabCount; t++ {
		if m.tabAvailable(t) {
			return t
		}
	}
	return TabProposal
}

func (m *Model) nextAvailableTab(current Tab, delta int) Tab {
	next := current
	for range int(tabCount) {
		next = Tab((int(next) + delta + int(tabCount)) % int(tabCount))
		if m.tabAvailable(next) {
			return next
		}
	}
	return current
}

func (m *Model) artifactPath() string {
	ch := m.current()
	if ch == nil {
		return ""
	}
	switch m.viewer.tab {
	case TabProposal:
		return filepath.Join(ch.Path, openspec.FileProposal)
	case TabDesign:
		return filepath.Join(ch.Path, openspec.FileDesign)
	case TabTasks:
		return filepath.Join(ch.Path, openspec.FileTasks)
	case TabSpecs:
		if m.viewer.specIdx < len(ch.SpecFiles) {
			specsDir := filepath.Join(ch.Path, openspec.DirSpecs)
			entries, err := os.ReadDir(specsDir)
			if err != nil {
				return ""
			}
			dirIdx := 0
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				if dirIdx == m.viewer.specIdx {
					p := filepath.Join(specsDir, e.Name(), openspec.FileSpec)
					if _, err := os.Stat(p); err == nil {
						return p
					}
					return ""
				}
				dirIdx++
			}
		}
	}
	return ""
}

// currentSpecPath returns the on-disk path of the project spec currently being
// viewed in ModeViewingSpec, or "" if the cursor is out of range. Requirements
// are sections within a single spec.md, so focus mode resolves to the same file.
func (m *Model) currentSpecPath() string {
	if m.spec.Cursor < 0 || m.spec.Cursor >= len(m.projectSpecs) {
		return ""
	}
	return filepath.Join(m.root, openspec.DirOpenspec, openspec.DirSpecs, m.projectSpecs[m.spec.Cursor].Name, openspec.FileSpec)
}

func (m *Model) currentArchive() *openspec.Change {
	if m.viewingWorktreeChange {
		return &m.worktreeViewChange
	}
	if m.index.ArchiveCursor < len(m.index.ArchiveChanges) {
		return &m.index.ArchiveChanges[m.index.ArchiveCursor]
	}
	return nil
}

func (m *Model) contentHeight() int {
	// Derived from the same chrome-row list View() renders, so the viewport can
	// never drift out of sync with the surrounding chrome.
	h := m.height - m.chromeRowCount()
	if h < 1 {
		h = 1
	}
	return h
}

// innerWidth is the content width inside the box's two vertical border columns.
func (m *Model) innerWidth() int {
	if m.width < 2 {
		return 0
	}
	return m.width - 2
}

// mergeReloadedChange updates in-memory state from a freshly reloaded Change
// and returns which artifacts changed. It does not handle cursor preservation
// or viewport refresh — the caller handles those.
func (m *Model) mergeReloadedChange(fresh openspec.Change) (tasksChanged bool, viewportDirty bool) {
	ch := m.current()
	if ch == nil {
		return false, false
	}

	if fresh.Tasks.Present != ch.Tasks.Present || fresh.Tasks.Content != ch.Tasks.Content {
		m.project.Changes[m.viewer.changeIdx].Tasks = fresh.Tasks
		m.tasks.Items = openspec.ParseTasks(fresh.Tasks.Content)
		tasksChanged = true
	}
	if fresh.Proposal.Present != ch.Proposal.Present || fresh.Proposal.Content != ch.Proposal.Content {
		m.project.Changes[m.viewer.changeIdx].Proposal = fresh.Proposal
		delete(m.renderCache, TabProposal)
		if m.viewer.tab == TabProposal {
			viewportDirty = true
		}
	}
	if fresh.Design.Present != ch.Design.Present || fresh.Design.Content != ch.Design.Content {
		m.project.Changes[m.viewer.changeIdx].Design = fresh.Design
		delete(m.renderCache, TabDesign)
		if m.viewer.tab == TabDesign {
			viewportDirty = true
		}
	}
	if fresh.Specs.Present != ch.Specs.Present || fresh.Specs.Content != ch.Specs.Content {
		m.project.Changes[m.viewer.changeIdx].Specs = fresh.Specs
		m.project.Changes[m.viewer.changeIdx].SpecFiles = fresh.SpecFiles
		if m.viewer.specIdx >= len(fresh.SpecFiles) {
			m.viewer.specIdx = 0
		}
		delete(m.renderCache, TabSpecs)
		if m.viewer.tab == TabSpecs {
			viewportDirty = true
		}
	}
	return
}
