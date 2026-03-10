import {requestWithAuth} from "~/composables/useAuth.js";

export function useMonitor() {
  const { backendUrl } = useAppConfig();

  async function getStatistics(id = null) {
    try {
      const resp = await $fetch(`${backendUrl}/monitor/statistics`, {
        method: "GET",
        query: id ? { id } : {},
      });

      if (resp && resp.code === 0) {
        return resp.data || [];
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

  async function listHosts(token) {
    try {
      const resp = await requestWithAuth(`/monitor/hosts`, "GET");

      if (resp && resp.code === 0) {
        return resp.data || [];
      }
      return [];
    } catch (error) {
      console.error("Failed to list hosts:", error);
      return [];
    }
  }

  async function addHost(token, name, target) {
    try {
      const resp = await requestWithAuth(`/monitor/hosts/add`, "POST", {
        body: JSON.stringify({ name, target })
      });

      if (resp && resp.code === 0) {
        return { success: true, data: resp.data };
      }
      return { success: false, message: resp?.message || "Failed to add host" };
    } catch (error) {
      console.error("Failed to add host:", error);
      return { success: false, message: error.message };
    }
  }

  async function removeHost(token, id) {
    try {
      const resp = await requestWithAuth(`/monitor/hosts/delete/${id}`, "POST");

      if (resp && resp.code === 0) {
        return { success: true };
      }
      return {
        success: false,
        message: resp?.message || "Failed to remove host",
      };
    } catch (error) {
      console.error("Failed to remove host:", error);
      return { success: false, message: error.message };
    }
  }

  async function listPendingHosts(token) {
    try {
      const resp = await requestWithAuth(`/monitor/admin/pending`, "GET");

      if (resp && resp.code === 0) {
        return resp.data || [];
      }
      return [];
    } catch (error) {
      console.error("Failed to list pending hosts:", error);
      return [];
    }
  }

  async function approveHost(token, id) {
    try {
      const resp = await requestWithAuth(`/monitor/admin/approve/${id}`, "POST");

      if (resp && resp.code === 0) {
        return { success: true };
      }
      return {
        success: false,
        message: resp?.message || "Failed to approve host",
      };
    } catch (error) {
      console.error("Failed to approve host:", error);
      return { success: false, message: error.message };
    }
  }

  async function rejectHost(token, id) {
    try {
      const resp = await requestWithAuth(`/monitor/admin/reject/${id}`, "POST");

      if (resp && resp.code === 0) {
        return { success: true };
      }
      return {
        success: false,
        message: resp?.message || "Failed to reject host",
      };
    } catch (error) {
      console.error("Failed to reject host:", error);
      return { success: false, message: error.message };
    }
  }

  return {
    getStatistics,
    getServerStatus,
    listHosts,
    addHost,
    removeHost,
    listPendingHosts,
    approveHost,
    rejectHost,
  };
}
