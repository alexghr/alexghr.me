CREATE TABLE assets (
  id            BLOB NOT NULL PRIMARY KEY, -- sha256, CYA against PK bug https://www.sqlite.org/lang_createtable.html#the_primary_key
  name          TEXT NOT NULL,
  content_type  TEXT NOT NULL,
  data          BLOB NOT NULL
);
