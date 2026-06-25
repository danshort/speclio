# change-index Specification

## Purpose
Provides a full-screen index view (`ModeIndex`) that centralizes navigation between active changes, archived changes, and project specs, accessible with `a` or `Esc` from any other mode.
## Requirements
### Requirement: Vista índice de pantalla completa
The TUI SHALL implement a `ModeIndex` mode that occupies the full screen with the same TUI chrome (borders, header, helpbar). The index SHALL show three sections: "Active Changes (N)", "Specifications (N)", and "Archived Changes (N)", where N is the count of items in that section; followed by the active changes, specifications, and archived changes respectively; the three separated by a section line. When a section has zero items, the empty-state message (e.g., "No active changes") SHALL be shown instead of a count. When a view background color is configured, the entire index view SHALL render with that background color filling the full terminal viewport, including all whitespace areas between elements and the empty area below the box frame.

#### Scenario: Active changes section title shows count
- **WHEN** the mode is `ModeIndex` and there are 3 active changes
- **THEN** the "Active Changes" section title reads "Active Changes (3)"

#### Scenario: Archived changes section title shows count
- **WHEN** the mode is `ModeIndex` and there are 5 archived changes
- **THEN** the "Archived Changes" section title reads "Archived Changes (5)"

#### Scenario: Specifications section title shows count
- **WHEN** the mode is `ModeIndex` and there are 2 project specs
- **THEN** the "Specifications" section title reads "Specifications (2)"

#### Scenario: Zero items still shows empty-state message
- **WHEN** the mode is `ModeIndex` and there are no active changes
- **THEN** the "Active Changes" section shows "No active changes" without a count

#### Scenario: Índice con activos, archivados y specs
- **WHEN** the mode is `ModeIndex` and active changes, archived changes, and project specs exist
- **THEN** the screen shows an "Active Changes" section, a "Specifications" section, and an "Archived Changes" section in that order, within the TUI chrome

#### Scenario: Índice sin activos
- **WHEN** the mode is `ModeIndex` and there are no active changes
- **THEN** the "Active Changes" section shows a message indicating there are no active changes

#### Scenario: Índice sin archivados
- **WHEN** the mode is `ModeIndex` and there are no archived changes
- **THEN** the "Archived Changes" section shows a message indicating there are no archived items

#### Scenario: Índice sin specs
- **WHEN** the mode is `ModeIndex` and there are no specs in `openspec/specs/`
- **THEN** the "Specifications" section shows a message indicating there are no specs available

#### Scenario: Índice con fondo configurado
- **WHEN** the mode is `ModeIndex` and a view background color is configured
- **THEN** the entire terminal viewport renders with that background color, with no visible terminal-default background in whitespace areas or below the box frame

### Requirement: Formato de cambios activos en el índice
Each active change SHALL be displayed with its name on the left and a progress bar `[█░] N/M` on the right, using the same bar style as the tab bar. The item under the cursor SHALL be visually highlighted.

#### Scenario: Change activo con progreso parcial
- **WHEN** an active change has 6 out of 10 tasks completed and is under the cursor
- **THEN** it shows `▶ nombre-del-change  [██████░░░░] 6/10` with highlighted style

#### Scenario: Change activo sin tareas
- **WHEN** an active change has no `tasks.md`
- **THEN** the name is shown without a progress bar

### Requirement: Formato de cambios archivados en el índice
Each archived change SHALL be displayed with the clean name (without date prefix) on the left and the date in ISO 8601 `YYYY-MM-DD` format in secondary style on the right, aligned in two columns. The name column width SHALL adjust to the longest name in the archived list.

#### Scenario: Archivado con formato de fecha ISO 8601
- **WHEN** the archive directory is named `2026-05-02-specs-subnav`
- **THEN** the item shows `specs-subnav  2026-05-02` with the date in grey aligned to the right of the name

#### Scenario: Varios archivados con nombres de distinta longitud
- **WHEN** there are archived items with names of different lengths
- **THEN** all dates appear aligned in the same column

### Requirement: Navegación en el índice
The cursor SHALL be able to move through all items (active and archived) with `j` (down) and `k` (up). Section separators are not selectable items. The cursor SHALL NOT go past the first or last item.

#### Scenario: Navegar de activos a archivados
- **WHEN** the cursor is on the last active change and the user presses `j`
- **THEN** the cursor jumps to the first archived item

#### Scenario: Sin overflow en los extremos
- **WHEN** the cursor is on the last item and the user presses `j`
- **THEN** the cursor does not change

