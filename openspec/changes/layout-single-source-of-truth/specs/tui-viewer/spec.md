## ADDED Requirements

### Requirement: Viewport height derived from the rendered chrome rows

The TUI SHALL size the content viewport from the same ordered list of chrome rows that `View()` renders, rather than from an independently maintained sum of layout constants. In every mode that renders a content viewport, the total number of rendered lines SHALL equal the terminal height. The empty-project welcome view, which renders fixed content without a sized viewport, is the single exception.

#### Scenario: Total rendered height equals terminal height

- **WHEN** the TUI renders a viewport-backed mode (help overlay closed) at a terminal height at least as large as that mode's chrome-row count
- **THEN** the number of rendered lines equals the terminal height, with no clipped row or trailing blank line

#### Scenario: Below minimum height the viewport clamps

- **WHEN** the terminal height is smaller than the mode's chrome-row count
- **THEN** the viewport height clamps to one row (rendered content necessarily exceeds the terminal), and this degenerate case is exempt from the height-equality invariant

#### Scenario: Height invariant holds across every viewport-backed mode

- **WHEN** the TUI renders in normal (with active changes), archive, index, spec, config, or worktrees mode
- **THEN** in each mode the total rendered height equals the terminal height

#### Scenario: Empty-project welcome view is exempt

- **WHEN** there are no active changes and the welcome view is shown
- **THEN** the view renders fixed welcome content without a sized viewport, and the height-invariant test excludes this mode rather than asserting against it

#### Scenario: Optional spec subnav row is accounted for

- **WHEN** the specs tab is active and the spec subnav row is present
- **THEN** the viewport is exactly one row shorter and the total rendered height still equals the terminal height
