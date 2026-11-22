#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
echo "Cleaning up CLI E2E environment..."
cd "$SCRIPT_DIR"
docker compose -f docker-compose.yml stop e2e_postgres 2>/dev/null || true
echo "CLI E2E cleanup complete!"