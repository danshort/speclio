## Why

Archived change dates on the index view render in `DD/MM/YYYY` format, which is ambiguous across locales (a reader cannot tell `02/05/2026` from a US-style month-first date). ISO 8601 (`YYYY-MM-DD`) is unambiguous, sorts lexically, and matches the date format already used in archive directory names. Resolves issue #16.

## What Changes

- Render archived change dates on the index in ISO 8601 `YYYY-MM-DD` format instead of `DD/MM/YYYY`.
- Update the `parseArchiveName` loader output to format the parsed date as `YYYY-MM-DD`.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `change-index`: The "Formato de cambios archivados en el índice" requirement changes the archived-change date format from `DD/MM/YYYY` to ISO 8601 `YYYY-MM-DD`.

## Impact

- `internal/openspec/loader.go`: `parseArchiveName` date layout (`02/01/2006` → `2006-01-02`).
- `internal/ui/view_test.go`: test fixtures using `DisplayDate` values in `DD/MM/YYYY` form.
- New unit coverage for `parseArchiveName` date formatting.
- No change to archive directory naming, navigation, or any other view.

## Non-goals

- Changing how dates are displayed anywhere other than the archived-changes section of the index.
- Adding configurability or locale-aware date formatting.
- Altering archive directory naming conventions or the date-prefix parsing rules.
