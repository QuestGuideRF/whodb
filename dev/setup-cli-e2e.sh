#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
echo "Setting up CLI E2E test environment..."
echo "Working from project root: $PROJECT_ROOT"
echo "Running cleanup first..."
bash "$SCRIPT_DIR/cleanup-cli-e2e.sh" 2>/dev/null || true
echo "Starting PostgreSQL..."
cd "$SCRIPT_DIR"
docker compose -f docker-compose.yml up -d e2e_postgres
echo "Waiting for PostgreSQL..."
COUNTER=0
MAX_WAIT=60
while [ $COUNTER -lt $MAX_WAIT ]; do
    if nc -z localhost 5432 2>/dev/null; then
        echo "PostgreSQL is ready"
        break
    fi
    sleep 1
    COUNTER=$((COUNTER + 1))
done
if [ $COUNTER -ge $MAX_WAIT ]; then
    echo "PostgreSQL failed to start within ${MAX_WAIT}s"
    exit 1
fi
echo "Building CLI..."
cd "$PROJECT_ROOT/cli"
go build -o whodb-cli .
echo "CLI built at cli/whodb-cli"
echo "CLI E2E setup complete!"