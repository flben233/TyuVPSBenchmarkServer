# Lolicon VPS

## 介绍

Lolicon VPS 是一个自动化的云服务器测试与监控平台，官网地址：[https://vps.lolicon.llc/](https://vps.lolicon.llc/)。

## Lolicon Monitor 使用说明

我们的监控点位于腾讯云上海服务器，使用Ping检测延迟和在线情况。

### 1. 登录

鼠标移动到左上角头像，使用 Github 授权登录

### 2. 进入探针页面

鼠标移动到左上角头像，点击 `Loli 探针`

### 3. 添加主机

点击右上角加号“+”，输入必要信息，点击提交，之后展开主机信息卡片，复制主机 ID 以供后续部署 Agent 使用，如果你不需要 Agent 监控流量等信息，到这一步就结束了。

## 4. Agent 端部署

### 4.1 一键部署脚本 (推荐)

该部署脚本等价于下面手工部署的方法2

```shell
curl -fsSL https://raw.githubusercontent.com/flben233/TyuVPSBenchmarkServer/refs/heads/master/install_agent.sh | bash -s
```

卸载：

```shell
curl -fsSL https://raw.githubusercontent.com/flben233/TyuVPSBenchmarkServer/refs/heads/master/install_agent.sh | bash -s -- uninstall
```

### 4.2 手工部署

#### 4.0 准备工作

选定一个需要监控流量的网络接口：

```shell
ip addr
```

应该会得到如下类似的输出：

```shell
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host noprefixroute 
       valid_lft forever preferred_lft forever
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP group default qlen 1000
    link/ether 52:54:00:63:f2:a0 brd ff:ff:ff:ff:ff:ff
    altname enp0s5
    altname ens5
    inet 10.0.16.15/22 brd 10.0.19.255 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 2402:4e00:1420:a00:228e:7354:e0c:0/128 scope global 
       valid_lft forever preferred_lft forever
    inet6 fe80::5054:ff:fe63:f2a0/64 scope link 
       valid_lft forever preferred_lft forever
```

lo 是回环接口，不包含任何外部流量，此处 eth0 是我们需要监控的接口（可能在不同服务器上名字不同，如 enp0s5、ens5 等），先记下

#### 4.1 Docker Compose 部署

首先根据 Docker [官方文档](https://docs.docker.com/compose/install/linux/) 安装 Docker 和 Docker Compose，随后执行以下命令初始化项目目录：

```shell
# 创建项目文件夹
mkdir lolicon-monitor
# 进入项目文件夹
cd lolicon-monitor
```

接下来编辑 `docker-compose.yml` 文件， 请将 `INSPECTOR_NETWORK_IFACE` 环境变量设置为上面选择的接口名称。

`docker-compose.yml` 文件内容如下：

```yaml
services:
  agent:
    image: ghcr.io/flben233/lolicon-vps/agent:latest
    restart: unless-stopped
    network_mode: host
    environment:
      INSPECTOR_HOST_ID: "1" # 上面创建的检测主机 ID
      INSPECTOR_SERVER_URL: "https://vps.lolicon.llc"
      INSPECTOR_NETWORK_IFACE: "eth0"  # 上面选择的接口名称
      HOST_PROC: /host/proc
      HOST_SYS: /host/sys
      HOST_ETC: /host/etc
      TZ: "Asia/Shanghai"
    volumes:
      - type: bind
        source: /proc
        target: /host/proc
        read_only: true
      - type: bind
        source: /sys
        target: /host/sys
        read_only: true
      - type: bind
        source: /etc
        target: /host/etc
        read_only: true
    read_only: true
    tmpfs:
      - /tmp
    cap_drop:
      - ALL
```

编辑 `docker-compose.yml` ，把上述内容复制到文件中

```shell
vi docker-compose.yml
```

按 `i` 进入编辑模式，粘贴内容后按 `Esc` 退出编辑模式，输入 `:wq` 保存并退出。

然后启动 Agent，2分钟左右就可以在网页上看到Agent上报的信息：

```shell
docker-compose up -d
```

#### 4.2 直接运行

如果不想使用 Docker Compose，或者不是x86架构，也可以直接运行 Agent 二进制文件。

首先要在[Release 页面](https://github.com/flben233/TyuVPSBenchmarkServer/releases)下载对应平台的 Agent 二进制文件，然后使用以下指令直接运行（注意替换其中的环境变量值和可执行文件名称）：

这里的Agent二进制文件名字为 `lolicon-monitor-agent-linux-amd64`，请注意替换成对应的文件名。

```shell

INSPECTOR_HOST_ID=刚才的主机ID INSPECTOR_SERVER_URL=https://vps.lolicon.llc INSPECTOR_NETWORK_IFACE=上面选择的接口名称 lolicon-monitor-agent-linux-amd64
```

也可以使用systemd进行管理，我们直接编辑 `/etc/systemd/system/lolicon-monitor-agent.service` 文件：

```shell
vi /etc/systemd/system/lolicon-monitor-agent.service
```

按 `i` 进入编辑模式，粘贴内容后按 `Esc` 退出编辑模式，输入 `:wq` 保存并退出。

文件内容如下，注意将 `ExecStart` 路径替换成你放置 Agent 二进制文件的路径：

```ini
[Unit]
Description=Lolicon Monitor Agent
After=network.target

[Service]
Type=simple
User=root
Restart=on-failure
RestartSec=5s
ExecStart=/usr/share/lolicon-monitor/lolicon-monitor-agent-linux-amd64
Environment="INSPECTOR_HOST_ID=刚才的主机ID"
Environment="INSPECTOR_SERVER_URL=https://vps.lolicon.llc"
Environment="INSPECTOR_NETWORK_IFACE=上面选择的接口名称"

[Install]
WantedBy=multi-user.target
```

保存后执行以下命令启动服务：

```shell
systemctl daemon-reload
systemctl start lolicon-monitor-agent
systemctl enable lolicon-monitor-agent
```

## 5. 注意事项

Host ID 是全局唯一标识，拥有此 ID 的 Agent 上报的数据会关联到对应的监控主机上，请勿泄露此 ID 以免数据被篡改。

## WebSSH 云同步使用说明

### 概述

WebSSH 云同步功能允许你将保存的 SSH 连接信息加密后存储到云端，方便在多设备间同步。你的加密密钥仅保存在本地浏览器中，永远不会发送到服务器。

#### 设置密钥

在每个浏览器上首次使用前需要先设置加密密钥，在所有设备上都要设置相同的密钥以确保连接信息能够正确加密和解密：

1. 进入 WebSSH 页面
2. 点击连接列表上方的 🔑 按钮
3. 输入至少 8 位的密钥（建议使用强密码）
4. 可选择"本地保存"以便下次直接使用，无需重复输入
5. 点击"确认"完成设置

#### 上传连接信息

将本地保存的连接信息加密后上传到云端：

1. 在本地创建或编辑好需要的连接
2. 点击 ⬆ 按钮
3. 上传成功后，连接信息即已加密存储到服务器

#### 下载连接信息

从云端获取并解密连接信息到本地：

1. 点击 ⬇ 按钮
2. 如果之前未设置密钥，系统会提示你先设置密钥
3. 下载成功后，云端连接会解密并合并到本地列表（已存在的连接会自动跳过）

#### 重置密钥

如果忘记密码或需要清空云端数据：

1. 点击 🔑 按钮
2. 选择"重置密钥"模式
3. 阅读警告提示后点击"确认重置"
4. 此操作将**永久删除**云端所有加密连接数据，且**不可恢复**

### 注意事项

如果解密后在浏览器上设置了新的密钥，再执行上传操作，会使用新的密钥重新加密连接信息并覆盖云端数据。这个操作也可以用来更换密钥，但请确保新密钥安全可靠，并且在所有设备上都要更新为新密钥以避免无法解密数据。

### 安全说明

- **密钥不传输**：你的密钥仅在前端用于加密/解密，永远不会发送到服务器
- **本地明文存储**：连接信息在本地浏览器中以明文保存，方便直接使用
- **云端加密存储**：只有上传到服务器的数据才会被加密
- **手动操作**：上传和下载均需手动触发，不会自动同步
- **妥善保管密钥**：忘记密钥将无法解密云端数据，请牢记或安全保存
