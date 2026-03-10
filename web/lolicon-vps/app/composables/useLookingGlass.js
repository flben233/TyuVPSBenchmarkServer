import {requestWithAuth} from "~/composables/useAuth.js";

export function useLookingGlass() {
  const { backendUrl } = useAppConfig();

  async function listPublicLookingGlass() {
    return useFetch(`${backendUrl}/lookingglass/list`, {
      method: "GET",
    });
  }

  async function listRecords(token) {
    try {
      const resp = await requestWithAuth(`/lookingglass/records`, "GET");

      if (resp && resp.code === 0) {
        return resp.data || [];
      }
      return [];
    } catch (error) {
      console.error("Failed to list records:", error);
      return [];
    }
  }

  async function addRecord(token, serverName, testUrl) {
    try {
      const resp = await requestWithAuth(`/lookingglass/records`, "POST", {
        body: JSON.stringify({ server_name: serverName, test_url: testUrl }),
      });

      if (resp && resp.code === 0) {
        return { success: true, data: resp.data };
      }
      return { success: false, message: resp?.message || "Failed to add record" };
    } catch (error) {
      console.error("Failed to add record:", error);
      return { success: false, message: error.message };
    }
  }

  async function removeRecord(token, id) {
    try {
      const resp = await requestWithAuth(`/lookingglass/records/delete/${id}`, "POST");

      if (resp && resp.code === 0) {
        return { success: true };
      }
      return { success: false, message: resp?.message || "Failed to remove record" };
    } catch (error) {
      console.error("Failed to remove record:", error);
      return { success: false, message: error.message };
    }
  }

  async function listPendingRecords(token) {
    try {
      const resp = await requestWithAuth(`/lookingglass/admin/pending`, "GET");

      if (resp && resp.code === 0) {
        return resp.data || [];
      }
      return [];
    } catch (error) {
      console.error("Failed to list pending records:", error);
      return [];
    }
  }

  async function approveRecord(token, id) {
    try {
      const resp = await requestWithAuth(`/lookingglass/admin/approve/${id}`, "POST");

      if (resp && resp.code === 0) {
        return { success: true };
      }
      return { success: false, message: resp?.message || "Failed to approve record" };
    } catch (error) {
      console.error("Failed to approve record:", error);
      return { success: false, message: error.message };
    }
  }

  async function rejectRecord(token, id) {
    try {
      const resp = await requestWithAuth(`/lookingglass/admin/reject/${id}`, "POST");

      if (resp && resp.code === 0) {
        return { success: true };
      }
      return { success: false, message: resp?.message || "Failed to reject record" };
    } catch (error) {
      console.error("Failed to reject record:", error);
      return { success: false, message: error.message };
    }
  }

  return {
    listPublicLookingGlass,
    listRecords,
    addRecord,
    removeRecord,
    listPendingRecords,
    approveRecord,
    rejectRecord,
  };
}