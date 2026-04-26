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
  monitorType: "ping",
  tagsInput: "",
  notify: false,
  notifyTolerance: 0,
  trafficSettlementDay: 0,
  monthlyTrafficLimit: 0,
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
      monitorType: props.host?.monitorType || "ping",
      tagsInput: Array.isArray(props.host?.tags) ? props.host.tags.join(", ") : "",
      notify: Boolean(props.host?.notify),
      notifyTolerance: Math.max(0, Math.floor(Number(props.host?.notifyTolerance) || 0)),
      trafficSettlementDay: Math.max(0, Math.min(31, Math.floor(Number(props.host?.trafficSettlementDay) || 0))),
      monthlyTrafficLimit: Math.max(0, (Number(props.host?.monthlyTrafficLimit) || 0) / 1024),
    };
  },
  { immediate: true },
);

function handleSubmit() {
  if (!form.value.name.trim() || !form.value.target.trim()) {
    warn("请填写服务器名称和目标地址");
    return;
  }

  if (form.value.monitorType === "tcp" && !form.value.target.includes(":")) {
    warn("TCP 监控目标请使用 host:port 格式");
    return;
  }

  if (form.value.monitorType === "http" && !/^https?:\/\//i.test(form.value.target.trim())) {
    warn("HTTPing 监控目标请使用 http:// 或 https:// 开头的完整 URL");
    return;
  }

  emit("submit", {
    name: form.value.name.trim(),
    target: form.value.target.trim(),
    monitor_type: form.value.monitorType,
    tags: stringifyTagList(form.value.tagsInput),
    notify: Boolean(form.value.notify),
    notify_tolerance: Math.max(0, Math.floor(Number(form.value.notifyTolerance) || 0)),
    traffic_settlement_day: Math.max(0, Math.min(31, Math.floor(Number(form.value.trafficSettlementDay) || 0))),
    monthly_traffic_limit: Math.max(0, (Number(form.value.monthlyTrafficLimit) || 0) * 1024),
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
        <el-input
          v-model="form.target"
          :placeholder="form.monitorType === 'http' ? '请输入完整 URL，例如 https://example.com' : form.monitorType === 'tcp' ? '请输入 host:port，例如 1.1.1.1:443' : '请输入 IP 或域名'"
          maxlength="256"
        />
      </el-form-item>
      <el-form-item label="监控类型">
        <el-radio-group v-model="form.monitorType">
          <el-radio-button label="ping">Ping</el-radio-button>
          <el-radio-button label="tcp">TCPing</el-radio-button>
          <el-radio-button label="http">HTTPing</el-radio-button>
        </el-radio-group>
        <div class="field-tip">Ping 使用 ICMP；TCPing 使用 TCP 建连；HTTPing 使用 HTTP 请求首包耗时。</div>
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
      <el-form-item label="每月结算日">
        <el-input-number
          v-model="form.trafficSettlementDay"
          :min="0"
          :max="31"
          :step="1"
          :precision="0"
          step-strictly
          controls-position="right"
          placeholder="不设置"
        />
        <div class="field-tip">0 表示不设置；设置后每月该日将作为流量统计的结算周期起点。</div>
      </el-form-item>
      <el-form-item label="每月流量上限">
        <div class="limit-input-row">
          <el-input-number
            v-model="form.monthlyTrafficLimit"
            :min="0"
            :step="0.1"
            :precision="1"
            controls-position="right"
            class="limit-input"
          />
          <span class="limit-unit">GB</span>
        </div>
        <div class="field-tip">0 表示不限制；设置后可在卡片上查看本周期已用流量。</div>
      </el-form-item>
      <el-form-item>
        <el-link type="primary" target="_blank" href="https://note.shirakawatyu.top/note/article/422">点我查看数据采集端部署说明</el-link>
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

.limit-input-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.limit-input {
  width: 200px;
}

.limit-unit {
  color: #909399;
  font-size: 13px;
  white-space: nowrap;
}
</style>
