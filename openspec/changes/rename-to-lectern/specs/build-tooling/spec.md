## MODIFIED Requirements

### Requirement: Binary named lectern
The project SHALL produce a binary named `lectern`. The entry point directory SHALL be `cmd/lectern/` so that `go install` names the binary by convention.

#### Scenario: go install produces correct binary name
- **WHEN** the developer runs `go install ./cmd/lectern/`
- **THEN** a binary named `lectern` is placed in `$GOPATH/bin`

#### Scenario: Binary is executable from PATH
- **WHEN** `$GOPATH/bin` is in `$PATH` and `make install` has been run
- **THEN** `lectern` is available as a command from any directory
