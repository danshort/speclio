## Context

The index view renders each archived change with its name and a creation date. The date string is produced by `parseArchiveName` in `internal/openspec/loader.go`, which parses the `YYYY-MM-DD` prefix of the archive directory name and reformats it with the Go layout `02/01/2006` (`DD/MM/YYYY`). The view layer (`internal/ui`) consumes the resulting `DisplayDate` string verbatim and only handles column alignment — it does not interpret the date.

## Goals / Non-Goals

**Goals:**
- Display archived-change dates in unambiguous ISO 8601 `YYYY-MM-DD` form.
- Keep the single source of truth for the display date in the loader.

**Non-Goals:**
- Locale-aware or configurable formatting.
- Any change to alignment, sorting, or other views.

## Decisions

- **Change the format layout in `parseArchiveName` only.** Replace the output layout `02/01/2006` with `2006-01-02`. The view already treats `DisplayDate` as an opaque string, so no view code changes are required for behavior — only test fixtures that hardcode the old format need updating.
  - *Alternative considered:* reformat in the view layer. Rejected — it would duplicate date-formatting logic and split responsibility; the loader is the natural single owner of the display string.
- **Reuse the already-parsed `time.Time`.** The directory prefix is already parsed with layout `2006-01-02`; the input prefix happens to equal the desired output, but we keep the parse-then-format round-trip so the displayed value stays normalized and decoupled from the raw directory string.

## Risks / Trade-offs

- [Existing tests assert `DD/MM/YYYY` fixtures (`internal/ui/view_test.go`)] → Update those fixtures to ISO form so they reflect the new contract; add a focused `parseArchiveName` unit test to lock the format.
- [Users accustomed to the old format] → Low impact; ISO 8601 is unambiguous and the change is cosmetic.
