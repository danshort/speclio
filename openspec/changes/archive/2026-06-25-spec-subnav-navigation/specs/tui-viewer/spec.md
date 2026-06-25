## MODIFIED Requirements

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

## ADDED Requirements

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
