import {requestWithAuth} from "~/composables/useAuth.js";

export function useReport() {
  const { backendUrl } = useAppConfig();

  async function pollTask(taskId, fallbackMessage = "任务执行失败") {
    const maxAttempts = 120;
    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      const resp = await requestWithAuth(`/task/status/${taskId}`, "GET");
      if (!resp || resp.code !== 0) {
        return { success: false, message: resp?.message || fallbackMessage };
      }

      const task = resp.data;
      if (task?.status === "done") {
        return { success: true, data: task.result, task };
      }

      await new Promise((resolve) => setTimeout(resolve, 1000));
    }

    return { success: false, message: "任务执行超时，请稍后重试" };
  }

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

  async function addReport(html, monitor_id, other_info) {
    try {
      const resp = await requestWithAuth(`/report/admin/add`, "POST", {
        body: JSON.stringify([{
          html: html,
          monitor_id: monitor_id,
          other_info: other_info || "",
        }]),
      });

      if (resp && resp.code === 0 && resp.data?.id) {
        const taskResult = await pollTask(resp.data.id, "报告导入失败");
        if (!taskResult.success) {
          return taskResult;
        }

        const failed = Array.isArray(taskResult.data?.failed) ? taskResult.data.failed : [];
        if (failed.length > 0) {
          return {
            success: false,
            data: taskResult.data,
            message: "报告解析失败",
          };
        }

        return { success: true, data: { id: resp.data.id, task: taskResult.task } };
      }
      return { success: false, message: resp?.message || "Failed to add report" };
    } catch (error) {
      console.error("Failed to add report:", error);
      return { success: false, message: error.message };
    }
  }

  async function deleteReport(id) {
    try {
      const resp = await requestWithAuth(`/report/admin/delete`, "POST", {
        body: JSON.stringify({ id })
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

  async function updateReport(id, monitorId, otherInfo) {
    try {
      const resp = await requestWithAuth(`/report/admin/update`, "POST", {
        body: JSON.stringify({ id, monitor_id: monitorId, other_info: otherInfo || "" })
      });

      if (resp && resp.code === 0) {
        return { success: true };
      }
      return { success: false, message: resp?.message || "Failed to update report" };
    } catch (error) {
      console.error("Failed to update report:", error);
      return { success: false, message: error.message };
    }
  }

  return {
    listReports,
    getReportDetails,
    addReport,
    deleteReport,
    updateReport,
  };
}
