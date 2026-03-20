<script setup>
const currentPage = defineModel("current-page");
const props = defineProps({
  total: {
    type: Number,
    required: true,
  },
  pageSize: {
    type: Number,
    default: 10,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
});
const paginationSize = ref("default");
const layout = ref("total, prev, pager, next, jumper");
const handleWidthChange = () => {
  if (window.innerWidth < 768) {
    paginationSize.value = "small";
  } else {
    paginationSize.value = "default";
  }

  if (window.innerWidth < 480) {
    layout.value = "total, prev, pager, next";
  } else {
    layout.value = "total, prev, pager, next, jumper";
  }
};
onMounted(() => {
  handleWidthChange();
  window.addEventListener("resize", handleWidthChange);
});
</script>

<template>
  <el-pagination
      v-model:current-page="currentPage"
      :disabled="disabled"
      :background="false"
      :layout="layout"
      :total="total"
      :page-size="pageSize"
      :size="paginationSize"
  />
</template>

<style scoped>

</style>