## 1. Fix archived-tasks arrow navigation (#7)

- [x] 1.1 Guard the Tasks-tab `j`/`down` and `k`/`up` branches in `updateViewer` with `m.mode == ModeNormal` so archive mode scrolls the viewport
- [x] 1.2 Guard the Tasks-tab `Space` (toggle) branch with `m.mode == ModeNormal`
- [x] 1.3 Add a regression test asserting the task cursor does not move on `j`/`k` in `ModeViewingArchive`

## 2. Improve dim-text contrast (#4)

- [x] 2.1 Introduce a `dimColor` constant (256-color `245`) in `styles.go`
- [x] 2.2 Apply `dimColor` to `helpStyle` and `taskDoneStyle`, and to `doneCodeStyle` in `tasks.go`
- [x] 2.3 Leave decorative chrome (borders, disabled tabs, empty progress segments) on ANSI `8`

## 3. Documentation

- [x] 3.1 Add `DEVELOPING.md` describing how to run a dev build alongside the Homebrew install
- [x] 3.2 Link `DEVELOPING.md` from `README.md` and `README.es.md`

## 4. Verification

- [x] 4.1 `go vet`, `gofmt`, and `go test -race ./...` pass
