# spec-validation Specification

## Purpose
TBD - created by archiving change validate-specs-and-changes. Update Purpose after archive.
## Requirements
### Requirement: Validate spec structure
The system SHALL validate each project spec against OpenSpec structural rules: the spec content MUST contain a `## Purpose` section and a `## Requirements` section, and every `### Requirement:` MUST contain at least one `#### Scenario:`. A spec violating any rule SHALL be considered invalid, with one error message per violation.

#### Scenario: Well-formed spec is valid
- **WHEN** a spec contains `## Purpose`, `## Requirements`, and each requirement has a scenario
- **THEN** validation returns no errors

#### Scenario: Spec missing the Requirements section
- **WHEN** a spec has `## Purpose` but no `## Requirements` section
- **THEN** validation returns an error indicating the missing `## Requirements` section

#### Scenario: Requirement without a scenario
- **WHEN** a spec contains a `### Requirement:` with no `#### Scenario:` before the next requirement
- **THEN** validation returns an error naming that requirement as missing a scenario

### Requirement: Validate change structure
The system SHALL validate each change: it MUST have a proposal, and every delta spec it contains MUST have at least one delta header (`## ADDED Requirements`, `## MODIFIED Requirements`, `## REMOVED Requirements`, or `## RENAMED Requirements`), with every requirement in an `ADDED` or `MODIFIED` section containing at least one `#### Scenario:`.

#### Scenario: Change without a proposal
- **WHEN** a change has no `proposal.md`
- **THEN** validation returns an error indicating the missing proposal

#### Scenario: Delta spec without a delta header
- **WHEN** a change's delta spec contains requirements but no `## ADDED/MODIFIED/REMOVED/RENAMED Requirements` header
- **THEN** validation returns an error for that delta spec

### Requirement: Surface validation errors in the index
The index SHALL display a visually distinct error marker next to any active change or project spec that fails validation. The marker SHALL NOT alter the line layout used for click hit-testing. Archived changes are frozen history and SHALL NOT be marked.

#### Scenario: Invalid spec is marked in the index
- **WHEN** the index renders a project spec that fails validation
- **THEN** an error marker is shown on that spec's row

#### Scenario: Valid item has no marker
- **WHEN** the index renders an active change or project spec that passes validation
- **THEN** no error marker is shown on that row

#### Scenario: Archived changes are not marked
- **WHEN** the index renders an archived change
- **THEN** no validation marker is shown, regardless of the archived change's content

### Requirement: Validation status is revalidated on view
The system SHALL recompute validation from the current in-memory artifact content each time the index is rendered, so the displayed status always reflects the latest polled content and is never stale.

#### Scenario: Fixing a spec clears its marker without restart
- **WHEN** an invalid spec is edited on disk to become valid and the index re-renders after the next poll
- **THEN** the error marker for that spec is no longer shown

### Requirement: Show validation errors when viewing an invalid spec
When viewing a spec that fails validation, the spec view SHALL display the specific validation error messages above the rendered spec content.

#### Scenario: Errors listed atop an invalid spec
- **WHEN** the user opens a spec that fails validation
- **THEN** the view shows a "Validation errors" block listing each error before the spec body

