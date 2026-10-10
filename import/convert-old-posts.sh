#!/usr/bin/env bash
set -euo pipefail

fail() { printf '%s\n' "$*" >&2; exit 1; }

[[ $# -eq 0 ]] || fail 'This script takes no arguments.'

# Use the same nixpkgs revision as devenv.lock's nixpkgs-unstable input.
if [[ ${IMPORT_OLD_POSTS_IN_NIX:-} != 1 ]]; then
  nixpkgs=github:NixOS/nixpkgs/73e728ddb6b7a12d18808f510813a13ee1fe4cce
  export IMPORT_OLD_POSTS_IN_NIX=1
  exec nix shell "$nixpkgs#bash" "$nixpkgs#git" "$nixpkgs#pandoc" \
    "$nixpkgs#jq" "$nixpkgs#coreutils" "$nixpkgs#gnused" "$nixpkgs#gnutar" \
    --command bash "$0"
fi

# Generate old-posts/ in the caller's working directory.
output=$(realpath -m -- old-posts)
[[ ! -e $output ]] || fail "Output already exists: $output"
repo=$(git -C "$(dirname -- "$(realpath -- "$0")")" rev-parse --show-toplevel)
commit=$(git -C "$repo" rev-parse --verify 'ag/old-site^{commit}')
work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT
mkdir -p "$work/source" "$output"
git -C "$repo" archive "$commit" www/src/content/blog www/src/assets www/public/blog \
  | tar -x -C "$work/source"

# Pandoc supplies its own Lua interpreter. The filter handles parsed Markdown
# links/images, so URLs inside code examples are left alone.
cat > "$work/import.lua" <<'LUA'
local source = assert(os.getenv('IMPORT_SOURCE'))
local output = assert(os.getenv('IMPORT_OUTPUT'))
local public = assert(os.getenv('IMPORT_PUBLIC'))
local slug = assert(os.getenv('IMPORT_SLUG'))
local copied, names = {}, {}

local function asset_url(url, is_image)
  if url:match('^%a[%w+.-]*:') or url:match('^//') or url:match('^#') then
    return url
  end
  local path, suffix = url:match('^([^?#]+)(.*)$')
  if not path then return url end
  local candidate
  if path:sub(1, 1) == '/' then
    candidate = public .. path
  else
    candidate = pandoc.path.directory(source) .. '/' .. path
  end
  local input = io.open(candidate, 'rb')
  if not input then
    if is_image or path:match('^%.%./%.%./assets/') then
      error('Missing local asset: ' .. url)
    end
    return url -- Post links and routes such as /rss.xml.
  end
  if copied[candidate] then
    input:close()
    return copied[candidate] .. suffix
  end
  local name = pandoc.path.filename(path)
  if names[name] then error('Asset filename collision: ' .. name) end
  names[name] = true
  local destination = assert(io.open(output .. '/assets/' .. name, 'wb'))
  destination:write(input:read('*a'))
  destination:close()
  input:close()
  copied[candidate] = 'assets/' .. name
  return copied[candidate] .. suffix
end

function Image(image)
  image.src = asset_url(image.src, true)
  return image
end

function Link(link)
  link.target = asset_url(link.target, false)
  return link
end

local function text(meta, key)
  return meta[key] and pandoc.utils.stringify(meta[key]) or ''
end

local function date(value)
  if value == '' then return nil end
  if value:match('^%d%d%d%d%-%d%d%-%d%d$') then
    return value .. 'T00:00:00Z'
  end
  return value
end

function Pandoc(doc)
  local meta = {
    slug = slug,
    title = text(doc.meta, 'title'),
    description = text(doc.meta, 'description'),
    publishedAt = date(text(doc.meta, 'pubDate')),
    updatedAt = date(text(doc.meta, 'updatedDate')),
    status = 'published',
  }
  assert(meta.title ~= '', 'Missing title: ' .. source)
  assert(meta.publishedAt, 'Missing pubDate: ' .. source)
  local file = assert(io.open(output .. '/meta.json', 'w'))
  file:write(pandoc.json.encode(meta), '\n')
  file:close()
  -- No current posts have a heroImage, but preserve one if present.
  local hero = text(doc.meta, 'heroImage')
  if hero ~= '' then
    doc.blocks:insert(1, pandoc.Para({pandoc.Image({}, asset_url(hero, true))}))
  end
  doc.meta = {}
  return doc
end
LUA

count=0
for source in "$work/source/www/src/content/blog/"*.md; do
  [[ -f $source ]] || fail 'No Markdown posts found on ag/old-site'
  name=$(basename -- "$source" .md)
  # Astro 2's github-slugger removes ':' from til-:has-css-pseudo-class.
  # Every other filename on this branch is already its published slug.
  slug=${name//:/}
  post_dir=$output/$slug
  [[ ! -e $post_dir ]] || fail "Duplicate slug: $slug"
  mkdir -p "$post_dir/assets"
  IMPORT_SOURCE=$source IMPORT_OUTPUT=$post_dir IMPORT_SLUG=$slug \
    IMPORT_PUBLIC=$work/source/www/public \
    pandoc "$source" \
      --from=markdown-smart-tex_math_dollars-raw_tex-implicit_figures \
      --to=html5 --wrap=none --no-highlight \
      --lua-filter="$work/import.lua" \
    | sed 's/{{/\&#123;\&#123;/g' > "$post_dir/post.html"
  # mkpost expects RFC3339 timestamps.
  jq -e '.publishedAt | fromdateiso8601' "$post_dir/meta.json" >/dev/null
  jq -e 'if .updatedAt then .updatedAt | fromdateiso8601 else true end' \
    "$post_dir/meta.json" >/dev/null
  count=$((count + 1))
  printf 'Converted %s\n' "$slug"
done
printf 'Converted %d posts from %s into %s\n' "$count" "$commit" "$output"
