## 1. Signature helper + cache

- [x] 1.1 Add a file-signature type `{mtime time.Time, size int64, present bool}` and a helper that derives it from `fileSystem.Stat` (→ `os.FileInfo`), treating a not-found stat as `present=false`
- [x] 1.2 Add a freshness cache (`map[string]fileSig`) on the `ui.Model`, plus a method like `changedSince(path) bool` that stats the path, compares to the cached signature, updates the cache, and reports whether it changed (unknown path → changed)
- [x] 1.3 Keep index and worktrees caches separate (each gates its own in-memory copy). No explicit eviction is needed: a removed change's stale entry is never queried, and a re-created change at the same path is detected via its new mtime/size.

## 2. Gate the poll paths

- [x] 2.1 `pollIndexMode` (`internal/ui/index.go`): in the no-structural-change branch, only `ReloadChange` an active change when its `tasks.md` signature changed; leave the three directory-list `ReadDir`s unconditional
- [x] 2.2 `pollWorktrees` (`internal/ui/worktrees.go`): only `ReloadChange` a sibling change when its `tasks.md` signature changed
- [x] 2.3 No explicit cache eviction on structural rebuild (see 1.3): the structural branch reloads via `LoadFrom` and the next tick re-gates correctly; stale entries for removed changes are harmless
- [x] 2.4 Leave `pollNormalModeContent` and `pollWorktreeChange` (single-change content reloads) and the editor-return path unchanged — none are gated

## 3. Tests

- [x] 3.1 With a counting fake `fileSystem`, assert an unchanged tick performs zero `ReadFile` re-reads (only stats) across the gated paths
- [x] 3.2 A changed `(mtime, size)` triggers exactly one reload and updates the cache; the next unchanged tick does not re-read
- [x] 3.3 A deleted `tasks.md` (present → absent) is detected and updates state
- [x] 3.4 First-seen path is treated as changed (loads once, then quiesces)
- [x] 3.5 Editor-return reload still re-reads regardless of cached signature

## 4. Verification

- [x] 4.1 `go test -race ./...` passes; `go vet` + `gofmt` clean
- [ ] 4.2 Manual: open the TUI on a project; confirm idle ticks do no file reads (e.g. observe via a debug count or `fs_usage`/`strace`), and that toggling/editing `tasks.md` externally still updates within ~500 ms
- [ ] 4.3 Manual: worktrees view still tracks sibling-change progress live after gating
