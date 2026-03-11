<script setup>
import { ref, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";

const props = defineProps({
  token: {
    type: String,
    required: true,
  },
});

const { listUsers, updateUser, deleteUser, listUserGroups } = useAdmin();
const { warn, err, success } = useMessage()

const loading = ref(false);
const users = ref([]);
const groups = ref([]);
const editDialogVisible = ref(false);
const editForm = ref({
  id: 0,
  name: "",
  login: "",
  group_id: 0,
});

onMounted(async () => {
  await loadData();
});

async function loadData() {
  if (!props.token) return;
  loading.value = true;
  try {
    const [userList, groupList] = await Promise.all([
      listUsers(),
      listUserGroups(),
    ]);
    users.value = userList;
    groups.value = groupList;
  } catch (error) {
    err("加载用户数据失败");
  } finally {
    loading.value = false;
  }
}

function getGroupName(groupId) {
  const group = groups.value.find((g) => g.id === groupId);
  return group ? group.name : "未分组";
}

function handleEdit(row) {
  editForm.value = {
    id: row.id,
    name: row.name,
    login: row.login,
    group_id: row.group_id,
  };
  editDialogVisible.value = true;
}

async function handleUpdateUser() {
  if (!editForm.value.name) {
    warn("用户名称不能为空");
    return;
  }

  loading.value = true;
  try {
    const result = await updateUser(editForm.value);
    if (result.success) {
      success("用户更新成功");
      editDialogVisible.value = false;
      await loadData();
    } else {
      err(result.message || "更新用户失败");
    }
  } catch (error) {
    err("更新用户失败");
  } finally {
    loading.value = false;
  }
}

async function handleDeleteUser(row) {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${row.name}" 吗？此操作不可恢复。`,
      "警告",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
      }
    );

    loading.value = true;
    const result = await deleteUser(row.id);
    if (result.success) {
      success("用户删除成功");
      await loadData();
    } else {
      err(result.message || "删除用户失败");
    }
  } catch (error) {
    if (error !== "cancel") {
      err("删除用户失败");
    }
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div>
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>用户管理</span>
          <el-button size="small" @click="loadData" :loading="loading">
            刷新
          </el-button>
        </div>
      </template>

      <el-empty
        v-if="users.length === 0 && !loading"
        description="暂无用户"
      />
      <el-table v-else :data="users" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="120" />
        <el-table-column prop="name" label="名称" min-width="150" />
        <el-table-column prop="login" label="登录名" min-width="150" />
        <el-table-column label="用户组" min-width="120">
          <template #default="{ row }">
            {{ getGroupName(row.group_id) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              @click="handleEdit(row)"
            >
              编辑
            </el-button>
            <el-button
              type="danger"
              size="small"
              @click="handleDeleteUser(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Edit User Dialog -->
    <el-dialog
      v-model="editDialogVisible"
      title="编辑用户"
      width="500px"
    >
      <el-form :model="editForm" label-width="100px">
        <el-form-item label="ID">
          <el-input :model-value="editForm.id" disabled />
        </el-form-item>
        <el-form-item label="用户组">
          <el-select
            v-model="editForm.group_id"
            placeholder="选择用户组"
            style="width: 100%"
          >
            <el-option
              v-for="group in groups"
              :key="group.id"
              :label="group.name"
              :value="group.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleUpdateUser" :loading="loading"
          >保存</el-button
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
