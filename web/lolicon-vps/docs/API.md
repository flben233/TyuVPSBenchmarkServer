---
title: Lolicon VPS API
language_tabs:
  - shell: Shell
  - http: HTTP
  - javascript: JavaScript
  - ruby: Ruby
  - python: Python
  - php: PHP
  - java: Java
  - go: Go
toc_footers: []
includes: []
search: true
code_clipboard: true
highlight_theme: darkula
headingLevel: 2
generator: "@tarslib/widdershins v4.0.30"

---

# Lolicon VPS API

Base URLs:

# Authentication

# monitor

## GET Get Monitoring Statistics

GET /monitor/statistics

Retrieve overall monitoring statistics (public).

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "data": [
    {
      "history": [
        0
      ],
      "name": "string",
      "uploader": "string"
    }
  ],
  "message": "string"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[common.APIResponse-array_response_StatisticsResponse](#schemacommon.apiresponse-array_response_statisticsresponse)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal Server Error|[common.APIResponse-any](#schemacommon.apiresponse-any)|

# 数据模型

<h2 id="tocS_common.APIResponse-any">common.APIResponse-any</h2>

<a id="schemacommon.apiresponse-any"></a>
<a id="schema_common.APIResponse-any"></a>
<a id="tocScommon.apiresponse-any"></a>
<a id="tocscommon.apiresponse-any"></a>

```json
{
  "code": 0,
  "data": null,
  "message": "string"
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|code|integer|false|none||none|
|data|any|false|none||none|
|message|string|false|none||none|

<h2 id="tocS_common.APIResponse-array_response_StatisticsResponse">common.APIResponse-array_response_StatisticsResponse</h2>

<a id="schemacommon.apiresponse-array_response_statisticsresponse"></a>
<a id="schema_common.APIResponse-array_response_StatisticsResponse"></a>
<a id="tocScommon.apiresponse-array_response_statisticsresponse"></a>
<a id="tocscommon.apiresponse-array_response_statisticsresponse"></a>

```json
{
  "code": 0,
  "data": [
    {
      "history": [
        0
      ],
      "name": "string",
      "uploader": "string"
    }
  ],
  "message": "string"
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|code|integer|false|none||none|
|data|[[response.StatisticsResponse](#schemaresponse.statisticsresponse)]|false|none||none|
|message|string|false|none||none|

<h2 id="tocS_response.StatisticsResponse">response.StatisticsResponse</h2>

<a id="schemaresponse.statisticsresponse"></a>
<a id="schema_response.StatisticsResponse"></a>
<a id="tocSresponse.statisticsresponse"></a>
<a id="tocsresponse.statisticsresponse"></a>

```json
{
  "history": [
    0
  ],
  "name": "string",
  "uploader": "string"
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|history|[number]|false|none||An array of latency (ms) in the past of this host|
|name|string|false|none||Host alias|
|uploader|string|false|none||Host uploader|

