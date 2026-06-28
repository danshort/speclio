# tasks-toggle Specification

## Purpose
Manages the TUI tasks tab: parsing `tasks.md`, cursor navigation between items, checkbox toggle with `Space`, per-section progress bars, and inline markdown rendering in task text.

## Requirements


### Requirement: Parse tasks.md into navigable items
The system SHALL parse `tasks.md` line by line and produce a flat list of items. Each line matching `- [ ] <text>` or `- [x] <text>` SHALL become a task item (pending or completed respectively). Lines matching a section heading (`## <text>`) SHALL become non-interactive section items. All other lines SHALL be ignored.

#### Scenario: Parsing a pending task
- **WHEN** the line is `- [ ] Initialize Go module`
- **THEN** a task item is produced with `done=false` and `text="Initialize Go module"`

#### Scenario: Parsing a completed task
- **WHEN** the line is `- [x] Initialize Go module`
- **THEN** a task item is produced with `done=true`

#### Scenario: Parsing a section
- **WHEN** the line is `## 1. Setup`
- **THEN** a section item is produced with `text="1. Setup"`, non-interactive

#### Scenario: Ignored line
- **WHEN** the line is a free-text paragraph
- **THEN** no item is produced

### Requirement: Navigate between tasks with j/k
In the `tasks` tab the cursor SHALL move between task items (not sections) with `j` (down) and `k` (up). The cursor SHALL skip sections automatically.

#### Scenario: Skip section when moving down
- **WHEN** the cursor is on the last task of section 1 and the user presses `j`
- **THEN** the cursor moves to the first task of section 2, skipping the section header

#### Scenario: Lower bound
- **WHEN** the cursor is on the last task and the user presses `j`
- **THEN** the cursor does not move

### Requirement: Toggle checkbox with Space
When pressing `Space` on a task item, the system SHALL:
1. Invert the `done` state of the item in memory
2. Re-read `tasks.md` and re-parse it, then locate the line to modify by matching the task's text against the fresh parse — never trusting a line index captured at render time
3. Modify only that line on disk, changing `[ ]` to `[x]` or vice versa, preserving the file's existing line endings
4. Update the render without reloading the entire file

If no task with the matching text is found in the fresh parse, the system SHALL make no change to the file. Task texts are expected to be unique; if the cursor's text matches more than one task in the fresh parse, the system SHALL make no change and surface a temporary error rather than toggling the first match.

#### Scenario: Mark task as completed
- **WHEN** the cursor is on `- [ ] Create structure` and the user presses `Space`
- **THEN** the line on disk becomes `- [x] Create structure` and the item shows the completed state

#### Scenario: Unmark completed task
- **WHEN** the cursor is on `- [x] Create structure` and the user presses `Space`
- **THEN** the line on disk becomes `- [ ] Create structure` and the item shows the pending state

#### Scenario: File shifted since render toggles the correct task
- **WHEN** the cursor is on `- [ ] 1.2 beta` but `tasks.md` has since gained lines above it (e.g. an agent inserted a section) so the rendered line index is now stale
- **THEN** pressing `Space` toggles the line whose text matches `1.2 beta`, not whatever line now sits at the old index

#### Scenario: Matching task no longer present
- **WHEN** the cursor task's text no longer exists in `tasks.md` at the moment of toggling
- **THEN** no write occurs and the file is left unchanged

#### Scenario: Ambiguous match makes no change
- **WHEN** the cursor's text matches more than one task in the fresh parse
- **THEN** no write occurs and the TUI shows a temporary error prompting disambiguation

#### Scenario: Write error
- **WHEN** `tasks.md` does not have write permissions and the user presses `Space`
- **THEN** the toggle is not applied and the TUI shows a temporary error message

### Requirement: Per-section progress bar
The TUI SHALL show a progress bar next to each section header along with the `completed/total` count of tasks in that section.

#### Scenario: Partially completed section
- **WHEN** a section has 2 completed tasks out of 5
- **THEN** `██░░░ 2/5` is shown next to the section header

#### Scenario: Fully completed section
- **WHEN** all tasks in a section are completed
- **THEN** the bar appears completely filled

### Requirement: Visual cursor indicator
The task item under the cursor SHALL be shown with a distinct visual indicator (e.g. `▶`) to differentiate it from the rest.

#### Scenario: Cursor on a task
- **WHEN** the cursor is on task N
- **THEN** that task shows the `▶` prefix and a differentiated visual style

### Requirement: Restore cursor by text after reload
When `tasks.md` is reloaded from disk, the system SHALL attempt to restore the cursor to the task whose text matches the text of the task that had the cursor before the reload. If the text is not found in the new list, the cursor SHALL be positioned on the first available task item.

#### Scenario: Task under the cursor still exists after reload
- **WHEN** the cursor was on the task with text `"1.3 Create structure"` and the reload does not remove that task
- **THEN** the cursor is positioned on the same task `"1.3 Create structure"`

#### Scenario: Task under the cursor removed during the reload
- **WHEN** the cursor was on a task that no longer exists in the new `tasks.md`
- **THEN** the cursor is positioned on the first available task item in the new list

### Requirement: Inline markdown rendering in task items
The TUI SHALL convert inline markdown marks present in the text of each task to ANSI styles before rendering the item with lipgloss. The supported patterns are `` `code` `` (backtick) and `**bold**` (double asterisk).

#### Scenario: Task with a code fragment
- **WHEN** the text of a task item contains `` `func main()` ``
- **THEN** the fragment is shown with the visual code style (distinct background or color) in the TUI

#### Scenario: Task with bold text
- **WHEN** the text of a task item contains `**important**`
- **THEN** the word is shown in bold in the TUI

#### Scenario: Multiple fragments in the same task
- **WHEN** the text of an item contains several `` `code` `` or `**bold**` fragments separated from each other
- **THEN** each fragment is rendered with its corresponding style independently

#### Scenario: Task without inline markdown
- **WHEN** the text of an item contains no backticks or double asterisks
- **THEN** the text is shown unchanged, without visual artifacts

### Requirement: Viewport scroll follows cursor
When navigating the tasks view, the viewport SHALL always scroll to keep the cursor-selected task fully visible, correctly accounting for task items that wrap across multiple terminal lines.

#### Scenario: Cursor moves below visible area
- **WHEN** the user navigates down and the selected task is below the bottom of the visible viewport
- **THEN** the viewport SHALL scroll down so the selected task is visible

#### Scenario: Cursor moves above visible area
- **WHEN** the user navigates up and the selected task is above the top of the visible viewport
- **THEN** the viewport SHALL scroll up so the selected task is visible

#### Scenario: Task text wraps across multiple lines
- **WHEN** a task's text is long enough that lipgloss renders it across more than one terminal line
- **THEN** the line counter SHALL advance by the actual rendered height of that item, not by 1

#### Scenario: Tasks beyond the initial visible area are reachable
- **WHEN** the task list contains more items than fit in the visible viewport height
- **THEN** navigating down with `j` SHALL eventually reach every task in the list
