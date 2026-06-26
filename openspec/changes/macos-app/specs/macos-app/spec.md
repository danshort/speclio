## ADDED Requirements

### Requirement: Native macOS reader for OpenSpec artifacts
The project SHALL provide a native macOS application that reads an OpenSpec project from the same `openspec/` layout the TUI reads and presents its changes and artifacts, without modifying or depending on the TUI.

#### Scenario: Open a project and browse changes
- **WHEN** the user opens a directory containing an `openspec/` folder
- **THEN** the app lists the project's active changes and lets the user navigate into each change's proposal, design, tasks, and specs

#### Scenario: TUI unaffected
- **WHEN** the macOS app is built and run
- **THEN** the Go TUI's behavior and code are unchanged

### Requirement: Top-level navigation modes
The app SHALL present a top-level mode switcher (a segmented control in the window toolbar) that switches between dedicated views for **Active Changes**, **Archived Changes**, **Specs**, and **Worktrees**. The selected mode determines the contents of both the sidebar list and the detail pane.

#### Scenario: Switching modes
- **WHEN** the user selects a mode in the switcher
- **THEN** the sidebar and detail update to that mode's content (active changes, archived changes, project specs, or worktrees)

#### Scenario: Active Changes is the default
- **WHEN** a project is opened
- **THEN** the app starts in the Active Changes mode

### Requirement: Project specs index
In the Specs mode, the app SHALL list the project's long-lived specs from `openspec/specs/` and render a selected spec with the same markdown and validation behavior as change artifacts.

#### Scenario: Browse project specs
- **WHEN** the user selects the Specs mode
- **THEN** the app lists the capability specs under `openspec/specs/` and renders the selected spec's `spec.md`

### Requirement: Archived changes browsing
In the Archived Changes mode, the app SHALL list the project's archived changes from `openspec/changes/archive/`, separate from active changes, and let the user open their artifacts.

#### Scenario: Browse archived changes
- **WHEN** the user selects the Archived Changes mode
- **THEN** the app lists the archived changes (newest first) and lets the user open each one's artifacts

### Requirement: Faithful domain behavior via a shared contract
The app SHALL obtain changes, tasks, validation results, and worktree data through a domain layer whose behavior matches the Go implementation as enforced by the shared fixture corpus.

#### Scenario: Loader parity
- **WHEN** the app parses a project that is also present in the shared corpus
- **THEN** the parsed result matches the corpus golden output, identical to the Go loader

### Requirement: Rendered markdown with unreadable-artifact handling
The app SHALL render artifact markdown natively — including tables, fenced code blocks, and nested lists — and SHALL surface an artifact that exists but cannot be read as a placeholder rather than as missing.

#### Scenario: Readable artifact renders
- **WHEN** an artifact file is present and readable
- **THEN** its markdown is displayed as formatted content, with tables, code blocks, and nested lists rendered

#### Scenario: Unreadable artifact is flagged
- **WHEN** an artifact file exists but cannot be read
- **THEN** the app shows a placeholder indicating the read failure, not an absent artifact

### Requirement: Inline validation banner on specs
When a spec or change has structural validation errors, the app SHALL surface them inline with the rendered artifact, mirroring the TUI, and SHALL omit the banner for an unreadable artifact (a read failure is not a validation failure).

#### Scenario: Spec with validation errors shows a banner
- **WHEN** a rendered spec fails structural validation
- **THEN** the app displays the validation messages together with the artifact content

#### Scenario: Unreadable artifact shows no validation banner
- **WHEN** an artifact exists but cannot be read
- **THEN** the app shows the read-failure placeholder and no validation banner

### Requirement: Requirement focus and navigation
The app SHALL let the user focus a single requirement within a spec and navigate to it, mirroring the TUI's requirement-extraction and jump behavior.

#### Scenario: Focus a single requirement
- **WHEN** the user selects a requirement in a spec
- **THEN** the app presents that requirement's block (from its `### Requirement:` heading to the next) and scrolls it into view

### Requirement: Native accessibility
The app SHALL meet baseline macOS accessibility expectations: VoiceOver labels, keyboard navigation of the sidebar and content, and Dynamic Type / contrast support.

#### Scenario: Keyboard-only navigation
- **WHEN** the user navigates with the keyboard only
- **THEN** they can move between the sidebar, artifact list, and content, and toggle a task, without a pointer

#### Scenario: VoiceOver reads the interface
- **WHEN** VoiceOver is enabled
- **THEN** changes, artifacts, and controls expose meaningful accessibility labels

### Requirement: Task toggling that preserves line endings
The app SHALL let the user toggle a task checkbox in `tasks.md`, writing the change to disk while preserving the file's existing line endings. Because the app reloads on disk changes, it SHALL re-read and re-parse `tasks.md` immediately before writing so a stale line index cannot toggle the wrong line.

#### Scenario: Toggle a task
- **WHEN** the user toggles a task in the app
- **THEN** the corresponding `- [ ]`/`- [x]` marker is flipped in `tasks.md` and the file's original line endings (LF or CRLF) are preserved

#### Scenario: File changed on disk before toggle
- **WHEN** `tasks.md` was modified by another process after it was rendered and the user then toggles a task
- **THEN** the app re-reads the current file before writing, so the intended task is toggled rather than a stale line

### Requirement: Worktrees overview
The app SHALL present the git worktrees of the project's repository, and SHALL degrade gracefully when git is unavailable or the directory is not a working tree.

#### Scenario: Worktrees listed
- **WHEN** the project is inside a git repository with multiple worktrees
- **THEN** the app lists the worktrees and marks the current one

#### Scenario: Git unavailable
- **WHEN** git is not on PATH or the directory is not a git working tree
- **THEN** the app shows an unavailable state instead of failing

### Requirement: Live reload on disk changes
The app SHALL reflect on-disk changes to the open project's `openspec/` tree without requiring a manual refresh.

#### Scenario: External edit refreshes the view
- **WHEN** a file under the open project's `openspec/` directory is modified by another process
- **THEN** the app updates its view to reflect the change
