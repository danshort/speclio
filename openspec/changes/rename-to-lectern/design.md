## Context

The project was `dossier` (upstream), renamed to `speclio` when this fork diverged, and is now renamed to `lectern`. The earlier name described the documents; `lectern` describes the reader, which is what the tool is, and fits the existing Roman/civic naming of the team's other tools. Only the maintainer has installed it, so there is no third-party migration cost.

## Goals / Non-Goals

**Goals:**
- A complete, consistent rename across code, build, docs, and live specs.
- Keep the release automation, ruleset, and tap intact (rename names within them only).

**Non-Goals:**
- Any behavior change.
- Rewriting historical records (CHANGELOG, archived changes) — they correctly reflect the old names.

## Decisions

- **Mechanical sweep, history preserved.** `speclio` → `lectern` across all active text files; `openspec/changes/archive/` and `CHANGELOG*` are excluded so the development history stays accurate. The `cmd/` directory and demo gif are `git mv`d.
- **Binary name is a speced requirement.** `build-tooling`'s "Binary named …" requirement is carried as a `MODIFIED` delta so the canonical spec reflects `lectern` after archive. Incidental name mentions in other specs are plain text edits (not requirement changes).
- **Breaking-change marker.** Renaming the binary breaks the existing install, so the PR is `feat!:` — release-please cuts the next version accordingly. The old `speclio.rb` is deleted from the tap after `lectern` first releases.
- **GitHub repo rename** uses GitHub's redirect, so existing links/clones keep working.

## Risks / Trade-offs

- **[Low] Third rename.** Name churn has a cost, but it is contained: no external installs, redirects cover old URLs, and the mechanical pass is identical to the prior `dossier→speclio` rename.
- **[Low] Mixed naming until release.** Until the rename releases and the tap is cleaned up, the published `speclio` artifacts still exist; harmless since only the maintainer is affected and will re-install.
