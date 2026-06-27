## ADDED Requirements

### Requirement: Automated release alongside the CLI
The macOS app SHALL be built and released automatically as part of each repository release, at the same version as the CLI, and its Homebrew cask SHALL be updated to that release so it can be installed with `brew install --cask`. The macOS app build SHALL NOT be able to block or fail the CLI release.

#### Scenario: App builds and publishes on release
- **WHEN** a release is created (the release PR is merged and a `v<version>` release is produced)
- **THEN** the macOS app is built at that same version and its `.app` archive is attached to that release

#### Scenario: Cask kept current
- **WHEN** the macOS app archive is published for a release
- **THEN** the Homebrew cask is updated with that version, the archive's checksum, and its URL, and published to the tap so `brew install --cask` installs the just-released build

#### Scenario: App build never blocks the CLI release
- **WHEN** the macOS app build (or its notarization) fails during a release
- **THEN** the CLI release and its Homebrew formula update still complete
