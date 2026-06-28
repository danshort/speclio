# Warn instead of silently skipping a spec directory with no spec.md

## Why

A spec capability directory that exists but contains no `spec.md` is malformed —
the directory declares a capability, but there's nothing in it. The loader
handles this case inconsistently and silently (#96):

- `loadSpecs` (a change's delta specs): the directory is **silently dropped**
  (`continue`), so it vanishes from the TUI/app entirely.
- `LoadProjectSpecsFrom` (project specs): the directory is surfaced but with
  **empty content and no warning**, so it looks like an intentional empty spec.

Either way the user gets no signal that something is wrong. The loader already
has the right pattern for this — an *unreadable* artifact is surfaced as a
*present* artifact with a ⚠ placeholder rather than vanishing. A spec directory
missing its `spec.md` should be treated the same way.

This is the surviving kernel of #96 after validating that OpenSpec exposes no
configurable layout (the directory structure, archive naming, and
`### Requirement:` heading are fixed conventions of the tool). The only real
gap was silent degradation on a missing artifact.

## What Changes

- A spec directory (a change's `specs/<cap>/` or the project's
  `specs/<cap>/`) that exists but has no `spec.md` SHALL be surfaced as a
  *present* spec carrying a visible placeholder and a read-error marker (⚠) —
  not dropped, and not shown as silently empty.
- The fix lands in both the Go loader (`internal/openspec/loader.go`) and the
  Swift `OpenSpecKit` port, keeping the cross-language golden contract in sync;
  the corpus gains a delta-spec-directory-without-spec.md fixture and goldens
  are regenerated.

## Non-goals

- Making OpenSpec's directory layout, archive naming, or requirement-heading
  format configurable or "tolerant" — these are fixed by the OpenSpec tool and
  enforced by `openspec validate --strict`. There is no config to read.
- Changing the malformed-archive-name behavior: an archive dir that doesn't
  match `YYYY-MM-DD-<name>` already stays visible (only the human date is
  omitted), which is acceptable.
- Any change to how genuinely *absent* primary artifacts (a change with no
  `design.md`, etc.) are handled — absent stays absent.
