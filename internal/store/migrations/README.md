# Migrations

Every `.sql` file in this directory is a schema migration. They are compiled into
the binary by `//go:embed` (see `../migrate.go`) and are never read from disk at
runtime.

Rules:

- **Name files `NNNN_snake_case.sql`**, zero-padded to four digits — `0001_settings.sql`.
  The padding makes filename order, byte order and version order the same thing, and
  the runner rejects anything that does not match.
- **Migrations are append-only.** A file that has shipped is never edited; correcting
  it means adding the next one. The runner applies each version exactly once and has
  no notion of re-running.
- **Do not write `BEGIN`, `COMMIT` or `ROLLBACK`.** The runner wraps every file in a
  transaction together with its `schema_migrations` row, so a file either applies
  completely and is recorded, or fails and leaves nothing behind.

This file is not a migration: the runner globs `*.sql` only.
