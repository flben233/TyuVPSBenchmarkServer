<script setup>
import { APPRISE_DOCS_URL, NOTIFY_PRESETS } from "~/utils/inspector";

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  settings: {
    type: Object,
    default: () => ({ notifyUrl: "", bgUrl: "" }),
  },
  submitting: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["update:modelValue", "save"]);

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit("update:modelValue", value),
});

const form = ref({
  notifyUrl: "",
  bgUrl: "",
});

const selectedPreset = ref("");
const builderVisible = ref(false);

watch(
  () => [props.settings, props.modelValue],
  () => {
    if (!props.modelValue) {
      return;
    }

    form.value = {
      notifyUrl: props.settings?.notifyUrl || "",
      bgUrl: props.settings?.bgUrl || "",
    };
    selectedPreset.value = "";
  },
  { immediate: true },
);

function openBuilder() {
  if (!selectedPreset.value) {
    return;
  }

  builderVisible.value = true;
}

function handleBuilderApply(url) {
  form.value.notifyUrl = url;
}

function handleSave() {
  emit("save", {
    notifyUrl: form.value.notifyUrl.trim(),
    bgUrl: form.value.bgUrl.trim(),
  });
}
</script>

<template>
  <el-dialog v-model="dialogVisible" title="Inspector 设置" width="640px" destroy-on-close>
    <el-form label-position="top">
      <el-form-item>
        <div class="settings-header">
          <span class="settings-title">消息通知 URL</span>
          <el-link :href="APPRISE_DOCS_URL" target="_blank" type="primary">查看 Apprise 文档</el-link>
        </div>
        <el-input
          v-model="form.notifyUrl"
          type="textarea"
          :rows="3"
          placeholder="直接粘贴 Apprise URL，例如 dingtalk://... 或 tgram://..."
        />
        <div class="settings-helper">
          如果不想手写 URL，可以先选择常见通知服务，再生成并回填到输入框。
        </div>
      </el-form-item>

      <el-form-item label="常见通知服务">
        <div class="preset-row">
          <el-select v-model="selectedPreset" placeholder="请选择通知服务" class="preset-select">
            <el-option
              v-for="preset in NOTIFY_PRESETS"
              :key="preset.value"
              :label="preset.label"
              :value="preset.value"
            />
          </el-select>
          <el-button :disabled="!selectedPreset" @click="openBuilder">生成 URL</el-button>
        </div>
      </el-form-item>

      <el-form-item label="页面背景图片 URL">
        <el-input
          v-model="form.bgUrl"
          placeholder="可填写一张公网图片地址，进入页面后会作为背景展示"
        />
        <div class="settings-helper">建议使用分辨率较高、对比度较低的图片，以保持卡片内容可读性。</div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSave">保存设置</el-button>
    </template>
  </el-dialog>

  <InspectorNotifyBuilderDialog
    v-model="builderVisible"
    :preset="selectedPreset"
    @apply="handleBuilderApply"
  />
</template>

<style scoped>
.settings-header {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.settings-title {
  color: #303133;
  font-weight: 600;
}

.settings-helper {
  margin-top: 8px;
  color: #909399;
  line-height: 1.6;
  font-size: 12px;
}

.preset-row {
  display: flex;
  gap: 12px;
  width: 100%;
}

.preset-select {
  flex: 1;
}

@media screen and (max-width: 768px) {
  .settings-header,
  .preset-row {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
