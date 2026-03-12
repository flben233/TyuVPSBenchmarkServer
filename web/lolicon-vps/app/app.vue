<script setup>
  const activePath = computed(() => {
    return useRoute().path;
  });
  const excludePaths = ["/slide/", "/inspector"];

  const { userInfo, login, refreshToken, token } = useAuth();

  onMounted(async () => {
    const params = new URLSearchParams(window.location.search);
    const code = params.get("code");
    console.log("OAuth code from URL:", code);
    if (code) {
      const cleanUrl = window.location.origin + window.location.pathname;
      window.history.replaceState({}, document.title, cleanUrl);
      await login(code);
      console.log("User Info after login:", userInfo.value);
    } else if (!userInfo.value && token.value) {
      await refreshToken();
    }
  });
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
    background: transparent !important;
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
