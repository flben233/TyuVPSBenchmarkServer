import {
  getDefaultInspectorQuery,
  mergeHostData,
  normalizeInspectorSettings,
  normalizeVisitorPage,
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
      console.error(error)
      return {
        success: false,
        data: null,
        message: resolveErrorMessage(error, fallbackMessage),
      };
    }
  }

  async function publicRequest(path, method, fallbackMessage = "请求失败", options = {}) {
    const { backendUrl } = useAppConfig();

    try {
      const response = await $fetch(`${backendUrl}${path}`, {
        method,
        ...options,
      });

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
      console.error(error);
      return {
        success: false,
        data: null,
        message: resolveErrorMessage(error, fallbackMessage),
      };
    }
  }

  async function listHosts() {
    return request(
      "/inspector/hosts",
      "GET",
      "获取服务器列表失败",
    );
  }

  async function queryData(query = getDefaultInspectorQuery()) {
    return request(
      "/inspector/data",
      "GET",
      "获取探针数据失败",
      { query }
    );
  }

  async function loadDashboard(query = getDefaultInspectorQuery()) {
    const [hostsResult, dataResult] = await Promise.all([listHosts(), queryData(query)]);
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

  async function createHost(payload) {
    return request(
      "/inspector/hosts/create",
        "POST",
      "创建服务器失败",
      { body: {
        target: payload.target,
        monitor_type: payload.monitor_type,
        name: payload.name,
        tags: payload.tags,
        notify: Boolean(payload.notify),
        notify_tolerance: payload.notify_tolerance,
        traffic_settlement_day: payload.traffic_settlement_day || 0,
        monthly_traffic_limit: payload.monthly_traffic_limit || 0,
      } }
    );
  }

  async function updateHost(id, payload) {
    return request(
      `/inspector/hosts/update/${id}`,
"POST",
      "更新服务器失败",
      { body: {
        name: payload.name,
        tags: payload.tags,
        target: payload.target,
        monitor_type: payload.monitor_type,
        notify: Boolean(payload.notify),
        notify_tolerance: payload.notify_tolerance,
        traffic_settlement_day: payload.traffic_settlement_day || 0,
        monthly_traffic_limit: payload.monthly_traffic_limit || 0,
      } }
    );
  }

  async function deleteHost(id) {
    return request(
      `/inspector/hosts/delete/${id}`,
      "POST",
      "删除服务器失败"
    );
  }

  async function getSettings() {
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

  async function updateSettings(payload) {
    const result = await request(
      "/inspector/settings/update",
      "POST",
      "保存设置失败",
      {
          body: {
            notify_url: payload.notifyUrl || null,
            bg_url: payload.bgUrl || null,
            visitor_enabled: Boolean(payload.visitorEnabled),
            allowed_host_ids: Array.isArray(payload.allowedHostIds)
              ? payload.allowedHostIds.map((id) => String(id))
              : [],
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

  async function testNotify(notifyUrl) {
    return request(
      "/inspector/notify/test",
      "POST",
      "测试通知失败",
      { body: { notify_url: notifyUrl } },
    );
  }

  async function getVisitorPage(owner, query = getDefaultInspectorQuery()) {
    const result = await publicRequest(
      "/inspector/visitor/" + owner,
      "GET",
      "获取访客页失败",
      {
        query: {
          ...query,
        },
      },
    );

    if (!result.success) {
      return result;
    }

    return {
      success: true,
      data: normalizeVisitorPage(result.data),
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
    testNotify,
    getVisitorPage,
  };
}
