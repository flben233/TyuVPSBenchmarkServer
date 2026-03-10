import {
  getDefaultInspectorQuery,
  mergeHostData,
  normalizeInspectorSettings,
} from "~/utils/inspector";
import {requestWithAuth} from "~/composables/useAuth.js";

export function useInspector() {
  function resolveErrorMessage(error, fallbackMessage) {
    return error?.data?.message || error?.message || fallbackMessage;
  }

  async function request(path, method, fallbackMessage = "请求失败", options = {}) {
    try {
      const response = await requestWithAuth(path, method, options);

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

  async function listHosts(token) {
    return request(
      "/inspector/hosts",
      "GET",
      "获取服务器列表失败",
    );
  }

  async function queryData(token, query = getDefaultInspectorQuery()) {
    return request(
      "/inspector/data",
      "GET",
      "获取探针数据失败",
      { query }
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
        "POST",
      "创建服务器失败",
      { body: payload }
    );
  }

  async function updateHost(token, id, payload) {
    return request(
      `/inspector/hosts/update/${id}`,
"POST",
      "更新服务器失败",
      { body: payload }
    );
  }

  async function deleteHost(token, id) {
    return request(
      `/inspector/hosts/delete/${id}`,
      "POST",
      "删除服务器失败"
    );
  }

  async function getSettings(token) {
    const result = await request(
      "/inspector/settings",
      "GET",
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
      "POST",
      "保存设置失败",
      {
          body: {
            notify_url: payload.notifyUrl || null,
            bg_url: payload.bgUrl || null,
          },
        }
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
