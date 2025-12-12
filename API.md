# VPSBenchmarkBackend REST API 文档

本文档基于代码仓库内现有路由与 handler 实现自动整理，放置于仓库根目录，便于开发和测试人员参考。

通用返回格式（JSON）:

```json
// 单次请求成功
{
  "code": 0,
  "message": "success",
  "data": { /* 实际数据 */ }
}

// 分页数据
{
  "code": 0,
  "message": "success",
  "data": [ /* items */ ],
  "total": 123,
  "page": 1,
  "page_size": 10
}

// 失败示例
{
  "code": -2,
  "message": "bad request"
}
```

鉴权与中间件:
- `GET /auth/user`、监控/管理类和 `report/admin` 路由受 JWT 验证保护；使用 `Authorization: Bearer <token>`。
- `report/admin` 额外使用 Admin 校验中间件。

模块路由总览（以 `base` 前缀为根，例如在代码中 RegisterRouter 的 `base`）：

1) Auth
- POST `/auth/github/login` — 使用 GitHub OAuth Code 交换登录 token。
- GET `/auth/user` — 获取当前 JWT 中的用户信息（需 Bearer token）。

2) Monitor
- GET `/monitor/statistics` — 获取监控统计（公开）。
- GET `/monitor/hosts` — 列出当前用户的监控主机（需 JWT）。
- POST `/monitor/hosts` — 新增监控主机（需 JWT）。
- POST `/monitor/hosts/:id` — 删除监控主机（需 JWT）。

3) Report（数据查询）
- GET `/report/data/list` — 分页列出报告。
- GET `/report/data/details` — 获取单条报告详情。
- GET or POST `/report/data/search` — 搜索报告，支持 JSON body 或 query params。
- GET `/report/data/media-names` — 列出所有媒体解锁名称
- GET `/report/data/virtualizations` — 列出所有虚拟化名称
- GET `/report/data/backroute-types` — 列出所有回程类型

4) Report（管理员）
- POST `/report/admin/add` — 添加报告（需 JWT + Admin）
- POST `/report/admin/delete` — 删除报告（需 JWT + Admin）

5) Tool
- GET/POST `/tool/ip` — IP 信息查询
- GET/POST `/tool/traceroute` — 路由追踪（依赖外部 `nexttrace`）
- GET/POST `/tool/whois` — WHOIS 查询（依赖外部 `whois`）

详细接口说明

**Auth**

- POST `/auth/github/login`
  - 描述：使用 GitHub OAuth `code` 换取本服务 JWT。 
  - 请求体 (JSON): `{ "code": "<github_oauth_code>" }`（必填）
  - 返回：`data` 为 `{"token": "<jwt>"}`

- GET `/auth/user`
  - 描述：返回当前 JWT 中保存的用户信息。
  - 需要：`Authorization: Bearer <token>`
  - 返回示例 `data`:
    ```json
    { "name": "username", "avatar_url": "https://..." }
    ```

**Monitor**

- GET `/monitor/statistics`
  - 描述：获取系统监控统计。

- GET `/monitor/hosts`
  - 描述：列出当前用户（或管理员）可见的监控主机。
  - 返回 `data`：主机列表，具体字段以 `internal/monitor/model/monitor.go` 中定义为准。

- POST `/monitor/hosts`
  - 描述：新增监控主机（需 JWT）。
  - 请求体 (JSON): `{ "target": "目标地址或域名", "name": "主机别名" }`
  - 成功返回 201 并包含 `{"id": <new_id>}`

- POST `/monitor/hosts/:id`
  - 描述：删除主机（路由实现为 POST 而非 DELETE）。
  - 参数：路径参数 `:id` 为主机 ID。

**Report（数据查询）**

- GET `/report/data/list`
  - Query: `page`, `page_size`
  - 返回：分页列表 `data` 为 `[]ReportInfoResponse`（查看 `internal/report/response/data.go`）

- GET `/report/data/details?id=...`
  - Query: `id`（必需）
  - 返回：单条完整报告数据（模型位于 `internal/report/model/*`）

- GET/POST `/report/data/search`
  - 支持 JSON body 或 query 参数。
  - JSON Body 示例（`SearchRequest`）:
    ```json
    {
      "name": "关键字",
      "media_unlocks": ["TikTok"],
      "virtualization": "KVM",
      "ipv6_support": true,
      "disk_level": 3,
      "ct_params": { "back_route": "某回程", "min_download": 50.0 }
    }
    ```
  - 分页：`page`, `page_size`

**Report（管理员）**

- POST `/report/admin/add` (需要 JWT + admin)
  - 支持 JSON `{ "html": "<report html>" }` 或者直接把 HTML 放在请求 body。
  - 返回 201, `data` 示例: `{ "report_id": "<id>" }`

- POST `/report/admin/delete` (需要 JWT + admin)
  - JSON Body: `{ "id": "<report id>" }` 或者 query `id`

**Tool**

- GET/POST `/tool/ip`
  - 请求体/参数: `target` (必需), `dataSource` (可选，`ipinfo` 或 `ip-api`)
  - 返回：第三方 IP 服务的解析数据（map）

- GET/POST `/tool/traceroute`
  - 请求体/参数: `target` (必需), `mode` (`icmp` 或 `tcp` 默认 `icmp`), `port` (可选)
  - 返回：`data` 包含 `{"raw": "nexttrace 输出 json 字符串"}`

- GET/POST `/tool/whois`
  - 请求体/参数: `target` (必需)
  - 返回：`data` 包含 `{ "raw": "whois 原始输出" }`

备注与已知差异
- 路由实现中有少数与常见 REST 语义不完全一致的点（例如删除主机使用 `POST /monitor/hosts/:id`，非 `DELETE`）。文档以代码中实际注册的路由为准。
- 管理类接口需要 JWT 且通过 `auth.GetAdminMiddleware()` 进一步限制，请确保 `config` 中 AdminID/权限配置正确。

定位代码
- 路由入口：`internal/*/router.go`（已扫描：`auth`, `monitor`, `report`, `tool`）
- 重要 handler：`internal/*/handler/*.go`
- 请求/响应模型：`internal/*/request`, `internal/*/response`, `internal/*/model`

如果你希望我再生成 `openapi.yaml`（OpenAPI 3.0 规范）或自动提取更多字段（如模型字段类型与示例），我可以继续把这些模型解析并输出一个完整的 OpenAPI 文件。

-- 生成于代码扫描结果
