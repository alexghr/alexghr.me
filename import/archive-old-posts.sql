-- Archive the five recommended posts that are currently published.
-- Keep draft posts private and preserve content, assets, and dates.
BEGIN TRANSACTION;

UPDATE posts
SET status = 'archived'
WHERE status = 'published'
  AND slug IN (
    'hetzner-nixos-server',
    'rubber-ducking-a-blog',
    'rubber-ducking-a-blog-2',
    'rebuild',
    'raspberry-pi-sd-card-failure'
  );

COMMIT;
