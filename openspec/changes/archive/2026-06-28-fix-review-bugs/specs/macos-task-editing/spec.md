## MODIFIED Requirements

### Requirement: Reorder tasks within a section
The system SHALL let the user drag a task to a new position within its section and SHALL recompute the ordinals of that section so they reflect the new visual order. Reordering SHALL place the dragged task at the drop position regardless of drag direction — a downward drag SHALL land the task immediately before the drop target, not one position past it.

#### Scenario: Drag a task upward
- **WHEN** the user drags `1.3` above `1.1` in a section ordered `1.1`, `1.2`, `1.3`
- **THEN** the dragged task becomes `1.1`, the former `1.1` becomes `1.2`, and the former `1.2` becomes `1.3`

#### Scenario: Drag a task downward
- **WHEN** the user drags `1.1` to just above `1.3` in a section ordered `1.1 a`, `1.2 b`, `1.3 c`, `1.4 d`
- **THEN** the order becomes `b`, `a`, `c`, `d` (renumbered `1.1`–`1.4`) — the dragged task lands immediately before the drop target, not after it
