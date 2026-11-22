#!/bin/bash
set -e
export WHODB_LOG_LEVEL="${WHODB_LOG_LEVEL:-error}"
HEADLESS="${1:-true}"
TARGET_DB="${2:-all}"
SPEC_FILE="${3:-}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DEFAULT_DATABASES="postgres mysql mysql8 mariadb tidb sqlite duckdb mongodb redis elasticsearch cockroachdb clickhouse memcached valkey dragonfly opensearch yugabytedb questdb ferretdb"
DEFAULT_CATEGORIES="postgres:sql mysql:sql mysql8:sql mariadb:sql tidb:sql sqlite:sql duckdb:sql mongodb:document redis:keyvalue elasticsearch:document cockroachdb:sql clickhouse:sql memcached:cache valkey:keyvalue dragonfly:keyvalue opensearch:document yugabytedb:sql questdb:sql ferretdb:document"
DATABASES_STR="${WHODB_DATABASES:-$DEFAULT_DATABASES}"
CATEGORIES_STR="${WHODB_DB_CATEGORIES:-$DEFAULT_CATEGORIES}"
VITE_EDITION="${WHODB_VITE_EDITION:-}"
SETUP_MODE="${WHODB_SETUP_MODE:-ce}"
EDITION_LABEL="${WHODB_EDITION_LABEL:-CE}"
EXTRA_WAIT="${WHODB_EXTRA_WAIT:-false}"
read -ra DATABASES <<< "$DATABASES_STR"
get_category() {
    local lookup_db="$1"
    for pair in $CATEGORIES_STR; do
        local db="${pair%%:*}"
        local cat="${pair#*:}"
        if [ "$db" = "$lookup_db" ]; then
            echo "$cat"
            return
        fi
    done
    echo "unknown"
}
echo "🚀 Running Playwright E2E tests ($EDITION_LABEL)"
echo "   Headless: $HEADLESS"
echo "   Target DB: $TARGET_DB"
echo "   Log Level: $WHODB_LOG_LEVEL"
if [ -n "$SPEC_FILE" ]; then
    echo "   Spec: $SPEC_FILE"
fi
if [ -n "$CDP_ENDPOINT" ]; then
    echo "   Mode: Gateway CDP ($CDP_ENDPOINT)"
fi
if [ "$TARGET_DB" != "all" ]; then
    FOUND=false
    for db in "${DATABASES[@]}"; do
        if [ "$db" = "$TARGET_DB" ]; then
            FOUND=true
            break
        fi
    done
    if [ "$FOUND" = "false" ]; then
        echo "❌ Unknown database: $TARGET_DB"
        echo "   Available: ${DATABASES[*]}"
        exit 1
    fi
    DATABASES=("$TARGET_DB")
