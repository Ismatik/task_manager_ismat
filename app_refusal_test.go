package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"reflect"
	"slices"
	"testing"
)

// The mechanical half of D30: nothing lets a bound method skip refuse().
//
// # Why this test exists at all (K16, S3-30)
//
// S3-08 gave refusals their own wire code and named the failed operation, and
// it did so BY HAND, at every call site, and until this file nothing forced the
// next method to call it. The same block had already decided the same question
// the other way —
// S3-01 argued that an eleven-times-by-hand rule needed a guard and built check
// 7 for it — so the inconsistency, not the risk, is the finding. Block B is
// where new bound methods arrive (NodeDetail, the field writers, the type
// switcher, tags, attachments, the time log, search, the archive, the enum
// sets), which is exactly the condition under which a hand-applied rule decays.
//
// # Why it reads app.go instead of using reflection
//
// Reflection can see that a method exists and what it returns; it cannot see
// what its body does with an error. The rule being enforced here is about the
// body, so the source is the only place the answer is written down. Reading a
// file across a boundary to check a rule is already this repository's habit —
// layout_test.go:45 reads frontend/src to compare class names, and
// internal/service/refusal_test.go:151 reads frontend/src/store/call.ts to
// compare a wire marker — and this is the same technique in the same language,
// with no new dependency, no new tool and no code generation.
//
// # Why it is a Go test and not a ninth `make guard` check
//
// `make guard` greps frontend/src. app.go is Go, so this belongs in `go test`,
// which is gate 1 and therefore a stronger place than a non-gate target. The
// guard stays at eight checks.
//
// # Why it is structural rather than a grep for the word "refuse" (D17)
//
// D17 exists because checks 3a/3b of `make guard` fire on IDENTIFIER NAMES, and
// a green name-based heuristic was offered as proof of something it never
// looked at. So this does not grep. It parses app.go, enumerates every exported
// method on *App from the syntax tree, and inspects EVERY return statement in
// each body:
//
//   - `return refuse(…)` / `return refuse[T](…)` — the error is routed. Good.
//   - `return a.SomeOtherBoundMethod(…)` — delegation to a method this same
//     test checks. Good, and it is what TimerStart and TimerStop do.
//   - `return value, nil` — a literal nil second result. There is no error to
//     route. Good.
//   - anything else with two results — `return domain.Node{}, err` — is the
//     defect, and it is named with its method, its line and its text.
//
// A body containing one refuse() and one bare `return x, err` therefore fails,
// which a grep for the word would not catch.

// refuseExceptions names any bound method that legitimately cannot route its
// error through refuse(), mapped to the reason.
//
// IT IS EMPTY AND SHOULD STAY EMPTY. There is no such method today. D30 allows
// an exception only as an explicit, commented entry here naming the method and
// why — never by loosening the enumeration above, because loosening the
// enumeration is how the rule stops being enforced for methods nobody has
// written yet.
var refuseExceptions = map[string]string{}

// The file the bound surface lives in, and the floor its enumeration must
// clear.
//
// The floor is D32: a walk that finds nothing must be RED, not green. The floor
// is the only number this file states, and the test above is what enforces it —
// every other count here is left to a command rather than to prose, which is
// the lesson of e172592 and of D32. What the surface actually measures today:
//
//	grep -cE '^func \(a \*App\) [A-Z]'        app.go   # bound methods  -> 22
//	grep -cE '^[[:space:]]*return refuse[([]' app.go   # refuse() sites -> 28
//
// D30 and S3-30 say "27 times across the 25 bound methods". Those two numbers
// are reported rather than corrected here: PLAN.md and TASKS.md are the PM's,
// and the argument they support — a rule applied by hand at every site, with
// nothing forcing the next one — is unaffected by which two integers it is.
//
// 20 is deliberately below 22, so that removing a method is not a spurious red,
// and far above zero, so that a parse which silently matched nothing is a real
// one.
const (
	boundSurfaceFile     = "app.go"
	boundSurfaceFloor    = 20
	refuseHelper         = "refuse"
	appReceiverTypeIdent = "App"
)

