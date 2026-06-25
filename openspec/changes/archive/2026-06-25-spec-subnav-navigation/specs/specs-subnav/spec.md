## RENAMED Requirements

- FROM: `### Requirement: Sub-navegación de specs`
- TO: `### Requirement: Specs sub-navigation chip row`

- FROM: `### Requirement: Ajuste de altura de contenido con sub-nav visible`
- TO: `### Requirement: Content height adjustment when the sub-nav is visible`

## REMOVED Requirements

### Requirement: Ciclo de specs con tecla 3
**Reason**: The `3` key's forward-only cycle was undiscoverable and one-directional, leaving users stuck on the first spec of a multi-spec change. It is replaced by bidirectional `[` / `]` navigation and left-click chip selection (see the new "Spec navigation with `[`, `]`, and mouse click" requirement). The `3` key now behaves as a plain primary-tab selector, consistent with `1`, `2`, and `4`.
**Migration**: To switch specs, press `[` (previous) or `]` (next), or left-click a spec chip. Pressing `3` selects the `specs` tab and no longer advances the spec.

## MODIFIED Requirements

### Requirement: Specs sub-navigation chip row
When the active tab is `specs` and at least one spec is available, the TUI SHALL display a chip row as the first line inside the content block, immediately after the horizontal separator (`├───┤`), showing the name of each spec. The chip of the currently visible spec SHALL be rendered with the active style (the same style as an active tab). The other chips SHALL be rendered with the inactive style. The row SHALL be static (it is not part of the scrollable viewport area). This applies identically in active changes (`ModeNormal`) and archived changes (`ModeViewingArchive`).

#### Scenario: Single spec
- **WHEN** the change has a single spec and the active tab is `specs`
- **THEN** the chip row is shown as the first line of the content block, with one chip representing that spec marked as active

#### Scenario: Multiple specs
- **WHEN** the change has two or more specs and the active tab is `specs`
- **THEN** the chip row is shown as the first line of the content block; the visible spec's chip is active and the others are inactive

#### Scenario: Row absent on other tabs
- **WHEN** the active tab is not `specs`
- **THEN** no specs chip row is shown in the content block

#### Scenario: Row does not disappear when scrolling
- **WHEN** the active tab is `specs` and the user scrolls down
- **THEN** the chip row remains visible as the first line of the content block

#### Scenario: Visual separation between the tab bar and the spec chips
- **WHEN** the active tab is `specs`
- **THEN** the horizontal separator `├───┤` appears between the tab bar and the specs chip row

### Requirement: Content height adjustment when the sub-nav is visible
When the specs sub-nav is visible, the content area SHALL reduce its height by one line to accommodate the extra row, preventing the viewport from overflowing outside the box.

#### Scenario: Reduced height on the specs tab
- **WHEN** the active tab is `specs` and specs are available
- **THEN** the viewport has one fewer line of height than on the other tabs

#### Scenario: Normal height on other tabs
- **WHEN** the active tab is `proposal`, `design`, or `tasks`
- **THEN** the viewport has the standard height

## ADDED Requirements

### Requirement: Spec navigation with `[`, `]`, and mouse click
When the active tab is `specs` and the change has more than one spec, the user SHALL be able to move between specs with the secondary sub-navigation keys `[` (previous spec) and `]` (next spec), wrapping around at the ends, and SHALL be able to select a specific spec by left-clicking its chip. Selecting a spec SHALL update the visible spec content. This navigation SHALL behave identically in active changes (`ModeNormal`) and archived changes (`ModeViewingArchive`). When the change has a single spec, `[` and `]` SHALL have no effect. The `←` / `→` arrows SHALL NOT change the active spec (they remain primary tab navigation). The selected spec SHALL persist while the same change is viewed: switching to another primary tab and back to `specs` SHALL preserve the selected spec. The selection SHALL reset to the first spec only when the active change changes.

#### Scenario: Next spec with `]`
- **WHEN** the active tab is `specs`, there are three specs, the second is active, and the user presses `]`
- **THEN** the third spec becomes active and the viewport shows its content

#### Scenario: Previous spec with `[`
- **WHEN** the active tab is `specs`, the second spec is active, and the user presses `[`
- **THEN** the first spec becomes active and the viewport shows its content

#### Scenario: `]` wraps from the last spec to the first
- **WHEN** the active tab is `specs`, the last spec is active, and the user presses `]`
- **THEN** the first spec becomes active

#### Scenario: `[` wraps from the first spec to the last
- **WHEN** the active tab is `specs`, the first spec is active, and the user presses `[`
- **THEN** the last spec becomes active

#### Scenario: Single spec, `[` / `]` do nothing
- **WHEN** the active tab is `specs`, there is only one spec, and the user presses `[` or `]`
- **THEN** the active spec remains the same and the content does not change

#### Scenario: Left-click selects a spec chip
- **WHEN** the active tab is `specs` and the user left-clicks the chip of a non-active spec
- **THEN** that spec becomes active and the viewport shows its content

#### Scenario: Navigation works in archived changes
- **WHEN** an archived change with multiple specs is open on the `specs` tab and the user presses `]` or clicks a chip
- **THEN** the active spec changes exactly as it does for an active change

#### Scenario: Arrows do not change the spec
- **WHEN** the active tab is `specs` with multiple specs and the user presses `→`
- **THEN** the active spec does not change (the arrow performs primary tab navigation only)

#### Scenario: Returning to the specs tab preserves the selected spec
- **WHEN** a non-first spec is active, the user switches to another tab and back to `specs` within the same change
- **THEN** the previously selected spec is still active

#### Scenario: Switching change resets the selected spec
- **WHEN** a non-first spec is active and the user moves to a different change with `h` or `l`
- **THEN** the selected spec resets to the first spec of the new change
