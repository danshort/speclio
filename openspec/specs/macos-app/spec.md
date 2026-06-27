# macos-app Specification

## Purpose
TBD - created by archiving change macos-app. Update Purpose after archive.
## Requirements
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

### Requirement: Project configuration view
When the project has a non-empty `openspec/config.yaml`, the Specs mode SHALL offer a "Project Config" entry that renders the configuration (context and rules) as formatted markdown.

#### Scenario: View project configuration
- **WHEN** the project has configuration content and the user selects the Project Config entry in the Specs mode
- **THEN** the app renders the context and rules as markdown

#### Scenario: No configuration
- **WHEN** the project has no `openspec/config.yaml` (or it is empty)
- **THEN** no Project Config entry is shown

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
The app SHALL present the git worktrees of the project's repository, the current worktree first and labeled with its state (current, detached, locked, prunable), and for each non-bare worktree SHALL surface its active changes with a task-progress indicator (completed / total). A non-bare worktree that has an `openspec/` project but no active changes SHALL show a "no active changes" affordance, distinct from a worktree with no project. Selecting a change under a worktree SHALL open it read-only. The view SHALL degrade gracefully when git is unavailable, a worktree is bare, or a worktree has no `openspec/` project.

#### Scenario: Worktrees listed with state
- **WHEN** the project is inside a git repository with multiple worktrees
- **THEN** the app lists the worktrees, current first, each labeled with its state (current, detached, locked, prunable as applicable)

#### Scenario: Per-worktree active changes and progress
- **WHEN** a non-bare worktree has active changes
- **THEN** each change is listed under that worktree with a completed/total task-progress indicator

#### Scenario: Worktree with a project but no active changes
- **WHEN** a non-bare worktree has an `openspec/` project but no active changes
- **THEN** the worktree shows a "no active changes" affordance rather than appearing empty

#### Scenario: Open a worktree's change read-only
- **WHEN** the user selects a change under a worktree
- **THEN** the app renders that change's artifacts read-only (no task toggling)

#### Scenario: Worktree without an OpenSpec project
- **WHEN** a worktree is bare or has no `openspec/` directory
- **THEN** it is listed with no changes, not as an error

#### Scenario: Git unavailable
- **WHEN** git is not on PATH or the directory is not a git working tree
- **THEN** the app shows an unavailable state instead of failing

### Requirement: Live reload on disk changes
The app SHALL reflect on-disk changes to the open project's `openspec/` tree without requiring a manual refresh, and a reload SHALL preserve the current selection when it still exists.

#### Scenario: External edit refreshes the view
- **WHEN** a file under the open project's `openspec/` directory is modified by another process
- **THEN** the app updates its view to reflect the change, keeping the current selection

### Requirement: Distributable signed build
The app SHALL be packaged as a `.app` bundle distributed via a Homebrew **cask** alongside the CLI. Signing is phased: until a Developer-ID certificate is available it SHALL be at least **ad-hoc** code-signed (required to run on Apple Silicon); once available it SHALL be **Developer-ID** signed (hardened runtime) and **notarized**. While the distributed build is unnotarized, the install instructions SHALL document the one-time Gatekeeper step; that guidance SHALL be removed once notarized builds ship.

#### Scenario: Build runs on Apple Silicon
- **WHEN** the app is packaged
- **THEN** the resulting `.app` carries at least an ad-hoc code signature and launches on Apple Silicon

#### Scenario: First-launch Gatekeeper guidance while unnotarized
- **WHEN** a user installs an unnotarized build
- **THEN** the install instructions document the one-time right-click → Open (or quarantine-removal) step

#### Scenario: Notarized build once the certificate exists
- **WHEN** a Developer-ID certificate and notary credentials are configured
- **THEN** the release produces a Developer-ID-signed, notarized, stapled build and the Gatekeeper caveat is dropped

### Requirement: Reveal and open the selected file
The app SHALL let the user reveal the currently selected item (artifact, project spec, config, or worktree) in Finder and open it in the system default application.

#### Scenario: Reveal in Finder
- **WHEN** the user chooses "Reveal in Finder" for a selection backed by a file or directory
- **THEN** Finder opens with that item selected

