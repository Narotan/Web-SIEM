#!/bin/bash

set -e

echo "=== SIEM Agent Uninstallation Script ==="

if [ "$EUID" -ne 0 ]; then 
    echo "Error: This script must be run as root"
    exit 1
fi

echo "1. Stopping service..."
systemctl stop siem-agent || true

echo "2. Disabling service..."
systemctl disable siem-agent || true

echo "3. Removing systemd service..."
rm -f /etc/systemd/system/siem-agent.service
systemctl daemon-reload

echo "4. Removing binary..."
rm -f /usr/local/bin/siem-agent

echo "5. Removing config (optional, press Ctrl+C to skip)..."
read -p "Remove configuration files? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf /etc/siem-agent
    echo "Configuration removed"
fi

echo "6. Removing logs (optional)..."
read -p "Remove log files? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf /var/log/siem-agent
    echo "Logs removed"
fi

echo "7. Removing buffer (optional)..."
read -p "Remove buffer data? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf /var/lib/siem-agent
    echo "Buffer removed"
fi

echo ""
echo "=== Uninstallation Complete ==="
