<template>
  <div id="site-title">Lolicon VPS</div>
  <div id="site-owner-container">
    <el-avatar size="large" src="/avatar.jpg" />
    <div id="site-owner">
      <div>ShirakawaTyu</div>
      <div style="display: grid; grid-template-columns: auto auto; gap: 8px">
        <el-link :underline="'never'" href="https://github.com/flben233">
          <img src="/github-mark.svg" alt="GitHub Icon" class="a-icon" />
          <div style="font-size: 13px">GitHub</div>
        </el-link>
        <el-link
          :underline="'never'"
          href="https://space.bilibili.com/21650658"
        >
          <img src="/bilibili-mark.svg" alt="Bilibili Icon" class="a-icon" />
          <div style="font-size: 13px">Bilibili</div>
        </el-link>
      </div>
    </div>
  </div>
  <el-divider />
  <slot></slot>
</template>

<style scoped>
#site-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 16px;
  color: var(--el-text-color-primary);
}
#site-owner-container {
  margin-bottom: 24px;
  display: flex;
  align-items: center;
}
#site-owner {
  margin-left: 8px;
  font-size: 16px;
  font-weight: 500;
  justify-content: space-between;
  display: flex;
  flex-direction: column;
  height: 56px;
  box-sizing: border-box;
  padding: 4px;
}
.a-icon {
  width: 13px;
  height: 13px;
  margin-right: 2px;
}
</style>
