package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// fts5ProbeTable is the name of the scratch virtual table the probe builds.
//
// It is created in the temp schema, which SQLite keeps private to a single
// connection and discards when that connection closes. That is what makes the
// probe safe to run against the user's live database: it writes nothing into
// the database file, takes no write lock on it, leaves no row in
// schema_migrations or the main sqlite_master, and cannot collide with a real
// table now or with one a later migration invents.
const fts5ProbeTable = "nexus_fts5_probe"

// The row the probe writes and looks for again. The search term is deliberately
// not a word: it has to come back from the index because FTS5 tokenised and
// matched it, not because something matched loosely.
const (
	fts5ProbeTitle = "nexus fts5 probe"
	fts5ProbeBody  = "capability probe body zzqwfts5xx marker"
	fts5ProbeTerm  = "zzqwfts5xx"
)

// FTS5Available reports whether the SQLite build behind db has the FTS5
// full-text search module compiled in.
//
// It does not ask; it exercises the module end to end. On a scratch virtual
// table in the temp schema it runs CREATE VIRTUAL TABLE ... USING fts5, inserts
// a row, runs a real MATCH query, checks that the row comes back, and drops the
// table again. Anything less would only prove that the module could be named.
//
// Exactly one failure means "unavailable" and returns (false, nil): SQLite
// refusing the CREATE with "no such module: fts5". Every other failure is
// returned as an error. The distinction is the whole point of the function — a
// probe that answered "false" to a read-only file, a full disk or a revoked
// permission would silently and permanently downgrade search over a problem
// that had nothing to do with FTS5.
//
// The result of the spike on this project, which S1-17 reads and does not
// re-litigate:
//
//	FTS5: AVAILABLE
//	Driver: modernc.org/sqlite v1.53.0
//	Probe: CREATE VIRTUAL TABLE ... USING fts5 + MATCH query -> table created, row inserted, MATCH returned the expected row, table dropped
//	Consequence for S1-17: FTS5 virtual table + triggers
//
// The probe stays in the tree rather than being deleted with the spike: the
// driver can be upgraded, and a capability that was verified once and then
// assumed forever is a capability nobody is checking.
func FTS5Available(ctx context.Context, db *sql.DB) (bool, error) {
	// The scratch table lives in the temp schema, which belongs to one
	// connection, so every statement of the probe has to run on that same
	// connection. database/sql hands out whichever pooled connection it likes
	// per statement, so the connection is pinned for the duration.
	conn, err := db.Conn(ctx)
	if err != nil {
		return false, fmt.Errorf("store: fts5 probe: acquiring a connection: %w", err)
	}
	defer conn.Close() //nolint:errcheck // returning a connection to the pool

	return fts5AvailableOnConn(ctx, conn)
}

// fts5AvailableOnConn is FTS5Available on an already-pinned connection, so that
// the tests can inspect the temp schema of the very connection the probe ran
// on and prove the scratch table is really gone rather than merely invisible.
func fts5AvailableOnConn(ctx context.Context, conn *sql.Conn) (available bool, err error) {
	const (
		createSQL = `CREATE VIRTUAL TABLE temp.` + fts5ProbeTable + ` USING fts5(title, description_md)`
		insertSQL = `INSERT INTO temp.` + fts5ProbeTable + ` (title, description_md) VALUES (?, ?)`
		matchSQL  = `SELECT title FROM temp.` + fts5ProbeTable + ` WHERE ` + fts5ProbeTable + ` MATCH ?`
		dropSQL   = `DROP TABLE temp.` + fts5ProbeTable
	)

	if _, createErr := conn.ExecContext(ctx, createSQL); createErr != nil {
		if isNoSuchFTS5Module(createErr) {
			return false, nil
		}
		// Note that no cleanup runs here: the table was not created, and if
		// the CREATE failed because something of that name already existed,
		// dropping it would destroy data the probe does not own.
		return false, fmt.Errorf("store: fts5 probe: creating the scratch table: %w", createErr)
	}

	// From here the scratch table exists and has to go, on every path.
	// context.WithoutCancel so that a caller whose context expires mid-probe
	// still gets its temp schema cleaned up rather than leaving the table
	// behind on a connection that goes straight back into the pool.
	defer func() {
		if _, dropErr := conn.ExecContext(context.WithoutCancel(ctx), dropSQL); dropErr != nil {
			dropErr = fmt.Errorf("store: fts5 probe: dropping the scratch table: %w", dropErr)
			available, err = false, errors.Join(err, dropErr)
		}
	}()

	if _, insertErr := conn.ExecContext(ctx, insertSQL, fts5ProbeTitle, fts5ProbeBody); insertErr != nil {
		return false, fmt.Errorf("store: fts5 probe: inserting the probe row: %w", insertErr)
	}

	var got string
	switch matchErr := conn.QueryRowContext(ctx, matchSQL, fts5ProbeTerm).Scan(&got); {
	case errors.Is(matchErr, sql.ErrNoRows):
		// The module loaded and then did not work. That is not "unavailable",
		// it is broken, and it must not be answered with a quiet false.
		return false, fmt.Errorf("store: fts5 probe: MATCH %q matched nothing, but the row is there", fts5ProbeTerm)
	case matchErr != nil:
		return false, fmt.Errorf("store: fts5 probe: running the MATCH query: %w", matchErr)
	case got != fts5ProbeTitle:
		return false, fmt.Errorf("store: fts5 probe: MATCH returned title %q, want %q", got, fts5ProbeTitle)
	}

	return true, nil
}

// isNoSuchFTS5Module reports whether err is SQLite refusing CREATE VIRTUAL
// TABLE because the FTS5 module is not compiled into the driver.
//
// The test is on the message, not on a result code, because SQLite reports a
// missing module as a plain SQLITE_ERROR — the same code it uses for a syntax
// error or an existing table — so the code carries no information here. It is
// applied only to the CREATE statement and only to this one message: a missing
// module is the single failure that means "this build cannot do full-text
// search". Everything else is a problem with the database, not with FTS5, and
// is reported as such.
func isNoSuchFTS5Module(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no such module") && strings.Contains(msg, "fts5")
}
