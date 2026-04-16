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
  submitting: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["update:modelValue", "save"]);

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

watch(
  () => props.settings,
  (v) => {
    form.value = {
      enabled: !!v?.enabled,
      apiBase: v?.apiBase || "",
      apiKey: v?.apiKey || "",
      model: v?.model || "",
    };
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
}
</script>

<template>
  <el-dialog v-model="visible" title="LLM API 设置" width="560px" :close-on-click-modal="false">
    <el-form label-width="120px">
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
        title="关闭“启用自定义 API”后，将使用系统默认模型配置"
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
