-- The core schema of PLAN.md §4: one `nodes` tree, plus the tables that hang
-- off it. Everything after this migration reads and writes these tables.
--
-- # One table for five kinds of thing
--
-- task, project, habit, note and bug are all rows in `nodes`, told apart by
-- `type`. A subtask and a project differ by type and position in the tree, not
-- by table. That is the modelling decision the whole product rests on: moving a
-- task under a project, or turning a note into a task, is an UPDATE rather than
-- a migration between tables.
--
-- # Derived state is not here
--
-- A parent's status and a subtree's progress are computed from the children
-- every time they are needed (D2, D7). There is no `progress` column and the
-- `status` column is meaningful on leaves only, so stored state can never drift
-- from the children it is supposed to summarise.
--
-- # Dates and timestamps are TEXT, and the two are not the same
--
-- SQLite has no date type, and mixing representations is how comparisons start
-- lying. This schema picks one of each and uses it everywhere:
--
--   timestamp  TEXT, ISO-8601 UTC, exactly 'YYYY-MM-DDTHH:MM:SSZ'
--   date       TEXT, exactly 'YYYY-MM-DD'
--
-- Both sort correctly as plain strings, which is what makes `due < :today` and
-- `ORDER BY started_at` mean what they read as. `nodes.due` and
-- `habit_checks.date` are DATES. `created_at`, `updated_at`, `completed_at`,
-- `archived_at`, `started_at` and `ended_at` are TIMESTAMPS.
--
-- # CHECK constraints on every enumerated column
--
-- The database is the last line of defence. A bug that writes status='Done'
-- instead of 'done' should fail loudly at the point it happens, not become a
-- row that nobody can explain a month later and that no view will ever show.
-- The Go enums in internal/domain and these constraints say the same thing
-- twice on purpose: one of them catches what the other lets through.

CREATE TABLE nodes (
    id             TEXT PRIMARY KEY,          -- uuid, generated in the service layer
    parent_id      TEXT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    type           TEXT NOT NULL CHECK (type IN ('task', 'project', 'habit', 'note', 'bug')),
    title          TEXT NOT NULL,
    description_md TEXT NOT NULL DEFAULT '',  -- Markdown; a bug's repro steps and severity live here as frontmatter
    status         TEXT NOT NULL CHECK (status IN ('backlog', 'week', 'today', 'doing', 'done')),
    due            TEXT NULL,                 -- date, 'YYYY-MM-DD'

    -- D1. Provenance of `due`, so that a move back to Backlog can clear a date
    -- the columns set and keep one the user typed. The default is 'manual'
    -- because a date that arrives without anyone saying where it came from is
    -- safer treated as the user's than as ours: being over-careful leaves a
    -- stale date on screen, being under-careful silently deletes the user's.
    due_source     TEXT NOT NULL DEFAULT 'manual' CHECK (due_source IN ('manual', 'auto')),

    priority       INTEGER NOT NULL DEFAULT 4 CHECK (priority BETWEEN 1 AND 4),  -- 1 is most urgent
    estimate_min   INTEGER NULL,              -- minutes
    recurrence     TEXT NULL,                 -- RRULE; required in practice for habits, enforced in the domain

    -- D4. The PMP KIT "Деятельность" field. These seven values are copied
    -- verbatim into the timelog form, so they are data rather than UI strings
    -- and are never translated. NULL means "not decided yet"; the default for
    -- a node's type is applied in the domain, not here, because it is
    -- overridable and a column DEFAULT could not express "depends on type".
    activity       TEXT NULL CHECK (activity IS NULL OR activity IN (
                       'Разработка',
                       'Анализ',
                       'Тестирование',
                       'Документация',
                       'Совещание',
                       'Согласование',
                       'Управление проектом'
                   )),

    sort_order     INTEGER NOT NULL DEFAULT 0,  -- position among siblings
    created_at     TEXT NOT NULL,               -- timestamp
    updated_at     TEXT NOT NULL,               -- timestamp
    completed_at   TEXT NULL,                   -- timestamp, set when the node becomes done
    archived_at    TEXT NULL                    -- timestamp; archived rows are hidden, never deleted
);

-- parent_id is self-referential with ON DELETE CASCADE, and the foreign_keys
-- pragma is on for every connection, so deleting a node really does take its
-- whole subtree with it rather than leaving orphans pointing at a missing row.
CREATE INDEX idx_nodes_parent_id ON nodes(parent_id);

-- The Kanban board reads by status, the calendar and "overdue" read by due, and
-- every view filters archived rows out, so all three get an index.
CREATE INDEX idx_nodes_status ON nodes(status);
CREATE INDEX idx_nodes_due ON nodes(due);
CREATE INDEX idx_nodes_archived_at ON nodes(archived_at);

CREATE TABLE tags (
    id    TEXT PRIMARY KEY,
    name  TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL DEFAULT ''   -- empty means "use the palette's default chip colour"
);

CREATE TABLE node_tags (
    node_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    tag_id  TEXT NOT NULL REFERENCES tags(id)  ON DELETE CASCADE,
    PRIMARY KEY (node_id, tag_id)
);

CREATE TABLE time_entries (
    id         TEXT PRIMARY KEY,
    node_id    TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    started_at TEXT NOT NULL,   -- timestamp
    ended_at   TEXT NULL        -- timestamp; NULL means the timer is still running
);

CREATE INDEX idx_time_entries_node_id ON time_entries(node_id);

-- At most one timer may be running at a time, anywhere in the tree.
--
-- The single-active invariant is a database constraint first and a service rule
-- second, because a race the service loses must not be able to corrupt the
-- data: two concurrent "start timer" calls that both pass an application-level
-- check would otherwise both commit, and the user would be billed twice for the
-- same minute with no way to tell which entry was real.
--
-- The trick is the expression: every row the partial index covers has
-- `ended_at IS NULL`, so the indexed expression is the constant 1 for all of
-- them, and UNIQUE therefore permits exactly one such row. Closed entries are
-- outside the WHERE clause and are unconstrained — a node may accumulate any
-- number of finished entries.
CREATE UNIQUE INDEX one_open_timer ON time_entries((ended_at IS NULL)) WHERE ended_at IS NULL;

CREATE TABLE attachments (
    id      TEXT PRIMARY KEY,
    node_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    path    TEXT NOT NULL,              -- path inside the app data directory; files are copied in, never linked
    mime    TEXT NOT NULL DEFAULT ''
);

-- One check per habit per day. The composite primary key is the rule: a day is
-- either done or not, and checking twice is the same fact, not two facts.
CREATE TABLE habit_checks (
    node_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    date    TEXT NOT NULL,              -- date, 'YYYY-MM-DD'
    PRIMARY KEY (node_id, date)
);
