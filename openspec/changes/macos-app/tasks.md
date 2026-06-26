## 1. Shared fixture corpus + Go golden tests (Phase 1 — stands alone)

- [x] 1.1 Create `testdata/corpus/` fixtures: `basic-project` (≥3 changes with mixed/equal `Created`, ≥3 spec dirs — exercises sort stability and unsorted `loadSpecs` order), `crlf-tasks` + `lf-tasks` (both with trailing newlines), `unreadable-artifact`, `malformed-archive-name` (`2026-13-99`, `2026-02-29`, `2024-02-29` for calendar validity), `malformed-meta` (bad `.openspec.yaml` → empty `Created`), `config-variants` (absent `rules`, `rules: {}`, multiline context), `delta-specs` (missing proposal, `HasPrefix` headers, empty-named requirement), `worktree-porcelain/*.txt` (captured porcelain text)
- [x] 1.2 Write the **serialization contract** doc (no `omitempty`; absent → `null`; empty slice → `[]`; empty map → `{}`; sorted keys; snake_case) and add JSON struct tags to the Go domain types accordingly
- [x] 1.3 `internal/openspec/golden_test.go`: golden tests for **every entry point** — `Project` (`*.json`), `ParseTasks` (`tasks.json`, incl. `LineNum`), `ExtractRequirement` (`requirements.json`), `parseWorktreeList` (`worktrees.json`), `ConfigToMarkdown` (`config.md`), validation (`validation.json`); `-update` flag to regenerate
- [x] 1.4 Byte-exact toggle write goldens for **both** LF and CRLF fixtures (`*.after-toggle.tasks.md`)
- [x] 1.5 Normalize the unreadable-artifact golden (presence + read-error flag + prefix-only content) so it is language-stable
- [x] 1.6 Generate goldens, confirm `go test ./internal/openspec/...` passes; keep existing `t.TempDir()` tests; wire into CI

## 2. OpenSpecKit — Swift domain port (Phase 2)

- [x] 2.1 Scaffold `macos/OpenSpecKit/` SwiftPM package (library + test target), no app dependency; pin Swift toolchain
- [x] 2.2 Port models as `Codable` structs with `CodingKeys` matching the Go JSON tags; honor the serialization contract (nil/empty/`[]`/`{}`)
- [x] 2.3 `protocol FileSystem` + `OSFileSystem`; `readDir` **sorts by name** (Swift `FileManager` is unsorted); not-found vs other-error distinction
- [x] 2.4 Parse YAML with **Yams** (not `Codable`); field-tolerant `.openspec.yaml` (swallow errors → empty `Created`) vs error-propagating `config.yaml`; preserve nil-vs-empty `rules`
- [x] 2.5 Port the loader using `components(separatedBy: "\n")` (Go `strings.Split` trailing/empty semantics); `loadFrom`, `loadFromPath` (grandparent `Project.Name`), archive listing (calendar-valid date gate), `loadSpecs` (separator, absent-on-empty), the two different spec-loading error semantics
- [x] 2.6 Port tasks: `parseTasks`; **separate** CRLF-safe `toggleTask` write path; `toggleTask` mutates the task list in place (`inout`) and re-reads before writing
- [x] 2.7 Port validation (incl. `proposal.Present`, `HasPrefix` headers vs anchored `deltaHeaderRe`, empty-named-requirement skip) and `extractRequirement`
- [x] 2.8 Port worktree porcelain parser + `normalizePath` as EvalSymlinks-then-lexical-Clean-fallback (not `resolvingSymlinksInPath`); define a `GitService` protocol (per 3.1) with the porcelain **parser separated from** the `Process` invocation, so the sandbox flip later swaps only the invocation
- [x] 2.9 `OpenSpecKitTests`: run every entry point against the shared `testdata/corpus/`, assert byte-equality vs the same goldens; unit-assert ToggleTask's in-memory mutation (no golden)
- [x] 2.10 Add a **required, non-path-filtered** macOS CI lane (`swift test`); both Go and Swift golden lanes green on every PR

## 3. Architecture decisions (gate Phases 4–6 — resolve before SwiftUI)

