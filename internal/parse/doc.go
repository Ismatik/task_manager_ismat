// Package parse implements the quick-add natural-language parser: it turns a
// single typed line such as
//
//	deploy KA Avto fri 15:00 !high #work >KA Avto ~2h
//
// into a structured node draft (date, priority, tags, parent project, estimate,
// type).
//
// Stage 0 delivers only a placeholder; the parser is Stage 4.
//
// # Imports
//
// parse may depend on nexus/internal/domain for its types. It must not import
// nexus/internal/store or nexus/internal/service — parsing produces a value, it
// does not persist one.
package parse
