# Design: surface a spec directory with no spec.md

## The two loaders

Both spec-loading paths have a three-way switch on the `spec.md` read result.
The `ErrNotExist` arm is where they diverge today:

| Loader | `spec.md` not found | Effect |
|---|---|---|
| `loadSpecs` (change delta specs) | `continue` | directory **dropped** |
| `LoadProjectSpecsFrom` (project specs) | leave `Content` empty, still append | silently **empty**, no ⚠ |
| both, on a non-not-found read error (`default`) | placeholder + `ReadErr` set | surfaced with ⚠ ✓ |

The fix makes the `ErrNotExist` arm behave like the `default` arm: surface the
directory as a *present* spec with a placeholder and a read-error marker.

## Sentinel + placeholder

A new sentinel error distinguishes "directory has no spec.md" from a real IO
failure, and a distinct placeholder prefix gives a clear message:

```go
var ErrNoSpecFile = errors.New("spec directory has no spec.md")
const missingSpecPrefix = "⚠ no spec.md in "
```

- `loadSpecs`: in the `ErrNotExist` arm, append a `NamedSpec{Name, Content:
  missingSpecPrefix + specDir, ReadErr: ErrNoSpecFile}` (and include it in the
  combined `parts`), instead of `continue`.
- `LoadProjectSpecsFrom`: in the `ErrNotExist` arm, set `ps.Content =
  missingSpecPrefix + specDir` and `ps.ReadErr = ErrNoSpecFile` before the
  existing unconditional append.

`ReadErr != nil` is what the UI already keys the ⚠ marker off
(`specMarker`, `changeHasReadErr` in `internal/ui/index.go`), so no UI change is
needed — the directory now shows a ⚠ and, when opened, the placeholder.

## Golden contract

The committed golden corpus (`testdata/corpus/golden/*.json`) is shared by the
Go golden test and Swift's `GoldenTests`. The placeholder embeds an absolute
path, so the Go test's `normContent` reduces an unreadable placeholder to
`prefix + relPath`, dropping the OS-specific tail. `normContent` is generalized
to recognize `missingSpecPrefix` as well as `unreadablePrefix`, so the missing-
spec placeholder normalizes the same portable way. The Swift golden
normalization mirrors this.

Behavioral golden deltas (intended):
- `basic-project.project-specs.json`: `empty-capability` flips from
  `{content:"", read_error:false}` to `{content:"⚠ no spec.md in specs/empty-capability", read_error:true}`.
- A new corpus fixture adds a change whose `specs/<cap>/` directory has no
  `spec.md`, locking the change-delta path; its golden shows the spec surfaced
  with the placeholder rather than dropped.

## Swift mirror

`OpenSpecKit/Sources/OpenSpecKit/Loader.swift` has the exact equivalents
(`loadProjectSpecsFrom`, `loadSpecs`) with the same "leave empty" / "skip"
comments. They get the same sentinel/placeholder treatment, and `Layout.swift`
gains the matching `missingSpecPrefix` constant so the cross-language strings
are byte-identical.

## Testing

- Go: extend `golden_test.go` coverage — the new fixture's project + the
  modified `empty-capability` exercise both loaders; regenerate with
  `go test ./internal/openspec/ -run TestGolden -update` and review the diff.
- Swift: `swift test` (OpenSpecKitGolden) verifies the port produces the same
  goldens. Headlessly testable end-to-end — no manual GUI step required.