fi
echo "   Databases: ${DATABASES[*]}"
echo "⚙️ Setting up test environment..."
export WHODB_SPEC_FILE="$SPEC_FILE"
bash "$SCRIPT_DIR/setup-e2e.sh" "$SETUP_MODE" "$TARGET_DB"
cd "$PROJECT_ROOT/frontend"
mkdir -p e2e/logs
find e2e/logs -name "*.log" ! -name "backend.log" -delete 2>/dev/null || true
rm -rf e2e/reports/* 2>/dev/null || sudo rm -rf e2e/reports/* 2>/dev/null || true
rm -rf e2e/.auth 2>/dev/null || true
mkdir -p e2e/reports/test-results e2e/reports/blobs e2e/reports/html
echo "🌐 Starting frontend dev server..."
BACKEND_PORT="${WHODB_BACKEND_PORT:-8080}"
if [ "$VITE_EDITION" = "ee" ]; then
    VITE_BACKEND_PORT="$BACKEND_PORT" NODE_ENV=test pnpm exec vite --config vite.ee.config.mts --port 3000 --clearScreen false --logLevel error > e2e/logs/frontend.log 2>&1 &
else
    VITE_BACKEND_PORT="$BACKEND_PORT" NODE_ENV=test pnpm exec vite --port 3000 --clearScreen false --logLevel error > e2e/logs/frontend.log 2>&1 &
fi
FRONTEND_PID=$!
echo "⏳ Waiting for frontend to be ready..."
COUNTER=0
while [ $COUNTER -lt 30 ]; do
    if nc -z localhost 3000 2>/dev/null; then
        echo "✅ Frontend is ready on port 3000"
        break
    fi
    sleep 0.5
    COUNTER=$((COUNTER + 1))
done
if ! nc -z localhost 3000 2>/dev/null; then
    echo "❌ Frontend failed to start"
    kill $FRONTEND_PID 2>/dev/null || true
    exit 1
fi
if [ "$EXTRA_WAIT" = "true" ]; then
    echo "⏳ Waiting for services..."
    MAX_WAIT=90 bash "$SCRIPT_DIR/wait-for-services.sh"
    echo "⏳ Giving services a moment to stabilize..."
    sleep 2
fi
FAILED_DBS=()
if [ -n "$SPEC_FILE" ]; then
    if [[ "$SPEC_FILE" == *.spec.* ]]; then
        SPEC_PATTERN="tests/features/$SPEC_FILE"
    else
        SPEC_PATTERN="tests/features/$SPEC_FILE.spec.mjs"
    fi
else
    SPEC_PATTERN=""
fi
PW_NO_DEPS="${WHODB_PW_NO_DEPS:-false}"
KEYBOARD_ONLY=false
if [ "$SPEC_FILE" = "keyboard-shortcuts" ] || [ "$SPEC_FILE" = "keyboard-shortcuts.spec.mjs" ]; then
    KEYBOARD_ONLY=true
fi
PW_PROJECT="standalone"
PW_PROJECT_MUTATING="standalone-mutating"
if [ -n "$CDP_ENDPOINT" ]; then
    PW_PROJECT="gateway"
    PW_PROJECT_MUTATING="gateway-mutating"
fi
PW_CONFIG="$PROJECT_ROOT/frontend/e2e/playwright.config.mjs"
if [ "$SPEC_FILE" = "accessibility" ] || [ "$SPEC_FILE" = "accessibility.spec.mjs" ]; then
    PW_CONFIG="$PROJECT_ROOT/frontend/e2e/playwright.a11y.config.mjs"
    PW_ARGS="--config=$PW_CONFIG --project=$PW_PROJECT"
else
    PW_ARGS="--config=$PW_CONFIG --project=$PW_PROJECT --project=$PW_PROJECT_MUTATING"
fi
if [ "$HEADLESS" = "false" ]; then
    PW_ARGS="$PW_ARGS --headed"
fi
if [ "$HEADLESS" = "true" ]; then
    DATABASE="${DATABASES[0]}" CATEGORY="$(get_category "${DATABASES[0]}")" \
        pnpm exec playwright test --config="$PW_CONFIG" --list > /dev/null 2>&1 || true
    echo "📋 Running ${#DATABASES[@]} database tests in parallel..."
    declare -A DB_PIDS
    for db in "${DATABASES[@]}"; do
        echo "🧪 Starting: $db ($(get_category "$db"))"
        (
            cd "$PROJECT_ROOT/frontend"
            LOG_FILE="$PROJECT_ROOT/frontend/e2e/logs/$db.log"
            if [ "$PW_NO_DEPS" = "true" ] && [ "$KEYBOARD_ONLY" = "true" ]; then
                DATABASE="$db" \
                CATEGORY="$(get_category "$db")" \
                pnpm exec playwright test \
                    --config="$PW_CONFIG" \
                    --project=setup \
                    > "$LOG_FILE" 2>&1
                DATABASE="$db" \
                CATEGORY="$(get_category "$db")" \
                pnpm exec playwright test \
                    --no-deps \
                    --config="$PW_CONFIG" \
                    --project="$PW_PROJECT_MUTATING" \
                    $SPEC_PATTERN \
                    >> "$LOG_FILE" 2>&1
            else
                DATABASE="$db" \
                CATEGORY="$(get_category "$db")" \
                pnpm exec playwright test \
                    $PW_ARGS \
                    $SPEC_PATTERN \
                    > "$LOG_FILE" 2>&1
            fi
        ) &
        DB_PIDS["$db"]=$!
    done
    echo ""
    echo "⏳ Waiting for all databases to complete..."
    declare -A DB_DONE
    DONE_COUNT=0
    TOTAL=${
    while [ $DONE_COUNT -lt $TOTAL ]; do
        for db in "${DATABASES[@]}"; do
            [ -n "${DB_DONE[$db]}" ] && continue
            if ! kill -0 "${DB_PIDS[$db]}" 2>/dev/null; then
                wait "${DB_PIDS[$db]}" && DB_DONE[$db]="pass" || DB_DONE[$db]="fail"
                DONE_COUNT=$((DONE_COUNT + 1))
                [ "${DB_DONE[$db]}" = "fail" ] && FAILED_DBS+=("$db")
                printf "\r\033[2K"
                if [ "${DB_DONE[$db]}" = "pass" ]; then
                    echo "✅ [$DONE_COUNT/$TOTAL] $db passed"
                else
                    echo "❌ [$DONE_COUNT/$TOTAL] $db failed (see e2e/logs/$db.log)"
                fi
            fi
        done
        if [ $DONE_COUNT -lt $TOTAL ]; then
            RUNNING=""
            for db in "${DATABASES[@]}"; do
                [ -n "${DB_DONE[$db]}" ] && continue
                LOG="$PROJECT_ROOT/frontend/e2e/logs/$db.log"
                SPEC=$(grep -o '› e2e/tests/features/[^:]*' "$LOG" 2>/dev/null | tail -1 | sed 's/.*features\///;s/\.spec\.mjs//')
                RUNNING="$RUNNING $db(${SPEC:-…})"
            done
            printf "\r\033[2K⏳ %d/%d done |%s" "$DONE_COUNT" "$TOTAL" "$RUNNING"
            sleep 2
        fi
    done
    echo ""
else
    echo "📋 Opening Playwright UI..."
    ENV_VARS=""
    if [ "$TARGET_DB" != "all" ]; then
        ENV_VARS="DATABASE=$TARGET_DB CATEGORY=$(get_category "$TARGET_DB")"
    fi
    (
        cd "$PROJECT_ROOT/frontend"
        env $ENV_VARS pnpm exec playwright test $PW_ARGS --ui
        exit $?
    ) && RESULT=0 || RESULT=$?
    if [ $RESULT -ne 0 ]; then
        FAILED_DBS+=("gui-session")
    fi
fi
if [ -d "$PROJECT_ROOT/frontend/e2e/reports/blobs" ] && ls "$PROJECT_ROOT/frontend/e2e/reports/blobs"/*.zip 1>/dev/null 2>&1; then
    echo ""
    echo "📊 Merging test reports..."
    (
        cd "$PROJECT_ROOT/frontend"
        PLAYWRIGHT_HTML_OPEN=never \
        PLAYWRIGHT_HTML_OUTPUT_DIR=e2e/reports/html \
        pnpm exec playwright merge-reports \
            e2e/reports/blobs \
            --reporter=html \
            2>/dev/null || true
    )
    echo "📁 HTML report: frontend/e2e/reports/html/index.html"
fi
echo ""
echo "🧹 Cleaning up..."
kill $FRONTEND_PID 2>/dev/null || true
bash "$SCRIPT_DIR/cleanup-e2e.sh" "$SETUP_MODE"
echo ""
echo "📊 Test Results:"
echo "   Total: ${#DATABASES[@]} databases"
echo "   Failed: ${#FAILED_DBS[@]} databases"
if [ ${
    echo ""
    echo "❌ Failed Tests:"
    for db in "${FAILED_DBS[@]}"; do
        echo ""
        echo "   [$db] ($(get_category "$db"))"
        LOG_FILE="$PROJECT_ROOT/frontend/e2e/logs/$db.log"
        if [ -f "$LOG_FILE" ]; then
            FAILURES=$(grep -E "(✘|FAILED|Error)" "$LOG_FILE" 2>/dev/null | tail -10)
            if [ -n "$FAILURES" ]; then
                echo "$FAILURES" | sed 's/^/     /'
            else
                echo "     (Could not parse test results)"
            fi
        else
            echo "     (Log file not found)"
        fi
        echo "     Log: $LOG_FILE"
    done
    exit 1
else
    echo "✅ All database tests passed!"
    echo "📁 Logs available in: frontend/e2e/logs/"
    echo "📁 HTML report: frontend/e2e/reports/html/"
    exit 0
fi