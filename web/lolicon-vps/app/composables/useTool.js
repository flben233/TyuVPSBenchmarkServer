export function useTool() {
  const { backendUrl } = useAppConfig();

  async function ipQuery(target, dataSource = "ipapi") {
    try {
      const resp = await useFetch(`${backendUrl}/tool/ip`, {
        method: "GET",
        query: { target, data_source: dataSource },
      });

      if (resp.data.value && resp.data.value.code === 0) {
        return resp.data.value;
      }
      return { code: -1, message: "Failed to query IP", data: null };
    } catch (error) {
      console.error("Failed to query IP:", error);
      return { code: -1, message: error.message, data: null };
    }
  }

  async function traceroute(target, mode = "icmp", port = null) {
    try {
      const query = { target, mode };
      if (port !== null) {
        query.port = port;
      }

      const resp = await useFetch(`${backendUrl}/tool/traceroute`, {
        method: "GET",
        query,
      });

      if (resp.data.value && resp.data.value.code === 0) {
        return resp.data.value;
      }
      return { code: -1, message: "Failed to perform traceroute", data: null };
    } catch (error) {
      console.error("Failed to perform traceroute:", error);
      return { code: -1, message: error.message, data: null };
    }
  }

  async function whois(target) {
    try {
      const resp = await useFetch(`${backendUrl}/tool/whois`, {
        method: "GET",
        query: { target },
      });

      if (resp.data.value && resp.data.value.code === 0) {
        return resp.data.value;
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
  };
}
