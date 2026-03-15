#!/bin/bash

# Lolicon Monitor Agent One-Click Install Script
# Repository: https://github.com/flben233/TyuVPSBenchmarkServer

set -e

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
  x86_64)
    BINARY_ARCH="amd64"
    ;;
  aarch64|arm64)
    BINARY_ARCH="arm64"
    ;;
  loongarch64)
    BINARY_ARCH="loong64"
    ;;
  mips64*)
    # Defaulting to mips64le as specified in supported list
    BINARY_ARCH="mips64le"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

if [ "$OS" != "linux" ] && [ "$OS" != "darwin" ]; then
    echo "Unsupported OS: $OS"
    exit 1
fi

# Check for root if Linux (for systemd)
if [ "$OS" = "linux" ] && [ "$EUID" -ne 0 ]; then
  echo "Please run as root to install systemd service"
  exit 1
fi

echo "--- Lolicon Monitor Agent Installer ---"
echo "Detected: $OS-$BINARY_ARCH"

# Show Available Network Interfaces
echo ""
echo "Available network interfaces:"
if [ "$OS" = "linux" ]; then
    ip -o link show | awk -F': ' '{print $2}' | grep -v "lo" | sed 's/^/  - /'
    DEFAULT_IFACE=$(ip route get 8.8.8.8 2>/dev/null | grep -oP 'dev \K\S+' || ip link | grep 'state UP' | awk '{print $2}' | sed 's/://' | head -n 1)
else
    ifconfig -l | tr ' ' '\n' | grep -v "lo" | sed 's/^/  - /'
    DEFAULT_IFACE=$(route -n get default 2>/dev/null | grep 'interface' | awk '{print $2}')
fi
echo ""

# Prompt for Configuration
read -p "Enter INSPECTOR_HOST_ID: " INSPECTOR_HOST_ID
if [ -z "$INSPECTOR_HOST_ID" ]; then
  echo "Error: INSPECTOR_HOST_ID is required."
  exit 1
fi

read -p "Enter INSPECTOR_SERVER_URL [https://vps.lolicon.llc]: " INSPECTOR_SERVER_URL
INSPECTOR_SERVER_URL=${INSPECTOR_SERVER_URL:-https://vps.lolicon.llc}

read -p "Enter INSPECTOR_NETWORK_IFACE [$DEFAULT_IFACE]: " INSPECTOR_NETWORK_IFACE
INSPECTOR_NETWORK_IFACE=${INSPECTOR_NETWORK_IFACE:-$DEFAULT_IFACE}

# Preparation
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="lolicon-monitor-agent"
FULL_BINARY_NAME="lolicon-monitor-agent-$OS-$BINARY_ARCH"
REPO="flben233/TyuVPSBenchmarkServer"

echo "Downloading agent binary ($FULL_BINARY_NAME)..."
LATEST_TAG=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep -oP '"tag_name": "\K[^"]+')

if [ -z "$LATEST_TAG" ]; then
  echo "Error: Could not fetch latest release tag."
  exit 1
fi

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/$FULL_BINARY_NAME"

curl -L -o "$INSTALL_DIR/$BINARY_NAME" "$DOWNLOAD_URL"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

if [ "$OS" = "linux" ]; then
    echo "Creating systemd service..."
    SERVICE_PATH="/etc/systemd/system/lolicon-monitor-agent.service"
    cat <<EOF > "$SERVICE_PATH"
[Unit]
Description=Lolicon Monitor Agent
After=network.target

[Service]
Type=simple
User=root
Restart=on-failure
RestartSec=5s
ExecStart=$INSTALL_DIR/$BINARY_NAME
Environment="INSPECTOR_HOST_ID=$INSPECTOR_HOST_ID"
Environment="INSPECTOR_SERVER_URL=$INSPECTOR_SERVER_URL"
Environment="INSPECTOR_NETWORK_IFACE=$INSPECTOR_NETWORK_IFACE"

[Install]
WantedBy=multi-user.target
EOF

    echo "Starting service..."
    systemctl daemon-reload
    systemctl enable lolicon-monitor-agent
    systemctl start lolicon-monitor-agent

    echo "--- Installation Complete ---"
    echo "Service status: systemctl status lolicon-monitor-agent"
    echo "Logs: journalctl -u lolicon-monitor-agent -f"
else
    echo "--- Download Complete ---"
    echo "Note: Automatic service creation is not supported on macOS in this script."
    echo "You can run the agent manually with:"
    echo "INSPECTOR_HOST_ID=$INSPECTOR_HOST_ID INSPECTOR_SERVER_URL=$INSPECTOR_SERVER_URL INSPECTOR_NETWORK_IFACE=$INSPECTOR_NETWORK_IFACE $INSTALL_DIR/$BINARY_NAME"
fi
