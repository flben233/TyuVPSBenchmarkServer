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
    <el-menu-item index="/webssh" class="menu-item webssh-menu-item">
      <el-icon size="24">
        <svg viewBox="0 0 1024 1024" width="24" height="24"><path d="M796.444444 284.444444h-96.711111l91.022223-187.733333-51.2-22.755555L637.155556 284.444444H386.844444L278.755556 73.955556l-45.511112 22.755555L324.266667 284.444444H227.555556c-62.577778 0-113.777778 51.2-113.777778 113.777778v455.111111c0 62.577778 51.2 113.777778 113.777778 113.777778h568.888888c62.577778 0 113.777778-51.2 113.777778-113.777778V398.222222c0-62.577778-51.2-113.777778-113.777778-113.777778z m56.888889 568.888889c0 34.133333-22.755556 56.888889-56.888889 56.888889H227.555556c-34.133333 0-56.888889-22.755556-56.888889-56.888889V398.222222c0-34.133333 22.755556-56.888889 56.888889-56.888889h568.888888c34.133333 0 56.888889 22.755556 56.888889 56.888889v455.111111zM0 455.111111h56.888889v341.333333H0zM967.111111 455.111111h56.888889v341.333333h-56.888889z" fill="currentColor"/><path d="M312.888889 597.333333m-85.333333 0a85.333333 85.333333 0 1 0 170.666666 0 85.333333 85.333333 0 1 0-170.666666 0Z" fill="currentColor"/><path d="M711.111111 597.333333m-85.333333 0a85.333333 85.333333 0 1 0 170.666666 0 85.333333 85.333333 0 1 0-170.666666 0Z" fill="currentColor"/></svg>
      </el-icon>
      <template #title>
        <span>WebSSH</span>
      </template>
    </el-menu-item>

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
