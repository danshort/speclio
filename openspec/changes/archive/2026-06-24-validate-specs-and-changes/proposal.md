## Why

speclio reads OpenSpec artifacts but never tells the user when one is malformed. A spec missing a `## Requirements` section, or a requirement with no scenario, renders without complaint — the reader has no signal that the artifact is broken. (This repo shipped exactly such a spec, `openspec-root-path`, which only surfaced when the `openspec` CLI was run separately.) Validation status should be visible in the tool that people actually read these artifacts in.

## What Changes

- Add a lightweight, dependency-free validator for OpenSpec artifacts that mirrors the structural rules the `openspec` CLI enforces:
  - A **spec** SHALL have a `## Purpose` section and a `## Requirements` section, and every `### Requirement:` SHALL contain at least one `#### Scenario:`.
  - A **change** SHALL have a proposal, and each of its delta specs SHALL contain at least one delta header (`## ADDED/MODIFIED/REMOVED/RENAMED Requirements`), with every requirement in an `ADDED`/`MODIFIED` section containing at least one scenario.
- Validate project specs and active changes and surface the result in the index: items with validation errors get a clearly visible error marker. Archived changes are frozen history and are intentionally not marked (see design).
- When viewing an invalid spec, list its specific validation errors at the top of the spec view.
- Validation is recomputed from the in-memory artifact content on every index render, so it always reflects the current on-disk state (content is already kept fresh by the existing 500 ms polling). No stale status.

## Non-goals

- Full parity with every rule in the `openspec` CLI. This covers the high-value structural checks, not the complete ruleset.
- A dedicated validation mode/panel or auto-fixing of artifacts.
- Configurable or pluggable validation rules.
- Shelling out to the external `openspec` binary (speclio stays standalone).

## Capabilities

### New Capabilities

- `spec-validation`: validate specs and changes against OpenSpec structural rules and surface errors in the index and spec view, revalidating on view.

## Impact

- `internal/openspec/` — new `validate.go` with pure validation functions + tests
- `internal/ui/index.go` — render an error marker on invalid index items
- `internal/ui/spec.go` / `internal/ui/viewport.go` — show validation errors atop an invalid spec
- `internal/ui/styles.go` — a marker style (reuse `errStyle`)
- No new dependencies, no filesystem writes
