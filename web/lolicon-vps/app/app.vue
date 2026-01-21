<script setup>
  const activePath = computed(() => {
    return useRoute().path;
  });
  const excludePaths = ["/slide/"];
</script>

<template>
  <div id="app">
    <NuxtLoadingIndicator color="#39C5BB" />
    <Sidebar v-if="!excludePaths.some(path => activePath.includes(path))"/>
    <NuxtPage/>
  </div>
</template>

<style>
  @media screen and (max-width: 768px) {
    #app {
      flex-direction: column;
    }
  }
  #app {
    display: flex;
    height: 100vh;
    width: 100vw;
    font-family: sans-serif;
  }
  body {
    margin: 0;
  }
  .el-input__wrapper, .el-select__wrapper {
    box-shadow: none !important;
    border-radius: 0 !important;
    border-bottom: 1px var(--el-border-color) solid;
    padding: 1px 8px;
    transition: var(--el-transition-border);
  }
  .el-input__wrapper:hover {
    box-shadow: none;
    border-bottom: 1px var(--el-input-hover-border-color) solid;
  }
  .el-input__wrapper.is-focus {
    box-shadow: none;
    border-bottom: 1px var(--el-color-primary) solid;
  }
  .el-input-number__decrease, .el-input-number__increase {
    background-color: transparent;
    border: none;
  }
  .el-input-number__decrease.is-hovering {
    box-shadow: none !important;
    border-bottom: 1px var(--el-border-color-hover) solid;
  }
  .el-select__wrapper.is-hovering {
    box-shadow: none !important;
    border-bottom: 1px var(--el-border-color-hover) solid;
  }
  .el-select__wrapper.is-focused {
    box-shadow: none !important;
    border-bottom: 1px var(--el-color-primary) solid;
  }
  .el-popper__arrow {
    display: none;
  }
</style>
