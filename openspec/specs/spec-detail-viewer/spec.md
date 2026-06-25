# spec-detail-viewer Specification

## Purpose
Allows viewing the content of a project spec (`openspec/specs/<name>/spec.md`) rendered as markdown directly from the index, in read-only mode with scrolling.

## Requirements

### Requirement: Spec viewing mode
The TUI SHALL implement a `ModeViewingSpec` mode that displays the content of `openspec/specs/<name>/spec.md` rendered as markdown in a read-only viewport. The mode SHALL be activated by pressing `Enter` on a spec in `ModeIndex`. `Esc` SHALL return to `ModeIndex`. There SHALL be no editing, tabs, or subnav in this mode.

#### Scenario: Open spec from the index
- **WHEN** the index cursor is on a spec and the user presses `Enter`
- **THEN** the TUI enters `ModeViewingSpec` and shows the content of the selected spec's `spec.md` rendered as markdown

#### Scenario: Scroll the content
- **WHEN** the mode is `ModeViewingSpec` and the user presses `j` or `k`
- **THEN** the viewport scrolls down or up respectively

#### Scenario: Return to the index
- **WHEN** the mode is `ModeViewingSpec` and the user presses `Esc`
- **THEN** the TUI returns to `ModeIndex` with the cursor on the spec that was being viewed

#### Scenario: Header in spec viewing mode
- **WHEN** the mode is `ModeViewingSpec`
- **THEN** the header shows `<project>  ·  <spec-name>  [spec]`

#### Scenario: HelpBar in spec viewing mode
- **WHEN** the mode is `ModeViewingSpec`
- **THEN** the help bar shows `j/k: scroll  Esc: index  q: quit`

### Requirement: Content field in ProjectSpec
`openspec.ProjectSpec` SHALL include a `Content string` field with the raw content of `spec.md`. `LoadProjectSpecs()` SHALL read and store this content when loading specs.

#### Scenario: Content populated
- **WHEN** `LoadProjectSpecs()` processes a spec with `spec.md` present
- **THEN** the returned `ProjectSpec` has `Content` with the full text of the file

#### Scenario: Content empty if spec.md absent
- **WHEN** a spec directory does not contain `spec.md`
- **THEN** `Content` is an empty string and the spec still appears in the list

### Requirement: Open spec with scroll to a specific requirement
When `ModeViewingSpec` is opened from an `indexKindRequirement` item, the TUI SHALL activate focus mode and render only the block of that requirement in the viewport, instead of showing the full spec scrolled to that requirement.

#### Scenario: Open spec from a requirement item
- **WHEN** the index cursor is on a requirement item and the user presses `Enter`
- **THEN** the TUI enters `ModeViewingSpec` in focus mode and the viewport shows only the content of that requirement

#### Scenario: Open spec from the spec item (no requirement target)
- **WHEN** the index cursor is on a spec item (not a requirement) and the user presses `Enter`
- **THEN** the TUI enters `ModeViewingSpec` showing the full spec from the beginning (existing behavior)

#### Scenario: Esc from spec opened via requirement returns to the index
- **WHEN** `ModeViewingSpec` was opened from a requirement item and the user presses `Esc`
- **THEN** the TUI returns to `ModeIndex` with the cursor on the requirement item from which it was opened
