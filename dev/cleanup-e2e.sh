#!/bin/bash
EDITION="${1:-ce}"
echo "🧹 Cleaning up $EDITION E2E environment..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
echo "📁 Working from project root: $PROJECT_ROOT"
echo "🧹 Cleaning up tmp directory..."
if [ -d "$PROJECT_ROOT/core/tmp" ]; then
    rm -f "$PROJECT_ROOT/core/tmp/e2e_test.db"
    rm -f "$PROJECT_ROOT/core/tmp/e2e_test.duckdb"
    rm -f "$PROJECT_ROOT/core/tmp/test-server.pid"
    find "$PROJECT_ROOT/core/tmp" -type f ! -name '.test-binary-hash' -delete 2>/dev/null || true
    echo "✅ tmp directory cleaned (preserved hash cache)"
else
    echo "ℹ️ No tmp directory to clean up"
fi
if [ -f "$PROJECT_ROOT/core/server.test" ]; then
    echo "ℹ️  Test binary preserved for caching: server.test"
    echo "   Size: $(du -h "$PROJECT_ROOT/core/server.test" | cut -f1)"
    echo "   To force rebuild, delete: rm $PROJECT_ROOT/core/server.test"
fi
if [ -f "$PROJECT_ROOT/core/coverage.out" ]; then
    echo "✅ Backend coverage saved to core/coverage.out"
fi
echo "🧹 Cleaning frontend coverage artifacts..."
if [ -d "$PROJECT_ROOT/frontend/.nyc_output" ]; then
    if [ -f "$PROJECT_ROOT/frontend/.nyc_output/out.json" ]; then
        echo "✅ Frontend coverage data preserved in .nyc_output/out.json"
    fi
fi
if [ "$EDITION" = "ee" ]; then
    EE_CLEANUP_SCRIPT="$PROJECT_ROOT/ee/dev/cleanup-ee-databases.sh"
    if [ -f "$EE_CLEANUP_SCRIPT" ]; then
        echo "🧹 Running EE-specific cleanup..."
        bash "$EE_CLEANUP_SCRIPT"
    fi
fi
echo "🐳 Stopping CE database services..."
cd "$SCRIPT_DIR"
docker compose -f docker-compose.yml --profile ssl rm -f -s -v 2>/dev/null || true
docker compose -f docker-compose.yml --profile ssl down --volumes --remove-orphans --timeout 0
echo "🔄 Pruning any dangling volumes..."
docker volume prune -f
echo "🛑 Stopping test server..."
if [ -f "$PROJECT_ROOT/core/tmp/test-server.pid" ]; then
    TEST_SERVER_PID=$(cat "$PROJECT_ROOT/core/tmp/test-server.pid")
    if ps -p $TEST_SERVER_PID > /dev/null 2>&1; then
        kill $TEST_SERVER_PID
        sleep 2
        echo "✅ Test server stopped (PID: $TEST_SERVER_PID)"
    fi
    rm -f "$PROJECT_ROOT/core/tmp/test-server.pid"
elif [ -n "$TEST_SERVER_PID" ] && ps -p $TEST_SERVER_PID > /dev/null 2>&1; then
    kill $TEST_SERVER_PID
    echo "✅ Test server stopped (PID: $TEST_SERVER_PID)"
else
    PIDS=$(pgrep -f "server.test" 2>/dev/null || true)
    if [ -n "$PIDS" ]; then
        echo "🔄 Found running server.test processes, stopping them..."
        echo $PIDS | xargs kill
        echo "✅ All server.test processes stopped"
    else
        echo "ℹ️ No test server processes found"
    fi
fi
echo "🔍 Ensuring port 3000 is free..."
lsof -ti:3000 | xargs kill -9 2>/dev/null || true
echo "🔍 Ensuring port 8080 is free..."
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
echo "✅ $EDITION E2E environment cleanup complete!"