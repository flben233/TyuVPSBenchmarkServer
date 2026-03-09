<script setup>
import { ElMessage } from "element-plus";
import { buildAppriseUrl, NOTIFY_PRESET_FIELDS } from "~/utils/inspector";

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  preset: {
    type: String,
    default: "",
  },
});

const emit = defineEmits(["update:modelValue", "apply"]);

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit("update:modelValue", value),
});

const form = ref({});

const presetConfig = computed(() => NOTIFY_PRESET_FIELDS[props.preset] || null);

watch(
  () => props.preset,
  (preset) => {
    const fields = NOTIFY_PRESET_FIELDS[preset]?.fields || [];
    form.value = fields.reduce((result, field) => {
      result[field.key] = "";
      return result;
    }, {});
  },
  { immediate: true },
);

function handleBuild() {
  const fields = presetConfig.value?.fields || [];
  const missingField = fields.find((field) => field.required && !String(form.value[field.key] || "").trim());
  if (missingField) {
    ElMessage.warning(`请填写${missingField.label}`);
    return;
  }

  const url = buildAppriseUrl(props.preset, form.value);
  if (!url) {
    ElMessage.error("生成通知 URL 失败，请检查填写内容");
    return;
  }

  emit("apply", url);
  dialogVisible.value = false;
}
</script>

<template>
  <el-dialog v-model="dialogVisible" :title="presetConfig?.label || '生成通知 URL'" width="560px" destroy-on-close>
    <template v-if="presetConfig">
      <p class="preset-description">{{ presetConfig.description }}</p>
      <el-form label-position="top">
        <el-form-item
          v-for="field in presetConfig.fields"
          :key="field.key"
          :label="field.label"
          :required="field.required"
        >
          <el-input
            v-model="form[field.key]"
            :type="field.type || 'text'"
            :placeholder="field.placeholder"
            :show-password="field.type === 'password'"
          />
        </el-form-item>
      </el-form>
    </template>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="handleBuild">生成 URL</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.preset-description {
  margin: 0 0 16px;
  color: #606266;
  line-height: 1.6;
}
</style>