func TestEveryBoundMethodRoutesItsErrorThroughRefuse(t *testing.T) {
	fset, file := parseBoundSurface(t)
	methods := boundMethodDecls(t, file)

	// D32. A parse that matched nothing, or a regexp-shaped mistake that
	// matched almost nothing, must not read as "every method passed".
	if len(methods) <= boundSurfaceFloor {
		t.Fatalf("enumerated only %d bound methods from %s, want more than %d — "+
			"the walk found (almost) nothing, which is a broken enumeration and not a clean bill of health",
			len(methods), boundSurfaceFile, boundSurfaceFloor)
	}

	for _, name := range sortedNames(methods) {
		decl := methods[name]

		t.Run(name, func(t *testing.T) {
			if why, skipped := refuseExceptions[name]; skipped {
				t.Skipf("%s is an explicit D30 exception: %s", name, why)
			}
			if !returnsAnError(decl) {
				// Which methods must return (T, error) is
				// TestEveryBoundMethodReturnsAnError's rule, asserted by
				// reflection a few lines away in app_test.go. Restating it here
				// would be the same rule spelled twice.
				t.Skipf("%s does not return an error, so there is nothing to route", name)
			}

			returns := methodReturns(decl)
			if len(returns) == 0 {
				t.Fatalf("%s has no return statement at all, so this test checked nothing about it", name)
			}

			receiver := receiverName(decl)
			for _, ret := range returns {
				if routesThroughRefuse(ret, receiver, methods) {
					continue
				}
				t.Errorf("%s:%d: %s returns an error that never passes through %s():\n\t%s\n"+
					"Every bound method's error must go through %s() so that service.Refuse can give it a wire code and the frontend can render it as a named refusal (D30, K16).\n"+
					"Write `return %s(a.svc.X.Y(a.context()))`, or on an early-out `return %s(zeroValue, err)`.",
					boundSurfaceFile, lineOf(t, fset, ret), name, refuseHelper, render(fset, ret),
					refuseHelper, refuseHelper, refuseHelper)
			}
		})
	}
}

// The enumeration above reads exactly one file, so a bound method added in a
// SECOND file would never be looked at — and the failure would be silence,
// which is the shape D32 forbids. Reflection sees the whole method set, so
// comparing the two is what keeps "every bound method" true rather than "every
// bound method that happens to live in app.go".
func TestTheRefuseWalkSawEveryBoundMethod(t *testing.T) {
	_, file := parseBoundSurface(t)
	walked := sortedNames(boundMethodDecls(t, file))

	var bound []string
	typ := reflect.TypeOf(&App{})
	for i := range typ.NumMethod() {
		bound = append(bound, typ.Method(i).Name)
	}
	slices.Sort(bound)

	if !slices.Equal(walked, bound) {
		t.Errorf("the methods %s enumerates are\n\t%v\nbut the bound surface is\n\t%v\n"+
			"A bound method outside %s is invisible to the refuse() walk. Move it into %s, or teach this test to read the file it lives in.",
			boundSurfaceFile, walked, bound, boundSurfaceFile, boundSurfaceFile)
	}
}

// A stale exception is an exemption nobody decided on: it names a method that
// no longer exists, and it would silently exempt a future method that reuses
// the name.
func TestEveryRefuseExceptionNamesALiveMethod(t *testing.T) {
	_, file := parseBoundSurface(t)
	methods := boundMethodDecls(t, file)

	for name, why := range refuseExceptions {
		if _, ok := methods[name]; !ok {
			t.Errorf("refuseExceptions names %q (%q), which is not a bound method in %s — delete the entry",
				name, why, boundSurfaceFile)
		}
	}
}

// parseBoundSurface reads and parses app.go, returning the file set (for
// positions) and the syntax tree.
func parseBoundSurface(t *testing.T) (*token.FileSet, *ast.File) {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, boundSurfaceFile, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing %s: %v", boundSurfaceFile, err)
	}
	return fset, file
}

