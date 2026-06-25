## Why

The name `speclio` (a coined spec+folio portmanteau) keyed off the *artifact* — the documents — which is the same thing the upstream name `dossier` already describes. The tool is the *reader/interface* for those artifacts, so the name should evoke reading, not the documents. `lectern` — the stand you read a document from — names the reader, fits the project's Roman/civic tooling family, is a real one-word name with no "what does that mean" tax, and is clear of Homebrew and namespace collisions. Renaming now is cheap: the tool has only been installed locally, so there is no migration burden.

## What Changes

- Rename the project and binary from `speclio` to `lectern`:
  - Module path `github.com/danshort/speclio` → `github.com/danshort/lectern`, all imports, `cmd/speclio/` → `cmd/lectern/`, the demo gif.
  - GoReleaser (project/binary/homepage/Homebrew formula `lectern.rb`), `Makefile`, `.gitignore`.
  - READMEs (EN/ES), `DEVELOPING.md`, `RELEASING.md`, `LICENSE` fork line, and active OpenSpec specs that name the binary.
- The GitHub repo is renamed `danshort/speclio` → `danshort/lectern` (old URLs redirect).
- Post-release cleanup: the orphaned `speclio.rb` formula is removed from the Homebrew tap once `lectern` releases.

## Non-goals

- No behavior changes — this is purely a rename.
- No change to the release automation, branch ruleset, or tap mechanism (only names within them).
- Historical records (`CHANGELOG`, `openspec/changes/archive/`) keep the old names as an accurate record of when the tool was `dossier`/`speclio`.

## Capabilities

### Modified Capabilities

- `build-tooling`: the produced binary and entry-point directory are renamed `speclio` → `lectern`.

## Impact

- `go.mod` + all imports, `cmd/lectern/`, `docs/lectern.gif`
- `.goreleaser.yaml`, `Makefile`, `.gitignore`, `release-please-config.json`/manifest (names only)
- `README.md`, `README.es.md`, `DEVELOPING.md`, `RELEASING.md`, `LICENSE`
- Active `openspec/specs/*` name mentions
- Breaking for the single existing local install: `brew uninstall speclio && brew install danshort/tap/lectern` (hence `feat!:`)
