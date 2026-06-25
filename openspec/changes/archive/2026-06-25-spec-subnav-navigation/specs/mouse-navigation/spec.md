## ADDED Requirements

### Requirement: Spec chip selection via left-click
When the active tab is `specs` and the specs chip row is visible, the TUI SHALL switch to the clicked spec when the user performs a left-click (press) on a spec chip. The coordinate mapping SHALL use the same scheme as the tab bar: each chip's rendered width including `Padding(0, 1)` plus one space between chips, starting from X=1 (past the `│` border) on the sub-nav row. A click that does not land on any chip (including the space between chips, or outside the chip range) SHALL be ignored. This behaves identically in `ModeNormal` and `ModeViewingArchive`.

#### Scenario: Click on a spec chip switches to it
- **WHEN** the active tab is `specs`, the change has multiple specs, and the user left-clicks a non-active spec chip
- **THEN** that spec becomes active and the content area shows the rendered spec

#### Scenario: Click on the active spec chip does nothing disruptive
- **WHEN** the user left-clicks the chip of the already-active spec
- **THEN** the active spec does not change

#### Scenario: Click between spec chips does nothing
- **WHEN** the user left-clicks the space between two spec chips
- **THEN** the active spec does not change

#### Scenario: Click outside the chip range does nothing
- **WHEN** the user left-clicks on the sub-nav row outside the X range of any spec chip
- **THEN** the active spec does not change

#### Scenario: Chip click works in archived changes
- **WHEN** an archived change with multiple specs is open on the `specs` tab and the user left-clicks a spec chip
- **THEN** that spec becomes active exactly as it does for an active change
