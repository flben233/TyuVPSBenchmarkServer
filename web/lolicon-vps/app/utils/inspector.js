export const APPRISE_DOCS_URL = "https://appriseit.com/services/";

export const NOTIFY_PRESETS = [
  { label: "钉钉", value: "dingtalk" },
  { label: "Email", value: "email" },
  { label: "飞书", value: "feishu" },
  { label: "Telegram", value: "telegram" },
  { label: "QQ", value: "qq" },
];

export const NOTIFY_PRESET_FIELDS = {
  dingtalk: {
    label: "钉钉",
    description: "填写钉钉 API Key，可选填签名 Secret 和接收手机号。",
    fields: [
      { key: "apiKey", label: "API Key", required: true, placeholder: "请输入钉钉 API Key" },
      { key: "secret", label: "Secret", required: false, placeholder: "可选：签名 Secret" },
      { key: "to", label: "接收手机号", required: false, placeholder: "可选：手机号，多个用逗号分隔" },
    ],
  },
  email: {
    label: "Email",
    description: "填写邮箱地址与授权码，Apprise 会根据邮箱域名自动识别 SMTP 配置。常见邮箱请使用授权码而非登录密码。",
    fields: [
      { key: "email", label: "邮箱地址", required: true, placeholder: "user@gmail.com" },
      { key: "password", label: "密码 / 授权码", required: true, placeholder: "请输入邮箱密码或授权码", type: "password" },
      { key: "to", label: "收件邮箱", required: false, placeholder: "可选：不填则发送给自己" },
      { key: "smtp", label: "自定义 SMTP 主机", required: false, placeholder: "以下邮箱此项选填：Gmail, Yahoo, fastmail, GMX, zoho, Yandex, SendGrid, QQ/Foxmail, 163.com" },
      { key: "port", label: "自定义端口", required: false, placeholder: "可选：默认 mailtos 587，mailto 25" },
      { key: "from", label: "自定义发件人", required: false, placeholder: "可选：noreply@example.com" },
      { key: "mode", label: "加密模式", required: false, placeholder: "可选：ssl / starttls" },
    ],
  },
  feishu: {
    label: "飞书机器人",
    description: "填写飞书自定义机器人 Webhook Token；若启用了签名可额外填写 Secret。",
    fields: [
      { key: "token", label: "Webhook Token", required: true, placeholder: "请输入 Webhook Token" },
    ],
  },
  telegram: {
    label: "Telegram Bot",
    description: "填写 Bot Token 与 Chat ID。",
    fields: [
      { key: "botToken", label: "Bot Token", required: true, placeholder: "123456:ABCDEF" },
      { key: "chatId", label: "Chat ID", required: false, placeholder: "Chat ID，可不填，此时需要先给自己的机器人发一条消息" },
    ],
  },
  qq: {
    label: "QQ Push",
    description: "填写 QQ Push 的 Token，来自 qmsg.zendee.cn。",
    fields: [
      { key: "token", label: "Token", required: true, placeholder: "请输入 QQ Push Token" },
    ],
  },
};

const EMPTY_SETTINGS = {
  notifyUrl: "",
  bgUrl: "",
  visitorEnabled: false,
  allowedHostIds: [],
};

export const INSPECTOR_INTERVAL_OPTIONS = [
  { label: "5 分钟", value: "5m" },
  { label: "15 分钟", value: "15m" },
  { label: "30 分钟", value: "30m" },
  { label: "1 小时", value: "1h" },
  { label: "6 小时", value: "6h" },
  { label: "12 小时", value: "12h" },
  { label: "1 天", value: "1d" },
];

const INTERVAL_UNIT_TO_MS = {
  s: 1000,
  m: 60 * 1000,
  h: 60 * 60 * 1000,
  d: 24 * 60 * 60 * 1000,
};

function toNumber(value, fallback = 0) {
  const numericValue = Number(value);
  return Number.isFinite(numericValue) ? numericValue : fallback;
}

function normalizeNotifyTolerance(value) {
  return Math.max(0, Math.floor(toNumber(value, 0)));
}

export function getEmptyInspectorSettings() {
  return { ...EMPTY_SETTINGS };
}

export function getDefaultInspectorQuery() {
  const end = Date.now() * 1_000_000;
  const start = (Date.now() - 24 * 60 * 60 * 1000) * 1_000_000;

  return {
    start,
    end,
    interval: "5m",
  };
}

export function parseIntervalToMs(interval) {
  const matched = String(interval || "").trim().match(/^(\d+)([smhd])$/);
  if (!matched) {
    return 0;
  }

  const value = Number(matched[1]);
  const unitMs = INTERVAL_UNIT_TO_MS[matched[2]] || 0;
  return value > 0 && unitMs > 0 ? value * unitMs : 0;
}

export function exceedsInspectorPointLimit(startMs, endMs, interval, maxPoints = 120) {
  const intervalMs = parseIntervalToMs(interval);
  if (!Number.isFinite(startMs) || !Number.isFinite(endMs) || startMs >= endMs || intervalMs <= 0) {
    return false;
  }

  return Math.ceil((endMs - startMs) / intervalMs) > maxPoints;
}

