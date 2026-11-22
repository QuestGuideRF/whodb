#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
echo "🧹 Cleaning all coverage data..."
if [ -f "$PROJECT_ROOT/core/coverage.out" ]; then
    rm "$PROJECT_ROOT/core/coverage.out"
    echo "✅ Backend coverage cleaned"
fi
if [ -f "$PROJECT_ROOT/core/coverage.tmp.out" ]; then
    rm "$PROJECT_ROOT/core/coverage.tmp.out"
fi
if [ -f "$PROJECT_ROOT/core/coverage.prev.out" ]; then
    rm "$PROJECT_ROOT/core/coverage.prev.out"
fi
cd "$PROJECT_ROOT/frontend"
if [ -d ".nyc_output" ] || [ -d "coverage" ]; then
    npm run coverage:clean 2>/dev/null || rm -rf .nyc_output coverage
    echo "✅ Frontend coverage cleaned"
fi
echo "✨ All coverage data cleaned!"
echo ""
echo "Next steps:"
echo "  - Run 'pnpm e2e:ce:headless' to generate fresh coverage"
echo "  - Run 'npm run view:coverage' to see frontend coverage"
echo "  - Run 'npm run view:coverage:backend' to see backend coverage"