-- Full-text search over nodes(title, description_md).
--
-- # Why FTS5 and not LIKE
--
-- S1-04 ran the spike rather than reading the documentation, on this exact
-- driver, and recorded:
--
--   FTS5: AVAILABLE
--   Driver: modernc.org/sqlite v1.53.0
--   Probe: CREATE VIRTUAL TABLE ... USING fts5 + MATCH query -> table created,
--          row inserted, MATCH returned the expected row, table dropped
--   Consequence for S1-17: FTS5 virtual table + triggers
--
-- So this is the FTS5 branch. The LIKE fallback that CLAUDE.md pre-approved is
-- not built: two search backends would be two behaviours to keep in step, and
-- the one that is never exercised is the one that would be broken.
--
-- # External content, and what that costs
--
-- content='nodes' makes this an EXTERNAL CONTENT table: the index stores the
-- tokens, and the column values themselves stay in `nodes`, where they already
-- are. Without it every title and description would be stored twice, and the
-- copy could disagree with the original.
--
-- The price is that SQLite does not keep the index up to date on its own — an
-- external content table trusts whoever writes it. That is what the three
-- triggers below are for, and it is why they are not optional decoration: an
-- index that can silently drift from its table is a bug generator, so the
-- triggers cover INSERT, UPDATE and DELETE, and nothing writes to `nodes`
-- without passing through one of them.
--
-- content_rowid is `nodes`.rowid, the implicit integer key every non-WITHOUT
-- ROWID table has. nodes.id is a uuid TEXT and cannot serve: FTS5 requires an
-- INTEGER rowid. The one hazard of that choice is VACUUM, which may renumber
-- the rowids of a table whose primary key is not INTEGER and would leave the
-- index pointing at the wrong rows. Nexus never issues VACUUM, and
-- SearchRepo.Rebuild exists so that recovering from one is a function call
-- rather than a re-install.

CREATE VIRTUAL TABLE nodes_fts USING fts5(
    title,
    description_md,
    content = 'nodes',
    content_rowid = 'rowid'
);

-- The 'delete' command rows below are FTS5's documented way of retracting a
-- document from an external content index: the OLD values have to be handed
-- back so that FTS5 can find and remove the exact tokens it indexed. Passing
-- the new values, or leaving the delete out of the UPDATE trigger, leaves the
-- old words matching forever — which is precisely the drift the sync test
-- looks for.

CREATE TRIGGER nodes_fts_after_insert AFTER INSERT ON nodes BEGIN
    INSERT INTO nodes_fts (rowid, title, description_md)
    VALUES (new.rowid, new.title, new.description_md);
END;

CREATE TRIGGER nodes_fts_after_delete AFTER DELETE ON nodes BEGIN
    INSERT INTO nodes_fts (nodes_fts, rowid, title, description_md)
    VALUES ('delete', old.rowid, old.title, old.description_md);
END;

CREATE TRIGGER nodes_fts_after_update AFTER UPDATE ON nodes BEGIN
    INSERT INTO nodes_fts (nodes_fts, rowid, title, description_md)
    VALUES ('delete', old.rowid, old.title, old.description_md);
    INSERT INTO nodes_fts (rowid, title, description_md)
    VALUES (new.rowid, new.title, new.description_md);
END;

-- Index whatever is already in `nodes`. On a fresh database that is nothing;
-- on a database that has been carrying data since 0002 it is everything, and
-- without this line those rows would be invisible to search forever while
-- every row added afterwards was findable — the worst kind of bug, because it
-- looks like search working.
INSERT INTO nodes_fts (nodes_fts) VALUES ('rebuild');
