## Context

The 500 ms tick (`update.go:72`, re-armed each tick) drives mode-specific poll helpers in `internal/ui/index.go` and `internal/ui/worktrees.go`. Each unconditionally re-reads and re-parses:

- `pollIndexMode` → `ReloadChange` for every active change (reads proposal + design + tasks + specs), though only `tasks.md` content drives the progress bars.
- `pollNormalModeContent` → `ReloadChange` for the visible change.
- `pollWorktrees` → `ReloadChange` for every change in every sibling worktree.

So with N changes the idle cost is ~4N reads + parses every half second. The observable contract — content reflects within ~500 ms — does not require reading unchanged files. `fileSystem.Stat` already returns `os.FileInfo` (with `ModTime()` and `Size()`), so a cheap signature check is available with no abstraction change.

## Goals / Non-Goals

**Goals:**
- An idle tick costs only a few `stat`s, not 4N reads+parses.
- No observable behavior change (latency, what updates, error surfacing all preserved).
- Uniform across index, normal-content, and worktrees modes.

**Non-Goals:**
- fsnotify / event-driven watching (see Decisions); configurable interval (#94 follow-up); gating the directory-list `ReadDir`s; content hashing.

## Decisions

### D1 — Signature gating, not filesystem watching
Keep the timer; before a `ReloadChange`, `stat` the file(s) that drive the refresh and skip the read+parse when `(mtime, size)` is unchanged.
- *Why over fsnotify:* on linux + darwin fsnotify is non-recursive (must walk + watch per dir, re-add on new dirs), needs Bubble Tea event plumbing, must watch multiple roots in worktrees mode plus the worktree-set change, and would still need a fallback poll. That is strictly more than gating, for a benefit (sub-500 ms latency, zero wakeups) the user does not need — the goal is killing the churn.

### D2 — `(mtime, size)` signature, not a hash
The gate is `stat`-based: modification time and size. A content hash would re-read the file, defeating the purpose.
- *Trade-off:* coarse-grained filesystems (1 s mtime) could miss a same-second, same-size edit until the next change; size catches most same-second edits and continued polling converges. Acceptable for markdown task files.

### D3 — Gate the two tasks-only, N-scaling paths
The cost that scales with the number of changes is the per-change `ReloadChange` in `pollIndexMode` (every active change, every tick) and `pollWorktrees` (every sibling change). Both only need `tasks.md` (index progress bars / worktree progress), so both gate on a single `tasks.md` signature per change. The single visible-change content reloads (`pollNormalModeContent`, `pollWorktreeChange`) are left unchanged: they reload one change per tick (non-scaling), and reliably gating their content — which includes nested `specs/**/spec.md` that `loadSpecs` walks — would need per-spec-file signatures, complexity not worth it for one reload.

### D4 — Two freshness caches on the model; explicit reloads bypass them
A `freshness` holds a `map[path]FileSig` (a not-found stat → the zero `FileSig`, `Present=false`). The model keeps **two**: `fresh` for index mode and `worktreeFresh` for worktrees mode, because each mode holds an independent in-memory copy of the current worktree's changes — a single shared cache could let a reload gated for one mode mask a needed reload for the other (returning to index after worktrees updated the file). First-seen path → treat as changed, load, record. The editor-return path (`editorReturnMsg`) re-reads unconditionally; gating applies only to the periodic tick, so a just-saved edit is always reflected.
- *Why lazy seeding (vs seeding at startup):* simpler; the first tick reads once then quiesces — harmless.
- *Why no explicit eviction:* a removed change's stale entry is never queried; a re-created change at the same path gets a new mtime/size and is detected as changed.

### D5 — Leave the directory-list `ReadDir`s unconditional
`pollIndexMode`'s three `ReadDir` name-comparisons (new/removed changes/specs/archives) stay every-tick: they are cheaper than file reads, and gating them on parent-dir mtime adds complexity for little gain. The big win is the per-change `ReloadChange`.

## Risks / Trade-offs

- **Coarse mtime granularity** → mitigated by also comparing size and by continued polling (D2).
- **Stale cache across a structural reload / mode switch** → addressed by the two-cache split (D4) plus mtime detection: a re-created change gets a new signature and is reloaded; a removed change's stale entry is simply never queried. No explicit eviction needed.
- **Symlinked/odd filesystems where mtime is unreliable** → worst case degrades to the current behavior's correctness on the next genuine change; never shows wrong content for longer than a real edit's mtime bump.

## Migration Plan

Additive, TUI-only, no new dependency, no spec-observable change. With the cache cold (startup) behavior matches today (one load), then ticks quiesce. Rollback is reverting the change. No macOS or golden-contract impact.

## Open Questions

- None outstanding; mechanism and scope are settled.