### Requirement: Seleccionar un change con Enter
Pressing `Enter` on an item or left-clicking an already-selected item SHALL close the index and open the selected change. If it is an active change, the mode switches to `ModeNormal` with that active change. If it is an archived change, the mode switches to `ModeViewingArchive` with that archived change.

#### Scenario: Seleccionar change activo
- **WHEN** the cursor is on an active change and the user presses `Enter` or left-clicks on it (when already selected)
- **THEN** the mode switches to `ModeNormal` showing that change

#### Scenario: Seleccionar change archivado
- **WHEN** the cursor is on an archived change and the user presses `Enter` or left-clicks on it (when already selected)
- **THEN** the mode switches to `ModeViewingArchive` showing the artifacts of that archived change

### Requirement: Helpbar del índice
The helpbar in `ModeIndex` SHALL show navigation hints and SHALL reflect the current sort mode via the `s` binding label:
- When sort mode is **name**: `j/k: navigate  Enter: open  Space: expand  click: select  s: sort by suffix  Esc: quit`
- When sort mode is **suffix**: `j/k: navigate  Enter: open  Space: expand  click: select  s: sort by name  Esc: quit`

#### Scenario: Helpbar en modo sort normal
- **WHEN** the mode is `ModeIndex` and the sort order is **name**
- **THEN** the helpbar shows `j/k: navigate  Enter: open  Space: expand  click: select  s: sort by suffix  Esc: quit`

#### Scenario: Helpbar en modo sort por sufijo
- **WHEN** the mode is `ModeIndex` and the sort order is **suffix**
- **THEN** the helpbar shows `j/k: navigate  Enter: open  Space: expand  click: select  s: sort by name  Esc: quit`

### Requirement: Actualización en tiempo real del índice
While the mode is `ModeIndex`, the TUI SHALL detect on each tick (≤ 500 ms) whether the list of active changes, the list of archived changes, or the list of project specs has changed on disk. If any structural change is detected, the index SHALL reload all three lists, rebuild the navigable items, and refresh the viewport without the user having to leave and re-enter `ModeIndex`. Additionally, when no structural change is detected, the TUI SHALL reload the task content of each active change from disk and, if any task content has changed, SHALL rebuild the index items and refresh the viewport so that progress bars reflect the latest task completion state. The cursor SHALL be preserved if the resulting index has at least as many items as the current position; otherwise it SHALL move to the last available item.

#### Scenario: Nuevo spec aparece en disco mientras el índice está abierto
- **WHEN** the mode is `ModeIndex` and a new directory is created in `openspec/specs/`
- **THEN** within a maximum of 500 ms the index shows the new spec in the "Specifications" section without user intervention

#### Scenario: Spec archivado desaparece de specs mientras el índice está abierto
- **WHEN** the mode is `ModeIndex` and a directory is deleted from `openspec/specs/`
- **THEN** within a maximum of 500 ms the spec disappears from the "Specifications" section

#### Scenario: Nuevo change archivado mientras el índice está abierto
- **WHEN** the mode is `ModeIndex` and a change is moved to `openspec/changes/archive/`
- **THEN** within a maximum of 500 ms the change appears in the "Archived Changes" section

#### Scenario: Nuevo change activo mientras el índice está abierto
- **WHEN** the mode is `ModeIndex` and a new change is created in `openspec/changes/`
- **THEN** within a maximum of 500 ms the change appears in the "Active Changes" section

#### Scenario: Cursor preservado cuando el ítem sigue existiendo
- **WHEN** the index reloads and the number of items does not decrease below the cursor position
- **THEN** the cursor stays at the same numeric position

#### Scenario: Cursor reajustado cuando el ítem desaparece
- **WHEN** the index reloads and the number of items is less than the current cursor position
- **THEN** the cursor moves to the last available item

#### Scenario: Tareas actualizadas en disco mientras el índice está abierto
- **WHEN** the mode is `ModeIndex` and the `tasks.md` file of an active change is externally modified (e.g., a checkbox is toggled)
- **THEN** within a maximum of 500 ms the progress bar for that change in the index reflects the updated `done/total` count without user intervention

### Requirement: Orden de cambios activos por fecha
Active changes in the index SHALL be displayed in creation date order, newest first, as provided by the loader.

#### Scenario: Índice con cambios de fechas variadas
- **WHEN** the index is rendered and active changes have different creation dates
- **THEN** the newest change appears first in the "Active Changes" section

