## ADDED Requirements

### Requirement: Hit-testing derives from the rendered layout

The TUI SHALL resolve mouse clicks using layout data emitted by the renderer — tab x-ranges produced by the tab bar and a line→item map produced by the index content — rather than re-deriving the geometry independently in the click handler. Tab label widths SHALL be measured by rendered display width (`lipgloss.Width`), not byte length. A click that does not fall on a rendered element SHALL resolve to nothing.

#### Scenario: Every rendered tab maps back to itself

- **WHEN** the tab bar is rendered and a left-click lands within a tab's emitted x-range
- **THEN** the resolved tab is exactly the one rendered at that x-range

#### Scenario: Every rendered index item maps back to itself

- **WHEN** an index item is rendered at content line L
- **THEN** a left-click at content line L resolves to that same item, so render position and hit-test agree by construction across empty, populated, expanded, and filtered states

#### Scenario: Clicks on non-item lines resolve to nothing

- **WHEN** a left-click lands on a section header, a blank separator line, or an empty-state line in the index
- **THEN** no item is resolved and the cursor does not move
