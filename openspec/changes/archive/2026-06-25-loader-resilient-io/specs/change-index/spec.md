## ADDED Requirements

### Requirement: Unreadable change artifact marked in the index

When an active change has an artifact (`proposal.md`, `design.md`, `tasks.md`, or a `spec.md`) that exists but could not be read, the index SHALL show a `⚠` marker on that change's row, and SHALL NOT emit a `✗` validation marker or a false "missing artifact" result for the unreadable file (an unreadable file is a read failure, not a structural one).

#### Scenario: Unreadable change artifact shows a warning marker

- **WHEN** an active change's `proposal.md` exists but is unreadable and the index is rendered
- **THEN** the change's row shows a `⚠` marker and no spurious `✗`

#### Scenario: Genuinely missing artifact still validates as missing

- **WHEN** an active change's `proposal.md` does not exist
- **THEN** validation still reports it missing (unchanged from today), distinct from the unreadable case
