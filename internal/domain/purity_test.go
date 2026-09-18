package domain_test

import (
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// nonTestGoFiles returns the package's own .go files, excluding test files.
// The test files themselves are allowed to import anything (this one imports os
// and go/build); the purity rule constrains the package, not its tests.
func nonTestGoFiles(t *testing.T) []string {
	t.Helper()

	pkg, err := build.ImportDir(".", 0)
	if err != nil {
		t.Fatalf("build.ImportDir(%q): %v", ".", err)
	}
	if len(pkg.GoFiles) == 0 {
		t.Fatal("no non-test .go files found in internal/domain")
	}
	return pkg.GoFiles
}

// TestDomainIsPure mechanically enforces the purity rule from ARCHITECTURE.md
// §2: internal/domain does no I/O and depends on no other internal package.
//
// The equivalent manual check is:
//
//	go list -f '{{join .Imports "\n"}}' ./internal/domain
//
// which must print nothing from the forbidden set below. Note that it must be
// .Imports (direct imports) and not `go list -deps`: fmt, for instance, pulls in
// os transitively, which says nothing about this package.
func TestDomainIsPure(t *testing.T) {
	pkg, err := build.ImportDir(".", 0)
	if err != nil {
		t.Fatalf("build.ImportDir(%q): %v", ".", err)
	}

	// Exact import paths domain may never import directly.
	forbiddenExact := map[string]string{
		"database/sql": "domain does no database access",
		"os":           "domain does no filesystem access",

		"nexus/internal/store":    "domain is the innermost layer (main -> service -> store -> domain)",
		"nexus/internal/service":  "domain is the innermost layer (main -> service -> store -> domain)",
		"nexus/internal/parse":    "domain is the innermost layer; parse depends on domain, not the reverse",
		"nexus/internal/platform": "domain is the innermost layer; platform is a leaf concern of main.go",
	}

	// Import path prefixes domain may never import directly.
	forbiddenPrefixes := map[string]string{
		"net":          "domain makes no network calls",
		"os/":          "domain does no filesystem or process access",
		"database/sql": "domain does no database access",
	}

	for _, imp := range pkg.Imports {
		if why, bad := forbiddenExact[imp]; bad {
			t.Errorf("internal/domain imports %q: %s", imp, why)
			continue
		}
		for prefix, why := range forbiddenPrefixes {
			if imp == prefix || strings.HasPrefix(imp, prefix+"/") {
				t.Errorf("internal/domain imports %q: %s", imp, why)
			}
		}
	}
}

// TestDomainReadsNoClock enforces the second half of the purity rule: time.Now
// is never called in domain. Anything needing the current time takes an
// injected now func() time.Time.
func TestDomainReadsNoClock(t *testing.T) {
	for _, name := range nonTestGoFiles(t) {
		src, err := os.ReadFile(filepath.Clean(name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if strings.Contains(string(src), "time.Now(") {
			t.Errorf("%s calls time.Now(): domain takes an injected now func() time.Time instead", name)
		}
	}
}
