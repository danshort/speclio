# openspec-loader spec delta

## MODIFIED Requirements

### Requirement: Unreadable files are surfaced, not silently dropped

The loader SHALL distinguish a file that does not exist from a file that exists but cannot be read. Not-found SHALL be detected with `errors.Is(err, fs.ErrNotExist)` (which unwraps). A genuine not-found SHALL be treated as absent (unchanged behavior). Any other read error on a primary artifact (`proposal.md`, `design.md`, `tasks.md`, or a spec's `spec.md`) SHALL produce a *present* artifact whose content is a visible placeholder describing the failure, rather than an absent or empty result. A single unreadable file SHALL NOT fail the rest of the load.

A spec capability directory (a change's `specs/<cap>/` or the project's `specs/<cap>/`) that exists but contains no `spec.md` SHALL likewise be surfaced as a *present* spec carrying a visible placeholder and a read-error marker, rather than being silently dropped or shown as silently empty. This applies to both the change delta-spec loader and the project-spec loader, so a malformed capability folder is visible (with a ⚠) instead of vanishing.

#### Scenario: Unreadable artifact surfaces a placeholder

- **WHEN** an artifact file exists but reading it returns a non-not-found error (e.g. permission denied)
- **THEN** the artifact is present and its content indicates it could not be read, and the rest of the change still loads normally

#### Scenario: Missing artifact stays absent

- **WHEN** an artifact file does not exist (`fs.ErrNotExist`)
- **THEN** the artifact is absent, exactly as before

#### Scenario: One unreadable spec does not sink the others

- **WHEN** one project spec's `spec.md` is unreadable among several specs
- **THEN** the readable specs load normally and only the unreadable one carries the placeholder

#### Scenario: A change's spec directory with no spec.md is surfaced, not dropped

- **WHEN** a change has a `specs/<cap>/` directory that contains no `spec.md`
- **THEN** that capability appears among the change's spec files as a present spec carrying a placeholder and a read-error marker, rather than being omitted

#### Scenario: A project spec directory with no spec.md carries a warning

- **WHEN** the project's `specs/<cap>/` directory contains no `spec.md`
- **THEN** that capability is listed with a placeholder and a read-error marker, rather than as a silently empty spec
