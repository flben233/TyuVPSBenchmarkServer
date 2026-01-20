export function useReport() {
  const { backendUrl } = useAppConfig();

  async function listReports(page = 1, pageSize = 10) {
      return useFetch(`${backendUrl}/report/data/list`, {
        method: "GET",
        query: { page: page, page_size: pageSize },
      });
  }

  async function getReportDetails(id) {
      return useFetch(`${backendUrl}/report/data/details`, {
        method: "GET",
        query: { id },
      });
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
