# tui-viewer Specification

## Purpose
Defines the layout and main behavior of the TUI: screen structure with borders and fixed zones, navigation between changes and tabs, markdown rendering with glamour, periodic polling for disk changes, and a welcome screen when there are no active changes.
## Requirements
### Requirement: Layout del TUI
The TUI SHALL divide the screen into fixed zones separated by horizontal lines: header (1 line), separator (1 line), tab bar (1 line), separator (1 line), content area (remainder), separator (1 line), help bar (1 line). In the `tasks` tab, a global progress bar is also added between the content area and the bottom separator. The header SHALL show `<project> · <change-name> [N/M]` where N is the position of the current change and M is the total number of active changes. The `View()` method SHALL return a `tea.View` struct with `AltScreen = true` and `BackgroundColor` set to the configured theme background color, instead of manually filling the background with padding.

#### Scenario: Separadores visibles entre zonas
- **WHEN** the TUI is rendered in any tab
- **THEN** a full-width horizontal line appears between the tab bar and the content, and another between the content and the help bar

#### Scenario: Alt screen and background color managed by tea.View
- **WHEN** the TUI renders any view
- **THEN** `tea.View.AltScreen` is set to `true` and `tea.View.BackgroundColor` reflects the configured theme

#### Scenario: Un solo change activo
- **WHEN** there is a single active change
- **THEN** the header shows `my-project · feat-a [1/1]`

#### Scenario: Varios changes activos
- **WHEN** there are three active changes and the second is selected
- **THEN** the header shows `my-project · feat-b [2/3]`

### Requirement: Navegación entre changes
The TUI SHALL allow navigating between active changes with `h` (previous) and `l` (next). Changing the change SHALL reset the selected tab to `proposal` if available, or to the first available artifact otherwise. Pressing `a` or `Esc` from `ModeNormal` SHALL open `ModeIndex`. Pressing `q` or `Ctrl+C` SHALL exit the application from any mode.

#### Scenario: Avanzar al siguiente change
- **WHEN** the user presses `l` while on change N
- **THEN** the TUI shows change N+1 (wrapping to the first if on the last)

#### Scenario: Retroceder al change anterior
- **WHEN** the user presses `h` while on change N
- **THEN** the TUI shows change N-1 (wrapping to the last if on the first)

#### Scenario: 'a' desde ModeNormal abre el índice
- **WHEN** the mode is `ModeNormal` and the user presses `a`
- **THEN** the mode switches to `ModeIndex`

#### Scenario: 'Esc' desde ModeNormal abre el índice
- **WHEN** the mode is `ModeNormal` and the user presses `Esc`
- **THEN** the mode switches to `ModeIndex`

#### Scenario: Salir con q desde cualquier modo
- **WHEN** the user presses `q` from any mode
- **THEN** the TUI exits

### Requirement: Tabs de artifact
The TUI SHALL show a tab bar with tabs `proposal`, `design`, `tasks`, `specs`. Tabs for absent artifacts SHALL be shown visually disabled and not selectable. The user SHALL be able to change tabs with keys `1`, `2`, `3`, `4`, with `Tab` / `→` (next available) and `Shift+Tab` / `←` (previous available), or by left-clicking on the tab label with the mouse. `Tab`, `Shift+Tab`, and the `←`/`→` arrows SHALL skip disabled tabs and wrap around at the ends; the arrows are secondary navigation that mirrors `Tab`/`Shift+Tab` and SHALL NOT cycle spec files. The `3` key SHALL select the `specs` tab exactly like the other numeric keys select their tabs (`1`→`proposal`, `2`→`design`, `4`→`tasks`); it SHALL NOT cycle specs. Moving between specs is the responsibility of the secondary sub-navigation (`[` / `]`) defined below and in the `specs-subnav` capability. If an absent artifact appears on disk during the session, the corresponding tab SHALL be enabled without needing to restart the TUI.

#### Scenario: Seleccionar tab disponible con tecla numérica
- **WHEN** the user presses `2` and `design.md` exists
- **THEN** the content area shows the rendered design

#### Scenario: Intentar seleccionar tab deshabilitada con tecla
- **WHEN** the user presses `2` and `design.md` does not exist
- **THEN** the tab does not change and no error occurs