#### Scenario: Open in default app
- **WHEN** the user chooses "Open in Default App" for a file-backed selection
- **THEN** the file opens in its default application

### Requirement: Persistent change progress bar
When the current selection belongs to a change (an active, archived, or worktree change) that has tasks, the detail pane SHALL show a persistent progress bar at the top reflecting that change's overall task completion, visible regardless of which of the change's artifacts is open. The bar SHALL be hidden for selections that are not changes (project specs, project config, worktree metadata) and for changes with no tasks.

#### Scenario: Progress visible across a change's artifacts
- **WHEN** the user views any artifact (proposal, design, a spec, or tasks) of a change that has tasks
- **THEN** a progress bar at the top of the detail pane shows that change's overall completed/total

#### Scenario: Hidden for non-change views
- **WHEN** the selection is a project spec, the project config, or worktree metadata (or a change with no tasks)
- **THEN** no persistent change progress bar is shown

### Requirement: Per-section task progress
In the Tasks view, each section heading SHALL show that section's task completion (completed/total of the tasks under it) alongside the heading. The Tasks view SHALL NOT show a separate overall progress bar — the change's overall progress is the persistent bar at the top of the detail pane.

#### Scenario: Section progress shown
- **WHEN** a tasks file has sections with tasks
- **THEN** each section heading displays the completed/total for the tasks in that section

#### Scenario: No overall bar in the Tasks view
- **WHEN** the Tasks view is shown
- **THEN** it shows per-section progress only, not a separate overall progress bar (overall progress is the persistent detail-pane bar)

### Requirement: Change progress in the sidebar
Every change row in the sidebar — active, archived, and worktree changes — SHALL show its task progress (completed/total), so progress is visible without opening the change.

#### Scenario: Active and archived changes show progress
- **WHEN** the sidebar lists active or archived changes that have tasks
- **THEN** each change row shows its completed/total task progress, consistent with worktree-change rows

### Requirement: App icon
The packaged `.app` SHALL bundle the project's app icon, compiled from the Icon Composer source (`lectern.icon`) into the bundle and referenced by `Info.plist`, so the app shows proper branding in Finder, the Dock, and About. When the build toolchain cannot compile the icon format, packaging SHALL warn and still produce a working (icon-less) build rather than fail.

#### Scenario: Packaged app includes its icon
- **WHEN** the app is packaged on a toolchain that supports the icon format
- **THEN** the bundle contains the compiled icon (`Assets.car` + `lectern.icns`) and `Info.plist` references it via `CFBundleIconName`

#### Scenario: Older toolchain degrades gracefully
- **WHEN** the build toolchain cannot compile the `.icon`
- **THEN** packaging warns and still produces a working build without the icon, rather than failing

### Requirement: Settings window
The app SHALL provide a standard macOS Settings window, opened via the "Settings…" menu item and ⌘,, as the home for user preferences. Preferences set there SHALL persist across launches.

#### Scenario: Open settings
- **WHEN** the user chooses Settings… (or presses ⌘,)
- **THEN** the app opens a Settings window with its preferences

#### Scenario: Preferences persist
- **WHEN** the user changes a preference and relaunches the app
- **THEN** the preference retains its value

### Requirement: Adjustable content font size
The app SHALL let the user adjust the size of rendered content (the artifact/markdown reading pane) via a font-size preference, applied as a multiplier on top of the system Dynamic Type size. It SHALL be adjustable from the Settings window and via keyboard shortcuts (⌘+ increase, ⌘− decrease, ⌘0 reset), and SHALL apply only to rendered content — not the sidebar or app chrome.

#### Scenario: Adjust from settings
- **WHEN** the user changes the content text-size control in Settings
- **THEN** rendered content resizes accordingly and the change persists

#### Scenario: Adjust via keyboard
- **WHEN** the user presses ⌘+, ⌘−, or ⌘0
- **THEN** the rendered content size increases, decreases, or resets, consistent with the Settings control (they share one stored value)

#### Scenario: Content-only scope
- **WHEN** the content font size is changed
- **THEN** only the rendered content resizes; the sidebar, toolbar, and other chrome are unaffected

