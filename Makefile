# Nexus — build entrypoints.
#
# ---------------------------------------------------------------------------
# SYSTEM PREREQUISITES (sudo required, installed once by the user):
#
#     sudo apt install pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev
#
# Until those land, `make build` and `make dev` WILL FAIL with a pkg-config /
# linker error. That is expected on a fresh machine, not a bug to debug.
# `make test` is unaffected and must stay green.
#
# ---------------------------------------------------------------------------
# WHY `-tags webkit2_41` IS MANDATORY (E1):
#
# Ubuntu 24.04 does NOT package `libwebkit2gtk-4.0-dev` — it does not exist in
# the archive. Only `libwebkit2gtk-4.1-dev` is available. Wails v2 probes for
# the webkit2gtk-4.0 pkg-config module by default and only looks for 4.1 when
# the `webkit2_41` build tag is set. Therefore EVERY `wails build` and
# `wails dev` invocation in this project must pass `-tags webkit2_41`.
#
# It is baked in below as $(TAGS) so nobody has to remember it. A wails command
# run without the tag failing to link is the expected outcome, not a defect.
#
# ---------------------------------------------------------------------------
# WHY $(WAILS) IS AN ABSOLUTE PATH (E2):
#
# The Wails CLI v2.16.0 lives at $(go env GOPATH)/bin/wails and is not
# guaranteed to be on PATH in non-login shells, so it is resolved explicitly
# rather than assumed callable as a bare `wails`.
# ---------------------------------------------------------------------------

WAILS ?= $(shell go env GOPATH)/bin/wails
TAGS  ?= webkit2_41

# No cgo, ever. The SQLite driver is modernc.org/sqlite (pure Go).
export CGO_ENABLED = 0

FRONTEND_DIR := frontend
DIST_DIR     := $(FRONTEND_DIR)/dist
BIN_DIR      := build/bin

.DEFAULT_GOAL := help

# Fail early and legibly if the Wails CLI is not where we expect it.
define require_wails
	@test -x "$(WAILS)" || { \
		echo "ERROR: wails CLI not found at '$(WAILS)'."; \
		echo "       Expected \$$(go env GOPATH)/bin/wails (see E2 in PLAN.md)."; \
		echo "       Install it with:"; \
		echo "         go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0"; \
		echo "       or override the path:  make $@ WAILS=/path/to/wails"; \
		exit 1; \
	}
endef

.PHONY: help
help: ## Show this help
	@echo "Nexus — available targets:"
	@echo
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
	@echo
	@echo "  WAILS = $(WAILS)"
	@echo "  TAGS  = $(TAGS)   (mandatory on every wails invocation — see E1 above)"
	@echo
	@echo "  Prerequisite: sudo apt install pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev"

.PHONY: dev
dev: ## Run the app in live-development mode (needs GTK/WebKit)
	$(require_wails)
	$(WAILS) dev -tags $(TAGS)

.PHONY: build
build: ## Build the production binary into build/bin (needs GTK/WebKit)
	$(require_wails)
	$(WAILS) build -tags $(TAGS)

.PHONY: test
test: $(DIST_DIR) ## Run the Go test suite
	go test ./...

.PHONY: clean
clean: ## Remove build/bin and frontend/dist
	rm -rf $(BIN_DIR) $(DIST_DIR)
	@mkdir -p $(DIST_DIR) && touch $(DIST_DIR)/.gitkeep

# main.go does `//go:embed all:frontend/dist`, so the directory must exist for
# `go build` / `go test` to compile package main at all. `clean` wipes its
# contents but recreates the directory for exactly this reason; this rule is the
# safety net for a fresh clone, where frontend/dist is gitignored and absent.
$(DIST_DIR):
	@mkdir -p $(DIST_DIR) && touch $(DIST_DIR)/.gitkeep