// boundMethodDecls returns every EXPORTED method declared on *App, keyed by
// name. Exported is the whole test of "bound": Wails binds the exported method
// set of the struct it is given, which is why startup, context and onIPCMessage
// are lower-case in the first place.
func boundMethodDecls(t *testing.T, file *ast.File) map[string]*ast.FuncDecl {
	t.Helper()

	found := map[string]*ast.FuncDecl{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) != 1 || fn.Body == nil {
			continue
		}
		star, ok := fn.Recv.List[0].Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		if ident, ok := star.X.(*ast.Ident); !ok || ident.Name != appReceiverTypeIdent {
			continue
		}
		if !fn.Name.IsExported() {
			continue
		}
		found[fn.Name.Name] = fn
	}
	return found
}

// returnsAnError reports whether the declaration's last result is `error`.
func returnsAnError(decl *ast.FuncDecl) bool {
	results := decl.Type.Results
	if results == nil || len(results.List) == 0 {
		return false
	}
	last := results.List[len(results.List)-1]
	ident, ok := last.Type.(*ast.Ident)
	return ok && ident.Name == "error"
}

// receiverName returns the receiver's identifier, or "" for `func (*App) X()`.
func receiverName(decl *ast.FuncDecl) string {
	names := decl.Recv.List[0].Names
	if len(names) == 0 {
		return ""
	}
	return names[0].Name
}

// methodReturns collects the return statements that belong to the METHOD.
//
// Returns inside a function literal belong to that literal and are deliberately
// not collected: they are not this method's returns, and the method's own
// `return` is still walked. There are no function literals in app.go today.
func methodReturns(decl *ast.FuncDecl) []*ast.ReturnStmt {
	var out []*ast.ReturnStmt
	ast.Inspect(decl.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.ReturnStmt:
			out = append(out, node)
		}
		return true
	})
	return out
}

// routesThroughRefuse is the rule, and it is the only place it is written down.
func routesThroughRefuse(ret *ast.ReturnStmt, receiver string, methods map[string]*ast.FuncDecl) bool {
	switch len(ret.Results) {
	case 1:
		call, ok := ret.Results[0].(*ast.CallExpr)
		if !ok {
			return false
		}
		// `return refuse(...)`, and `return refuse[T](...)` with the type
		// argument written out, which CheckHabitToday and UncheckHabitToday
		// need because they pass an untyped nil.
		if calleeName(call.Fun) == refuseHelper {
			return true
		}
		// `return a.TimerCurrent()` — delegation to another bound method, which
		// this same test checks. TimerStart and TimerStop both do this.
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok && receiver != "" {
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == receiver {
				if _, bound := methods[sel.Sel.Name]; bound {
					return true
				}
			}
		}
		return false

	case 2:
		// `return service.TimerView{}, nil` — a literal nil error. There is
		// nothing to route, and refuse(x, nil) would return the same thing.
		ident, ok := ret.Results[1].(*ast.Ident)
		return ok && ident.Name == "nil"

	default:
		// Zero results is a naked return under named results, where the error
		// value is invisible at the return site; anything else is not the
		// (T, error) shape this file is about.
		return false
	}
}

// calleeName returns the identifier being called, seeing through the generic
// instantiation forms `f[T](…)` and `f[T, U](…)`.
func calleeName(fun ast.Expr) string {
	switch node := fun.(type) {
	case *ast.Ident:
		return node.Name
	case *ast.IndexExpr:
		return calleeName(node.X)
	case *ast.IndexListExpr:
		return calleeName(node.X)
	}
	return ""
}

// lineOf renders a node's line number so a failure points at a line of app.go.
func lineOf(t *testing.T, fset *token.FileSet, node ast.Node) int {
	t.Helper()
	return fset.Position(node.Pos()).Line
}

// sortedNames keeps subtest order, and therefore failure order, stable.
func sortedNames(methods map[string]*ast.FuncDecl) []string {
	names := make([]string, 0, len(methods))
	for name := range methods {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// render prints a node back as Go source, so a failure quotes the offending
// line rather than describing it.
func render(fset *token.FileSet, node ast.Node) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return "<unprintable>"
	}
	return buf.String()
}
