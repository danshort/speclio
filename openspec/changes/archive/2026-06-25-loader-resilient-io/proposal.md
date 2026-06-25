## Why

The loader (`internal/openspec/loader.go`) has ~5 near-identical "ReadDir → handle not-found → filter dirs → sort" blocks, and the openspec path segments (`openspec`/`changes`/`specs`/`archive`, the artifact filenames) plus the `"### Requirement: "` schema literal are duplicated across `loader.go`, `validate.go`, and `model.go`. More importantly, read errors are swallowed: `loadFile`, `loadSpecs`, and `LoadProjectSpecsFrom` treat *any* error as "absent", so a file that exists but can't be read (permissions, `EISDIR`, transient I/O) silently vanishes — the worst failure mode for a tool whose whole job is to *show* you what's there. (Issue #36.)

## What Changes

- **Dedup:** one `listDirs(path, excludeArchive)` helper replaces the repeated ReadDir→filter loops in `ListChangeNamesFrom`, `ListArchiveNamesFrom`, `ListSpecNamesFrom`, `LoadProjectSpecsFrom`, and `ListArchiveChangesFrom` (callers keep their own sort).
- **Constants:** hoist the openspec dir names, artifact filenames, and the `"### Requirement: "` prefix to package-level constants in `internal/openspec`, consumed by `validate.go` and `model.go:currentSpecPath`.
- **Resilience:** replace `os.IsNotExist(err)` with `errors.Is(err, fs.ErrNotExist)` (unwraps) at all Stat/ReadDir/ReadFile sites. A genuine not-found stays benign (absent); any *other* read error on a primary file no longer vanishes — `loadFile`/`loadSpecs`/`LoadProjectSpecsFrom` return a *present* artifact with a `ReadErr` set and placeholder content (`"⚠ couldn't read <file>: <err>"`), so the failure is visible on open and machine-detectable.
- **Minimal model signal:** add `ReadErr error` to `Artifact` and `ProjectSpec` (the placeholder string alone can't be distinguished from real content and would pollute the validation layer). Not-found ⇒ absent (no `ReadErr`); unreadable ⇒ `Present: true` + `ReadErr`.
- **Validation stays honest:** `ValidateChange`/`validateDeltaSpec`/the index `✗` path skip artifacts/specs with `ReadErr` — an unreadable file is a read failure, not a structural/"missing" failure, so it must not produce a spurious `✗` or a false "missing proposal.md".
- **Index visibility:** an unreadable project spec **or** an unreadable artifact of an active change gets a `⚠` marker in the index (in place of the `✗`), so it reads as "unreadable", not "empty" or "valid".

## Capabilities

### New Capabilities
<!-- None: this hardens existing loader behavior. -->

### Modified Capabilities
- `openspec-loader`: a file that exists but cannot be read is surfaced (`Present` + `ReadErr` + placeholder), not conflated with "absent"; not-found is detected with `errors.Is(fs.ErrNotExist)`; one unreadable file does not sink the load.
- `index-specs-section`: a project spec whose `spec.md` is unreadable shows a `⚠` marker (in place of the `✗`).
- `change-index`: an active change with an unreadable artifact shows a `⚠` marker; an unreadable artifact does not produce a spurious `✗` or a false "missing" validation result.

## Non-goals

- No coarse "fail the whole load on one bad file" — the loader keeps degrading gracefully (one unreadable file must not sink the rest of the index).
- No change to how genuinely-absent files behave, to optional metadata (`.openspec.yaml`) swallowing, or to the directory-level error propagation that already exists.
- Not a rewrite of parsing or the `fileSystem` interface.
- The `ReadErr` field is the *minimal* signal needed; not a general richer error/diagnostics model.

## Impact

- Code: `internal/openspec/loader.go` (helper, constants, predicate, `ReadErr` + placeholder), `internal/openspec/validate.go` (skip `ReadErr` artifacts/specs; consume constants), `internal/ui/model.go` (`currentSpecPath` **and** `artifactPath` consume path constants), `internal/ui/index.go` (the `⚠` marker for spec rows and active-change rows).
- Tests: error-path coverage (unreadable ⇒ present+`ReadErr`+placeholder; not-found ⇒ absent; one bad file among many ⇒ others load) via a fake `fileSystem` returning a non-not-found error; index marker tests.
- No dependency, CLI, or on-disk format changes. Delivered on one branch / one PR.
