## ADDED Requirements

### Requirement: Unreadable spec marked in the index

When a project spec's `spec.md` exists but could not be read, the index SHALL show a `⚠` marker next to that spec's row, in place of (not in addition to) the `✗` validation marker, so an unreadable spec is distinguishable from an empty or structurally-invalid one.

#### Scenario: Unreadable spec shows a warning marker, not a validation cross

- **WHEN** a project spec's `spec.md` is unreadable and the index is rendered
- **THEN** that spec's row shows a `⚠` marker and not a `✗`

#### Scenario: Readable spec shows no warning marker

- **WHEN** a project spec loads normally
- **THEN** no `⚠` marker appears on its row
