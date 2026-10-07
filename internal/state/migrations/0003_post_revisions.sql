CREATE TABLE post_revisions (
  id           INTEGER PRIMARY KEY,
  post_id      INTEGER NOT NULL REFERENCES posts(id),
  saved_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  slug         TEXT NOT NULL,
  title        TEXT NOT NULL,
  description  TEXT NOT NULL,
  body         TEXT NOT NULL,
  body_type    TEXT NOT NULL,
  created_at   TEXT NOT NULL,
  published_at TEXT,
  updated_at   TEXT
);
CREATE INDEX post_revisions_post_id ON post_revisions(post_id, id);

CREATE TABLE post_revision_tags (
  revision_id INTEGER NOT NULL REFERENCES post_revisions(id),
  tag_id      INTEGER NOT NULL REFERENCES tags(id),
  PRIMARY KEY (revision_id, tag_id)
);

-- blobs remain in assets
CREATE TABLE post_revision_assets (
  revision_id INTEGER NOT NULL REFERENCES post_revisions(id),
  asset_id    BLOB NOT NULL REFERENCES assets(id),
  local_path  TEXT NOT NULL,
  PRIMARY KEY (revision_id, local_path)
);
