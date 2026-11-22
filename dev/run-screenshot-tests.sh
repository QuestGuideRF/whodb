#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SCREENSHOT_DIR="$PROJECT_ROOT/frontend/e2e/screenshots/postgres"
LOG_DIR="$PROJECT_ROOT/frontend/e2e/logs"
BACKEND_LOG="$LOG_DIR/screenshot-backend.log"
FRONTEND_LOG="$LOG_DIR/screenshot-frontend.log"
echo "📸 WhoDB Screenshot Generation Environment"
echo "=========================================="
echo ""
cleanup() {
    echo ""
    echo "🧹 Cleaning up screenshot environment..."
    if [ -n "${TEST_SERVER_PID:-}" ] && ps -p "$TEST_SERVER_PID" > /dev/null 2>&1; then
        echo "   Stopping test server (PID: $TEST_SERVER_PID)..."
        kill "$TEST_SERVER_PID" || true
        sleep 2
        if ps -p "$TEST_SERVER_PID" > /dev/null 2>&1; then
            kill -9 "$TEST_SERVER_PID" || true
        fi
    elif [ -f "$PROJECT_ROOT/core/tmp/screenshot-server.pid" ]; then
        TEST_SERVER_PID=$(cat "$PROJECT_ROOT/core/tmp/screenshot-server.pid")
        if ps -p "$TEST_SERVER_PID" > /dev/null 2>&1; then
            echo "   Stopping test server (PID: $TEST_SERVER_PID)..."
            kill "$TEST_SERVER_PID" || true
            sleep 2
            if ps -p "$TEST_SERVER_PID" > /dev/null 2>&1; then
                kill -9 "$TEST_SERVER_PID" || true
            fi
        fi
    fi
    rm -f "$PROJECT_ROOT/core/tmp/screenshot-server.pid"
    if [ -n "${FRONTEND_PID:-}" ] && ps -p "$FRONTEND_PID" > /dev/null 2>&1; then
        echo "   Stopping frontend dev server (PID: $FRONTEND_PID)..."
        kill "$FRONTEND_PID" || true
        sleep 2
        if ps -p "$FRONTEND_PID" > /dev/null 2>&1; then
            kill -9 "$FRONTEND_PID" || true
        fi
    fi
    echo "   Stopping Docker services..."
    cd "$SCRIPT_DIR"
    docker compose -f docker-compose.screenshot.yaml down --volumes --remove-orphans || true
    echo "✅ Cleanup complete"
}
trap cleanup EXIT
echo "🔨 Building test binary..."
cd "$PROJECT_ROOT/core"
export GOCACHE="${GOCACHE:-$PROJECT_ROOT/core/tmp/go-cache}"
export GOTMPDIR="${GOTMPDIR:-$PROJECT_ROOT/core/tmp/go-tmp}"
mkdir -p "$GOCACHE" "$GOTMPDIR"
BINARY_PATH="$PROJECT_ROOT/core/server.test"
echo "   Building test binary..."
go test -coverpkg=./... -c -o server.test ./cmd/whodb
echo "   ✅ Test binary built"
echo ""
echo "🗄️  Setting up SQLite database..."
SQLITE_DB="$PROJECT_ROOT/core/tmp/e2e_test.db"
mkdir -p "$PROJECT_ROOT/core/tmp"
if [ -f "$SQLITE_DB" ]; then
    if sqlite3 "$SQLITE_DB" "SELECT name FROM sqlite_master WHERE type='table' AND name='users';" 2>/dev/null | grep -q users; then
        echo "   SQLite database already initialized"
    else
        echo "   Reinitializing SQLite database..."
        rm -f "$SQLITE_DB"
        sqlite3 "$SQLITE_DB" < "$SCRIPT_DIR/sample-data/sqlite3/data.sql"
        chmod 644 "$SQLITE_DB"
    fi
else
    echo "   Creating SQLite database..."
    sqlite3 "$SQLITE_DB" < "$SCRIPT_DIR/sample-data/sqlite3/data.sql"
    chmod 644 "$SQLITE_DB"