#### Scenario: Cambio sin fecha aparece al final
- **WHEN** an active change has no `created` date
- **THEN** it appears after all dated changes in the "Active Changes" section

### Requirement: Filtrar el índice con /
While the mode is `ModeIndex`, pressing `/` SHALL enter a filter typing state where the help bar is replaced by a prompt showing `/` followed by the current query text. While in this typing state, every printable character typed SHALL be appended to the query, every `Backspace` SHALL remove the last character, and the index items SHALL be filtered in real-time using case-insensitive substring matching against the change name (active and archived), the spec name, and the requirement name. Pressing `Enter` SHALL confirm the query (filter stays applied, help bar returns with `[/query]` indicator). Pressing `Esc` while typing SHALL cancel the editing and restore the previous filter state. While a filter is applied but not typing, `Esc` SHALL clear it.

#### Scenario: Pressing / enters filter typing state
- **WHEN** the mode is `ModeIndex` and the user presses `/`
- **THEN** the help bar shows `/` with a visible input cursor and no filtering occurs yet

#### Scenario: Typing filters items in real-time
- **WHEN** the mode is `ModeIndex`, the user presses `/`, then types "foo"
- **THEN** within the same frame, only items whose name contains "foo" (case-insensitive) remain visible, and items without "foo" are hidden

#### Scenario: Enter confirms filter, Esc during typing cancels
- **WHEN** the mode is `ModeIndex`, the user types "bar" after `/` and presses `Enter`
- **THEN** the filter remains active with `[/bar]` indicator

#### Scenario: Esc with filter active clears it
- **WHEN** the mode is `ModeIndex`, a filter "foo" is active, and the user presses `Esc`
- **THEN** the filter is cleared and all items are shown

#### Scenario: Backspace during typing removes last character
- **WHEN** the mode is `ModeIndex`, the user types "foobar" after `/`, then presses `Backspace` twice
- **THEN** the query becomes "foob" and the filter updates accordingly

#### Scenario: / reopens editing with pre-filled query
- **WHEN** the mode is `ModeIndex`, a filter "foo" is active, and the user presses `/`
- **THEN** the filter typing state opens with the query pre-filled as "foo"

#### Scenario: Case-insensitive matching
- **WHEN** the mode is `ModeIndex`, a change named "MyFeature" exists, and the user types `/myfeature`
- **THEN** "MyFeature" is shown as a matching item

### Requirement: Secciones sin coincidencias muestran mensaje
When a filter is active and a section has no items matching the filter, that section SHALL display "No items match '<query>'" in help style. Sections that have no items without a filter SHALL keep their existing empty messages.

#### Scenario: Sección activa sin match muestra mensaje
- **WHEN** the mode is `ModeIndex`, a filter is active, and no active change matches it
- **THEN** the Active Changes section shows "No items match '<query>'" instead of the list

#### Scenario: Otras secciones con matches aún se muestran
- **WHEN** the mode is `ModeIndex`, a filter matches some specs but no active changes and no archived changes
- **THEN** the Active Changes and Archived Changes sections show the no-match message, while Specifications shows matching specs

### Requirement: Cursor preservado al filtrar
When a filter is applied or changed, the cursor SHALL be preserved on the same logical item if it still matches. If it no longer matches, the cursor SHALL move to the first item in the filtered list. If no items match, the cursor SHALL be set to 0.

#### Scenario: Cursor stays on same item when it still matches
- **WHEN** the cursor is on "data-export" and the user types `/data`
- **THEN** the cursor remains on "data-export" in the filtered list

#### Scenario: Cursor moves to first item when current item is filtered out
- **WHEN** the cursor is on "auth" and the user types `/data`
- **THEN** the cursor moves to the first matching item

### Requirement: Expand an archived change with Space
While the mode is `ModeIndex` and the cursor is on an archived change, pressing `Space` SHALL toggle the expansion of that archived change. When expanded, the index SHALL insert, immediately after the archived change row, one nested sub-item for each artifact type that is present on that change, in the fixed order proposal, design, specs, tasks. Artifact types that are not present SHALL be omitted. Pressing `Space` again SHALL collapse the archived change and remove its sub-items. After toggling, the cursor SHALL remain anchored on the toggled archived change.

#### Scenario: Expanding an archived change reveals its present artifacts
- **WHEN** the mode is `ModeIndex`, the cursor is on an archived change that has a proposal, specs, and tasks (but no design), and the user presses `Space`
- **THEN** three nested sub-items labelled "proposal", "specs", and "tasks" appear immediately below the archived change, in that order, and "design" is not shown