export function parseTagList(rawTags) {
  if (!rawTags) {
    return [];
  }

  if (Array.isArray(rawTags)) {
    return rawTags.filter(Boolean).map((tag) => String(tag).trim()).filter(Boolean);
  }

  if (typeof rawTags !== "string") {
    return [];
  }

  const trimmed = rawTags.trim();
  if (!trimmed) {
    return [];
  }

  try {
    const parsed = JSON.parse(trimmed);
    if (Array.isArray(parsed)) {
      return parsed.map((tag) => String(tag).trim()).filter(Boolean);
    }
  } catch {
    return trimmed
      .split(/[，,]/)
      .map((tag) => tag.trim())
      .filter(Boolean);
  }

  return [];
}

export function stringifyTagList(tagInput) {
  const tags = Array.isArray(tagInput)
    ? tagInput
    : String(tagInput || "")
        .split(/[，,]/)
        .map((tag) => tag.trim())
        .filter(Boolean);

  if (tags.length === 0) {
    return "";
  }

  return JSON.stringify(tags);
}

export function normalizeInspectorSettings(payload) {
  return {
    notifyUrl: payload?.notify_url || "",
    bgUrl: payload?.bg_url || "",
    visitorEnabled: Boolean(payload?.visitor_enabled),
    allowedHostIds: Array.isArray(payload?.allowed_host_ids)
      ? payload.allowed_host_ids.map((id) => String(id))
      : [],
  };
}

export function normalizeHost(host = {}) {
  let lastUpdate = host.last_update;
  if (new Date(host.last_update).getTime() <= 0) {
    lastUpdate = null;
  }
  return {
    id: host.id,
    target: host.target || "",
    monitorType: host.monitor_type || "ping",
    name: host.name || `Server ${host.id || ""}`.trim(),
    tags: parseTagList(host.tags),
    rawTags: host.tags || "",
    notify: Boolean(host.notify),
    notifyTolerance: normalizeNotifyTolerance(host.notify_tolerance),
    trafficSettlementDay: toNumber(host.traffic_settlement_day, 0),
    monthlyTrafficLimit: toNumber(host.monthly_traffic_limit, 0),
    customOrder: toNumber(host.custom_order, 2147483647),
    latestPing: toNumber(host.latest_ping),
    uptimeSeconds: toNumber(host.uptime_seconds),
    lastUpdate: lastUpdate,
    cpuUsagePercent: toNumber(host.cpu_usage_percent),
    memoryTotalBytes: toNumber(host.memory_total_bytes),
    memoryUsedBytes: toNumber(host.memory_used_bytes),
    memoryUsagePercent: toNumber(host.memory_usage_percent),
    uploadMbps: toNumber(host.upload_mbps),
    downloadMbps: toNumber(host.download_mbps),
    system: host.system || "未知系统",
  };
}

export function normalizeHostData(item = {}) {
  const host = normalizeHost(item);

  return {
    ...host,
    sent: toNumber(item.sent),
    recv: toNumber(item.recv),
    loss: toNumber(item.loss),
    trafficUsage: toNumber(item.traffic_usage),
    trafficSettlementDay: toNumber(item.traffic_settlement_day, 0),
    monthlyTrafficLimit: toNumber(item.monthly_traffic_limit, 0),
    ping: Array.isArray(item.ping)
      ? item.ping.map((point) => ({
          hostId: toNumber(point.host_id),
          latency: toNumber(point.latency),
          time: point.time,
        }))
      : [],
  };
}

export function normalizeVisitorHost(item = {}) {
  const normalized = normalizeHost({
    ...item,
    id: "",
    target: "",
  });

  return {
    ...normalized,
    id: "",
    target: "",
    sent: toNumber(item.sent),
    recv: toNumber(item.recv),
    loss: toNumber(item.loss),
    ping: Array.isArray(item.ping)
      ? item.ping.map((point) => ({
          hostId: toNumber(point.host_id),
          latency: toNumber(point.latency),
          time: point.time,
        }))
      : [],
  };
}

export function normalizeVisitorPage(payload = {}) {
  const hosts = Array.isArray(payload.hosts) ? payload.hosts.map((host) => normalizeVisitorHost(host)) : [];

  return {
    ownerName: payload.owner_name || "",
    ownerId: payload.owner_id || "",
    bgUrl: payload.bg_url || "",
    hosts: hosts,
  };
}

export function mergeHostData(hosts = [], queryData = []) {
  const dataMap = new Map(
    queryData.map((entry) => {
      const normalizedEntry = normalizeHostData(entry);
      return [normalizedEntry.id, normalizedEntry];
    }),
  );

  const merged = hosts.map((host) => {
    const normalizedHost = normalizeHost(host);
    const metrics = dataMap.get(normalizedHost.id);

    if (!metrics) {
      return {
        ...normalizedHost,
        sent: 0,
        recv: 0,
        latestPing: normalizedHost.latestPing,
        ping: [],
      };
    }

    return {
      ...metrics,
      ...normalizedHost,
      sent: metrics.sent,
      recv: metrics.recv,
      latestPing: metrics.latestPing,
      ping: metrics.ping,
    };
  });

  return merged;
}

