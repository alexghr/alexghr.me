-- Restore the three legacy asset URLs from the HN/RSS article.
-- Apply migration 0005_redirects.sql before running this script.
BEGIN TRANSACTION;

INSERT INTO redirects (from_path, to_url, active, permanent)
VALUES
  (
    '/blog/effects-of-hn-personal-blogs-thread/rss-hits-per-day.svg',
    '/blog/effects-of-hn-personal-blogs-thread/assets/rss-hits-per-day.svg',
    1,
    1
  ),
  (
    '/blog/effects-of-hn-personal-blogs-thread/rss-hits-per-day.csv',
    '/blog/effects-of-hn-personal-blogs-thread/assets/rss-hits-per-day.csv',
    1,
    1
  ),
  (
    '/blog/effects-of-hn-personal-blogs-thread/rss-ua.txt',
    '/blog/effects-of-hn-personal-blogs-thread/assets/rss-ua.txt',
    1,
    1
  )
ON CONFLICT (from_path) DO UPDATE SET
  to_url = excluded.to_url,
  active = excluded.active,
  permanent = excluded.permanent;

COMMIT;