### Requirement: Persistent project label
The app SHALL keep the open project's name visible across all navigation states, so the user always knows which project they are viewing. The project name SHALL be shown as a header at the top of the sidebar AND as the window title, with the current location shown as the window subtitle.

#### Scenario: Project name in the sidebar header
- **WHEN** a project is open
- **THEN** the project's name is shown as a prominent header at the top of the sidebar, above the navigation list, regardless of the current mode or selection

#### Scenario: Project name in the window title
- **WHEN** a project is open
- **THEN** the window title shows the project's name, and the window subtitle shows the current location (mode, and the selected change/spec/artifact where applicable)

#### Scenario: Persists across navigation
- **WHEN** the user switches modes (Active, Archived, Specs, Worktrees) or selects different items
- **THEN** the project name remains visible in both the sidebar header and the window title; only the window subtitle changes to reflect the new location

#### Scenario: No project open
- **WHEN** no project is open
- **THEN** no project header or project title is shown, and the empty state communicates that no project is open

### Requirement: Multiple project windows
The app SHALL allow more than one project to be open at the same time, each in its own window (and macOS window tab), with independent navigation state (mode, selection, sidebar, and live reload). Opening a project SHALL NOT close or replace any already-open project.

#### Scenario: Open a second project without closing the first
- **WHEN** a project is open and the user opens another project
- **THEN** the second project opens in its own window and the first remains open and unchanged

#### Scenario: Independent navigation per window
- **WHEN** two project windows are open
- **THEN** changing the mode or selection in one window does not affect the other

#### Scenario: Window tabs hold different projects
- **WHEN** the user creates a new tab and opens a project in it
- **THEN** that tab shows the newly opened project, independent of the other tabs

### Requirement: Open into an empty window or a new one
The app SHALL open a chosen project into the current window when that window has no project loaded, and otherwise SHALL open it in a new window, so that opening a project never evicts another. Each project SHALL have at most one window: opening a project that is already open SHALL focus its existing window rather than creating a duplicate.

#### Scenario: Reuse an empty window
- **WHEN** the focused window has no project loaded and the user opens a project
- **THEN** the project loads into that window and no stray empty window remains

#### Scenario: Open in a new window when the current one is in use
- **WHEN** the focused window already has a project loaded and the user opens another project
- **THEN** the project opens in a new window and the focused window is unchanged

#### Scenario: Opening an already-open project focuses it
- **WHEN** the user opens a project that is already open in another window
- **THEN** that existing window is brought to the front and no duplicate window is created

### Requirement: Open Recent
The app SHALL provide a File ▸ Open Recent menu listing recently opened projects, most-recent first. Selecting an entry SHALL open the project (or focus it if already open). The menu SHALL include a Clear Menu action.

#### Scenario: Reopen a recent project
- **WHEN** the user chooses a project from Open Recent
- **THEN** the project opens in a window, or its existing window is focused if it is already open

#### Scenario: Recents update on open
- **WHEN** the user opens a project
- **THEN** it appears at the top of the Open Recent menu, without duplicate entries

#### Scenario: Clear the menu
- **WHEN** the user chooses Clear Menu
- **THEN** the Open Recent list is emptied, and any currently-open project windows are unaffected

### Requirement: Reopen open projects on launch
The app SHALL remember the set of projects that were open when it last quit and reopen a window for each on the next launch.

#### Scenario: Restore multiple projects
- **WHEN** the user quits with two or more projects open and relaunches the app
- **THEN** a window reopens for each previously-open project

#### Scenario: Unresolvable project on restore
- **WHEN** a remembered project can no longer be resolved (moved, deleted, or permission revoked)
- **THEN** the app still launches and that window shows the empty/error state instead of failing

### Requirement: Per-window commands
Menu and toolbar commands that act on a project SHALL act on the focused window's project, so commands invoked in one window do not affect another. The content text-size preference SHALL remain global across all windows.

#### Scenario: Reload affects only the focused window
- **WHEN** the user invokes Reload (or a file action) with multiple windows open
- **THEN** only the focused window's project is reloaded / acted upon

#### Scenario: Text size applies everywhere
- **WHEN** the user changes the content text size
- **THEN** the change applies to rendered content in all open windows