#### Scenario: Collapsing an archived change hides its artifacts
- **WHEN** the mode is `ModeIndex`, an archived change is expanded, and the user presses `Space` on it again
- **THEN** the nested artifact sub-items are removed and only the archived change row remains

#### Scenario: Cursor stays on the toggled archived change
- **WHEN** the mode is `ModeIndex`, the cursor is on an archived change, and the user presses `Space`
- **THEN** the cursor remains on that same archived change row

#### Scenario: Space on an archived change with no artifacts does nothing visible
- **WHEN** the mode is `ModeIndex`, the cursor is on an archived change that has no present artifacts, and the user presses `Space`
- **THEN** no sub-items are added and the cursor does not move

### Requirement: Navigate and display archived artifact sub-items
Archived artifact sub-items SHALL be navigable with `j` (down) and `k` (up) like any other index item, and SHALL be rendered indented below their parent archived change with the artifact-type name. The sub-item under the cursor SHALL be visually highlighted using the same cursor style as requirement sub-items.

#### Scenario: Navigate from archived change into its artifacts
- **WHEN** the mode is `ModeIndex`, an archived change is expanded, the cursor is on that archived change, and the user presses `j`
- **THEN** the cursor moves to the first artifact sub-item below it

#### Scenario: Artifact sub-item under the cursor is highlighted
- **WHEN** the mode is `ModeIndex` and the cursor is on an archived artifact sub-item
- **THEN** that sub-item is rendered indented with the highlighted cursor marker

### Requirement: Open an archived artifact with Enter
Pressing `Enter` on an archived artifact sub-item, or left-clicking it when already selected, SHALL switch the mode to `ModeViewingArchive` for the parent archived change with the active tab set to the selected artifact type.

#### Scenario: Enter on an artifact sub-item opens that tab
- **WHEN** the mode is `ModeIndex`, the cursor is on the "design" sub-item of an expanded archived change, and the user presses `Enter`
- **THEN** the mode switches to `ModeViewingArchive` for that archived change with the active tab set to `design`

#### Scenario: Click on a selected artifact sub-item opens that tab
- **WHEN** the mode is `ModeIndex` and the user left-clicks the already-selected "tasks" sub-item of an expanded archived change
- **THEN** the mode switches to `ModeViewingArchive` for that archived change with the active tab set to `tasks`

### Requirement: Filtering keeps archived artifact sub-items with matching parents
When a filter is active, an archived artifact sub-item SHALL be considered a match when its parent archived change name matches the filter or its artifact-type label matches the filter, so that the sub-items of a matching archived change remain visible while that change is expanded.

#### Scenario: Sub-items remain visible when the parent change matches
- **WHEN** the mode is `ModeIndex`, an archived change named "data-export" is expanded, and the user types `/data`
- **THEN** the "data-export" archived change and its visible artifact sub-items remain shown

### Requirement: Open the worktrees view from the index
While the mode is `ModeIndex`, pressing `w` SHALL switch the mode to `ModeWorktrees`. The index helpbar SHALL advertise this with a static `w` affordance. The index SHALL NOT compute or display live cross-worktree counts, so that opening the index does not trigger any cross-worktree discovery or polling.

#### Scenario: Pressing w opens the worktrees view
- **WHEN** the mode is `ModeIndex` and the user presses `w`
- **THEN** the mode switches to `ModeWorktrees`

#### Scenario: Helpbar advertises the worktrees view
- **WHEN** the mode is `ModeIndex`
- **THEN** the helpbar includes a static `w` entry for the worktrees view and shows no live cross-worktree counts

### Requirement: Unreadable change artifact marked in the index

When an active change has an artifact (`proposal.md`, `design.md`, `tasks.md`, or a `spec.md`) that exists but could not be read, the index SHALL show a `⚠` marker on that change's row, and SHALL NOT emit a `✗` validation marker or a false "missing artifact" result for the unreadable file (an unreadable file is a read failure, not a structural one).

#### Scenario: Unreadable change artifact shows a warning marker

- **WHEN** an active change's `proposal.md` exists but is unreadable and the index is rendered
- **THEN** the change's row shows a `⚠` marker and no spurious `✗`

#### Scenario: Genuinely missing artifact still validates as missing

- **WHEN** an active change's `proposal.md` does not exist
- **THEN** validation still reports it missing (unchanged from today), distinct from the unreadable case