export function formatPercent(value, digits = 1) {
  return `${toNumber(value).toFixed(digits)}%`;
}

export function formatBytes(bytes) {
  const value = toNumber(bytes);
  if (value <= 0) {
    return "0 B";
  }

  const units = ["B", "KiB", "MiB", "GiB", "TiB"];
  let unitIndex = 0;
  let displayValue = value;

  while (displayValue >= 1024 && unitIndex < units.length - 1) {
    displayValue /= 1024;
    unitIndex += 1;
  }

  return `${displayValue.toFixed(displayValue >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`;
}

export function formatTrafficAmount(valueInMb) {
  const value = toNumber(valueInMb);
  if (value <= 0) {
    return "0 MB";
  }

  if (value >= 1024 * 1024) {
    return `${(value / (1024 * 1024)).toFixed(2)} TB`;
  }

  if (value >= 1024) {
    return `${(value / 1024).toFixed(2)} GB`;
  }

  return `${value.toFixed(value >= 10 ? 0 : 1)} MB`;
}

export function formatNetworkSpeed(valueInMbps) {
  const value = toNumber(valueInMbps);
  return `${value.toFixed(value >= 100 ? 0 : 1)} Mbps`;
}

export function formatLatency(value) {
  if (value === undefined || value === null || value === "") {
    return "暂无数据";
  }

  const latency = toNumber(value, NaN);
  if (!Number.isFinite(latency)) {
    return "暂无数据";
  }

  return latency > 0 ? `${latency.toFixed(1)} ms` : "超时";
}

export function formatUptime(seconds) {
  const totalSeconds = Math.max(0, Math.floor(toNumber(seconds)));
  if (!totalSeconds) {
    return "未上报";
  }

  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);

  if (days > 0) {
    return `${days}天 ${hours}小时`;
  }

  if (hours > 0) {
    return `${hours}小时 ${minutes}分钟`;
  }

  return `${minutes}分钟`;
}

export function formatTimestamp(timestamp) {
  if (!timestamp) {
    return "-";
  }

  return new Date(timestamp).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function getLatestPingValue(points = [], latestPing = null) {
  if (latestPing !== null && latestPing !== undefined && latestPing !== "") {
    return formatLatency(latestPing);
  }

  if (!Array.isArray(points) || points.length === 0) {
    return "暂无数据";
  }

  return formatLatency(points[points.length - 1]?.latency);
}

function buildQueryString(query = {}) {
  const searchParams = new URLSearchParams();
  Object.entries(query).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") {
      searchParams.set(key, String(value));
    }
  });
  const queryString = searchParams.toString();
  return queryString ? `?${queryString}` : "";
}

function encodeSegment(value) {
  return encodeURIComponent(String(value || "").trim());
}

export function buildAppriseUrl(type, form = {}) {
  switch (type) {
    case "dingtalk": {
      const apiKey = encodeSegment(form.apiKey);
      const secret = encodeSegment(form.secret);
      const toPhones = String(form.to || "").trim()
        .split(/[,，]/)
        .map((s) => s.trim())
        .filter(Boolean)
        .map(encodeSegment)
        .join("/");
      if (!apiKey) return "";
      const auth = secret ? `${secret}@` : "";
      const target = toPhones ? `/${toPhones}` : "";
      return `dingtalk://${auth}${apiKey}${target}`;
    }
    case "email": {
      const email = String(form.email || "").trim();
      const password = encodeSegment(form.password);
      const to = encodeSegment(form.to);
      const smtp = String(form.smtp || "").trim();
      const port = String(form.port || "").trim();
      const from = String(form.from || "").trim();
      const mode = String(form.mode || "").trim();

      const atIdx = email.indexOf("@");
      if (atIdx < 0 || !password) {
        return "";
      }

      const user = encodeSegment(email.substring(0, atIdx));
      const domain = encodeSegment(email.substring(atIdx + 1));

      if (!user || !domain) {
        return "";
      }

      const schema = "mailtos";
      const base = `${schema}://${user}:${password}@${domain}${port ? `:${port}` : ""}${to ? `/${to}` : ""}`;
      const query = buildQueryString({ smtp, from, mode });
      return `${base}${query}`;
    }
    case "feishu": {
      const token = encodeSegment(form.token);
      return token ? `feishu://${token} : ""}` : "";
    }
    case "telegram": {
      const botToken = encodeSegment(form.botToken);
      const chatId = encodeSegment(form.chatId);
      return botToken ? `tgram://${botToken}/${chatId}` : "";
    }
    case "qq": {
      const token = encodeSegment(form.token);
      return token ? `qq://${token}` : "";
    }
    default:
      return "";
  }
}

export function generateIdStar(id) {
  let stars = [];
  for (let i = 0; i < String(id).length; i++) {
    stars.push("*");
  }
  return stars.join("");
}

export function generateIpStar(ip) {
  let isV6 = ip.includes(":");
  let segments = isV6 ? ip.split(":") : ip.split(".");
  let stars = segments.map(() => "***");
  return isV6 ? stars.join(":") : stars.join(".");
}
