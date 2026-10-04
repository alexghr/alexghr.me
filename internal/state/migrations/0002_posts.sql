CREATE TABLE posts (
  id            INTEGER PRIMARY KEY,
  slug          TEXT NOT NULL UNIQUE,
  title         TEXT NOT NULL,
  description   TEXT NOT NULL,
  body          TEXT NOT NULL,
  body_type     TEXT NOT NULL, -- e.g. "html-template"
  created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')), -- ISO 8601
  published_at  TEXT, -- ISO 8601
  updated_at    TEXT  -- ISO 8601
);

CREATE TABLE tags (
  id    INTEGER PRIMARY KEY,
  name  TEXT NOT NULL UNIQUE
);

CREATE TABLE post_tags (
  post_id  INTEGER NOT NULL REFERENCES posts(id),
  tag_id   INTEGER NOT NULL REFERENCES tags(id),

  PRIMARY KEY (post_id, tag_id)
);

CREATE TABLE post_assets (
  post_id     INTEGER NOT NULL REFERENCES posts(id),
  asset_id    BLOB    NOT NULL REFERENCES assets(id),
  local_path  TEXT    NOT NULL,

  PRIMARY KEY (post_id, local_path)
);
