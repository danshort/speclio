## Context

`internal/openspec/loader.go` repeats a "ReadDir → handle not-found → filter `IsDir` → sort" block in five places, scatters the openspec path segments and artifact filenames as string literals (also duplicated in `validate.go` and `model.go:currentSpecPath`), and the `"### Requirement: "` prefix appears in `loader.go` and `validate.go`. Separately, three read sites swallow *all* errors as "absent":

- `loadFile` (loader.go:289) → `Artifact{}` on any error.
- `loadSpecs` (loader.go:308) → `continue` on any per-spec read error.
- `LoadProjectSpecsFrom` (loader.go:149) → skip parse on any read error.

So "absent" and "present but unreadable" are conflated — a permissions/`EISDIR`/I/O failure makes a file vanish with no signal. Directory-level errors are *already* propagated (and several `*returns error on failure* requirements exist); the gap is per-file reads.

## Goals / Non-Goals

**Goals:**
- One `listDirs` helper; path/schema constants in one place.
- Not-found detected with `errors.Is(fs.ErrNotExist)`; unreadable primary files surfaced, not dropped; graceful degradation preserved.
- Unreadable project specs visibly marked in the index.

**Non-Goals:**
- No `Artifact`/`ProjectSpec` error field (see decision 3 for how the marker works without one).
- No coarse "one bad file fails the whole load."
- No change to absent-file behavior, optional `.openspec.yaml` swallowing, or the `fileSystem` interface.

## Decisions

**1. `listDirs(path string, excludeArchive bool) ([]string, error)`** — the only axes of variation across the five call sites are "exclude the `archive` dir?" and sort order. The helper does ReadDir → `errors.Is(fs.ErrNotExist)` returns `(nil, nil)` → filter `IsDir()` (and `name != archiveDir` when `excludeArchive`) → return names; **callers sort** (changes by created-date elsewhere, archive reverse, specs ascending). Non-not-found ReadDir errors propagate (unchanged).

**2. Path/schema constants** in `internal/openspec` (e.g. `dirOpenspec = "openspec"`, `dirChanges`, `dirSpecs`, `dirArchive`, `fileProposal`/`fileDesign`/`fileTasks`/`fileSpec`/`fileConfig`/`fileMeta`, `reqPrefix = "### Requirement: "`). `validate.go` and `model.go:currentSpecPath` consume them. Pure mechanical, no behavior change.

**3. Add a minimal `ReadErr error` to `Artifact` and `ProjectSpec` (revised after plan verification).** The original idea was placeholder content only, no model field — but verification showed that conflates with the validation layer: making an unreadable file `Present:true` with placeholder content silences `ValidateChange`'s "missing proposal.md" and makes `ValidateSpec`/`validateDeltaSpec` emit a *spurious* `✗` on the placeholder text. A single field is far less churn than getting a string-sniff right across the loader (producer), the index, and validation (consumers), and it can't collide with real content. So:
- `loadFile`: not-found ⇒ `Artifact{}` (absent, unchanged). Other read error ⇒ `Artifact{Present: true, ReadErr: err, Content: unreadablePrefix + path + ": " + err}`. `Present:true` keeps "missing" from firing; `ReadErr` is the machine signal; `Content` is the human view on open.
- `loadSpecs` / `LoadProjectSpecsFrom`: not-found ⇒ skip/absent; other error ⇒ include the spec with `ReadErr` + placeholder content (do not drop it); loops keep going (graceful degradation).

**4. Validation skips `ReadErr` items; the index marks them `⚠`.**
- `validate.go`: `ValidateChange` treats an artifact with `ReadErr` as present-but-not-validatable (no "missing", no content parse); `validateDeltaSpec` skips `SpecFiles` with `ReadErr`. So an unreadable file never yields a `✗`.
- `index.go`: per spec row and per active-change row, if the spec/any artifact has `ReadErr`, render `⚠` (a new warn-styled marker mirroring `validationMarker`) **instead of** computing the `✗` validation marker for that item. Covers both `index-specs-section` (specs) and `change-index` (active changes).

**5. Mechanical-only sites.** The `Stat` checks (loader.go:78/247/250) and `LoadConfigFrom` (loader.go:121) get the `os.IsNotExist`→`errors.Is(fs.ErrNotExist)` swap but **no** placeholder: Stat failures fall through to a downstream Read that surfaces the error, and `LoadConfigFrom` already propagates non-not-found errors. Optional `.openspec.yaml` metadata (loader.go:275) stays swallowed by design.

## Risks / Trade-offs

- [Adding `ReadErr` touches the data model] → It's one field on two structs; verification showed it's *less* churn and risk than the stringly-typed alternative, and it keeps validation honest.
- [Placeholder content mistaken for real content] → It carries `ReadErr` for logic; the visible `⚠ couldn't read …` text is only for the viewport and reads clearly as an error.
- [`os.IsNotExist` → `errors.Is(fs.ErrNotExist)`] → Strictly more correct (unwraps); `OSFS` returns raw os errors so direct cases are unchanged, wrapped cases now handled.
- [Unreadable change artifact silences a real "missing" error?] → No: "missing" means absent (`!Present`); unreadable is `Present:true + ReadErr` and is surfaced via `⚠`, a distinct, more accurate signal.

## Open Questions

- None blocking. (Sentinel-vs-field resolved in favor of the field after verification.)
