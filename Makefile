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
	@echo "  COVER_MIN = $(COVER_MIN)   (make cover only; the five gates of make check are unchanged)"
	@echo
	@echo "  Prerequisite: sudo apt install pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev"

.PHONY: check
check: test vet lint typecheck build ## Run all five gates in order, stopping at the first failure
	@echo
	@echo "All five gates passed."

# ---------------------------------------------------------------------------
# COVERAGE — a measurement, NOT a sixth gate.
#
# `cover` is deliberately not a prerequisite of `check`. The gate definition is
# the five commands above and stays exactly five; Stage 1's >=90% ACCEPT is a
# separate, separately-run number. Adding cover to check would quietly redefine
# what "green" means in every earlier ticket and every review note.
#
# The two packages are measured SEPARATELY and each must clear the bar on its
# own — not combined, not averaged, not "the repo overall". internal/store is
# not measured here on purpose: rules live in domain and orchestration in
# service, so store is thin by design and its remaining error paths are
# driver-dependent.
#
# The comparison runs under LC_ALL=C. awk parses numbers with the locale's
# decimal separator, and under a comma-decimal locale — which this machine has
# — `awk` reads a coverage total of 77.5 as 77, silently measuring the wrong
# number against the threshold. This is a measurement concern internal to this
# target and is unrelated to K1's deferral, which is about the GTK window.
COVER_MIN  ?= 90.0
COVER_PKGS := domain service

.PHONY: cover
cover: ## Measure internal/domain and internal/service against COVER_MIN (not a gate)
	@status=0; \
	for pkg in $(COVER_PKGS); do \
		go test -covermode=atomic -coverprofile=coverage.$$pkg.out ./internal/$$pkg/... || exit 1; \
		total=$$(go tool cover -func=coverage.$$pkg.out | tail -1 | awk '{ print $$NF }' | tr -d '%'); \
		LC_ALL=C awk -v pkg="internal/$$pkg" -v total="$$total" -v min="$(COVER_MIN)" 'BEGIN { if (total + 0 < min + 0) { printf "  FAIL  %s is at %.1f%% statement coverage, below the required %.1f%% (short by %.1f points)\n", pkg, total, min, min - total; exit 1 } printf "  ok    %s is at %.1f%% statement coverage (required %.1f%%)\n", pkg, total, min }' || status=1; \
	done; \
	echo; \
	if [ $$status -ne 0 ]; then \
		echo "make cover FAILED: see the FAIL line(s) above. Threshold COVER_MIN = $(COVER_MIN)%."; \
		exit 1; \
	fi; \
	echo "Coverage threshold met: every measured package is at or above $(COVER_MIN)%."

# ---------------------------------------------------------------------------
# THE FRONTEND TEST SUITE — a measurement, NOT a sixth gate.
#
# Same precedent as `cover` above: `make check` is the five gate commands and
# stays exactly five. This is run alongside them and is an acceptance criterion
# on every frontend ticket from S2-10 onwards, which is a different thing from
# being a gate.
#
# vitest is interactive by default (it watches), so the non-interactive `--run`
# is baked in here — a target that never exits is a target no CI and no
# reviewer can use.

.PHONY: front-test
front-test: $(NODE_MODULES) ## Run the vitest suite in frontend/ (not a gate)
	cd $(FRONTEND_DIR) && npm run test -- --run

# ---------------------------------------------------------------------------
# GUARD — the mechanical rules check. Also NOT a gate.
#
# Stage 1 failed review three times on one defect: a rule written down twice and
# edited once. Stage 2's version of that defect is a rule re-derived in
# TypeScript, and the one defence that does not depend on a reviewer's attention
# is a grep that runs every time. That is all this is.
#
# Five checks. Checks 1, 2 and 5 are EXACT — they look for a literal that has no
# legitimate reason to exist in frontend/src, so a hit is a defect and there is
# nothing to argue about. Checks 3 and 4 are NECESSARILY HEURISTIC: "is this a
# computation or a field read" and "is this string user-visible" are not
# questions a regular expression can answer, and both are stated as heuristics
# on purpose. They are tuned to be tight enough to be actionable (every hit
# names a file and a line) and loose enough not to cry wolf, because a guard
# that produces false positives gets disabled, which is worse than no guard.
#
# Every check searches TRACKED AND UNTRACKED files (`git grep --untracked`).
# A guard that only sees `git add`-ed files is a guard a new component evades by
# simply not being staged yet, which is exactly when it is being written.
#
# Where a genuine exception is needed it goes in GUARD_ALLOW_RE below — a named,
# commented entry in one place — and never in an inline suppression comment
# scattered through the source.

# The identifier family that checks 3a and 3b look for: a name CONTAINING one of
# the derivation words. The first branch is the bare word; the second requires a
# lower-case first letter, which is what spares a React component called
# ProgressBar from being read as a computation.
GUARD_DERIVED_ID = ((overdue|derive|streak|progress|percent)|[a-z][A-Za-z0-9_$$]*(overdue|derive|streak|progress|percent|Overdue|Derive|Streak|Progress|Percent))[A-Za-z0-9_$$]*

