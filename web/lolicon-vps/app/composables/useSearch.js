export function useSearch() {
  const { backendUrl } = useAppConfig();

  // Get all backroute types
  async function getBackrouteTypes() {
    try {
      const resp = await $fetch(`${backendUrl}/report/data/backroute-types`, {
        method: "GET",
      });

      if (resp && resp.code === 0) {
        return resp.data || [];
      }
      return [];
    } catch (error) {
      console.error("Failed to get backroute types:", error);
      return [];
    }
  }

  // Get all media names
  async function getMediaNames() {
    try {
      const resp = await $fetch(`${backendUrl}/report/data/media-names`, {
        method: "GET",
      });

      if (resp && resp.code === 0) {
        return resp.data || [];
      }
      return [];
    } catch (error) {
      console.error("Failed to get media names:", error);
      return [];
    }
  }

  // Get all virtualization types
  async function getVirtualizations() {
    try {
      const resp = await $fetch(`${backendUrl}/report/data/virtualizations`, {
        method: "GET",
      });

      if (resp && resp.code === 0) {
        return resp.data || [];
      }
      return [];
    } catch (error) {
      console.error("Failed to get virtualizations:", error);
      return [];
    }
  }

  // Search reports with filters
  async function searchReports(searchParams, page = 1, pageSize = 10) {
    try {
      const resp = await $fetch(`${backendUrl}/report/data/search`, {
        method: "POST",
        query: { page, page_size: pageSize },
        body: searchParams,
      });

      if (resp && resp.code === 0) {
        return resp;
      }
      return { data: [], total: 0, page: 1, page_size: pageSize };
    } catch (error) {
      console.error("Failed to search reports:", error);
      return { data: [], total: 0, page: 1, page_size: pageSize };
    }
  }

  return {
    getBackrouteTypes,
    getMediaNames,
    getVirtualizations,
    searchReports,
  };
}
