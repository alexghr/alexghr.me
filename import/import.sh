#!/usr/bin/env bash
set -euo pipefail

repo=$(git -C "$(dirname -- "$(realpath -- "$0")")" rev-parse --show-toplevel)
svc="alexghr-me.service"
remote_path="/var/lib/alexghr.me/db.sqlite3"
local_dir="$(mktemp -d)"
cd "$local_dir"

function cleanup() {
  ssh trip systemctl start "$svc"
  rm -rf "$local_dir"
}

trap cleanup EXIT SIGINT SIGTERM

ssh trip "sqlite3 -readonly '$remote_path' \".backup '$remote_path.bak'\""
scp "trip:$remote_path.bak" "db.sqlite3"

"$repo/import/convert-old-posts.sh"
for post in ./old-posts/*; do 
  "$repo/mkpost" -db "db.sqlite3" -dir "$post"
done

sqlite3 -bail "db.sqlite3" < "$repo/import/archive-old-posts.sql"
sqlite3 -bail "db.sqlite3" < "$repo/import/setup-redirects.sql"

scp "db.sqlite3" "trip:$remote_path.new"
ssh trip chown --reference="$remote_path" "$remote_path.new"
ssh trip chmod --reference="$remote_path" "$remote_path.new"

ssh trip systemctl stop "$svc"

ssh trip mv "$remote_path.new" "$remote_path"
