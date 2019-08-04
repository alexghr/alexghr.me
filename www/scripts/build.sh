#!/usr/bin/env sh

RENDER_SCRIPT=$(realpath $(dirname $0))/render.sh

find site/pages -mindepth 1 -type d -print0 | xargs -0 -r -n 1 -I FILE sh -c "$RENDER_SCRIPT FILE public"
cp -r site/assets public/
cp site/robots.txt public/
