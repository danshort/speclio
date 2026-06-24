## ADDED Requirements

### Requirement: Legible secondary text
Secondary ("minimized") text — help/footer text, empty-state messages such as "No active changes", requirement count labels, and completed tasks — SHALL use a foreground color that is readable on a standard dark terminal while remaining visually subordinate to primary text.

#### Scenario: Help and footer text is readable
- **WHEN** the TUI renders the help bar or an empty-state message
- **THEN** the text uses the minimized color `245` (a mid-gray), not ANSI `8` (bright black)

#### Scenario: Requirement labels are readable
- **WHEN** the index renders a project spec's requirement-count label
- **THEN** the label uses the minimized color and is legible against the background

#### Scenario: Completed tasks remain de-emphasized but readable
- **WHEN** the Tasks tab renders a completed (`- [x]`) task
- **THEN** the task text uses the minimized color, dimmer than pending tasks yet legible