#### Scenario: Seleccionar tab disponible con click del mouse
- **WHEN** the user left-clicks on the "design" tab label and `design.md` exists
- **THEN** the content area shows the rendered design

#### Scenario: Intentar seleccionar tab deshabilitada con click
- **WHEN** the user left-clicks on a disabled tab label and the artifact does not exist
- **THEN** the tab does not change and no error occurs

#### Scenario: Tab se habilita al aparecer artifact
- **WHEN** the TUI starts without `proposal.md` and an external process creates that file
- **THEN** within a maximum of 500 ms the `proposal` tab is shown as enabled and is selectable

#### Scenario: Tecla 3 desde otra tab va a specs
- **WHEN** the active tab is `proposal` and the user presses `3`
- **THEN** the active tab changes to `specs`

#### Scenario: Tecla 3 en specs no cambia el spec activo
- **WHEN** the active tab is already `specs`, the change has multiple specs, and the user presses `3`
- **THEN** the active tab remains `specs` and the active spec does not change

#### Scenario: Ciclar hacia adelante con Tab
- **WHEN** the active tab is `proposal`, `design` is disabled, and `specs` is available
- **THEN** the user pressing `Tab` changes the active tab to `specs` (skipping disabled `design`)

#### Scenario: Ciclar hacia atrás con Shift+Tab
- **WHEN** the active tab is `tasks`, `specs` is disabled, and `design` is available
- **THEN** the user pressing `Shift+Tab` changes the active tab to `design` (skipping disabled `specs`)

#### Scenario: Avanzar tab con flecha derecha
- **WHEN** the active tab is `proposal`, `design` is disabled, and `specs` is available
- **THEN** the user pressing `→` changes the active tab to `specs` (skipping disabled `design`), identically to `Tab`

#### Scenario: Retroceder tab con flecha izquierda
- **WHEN** the active tab is `tasks`, `specs` is disabled, and `design` is available
- **THEN** the user pressing `←` changes the active tab to `design` (skipping disabled `specs`), identically to `Shift+Tab`

#### Scenario: Tab da la vuelta al final
- **WHEN** the active tab is the last available tab and the user presses `Tab`
- **THEN** the active tab wraps around to the first available tab

#### Scenario: Shift+Tab da la vuelta al principio
- **WHEN** the active tab is the first available tab and the user presses `Shift+Tab`
- **THEN** the active tab wraps around to the last available tab

#### Scenario: Tab no actúa en modo configuración
- **WHEN** the mode is `ModeViewingConfig` and the user presses `Tab`
- **THEN** the tab does not change and the key is handled by the text input instead

### Requirement: Render de markdown con glamour
The TUI SHALL render `proposal`, `design`, and `specs` artifacts using glamour with the width of the content area. The content area SHALL be scrollable with `j`/`k` or the arrow keys.

#### Scenario: Scroll en contenido largo
- **WHEN** the artifact has more content than the screen height and the user presses `j`
- **THEN** the content scrolls down one line

#### Scenario: Wrap de glamour ajustado al ancho
- **WHEN** the terminal is 80 columns wide
- **THEN** glamour renders the markdown without exceeding those 80 columns

### Requirement: Pantalla de bienvenida sin changes activos
When the TUI starts and there are no active changes, it SHALL open directly in the index view (`ModeIndex`), showing the "Active Changes" section (empty), "Specifications", and "Archived Changes" sections within the full TUI chrome, with the index help bar. If the TUI enters `ModeNormal` while there are no active changes (e.g., all changes were deleted during the session), it SHALL show an informational message with the available actions.

#### Scenario: Arranque sin changes activos muestra el índice
- **WHEN** the TUI starts and `openspec/changes/` contains no active subdirectories
- **THEN** the TUI shows the index view with "Active Changes", "Specifications", and "Archived Changes" sections and the help bar `j/k: navigate  Enter: open  Space: expand  s: sort by suffix  i: info  Esc: quit`

#### Scenario: Sin changes activos desde ModeNormal
- **WHEN** the mode is `ModeNormal` and there are no active changes
- **THEN** the TUI shows `"No active changes. Create one with /opsx:propose"` and the help line `a/Esc: index  q: quit`

### Requirement: Salir del TUI
The user SHALL be able to exit the TUI at any time with `q` or `Ctrl+C`.

