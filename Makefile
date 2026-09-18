# Nexus — build entrypoints.
#
# `make check` is the single entry point: it runs the five gate commands in
# order and stops at the first failure.
#
#   1  go test     (repo root)      4  npm run typecheck   (frontend/)
#   2  go vet      (repo root)      5  wails build -tags webkit2_41
#   3  npm run lint (frontend/)
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
NODE_MODULES := $(FRONTEND_DIR)/node_modules
BIN_DIR      := build/bin

# The Go packages that are actually ours.
#
# `./...` is WRONG for this repo: frontend/node_modules contains real Go source
# — the `flatted` npm package ships golang/pkg/flatted — so `go list ./...`
# yields nexus/frontend/node_modules/flatted/golang/pkg/flatted and gates 1 and
# 2 would be testing and vetting somebody else's vendored code, which we neither
# own nor can fix. Recursively expanded (=, not :=) so that it is evaluated when
# a recipe runs, i.e. after $(DIST_DIR) exists and package main can be loaded.
GOPKGS = $(shell go list ./... | grep -v /node_modules/)

# A gate that can pass while checking nothing is worse than no gate at all.
#
# `$(shell ...)` swallows a failing `go list` and hands back the empty string,
# and `go test` / `go vet` with no package argument quietly falls back to the
# package in the current directory — so an empty $(GOPKGS) would turn gates 1
# and 2 into green no-ops that verified nothing. Gates therefore use
# $(CHECKED_GOPKGS), which is $(GOPKGS) when it is non-empty and a hard make
# error otherwise. Recursively expanded (=, not :=) so the $(error) fires only
# when a gate recipe actually runs, not on every parse of this file.
CHECKED_GOPKGS = $(if $(strip $(GOPKGS)),$(GOPKGS),$(error the Go package list is empty: `go list ./...` failed or matched nothing (see its error above). Refusing to run this gate — a bare `go test` / `go vet` would silently check only the current directory and report success))

.DEFAULT_GOAL := help

# The gates are ordered on purpose and must stay ordered even under `make -j`.
.NOTPARALLEL:

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

.PHONY: check
check: test vet lint typecheck build ## Run all five gates in order, stopping at the first failure
	@echo
	@echo "All five gates passed."

.PHONY: dev
dev: ## Run the app in live-development mode (needs GTK/WebKit)
	$(require_wails)
	@$(WAILS) dev -tags $(TAGS) || { $(MAKE) --no-print-directory apt-hint; exit 1; }

.PHONY: build
build: $(DIST_DIR) ## Gate 5 — build the production binary into build/bin (needs GTK/WebKit)
	$(require_wails)
	@$(WAILS) build -tags $(TAGS) || { $(MAKE) --no-print-directory apt-hint; exit 1; }

# Not a gate and not meant to be run directly: printed after a failed wails
# invocation, and only when the GTK/WebKit development packages really are the
# reason. There is no point shouting about apt when the build broke on a Go
# compile error.
.PHONY: apt-hint
apt-hint:
	@command -v pkg-config >/dev/null 2>&1 \
		&& pkg-config --exists gtk+-3.0 \
		&& pkg-config --exists webkit2gtk-4.1 \
		&& exit 0; \
	echo ""; \
	echo "  The GTK/WebKit development packages are missing, which is why this"; \
	echo "  failed. Install them with:"; \
	echo ""; \
	echo "    sudo apt install pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev"; \
	echo ""

.PHONY: test
test: $(DIST_DIR) ## Gate 1 — run the Go test suite
	go test $(CHECKED_GOPKGS)

.PHONY: vet
vet: $(DIST_DIR) ## Gate 2 — run go vet
	go vet $(CHECKED_GOPKGS)

.PHONY: lint
lint: $(NODE_MODULES) ## Gate 3 — run ESLint over the frontend
	cd $(FRONTEND_DIR) && npm run lint

.PHONY: typecheck
typecheck: $(NODE_MODULES) ## Gate 4 — type-check the frontend
	cd $(FRONTEND_DIR) && npm run typecheck

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

# Gates 3 and 4 need the dev dependencies. On a fresh clone node_modules is
# absent and `npm run lint` would fail with something unhelpful about a missing
# eslint, so install first. `npm ci` rather than `npm install`: it installs
# exactly what package-lock.json pins and never rewrites the lockfile.
$(NODE_MODULES):
	cd $(FRONTEND_DIR) && npm ci
