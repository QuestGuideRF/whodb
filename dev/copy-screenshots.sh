#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SCREENSHOTS_SOURCE="$PROJECT_ROOT/frontend/e2e/screenshots/postgres"
DOCS_IMAGES_DEST="$PROJECT_ROOT/docs/images"
echo "📸 WhoDB Screenshot Copy Utility"
echo "=================================="
echo ""
if [ ! -d "$SCREENSHOTS_SOURCE" ]; then
    echo "❌ Error: Screenshots directory not found at:"
    echo "   $SCREENSHOTS_SOURCE"
    echo ""
    echo "💡 Tip: Run 'pnpm screenshot' first to generate screenshots"
    exit 1
fi
SCREENSHOT_COUNT=$(find "$SCREENSHOTS_SOURCE" -name "*.png" 2>/dev/null | wc -l | tr -d ' ')
if [ "$SCREENSHOT_COUNT" -eq 0 ]; then
    echo "❌ Error: No screenshots found in:"
    echo "   $SCREENSHOTS_SOURCE"
    echo ""
    echo "💡 Tip: Run 'pnpm screenshot' first to generate screenshots"
    exit 1
fi
echo "📁 Source: $SCREENSHOTS_SOURCE"
echo "📁 Destination: $DOCS_IMAGES_DEST"
echo "📊 Found $SCREENSHOT_COUNT screenshot(s)"
echo ""
if [ ! -d "$DOCS_IMAGES_DEST" ]; then
    echo "📂 Creating destination directory..."
    mkdir -p "$DOCS_IMAGES_DEST"
fi
echo "🚀 Copying screenshots..."
cp -v "$SCREENSHOTS_SOURCE"/*.png "$DOCS_IMAGES_DEST/" 2>&1 | while read line; do
    filename=$(basename "$line" | sed "s/'//g" | awk '{print $NF}')
    echo "   ✓ $filename"
done
echo ""
echo "=================================="
echo "✅ Successfully copied $SCREENSHOT_COUNT screenshot(s) to docs/images/"
echo ""
echo "💡 Screenshots are now available at:"
echo "   $DOCS_IMAGES_DEST"
echo "=================================="