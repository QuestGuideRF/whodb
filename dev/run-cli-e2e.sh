#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
echo "Running CLI E2E tests..."
bash "$SCRIPT_DIR/setup-cli-e2e.sh"
echo "Running CLI E2E tests..."
cd "$PROJECT_ROOT/cli"
go test -tags=e2e_postgres -v ./e2e/...
RESULT=$?
bash "$SCRIPT_DIR/cleanup-cli-e2e.sh"
exit $RESULT