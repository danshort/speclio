## Context

`release.yml` (on `main`) runs release-please on every push; when its release PR merges, `release_created` is true and a `goreleaser` job builds the CLI + updates the tap formula. `macos-app-release.yml` is separate, fires only on a manual `macos-app-v*` tag, and just packages + creates a standalone GitHub release — it never updates the cask. The tap has only `lectern.rb` (no `Casks/`), and `lectern-app.rb` carries placeholder `version "0.0.0"` / `sha256 :no_check`. Net: the app's Homebrew path is non-functional.

## Decisions

### Couple the trigger, not the success

Add a `macos-app` job to `release.yml` as a **sibling** of `goreleaser` — both `needs: release-please` and `if: release_created`, neither depends on the other. A macOS build/notarization failure therefore cannot fail or roll back the CLI formula release. This keeps the original decoupling guarantee while making the app build automatic.

### Reusable workflow, not duplicated logic

Convert `macos-app-release.yml` to `on: workflow_call` (inputs: `version`) **+** `workflow_dispatch` (manual rebuilds), and drop the `push: tags: macos-app-v*` trigger. `release.yml`'s `macos-app` job calls it with `version: ${{ needs.release-please.outputs.tag_name }}` and `secrets: inherit`. Single source of build logic.

### Unified version

The job normalizes the tag (`v0.20.0` → `0.20.0`) and builds at that version, so the app and CLI always share a version. The zip is uploaded to the **existing `v<version>` GitHub release** (created by release-please, the same one GoReleaser appends CLI binaries to) via `gh release upload`, rather than a separate `macos-app-v*` release.

### Cask kept current and actually published

After packaging:
1. `sha256` of `Lectern-<version>.zip` via `shasum -a 256`.
2. Render the cask from the in-repo `macos/Casks/lectern-app.rb` (single source) by substituting the real `version` and `sha256` (the `url` already resolves from `version`).
3. Clone `danshort/homebrew-tap` with `HOMEBREW_TAP_TOKEN` (the token GoReleaser already uses for the formula), write `Casks/lectern-app.rb`, commit, push.

So `brew install --cask danshort/tap/lectern-app` resolves to the just-released, version-matched zip.

### Interim unnotarized; notarization slots in later

`package.sh` already does ad-hoc when `SIGN_IDENTITY` is unset and Developer-ID when set. The cask retains its Gatekeeper caveat for now. When #67 lands, a `notarytool submit --wait` + `stapler staple` step is added to this same job and the caveat is dropped — no structural change.

## Risks / Trade-offs

- **Release-workflow changes can't be fully CI-tested pre-merge** (they only run on a real release). Mitigated by `actionlint`, keeping logic in shell steps that are individually simple, and the non-blocking design (a failure can't corrupt the CLI release). First real exercise is the `0.20.0` release.
- **Tap write access** depends on `HOMEBREW_TAP_TOKEN` being scoped to the tap repo (already true for the formula).
- **Unnotarized cask installs** are Gatekeeper-blocked until #67 — accepted (internal users, documented caveat).
- **`gh release upload` race** with GoReleaser appending to the same release — both only add distinct assets; `--clobber` guards re-runs.
