## Why

The TUI's 500 ms poll re-reads and re-parses on every tick whether anything changed or not: `pollIndexMode` calls `ReloadChange` (reading proposal/design/tasks/specs) for **every** active change, `pollNormalModeContent` reloads the visible change, and `pollWorktrees` reloads every change in every sibling worktree. With N active changes that's ~4N file reads + parses every half second, forever — wasted CPU and battery that scales badly. The observable goal (changes appear within ~500 ms) does not require re-reading unchanged files.

The fix is to keep the timer but make a tick cheap: detect changes with a file **signature** (mtime + size) `stat` and re-read/re-parse only what actually changed. This was chosen over filesystem watching (fsnotify): on the TUI's targets (linux + darwin) fsnotify is non-recursive, needs Bubble Tea event plumbing and multi-root worktree handling, and would still require a fallback poll — far more machinery for the same practical result, since killing the churn (not sub-500 ms latency) is the goal.

## What Changes

- Add a per-file freshness cache (`path → {mtime, size}`) on the model. On each tick, `stat` `tasks.md` for the relevant changes and **skip the `ReloadChange` (4 reads + specs walk) when the signature is unchanged**. This targets the two paths whose cost scales with the number of changes:
  - Index mode (`pollIndexMode`): gate each active change on its `tasks.md` signature — one cheap `stat` per change instead of `ReloadChange` for every change every tick. (The index progress bars only need `tasks.md`.)
  - Worktrees mode (`pollWorktrees`): gate each sibling change on its `tasks.md` signature. (The worktrees view only shows task progress.)
- A missing file (e.g. `tasks.md` deleted) counts as a change (present → absent).
- Explicit reloads — the editor-return path (`e`) — remain **unconditional**; gating applies only to the periodic tick.
- Keep the 500 ms interval (latency and the snappy feel are unchanged; an idle tick is now just a handful of `stat`s). A configurable interval is a possible future follow-up now that the config file exists (#94), but is out of scope here.
- No filesystem-abstraction change is needed: `fileSystem.Stat` already returns `os.FileInfo`, which exposes `ModTime()` and `Size()`.

## Capabilities

### New Capabilities
- `reload-freshness`: the periodic reload SHALL re-read/re-parse a change only when its `tasks.md` signature (mtime + size) has changed since last seen; this governs the index task-progress reload and the worktrees per-change reload. Explicit reloads bypass gating; deletions are detected; a first-seen file is treated as changed.

### Modified Capabilities
- `change-index`: the "Real-time index updates" requirement's per-tick task reload is refined to re-read only the changes whose `tasks.md` signature changed, rather than every change every tick.

## Impact

- **Code:** a freshness cache on `internal/ui` `Model` and a small `Loader.Signature(path)` helper (using the existing `fs.Stat` → `os.FileInfo`); gate the `ReloadChange` calls in `pollIndexMode` (`internal/ui/index.go`) and `pollWorktrees` (`internal/ui/worktrees.go`).
- **Tests:** with a counting/fake `fileSystem`, assert that an unchanged tick performs no re-read+parse, that a changed signature triggers exactly one reload, that a deleted file is detected, and that the editor-return path still reloads unconditionally.
- **No new dependency.** No change to the cross-language golden contract (signatures are runtime-only, never serialized). macOS app unaffected.

## Non-goals

- Gating the single visible-change content reload (`pollNormalModeContent`) and the open foreign-change reload (`pollWorktreeChange`) — these reload **one** change per tick (not N), so they don't drive the scaling churn, and reliably gating their *content* (which includes nested `specs/**/spec.md`) would need per-spec-file signatures — complexity disproportionate to a single-change reload. Left as-is.
- Filesystem watching / fsnotify — explicitly rejected for this change (see Why); event-driven could be revisited later if instant updates ever become a requirement.
- A configurable poll interval — future, building on #94.
- Gating the directory-list `ReadDir`s that detect new/removed changes/specs/archives — they are cheap relative to file reads and left unconditional in v1.
- Content hashing — `stat`-based mtime+size is the cheap signal; hashing would re-read the file and defeat the purpose.