#### Scenario: Salir con q
- **WHEN** the user presses `q`
- **THEN** the TUI exits and the terminal is left in a clean state

#### Scenario: Salir con Ctrl+C
- **WHEN** the user presses `Ctrl+C`
- **THEN** the TUI exits and the terminal is left in a clean state

### Requirement: Barra de ayuda de teclado
The TUI SHALL show a fixed help line at the bottom listing the shortcuts active in the current context. On the `specs` tab, when the change has more than one spec, the help line SHALL advertise `[` / `]` for moving between specs. The help line SHALL NOT advertise the removed `3`-cycle.

#### Scenario: Tasks tab help line
- **WHEN** the active tab is `tasks` and the mode is `ModeNormal`
- **THEN** the help line includes a `h/l: change` hint, artifact navigation, `Space: toggle`, `e: edit`, and `Esc: index`

#### Scenario: Specs tab advertises spec navigation
- **WHEN** the active tab is `specs` with more than one spec and the mode is `ModeNormal`
- **THEN** the help line includes a `[` / `]` hint for previous/next spec

#### Scenario: Proposal or design tab help line
- **WHEN** the active tab is `proposal` or `design` and the mode is `ModeNormal`
- **THEN** the help line includes a `h/l: change` hint, artifact navigation, `j/k: scroll`, `e: edit`, and `Esc: index`

### Requirement: Polling periódico de artifacts
The TUI SHALL start a polling cycle every 500 ms on startup. On each tick it SHALL compare the on-disk content of the artifacts of the currently visible change with the in-memory content, AND detect changes in artifact presence (absent → present). If at the moment of the tick `len(m.project.Changes) == 0`, the tick SHALL attempt to reload the change list from disk and adopt the new state if at least one change is available. The cycle SHALL continue while the TUI is active.

#### Scenario: Tick sin cambios
- **WHEN** no file in the change has changed on disk
- **THEN** the TUI does not update any state or re-render anything

#### Scenario: Tick detecta cambio en tasks.md
- **WHEN** the content of `tasks.md` on disk differs from the in-memory content
- **THEN** the TUI re-parses the tasks, restores the cursor, and refreshes the view if the active tab is `tasks`

#### Scenario: Tick detecta cambio en artifact de markdown
- **WHEN** the content of `proposal.md`, `design.md`, or a `spec.md` on disk differs from the in-memory content
- **THEN** the TUI invalidates the corresponding entry in the render cache; the next time the user accesses that tab it is re-rendered with the new content

#### Scenario: Tick detecta aparición de artifact ausente
- **WHEN** an artifact that did not exist in the previous tick now exists on disk
- **THEN** the TUI updates the artifact presence state and enables the corresponding tab

#### Scenario: TUI arranca sin changes activos y se crea uno
- **WHEN** the TUI starts with `len(m.project.Changes) == 0` and during the session a change is created on disk
- **THEN** within a maximum of 500 ms the TUI reloads the change list and shows the new change

### Requirement: Actualización de tasks visible en tiempo real
When the TUI detects a change in `tasks.md` and the active tab is `tasks`, it SHALL refresh the view immediately without user intervention.

#### Scenario: Agente marca tarea como completada
- **WHEN** an external process changes `- [ ] tarea` to `- [x] tarea` in `tasks.md`
- **THEN** within a maximum of 500 ms the TUI shows the task as completed with the updated progress bar

### Requirement: Word wrap en todos los items de tarea
In the `tasks` tab, all task items SHALL word-wrap to the width of the content area (`m.width - 2`), regardless of whether the item is under the cursor or not.

#### Scenario: Item largo sin cursor
- **WHEN** a task item has more characters than the terminal width and the cursor is not on it
- **THEN** the text word-wraps and is shown in full across multiple lines

#### Scenario: Item largo con cursor
- **WHEN** a task item has more characters than the terminal width and the cursor is on it
- **THEN** the text word-wraps and is shown in full across multiple lines with the cursor style

### Requirement: Barra de progreso global en la vista de tasks
The TUI SHALL show a global progress bar as the first content line of the `tasks` tab, before any section. The bar SHALL reflect the total completed tasks over the total tasks in the change.

#### Scenario: Change con tareas parcialmente completadas
- **WHEN** a change has 3 completed tasks out of 8 total
- **THEN** the first line of the tasks view shows a progress bar with `3/8` and a proportional fraction of filled blocks

