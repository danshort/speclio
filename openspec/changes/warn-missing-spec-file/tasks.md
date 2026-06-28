# Tasks: warn instead of silently skipping a spec dir with no spec.md

## 1. Go loader

- [x] 1.1 Add `ErrNoSpecFile` sentinel and `missingSpecPrefix = "⚠ no spec.md in "` to `internal/openspec/loader.go`.
- [x] 1.2 In `loadSpecs`, change the `ErrNotExist` arm from `continue` to appending a `NamedSpec{Name, Content: missingSpecPrefix + name, ReadErr: ErrNoSpecFile}` and including it in the combined `parts`. (Placeholder names the capability — not a path — so the un-normalized combined `Specs` content stays portable.)
- [x] 1.3 In `LoadProjectSpecsFrom`, change the `ErrNotExist` arm to set `ps.Content = missingSpecPrefix + name` and `ps.ReadErr = ErrNoSpecFile`.
- [x] 1.4 Generalize `golden_test.go`'s `normContent` to normalize the `missingSpecPrefix` placeholder (prefix + relPath) as well as `unreadablePrefix`.

## 2. Corpus

- [x] 2.1 Add a corpus fixture: a change with a `specs/<cap>/` directory containing no `spec.md` (use a `.gitkeep` like `empty-capability`), and wire a golden test case for it.
- [x] 2.2 Regenerate goldens (`go test ./internal/openspec/ -run TestGolden -update`) and review the diff: `empty-capability` flips to a placeholder + `read_error:true`; the new fixture surfaces its spec.

## 3. Swift port

- [x] 3.1 Mirror the sentinel/placeholder in `OpenSpecKit/Sources/OpenSpecKit/Loader.swift` (`loadSpecs`, `loadProjectSpecsFrom`) and add `missingSpecPrefix` to `Layout.swift` (byte-identical string).
- [x] 3.2 Mirror the golden placeholder normalization in `OpenSpecKitGolden/Golden.swift`'s `normContent` (add the `missingSpecPrefix` branch) so the Swift port matches the regenerated goldens.

## 4. Verify

- [x] 4.1 `go build ./...`, `go vet ./internal/...`, `gofmt -l` clean; `go test ./internal/openspec/ ./internal/ui/` green.
- [x] 4.2 `swift test` (OpenSpecKit) green against the regenerated goldens.
- [x] 4.3 `openspec validate warn-missing-spec-file --strict` passes.
