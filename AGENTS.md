# AGENTS.md — working on lectern

Conventions and architecture for AI agents (and humans) working in this repo.
Read this first; it captures the project-specific workflow that isn't obvious
from the code alone.

## What lectern is

A reader/navigator for [OpenSpec](https://github.com/openspec) project artifacts
(proposals, designs, specs, tasks), shipped as **two front-ends over one shared
domain model**:

```
        ┌──────────────────────────┐      ┌──────────────────────────────┐
        │  Go TUI (Bubble Tea)     │      │  macOS app (SwiftUI/AppKit)   │
        │  cmd/lectern, internal/ui│      │  macos/LecternApp             │
        └────────────┬─────────────┘      └───────────────┬──────────────┘
                     │                                     │
        ┌────────────▼─────────────┐      ┌───────────────▼──────────────┐
        │  internal/openspec (Go)  │◀────▶│  OpenSpecKit (Swift)          │
        │  loader, tasks, validate │ same │  macos/OpenSpecKit            │
        └──────────────────────────┘ bytes└───────────────────────────────┘
                     └──────── shared golden corpus ───────┘
                            testdata/corpus/
```

- **Go TUI** — `cmd/lectern`, `internal/ui` (Elm-architecture Bubble Tea),
  `internal/openspec` (domain: load/parse/validate/toggle), `internal/config`.
- **macOS app** — `macos/LecternApp` (SwiftUI/AppKit reader + interactive task
  editing) on top of `macos/OpenSpecKit` (a Swift port of `internal/openspec`).
- **The contract that ties them together** is the **shared golden corpus** at
  `testdata/corpus/` (see `testdata/corpus/README.md`). Both languages must
  produce **byte-identical** output for the same fixtures, checked in CI. **A
  change to loader/domain behavior in one language is a failing build unless the
  other language and the goldens are updated to match.** This is the single most
  important invariant in the repo.

OpenSpec's directory layout (`openspec/ → changes/ · specs/ · changes/archive/`),
the `YYYY-MM-DD-<name>` archive naming, and the `### Requirement:` heading are
**fixed conventions of the OpenSpec tool** — not configurable. Don't try to make
them configurable; do surface malformed/missing artifacts with a ⚠ rather than
dropping them silently.

## Build, test, run

**Go TUI:**
```bash
make build      # -> ./lectern (dev binary in repo root; run with ./lectern)
make test       # go test -race -cover ./...
make lint       # golangci-lint run ./...   (needs golangci-lint)
make fmt        # goimports -w .            (needs goimports)
go run ./cmd/lectern            # build+run against the cwd's openspec/
```

**macOS app (needs Xcode):**
```bash
cd macos/OpenSpecKit && swift test     # domain + golden tests (21+ tests)
cd macos/LecternApp  && swift build    # compile the app
cd macos/LecternApp  && swift run      # run it (⌘O to choose a project)
macos/LecternApp/scripts/package.sh 0.1.0   # -> dist/Lectern.app + .zip
```

**The golden corpus (cross-language contract):**
```bash
# After ANY change to internal/openspec loader/domain behavior:
go test ./internal/openspec/ -run TestGolden -update   # regenerate goldens
# then review the diff, mirror the behavior in macos/OpenSpecKit, and run:
cd macos/OpenSpecKit && swift test                     # must match the new goldens
```
Goldens embed no absolute paths or OS-specific text (the test normalizes them);
keep placeholder content path-free so it stays portable across machines and
languages.

**CI** (`.github/workflows/ci.yml`, plus `pr-title-lint.yml`):
- `test` — `go vet` + `go test -race -cover`
- `swift` — `swift build`, `swift run oskgolden` (golden byte-check), `swift test`
- `lint` — **PR-title** conventional-commit lint (the title becomes the changelog)

## The development workflow (OpenSpec, spec-driven)

This repo dogfoods OpenSpec. **Do not jump straight to code.** Each unit of work
goes through the OpenSpec pipeline, driven by the `opsx:*` skills:

1. **explore** (`/opsx:explore`) — think, don't implement.
2. **propose** (`/opsx:propose`) — create the change: `proposal.md`, `design.md`,
   delta `specs/<cap>/spec.md` (ADDED/MODIFIED Requirements with
   `#### Scenario:` WHEN/THEN), `tasks.md`. Validate with
   `openspec validate <change> --strict`.
3. **apply** (`/opsx:apply`) — implement tasks, checking them off.
4. **verify** (`/opsx:verify`) — ALWAYS actually run the verify skill before
   archiving; don't just assert it inline. For UI-only behavior that isn't
   headlessly testable (AppKit/SwiftUI), get a **manual QA pass from the user**
   before running verify.
5. **archive** (`/opsx:archive`) — sync the delta spec into
   `openspec/specs/<cap>/spec.md`, then move the change dir to
   `openspec/changes/archive/YYYY-MM-DD-<name>/`.
6. **PR → CI → merge** (see below).

One issue at a time, kept clean and tight. Verify your risky assumptions against
the actual tool/docs before asserting them (e.g. "is X configurable?" — check,
don't guess).

## Git / PR / release conventions

- **`main` is ruleset-protected and squash-only.** Never commit directly; branch
  from up-to-date `origin/main`, open a PR.
- **The PR title becomes the changelog entry** (release-please + squash). Use
  Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`) — the `lint` CI check
  enforces this. For a richer entry, use a `BEGIN_COMMIT_OVERRIDE` block (see
  `RELEASING.md`).
- **Auto-merge is disabled.** Watch CI (`gh pr checks <n> --watch`), then
  `gh pr merge <n> --squash`.
- **Head branches auto-delete on merge.** Locally, `git branch -D <branch>`
  (squash-merged branches aren't recognized by `git branch -d`).
- Commit trailer: `Co-Authored-By: <model> <noreply@anthropic.com>`.
- Releases are automated alongside the CLI; `CHANGELOG.md` is **release-please
  managed — never hand-edit it**.

## Where the backlog and project state live

- **GitHub Issues + Project #5** (board fields: Priority P0–P2, Size XS–XL,
  Status; closed issues auto-move to Done). This is the durable source of "what's
  next" — it travels across machines/accounts; local agent memory does not.
- `openspec/specs/` is the living spec of current behavior; archived changes
  under `openspec/changes/archive/` are the development history.

## Current state / picking up where we left off (2026-06-30)

Recently shipped (all merged to `main`): macOS editor override + ⌘E (#110);
undo/redo for all macOS task mutations via byte-exact file snapshots (#103);
FSEvents-driven worktree progress (#98); the render-cache archive cleanup (#82);
and "warn-don't-skip a spec dir with no spec.md" across both loaders + golden
corpus (#96). No active (un-archived) OpenSpec changes remain.

Open backlog (see GitHub for current priority):
- **#102** macOS VoiceOver / accessibility — self-contained, a good next pick.
- **#100** expand macOS Preferences (partly chipped by #110/#103).
- **#111** TUI in-app settings screen.
- **#104** *Operate* the OpenSpec lifecycle (TUI + macOS), not just view it — the
  flagship, **large**; start with `/opsx:explore` before any code.
- **#77** macOS window tab groups across relaunch.
- **#67** Notarize the macOS app — **blocked on an Apple Developer account.**

There is a standing queue of **manual GUI QA** for recent macOS merges that only
the user can run (live worktree update, undo behaviors, editor override).
