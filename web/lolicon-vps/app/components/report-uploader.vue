<script setup>
import { ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";

const props = defineProps({
  token: {
    type: String,
    required: true,
  },
  hosts: {
    type: Array,
    default: () => [],
  },
});

const { addReport, deleteReport } = useReport();

// Report data
const fileInput = useTemplateRef("report-selector");
const addReportDialogVisible = ref(false);
const reportFile = ref(null);
const selectedMonitorId = ref(null);
const reportFiles = ref([]);
const fileMonitorMapping = ref([]);
const uploadProgress = ref(0);
const uploadingBatch = ref(false);
const loading = ref(false);

// Report functions
async function handleFileSelect(event) {
  if (!event.target.files.length) {
    ElMessage.warning("请选择HTML文件");
    return;
  }

  const files = Array.from(event.target.files);

  if (files.length === 1) {
    // Single file upload
    reportFile.value = files[0];
    reportFiles.value = [];
    fileMonitorMapping.value = [];
  } else {
    // Multiple files upload
    reportFiles.value = files;
    reportFile.value = null;
    // Initialize monitor mapping for each file
    fileMonitorMapping.value = files.map((file) => ({
      file: file,
      monitorId: null,
    }));
  }

  // Reset file input
  event.target.value = "";
  // Open dialog to select monitor
  addReportDialogVisible.value = true;
}

async function handleAddReport() {
  if (!reportFile.value && reportFiles.value.length === 0) {
    ElMessage.warning("请选择HTML文件");
    return;
  }

  loading.value = true;
  uploadingBatch.value = reportFiles.value.length > 0;

  try {
    if (reportFiles.value.length > 0) {
      // Batch upload
      const totalFiles = reportFiles.value.length;
      let successCount = 0;
      let failedCount = 0;

      for (let i = 0; i < totalFiles; i++) {
        uploadProgress.value = Math.round(((i + 1) / totalFiles) * 100);

        try {
          const htmlContent = await reportFiles.value[i].text();
          const monitorId = fileMonitorMapping.value[i].monitorId;
          const result = await addReport(
            props.token,
            htmlContent,
            monitorId || undefined
          );

          if (result.success) {
            successCount++;
          } else {
            failedCount++;
            console.error(
              `Failed to upload ${reportFiles.value[i].name}:`,
              result.message
            );
          }
        } catch (error) {
          failedCount++;
          console.error(`Error uploading ${reportFiles.value[i].name}:`, error);
        }
      }

      if (successCount > 0) {
        ElMessage.success(
          `成功上传 ${successCount} 个报告${failedCount > 0 ? `，失败 ${failedCount} 个` : ""}`
        );
      } else {
        ElMessage.error("所有报告上传失败");
      }

      addReportDialogVisible.value = false;
      reportFiles.value = [];
      fileMonitorMapping.value = [];
      uploadProgress.value = 0;
    } else {
      // Single file upload
      const htmlContent = await reportFile.value.text();
      const result = await addReport(
        props.token,
        htmlContent,
        selectedMonitorId.value || undefined
      );
      if (result.success) {
        ElMessage.success(`报告添加成功。ID: ${result.data.id}`);
        addReportDialogVisible.value = false;
        reportFile.value = null;
        selectedMonitorId.value = null;
      } else {
        ElMessage.error(result.message || "添加报告失败");
      }
    }
  } catch (error) {
    ElMessage.error("添加报告失败");
  } finally {
    loading.value = false;
    uploadingBatch.value = false;
  }
}

async function handleDeleteReport() {
  try {
    const { value: reportId } = await ElMessageBox.prompt(
      "请输入报告 ID",
      "删除报告",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        inputPattern: /.+/,
        inputErrorMessage: "报告 ID 是必填项",
      }
    );

    loading.value = true;
    const result = await deleteReport(props.token, reportId);
    if (result.success) {
      ElMessage.success("报告删除成功");
    } else {
      ElMessage.error(result.message || "删除报告失败");
    }
  } catch (error) {
    if (error !== "cancel") {
      ElMessage.error("删除报告失败");
    }
  } finally {
    loading.value = false;
  }
}

function showFileSelector() {
  if (fileInput) {
    fileInput.value.click();
  }
}
</script>

<template>
  <div>
    <el-card shadow="never">
      <template #header>
        <span>报告管理（仅管理员）</span>
      </template>

      <el-space direction="vertical" style="width: 100%" :size="16">
        <input
          type="file"
          ref="report-selector"
          style="display: none"
          accept=".html"
          multiple
          @change="handleFileSelect"
        />
        <el-button
          type="primary"
          @click="showFileSelector"
          :loading="loading"
        >
          添加报告
        </el-button>
        <el-button type="danger" @click="handleDeleteReport">
          删除报告
        </el-button>
      </el-space>
    </el-card>

    <!-- Add Report Dialog -->
    <el-dialog v-model="addReportDialogVisible" title="添加报告" width="600px">
      <el-form label-width="120px">
        <!-- Single file upload -->
        <template v-if="reportFile">
          <el-form-item label="文件">
            <span>{{ reportFile.name }}</span>
          </el-form-item>
          <el-form-item label="关联监控主机">
            <el-select
              v-model="selectedMonitorId"
              placeholder="选择监控主机（可选）"
              clearable
              style="width: 100%"
            >
              <el-option
                v-for="host in hosts"
                :key="host.id"
                :label="host.name"
                :value="host.id"
              />
            </el-select>
          </el-form-item>
        </template>

        <!-- Batch upload -->
        <template v-else-if="fileMonitorMapping.length > 0">
          <el-form-item label="文件列表">
            <div style="width: 100%">
              <div style="margin-bottom: 8px; color: #606266">
                已选择 {{ fileMonitorMapping.length }} 个文件
              </div>
              <div
                style="
                  max-height: 400px;
                  overflow-y: auto;
                  border: 1px solid #dcdfe6;
                  border-radius: 4px;
                  padding: 12px;
                "
              >
                <div
                  v-for="(mapping, index) in fileMonitorMapping"
                  :key="index"
                  style="
                    margin-bottom: 16px;
                    padding-bottom: 16px;
                    border-bottom: 1px solid #f0f0f0;
                  "
                  :style="
                    index === fileMonitorMapping.length - 1
                      ? 'border-bottom: none; margin-bottom: 0; padding-bottom: 0;'
                      : ''
                  "
                >
                  <div
                    style="
                      font-size: 13px;
                      color: #303133;
                      margin-bottom: 8px;
                      font-weight: 500;
                    "
                  >
                    {{ index + 1 }}. {{ mapping.file.name }}
                  </div>
                  <el-select
                    v-model="mapping.monitorId"
                    placeholder="选择监控主机（可选）"
                    clearable
                    size="small"
                    style="width: 100%"
                  >
                    <el-option
                      v-for="host in hosts"
                      :key="host.id"
                      :label="host.name"
                      :value="host.id"
                    />
                  </el-select>
                </div>
              </div>
            </div>
          </el-form-item>
        </template>

        <template v-else>
          <el-form-item label="文件">
            <span>未选择文件</span>
          </el-form-item>
        </template>

        <el-form-item v-if="uploadingBatch" label="上传进度">
          <el-progress :percentage="uploadProgress" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addReportDialogVisible = false" :disabled="loading"
          >取消</el-button
        >
        <el-button type="primary" @click="handleAddReport" :loading="loading">
          {{
            reportFiles.length > 1
              ? `批量上传 (${reportFiles.length} 个文件)`
              : "添加"
          }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
</style>
