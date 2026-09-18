-- Application settings: a key/value table read at startup for the palette,
-- theme, accent override and language (D6).
--
-- The defaults are not seeded here. Seeding is idempotent and must never
-- overwrite a value the user has already chosen, which is a rule about the
-- running application rather than about the shape of the schema, so it lives in
-- SettingsRepo.SeedDefaults where it can be tested as such.
CREATE TABLE settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
