package store

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"nexus/internal/domain"
)

// ftsTable is the name of the full-text index created by migration 0003.
//
// It is referenced as a bare table name in every query, never through an alias.
// That is not style: this driver rejects `... FROM nodes_fts AS f WHERE f MATCH
// ?` with "no such column: f", which S1-04 hit and recorded. The MATCH operator
// wants the table's own name on its left-hand side.
const ftsTable = "nodes_fts"

// SearchOptions are the knobs on a search. The zero value is the one the UI
// uses: live nodes only, no limit.
type SearchOptions struct {
	// IncludeArchived brings archived nodes back into the results. Off by
	// default, like every other list in this package: archiving is how a user
	// hides something, and search is the easiest place to undo that by
	// accident.
	IncludeArchived bool

	// Limit caps the number of results. Zero or negative means no cap.
	Limit int
}

// SearchRepo is full-text search over nodes(title, description_md).
//
// # The backend is not part of the API
//
// Search takes a plain string and returns []domain.Node. Nothing in the
// signature, the options or the errors says "FTS5", and no caller has to write
// a MATCH expression, escape one, or know that a `*` means something. That is
// deliberate: the service layer above must be able to search without importing
// SQL, and the backend must be replaceable without touching it.
//
// # Ranking
//
// Results come back in bm25 order — SQLite's relevance ranking, which weighs a
// term by how rare it is and how short the field containing it is, so a word in
// a five-word title outranks the same word buried in a long description. Ties
// break on updated_at DESC then id, so that two equally relevant nodes never
// swap places between two runs of the same search.
type SearchRepo struct {
	exec Executor
}

// NewSearchRepo returns a repository over exec. Migration 0003 must have run.
func NewSearchRepo(exec Executor) *SearchRepo { return &SearchRepo{exec: exec} }

// WithExecutor returns a copy bound to exec.
func (r *SearchRepo) WithExecutor(exec Executor) *SearchRepo { return &SearchRepo{exec: exec} }

// Search returns the nodes whose title or description_md match query, most
// relevant first.
//
// The query is the user's raw typing. It is treated as literal text throughout:
// a query of `100% done` looks for the words "100" and "done", not for a
// wildcard; a query of `"` or `*` or `AND` is a word, or nothing, and never a
// syntax error thrown at somebody who was typing a task name. See ftsQuery.
//
// A query with nothing searchable in it — empty, whitespace, or only
// punctuation — returns an empty result rather than every row. Returning the
// whole table for an empty search box is how a search feature becomes a way to
// accidentally list ten thousand rows.
func (r *SearchRepo) Search(ctx context.Context, query string, opts SearchOptions) ([]domain.Node, error) {
	match, ok := ftsQuery(query)
	if !ok {
		return []domain.Node{}, nil
	}

	sql := "SELECT " + nodeColumns + " FROM " + ftsTable +
		" JOIN nodes ON nodes.rowid = " + ftsTable + ".rowid" +
		" WHERE " + ftsTable + " MATCH ?" +
		archivedFilter(opts.IncludeArchived) +
		" ORDER BY bm25(" + ftsTable + "), nodes.updated_at DESC, nodes.id"

	args := []any{match}
	if opts.Limit > 0 {
		sql += " LIMIT ?"
		args = append(args, opts.Limit)
	}

	rows, err := r.exec.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("store: searching for %q: %w", query, err)
	}
	defer rows.Close() //nolint:errcheck // the error surfaces from rows.Err below

	out := []domain.Node{}
	for rows.Next() {
		n, err := scanNode(rows)
		if err != nil {
			return nil, fmt.Errorf("store: searching for %q: %w", query, err)
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: searching for %q: %w", query, err)
	}
	return out, nil
}

// Rebuild reindexes every node from scratch.
//
// The index is kept current by the triggers migration 0003 installs, so this is
// a recovery path rather than routine maintenance. It is needed after anything
// that can renumber `nodes`.rowid — VACUUM, which Nexus never runs but a user
// with a sqlite3 shell might — and after restoring a database file from a
// backup that predates the index.
//
// It is a function rather than a note in a README because the alternative to
// having it is telling somebody to delete their database.
func (r *SearchRepo) Rebuild(ctx context.Context) error {
	const stmt = "INSERT INTO " + ftsTable + "(" + ftsTable + ") VALUES ('rebuild')"

	if _, err := r.exec.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("store: rebuilding the search index: %w", err)
	}
	return nil
}

// ftsQuery turns the user's typing into an FTS5 query expression, and reports
// whether anything searchable was left.
//
// # The rule: every token is a quoted phrase, and nothing else survives
//
// The input is cut into runs of letters and digits; everything else is a
// separator. Each run is wrapped in double quotes, which is FTS5's own syntax
// for "this is a string, not an operator", and the runs are joined with spaces,
// which FTS5 reads as AND. `report 2026` finds nodes containing both.
//
// That single rule is what makes every awkward input safe, without a list of
// characters to escape:
//
//   - `*`, `%`, `_`, `'`, `"`, `(`, `^`, `:` and friends are separators, so
//     they cannot become wildcards, operators or a syntax error. A search for
//     `50%` searches for "50".
//   - AND, OR, NOT and NEAR come through as quoted words, so a node titled
//     "and then" is findable by typing "and" instead of producing a query that
//     means something entirely different.
//   - Unicode is preserved: unicode.IsLetter is true for Cyrillic, so `Отчёт`
//     is one token and stays one token. The app is EN/RU and a tokenizer that
//     only understood ASCII would make half of it unsearchable.
//
// A quote inside a token is impossible by construction — quotes are separators
// — but the doubling is applied anyway, because "impossible by construction"
// stops being true the first time somebody widens the token rule.
func ftsQuery(input string) (string, bool) {
	tokens := strings.FieldsFunc(input, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	if len(tokens) == 0 {
		return "", false
	}

	quoted := make([]string, len(tokens))
	for i, t := range tokens {
		quoted[i] = `"` + strings.ReplaceAll(t, `"`, `""`) + `"`
	}
	return strings.Join(quoted, " "), true
}
