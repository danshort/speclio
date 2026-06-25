## 1. Constants (b)

- [x] 1.1 Add package-level constants in `internal/openspec` for the openspec dir names (`openspec`/`changes`/`specs`/`archive`), artifact filenames (`proposal.md`/`design.md`/`tasks.md`/`spec.md`/`config.yaml`/`.openspec.yaml`), and the `"### Requirement: "` prefix.
- [x] 1.2 Replace the scattered literals in `loader.go` and `validate.go` with the constants. Update `internal/ui/model.go` path-building — **both `currentSpecPath` and `artifactPath`** (the latter duplicates every artifact filename) — to consume the exported path constants.
- [x] 1.3 `gofmt`/`vet`/`go test ./...` clean; behavior unchanged.

## 2. Dedup directory listing (a)

- [x] 2.1 Add `func (l *Loader) listDirs(path string, excludeArchive bool) ([]string, error)`: ReadDir → `errors.Is(err, fs.ErrNotExist)` ⇒ `(nil, nil)` → keep `IsDir()` entries (excluding the archive dir when `excludeArchive`) → return names (no sort).
- [x] 2.2 Rewrite `ListChangeNamesFrom`, `ListArchiveNamesFrom`, `ListSpecNamesFrom`, and the dir-collection loops in `LoadProjectSpecsFrom` / `ListArchiveChangesFrom` to call `listDirs`, then apply each caller's existing sort.
- [x] 2.3 `gofmt`/`vet`/`go test ./...` clean; behavior unchanged (existing loader tests pass).

## 3. Resilient reads (c)

- [x] 3.1 Replace every `os.IsNotExist(err)` with `errors.Is(err, fs.ErrNotExist)` — loader.go Stat sites (78, 247, 250), `LoadConfigFrom` (121), and all ReadDir/ReadFile sites. Add a shared `unreadablePrefix` constant. Add `ReadErr error` to `Artifact` and `ProjectSpec`.
- [x] 3.2 `loadFile`: not-found ⇒ `Artifact{}` (absent, unchanged); any other read error ⇒ `Artifact{Present: true, ReadErr: err, Content: unreadablePrefix + path + ": " + err}`.
- [x] 3.3 `loadSpecs` and `LoadProjectSpecsFrom`: not-found ⇒ skip/absent as today; other read error on a `spec.md` ⇒ include the spec/NamedSpec with `ReadErr` + placeholder content (do not drop it); the loop must keep going (graceful degradation). Stat sites get the predicate swap only (no placeholder) — a non-not-found Stat falls through to a Read that surfaces it.
- [x] 3.4 Validation honesty (`validate.go`): `ValidateChange` treats an artifact with `ReadErr` as present-but-not-validatable (no "missing", no parse); `validateDeltaSpec` skips `SpecFiles` with `ReadErr`. So an unreadable file never yields a `✗`.
- [x] 3.5 Index `⚠` marker (`index.go` `renderIndexContent`): for a **project spec** with `ReadErr` and for an **active change** with any `ReadErr` artifact, render a warn-styled `⚠` **in place of** the `✗` validation marker (short-circuit the `ValidateSpec`/`ValidateChange` call for that row).

## 4. Tests

- [x] 4.1 Add a test `fileSystem` (or extend the existing fake) that returns a permission/other error for a chosen path, distinct from not-found.
- [x] 4.2 Loader tests: unreadable `proposal.md` ⇒ artifact `Present` with `ReadErr` set + placeholder content; missing ⇒ absent (no `ReadErr`); one unreadable spec among several ⇒ others load, only it carries `ReadErr`; not-found dir ⇒ empty (no error).
- [x] 4.3 Validation tests: a change with an unreadable `proposal.md` is NOT reported "missing proposal.md" and emits no `✗`; an unreadable delta `spec.md` is skipped (no spurious delta-header error).
- [x] 4.4 Index tests: an unreadable project spec renders `⚠` and not `✗`; an active change with an unreadable artifact renders `⚠`; readable items render neither.
- [x] 4.5 `gofmt`/`vet`/full `go test ./...` clean.

## 5. Validate

- [x] 5.1 Run `openspec validate loader-resilient-io` and resolve any issues.
