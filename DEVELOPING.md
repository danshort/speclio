# Developing lectern locally

This guide is for hacking on lectern while you (likely) already have a stable
copy installed via Homebrew. The goal: change something, then run your dev build
**right next to** the global `lectern` without the two stepping on each other.

## Prerequisites

- **Go 1.25+** (`go version`)
- A clone of this repo:
  ```bash
  git clone https://github.com/danshort/lectern
  cd lectern
  ```

## TL;DR — the fast loop

`make build` compiles a binary named `lectern` into the **repo root**. Run it
with an explicit `./` so the shell uses your dev build, not the one on `PATH`:

```bash
make build      # -> ./lectern
./lectern        # your dev build (note the leading ./)
lectern          # the Homebrew-installed build (resolved from PATH)
```

So `./lectern` = what you're working on, `lectern` = the stable install. They
coexist with zero conflict because the dev binary lives in the repo, not on
your `PATH`.

Even faster, skipping the binary entirely:

```bash
go run ./cmd/lectern            # build + run in one step
go run ./cmd/lectern /path/to/change-dir
```

### Telling them apart at a glance

Release builds embed a version; local `make build` / `go run` builds don't, so
`--version` prints an empty version. To stamp your dev build:

```bash
go build -ldflags "-X main.version=dev" -o lectern ./cmd/lectern
./lectern --version     # -> lectern dev
lectern --version       # -> lectern 0.15.0   (the Homebrew one)
```

If you'd rather run a dev build from anywhere without `./`, give it a distinct
name so it never shadows the real one:

```bash
go build -o lectern-dev ./cmd/lectern
# put it somewhere on PATH, or call it directly:
./lectern-dev
```

## Running against a project

lectern looks for an `openspec/` directory in the current working directory, or
takes a path to a single change directory as an argument:

```bash
cd ~/some/openspec-project && /full/path/to/lectern   # whole project
/full/path/to/lectern ~/some/openspec-project/openspec/changes/my-change
```

This repo is itself an OpenSpec project, so you can dogfood from the repo root:

```bash
make build && ./lectern
```

## Tests, lint, format

```bash
make test     # go test -race -cover ./...
make lint     # golangci-lint run ./...   (needs golangci-lint installed)
make fmt      # goimports -w .            (needs goimports installed)
```

Run `make test` before opening a PR; CI runs the same checks plus `go vet` and a
`govulncheck` dependency scan.

## A change, end to end

```bash
# 1. edit code...
# 2. build your dev binary
make build
# 3. run it against a real project and eyeball the change
cd ~/some/openspec-project && /full/path/to/repo/lectern
# 4. run the tests
cd /full/path/to/repo && make test
```

## Avoiding conflicts with the Homebrew install

- **Prefer `./lectern`** (or `go run ./cmd/lectern`) for testing. It's
  unambiguous and never touches your stable install.
- **`make install` is a footgun here.** It runs `go install`, dropping a
  `lectern` into `$(go env GOPATH)/bin` (usually `~/go/bin`). If that directory
  comes before Homebrew on your `PATH`, a bare `lectern` will silently start
  resolving to your dev build. Check with `which lectern` / `type -a lectern`.
  To undo it: `rm "$(go env GOPATH)/bin/lectern"`.
- `make clean` removes the `./lectern` dev binary from the repo root.

## Cleaning up

```bash
make clean                              # remove ./lectern
rm "$(go env GOPATH)/bin/lectern"       # only if you ran `make install`
```

The Homebrew copy is unaffected by any of the above; manage it with
`brew upgrade lectern` / `brew uninstall lectern` as usual.

## The macOS app

The native reader lives under `macos/` as two SwiftPM packages (needs Xcode):

- **`macos/OpenSpecKit`** — the domain layer: a Swift port of Go's
  `internal/openspec` (loading, parsing, validation, task editing).
- **`macos/LecternApp`** — the SwiftUI/AppKit app built on top of it.

```bash
cd macos/OpenSpecKit && swift test     # domain + golden tests
cd macos/LecternApp  && swift build    # compile the app
cd macos/LecternApp  && swift run      # run it (⌘O to choose a project folder)
macos/LecternApp/scripts/package.sh 0.1.0   # → dist/Lectern.app + .zip
```

### The cross-language golden contract — read before touching the loader

`internal/openspec` (Go) and `OpenSpecKit` (Swift) must produce **byte-identical**
output for a shared corpus of fixtures under [`testdata/corpus/`](testdata/corpus/README.md).
CI runs both toolchains against the committed `testdata/corpus/golden/` files, so
a behavior change in **one** language without the other is a failing build.

When you change loader/domain behavior:

```bash
# 1. change Go behavior, then regenerate the goldens and REVIEW the diff:
go test ./internal/openspec/ -run TestGolden -update
# 2. mirror the same behavior in macos/OpenSpecKit (same logic, same strings), then:
cd macos/OpenSpecKit && swift test          # must reproduce the new goldens
```

Keep placeholder/diagnostic strings **path-free** (name the capability, not an
absolute path) so goldens stay portable across machines and languages — the test
normalizer only scrubs paths inside known prefixes. Add or extend a corpus
fixture whenever you add a behavior worth pinning.

CI for the app runs in the `swift` job: `swift build`, `swift run oskgolden`
(the executable golden byte-check), then `swift test`.
