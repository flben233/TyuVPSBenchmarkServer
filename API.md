# VPSBenchmarkBackend API 文档

本文件详细说明了 VPSBenchmarkBackend 的 API 接口、请求参数及返回结构。

- **基础路径 (BasePath):** `/api`
- **公共响应结构:**
  所有接口通常返回以下 JSON 结构：
  ```json
  {
    "code": 200,      // 状态码 (200 为成功)
    "message": "OK",  // 提示信息
    "data": { ... }   // 实际数据内容
  }
  ```

---

## 🔒 认证模块 (Auth)

### 1. GitHub OAuth 登录
`GET /auth/github/login`
- **参数:** `code` (query, string) - GitHub 返回的临时 code
- **返回:** `common.APIResponse-response_LoginResponse`
  - `data.token`: JWT Token

### 2. 刷新 JWT Token
`POST /auth/refresh`
- **返回:** `common.APIResponse-response_LoginResponse`
  - `data.token`: 新的 JWT Token

### 3. 获取当前用户信息
`GET /auth/user` (需 JWT)
- **返回:** `common.APIResponse-response_UserInfo`
  - `data.name`: 用户名
  - `data.avatar_url`: 头像链接

### 4. 后台用户管理 (Admin)
- `GET /auth/admin/users`: 列出用户。返回 `[]model.User`。
- `POST /auth/admin/user`: 更新用户。请求体 `model.User` { `id`, `name`, `group_id` }。
- `POST /auth/admin/user/delete`: 删除用户。请求体 `{ "UserID": 123 }`。
- `GET /auth/admin/groups`: 列出所有用户组。返回 `[]model.UserGroup`。
- `POST /auth/admin/group/create`: 创建用户组。请求体 `model.UserGroup` { `name`, `permissions` }。
- `POST /auth/admin/group`: 更新用户组。请求体 `model.UserGroup` { `id`, `name`, `permissions` }。
- `POST /auth/admin/group/delete`: 删除用户组。请求体 `{ "GroupID": 123 }`。

---

## 📊 监控模块 (Monitor)

### 1. 获取服务器运行状态
`GET /monitor/status`
- **返回:** `response.ServerStatusResponse`
  - `cpu_usage_percent`: CPU 使用率
  - `memory_usage_percent`: 内存使用率
  - `upload_mbps`/`download_mbps`: 实时网络 IO

### 2. 获取监控统计信息
`GET /monitor/statistics`
- **返回:** `response.StatisticsResponse`
  - `total_hosts`: 监控主机总数
  - `online_hosts`: 在线主机数
  - `offline_hosts`: 离线主机数

### 3. 列出我的监控主机
`GET /monitor/hosts` (需 JWT)
- **返回:** `[]model.Host` (简要信息)

### 4. 新增监控主机
`POST /monitor/hosts/add` (需 JWT)
- **请求体:** `request.HostRequest`
  ```json
  {
    "name": "主机名",
    "target": "检测目标 (IP/域名)"
  }
  ```
- **返回:** `{ "data": { "id": 123 } }`

### 5. 删除监控主机
`POST /monitor/hosts/delete/{id}` (需 JWT)
- **参数:** `id` (path, int) - 主机 ID
- **返回:** `{ "code": 200, "message": "删除成功" }`

### 6. 审核管理 (Admin)
- `GET /monitor/admin/pending`: 列出待审核主机。返回 `[]model.Host`。
- `POST /monitor/admin/approve/{id}`: 审核通过主机。参数 `id` (path, int)。
- `POST /monitor/admin/reject/{id}`: 审核拒绝主机。参数 `id` (path, int)。

---

## 📑 报告模块 (Report)

### 1. 分页获取报告列表
`GET /report/data/list`
- **参数:** `page`, `page_size` (query, int)
- **返回:** `common.PaginatedResponse-array_response_ReportInfoResponse`

### 2. 获取报告详情
`GET /report/data/details`
- **参数:** `id` (query, string)
- **返回:** `model.BenchmarkResult` (包含详细的 CPU, Disk, Speedtest, Media 解锁等数据)

### 3. 搜索报告
`POST /report/data/search`
- **请求体:** `request.SearchRequest`
  ```json
  {
    "name": "关键词",
    "virtualization": "KVM",
    "ipv6_support": true,
    "media_unlocks": ["Netflix", "Disney+"],
    "cm_params": { "back_route": "CN2 GIA", "latency": 150 }
  }
  ```
