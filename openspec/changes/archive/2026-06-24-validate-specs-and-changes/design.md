## Context

speclio loads every artifact's content into memory (`ProjectSpec.Content`, `Change.Proposal/Design/Tasks/SpecFiles`) and keeps it fresh via 500 ms polling (`handleTick` → `pollIndexMode`/`pollNormalModeContent`). The index (`ModeIndex`) lists active changes, project specs, and archived changes. There is currently no notion of artifact validity anywhere in the model or render path.

The `openspec` CLI already defines the canonical structural rules; we observed its exact messages (e.g. "Spec must have a Requirements section", "Each requirement MUST include at least one #### Scenario: block"). speclio is distributed standalone (Homebrew, no `openspec` dependency), so we reimplement the high-value subset in Go rather than shelling out.

## Goals / Non-Goals

**Goals:**
- Pure, table-testable validation functions over artifact content (no I/O).
- Visible error markers in the index for invalid specs and changes.
- Specific error messages shown when viewing an invalid spec.
- Always-fresh status without a bespoke cache.

**Non-Goals:**
- Complete CLI rule parity, a validation panel, auto-fix, or configurable rules.
- A separate mtime cache: see decision below.

## Decisions

- **Pure functions in `internal/openspec/validate.go`.**
  - `ValidateSpec(content string) []string` — returns error messages (empty/nil = valid). Checks: contains `## Purpose`; contains `## Requirements`; every `### Requirement:` block (lines until the next `### ` or `## `) contains at least one `#### Scenario:`.
  - `ValidateChange(ch Change) []string` — proposal present; every delta spec (`ch.SpecFiles`) contains at least one delta header matching `^## (ADDED|MODIFIED|REMOVED|RENAMED) Requirements`; every `### Requirement:` inside an `ADDED`/`MODIFIED` section contains at least one `#### Scenario:` (`REMOVED`/`RENAMED` sections are exempt).
  - These mirror the CLI's observed rules and are independently unit-tested.
- **Compute at render, not via a cache.** Validation is cheap string scanning over already-resident, already-polled content. `renderIndexContent` calls the validators per item as it renders. Because the index re-renders on poll, navigation, and entry, status is inherently current — satisfying "revalidate every time the page is viewed" with no separate cache or invalidation logic. The issue's mtime-cache suggestion is unnecessary at this scale (tens of small files); recorded as a future option if profiling ever shows cost.
- **Marker, not layout change.** Invalid items get a trailing red marker rendered with the existing `errStyle` (e.g. ` ✗`). A trailing marker preserves column alignment and, crucially, adds no lines — so the index's click-hit-testing line math (`indexItemAtContentLine`) is unaffected.
- **Spec-view detail.** When `ModeViewingSpec` shows an invalid spec, prepend a short, styled "Validation errors:" block (the messages from `ValidateSpec`) above the rendered markdown. This is the "detail view" for the selected item without introducing a new mode.
- **Markers only on actionable artifacts (project specs + active changes); archived changes are not marked.** An audit of this repo found project specs and the active change clean, but 7 of 58 archived changes flagged — mostly historical noise: old "no-spec-changes" placeholder spec files (no delta header) and prose-only `MODIFIED` deltas from looser early conventions. Archived changes are immutable history; a user cannot act on a marker there, so marking them is noise without recourse. Validation therefore runs only on project specs and active changes, where errors are actionable. (The `MODIFIED`-requires-a-scenario rule is kept for active changes because an archived delta replaces the whole requirement, so a MODIFIED requirement should restate its scenarios.)

## Risks / Trade-offs

- **[Low] Rule drift from the `openspec` CLI.** We implement a subset; the CLI may add rules. Mitigation: validators live in one small file with tests, easy to extend; scope is documented as a subset.
- **[Low] Recompute on every render.** Negligible for realistic project sizes; if it ever matters, cache keyed by content hash/mtime is a drop-in follow-up.
- **[Low] False sense of completeness.** A "valid" badge means "passes these structural checks," not "semantically correct." The marker only flags errors; absence of a marker is intentionally quiet.
