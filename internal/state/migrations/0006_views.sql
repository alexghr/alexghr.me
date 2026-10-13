CREATE TABLE views (
  id             INTEGER PRIMARY KEY,
  request_id     TEXT NOT NULL UNIQUE,
  path           TEXT NOT NULL,
  user_agent     TEXT,
  referrer       TEXT,
  country        TEXT,
  status         INTEGER,
  created_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')) -- ISO 8601
);

CREATE INDEX views_created_at_idx       ON views(created_at);
CREATE INDEX views_path_created_at_idx  ON views(path, created_at);
