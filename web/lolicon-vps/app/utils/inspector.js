export const APPRISE_DOCS_URL = "https://github.com/caronc/apprise/wiki";

export const NOTIFY_PRESETS = [
  { label: "钉钉", value: "dingtalk" },
  { label: "Email", value: "email" },
  { label: "飞书", value: "feishu" },
  { label: "Telegram", value: "telegram" },
  { label: "QQ", value: "qq" },
];

export const NOTIFY_PRESET_FIELDS = {
  dingtalk: {
    label: "钉钉机器人",
    description: "填写钉钉机器人 Access Token，若启用了签名可额外填写 Secret。",
    fields: [
      { key: "accessToken", label: "Access Token", required: true, placeholder: "请输入机器人 Access Token" },
      { key: "secret", label: "Secret", required: false, placeholder: "可选：签名 Secret" },
    ],
  },
  email: {
    label: "Email SMTP",
    description: "填写 SMTP 服务器、账号与收件邮箱，用于生成 mailto 协议 URL。",
    fields: [
      { key: "username", label: "用户名", required: true, placeholder: "example@gmail.com" },
      { key: "password", label: "密码", required: true, placeholder: "请输入邮箱密码或授权码", type: "password" },
      { key: "host", label: "SMTP 主机", required: true, placeholder: "smtp.gmail.com" },
      { key: "port", label: "端口", required: false, placeholder: "587" },
      { key: "to", label: "收件邮箱", required: true, placeholder: "receiver@example.com" },
      { key: "from", label: "发件邮箱", required: false, placeholder: "可选：显示发件地址" },
      { key: "secure", label: "启用 TLS", required: false, placeholder: "true / false" },
    ],
  },
  feishu: {
    label: "飞书机器人",
    description: "填写飞书自定义机器人 Webhook Token；若启用了签名可额外填写 Secret。",
    fields: [
      { key: "token", label: "Webhook Token", required: true, placeholder: "请输入 Webhook Token" },
      { key: "secret", label: "Secret", required: false, placeholder: "可选：签名 Secret" },
    ],
  },
  telegram: {
    label: "Telegram Bot",
    description: "填写 Bot Token 与 Chat ID。",
    fields: [
      { key: "botToken", label: "Bot Token", required: true, placeholder: "123456:ABCDEF" },
      { key: "chatId", label: "Chat ID", required: true, placeholder: "请输入 Chat ID" },
    ],
  },
  qq: {
    label: "QQ（Qmsg）",
    description: "使用常见的 Qmsg 酱方式生成 URL，填写 Key 与 QQ 号。",
    fields: [
      { key: "key", label: "Key", required: true, placeholder: "请输入 Qmsg Key" },
      { key: "qq", label: "QQ 号", required: true, placeholder: "请输入接收 QQ 号" },
    ],
  },
};

const EMPTY_SETTINGS = {
  notifyUrl: "",
  bgUrl: "",
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
    name: host.name || `Server ${host.id || ""}`.trim(),
    tags: parseTagList(host.tags),
    rawTags: host.tags || "",
    notify: Boolean(host.notify),
    notifyTolerance: normalizeNotifyTolerance(host.notify_tolerance),
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
    ping: Array.isArray(item.ping)
      ? item.ping.map((point) => ({
          hostId: toNumber(point.host_id),
          latency: toNumber(point.latency),
          time: point.time,
        }))
      : [],
  };
}

export function mergeHostData(hosts = [], queryData = []) {
  const dataMap = new Map(
    queryData.map((entry) => {
      const normalizedEntry = normalizeHostData(entry);
      return [normalizedEntry.id, normalizedEntry];
    }),
  );

  return hosts.map((host) => {
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
      const accessToken = encodeSegment(form.accessToken);
      const secret = encodeSegment(form.secret);
      return accessToken ? `dingtalk://${accessToken}${secret ? `/${secret}` : ""}` : "";
    }
    case "email": {
      const username = encodeSegment(form.username);
      const password = encodeSegment(form.password);
      const host = String(form.host || "").trim();
      const port = String(form.port || "").trim();
      const to = encodeSegment(form.to);
      const from = String(form.from || "").trim();
      const secure = String(form.secure || "").trim();

      if (!username || !password || !host || !to) {
        return "";
      }

      const base = `mailto://${username}:${password}@${host}${port ? `:${port}` : ""}/${to}`;
      const query = buildQueryString({ from, secure });
      return `${base}${query}`;
    }
    case "feishu": {
      const token = encodeSegment(form.token);
      const secret = encodeSegment(form.secret);
      return token ? `feishu://${token}${secret ? `/${secret}` : ""}` : "";
    }
    case "telegram": {
      const botToken = encodeSegment(form.botToken);
      const chatId = encodeSegment(form.chatId);
      return botToken && chatId ? `tgram://${botToken}/${chatId}` : "";
    }
    case "qq": {
      const key = encodeSegment(form.key);
      const qq = encodeSegment(form.qq);
      return key && qq ? `qmsg://${key}/${qq}` : "";
    }
    default:
      return "";
  }
}
