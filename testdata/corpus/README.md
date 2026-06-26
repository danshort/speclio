# Shared fixture corpus

A committed corpus of OpenSpec project fixtures plus **golden output** that every
implementation of the domain layer — the Go `internal/openspec` package and the
planned Swift `OpenSpecKit` port — must reproduce exactly. It is the single
source of truth for *expected loader behavior*, run in CI on both toolchains so a
one-sided behavior change is a failing build, not a latent drift.

This is Phase 1 of the macOS-app change (`openspec/changes/macos-app/`). It is
useful on its own: it hardens the existing Go loader regardless of whether the
Swift app is ever built.

## Layout

| Fixture | Pins |
|---|---|
| `basic-project/` | change sort (desc `Created`, empty last, stable), `loadSpecs` ordering (≥3 spec dirs), artifact presence |
| `lf-tasks/`, `crlf-tasks/` | `ParseTasks` line numbers; byte-exact toggle write (LF + CRLF preserved) |
| `unreadable-artifact/` | present-but-unreadable artifact (read fault injected in the test; see below) |
| `malformed-archive-name/` | archive date gate — regex match **and** calendar validity (`2026-13-99`, `2026-02-29` rejected; `2024-02-29` accepted; `plain-name` → no date) |
| `malformed-meta/` | tolerant `.openspec.yaml` decode — a malformed file yields an empty `Created`, never an error |
| `config-variants/` | `LoadConfig` + `ConfigToMarkdown`: absent `rules`, `rules: {}`, multiline `context` |
| `delta-specs/` | `ValidateSpec` / `ValidateChange`: missing sections, missing proposal, `HasPrefix` header matching, requirement-without-scenario, empty-named requirement |
| `worktree-porcelain/*.txt` | `parseWorktreeList` over captured `git worktree list --porcelain` text (no live git) |

Goldens live under `golden/`. Regenerate with `go test ./internal/openspec/ -run TestGolden -update`.

## Serialization contract

Goldens are byte-compared, so both languages must emit **identical bytes**. The
contract:

- **Sorted keys.** Every JSON object's keys are sorted alphabetically.
  - Go: marshal, round-trip through `any`, re-encode (encoder sorts map keys).
  - Swift: `JSONEncoder.outputFormatting = [.sortedKeys, .prettyPrinted, .withoutEscapingSlashes]`.
- **Pretty-printed**, 2-space indent, single trailing newline.
- **No HTML escaping.** Go: `encoder.SetEscapeHTML(false)`. (`<`, `>`, `&`, `/` stay literal.)
- **Field names are snake_case**, defined by the Go `json:"..."` struct tags; Swift mirrors them via `CodingKeys`.
- **No `omitempty`.** Absent/`nil` → `null`; empty slice → `[]` (never `null`); empty map → `{}`.
- **`ItemKind`** serializes as the string `"section"` / `"task"`, not an int.

### Machine- and language-independence

Two values are *not* portable and are normalized before comparison:

1. **Absolute paths.** `Change.Path` and any path embedded in content are
   relativized to the fixture root (forward slashes). A final pass also replaces
   the absolute fixture root with `<root>` everywhere as a safety net.
2. **Read errors.** A Go `error` string is OS/locale-specific and differs from a
   Swift `NSError`. The error field is `json:"-"`; instead the golden records a
   derived `read_error` boolean and **prefix-only** placeholder content
   (`⚠ couldn't read <relative-path>`), dropping the error tail. The
   `unreadable-artifact` case is produced by injecting a synthetic read fault in
   the test (a real unreadable file cannot survive a git checkout), so the result
   is deterministic on every platform.

## What the corpus cannot pin

Golden coverage ≠ behavioral completeness. These differ by environment and need
targeted unit/integration tests, not goldens: cross-platform filesystem and
Unicode normalization (APFS NFD vs Linux), `normalizePath` symlink resolution
(macOS-only), YAML parsing edge cases (yaml.v3 vs Yams), and regex dialect (Go
RE2 vs Swift ICU). Any change to loader/tasks/validation/worktree-parsing
behavior **must** add or update a fixture here.
