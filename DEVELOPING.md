# Developing speclio locally

This guide is for hacking on speclio while you (likely) already have a stable
copy installed via Homebrew. The goal: change something, then run your dev build
**right next to** the global `speclio` without the two stepping on each other.

## Prerequisites

- **Go 1.25+** (`go version`)
- A clone of this repo:
  ```bash
  git clone https://github.com/danshort/speclio
  cd speclio
  ```

## TL;DR — the fast loop

`make build` compiles a binary named `speclio` into the **repo root**. Run it
with an explicit `./` so the shell uses your dev build, not the one on `PATH`:

```bash
make build      # -> ./speclio
./speclio        # your dev build (note the leading ./)
speclio          # the Homebrew-installed build (resolved from PATH)
```

So `./speclio` = what you're working on, `speclio` = the stable install. They
coexist with zero conflict because the dev binary lives in the repo, not on
your `PATH`.

Even faster, skipping the binary entirely:

```bash
go run ./cmd/speclio            # build + run in one step
go run ./cmd/speclio /path/to/change-dir
```

### Telling them apart at a glance

Release builds embed a version; local `make build` / `go run` builds don't, so
`--version` prints an empty version. To stamp your dev build:

```bash
go build -ldflags "-X main.version=dev" -o speclio ./cmd/speclio
./speclio --version     # -> speclio dev
speclio --version       # -> speclio 0.15.0   (the Homebrew one)
```

If you'd rather run a dev build from anywhere without `./`, give it a distinct
name so it never shadows the real one:

```bash
go build -o speclio-dev ./cmd/speclio
# put it somewhere on PATH, or call it directly:
./speclio-dev
```

## Running against a project

speclio looks for an `openspec/` directory in the current working directory, or
takes a path to a single change directory as an argument:

```bash
cd ~/some/openspec-project && /full/path/to/speclio   # whole project
/full/path/to/speclio ~/some/openspec-project/openspec/changes/my-change
```

This repo is itself an OpenSpec project, so you can dogfood from the repo root:

```bash
make build && ./speclio
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
cd ~/some/openspec-project && /full/path/to/repo/speclio
# 4. run the tests
cd /full/path/to/repo && make test
```

## Avoiding conflicts with the Homebrew install

- **Prefer `./speclio`** (or `go run ./cmd/speclio`) for testing. It's
  unambiguous and never touches your stable install.
- **`make install` is a footgun here.** It runs `go install`, dropping a
  `speclio` into `$(go env GOPATH)/bin` (usually `~/go/bin`). If that directory
  comes before Homebrew on your `PATH`, a bare `speclio` will silently start
  resolving to your dev build. Check with `which speclio` / `type -a speclio`.
  To undo it: `rm "$(go env GOPATH)/bin/speclio"`.
- `make clean` removes the `./speclio` dev binary from the repo root.

## Cleaning up

```bash
make clean                              # remove ./speclio
rm "$(go env GOPATH)/bin/speclio"       # only if you ran `make install`
```

The Homebrew copy is unaffected by any of the above; manage it with
`brew upgrade speclio` / `brew uninstall speclio` as usual.
