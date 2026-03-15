#!/bin/bash

# Lolicon 探针一键安装/卸载脚本
# 项目地址: https://github.com/flben233/TyuVPSBenchmarkServer

set -e

# 检测操作系统和架构
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
    BINARY_ARCH="mips64le"
    ;;
  *)
    echo "错误: 不支持的架构: $ARCH"
    exit 1
    ;;
esac

if [ "$OS" != "linux" ] && [ "$OS" != "darwin" ]; then
    echo "错误: 不支持的操作系统: $OS"
    exit 1
fi

INSTALL_DIR="/usr/local/bin"
BINARY_NAME="lolicon-monitor-agent"
SERVICE_NAME="lolicon-monitor-agent.service"

# 卸载逻辑
if [ "$1" = "uninstall" ]; then
    echo "--- Lolicon 探针卸载程序 ---"
    if [ "$OS" = "linux" ]; then
        if [ "$EUID" -ne 0 ]; then
            echo "错误: 请以 root 权限运行卸载程序。"
            exit 1
        fi
        echo "正在停止并禁用服务..."
        systemctl stop lolicon-monitor-agent || true
        systemctl disable lolicon-monitor-agent || true
        echo "正在删除 systemd 服务文件..."
        rm -f "/etc/systemd/system/$SERVICE_NAME"
        systemctl daemon-reload
    fi
    echo "正在删除二进制文件..."
    rm -f "$INSTALL_DIR/$BINARY_NAME"
    echo "--- 卸载完成 ---"
    exit 0
fi

# Linux 环境下检查 root 权限（安装 systemd 服务需要）
if [ "$OS" = "linux" ] && [ "$EUID" -ne 0 ]; then
  echo "错误: 请以 root 权限运行安装程序。"
  exit 1
fi

echo "--- Lolicon 探针安装程序 ---"
echo "检测到系统架构: $OS-$BINARY_ARCH"

# 显示可用的网络接口
echo ""
echo "当前可用的网络接口:"
if [ "$OS" = "linux" ]; then
    ip -o link show | awk -F': ' '{print $2}' | grep -v "lo" | sed 's/^/  - /'
    DEFAULT_IFACE=$(ip route get 8.8.8.8 2>/dev/null | grep -oP 'dev \K\S+' || ip link | grep 'state UP' | awk '{print $2}' | sed 's/://' | head -n 1)
else
    ifconfig -l | tr ' ' '\n' | grep -v "lo" | sed 's/^/  - /'
    DEFAULT_IFACE=$(route -n get default 2>/dev/null | grep 'interface' | awk '{print $2}')
fi
echo ""

# 配置输入
read -p "请输入主机 ID (INSPECTOR_HOST_ID): " INSPECTOR_HOST_ID
if [ -z "$INSPECTOR_HOST_ID" ]; then
  echo "错误: 必须填写主机 ID。"
  exit 1
fi

read -p "请输入服务端地址 (INSPECTOR_SERVER_URL) [默认: https://vps.lolicon.llc]: " INSPECTOR_SERVER_URL
INSPECTOR_SERVER_URL=${INSPECTOR_SERVER_URL:-https://vps.lolicon.llc}

read -p "请输入要监控的网卡名称 (INSPECTOR_NETWORK_IFACE) [默认: $DEFAULT_IFACE]: " INSPECTOR_NETWORK_IFACE
INSPECTOR_NETWORK_IFACE=${INSPECTOR_NETWORK_IFACE:-$DEFAULT_IFACE}

# 准备安装
FULL_BINARY_NAME="lolicon-monitor-agent-$OS-$BINARY_ARCH"
REPO="flben233/TyuVPSBenchmarkServer"

echo "正在从 GitHub 获取最新版本信息..."
LATEST_TAG=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep -oP '"tag_name": "\K[^"]+')

if [ -z "$LATEST_TAG" ]; then
  echo "错误: 无法获取最新版本号，请检查网络连接。"
  exit 1
fi

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/$FULL_BINARY_NAME"

echo "正在下载探针二进制文件 ($FULL_BINARY_NAME)..."
curl -L -o "$INSTALL_DIR/$BINARY_NAME" "$DOWNLOAD_URL"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

if [ "$OS" = "linux" ]; then
    echo "正在创建 systemd 服务..."
    SERVICE_PATH="/etc/systemd/system/$SERVICE_NAME"
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

    echo "正在启动服务..."
    systemctl daemon-reload
    systemctl enable lolicon-monitor-agent
    systemctl start lolicon-monitor-agent

    echo ""
    echo "--- 安装完成 ---"
    echo "服务状态: systemctl status lolicon-monitor-agent"
    echo "查看日志: journalctl -u lolicon-monitor-agent -f"
    echo "如需卸载，请运行: sudo $0 uninstall"
else
    echo ""
    echo "--- 下载完成 ---"
    echo "提示: 本脚本暂不支持在 macOS 上自动创建服务。"
    echo "您可以手动运行探针:"
    echo "INSPECTOR_HOST_ID=$INSPECTOR_HOST_ID INSPECTOR_SERVER_URL=$INSPECTOR_SERVER_URL INSPECTOR_NETWORK_IFACE=$INSPECTOR_NETWORK_IFACE $INSTALL_DIR/$BINARY_NAME"
    echo "如需卸载，请运行: rm $INSTALL_DIR/$BINARY_NAME"
fi
