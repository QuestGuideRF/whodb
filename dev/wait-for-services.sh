#!/bin/bash
set -e
BACKEND_URL="${BACKEND_URL:-http://localhost:8080}"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:3000}"
MAX_WAIT="${MAX_WAIT:-60}"
echo "⏳ Waiting for services to be ready..."
echo "   Backend:  $BACKEND_URL"
echo "   Frontend: $FRONTEND_URL"
echo "   Timeout:  ${MAX_WAIT}s"
check_port() {
    local host=$1
    local port=$2
    nc -z "$host" "$port" 2>/dev/null
}
BACKEND_HOST=$(echo "$BACKEND_URL" | sed -e 's|^[^/]*//||' -e 's|:.*||')
BACKEND_PORT=$(echo "$BACKEND_URL" | sed -e 's|.*:||' -e 's|/.*||')
BACKEND_HOST=${BACKEND_HOST:-localhost}
BACKEND_PORT=${BACKEND_PORT:-8080}
FRONTEND_HOST=$(echo "$FRONTEND_URL" | sed -e 's|^[^/]*//||' -e 's|:.*||')
FRONTEND_PORT=$(echo "$FRONTEND_URL" | sed -e 's|.*:||' -e 's|/.*||')
FRONTEND_HOST=${FRONTEND_HOST:-localhost}
FRONTEND_PORT=${FRONTEND_PORT:-3000}
wait_for_port_async() {
    local name=$1
    local host=$2
    local port=$3
    local max_wait=$4
    local counter=0
    echo "⏳ Waiting for $name ($host:$port)..."
    while [ $counter -lt $max_wait ]; do
        if check_port "$host" "$port"; then
            echo "✅ $name is ready!"
            return 0
        fi
        sleep 1
        counter=$((counter + 1))
    done
    echo "❌ $name timeout after ${max_wait}s"
    return 1
}
wait_for_port_async "Backend" "$BACKEND_HOST" "$BACKEND_PORT" "$MAX_WAIT" &
PID_BACKEND=$!
wait_for_port_async "Frontend" "$FRONTEND_HOST" "$FRONTEND_PORT" "$MAX_WAIT" &
PID_FRONTEND=$!
FAILED=false
if ! wait $PID_BACKEND; then
    echo "Backend failed to start within ${MAX_WAIT} seconds"
    FAILED=true
fi
if ! wait $PID_FRONTEND; then
    echo "Frontend failed to start within ${MAX_WAIT} seconds"
    FAILED=true
fi
if [ "$FAILED" = "true" ]; then
    exit 1
fi
echo "🎉 All services are ready!"