CREATE TABLE redirects (
  id          INTEGER PRIMARY KEY,
  active      BOOL NOT NULL DEFAULT true,
  from_path   TEXT NOT NULL UNIQUE,
  to_url      TEXT NOT NULL,
  permanent   BOOL NOT NULL DEFAULT FALSE,
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),

  CHECK (from_path != to_url)
);
