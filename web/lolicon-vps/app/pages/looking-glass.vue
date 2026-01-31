<script setup>
useHead({
  title: 'Looking Glass 列表 - Lolicon VPS',
  meta: [
    { name: 'description', content: 'Looking Glass 工具列表，提供全球各地服务器的网络测试工具，可进行路由追踪、Ping测试等网络诊断' },
    { name: 'keywords', content: 'Looking Glass,LG,路由追踪,网络测试,Ping测试,网络诊断,traceroute,BGP' },
    { property: 'og:title', content: 'Looking Glass 列表 - Lolicon VPS' },
    { property: 'og:description', content: 'Looking Glass 工具列表，提供全球各地服务器的网络测试工具' },
    { property: 'og:type', content: 'website' }
  ]
});

const lgData = ref([]);
const { listPublicLookingGlass } = useLookingGlass();
const data = await listPublicLookingGlass();
lgData.value = data;
</script>

<template>
  <div id="lg-root">
    <div id="lg-title"> Looking Glass 列表 </div>
    <el-empty v-if="lgData.length === 0" description="暂无 Looking Glass 数据" />
    <el-card class="lg-item" v-for="item in lgData" :key="item.id" shadow="never">
      <div class="lg-item-header">
        {{ item.server_name }}
      </div>
        <el-link :href="item.test_url" target="_blank" type="primary" class="lg-item-url">
        {{ item.test_url }}
      </el-link>
      <div class="lg-item-uploader">
        上传者: {{ item.uploader_name }}
      </div>
    </el-card>
  </div>
</template>

<style scoped>
  #lg-root {
    padding: 16px;
    box-sizing: border-box;
    width: 100%;
    height: 100%;
    overflow-y: auto;
  }
  #lg-title {
    font-size: 28px;
    font-weight: 300;
    margin-bottom: 16px;
    margin-top: -3px;
    font-family: Noto Sans SC, sans-serif;
  }
  .lg-item {
    margin-bottom: 16px;
  }
  .lg-item-header {
    font-weight: 600;
    font-size: 18px;
    margin-bottom: 8px;
  }
  .lg-item-url {
    margin-bottom: 2px;
    font-size: 16px;
  }
  .lg-item-uploader {
    font-size: 14px;
    color: #909399;
  }
</style>