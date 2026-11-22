#!/bin/sh
set -e
COCKROACH_HOST="${COCKROACH_HOST:-e2e_cockroachdb}"
COCKROACH_PORT="${COCKROACH_PORT:-26257}"
MAX_RETRIES=30
RETRY_INTERVAL=2
if [ -n "$COCKROACH_CERTS_DIR" ]; then
    CONN_FLAGS="--certs-dir=$COCKROACH_CERTS_DIR"
    echo "Using secure mode with certs from $COCKROACH_CERTS_DIR"
else
    CONN_FLAGS="--insecure"
    echo "Using insecure mode"
fi
echo "Waiting for CockroachDB at ${COCKROACH_HOST}:${COCKROACH_PORT}..."
retries=0
while [ $retries -lt $MAX_RETRIES ]; do
    if cockroach sql $CONN_FLAGS --host="${COCKROACH_HOST}:${COCKROACH_PORT}" -e "SELECT 1" > /dev/null 2>&1; then
        echo "CockroachDB is ready!"
        break
    fi
    retries=$((retries + 1))
    echo "Attempt $retries/$MAX_RETRIES - CockroachDB not ready yet, retrying in ${RETRY_INTERVAL}s..."
    sleep $RETRY_INTERVAL
done
if [ $retries -ge $MAX_RETRIES ]; then
    echo "ERROR: CockroachDB did not become ready within $((MAX_RETRIES * RETRY_INTERVAL))s"
    exit 1
fi
echo "Running init SQL..."
cockroach sql $CONN_FLAGS --host="${COCKROACH_HOST}:${COCKROACH_PORT}" < /data.sql
echo "CockroachDB initialization complete!"