<script setup>
import {
  UserFilled,
  House,
  Search,
  Odometer,
  Notebook,
  More,
} from "@element-plus/icons-vue";

const mode = ref("vertical");
const { userInfo, logout } = useAuth();
const avatarUrl = ref(null);
const activePath = computed(() => {
  return useRoute().path;
});
const ellipse = ref(false);

if (userInfo.value) {
  avatarUrl.value = userInfo.value.avatarUrl;
}

function handleWidthChange() {
  if (window.innerWidth < 768) {
    mode.value = "horizontal";
  } else {
    mode.value = "vertical";
  }

  ellipse.value = window.innerWidth < 384;
}

watch(userInfo, (newVal) => {
  if (newVal) {
    avatarUrl.value = newVal.avatarUrl;
  } else {
    avatarUrl.value = null;
  }
});

onMounted(async () => {
  handleWidthChange();
  window.addEventListener("resize", handleWidthChange);
});

onUnmounted(() => {
  window.removeEventListener("resize", handleWidthChange);
});

const avatarNav = (command) => {
  if (command === "login") {
    const { clientId } = useAppConfig();
    window.location.href = "https://github.com/login/oauth/authorize?client_id=" + clientId;
  } else if (command === "logout") {
    logout();
  } else if (command === "center") {
    useRouter().push("/center");
  } else if (command === "inspector") {
    useRouter().push("/inspector");
  }
};
</script>

<template>
  <el-menu
    id="m-root"
    :mode="mode"
    :collapse="true"
    :default-active="activePath"
    router
    :ellipsis="ellipse"
  >
    <div class="m-avatar-container">
      <el-dropdown @command="avatarNav" class="m-avatar-dropdown">
        <el-avatar v-if="avatarUrl" :src="avatarUrl" />
        <el-avatar v-else :icon="UserFilled" />
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item v-if="!userInfo" command="login">
              使用 GitHub 登录
            </el-dropdown-item>
            <template v-else>
              <el-dropdown-item command="center">个人中心</el-dropdown-item>
              <el-dropdown-item command="inspector">Loli 探针</el-dropdown-item>
              <el-dropdown-item command="logout">退出登录</el-dropdown-item>
            </template>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <el-menu-item index="/" class="menu-item first-item">
      <el-icon size="24">
        <House />
      </el-icon>
      <template #title>
        <span>主页</span>
      </template>
    </el-menu-item>
    <el-menu-item index="/search" class="menu-item">
      <el-icon size="24">
        <Search />
      </el-icon>
      <template #title>
        <span>搜索</span>
      </template>
    </el-menu-item>
    <el-sub-menu index="a">
      <template #title>
        <el-icon size="24">
          <Odometer />
        </el-icon>
      </template>
      <el-menu-item index="/monitor" class="menu-item">
        公共服务器监控
      </el-menu-item>
      <el-menu-item index="/inspector" class="menu-item">
        Lolicon Monitor
      </el-menu-item>
    </el-sub-menu>

    <el-menu-item index="/looking-glass" class="menu-item">
      <el-icon size="24">
        <Notebook />
      </el-icon>
      <template #title>
        <span>Looking Glass</span>
      </template>
    </el-menu-item>
    <el-sub-menu index="b">
      <template #title>
        <el-icon size="24">
          <More />
        </el-icon>
      </template>
      <el-menu-item index="/tools/ipquery" class="menu-item">
        IP查询
      </el-menu-item>
      <el-menu-item index="/tools/traceroute" class="menu-item">
        路由追踪
      </el-menu-item>
      <el-menu-item index="/tools/whois" class="menu-item">
        Whois查询
      </el-menu-item>
    </el-sub-menu>
  </el-menu>
</template>

<style scoped>
.m-avatar-container {
  text-align: center;
  padding: 8px;
  box-sizing: border-box;
  margin: auto;
}

.menu-item {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 64px;
}

.menu-item:nth-child(2) {
  margin-left: auto;
}

@media screen and (max-width: 768px) {
  #m-root {
    width: 100% !important;
    padding: 0 16px;
    align-items: center;
  }

  .m-avatar-dropdown {
    align-items: center;
    justify-content: center;
    padding: 6px 0;
    width: 64px;
    box-sizing: border-box;
  }

  .m-avatar-container {
    padding: 0;
    margin: 8px 0;
  }
}
</style>
