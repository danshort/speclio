## Context

The toolbar segmented picker always shows the **mode** (Active/Archived/Specs/Worktrees); the detail pane sets `.navigationTitle` to the selected artifact's name. The **project name** is shown nowhere persistently. #70 asks to elevate it so it stays visible across all navigation states.

## Decisions

### Two complementary homes for project identity

1. **Sidebar header** — primary. A prominent, non-selectable header row pinned at the top of the sidebar `List`, showing `model.project?.name` with a `books.vertical` icon. Mirrors the Xcode navigator root / Finder source-list pattern. Always visible while the sidebar is open. Rendered only when a project is open (the empty state already names its own situation).

2. **Window title + subtitle** — complement, so identity survives a collapsed sidebar. The detail pane sets:
   - `.navigationTitle(project.name)` — project as the primary title.
   - `.navigationSubtitle(locationTrail)` — the current location as a `·`-separated trail (mode, then change/spec, then artifact where applicable), e.g. `Active · macos-app · Design`, `Specs · shared-fixture-corpus`, `Worktrees`.

   The subtitle replaces the previous per-view `.navigationTitle(artifactName)`, so no information is lost — the artifact name moves from title to the end of the subtitle, and the title now carries the project.

### Why not a breadcrumb strip (option A)

A full breadcrumb bar below the toolbar would re-show the mode (already in the picker) and the artifact (already available), adding chrome to surface one missing field. The subtitle trail gives the same orientation in a native location with zero new chrome.

### Trail construction

A single helper derives the subtitle from `model.mode` and `model.selection` so the two homes never drift:
- `.artifact(ref)` → `mode · changeName · artifactLabel`
- `.projectSpec(name)` → `mode · name`
- `.worktree` / `.worktreeArtifact` → mode (+ worktree title / change · artifact)
- `.config` → `mode · Project Config`
- `nil` selection → just the mode

`artifactLabel` maps `ArtifactKind` to a display string (Proposal, Design, Tasks, or the spec capability name).

## Risks / Trade-offs

- **Sidebar collapse** hides the header — mitigated by the title-bar copy (decision 2).
- **Subtitle truncation** on narrow windows — acceptable; the sidebar header remains the full, untruncated source of project identity.
- Empty/no-project state already communicates "No project open"; the header and title only apply when a project is loaded.
