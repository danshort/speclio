## Context

Today `LecternApp` owns one `@StateObject AppModel` and injects it into a single `WindowGroup`. Every window/tab shares that instance, so the app is single-project and macOS auto-tabbing is useless. #69 wants multiple projects open at once; we chose **separate windows/tabs (Option A)** with **reopen-all-on-launch**. #70 already made the project label per-view (sidebar header + window title), so the per-window UI is ready — this change is the plumbing underneath it.

## Decisions

### Per-window model via value-based WindowGroup

Switch to `WindowGroup(for: ProjectRef.self) { ref in RootView(ref: ref) }`. `RootView` owns `@StateObject AppModel` and loads `ref` on appear. SwiftUI then:
- gives every window/tab its **own** `AppModel` (independent mode, selection, watcher, security scope), and
- gives each `ProjectRef` a **single window** (open the same ref → focus, not duplicate).

`ProjectRef` is a small `Codable & Hashable` identifier. It carries the **stable project id = the folder path**, used both as the window's identity and as the key into the bookmark store.

### Reopen-all: self-managed, not OS window restoration

macOS's built-in window restoration is **not** usable here: it requires a signed `.app` bundle and a clean quit (it keys saved state off the bundle id), so it never works under `swift run`, and a window opened *in place* (the first project, into the launch window) carries no `ProjectRef` value to restore anyway. So reopen-all is **self-managed**:
- `ProjectStore` keeps an ordered `openProjects` list in `UserDefaults`, written **eagerly on open** (`markOpen`) so even a hard kill preserves it. An entry is removed only when the **user** closes a window (`markClosed` in `AppModel.teardown`), not on app termination — distinguished by an `AppDelegate.isTerminating` flag set in `applicationShouldTerminate` (which runs before windows tear down).
- On launch the **launch window** (the one with no `ref`) calls `reopenProjectsOnLaunch`: it `openWindow(value:)`s each saved path and dismisses itself. Single-instance dedup means it's harmless if the OS also restored some.

This works identically under `swift run` and a packaged `.app`, and survives both ⌘Q and a terminal kill.

### Bookmarks: one per project, keyed by path

Replace the single `projectBookmark` key with a **dictionary store** in `UserDefaults`: `path → security-scoped bookmark data`. Flow:
- **Open** → `NSOpenPanel` → create bookmark, upsert into the store under its path → `openWindow(value: ProjectRef(path:))`.
- **Restore** → `reopenProjectsOnLaunch` reopens a window per saved path; each `RootView`/`AppModel` resolves the bookmark from the store by path, starts the security scope, and loads. Stale bookmarks refresh; unresolvable ones show the existing empty/error state rather than failing the launch.
- Each window **stops** its own security scope and watcher on close. Removing a project from the store is out of scope (handled implicitly when a bookmark goes stale / by a future recents UI).

### Open semantics (open without closing, one window per project)

- **Single instance per project.** `openWindow(value: ProjectRef(path:))` focuses the existing window when one already presents that `ProjectRef`, and only creates a new window otherwise. So opening an already-open project raises its window instead of duplicating it — free from value-based windows, no manual bookkeeping.
- **"Open Project…" (⌘O):** show `NSOpenPanel`, persist the bookmark, record the recent. Then, **if the focused window has no project loaded** (empty state, e.g. first launch or a fresh tab), load the project **in place** into that window's `AppModel` — no new window, no stray empty window. Otherwise `openWindow(value:)` (new window, or focus if already open). Same logic backs the toolbar Open button, the empty-state button, and the command.
- **Trade-off:** the in-place path doesn't consult other windows, so opening a project that is *already* open *from an empty window* shows a second view of it. Every other path (Open from a loaded window, Open Recent, restore) goes through `openWindow(value:)` and enforces single instance; this lone empty-window edge isn't worth the window-coordination complexity of an open-then-dismiss dance.
- **New Tab (⌘T) / Show Tab Bar:** kept enabled (automatic window tabbing stays on). A new tab is an empty-project window; ⌘O then fills it in place — so tabs now hold different projects, fixing the previously-dead menu item.

### Open Recent

Maintain an **ordered recents list of project paths** in `UserDefaults` (most-recent first, de-duplicated, capped), updated on every successful open. The same `path → bookmark data` store provides access. Render a standard **File ▸ Open Recent** submenu via `CommandGroup(after: .newItem)`: each entry calls `openWindow(value: ProjectRef(path:))` (so it also benefits from single-instance focusing), plus a **Clear Menu** item that empties the recents list (bookmarks for still-open windows are untouched).

### Command routing with focused values

App-level `.commands` can't reach a specific window's `AppModel`. Expose the focused window's model via `FocusedValues` (`@FocusedObject`/`@FocusedValue`), set from `RootView`. `Open` is inherently app-level (it calls `openWindow`); `Reload` and the file-actions menu already live in the per-window toolbar and stay there. The text-size commands keep using global `@AppStorage` — unchanged.

### What stays the same

- The entire per-window UI: sidebar, modes, detail pane, progress bars, worktrees, #70 label/title.
- `OpenSpecKit` / loader — `loadFrom(path)` already loads any project path; no domain changes.
- Non-sandboxed, bookmark-based access (Option C) — just one bookmark per open project now.

## Risks / Trade-offs

- **Restore correctness.** Self-managed reopen avoids the OS-restoration pitfalls (bundle id, clean quit) but adds its own: the `isTerminating` distinction between user-close and quit, and the one-shot launch-window trigger. Mitigated by eager `markOpen` (hard-kill safe) and a clear fallback — a window whose bookmark won't resolve shows the empty/error state, never a crash or blank launch.
- **Multiple concurrent security scopes.** Each window holds its own `startAccessingSecurityScopedResource`; must be balanced by a stop on window close to avoid leaks.
- **Multiple FSEvents watchers** (one per window) — acceptable; each is cheap and scoped to its project's `openspec/`.
- **Single instance per project.** Value-based windows give this for free (same `ProjectRef` → same window), so a project can't be opened twice; the only edge is the empty-window dismiss after Open, handled explicitly above.
- **Per-window vs. per-scene state.** Mode/selection are deliberately per-window (not persisted globally) so restored windows open at a sensible default; persisting per-window selection across launches is a possible later refinement.

## Limitations

- **Tab grouping is not restored.** Reopen-all brings every project back, but each as its own window — windows that were grouped into tabs reopen flattened. Preserving tab topology would require persisting tab-group membership and reconstructing it via AppKit (`NSWindow.addTabbedWindow`) after SwiftUI creates the windows, which is fragile (window↔ref matching, async timing) for what is transient layout state. Deferred to a follow-up — tracked in #77.
