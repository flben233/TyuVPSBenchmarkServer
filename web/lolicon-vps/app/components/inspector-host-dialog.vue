<script setup>
import { stringifyTagList } from "~/utils/inspector";

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  mode: {
    type: String,
    default: "create",
  },
  host: {
    type: Object,
    default: null,
  },
  submitting: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["update:modelValue", "submit"]);
const { warn } = useMessage()

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit("update:modelValue", value),
});

const form = ref({
  name: "",
  target: "",
  tagsInput: "",
  notify: false,
  notifyTolerance: 0,
});

const isEditMode = computed(() => props.mode === "edit");

watch(
  () => [props.host, props.modelValue, props.mode],
  () => {
    if (!props.modelValue) {
      return;
    }

    form.value = {
      name: props.host?.name || "",
      target: props.host?.target || "",
      tagsInput: Array.isArray(props.host?.tags) ? props.host.tags.join(", ") : "",
      notify: Boolean(props.host?.notify),
      notifyTolerance: Math.max(0, Math.floor(Number(props.host?.notifyTolerance) || 0)),
    };
  },
  { immediate: true },
);

function handleSubmit() {
  if (!form.value.name.trim() || !form.value.target.trim()) {
    warn("请填写服务器名称和目标地址");
    return;
  }

  emit("submit", {
    name: form.value.name.trim(),
    target: form.value.target.trim(),
    tags: stringifyTagList(form.value.tagsInput),
    notify: Boolean(form.value.notify),
    notify_tolerance: Math.max(0, Math.floor(Number(form.value.notifyTolerance) || 0)),
  });
}
</script>

<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isEditMode ? '编辑服务器' : '添加服务器'"
    width="560px"
    destroy-on-close
  >
    <el-form label-position="top">
      <el-form-item label="服务器名称">
        <el-input v-model="form.name" placeholder="例如：Tokyo CN2" maxlength="64" />
      </el-form-item>
      <el-form-item label="目标地址">
        <el-input v-model="form.target" placeholder="请输入 IP 或域名" maxlength="128" />
      </el-form-item>
      <el-form-item label="标签">
        <el-input
          v-model="form.tagsInput"
          placeholder="使用逗号分隔，例如：日本, CN2, 512M"
          maxlength="256"
        />
      </el-form-item>
      <el-form-item label="离线通知">
        <el-switch v-model="form.notify" />
        <div class="field-tip">开启后，目标离线/恢复时将按当前用户设置发送通知。</div>
      </el-form-item>
      <el-form-item label="通知容错次数" v-if="form.notify">
        <el-input-number
          v-model="form.notifyTolerance"
          :min="0"
          :step="1"
          :precision="0"
          step-strictly
          controls-position="right"
        />
        <div class="field-tip">0 表示立即通知；大于 0 表示连续异常达到该次数后再通知。</div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">
        {{ isEditMode ? '保存' : '创建' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.field-tip {
  margin-left: 8px;
  color: #909399;
  font-size: 12px;
}
</style>
