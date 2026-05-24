# change-index Specification (Delta)

## MODIFIED Requirements

### Requirement: Vista índice de pantalla completa
The TUI SHALL implement a `ModeIndex` mode that occupies the full screen with the same TUI chrome (borders, header, helpbar). The index SHALL show three sections: "Active Changes" with active changes, "Specifications" with the project specs in `openspec/specs/`, and "Archived Changes" with changes in `openspec/changes/archive/`; the three separated by a section line. When a view background color is configured, the entire index view SHALL render with that background color filling the full terminal viewport, including all whitespace areas between elements and the empty area below the box frame.

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
