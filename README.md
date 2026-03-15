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
