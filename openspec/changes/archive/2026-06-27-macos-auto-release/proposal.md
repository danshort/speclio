## Why

The macOS `.app` isn't wired into the release flow: `macos-app-release.yml` only runs on a manual `macos-app-v*` tag, and nothing updates the Homebrew **cask** (`lectern-app.rb` is stuck at `version "0.0.0"`, `sha256 :no_check`, and was never published to the `danshort/homebrew-tap` tap — it has no `Casks/`). So `brew install --cask danshort/tap/lectern-app` doesn't actually work, and a release ships a current CLI formula but a stale/absent app cask. We want merging the release PR to also build the app and keep its cask correct, automatically, at the same version as the CLI.

## What Changes

- **Build + publish the `.app` automatically on each release.** Add a macOS-app job to `release.yml` gated on release-please's `release_created`, building at the **release version** (unified with the CLI — no separate `macos-app-v*` line). It uploads `Lectern-<version>.zip` to the same `v<version>` GitHub release.
- **Keep the cask current automatically.** After packaging, compute the zip's `sha256`, render `lectern-app.rb` with the real `version`/`sha256`/`url`, and publish it to `danshort/homebrew-tap` under `Casks/`, so `brew install --cask danshort/tap/lectern-app` installs the just-released build.
- **Non-blocking by design.** The macOS-app job is independent of the GoReleaser/CLI job: if the app build (or later, notarization) fails, the CLI formula release is unaffected — preserving the original "a macOS outage can't block `brew install lectern`" property by coupling the *trigger*, not the *success*.
- **Interim unnotarized.** Per decision, this ships now with ad-hoc-signed builds and the existing Gatekeeper caveat in the cask; the notarize step (#67) slots into the same job later and the caveat is removed then.
- Convert `macos-app-release.yml` to a **reusable workflow** (`workflow_call`) invoked by `release.yml`, keeping `workflow_dispatch` for manual rebuilds; drop the standalone `macos-app-v*` tag trigger.

## Non-goals

- Notarization (#67) — slots into this same job once the account/secrets exist.
- Changing the CLI formula flow (already automated via GoReleaser).
- Any app behavior / `OpenSpecKit` / corpus change — this is release plumbing only.

## Capabilities

### Modified Capabilities

- `macos-app`: the distributable build is now produced and its Homebrew cask updated automatically with each release, version-unified with the CLI, without blocking the CLI release.

## Impact

- `.github/workflows/release.yml` — new non-blocking `macos-app` job (calls the reusable workflow on `release_created`).
- `.github/workflows/macos-app-release.yml` — becomes `workflow_call` + `workflow_dispatch`; uploads to the `v<version>` release; adds the cask sha256/version render + push to the tap.
- `macos/Casks/lectern-app.rb` — URL points at the unified `v<version>` release asset.
- No app/domain/corpus changes.