# What makes a right-hand side a COMPUTATION rather than a read. Deliberately
# not "anything with an operator": `?.` and `??` are excluded (the generated
# Wails client sends null where it declares undefined, so `?? null` is the
# prescribed way to read a nullable field), `<`/`>` are only counted when
# surrounded by spaces so that JSX and TS generics do not trip it, and `/` and
# `*` likewise so that a path, a comment or a regex does not.
GUARD_COMPUTE_RE = ([[:space:]][<>]=?[[:space:]]|&&|\|\||[^?.]\?[^?.]|[[:space:]][-+/*][[:space:]]|\.filter\(|\.reduce\(|\.every\(|\.some\(|Math\.|new Date|Date\.)

# THE ALLOW-LIST. One ERE, alternation-separated, matched against the
# `file:line:text` output of any check; a matching hit is dropped. Every entry
# must carry a comment naming what it is and why.
#
# Empty at S2-10: nothing in frontend/src needs an exception yet, and an
# allow-list that starts out populated is an allow-list nobody reads.
GUARD_ALLOW_RE =

.PHONY: guard
guard: ## The mechanical rules greps over frontend/src (not a gate)
	@allow='$(GUARD_ALLOW_RE)'; \
	keep() { if [ -z "$$allow" ]; then cat; else grep -vE "$$allow" || true; fi; }; \
	status=0; \
	fail() { status=1; echo; echo "  guard FAIL — $$1"; echo "$$2" | sed 's/^/      /'; }; \
	out=$$(git grep -n --untracked -E '#[0-9a-fA-F]{3,8}' -- frontend/src | keep); \
	[ -z "$$out" ] || { fail "check 1, hex literal. Colours resolve only through the Tailwind token names (bg, surface, elevated, line, ink, muted, accent, accent-2, on-accent, danger, warning, success) — design/ owns the values." "$$out"; }; \
	out=$$(git grep -n --untracked -E "['\"\`](backlog|week|today|doing|done)['\"\`]" -- frontend/src ':!frontend/src/locales' | keep); \
	[ -z "$$out" ] || { fail "check 2, status string literal. Which strings are Kanban columns is domain.Status's answer; a quoted copy in TypeScript is a second spelling of it. Compare against a value Go returned, or key off the ColumnView the board handed you." "$$out"; }; \
	out=$$(git grep -n --untracked -E '(const|let|var)[[:space:]]+$(GUARD_DERIVED_ID)[[:space:]]*(:[^=]*)?=' -- frontend/src | grep -E '$(GUARD_COMPUTE_RE)' | keep); \
	[ -z "$$out" ] || { fail "check 3a (heuristic), a derived value COMPUTED rather than read. overdue, status, progress, percent and streak are fields Go already filled in on the DTO — see internal/service/dto.go. If the value you need is not on the DTO, the fix is a Go change, not a TypeScript one." "$$out"; }; \
	out=$$(git grep -n --untracked -E '(function[[:space:]]+$(GUARD_DERIVED_ID)[[:space:]]*\(|(const|let|var)[[:space:]]+$(GUARD_DERIVED_ID)[[:space:]]*(:[^=]*)?=[[:space:]]*(async[[:space:]]+)?(\([^()]*\)[[:space:]]*(:[^=]*)?=>|[A-Za-z0-9_$$]+[[:space:]]*=>|function))' -- frontend/src | keep); \
	[ -z "$$out" ] || { fail "check 3b (heuristic), a function named for a derivation. A function called deriveX/computeProgress/isOverdue/streakOf is a rule with a second implementation. Go owns the rule." "$$out"; }; \
	out=$$(git grep -n --untracked -E "[[:space:]](title|aria-label|aria-description|placeholder|alt)=[\"']" -- frontend/src | keep); \
	[ -z "$$out" ] || { fail "check 4a, a user-visible attribute holding a string literal. Every user-visible string is a key in frontend/src/locales/en.json AND ru.json: write title={t('...')}." "$$out"; }; \
	out=$$(git grep -n --untracked -E '[^=<>]>[^<>{}]*[[:alpha:]][^<>{}]*<' -- 'frontend/src/*.tsx' 'frontend/src/**/*.tsx' | grep -vE '^[^:]*:[0-9]+:[[:space:]]*(//|\*|/\*)' | keep); \
	[ -z "$$out" ] || { fail "check 4b (heuristic), a bare JSX text node. Put the text in en.json and ru.json and render {t('...')}." "$$out"; }; \
	out=$$(git grep -n --untracked -E "^[[:space:]]*[[:upper:]][[:alpha:]']*([[:space:]]+[[:alpha:]']+)+[.!?]?$$" -- 'frontend/src/*.tsx' 'frontend/src/**/*.tsx' | grep -vE '^[^:]*:[0-9]+:[[:space:]]*(//|\*|/\*)' | keep); \
	[ -z "$$out" ] || { fail "check 4c (heuristic), a line of bare prose in a .tsx file — a JSX text node on its own line. Same fix as 4b." "$$out"; }; \
	out=$$(git grep -n --untracked -E '(^|[^A-Za-z0-9_])P[0-2]([^A-Za-z0-9_]|$$)' -- frontend/src ':!frontend/src/locales' ':!frontend/src/lib/priority.ts' | keep); \
	[ -z "$$out" ] || { fail "check 5, a P0/P1/P2 chip label outside its one module. The priority->chip mapping (ARCHITECTURE.md section 6: 1->P0, 2->P1, 3->P2, 4->no chip) lives in frontend/src/lib/priority.ts and nowhere else." "$$out"; }; \
	echo; \
	if [ $$status -ne 0 ]; then \
		echo "make guard FAILED: see the FAIL block(s) above."; \
		echo "Checks 3 and 4 are heuristics. If a hit is genuinely a false positive, add a"; \
		echo "commented entry to GUARD_ALLOW_RE in the Makefile — never an inline suppression."; \
		exit 1; \
	fi; \
	echo "make guard: all five mechanical rules checks passed over frontend/src."

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
