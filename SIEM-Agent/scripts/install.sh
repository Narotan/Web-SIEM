#!/bin/bash

set -e

echo "=== SIEM Agent Installation Script ==="

if [ "$EUID" -ne 0 ]; then 
    echo "Error: This script must be run as root"
    exit 1
fi

INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/siem-agent"
LOG_DIR="/var/log/siem-agent"
BUFFER_DIR="/var/lib/siem-agent/buffer"
BINARY_NAME="siem-agent"

echo "1. Creating directories..."
mkdir -p "$CONFIG_DIR"
mkdir -p "$LOG_DIR"
mkdir -p "$BUFFER_DIR"

echo "2. Building binary..."
cd "$(dirname "$0")/.."
go build -o "$BINARY_NAME" cmd/agent/main.go

echo "3. Installing binary..."
mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo "4. Installing config..."
if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
    cp configs/config.yaml "$CONFIG_DIR/config.yaml"
    echo "Config installed to $CONFIG_DIR/config.yaml"
else
    echo "Config already exists, skipping..."
fi

echo "5. Installing systemd service..."
cp deployments/systemd/siem-agent.service /etc/systemd/system/
systemctl daemon-reload

echo "6. Setting permissions..."
chown -R root:root "$CONFIG_DIR"
chown -R root:root "$LOG_DIR"
chown -R root:root "$BUFFER_DIR"
chmod 755 "$LOG_DIR"
chmod 755 "$BUFFER_DIR"

echo ""
echo "=== Installation Complete ==="
echo ""
echo "To start the agent:"
echo "  sudo systemctl start siem-agent"
echo ""
echo "To enable auto-start on boot:"
echo "  sudo systemctl enable siem-agent"
echo ""
echo "To check status:"
echo "  sudo systemctl status siem-agent"
echo ""
echo "To view logs:"
echo "  sudo journalctl -u siem-agent -f"
echo "  or"
echo "  sudo tail -f $LOG_DIR/agent.log"
echo ""
echo "Configuration file: $CONFIG_DIR/config.yaml"
