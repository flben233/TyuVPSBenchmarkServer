<script setup>
import { ref, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";

const props = defineProps({
  token: {
    type: String,
    required: true,
  },
});

const { listUserGroups, createUserGroup, updateUserGroup, deleteUserGroup } = useAdmin();

const loading = ref(false);
const groups = ref([]);
const dialogVisible = ref(false);
const isEditing = ref(false);
const groupForm = ref({
  id: 0,
  name: "",
  is_admin: false,
  inspector_num: 0,
  max_host_num: 0,
});

onMounted(async () => {
  await loadData();
});

async function loadData() {
  if (!props.token) return;
  loading.value = true;
  try {
    groups.value = await listUserGroups(props.token);
  } catch (error) {
    ElMessage.error("加载用户组数据失败");
  } finally {
    loading.value = false;
  }
}

function handleAdd() {
  isEditing.value = false;
  groupForm.value = {
    id: 0,
    name: "",
    is_admin: false,
    inspector_num: 0,
    max_host_num: 0,
  };
  dialogVisible.value = true;
}

function handleEdit(row) {
  isEditing.value = true;
  groupForm.value = {
    id: row.id,
    name: row.name,
    is_admin: row.is_admin,
    inspector_num: row.inspector_num,
    max_host_num: row.max_host_num,
  };
  dialogVisible.value = true;
}

async function handleSubmit() {
  if (!groupForm.value.name) {
    ElMessage.warning("用户组名称不能为空");
    return;
  }

  loading.value = true;
  try {
    let result;
    if (isEditing.value) {
      result = await updateUserGroup(props.token, groupForm.value);
    } else {
      result = await createUserGroup(props.token, groupForm.value);
    }

    if (result.success) {
      ElMessage.success(isEditing.value ? "用户组更新成功" : "用户组创建成功");
      dialogVisible.value = false;
      await loadData();
    } else {
      ElMessage.error(
        result.message || (isEditing.value ? "更新用户组失败" : "创建用户组失败")
      );
    }
  } catch (error) {
    ElMessage.error(isEditing.value ? "更新用户组失败" : "创建用户组失败");
  } finally {
    loading.value = false;
  }
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户组 "${row.name}" 吗？此操作不可恢复。`,
      "警告",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
      }
    );

    loading.value = true;
    const result = await deleteUserGroup(props.token, row.id);
    if (result.success) {
      ElMessage.success("用户组删除成功");
      await loadData();
    } else {
      ElMessage.error(result.message || "删除用户组失败");
    }
  } catch (error) {
    if (error !== "cancel") {
      ElMessage.error("删除用户组失败");
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
          <span>用户组管理</span>
          <el-button type="primary" size="small" @click="handleAdd">
            添加用户组
          </el-button>
        </div>
      </template>

      <el-empty
        v-if="groups.length === 0 && !loading"
        description="暂无用户组"
      />
      <el-table v-else :data="groups" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" min-width="150" />
        <el-table-column label="管理员" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_admin ? 'danger' : 'info'" effect="dark">
              {{ row.is_admin ? "是" : "否" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="inspector_num"
          label="审核员数量"
          width="120"
        />
        <el-table-column
          prop="max_host_num"
          label="最大主机数"
          width="120"
        />
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
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Add/Edit Group Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑用户组' : '添加用户组'"
      width="500px"
    >
      <el-form :model="groupForm" label-width="120px">
        <el-form-item label="ID" v-if="isEditing">
          <el-input :model-value="groupForm.id" disabled />
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="groupForm.name" placeholder="输入用户组名称" />
        </el-form-item>
        <el-form-item label="管理员权限">
          <el-switch v-model="groupForm.is_admin" />
        </el-form-item>
        <el-form-item label="最大探针数">
          <el-input-number
            v-model="groupForm.inspector_num"
            :min="0"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="最大主机数">
          <el-input-number
            v-model="groupForm.max_host_num"
            :min="0"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="loading">
          {{ isEditing ? "保存" : "添加" }}
        </el-button>
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
