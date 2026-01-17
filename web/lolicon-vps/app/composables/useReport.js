export function useReport() {
  const { backendUrl } = useAppConfig();

  async function listReports(page = 1, pageSize = 10) {
    try {
      const resp = await useFetch(`${backendUrl}/report/data/list`, {
        method: "GET",
        query: { page: page, page_size: pageSize },
      });

      if (resp.data.value && resp.data.value.code === 0) {   
        return resp.data.value;
      }
      return { data: [], total: 0 };
    } catch (error) {
      console.error("Failed to list reports:", error);
      return { data: [], total: 0 };
    }
  }

  async function getReportDetails(id) {
    try {
      const resp = await useFetch(`${backendUrl}/report/data/details`, {
        method: "GET",
        query: { id },
      });
      if (resp.data.value && resp.data.value.code === 0) {
        return resp.data.value;
      }
      return null;
    } catch (error) {
      console.error("Failed to get report details:", error);
      return null;
    }
  }

  async function addReport(token, html, monitor_id) {
    try {
      const resp = await $fetch(`${backendUrl}/report/admin/add`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ 
          html: html,
          monitor_id: monitor_id
        }),
      });

      if (resp && resp.code === 0) {
        return { success: true, data: resp.data };
      }
      return { success: false, message: resp?.message || "Failed to add report" };
    } catch (error) {
      console.error("Failed to add report:", error);
      return { success: false, message: error.message };
    }
  }

  async function deleteReport(token, id) {
    try {
      const resp = await $fetch(`${backendUrl}/report/admin/delete`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ id }),
      });

      if (resp && resp.code === 0) {
        return { success: true };
      }
      return { success: false, message: resp?.message || "Failed to delete report" };
    } catch (error) {
      console.error("Failed to delete report:", error);
      return { success: false, message: error.message };
    }
  }

  return {
    listReports,
    getReportDetails,
    addReport,
    deleteReport,
  };
}
