import {
  getDefaultInspectorQuery,
  mergeHostData,
  normalizeInspectorSettings,
} from "~/utils/inspector";

export function useInspector() {
  const { backendUrl } = useAppConfig();

  function resolveErrorMessage(error, fallbackMessage) {
    return error?.data?.message || error?.message || fallbackMessage;
  }

  async function request(path, options = {}, fallbackMessage = "请求失败") {
    try {
      const response = await $fetch(`${backendUrl}${path}`, options);

      if (response?.code === 0) {
        return {
          success: true,
          data: response.data,
          message: response.message || "",
        };
      }

      return {
        success: false,
        data: null,
        message: response?.message || fallbackMessage,
      };
    } catch (error) {
      return {
        success: false,
        data: null,
        message: resolveErrorMessage(error, fallbackMessage),
      };
    }
  }

  function withAuth(token, options = {}) {
    return {
      ...options,
      headers: {
        ...(options.headers || {}),
        Authorization: `Bearer ${token}`,
      },
    };
  }

  async function listHosts(token) {
    return request(
      "/inspector/hosts",
      withAuth(token, { method: "GET" }),
      "获取服务器列表失败",
    );
  }

  async function queryData(token, query = getDefaultInspectorQuery()) {
    return request(
      "/inspector/data",
      withAuth(token, { method: "GET", query }),
      "获取探针数据失败",
    );
  }

  async function loadDashboard(token, query = getDefaultInspectorQuery()) {
    const [hostsResult, dataResult] = await Promise.all([listHosts(token), queryData(token, query)]);
    if (!hostsResult.success) {
      return hostsResult;
    }

    if (!dataResult.success) {
      return dataResult;
    }

    return {
      success: true,
      data: mergeHostData(hostsResult.data || [], dataResult.data || []),
      message: "",
    };
  }

  async function createHost(token, payload) {
    return request(
      "/inspector/hosts/create",
      withAuth(token, {
        method: "POST",
        body: payload,
      }),
      "创建服务器失败",
    );
  }

  async function updateHost(token, id, payload) {
    return request(
      `/inspector/hosts/update/${id}`,
      withAuth(token, {
        method: "POST",
        body: payload,
      }),
      "更新服务器失败",
    );
  }

  async function deleteHost(token, id) {
    return request(
      `/inspector/hosts/delete/${id}`,
      withAuth(token, { method: "POST" }),
      "删除服务器失败",
    );
  }

  async function getSettings(token) {
    const result = await request(
      "/inspector/settings",
      withAuth(token, { method: "GET" }),
      "获取设置失败",
    );

    if (!result.success) {
      return result;
    }

    return {
      success: true,
      data: normalizeInspectorSettings(result.data),
      message: "",
    };
  }

  async function updateSettings(token, payload) {
    const result = await request(
      "/inspector/settings/update",
      withAuth(token, {
        method: "POST",
        body: {
          notify_url: payload.notifyUrl || null,
          bg_url: payload.bgUrl || null,
        },
      }),
      "保存设置失败",
    );

    if (!result.success) {
      return result;
    }

    return {
      success: true,
      data: normalizeInspectorSettings(result.data),
      message: "",
    };
  }

  return {
    getDefaultInspectorQuery,
    listHosts,
    queryData,
    loadDashboard,
    createHost,
    updateHost,
    deleteHost,
    getSettings,
    updateSettings,
  };
}
