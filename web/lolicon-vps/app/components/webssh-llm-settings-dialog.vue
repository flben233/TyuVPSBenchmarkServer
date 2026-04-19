<script setup>
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  settings: {
    type: Object,
    default: () => ({
      enabled: false,
      apiBase: "",
      apiKey: "",
      model: "",
    }),
  },
  allowedCommands: {
    type: Array,
    default: () => [],
  },
  submitting: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["update:modelValue", "save", "save-whitelist"]);

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit("update:modelValue", v),
});

const form = ref({
  enabled: false,
  apiBase: "",
  apiKey: "",
  model: "",
});

const whitelist = ref([]);
const newCommand = ref("");

watch(
  () => [props.settings, props.allowedCommands],
  () => {
    form.value = {
      enabled: !!props.settings?.enabled,
      apiBase: props.settings?.apiBase || "",
      apiKey: props.settings?.apiKey || "",
      model: props.settings?.model || "",
    };
    whitelist.value = [...(props.allowedCommands || [])];
  },
  { immediate: true, deep: true }
);

function handleSave() {
  emit("save", {
    enabled: !!form.value.enabled,
    apiBase: form.value.apiBase?.trim() || "",
    apiKey: form.value.apiKey?.trim() || "",
    model: form.value.model?.trim() || "",
  });
  emit("save-whitelist", whitelist.value.filter((c) => c.trim()));
}

function addCommand() {
  const cmd = newCommand.value.trim().toLowerCase();
  if (!cmd) return;
  if (!whitelist.value.includes(cmd)) {
    whitelist.value.push(cmd);
  }
  newCommand.value = "";
}

function removeCommand(cmd) {
  whitelist.value = whitelist.value.filter((c) => c !== cmd);
}

function handleCommandKeydown(e) {
  if (e.key === "Enter") {
    e.preventDefault();
    addCommand();
  }
}
</script>

<template>
  <el-dialog v-model="visible" title="LLM 设置" width="600px" :close-on-click-modal="false">
    <el-form label-width="120px">
      <el-divider content-position="left">LLM API</el-divider>
      <el-form-item label="启用自定义 API">
        <el-switch v-model="form.enabled" />
      </el-form-item>
      <el-form-item label="API Base" required>
        <el-input v-model="form.apiBase" placeholder="https://api.example.com/v1" :disabled="!form.enabled" />
      </el-form-item>
      <el-form-item label="API Key" required>
        <el-input v-model="form.apiKey" placeholder="sk-..." show-password :disabled="!form.enabled" />
      </el-form-item>
      <el-form-item label="Model" required>
        <el-input v-model="form.model" placeholder="gpt-4o-mini" :disabled="!form.enabled" />
      </el-form-item>
      <el-alert
        title='关闭后将使用免费模型。点击连接列表的 ↑ 按钮会同时上传API设置，纯前端AES加密后存储在云端，服务器不保存明文。'
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 16px"
      />

      <el-divider content-position="left">命令白名单</el-divider>
      <el-form-item label="允许的命令">
        <div class="whitelist-tags">
          <el-tag
            v-for="cmd in whitelist"
            :key="cmd"
            closable
            @close="removeCommand(cmd)"
            class="whitelist-tag"
          >
            {{ cmd }}
          </el-tag>
          <el-input
            v-model="newCommand"
            placeholder="输入命令名"
            size="small"
            class="whitelist-input"
            @keydown="handleCommandKeydown"
          >
            <template #append>
              <el-button @click="addCommand" :disabled="!newCommand.trim()">+</el-button>
            </template>
          </el-input>
        </div>
      </el-form-item>
      <el-alert
        title="默认已允许只读命令（ls、cat、grep 等），此处添加的是额外允许的命令。控制粒度为命令级别，例如添加 rm 将允许 rm 的任何参数。"
        type="info"
        :closable="false"
        show-icon
      />
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSave">保存</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.whitelist-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
  width: 100%;
}

.whitelist-tag {
  font-family: 'Cascadia Mono', Consolas, monospace;
}

.whitelist-input {
  width: 200px;
}
</style>

<style>
.el-input-group__append {
  border-radius: 0 !important;
  box-shadow: 0 -1px 0 0 var(--el-input-border-color) inset,0 -1px 0 0 var(--el-input-border-color) inset,0px 0 0 0 var(--el-input-border-color) inset;
}
</style>