- [x] 3.1 **App Sandbox posture — DECIDED: Option C.** Ship Developer-ID **non-sandboxed** now (full features), architected for a later sandbox/App-Store flip: route all FS access through `FileSystem` and all git through a new `GitService` protocol (2.3/2.8), use security-scoped-bookmark access patterns from day one (4.1), and defer the git + sibling-worktree sandbox problem (scope-to-repo-root vs drop vs XPC helper) to the App Store decision (~1 yr out). Don't hard-wire Sparkle as the only updater (App Store bans it) — see 6.4.
- [x] 3.2 **Markdown renderer — DECIDED: swift-markdown + custom SwiftUI views** (Apache-2.0; renders tables/code-fences/nested-lists that `AttributedString` can't).

## 4. SwiftUI reader shell (Phase 4 — read-only)

- [x] 4.1 App target `macos/LecternApp/` depending on `OpenSpecKit`; project picker that persists a **security-scoped bookmark from day one** (per 3.1) and accesses all files through it, so the model is unchanged if the app is sandboxed later
- [x] 4.2 `NavigationSplitView`: sidebar of changes → artifacts; detail pane
- [x] 4.3 Markdown rendering (tables, code fences, nested lists), the `⚠ couldn't read` placeholder, and the inline validation banner (omitted for unreadable artifacts)
- [x] 4.4 **Top-level mode switcher** (toolbar segmented control): Active Changes, Archived Changes, Specs, Worktrees — the selected mode drives the sidebar list and detail pane (default: Active Changes)
- [x] 4.5 Port `loadProjectSpecsFrom` into `OpenSpecKit` (deferred in Phase 2) **+ a corpus fixture with `openspec/specs/` and a Go + Swift golden**, so the Specs mode stays in the cross-language contract
- [x] 4.6 **Specs mode**: list `openspec/specs/` capability specs; render the selected `spec.md` with markdown + the validation banner
- [x] 4.7 **Archived Changes mode**: list `openspec/changes/archive/` (newest first) via `listArchiveChangesFrom`, reusing change rendering (rich archive-specific display deferred)
- [x] 4.8 **Worktrees mode**: list git worktrees via `GitService`/`ProcessGitService`; graceful "unavailable" when git is absent
- [x] 4.9 Requirement focus/extract + jump-to navigation (via `extractRequirement`) within a rendered spec
- [x] 4.10 Project config view (`openspec/config.yaml` → `configToMarkdown`)

## 5. Interaction + OS integration (Phase 5)

- [x] 5.1 Task checkbox toggle through the CRLF-safe `toggleTask` with a re-read before write; integration test on LF and CRLF fixtures and on an externally-modified file
- [x] 5.2 Harden the worktrees `Process` invocation (5 s watchdog, error states) behind `GitService` — the Worktrees *view* itself lands in Phase 4 (4.8)
- [ ] 5.3 FSEvents live reload of `openspec/` (debounced); integration test that an external edit refreshes the view
- [x] 5.4 Open-in-editor / reveal-in-Finder (subject to the sandbox decision)

## 6. Packaging, signing, distribution (Phase 6)

- [ ] 6.1 **Ad-hoc** code-sign (`codesign --sign -`, required for Apple Silicon to run); produce a `.dmg`/zip. Developer-ID signing **deferred** (no Apple Developer account yet — see 3.1/distribution decision)
- [ ] 6.2 **Deferred:** notarize + staple (requires the Apple Developer account). When added, run it in a **decoupled** macOS job that cannot fail the existing goreleaser/CLI release
- [ ] 6.3 Publish a Homebrew **cask** alongside the CLI formula; document the first-launch Gatekeeper step (right-click → Open, or `--no-quarantine`) since the build is unnotarized
- [x] 6.4 **Update mechanism — DECIDED: `brew upgrade` only** (no Sparkle now; nothing foreclosed — App Store updates later if that path is taken)
- [ ] 6.5 Accessibility pass (VoiceOver, keyboard nav, Dynamic Type, contrast); update `README.md` with screenshots

## 7. Verification

- [ ] 7.1 Both golden lanes green in CI (Go + Swift, non-path-filtered) on every PR
- [ ] 7.2 Manual QA matrix: browse, render (tables/code/lists), validation banner, requirement focus, toggle (LF + CRLF + externally-modified), worktrees, live reload, missing-git, unreadable-artifact
- [ ] 7.3 Confirm the TUI is unchanged (no diffs under `internal/ui`, `cmd/`, `internal/openspec` logic — only added tests/tags)
- [ ] 7.4 Verify a signed+notarized build installs cleanly past Gatekeeper on a clean machine
