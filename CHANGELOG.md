# Changelog

## [0.20.1](https://github.com/danshort/lectern/compare/v0.20.0...v0.20.1) (2026-06-27)


### Bug Fixes

* ship the macOS app icon in released builds ([#88](https://github.com/danshort/lectern/issues/88)) ([fa8514f](https://github.com/danshort/lectern/commit/fa8514fa28e49f1de8ea138cfcd89989a698c837))

## [0.20.0](https://github.com/danshort/lectern/compare/v0.19.0...v0.20.0) (2026-06-27)


### Features

* add native macOS app for browsing OpenSpec specs and changes ([1af4d08](https://github.com/danshort/lectern/commit/1af4d08d47fb427b63701f6afeef90446862f96b))


### Bug Fixes

* invalidate render cache on editor return ([82eff28](https://github.com/danshort/lectern/commit/82eff28ad16f0f4b503466e3b8beb5000f44cf89))
* invalidate render cache on editor return ([#82](https://github.com/danshort/lectern/issues/82)) ([8cb2439](https://github.com/danshort/lectern/commit/8cb24396f29a62d68332140c6a4b3ca140437e52))

## [0.19.0](https://github.com/danshort/lectern/compare/v0.18.0...v0.19.0) (2026-06-25)


### Features

* navigate sub-specs with [/] and clickable chips, drop 3-cycle ([#51](https://github.com/danshort/lectern/issues/51)) ([#54](https://github.com/danshort/lectern/issues/54)) ([98ad067](https://github.com/danshort/lectern/commit/98ad067f42acdf60326846aaf992528625b0d78f))


### Bug Fixes

* normalize CRLF when parsing specs and tasks ([#43](https://github.com/danshort/lectern/issues/43)) ([d541c24](https://github.com/danshort/lectern/commit/d541c2497bd68ee992bc3fc9a15823ac2c879c37))
* surface unreadable files instead of silently dropping them ([#36](https://github.com/danshort/lectern/issues/36)) ([#52](https://github.com/danshort/lectern/issues/52)) ([6479362](https://github.com/danshort/lectern/commit/6479362a5b1aadc747889996d47f52375665e0f8))

## [0.18.0](https://github.com/danshort/lectern/compare/v0.17.0...v0.18.0) (2026-06-25)


### Features

* add `?` keyboard shortcut help overlay ([#32](https://github.com/danshort/lectern/issues/32)) ([5bae0cb](https://github.com/danshort/lectern/commit/5bae0cbef20c94c2de0b95cd0d46d71e2891090b))
* add `e` (open in editor) shortcut to spec views ([#24](https://github.com/danshort/lectern/issues/24)) ([52399d1](https://github.com/danshort/lectern/commit/52399d1149d6a90ed619d0e8e2dc092603e7bd86))
* add expand (space) to archived changes in index ([#31](https://github.com/danshort/lectern/issues/31)) ([bd474e3](https://github.com/danshort/lectern/commit/bd474e3d65ce3d56124f69376e5472fb262e35bf))
* add worktrees view (w) to survey active changes and live task progress across all git worktrees of the repository, with read-only viewing of changes in sibling worktrees ([9e8cd05](https://github.com/danshort/lectern/commit/9e8cd05009eda079acb744e2ae747f0182cb93ca))


### Bug Fixes

* index-mode panic, worktree-edit corruption, and robustness gaps (+ English-only docs) ([#34](https://github.com/danshort/lectern/issues/34)) ([655900d](https://github.com/danshort/lectern/commit/655900d5b270c983204b44564413a494b0879e2c))
* render archived change dates in ISO 8601 format ([#27](https://github.com/danshort/lectern/issues/27)) ([072099c](https://github.com/danshort/lectern/commit/072099c72f9da30e0fca48700e456208c958c71a))

## [0.17.0](https://github.com/danshort/lectern/compare/v0.16.0...v0.17.0) (2026-06-25)


### ⚠ BREAKING CHANGES

* the binary is renamed `speclio` → `lectern`. Reinstall with `brew uninstall speclio && brew install danshort/tap/lectern`.

### Features

* rename the tool and binary from speclio to lectern ([d6e8dea](https://github.com/danshort/lectern/commit/d6e8dea4f91085d29d9265a530b62a8b5067cda0))


### Bug Fixes

* **ci:** keep release-please in 0.x for pre-1.0 breaking changes ([#21](https://github.com/danshort/lectern/issues/21)) ([1635f55](https://github.com/danshort/lectern/commit/1635f55ed9a2d92391266723533fe0b48c6d51b0))

## [0.16.0](https://github.com/danshort/speclio/compare/v0.15.0...v0.16.0) (2026-06-25)


### Features

* arrow keys switch tabs when viewing a change ([#6](https://github.com/danshort/speclio/issues/6)) ([f8639b7](https://github.com/danshort/speclio/commit/f8639b7662113acf16aeb07985c5b46acc77a7f6))
* validate specs and changes, surface errors in UI ([#5](https://github.com/danshort/speclio/issues/5)) ([3578e3b](https://github.com/danshort/speclio/commit/3578e3b59e2e6d48474c2c5626bf039c6b00cdf1))


### Bug Fixes

* archive tasks-tab arrows ([#7](https://github.com/danshort/speclio/issues/7)), dim-text contrast ([#4](https://github.com/danshort/speclio/issues/4)); add DEVELOPING.md ([621a65e](https://github.com/danshort/speclio/commit/621a65e4499af02b5503355c94e15ea56ab94569))
* **ci:** release-please should use v-prefixed tags (not speclio-v) ([#13](https://github.com/danshort/speclio/issues/13)) ([c02ab28](https://github.com/danshort/speclio/commit/c02ab286c6c3c8c5a3ddf66e897b5536e6b9552c))
* **openspec:** normalize openspec-root-path spec header ([b93bb63](https://github.com/danshort/speclio/commit/b93bb6337fb71f4391ca306592c3b199e2d65d45))

## v0.15.0

### Changed
- **Renamed the project and binary from `dossier` to `speclio`.** speclio is a fork of [dossier](https://github.com/fselich/dossier) by fselich, maintained independently as it diverges. Install with `brew tap danshort/tap && brew install speclio`.

### Added
- Prebuilt **macOS** binaries (amd64 + arm64) alongside Linux in every release, distributed via a Homebrew tap.
- `govulncheck` dependency scanning in CI; `RELEASING.md` documenting the release process.

### Fixed
- `$EDITOR` values with arguments (e.g. `code --wait`, `emacs -nw`) now launch correctly.

## v0.14.1

### Fixed
- Done-task code spans in the task list no longer show the first letter in a different color. Lipgloss renders underlined text character by character, resetting the foreground between them. The fix combines underline with the foreground color so each character inherits both.

## v0.14.0

### Added
- Press `/` in the index view to filter changes, specs, and archived items by name in real-time. Type to narrow down, `Enter` to lock the filter, `Esc` to clear it. A search box, basically.

## v0.13.0

### Internal
- Split the monolithic `handleKeyPress()` into per-mode update functions, each in its own file: `viewer.go`, `index.go`, `spec.go`, `config.go`. `update.go` is now a thin dispatcher.
- Introduced a `fileSystem` interface and `Loader` struct in `openspec`, so the package no longer depends on `os` directly. All public functions preserved via backward-compatible wrappers.
- Added `.golangci.yml` with errcheck, staticcheck, govet, unused, gofmt, goimports, and a `Makefile` with `test`, `lint`, and `fmt` targets.
- Eliminated all silent `log.Printf` error calls. Archive and spec load errors are now displayed in the help bar for 3 seconds via `m.errMsg`, exactly like toggle errors.

### Changed
- Tab bar `parts` slice is now preallocated to exactly 4 entries, and the tasks `items` slice preallocates to the line count. Everything is now 3 nanoseconds faster. Totally worth the token spend.
- The `taskCounts` function no longer uses naked returns (which were confusing to anyone who scrolled past line 491 of index.go).
- Layout constants (`chromeTop`, `chromeHeader`, etc.) replace magic numbers in `contentHeight()`. Now you know why it was subtracting 6.
- The reload-merge logic that was copy-pasted in two places is now a single `mergeReloadedChange()` method. DRY*2.

## v0.12.0

### Fixed
- Starting dossier with no pending changes now shows the index view with specs and archived changes instead of a blank screen.
- Task content updates inside existing changes now trigger a live refresh of the index list instead of silently ignoring them.
- The loading placeholder (`"Loading..."` / `"Cargando..."`) was removed. Raw markdown is shown immediately while the styled version renders in the background. Goodbye to the involuntary epilepsy mode.

### Changed
- Change list in the index view is now sorted by `created` date (descending). Before, they were sorted by whatever the filesystem felt like.

## v0.11.0

### Fixed
- Mouse stopped working after returning from the external editor (`e`). Turns out Bubble Tea v1 didn't save mouse state when suspending the terminal. It works now, but it doesn't matter because nobody should be using a mouse anyway.
- App would crash on startup if `archive/`, `specs/`, or `changes/` directories didn't exist. Now it returns empty lists as it should, without making a scene.
- The app background was black instead of the terminal's default color. `NoColor` means "no color," not "black." Who knew.
- `go.mod` had all dependencies marked as indirect. All of them. Including Bubble Tea, which is literally what the app is about.

### Changed
- Full migration to Bubble Tea v2, Bubbles v2, Lip Gloss v2, and Glamour v2. New imports, new declarative API for `View()`, key and mouse messages split into separate types. About 1300 lines touched. Don't ask for whom.
- `renderWithBackground()` and `bgSGRRestore()` removed. Bubble Tea v2 handles the background on its own. One less function to maintain.

### Added
- Unit tests. Yes, finally. ~30 tests across `loader_test.go`, `tasks_test.go`, and `view_test.go`. 74% coverage in `openspec`. UI tests are harder, don't judge me.
- CI via GitHub Actions: `go vet`, `go test -race`, and coverage on every push and PR to `main`. Failures are now caught before merging, not after.

### Internal
- The `openspec` package now accepts an explicit root path in all its functions (`LoadFrom`, `LoadConfigFrom`, etc.) instead of calling `os.Getwd()` internally. More testable, less coupled to global state.
- All loader functions now return `error` instead of silently swallowing failures. Malformed YAML errors are no longer swept under the rug.

## v0.10.0

### Added
- Tab bar now shows a distinct color (cyan) for progress bars that reach 100% completion. This change alone deserved a jump straight to v1.0, I know.
- New project info view: press `i` to see `openspec/config.yaml` rendered as markdown. Still can't edit it. I forgot to add that.
- Mouse support: click on tabs to switch between them, scroll wheel works on viewports. Still, don't use a mouse. It's for cowards.
- `Tab` / `Shift+Tab` cycle forward and backward through available tabs. Welcome to the world of keybinding incompatibilities between the app and the window system.
- `--version` / `-v` flag to print the current version. The AI did this on its own, without being asked.

### Changed
- Progress bar at 100% completion now renders in cyan instead of green. Cyan is like light blue, in case I forget.
- Goreleaser releases are now fully automated (no more drafts). Boring.
- Help bar updated to include `Tab` and mouse shortcuts.

### Internal
- Split `internal/ui/model.go` into six focused files (`model.go`, `update.go`, `viewport.go`, `index.go`, `tasks.go`, `view.go`). Super boring.
