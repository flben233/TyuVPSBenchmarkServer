export function useLookingGlass() {
  const { backendUrl } = useAppConfig();

  async function listPublicLookingGlass() {
    try {
      const resp = await useFetch(`${backendUrl}/lookingglass/list`, {
        method: "GET",
      });
      if (resp.data.value && resp.data.value.code === 0) {
        return resp.data.value.data || [];
      }
    } catch (error) {
      console.error("Failed to list public looking glass:", error);
    }
    return [];
  }

  async function listRecords(token) {
    try {
      const resp = await $fetch(`${backendUrl}/lookingglass/records`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

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
      const resp = await $fetch(`${backendUrl}/lookingglass/records`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
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
      const resp = await $fetch(`${backendUrl}/lookingglass/records/${id}`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

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
      const resp = await $fetch(`${backendUrl}/lookingglass/admin/pending`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

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
      const resp = await $fetch(`${backendUrl}/lookingglass/admin/approve/${id}`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

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
      const resp = await $fetch(`${backendUrl}/lookingglass/admin/reject/${id}`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

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