- **分页参数:** `page`, `page_size` (query)
- **返回:** `common.PaginatedResponse-array_response_ReportInfoResponse`

### 4. 添加报告
`POST /report/admin/add` (需 JWT)
- **请求体:** `AddReportRequest` (body/HTML)
- **返回:** `{ "code": 200, "message": "添加成功", "data": { "id": 123 } }`

### 5. 更新报告对应 Monitor ID
`POST /report/admin/delete` (需 JWT)
- **请求体:** `UpdateReportRequest` (body)
- **返回:** `{ "code": 200, "message": "更新成功" }`

---

## 🔍 检测器模块 (Inspector)

### 1. 列出检测主机
`GET /inspector/hosts` (需 JWT)
- **返回:** `[]model.InspectorHost` (简要信息)

### 2. 创建检测主机
`POST /inspector/hosts/create` (需 JWT)
- **请求体:** `CreateHostRequest` (body)
  ```json
  {
    "name": "主机名",
    "target": "检测目标 (IP/域名)"
  }
  ```
- **返回:** `{ "data": { "id": 123 } }`

### 3. 更新检测主机
`POST /inspector/hosts/update/{id}` (需 JWT)
- **参数:** `id` (path, int) - 主机 ID
- **请求体:** `UpdateHostRequest` (body)
- **返回:** `{ "code": 200, "message": "更新成功" }`

### 4. 删除检测主机
`POST /inspector/hosts/delete/{id}` (需 JWT)
- **参数:** `id` (path, int) - 主机 ID
- **返回:** `{ "code": 200, "message": "删除成功" }`

### 5. 查询检测历史数据
`GET /inspector/data` (需 JWT)
- **参数:**
  - `start`: 开始时间 (纳秒时间戳)
  - `end`: 结束时间 (纳秒时间戳)
  - `interval`: 聚合间隔 (如 `1h`, `30m`)
- **返回:** `[]response.HostData` (包含一段时间内的 Ping 和流量趋势)

### 6. 上报检测数据 (Agent 使用)
`POST /inspector/data/put`
- **请求体:** `request.PutDataRequest`
  ```json
  {
    "host_id": 1,
    "hostInfo": { "cpu_usage_percent": 10.5, ... },
    "traffic": [ { "time": "...", "sent": 100, "recv": 200 } ]
  }
  ```

### 7. 获取个人检测设置
`GET /inspector/settings` (需 JWT)
- **返回:** `response.InspectorSettingsResponse`

### 8. 更新个人检测设置
`POST /inspector/settings/update` (需 JWT)
- **请求体:** `UpdateInspectorSettingRequest` (body)
- **返回:** `{ "code": 200, "message": "更新成功" }`

---

## 🔍 Looking Glass 模块

### 1. 列出公开的 LG 记录
`GET /lookingglass/list`
- **返回:** `[]model.LookingGlass` (公开记录)

### 2. 列出我的 LG 记录
`GET /lookingglass/records` (需 JWT)
- **返回:** `[]model.LookingGlass` (我的记录)

### 3. 添加 LG 记录
`POST /lookingglass/records` (需 JWT)
- **请求体:** `request.LookingGlassRequest`
  ```json
  {
    "server_name": "名称",
    "test_url": "测速文件地址"
  }
  ```
- **返回:** `{ "code": 200, "message": "添加成功", "data": { "id": 123 } }`

### 4. 删除 LG 记录
`POST /lookingglass/records/{id}` (需 JWT)
- **参数:** `id` (path, int) - 记录 ID
- **返回:** `{ "code": 200, "message": "删除成功" }`

### 5. 审核管理 (Admin)
- `GET /lookingglass/admin/pending`: 列出待审核 LG 记录。返回 `[]model.LookingGlass`。
- `POST /lookingglass/admin/approve/{id}`: 审核通过 LG 记录。参数 `id` (path, int)。
- `POST /lookingglass/admin/reject/{id}`: 审核拒绝 LG 记录。参数 `id` (path, int)。

---

## 📂 数据结构详解 (Definitions)

### model.User
```json
{
  "id": 12345,        // GitHub ID
  "login": "username",
  "name": "Nickname",
  "group_id": 1
}
```

### model.BenchmarkResult (报告详情核心结构)
- `cpu`: { `single`, `multi` }
- `disk`: { `seq_read`, `seq_write` }
- `spdtest`: Speedtest 结果列表
- `media`: 流媒体解锁情况 (IPv4/IPv6)
- `itdog`: Ping 和路由追踪结果
