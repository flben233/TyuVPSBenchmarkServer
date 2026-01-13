export function useMonitor() {
  const { backendUrl } = useAppConfig();

  async function getStatistics() {
    try {
      const resp = await useFetch(`${backendUrl}/monitor/statistics`, {
        method: "GET",
      });

      if (resp.data.value && resp.data.value.code === 0) {
        return resp.data.value.data || [];
      }
      return [];
    } catch (error) {
      console.error("Failed to get monitor statistics:", error);
      return [];
    }
  }

  async function getServerStatus() {
    try {
      const resp = await $fetch(`${backendUrl}/monitor/status`, {
        method: "GET",
      });
      
      if (resp && resp.code === 0) {
        return resp.data;
      }
      return null;
    } catch (error) {
      console.error("Failed to get server status:", error);
      return null;
    }
  }

  return {
    getStatistics,
    getServerStatus,
  };
}
