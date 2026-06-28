## Context

A parallel code review of the recently merged work flagged four contained bugs (#114). Each has a small, local fix; this change batches them with a few comment/consistency cleanups. The deeper task-identity limitation is out of scope (#115).

## Goals / Non-Goals

**Goals:** correct the four bugs with regression tests where testable; tidy the flagged comments and inconsistencies. No behavior change beyond the fixes.

**Non-Goals:** duplicate-identity disambiguation and `ForEach(id:)` stability (#115); any new feature.

## Decisions

### B1 — Same-section downward reorder index correction
`moveTask` removes the source line, re-parses, then inserts at the destination section's post-removal `slots[toIndex]`. For a **same-section** move where the source's original position is **before** `toIndex`, removal shifts the destination slots left by one, so the effective insertion index must be `toIndex - 1`. Compute the source's ordinal position within its section before removal; when `fromPrefix == toPrefix && srcPos < toIndex`, use `toIndex - 1`. Upward moves (`srcPos >= toIndex`) and cross-section moves are unaffected. Pin with a downward-reorder test.

### B2 — Reap the detached launcher
After a successful `cmd.Start()` in the `OpenDetached` branch, spawn `go func() { _ = cmd.Wait() }()` so the short-lived handler (`open`/`xdg-open`) is reaped instead of lingering as a zombie. The goroutine ends when the launcher exits — bounded, one per open.

### B3 — Explicit cancel flag for Esc
Introduce a transient `cancellingEdit` bool. `onExitCommand` sets it and clears `editingID`; the `onChange(of: editorFocused)` blur handler commits only when `!cancellingEdit`. This removes the dependency on SwiftUI's undocumented Esc-vs-blur event ordering.

### B4 — Engine-level newline sanitization
`editTaskText` and `addTask` collapse `\r` and `\n` to a single space in the description before splicing, enforcing the single-line task invariant at the domain boundary regardless of caller. (The UI already collapses, so this is defense-in-depth; it also makes the OpenSpecKit API safe to call directly.)

### Polish
`doToggle` uses `filepath.Join(ch.Path, openspec.FileTasks)` (cross-platform + constant reuse) and `clearErrAfter()` (DRY). `ResolveOpener` trims its input. Three comments corrected to match behavior (`taskIdentity` strips all `~~`; `#nosec` note acknowledges the Windows `cmd /c start` shell; `addTask` doc drops the unimplemented fallback claim).

## Risks / Trade-offs

- **B1 index math** → covered by upward + downward + cross-section + end-of-section tests so the correction can't silently regress.
- **B3 flag lifetime** → reset `cancellingEdit` to false whenever an edit begins, so it never leaks into a later edit.

## Open Questions

- None; all four fixes are well-understood. Identity uniqueness remains deferred to #115.
