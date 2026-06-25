## 1. Code & build rename

- [x] 1.1 Module path `github.com/danshort/speclio` → `github.com/danshort/lectern` (go.mod + all imports); `cmd/speclio/` → `cmd/lectern/`; `docs/speclio.gif` → `docs/lectern.gif`
- [x] 1.2 GoReleaser (project/binary/homepage/`lectern.rb`), `Makefile`, `.gitignore`, release-please config/manifest name mentions
- [x] 1.3 `go build`, `go test`, `gofmt`, `goreleaser check` all pass; binary is named `lectern`

## 2. Docs & specs

- [x] 2.1 READMEs (EN/ES), `DEVELOPING.md`, `RELEASING.md`, `LICENSE` fork line
- [x] 2.2 Active OpenSpec spec name mentions; `build-tooling` binary-name requirement carried as a MODIFIED delta

## 3. Repo & release

- [x] 3.1 Rename GitHub repo `danshort/speclio` → `danshort/lectern`
- [x] 3.2 Post-merge: cut the `lectern` release (release-please) and remove the orphaned `speclio.rb` from the Homebrew tap
- [x] 3.3 Post-release: `brew uninstall speclio && brew install danshort/tap/lectern`

## 4. Verification

- [x] 4.1 `openspec validate rename-to-lectern --strict` passes
- [x] 4.2 Post-merge: PR passes the `lint`/`test` ruleset checks and squash-merges
