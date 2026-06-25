## MODIFIED Requirements

### Requirement: Formato de cambios archivados en el índice
Each archived change SHALL be displayed with the clean name (without date prefix) on the left and the date in ISO 8601 `YYYY-MM-DD` format in secondary style on the right, aligned in two columns. The name column width SHALL adjust to the longest name in the archived list.

#### Scenario: Archivado con formato de fecha ISO 8601
- **WHEN** the archive directory is named `2026-05-02-specs-subnav`
- **THEN** the item shows `specs-subnav  2026-05-02` with the date in grey aligned to the right of the name

#### Scenario: Varios archivados con nombres de distinta longitud
- **WHEN** there are archived items with names of different lengths
- **THEN** all dates appear aligned in the same column
