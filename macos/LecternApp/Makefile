# Convenience targets for the Lectern macOS app.
# Run from macos/LecternApp/ — e.g. `make run`, `make app`.
#
# The npm-script equivalent for a Swift Package: `make <target>`.

VERSION ?= 0.0.0-dev

.PHONY: run app help

## run: build + launch the app in debug mode for fast iteration (Ctrl-C to quit)
run:
	@scripts/run-dev.sh

## app: build the signed .app bundle and reveal it in Finder (VERSION=x.y.z)
app:
	@scripts/build-and-open.sh $(VERSION)

## help: list available targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /'

.DEFAULT_GOAL := help
