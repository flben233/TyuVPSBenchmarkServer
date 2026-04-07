<script setup>
import { generateId } from "~/utils/webssh-storage";

const props = defineProps({
  modelValue: Boolean,
  editConnection: {
    type: Object,
    default: null,
  },
});

const emit = defineEmits(["update:modelValue", "save"]);

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit("update:modelValue", val),
});

const form = ref({
  name: "",
  host: "",
  port: 22,
  username: "root",
  authType: "password",
  password: "",
  privateKey: "",
});

watch(
  () => props.modelValue,
  (val) => {
    if (val && props.editConnection) {
      form.value = { ...props.editConnection };
    } else if (val) {
      form.value = {
        name: "",
        host: "",
        port: 22,
        username: "root",
        authType: "password",
        password: "",
        privateKey: "",
      };
    }
  }
);

function handleSave() {
  if (!form.value.name || !form.value.host || !form.value.username) {
    return;
  }

  const connection = {
    ...form.value,
    id: form.value.id || generateId(),
    port: Number(form.value.port) || 22,
    lastConnected: form.value.lastConnected || null,
  };

  emit("save", connection);
  visible.value = false;
}
</script>

<template>
  <el-dialog
    v-model="visible"
    :title="editConnection ? '编辑连接' : '新建连接'"
    width="480px"
    :close-on-click-modal="false"
  >
    <el-form label-width="90px" label-position="left">
      <el-form-item label="名称" required>
        <el-input v-model="form.name" placeholder="例如: My Server" />
      </el-form-item>
      <el-form-item label="主机" required>
        <el-input v-model="form.host" placeholder="IP 地址或域名" />
      </el-form-item>
      <el-form-item label="端口">
        <el-input-number v-model="form.port" :min="1" :max="65535" />
      </el-form-item>
      <el-form-item label="用户名" required>
        <el-input v-model="form.username" placeholder="例如: root" />
      </el-form-item>
      <el-form-item label="认证方式">
        <el-radio-group v-model="form.authType">
          <el-radio value="password">密码</el-radio>
          <el-radio value="privateKey">私钥</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item v-if="form.authType === 'password'" label="密码">
        <el-input
          v-model="form.password"
          type="password"
          show-password
          placeholder="SSH 密码"
        />
      </el-form-item>
      <el-form-item v-if="form.authType === 'privateKey'" label="私钥">
        <el-input
          v-model="form.privateKey"
          type="textarea"
          :rows="6"
          placeholder="粘贴 SSH 私钥内容"
        />
      </el-form-item>
      <el-form-item v-if="form.authType === 'privateKey'">
        <el-alert
          type="warning"
          :closable="false"
          show-icon
          title="私钥将保存在浏览器本地存储中，请确保设备安全"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleSave">
        {{ editConnection ? "保存" : "创建" }}
      </el-button>
    </template>
  </el-dialog>
</template>
