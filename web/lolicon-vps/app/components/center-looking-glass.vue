<script setup>
import { ref, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";

const props = defineProps({
  token: {
    type: String,
    required: true,
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
});

const {
  listRecords,
  addRecord,
  removeRecord,
  listPendingRecords,
  approveRecord,
  rejectRecord,
} = useLookingGlass();

const loading = ref(false);
const records = ref([]);
const pendingRecords = ref([]);
const addRecordForm = ref({
  serverName: "",
  testUrl: "",
});
const addRecordDialogVisible = ref(false);

onMounted(async () => {
  await loadData();
});

async function loadData() {
  if (!props.token) return;
  loading.value = true;
  try {
    records.value = await listRecords(props.token);
    if (props.isAdmin) {
      pendingRecords.value = await listPendingRecords(props.token);
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
      props.token,
      addRecordForm.value.serverName,
      addRecordForm.value.testUrl
    );
    if (result.success) {
      ElMessage.success("记录添加成功");
      addRecordDialogVisible.value = false;
      addRecordForm.value = { serverName: "", testUrl: "" };
      await loadData();
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
    const result = await removeRecord(props.token, id);
    if (result.success) {
      ElMessage.success("记录删除成功");
      await loadData();
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
    const result = await approveRecord(props.token, id);
    if (result.success) {
      ElMessage.success("记录审核通过");
      await loadData();
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
    const result = await rejectRecord(props.token, id);
    if (result.success) {
      ElMessage.success("记录已拒绝");
      await loadData();
    } else {
      ElMessage.error(result.message || "拒绝记录失败");
    }
  } catch (error) {
    ElMessage.error("拒绝记录失败");
  } finally {
    loading.value = false;
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
  <div>
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
            <el-tag :type="getReviewStatusType(row.review_status)" effect="dark">
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
  </div>
</template>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
