# Tasks

## 1. Reusable macOS-app workflow
- [x] 1.1 Convert `macos-app-release.yml` to `on: workflow_call` (input: `version`) + `workflow_dispatch`; drop the `macos-app-v*` push trigger
- [x] 1.2 Normalize the version (strip a leading `v`) and build via `package.sh` at that version

## 2. Publish to the unified release + cask
- [x] 2.1 Upload `Lectern-<version>.zip` to the existing `v<version>` GitHub release (`gh release upload --clobber`)
- [x] 2.2 Compute the zip `sha256`, render `macos/Casks/lectern-app.rb` with real `version`/`sha256`, and push it to `danshort/homebrew-tap` under `Casks/` using `HOMEBREW_TAP_TOKEN`
- [x] 2.3 Point the cask `url` at the unified `v<version>` release asset

## 3. Wire into the release flow (non-blocking)
- [x] 3.1 Add a `macos-app` job to `release.yml`: `needs: release-please`, `if: release_created`, `uses: ./.github/workflows/macos-app-release.yml` with `version: <tag_name>`, `secrets: inherit`
- [x] 3.2 Ensure it is a sibling of `goreleaser` (no mutual dependency) so an app-build failure can't fail the CLI release

## 4. Verify
- [x] 4.1 `actionlint` (or YAML parse) passes for both workflows
- [x] 4.2 Confirm logic by inspection: version unification, non-blocking topology, cask render correctness (real release exercises it on the next `vX.Y.Z`)
