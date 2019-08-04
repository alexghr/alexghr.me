#!/usr/bin/env sh

set -o noclobber

RENDER_SCRIPT=$(dirname $0)/render.js

if [ -f "$1/index.js" ]; then
  node $RENDER_SCRIPT "$1" "$2"
fi
