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

  return {
    listReports,
    getReportDetails,
  };
}