#### Scenario: Change con todas las tareas completadas
- **WHEN** all tasks in the change are marked as completed
- **THEN** the progress bar appears completely filled and shows the total as `N/N`

#### Scenario: Change sin tareas
- **WHEN** the change has no task items
- **THEN** no global progress bar is shown

### Requirement: Actualización inmediata del contador de progreso tras toggle
When the user toggles a task with `Space`, the progress counter in the tab bar SHALL update in the same frame, without waiting for the next polling cycle.

#### Scenario: Marcar tarea completa actualiza tab bar
- **WHEN** the user presses `Space` on a pending task and the disk write succeeds
- **THEN** the `N/M` counter and the progress bar in the tab bar update immediately in the same render

#### Scenario: Desmarcar tarea actualiza tab bar
- **WHEN** the user presses `Space` on a completed task and the disk write succeeds
- **THEN** the `N/M` counter and the progress bar in the tab bar decrement immediately in the same render

### Requirement: Viewport height derived from the rendered chrome rows

The TUI SHALL size the content viewport from the same ordered list of chrome rows that `View()` renders, rather than from an independently maintained sum of layout constants. In every mode that renders a content viewport, the total number of rendered lines SHALL equal the terminal height. The empty-project welcome view, which renders fixed content without a sized viewport, is the single exception.

#### Scenario: Total rendered height equals terminal height

- **WHEN** the TUI renders a viewport-backed mode (help overlay closed) at a terminal height at least as large as that mode's chrome-row count
- **THEN** the number of rendered lines equals the terminal height, with no clipped row or trailing blank line

#### Scenario: Below minimum height the viewport clamps

- **WHEN** the terminal height is smaller than the mode's chrome-row count
- **THEN** the viewport height clamps to one row (rendered content necessarily exceeds the terminal), and this degenerate case is exempt from the height-equality invariant

#### Scenario: Height invariant holds across every viewport-backed mode

- **WHEN** the TUI renders in normal (with active changes), archive, index, spec, config, or worktrees mode
- **THEN** in each mode the total rendered height equals the terminal height

#### Scenario: Empty-project welcome view is exempt

- **WHEN** there are no active changes and the welcome view is shown
- **THEN** the view renders fixed welcome content without a sized viewport, and the height-invariant test excludes this mode rather than asserting against it

#### Scenario: Optional spec subnav row is accounted for

- **WHEN** the specs tab is active and the spec subnav row is present
- **THEN** the viewport is exactly one row shorter and the total rendered height still equals the terminal height

### Requirement: Secondary sub-navigation convention
The viewer MAY render a secondary sub-navigation row (a chip row) inside the content block, below the primary artifact tab bar, for a tab that contains multiple sub-items. Wherever such a sub-nav is present, the user SHALL move between its items with `[` (previous item) and `]` (next item), wrapping around at the ends, and SHALL be able to select an item by left-clicking its chip. The primary tab-navigation keys (`1`–`4`, `Tab`/`Shift+Tab`, `←`/`→`) SHALL continue to operate on the primary tab bar only and SHALL NOT move between secondary sub-items. When no secondary sub-nav is present, `[` and `]` SHALL have no effect. The specs chip row is the only secondary sub-nav today (see the `specs-subnav` capability); any future sub-nav inherits this convention.

#### Scenario: `]` advances the secondary sub-nav
- **WHEN** a tab with a multi-item secondary sub-nav is active and the user presses `]`
- **THEN** the next sub-item becomes active, wrapping to the first after the last

#### Scenario: `[` goes back in the secondary sub-nav
- **WHEN** a tab with a multi-item secondary sub-nav is active and the user presses `[`
- **THEN** the previous sub-item becomes active, wrapping to the last before the first

#### Scenario: Primary keys do not move secondary sub-items
- **WHEN** a tab with a secondary sub-nav is active and the user presses `Tab`
- **THEN** the active artifact tab changes and the secondary sub-item selection is governed only by `[` / `]` and clicks

#### Scenario: `[` / `]` are inert without a sub-nav
- **WHEN** the active tab has no secondary sub-nav (e.g. `proposal`) and the user presses `[` or `]`
- **THEN** nothing changes and no error occurs

