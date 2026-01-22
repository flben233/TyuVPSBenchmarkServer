import { da } from "element-plus/es/locale/index.mjs";

export function useTool() {
  const { backendUrl } = useAppConfig();
  const dataSourceOptions = [
    { label: "ipapi.is", value: "ipapi" },
    { label: "ipinfo.io", value: "ipinfo" },
  ];

  async function ipQuery(target, dataSource = "ipapi") {
    if (dataSource === "ipinfo") {
      if (target === "") {
        target = await $fetch("https://ipinfo.io/ip", {
          method: "GET",
        });
        target = target.trim();
      }
      const resp = await $fetch(`https://ipinfo.io/widget/demo/${target}`, {
        method: "GET",
      });
      return resp;
    } else if (dataSource === "ipapi") {
      if (target === "") {
        const resp = await $fetch(`https://api.ipapi.is/`, {
          method: "GET",
        });
        return resp;
      }
      const resp = await $fetch(`https://api.ipapi.is/?ip=${target}`, {
        method: "GET",
      });
      return resp;
    }
  }

  async function traceroute(target, mode = "icmp", port = null) {
    try {
      const query = { target, mode };
      if (port !== null) {
        query.port = port;
      }

      const resp = await $fetch(`${backendUrl}/tool/traceroute`, {
        method: "GET",
        query,
      });

      if (resp && resp.code === 0) {
        return resp;
      }
      return { code: -1, message: "Failed to perform traceroute", data: null };
    } catch (error) {
      console.error("Failed to perform traceroute:", error);
      return { code: -1, message: error.message, data: null };
    }
  }

  async function whois(target) {
    try {
      const resp = await $fetch(`${backendUrl}/tool/whois`, {
        method: "GET",
        query: { target },
      });

      if (resp && resp.code === 0) {
        return resp;
      }
      return { code: -1, message: "Failed to query WHOIS", data: null };
    } catch (error) {
      console.error("Failed to query WHOIS:", error);
      return { code: -1, message: error.message, data: null };
    }
  }

  return {
    ipQuery,
    traceroute,
    whois,
    dataSourceOptions,
  };
}
