#!/bin/bash
# Generate test data for SIEM dashboard

DB_HOST="localhost"
DB_PORT="5140"

# Function to send insert command to NoSQLdb
send_insert() {
    local data="$1"
    echo "$data" | nc -q 1 $DB_HOST $DB_PORT
}

# Get current timestamp components for generating recent data
NOW=$(date +%s)

# Event types
EVENT_TYPES=("user_login" "auth_failure" "system_event" "file_access" "network_connection" "process_start" "privilege_escalation" "config_change")
SEVERITIES=("low" "medium" "high" "critical")
USERS=("admin" "root" "developer" "operator" "security" "guest" "service_account" "backup_user")
PROCESSES=("sshd" "sudo" "systemd" "nginx" "postgres" "docker" "cron" "auditd")
AGENTS=("agent-ubuntu-01" "agent-centos-02" "agent-debian-03" "agent-arch-04")
SOURCES=("auth" "syslog" "audit" "kernel" "application")

echo "Generating test data for SIEM..."

# Generate 100 events over the last 24 hours
for i in $(seq 1 100); do
    # Random time in last 24 hours
    OFFSET=$((RANDOM % 86400))
    TIMESTAMP=$(date -d "@$((NOW - OFFSET))" --iso-8601=seconds)
    
    # Random selections
    EVENT_TYPE=${EVENT_TYPES[$((RANDOM % ${#EVENT_TYPES[@]}))]}
    SEVERITY=${SEVERITIES[$((RANDOM % ${#SEVERITIES[@]}))]}
    USER=${USERS[$((RANDOM % ${#USERS[@]}))]}
    PROCESS=${PROCESSES[$((RANDOM % ${#PROCESSES[@]}))]}
    AGENT=${AGENTS[$((RANDOM % ${#AGENTS[@]}))]}
    SOURCE=${SOURCES[$((RANDOM % ${#SOURCES[@]}))]}
    
    # Generate IP
    IP="192.168.$((RANDOM % 256)).$((RANDOM % 256))"
    
    # Generate message based on event type
    case $EVENT_TYPE in
        "user_login")
            MESSAGE="Accepted password for $USER from $IP port 22 ssh2"
            ;;
        "auth_failure")
            MESSAGE="Failed password for $USER from $IP port 22 ssh2"
            ;;
        "system_event")
            MESSAGE="System event: service $PROCESS status changed"
            ;;
        "file_access")
            MESSAGE="File access: $USER accessed /etc/passwd"
            ;;
        "network_connection")
            MESSAGE="New connection from $IP to port $((RANDOM % 65535))"
            ;;
        "process_start")
            MESSAGE="Process $PROCESS started by $USER"
            ;;
        "privilege_escalation")
            MESSAGE="$USER escalated privileges via sudo"
            ;;
        "config_change")
            MESSAGE="Configuration changed in /etc/sysconfig by $USER"
            ;;
    esac
    
    # Create JSON payload
    PAYLOAD=$(cat <<EOF
{
    "database": "security_events",
    "operation": "insert",
    "data": [{
        "agent_id": "$AGENT",
        "timestamp": "$TIMESTAMP",
        "event_type": "$EVENT_TYPE",
        "severity": "$SEVERITY",
        "user": "$USER",
        "process": "$PROCESS",
        "source": "$SOURCE",
        "source_ip": "$IP",
        "hostname": "testhost-$((RANDOM % 10))",
        "message": "$MESSAGE",
        "raw_log": "$(date -d "@$((NOW - OFFSET))" '+%b %d %H:%M:%S') testhost $PROCESS[$$]: $MESSAGE"
    }]
}
EOF
)
    
    echo "Inserting event $i: $EVENT_TYPE ($SEVERITY)"
    echo "$PAYLOAD" | nc -q 1 $DB_HOST $DB_PORT
    
done

echo "Done! Generated 100 test events."
