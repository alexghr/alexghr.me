ALTER TABLE posts ADD COLUMN status TEXT NOT NULL DEFAULT 'draft';

UPDATE posts SET status = 'published'
WHERE published_at IS NOT NULL;

CREATE INDEX post_status ON posts(status);
