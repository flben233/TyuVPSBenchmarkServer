<script setup>
import { ref, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";

const { token, isAdmin } = useAuth();
const {
  listHosts,
  addHost,
  removeHost,
  listPendingHosts,
  approveHost,
  rejectHost,
} = useMonitor();
const {
  listRecords,
  addRecord,
  removeRecord,
  listPendingRecords,
  approveRecord,
  rejectRecord,
} = useLookingGlass();
const { addReport, deleteReport } = useReport();

const activeTab = ref("monitor");
const loading = ref(true);

// Monitor data
const hosts = ref([]);
const pendingHosts = ref([]);
const addHostForm = ref({
  name: "",
  target: "",
});
const addHostDialogVisible = ref(false);

// Looking Glass data
const records = ref([]);
const pendingRecords = ref([]);
const addRecordForm = ref({
  serverName: "",
  testUrl: "",
});
const addRecordDialogVisible = ref(false);

// Report data
const fileInput = useTemplateRef("report-selector");
const addReportDialogVisible = ref(false);
const reportFile = ref(null);
const selectedMonitorId = ref(null);

// Load data on mount
onMounted(async () => {
  await loadMonitorData();
  await loadLookingGlassData();
});

// Monitor functions
async function loadMonitorData() {
  if (!token.value) return;
  loading.value = true;
  try {
    hosts.value = await listHosts(token.value);
    if (isAdmin.value) {
      pendingHosts.value = await listPendingHosts(token.value);
    }
  } catch (error) {
    ElMessage.error("加载监控数据失败");
  } finally {
    loading.value = false;
  }
}

async function handleAddHost() {
  if (!addHostForm.value.name || !addHostForm.value.target) {
    ElMessage.warning("请填写所有字段");
    return;
  }

  loading.value = true;
  try {
    const result = await addHost(
      token.value,
      addHostForm.value.name,
      addHostForm.value.target
    );
    if (result.success) {
      ElMessage.success("主机添加成功");
      addHostDialogVisible.value = false;
      addHostForm.value = { name: "", target: "" };
      await loadMonitorData();
    } else {
      ElMessage.error(result.message || "添加主机失败");
    }
  } catch (error) {
    ElMessage.error("添加主机失败");
  } finally {
    loading.value = false;
  }
}

async function handleRemoveHost(id) {
  try {
    await ElMessageBox.confirm("确定要删除此主机吗？", "警告", {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    });

    loading.value = true;
    const result = await removeHost(token.value, id);
    if (result.success) {
      ElMessage.success("主机删除成功");
      await loadMonitorData();
    } else {
      ElMessage.error(result.message || "删除主机失败");
    }
  } catch (error) {
    if (error !== "cancel") {
      ElMessage.error("删除主机失败");
    }
  } finally {
    loading.value = false;
  }
}

async function handleApproveHost(id) {
  loading.value = true;
  try {
    const result = await approveHost(token.value, id);
    if (result.success) {
      ElMessage.success("主机审核通过");
      await loadMonitorData();
    } else {
      ElMessage.error(result.message || "审核主机失败");
    }
  } catch (error) {
    ElMessage.error("审核主机失败");
  } finally {
    loading.value = false;
  }
}

async function handleRejectHost(id) {
  loading.value = true;
  try {
    const result = await rejectHost(token.value, id);
    if (result.success) {
      ElMessage.success("主机已拒绝");
      await loadMonitorData();
    } else {
      ElMessage.error(result.message || "拒绝主机失败");
    }
  } catch (error) {
    ElMessage.error("拒绝主机失败");
  } finally {
    loading.value = false;
  }
}

// Looking Glass functions
async function loadLookingGlassData() {
  if (!token.value) return;
  loading.value = true;
  try {
    records.value = await listRecords(token.value);
    if (isAdmin.value) {
      pendingRecords.value = await listPendingRecords(token.value);
    }
  } catch (error) {
    ElMessage.error("加载 Looking Glass 数据失败");
  } finally {
    loading.value = false;
  }
}

async function handleAddRecord() {
  if (!addRecordForm.value.serverName || !addRecordForm.value.testUrl) {
    ElMessage.warning("请填写所有字段");
    return;
  }

  loading.value = true;
  try {
    const result = await addRecord(
      token.value,
      addRecordForm.value.serverName,
      addRecordForm.value.testUrl
    );
    if (result.success) {
      ElMessage.success("记录添加成功");
      addRecordDialogVisible.value = false;
      addRecordForm.value = { serverName: "", testUrl: "" };
      await loadLookingGlassData();
    } else {
      ElMessage.error(result.message || "添加记录失败");
    }
  } catch (error) {
    ElMessage.error("添加记录失败");
  } finally {
    loading.value = false;
  }
}

async function handleRemoveRecord(id) {
  try {
    await ElMessageBox.confirm("确定要删除此记录吗？", "警告", {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    });

    loading.value = true;
    const result = await removeRecord(token.value, id);
    if (result.success) {
      ElMessage.success("记录删除成功");
      await loadLookingGlassData();
    } else {
      ElMessage.error(result.message || "删除记录失败");
    }
  } catch (error) {
    if (error !== "cancel") {
      ElMessage.error("删除记录失败");
    }
  } finally {
    loading.value = false;
  }
}

async function handleApproveRecord(id) {
  loading.value = true;
  try {
    const result = await approveRecord(token.value, id);
    if (result.success) {
      ElMessage.success("记录审核通过");
      await loadLookingGlassData();
    } else {
      ElMessage.error(result.message || "审核记录失败");
    }
  } catch (error) {
    ElMessage.error("审核记录失败");
  } finally {
    loading.value = false;
  }
}

async function handleRejectRecord(id) {
  loading.value = true;
  try {
    const result = await rejectRecord(token.value, id);
    if (result.success) {
      ElMessage.success("记录已拒绝");
      await loadLookingGlassData();
    } else {
      ElMessage.error(result.message || "拒绝记录失败");
    }
  } catch (error) {
    ElMessage.error("拒绝记录失败");
  } finally {
    loading.value = false;
  }
}

// Report functions
async function handleFileSelect(event) {
  if (!event.target.files.length) {
    ElMessage.warning("请选择HTML文件");
    return;
  }

  reportFile.value = event.target.files[0];
  // Reset file input
  event.target.value = "";
  // Open dialog to select monitor
  addReportDialogVisible.value = true;
}

async function handleAddReport() {
  if (!reportFile.value) {
    ElMessage.warning("请选择HTML文件");
    return;
  }

  loading.value = true;
  try {
    const htmlContent = await reportFile.value.text();
    const result = await addReport(
      token.value,
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
  } catch (error) {
    ElMessage.error("添加报告失败");
  } finally {
    loading.value = false;
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
    const result = await deleteReport(token.value, reportId);
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
    console.log(fileInput);
    fileInput.value.click();
  }
}

function getReviewStatusText(status) {
  switch (status) {
    case 0:
      return "待审核";
    case 1:
      return "已通过";
    case 2:
      return "已拒绝";
    default:
      return "未知";
  }
}

function getReviewStatusType(status) {
  switch (status) {
    case 0:
      return "warning";
    case 1:
      return "success";
    case 2:
      return "danger";
    default:
      return "info";
  }
}
</script>

<template>
  <el-row id="center-root">
    <el-col :span="24" id="content-area">
      <h1 class="page-title">个人中心</h1>
      <div v-if="!token && !loading" class="login-prompt">
        <el-empty description="请登录以访问个人中心">
          <el-button type="primary" @click="$router.push('/login')"
            >前往登录</el-button
          >
        </el-empty>
      </div>

      <div v-else>
        <el-tabs v-model="activeTab" class="center-tabs">
          <!-- Monitor Hosts Tab -->
          <el-tab-pane label="监控主机" name="monitor">
            <el-card shadow="never">
              <template #header>
                <div class="card-header">
                  <span>我的监控主机</span>
                  <el-button
                    type="primary"
                    size="small"
                    @click="addHostDialogVisible = true"
                  >
                    添加主机
                  </el-button>
                </div>
              </template>

              <el-empty
                v-if="hosts.length === 0 && !loading"
                description="暂无主机"
              />
              <el-table v-else :data="hosts" v-loading="loading" border stripe>
                <el-table-column prop="id" label="ID" width="80" />
                <el-table-column prop="name" label="名称" min-width="150" />
                <el-table-column prop="target" label="目标" min-width="200" />
                <el-table-column
                  prop="uploader_name"
                  label="上传者"
                  min-width="120"
                />
                <el-table-column prop="review_status" label="状态" width="120">
                  <template #default="{ row }">
                    <el-tag :type="getReviewStatusType(row.review_status)">
                      {{ getReviewStatusText(row.review_status) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="120">
                  <template #default="{ row }">
                    <el-button
                      type="danger"
                      size="small"
                      @click="handleRemoveHost(row.id)"
                    >
                      删除
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </el-card>

            <!-- Admin: Pending Hosts -->
            <el-card shadow="never" v-if="isAdmin" style="margin-top: 16px">
              <template #header>
                <span>待审核主机（管理员）</span>
              </template>
              <el-empty
                v-if="pendingHosts.length === 0 && !loading"
                description="暂无待审核主机"
              />
              <el-table
                v-else
                :data="pendingHosts"
                v-loading="loading"
                border
                stripe
              >
                <el-table-column prop="id" label="ID" width="80" />
                <el-table-column prop="name" label="名称" min-width="150" />
                <el-table-column prop="target" label="目标" min-width="200" />
                <el-table-column
                  prop="uploader_name"
                  label="上传者"
                  min-width="120"
                />
                <el-table-column label="操作" width="200">
                  <template #default="{ row }">
                    <el-button
                      type="success"
                      size="small"
                      @click="handleApproveHost(row.id)"
                    >
                      通过
                    </el-button>
                    <el-button
                      type="danger"
                      size="small"
                      @click="handleRejectHost(row.id)"
                    >
                      拒绝
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </el-card>
          </el-tab-pane>

          <!-- Looking Glass Tab -->
          <el-tab-pane label="Looking Glass" name="lookingglass">
            <el-card shadow="never">
              <template #header>
                <div class="card-header">
                  <span>我的 Looking Glass 记录</span>
                  <el-button
                    type="primary"
                    size="small"
                    @click="addRecordDialogVisible = true"
                  >
                    添加记录
                  </el-button>
                </div>
              </template>
              <el-empty
                v-if="records.length === 0 && !loading"
                description="暂无记录"
              />
              <el-table
                v-else
                :data="records"
                v-loading="loading"
                border
                stripe
              >
                <el-table-column prop="id" label="ID" width="80" />
                <el-table-column
                  prop="server_name"
                  label="服务器名称"
                  min-width="150"
                />
                <el-table-column
                  prop="test_url"
                  label="测试 URL"
                  min-width="200"
                >
                  <template #default="{ row }">
                    <el-link
                      :href="row.test_url"
                      target="_blank"
                      type="primary"
                    >
                      {{ row.test_url }}
                    </el-link>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="uploader_name"
                  label="上传者"
                  min-width="120"
                />
                <el-table-column prop="review_status" label="状态" width="120">
                  <template #default="{ row }">
                    <el-tag :type="getReviewStatusType(row.review_status)">
                      {{ getReviewStatusText(row.review_status) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="120">
                  <template #default="{ row }">
                    <el-button
                      type="danger"
                      size="small"
                      @click="handleRemoveRecord(row.id)"
                    >
                      删除
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </el-card>

            <!-- Admin: Pending Records -->
            <el-card shadow="never" v-if="isAdmin" style="margin-top: 16px">
              <template #header>
                <span>待审核记录（管理员）</span>
              </template>
              <el-empty
                v-if="pendingRecords.length === 0 && !loading"
                description="暂无待审核记录"
              />
              <el-table
                v-else
                :data="pendingRecords"
                v-loading="loading"
                border
                stripe
              >
                <el-table-column prop="id" label="ID" width="80" />
                <el-table-column
                  prop="server_name"
                  label="服务器名称"
                  min-width="150"
                />
                <el-table-column
                  prop="test_url"
                  label="测试 URL"
                  min-width="200"
                >
                  <template #default="{ row }">
                    <el-link
                      :href="row.test_url"
                      target="_blank"
                      type="primary"
                    >
                      {{ row.test_url }}
                    </el-link>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="uploader_name"
                  label="上传者"
                  min-width="120"
                />
                <el-table-column label="操作" width="200">
                  <template #default="{ row }">
                    <el-button
                      type="success"
                      size="small"
                      @click="handleApproveRecord(row.id)"
                    >
                      通过
                    </el-button>
                    <el-button
                      type="danger"
                      size="small"
                      @click="handleRejectRecord(row.id)"
                    >
                      拒绝
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </el-card>
          </el-tab-pane>

          <!-- Admin: Report Management Tab -->
          <el-tab-pane label="报告管理" name="report" v-if="isAdmin">
            <el-card shadow="never">
              <template #header>
                <span>报告管理（仅管理员）</span>
              </template>

              <el-space direction="vertical" style="width: 100%" :size="16">
                <input
                  type="file"
                  ref="report-selector"
                  style="display: none"
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
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-col>

    <!-- Add Host Dialog -->
    <el-dialog
      v-model="addHostDialogVisible"
      title="添加监控主机"
      width="500px"
    >
      <el-form :model="addHostForm" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="addHostForm.name" placeholder="输入主机名称" />
        </el-form-item>
        <el-form-item label="目标">
          <el-input
            v-model="addHostForm.target"
            placeholder="输入目标（IP 或域名）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addHostDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddHost" :loading="loading"
          >添加</el-button
        >
      </template>
    </el-dialog>

    <!-- Add Record Dialog -->
    <el-dialog
      v-model="addRecordDialogVisible"
      title="添加 Looking Glass 记录"
      width="500px"
    >
      <el-form :model="addRecordForm" label-width="120px">
        <el-form-item label="服务器名称">
          <el-input
            v-model="addRecordForm.serverName"
            placeholder="输入服务器名称"
          />
        </el-form-item>
        <el-form-item label="测试 URL">
          <el-input
            v-model="addRecordForm.testUrl"
            placeholder="输入测试 URL"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addRecordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddRecord" :loading="loading"
          >添加</el-button
        >
      </template>
    </el-dialog>

    <!-- Add Report Dialog -->
    <el-dialog v-model="addReportDialogVisible" title="添加报告" width="500px">
      <el-form label-width="120px">
        <el-form-item label="文件">
          <span>{{ reportFile?.name || "未选择文件" }}</span>
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
      </el-form>
      <template #footer>
        <el-button @click="addReportDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddReport" :loading="loading"
          >添加</el-button
        >
      </template>
    </el-dialog>
  </el-row>
</template>

<style scoped>
#center-root {
  width: 100%;
  padding: 16px;
  box-sizing: border-box;
  overflow-x: hidden;
}

#content-area {
  margin: 0 auto;
}

.page-title {
  font-size: 32px;
  font-weight: 300;
  margin: 0 0 24px 0;
  color: #303133;
}

.login-prompt {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}

.center-tabs {
  margin-top: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
