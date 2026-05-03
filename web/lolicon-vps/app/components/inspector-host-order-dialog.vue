<script setup>
import { ArrowDown, ArrowUp, Rank } from "@element-plus/icons-vue";

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  hosts: {
    type: Array,
    default: () => [],
  },
});

const emit = defineEmits(["update:modelValue", "saved"]);

const { updateHostOrder } = useInspector();
const { success, err } = useMessage();

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit("update:modelValue", value),
});

const orderedHosts = ref([]);
const saving = ref(false);
const draggingIndex = ref(-1);
const listRef = ref(null);

watch(
  () => [props.hosts, props.modelValue],
  () => {
    if (!props.modelValue) {
      return;
    }
    orderedHosts.value = [...props.hosts];
  },
  { immediate: true },
);

function getItemHeight() {
  const item = listRef.value?.querySelector(".order-item");
  return item ? item.offsetHeight + 1 : 49;
}

function getIndexFromY(clientY) {
  if (!listRef.value) return -1;
  const rect = listRef.value.getBoundingClientRect();
  const relativeY = clientY - rect.top + listRef.value.scrollTop;
  const index = Math.floor(relativeY / getItemHeight());
  return Math.max(0, Math.min(index, orderedHosts.value.length - 1));
}

function moveTo(fromIndex, toIndex) {
  if (fromIndex === toIndex) return;
  const item = orderedHosts.value.splice(fromIndex, 1)[0];
  orderedHosts.value.splice(toIndex, 0, item);
}

// Mouse events
function onMouseDown(e, index) {
  e.preventDefault();
  draggingIndex.value = index;

  const onMouseMove = (moveEvent) => {
    const newIndex = getIndexFromY(moveEvent.clientY);
    if (newIndex !== -1 && newIndex !== draggingIndex.value) {
      moveTo(draggingIndex.value, newIndex);
      draggingIndex.value = newIndex;
    }
  };

  const onMouseUp = () => {
    draggingIndex.value = -1;
    document.removeEventListener("mousemove", onMouseMove);
    document.removeEventListener("mouseup", onMouseUp);
  };

  document.addEventListener("mousemove", onMouseMove);
  document.addEventListener("mouseup", onMouseUp);
}

// Touch events
function onTouchStart(e, index) {
  draggingIndex.value = index;
}

function onTouchMove(e) {
  if (draggingIndex.value === -1) return;
  e.preventDefault();

  const newIndex = getIndexFromY(e.touches[0].clientY);
  if (newIndex !== -1 && newIndex !== draggingIndex.value) {
    moveTo(draggingIndex.value, newIndex);
    draggingIndex.value = newIndex;
  }
}

function onTouchEnd() {
  draggingIndex.value = -1;
}

function moveUp(index) {
  if (index <= 0) return;
  const temp = orderedHosts.value[index];
  orderedHosts.value[index] = orderedHosts.value[index - 1];
  orderedHosts.value[index - 1] = temp;
}

function moveDown(index) {
  if (index >= orderedHosts.value.length - 1) return;
  const temp = orderedHosts.value[index];
  orderedHosts.value[index] = orderedHosts.value[index + 1];
  orderedHosts.value[index + 1] = temp;
}

async function handleSave() {
  saving.value = true;
  const hostIds = orderedHosts.value.map((host) => host.id);
  const result = await updateHostOrder(hostIds);
  saving.value = false;

  if (!result.success) {
    err(result.message || "保存排序失败");
    return;
  }

  success("排序已保存");
  emit("saved");
  dialogVisible.value = false;
}
</script>

<template>
  <el-dialog
    v-model="dialogVisible"
    title="调整主机排序"
    width="480px"
    destroy-on-close
  >
    <div class="order-tip">拖拽主机或使用箭头按钮调整顺序，排序将应用于所有页面。</div>
    <div
      ref="listRef"
      class="order-list"
      @touchmove="onTouchMove"
      @touchend="onTouchEnd"
    >
      <div
        v-for="(host, index) in orderedHosts"
        :key="host.id"
        class="order-item"
        :class="{ dragging: draggingIndex === index }"
        @mousedown="(e) => onMouseDown(e, index)"
        @touchstart.passive="(e) => onTouchStart(e, index)"
      >
        <div class="order-item-content">
          <el-icon class="drag-handle">
            <img src="/drag.svg" alt="拖动" />
          </el-icon>
          <span class="host-name">{{ host.name || `主机 ${host.id}` }}</span>
        </div>
        <div class="order-actions">
          <el-button
            link
            :icon="ArrowUp"
            :disabled="index === 0"
            @click.stop="moveUp(index)"
          />
          <el-button
            link
            :icon="ArrowDown"
            :disabled="index === orderedHosts.length - 1"
            @click.stop="moveDown(index)"
          />
        </div>
      </div>
    </div>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSave">保存排序</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.order-tip {
  color: #909399;
  font-size: 12px;
  margin-bottom: 12px;
  line-height: 1.6;
}

.order-list {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  -webkit-overflow-scrolling: touch;
}

.order-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-bottom: 1px solid #e4e7ed;
  background: #fff;
  transition: background-color 0.2s;
  user-select: none;
  -webkit-user-select: none;
  cursor: grab;
}

.order-item:last-child {
  border-bottom: none;
}

.order-item:hover {
  background: #f5f7fa;
}

.order-item.dragging {
  background: #ecf5ff;
  opacity: 0.8;
  cursor: grabbing;
}

.order-item-content {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
  pointer-events: none;
}

.drag-handle {
  color: #909399;
  flex-shrink: 0;
}

.host-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.order-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}
</style>
