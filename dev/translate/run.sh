#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
LOCALE_ARGS=""
if [ $
    LOCALE_ARGS=$(IFS=,; echo "$*")
fi
echo "=== Step 1: Detecting translation drift ==="
echo ""
if [ -n "$LOCALE_ARGS" ]; then
    uv run python detect.py -l "$LOCALE_ARGS"
else
    uv run python detect.py
fi
echo ""
echo "=== Step 2: Translating and applying ==="
echo ""
if [ $
    node translate.mjs "$@"
else
    node translate.mjs
fi