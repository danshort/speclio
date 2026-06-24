## Context

`updateViewer` (`internal/ui/viewer.go`) handles key input while viewing a change (shared by `ModeNormal` and `ModeViewingArchive`). It already implements `Tab` → `nextAvailableTab(+1)` and `Shift+Tab` → `nextAvailableTab(-1)`, both of which skip disabled tabs and wrap. `h`/`l` move between changes. Left/Right arrows are unbound. In bubbletea v2 these arrive as `KeyPressMsg.String()` values `"left"` / `"right"`, the same way `"up"`/`"down"` already reach the index and scroll handlers.

## Goals / Non-Goals

**Goals:**
- `←`/`→` switch tabs exactly like `Shift+Tab`/`Tab`.
- Reuse the existing `nextAvailableTab` logic (skip disabled, wrap).

**Non-Goals:**
- Repurposing `h`/`l`.
- Spec-file cycling via arrows (stays on `3`).

## Decisions

- **Fold arrows into the existing cases.** `case "tab", "right":` and `case "shift+tab", "left":` — the smallest possible change, guaranteeing identical behavior (disabled-skip and wrap) to the keys they mirror. No separate code path to drift.
- **No mode guard.** Like `Tab`, the arrows work wherever the change viewer is active (normal and archive), since both render the tab bar. This is consistent and harmless; the change-viewing footer (normal/tasks) is what advertises it, per the issue's "when viewing a change" framing.
- **Footer wording.** `1-4/Tab/←→: artifact` keeps all three input methods discoverable in the normal and tasks help bars.

## Risks / Trade-offs

- **[Low] Arrow collision with scrolling.** Up/Down already scroll (or move the task cursor); Left/Right were free, so binding them to tabs introduces no conflict.
- **[Low] Discoverability vs. footer length.** The footer grows slightly; acceptable, and the README documents the full set.
