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
  listHosts,
  addHost,
  removeHost,
  listPendingHosts,
  approveHost,
  rejectHost,
} = useMonitor();
const { warn, err, success } = useMessage()

const loading = ref(false);
const hosts = ref([]);
const pendingHosts = ref([]);
const addHostForm = ref({
  name: "",
  target: "",
});
const addHostDialogVisible = ref(false);

defineExpose({ hosts });

onMounted(async () => {
  await loadData();
});

async function loadData() {
  if (!props.token) return;
  loading.value = true;
  try {
    hosts.value = await listHosts();
    if (props.isAdmin) {
      pendingHosts.value = await listPendingHosts();
    }
  } catch (error) {
    err("加载监控数据失败");
  } finally {
    loading.value = false;
  }
}

async function handleAddHost() {
  if (!addHostForm.value.name || !addHostForm.value.target) {
    warn("请填写所有字段");
    return;
  }

  loading.value = true;
  try {
    const result = await addHost(
      addHostForm.value.name,
      addHostForm.value.target
    );
    if (result.success) {
      success("主机添加成功");
      addHostDialogVisible.value = false;
      addHostForm.value = { name: "", target: "" };
      await loadData();
    } else {
      err(result.message || "添加主机失败");
    }
  } catch (error) {
    err("添加主机失败");
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
    const result = await removeHost(id);
    if (result.success) {
      success("主机删除成功");
      await loadData();
    } else {
      err(result.message || "删除主机失败");
    }
  } catch (error) {
    if (error !== "cancel") {
      err("删除主机失败");
    }
  } finally {
    loading.value = false;
  }
}

async function handleApproveHost(id) {
  loading.value = true;
  try {
    const result = await approveHost(id);
    if (result.success) {
      success("主机审核通过");
      await loadData();
    } else {
      err(result.message || "审核主机失败");
    }
  } catch (error) {
    err("审核主机失败");
  } finally {
    loading.value = false;
  }
}

async function handleRejectHost(id) {
  loading.value = true;
  try {
    const result = await rejectHost(id);
    if (result.success) {
      success("主机已拒绝");
      await loadData();
    } else {
      err(result.message || "拒绝主机失败");
    }
  } catch (error) {
    err("拒绝主机失败");
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
  </div>
</template>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
