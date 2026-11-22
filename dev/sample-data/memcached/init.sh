#!/bin/sh
MEMCACHED_HOST="${MEMCACHED_HOST:-e2e_memcached}"
MEMCACHED_PORT="${MEMCACHED_PORT:-11211}"
MEMCACHED_TLS="${MEMCACHED_TLS:-false}"
mc_send() {
    if [ "$MEMCACHED_TLS" = "true" ]; then
        timeout 2 openssl s_client -connect "$MEMCACHED_HOST:$MEMCACHED_PORT" -quiet -ign_eof 2>/dev/null
    else
        nc -w 1 "$MEMCACHED_HOST" "$MEMCACHED_PORT"
    fi
}
echo "Waiting for Memcached to be ready at ${MEMCACHED_HOST}:${MEMCACHED_PORT} (TLS=${MEMCACHED_TLS})..."
ready=false
for i in $(seq 1 30); do
    if printf "version\r\n" | mc_send 2>/dev/null | grep -q "VERSION"; then
        echo "Memcached is ready!"
        ready=true
        break
    fi
    echo "Attempt $i: Memcached not ready, waiting..."
    sleep 1
done
if [ "$ready" != "true" ]; then
    echo "ERROR: Memcached did not become ready"
    exit 1
fi
mc_set() {
    key="$1"
    flags="$2"
    exptime="$3"
    value="$4"
    bytes=$(printf '%s' "$value" | wc -c | tr -d ' ')
    printf "set %s %s %s %s\r\n%s\r\n" "$key" "$flags" "$exptime" "$bytes" "$value" | mc_send
}
echo "Seeding Memcached with test data..."
mc_set "user:1" 0 0 '{"id":1,"username":"johndoe","email":"john@example.com","created_at":"2023-01-15T10:30:00Z"}'
mc_set "user:2" 0 0 '{"id":2,"username":"janedoe","email":"jane@example.com","created_at":"2023-02-20T14:45:00Z"}'
mc_set "user:3" 0 0 '{"id":3,"username":"bobsmith","email":"bob@example.com","created_at":"2023-03-10T09:15:00Z"}'
mc_set "user:4" 0 0 '{"id":4,"username":"alicejones","email":"alice@example.com","created_at":"2023-04-05T16:20:00Z"}'
mc_set "user:5" 0 0 '{"id":5,"username":"charlie","email":"charlie@example.com","created_at":"2023-05-12T11:00:00Z"}'
mc_set "product:1" 1 0 '{"id":1,"name":"Laptop","price":999.99,"category":"electronics"}'
mc_set "product:2" 1 0 '{"id":2,"name":"Headphones","price":79.99,"category":"electronics"}'
mc_set "product:3" 1 0 '{"id":3,"name":"Keyboard","price":129.99,"category":"accessories"}'
mc_set "product:4" 1 0 '{"id":4,"name":"Mouse","price":49.99,"category":"accessories"}'
mc_set "product:5" 1 0 '{"id":5,"name":"Monitor","price":399.99,"category":"electronics"}'
mc_set "inventory:product:1" 2 0 "50"
mc_set "inventory:product:2" 2 0 "150"
mc_set "inventory:product:3" 2 0 "75"
mc_set "inventory:product:4" 2 0 "200"
mc_set "inventory:product:5" 2 0 "30"
mc_set "session:abc123" 0 0 '{"user_id":1,"login_time":"2023-06-01T08:00:00Z","ip":"192.168.1.1"}'
mc_set "session:def456" 0 0 '{"user_id":2,"login_time":"2023-06-01T09:30:00Z","ip":"192.168.1.2"}'
mc_set "session:ghi789" 0 0 '{"user_id":3,"login_time":"2023-06-01T10:15:00Z","ip":"192.168.1.3"}'
mc_set "order:1" 1 0 '{"id":1,"user_id":1,"total":1079.98,"status":"delivered","created_at":"2023-06-15"}'
mc_set "order:2" 1 0 '{"id":2,"user_id":2,"total":79.99,"status":"shipped","created_at":"2023-06-16"}'
mc_set "order:3" 1 0 '{"id":3,"user_id":3,"total":179.98,"status":"processing","created_at":"2023-06-17"}'
mc_set "cache:homepage" 3 0 '{"featured_products":[1,3,5],"banner":"Summer Sale"}'
mc_set "cache:categories" 3 0 '["electronics","accessories","computers"]'
mc_set "counter:page_views" 2 0 "42857"
mc_set "counter:api_calls" 2 0 "128934"
echo "Memcached seeding complete! Verifying..."
printf "stats\r\nquit\r\n" | mc_send | grep curr_items
echo "Done."