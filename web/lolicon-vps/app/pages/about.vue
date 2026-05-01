<script setup>
useHead({
  title: '关于 - Lolicon VPS',
});
</script>

<template>
  <div id="about-root">
    <div id="about-title">关于</div>

    <el-card shadow="never" class="about-section">
      <template #header>作者信息</template>
      <div class="info-row">
        <span class="label">名称</span>
        <span>ShirakawaTyu</span>
      </div>
      <div class="info-row">
        <span class="label">GitHub</span>
        <el-link href="https://github.com/flben233" target="_blank" type="primary">flben233</el-link>
      </div>
      <div class="info-row">
        <span class="label">Bilibili</span>
        <el-link href="https://space.bilibili.com/21650658" target="_blank" type="primary">21650658</el-link>
      </div>
    </el-card>

    <el-card shadow="never" class="about-section">
      <template #header>网站信息</template>
      <div class="info-row">
        <span class="label">开源协议</span>
        <span>AGPL-3.0</span>
      </div>
      <div class="info-row">
        <span class="label">开源地址</span>
        <el-link href="https://github.com/flben233/TyuVPSBenchmarkServer" target="_blank" type="primary">TyuVPSBenchmarkServer</el-link>
      </div>
    </el-card>

    <el-card shadow="never" class="about-section">
      <template #header>测试脚本</template>
      <div class="info-row">
        <span class="label">IP质量</span>
        <el-link href="https://github.com/xykt/IPQuality" target="_blank" type="primary">IPQuality</el-link>
      </div>
      <div class="info-row">
        <span class="label">网络测速</span>
        <el-link href="https://github.com/flben233/cdn-speed" target="_blank" type="primary">cdn-speed</el-link>
      </div>
      <div class="info-row">
        <span class="label">融合怪</span>
        <el-link href="https://github.com/spiritLHLS/ecs" target="_blank" type="primary">ecs</el-link>
      </div>
      <div class="info-row">
        <span class="label">磁盘测试</span>
        <el-link href="https://github.com/flben233/tyu-disk" target="_blank" type="primary">tyu-disk</el-link>
      </div>
      <div class="info-row">
        <span class="label">流媒体解锁</span>
        <el-link href="https://github.com/lmc999/RegionRestrictionCheck" target="_blank" type="primary">RegionRestrictionCheck</el-link>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
#about-root {
  padding: 16px;
  box-sizing: border-box;
  overflow-x: hidden;
}
#about-title {
  font-size: 28px;
  font-weight: 300;
  margin-bottom: 16px;
}
.about-section {
  margin-bottom: 16px;
}
.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
}
.label {
  color: var(--el-text-color-regular);
}
</style>
