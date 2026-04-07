<script setup>
import { savePassword, getPassword, clearPassword } from "~/utils/webssh-storage";

const props = defineProps({
  modelValue: Boolean,
  mode: {
    type: String,
    default: "set", // set | change | reset
  },
});

const emit = defineEmits(["update:modelValue", "saved", "reset"]);

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit("update:modelValue", val),
});

const form = ref({
  oldPassword: "",
  newPassword: "",
  confirmPassword: "",
  saveLocally: true,
});

const errorMsg = ref("");

const isResetMode = ref(props.mode === "reset");

watch(
  () => props.modelValue,
  (val) => {
    if (val) {
      errorMsg.value = "";
      isResetMode.value = props.mode === "reset";
      if (props.mode === "change" && !isResetMode.value) {
        form.value.oldPassword = getPassword() || "";
      }
      form.value.newPassword = "";
      form.value.confirmPassword = "";
    }
  }
);

const dialogTitle = computed(() => {
  if (isResetMode.value) return "重置密钥";
  if (props.mode === "set") return "设置密钥";
  return "修改密钥";
});

const effectiveMode = computed(() => {
  if (isResetMode.value) return "reset";
  return props.mode;
});

function validate() {
  errorMsg.value = "";

  if (!isResetMode.value && props.mode === "change" && !form.value.oldPassword) {
    errorMsg.value = "请输入当前密钥";
    return false;
  }

  if (!isResetMode.value && (!form.value.newPassword || form.value.newPassword.length < 8)) {
    errorMsg.value = "密钥至少 8 位";
    return false;
  }

  if (!isResetMode.value && form.value.newPassword !== form.value.confirmPassword) {
    errorMsg.value = "两次输入的密钥不一致";
    return false;
  }

  return true;
}

async function handleConfirm() {
  if (!validate()) return;

  if (isResetMode.value) {
    emit("reset");
    visible.value = false;
    return;
  }

  savePassword(form.value.newPassword);
  emit("saved", form.value.newPassword);
  visible.value = false;
}
</script>

<template>
  <el-dialog
    v-model="visible"
    :title="dialogTitle"
    width="420px"
    :close-on-click-modal="false"
  >
    <el-form label-width="100px" label-position="left">
      <el-form-item v-if="effectiveMode === 'change'" label="当前密钥">
        <el-input
          v-model="form.oldPassword"
          type="password"
          show-password
          placeholder="输入当前密钥"
        />
      </el-form-item>

      <el-form-item v-if="effectiveMode !== 'reset'" label="新密钥" required>
        <el-input
          v-model="form.newPassword"
          type="password"
          show-password
          placeholder="至少 8 位"
        />
      </el-form-item>

      <el-form-item v-if="effectiveMode !== 'reset'" label="确认新密钥" required>
        <el-input
          v-model="form.confirmPassword"
          type="password"
          show-password
          placeholder="再次输入新密钥"
        />
      </el-form-item>

      <el-form-item v-if="effectiveMode !== 'reset'" label="本地保存">
        <el-switch v-model="form.saveLocally" />
        <span class="hint-text">保存后可直接使用，无需重复输入</span>
      </el-form-item>

      <el-form-item v-if="props.mode !== 'set'" label="模式">
        <el-switch
          v-model="isResetMode"
          active-text="重置密钥"
          inactive-text="修改密钥"
          inline-prompt
          @change="form.newPassword = ''; form.confirmPassword = ''"
        />
      </el-form-item>

      <el-alert
        v-if="effectiveMode === 'reset'"
        type="error"
        :closable="false"
        show-icon
        title="此操作将永久删除云端所有加密连接数据，不可恢复"
      />

      <el-alert v-if="errorMsg" type="error" :closable="false" :title="errorMsg" />
    </el-form>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button
        v-if="effectiveMode === 'reset'"
        type="danger"
        @click="handleConfirm"
      >
        确认重置
      </el-button>
      <el-button
        v-else
        type="primary"
        @click="handleConfirm"
      >
        确认
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.hint-text {
  font-size: 12px;
  color: #909399;
  margin-left: 8px;
}
</style>
