#!/bin/bash

set -e

TEST_DIR="/tmp/test-siem-logs"
LOG_DIR="/tmp/siem-agent-logs"
BUFFER_DIR="/tmp/siem-agent-buffer"
SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG_FILE="$SCRIPT_DIR/configs/config.yaml"
BINARY="$SCRIPT_DIR/bin/siem-agent"

mkdir -p "$TEST_DIR" "$LOG_DIR" "$BUFFER_DIR"

if [ ! -f "$BINARY" ]; then
    echo "Building agent..."
    cd "$SCRIPT_DIR"
    go build -o bin/siem-agent ./cmd/agent/main.go
fi

echo "Checking NoSQLdb connection..."
if ! socat -u /dev/null TCP:127.0.0.1:5140,connect-timeout=2 2>/dev/null; then
    echo "ERROR: NoSQLdb not running on port 5140"
    exit 1
fi
echo "OK: NoSQLdb is running"

> "$TEST_DIR/auth.log"

sed -i 's|/tmp/test-siem-logs/app.log|/tmp/test-siem-logs/auth.log|g' "$CONFIG_FILE" 2>/dev/null || true

CONFIG_PATH="$CONFIG_FILE" "$BINARY" &
AGENT_PID=$!
sleep 2

if ! ps -p $AGENT_PID > /dev/null 2>&1; then
    echo "ERROR: Agent failed to start"
    exit 1
fi
echo "OK: Agent started (PID: $AGENT_PID)"

echo "Dec 15 12:00:01 testhost sshd[1001]: Accepted password for admin from 192.168.1.100 port 22 ssh2" >> "$TEST_DIR/auth.log"
echo "Dec 15 12:00:05 testhost sshd[1002]: Failed password for root from 10.0.0.50 port 22 ssh2" >> "$TEST_DIR/auth.log"
echo "Dec 15 12:00:10 testhost sudo[1003]:    admin : TTY=pts/0 ; PWD=/root ; USER=root ; COMMAND=/bin/cat /etc/shadow" >> "$TEST_DIR/auth.log"

sleep 7

kill $AGENT_PID 2>/dev/null || true
wait $AGENT_PID 2>/dev/null || true
echo "OK: Agent stopped"

echo "Checking database..."
RESULT=$(echo '{"database":"security_events","operation":"find","filter":{}}' | socat - TCP:127.0.0.1:5140 2>/dev/null)

if echo "$RESULT" | grep -q '"status":"success"'; then
    COUNT=$(echo "$RESULT" | grep -o '"count":[0-9]*' | cut -d':' -f2)
    echo "OK: Found $COUNT events in database"
    echo ""
    echo "Sample data:"
    echo "$RESULT" | python3 -m json.tool 2>/dev/null | head -30 || echo "$RESULT"
else
    echo "ERROR: Failed to query database"
    echo "$RESULT"
    exit 1
fi

echo ""
echo "TEST PASSED"