fi
echo "   ✅ SQLite database ready"
echo ""
echo "🐳 Starting Docker services for screenshots..."
cd "$SCRIPT_DIR"
docker compose -f docker-compose.screenshot.yaml down --volumes --remove-orphans || true
docker compose -f docker-compose.screenshot.yaml up -d --remove-orphans
echo "   Waiting for PostgreSQL to be ready..."
COUNTER=0
MAX_WAIT=60
while [ $COUNTER -lt $MAX_WAIT ]; do
    POSTGRES_STATUS=$(docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' screenshot_postgres 2>/dev/null || true)
    if [ "$POSTGRES_STATUS" = "healthy" ]; then
        echo "   ✅ PostgreSQL is ready"
        break
    fi
    if [ "$POSTGRES_STATUS" = "unhealthy" ]; then
        echo "   ❌ PostgreSQL healthcheck failed"
        docker logs screenshot_postgres || true
        exit 1
    fi
    sleep 1
    COUNTER=$((COUNTER + 1))
done
if [ $COUNTER -ge $MAX_WAIT ]; then
    echo "   ❌ PostgreSQL failed to start within ${MAX_WAIT} seconds"
    exit 1
fi
echo ""
echo "🚀 Starting backend test server..."
cd "$PROJECT_ROOT/core"
mkdir -p "$LOG_DIR" "$SCREENSHOT_DIR"
ENVIRONMENT=dev WHODB_DISABLE_MOCK_DATA_GENERATION='' \
    ./server.test -test.run=^TestMain$ \
    > "$BACKEND_LOG" 2>&1 &
TEST_SERVER_PID=$!
echo "$TEST_SERVER_PID" > "$PROJECT_ROOT/core/tmp/screenshot-server.pid"
echo "   Waiting for backend server..."
COUNTER=0
MAX_WAIT=30
while [ $COUNTER -lt $MAX_WAIT ]; do
    if ! ps -p "$TEST_SERVER_PID" > /dev/null 2>&1; then
        echo "   ❌ Backend server exited before becoming ready"
        echo "   Backend log: $BACKEND_LOG"
        tail -n 80 "$BACKEND_LOG" || true
        exit 1
    fi
    if curl -sf http://localhost:8080/health >/dev/null 2>&1; then
        echo "   ✅ Backend server is ready (PID: $TEST_SERVER_PID)"
        break
    fi
    sleep 1
    COUNTER=$((COUNTER + 1))
done
if [ $COUNTER -ge $MAX_WAIT ]; then
    echo "   ❌ Backend server failed to start within ${MAX_WAIT} seconds"
    exit 1
fi
echo ""
echo "🎨 Starting frontend dev server..."
cd "$PROJECT_ROOT/frontend"
VITE_E2E_TEST=true NODE_ENV=test pnpm exec vite --host 127.0.0.1 --port 3000 --strictPort > "$FRONTEND_LOG" 2>&1 &
FRONTEND_PID=$!
echo "   Waiting for frontend server..."
COUNTER=0
MAX_WAIT=60
while [ $COUNTER -lt $MAX_WAIT ]; do
    if ! ps -p "$FRONTEND_PID" > /dev/null 2>&1; then
        echo "   ❌ Frontend server exited before becoming ready"
        echo "   Frontend log: $FRONTEND_LOG"
        tail -n 80 "$FRONTEND_LOG" || true
        exit 1
    fi
    if curl -sf http://localhost:3000 >/dev/null 2>&1; then
        echo "   ✅ Frontend server is ready (PID: $FRONTEND_PID)"
        break
    fi
    sleep 1
    COUNTER=$((COUNTER + 1))
done
if [ $COUNTER -ge $MAX_WAIT ]; then
    echo "   ❌ Frontend server failed to start within ${MAX_WAIT} seconds"
    exit 1
fi
echo ""
echo "📸 Running screenshot tests..."
echo ""
cd "$PROJECT_ROOT/frontend"
echo "   Test spec: tests/postgres-screenshots.spec.mjs"
cd "$PROJECT_ROOT/frontend/e2e"
set +e
WHODB_SCREENSHOT_DIR="$SCREENSHOT_DIR" NODE_ENV=test pnpm exec playwright test \
    --config=playwright.screenshot.config.mjs \
    --project=screenshots \
    tests/postgres-screenshots.spec.mjs
PW_EXIT_CODE=$?
set -e
echo ""
echo "=========================================="
if [ $PW_EXIT_CODE -eq 0 ]; then
    echo "✅ Screenshot generation completed successfully"
    echo ""
    echo "📁 Screenshots saved to:"
    echo "   $PROJECT_ROOT/frontend/e2e/screenshots/postgres/"
    echo ""
    echo "💡 Tip: You can find all screenshots organized by test number and name"
else
    echo "❌ Screenshot generation failed with exit code: $PW_EXIT_CODE"
    echo ""
    echo "Check the Playwright output above for details"
fi
echo "=========================================="
echo ""
exit $PW_EXIT_CODE