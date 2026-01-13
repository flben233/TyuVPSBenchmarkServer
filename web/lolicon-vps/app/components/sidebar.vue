<script setup>
import {
  UserFilled,
  House,
  Search,
  Odometer,
  Notebook,
  More,
} from "@element-plus/icons-vue";
const { userInfo } = useAuth();
const avatarUrl = ref(null);
const activePath = computed(() => {
  console.log("Current Route Path:", useRoute().path);
  return useRoute().path;
});

if (userInfo.value) {
  console.log("User Info:", userInfo.value);
  avatarUrl.value = userInfo.value.avatarUrl;
} else {
  console.log("No user info available");
}

const token = useRoute().params.token;
if (token) {
  const { login } = useAuth();
  login(token);
}

const avatarNav = (command) => {
  if (command === "login") {
    window.location.href = "https://github.com/login/oauth/authorize?client_id=Ov23limxDDoGO9of9P4m";
  } else if (command === "logout") {
    const { logout } = useAuth();
    logout();
  } else if (command === "profile") {
    useRouter().push("/profile");
  }
};
</script>

<template>
  <el-menu :collapse="true" :default-active="activePath" router>
    <div class="m-avatar-container">
      <el-dropdown @command="avatarNav">
        <el-avatar v-if="avatarUrl" :src="avatarUrl" />
        <el-avatar v-else :icon="UserFilled" />
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item v-if="!userInfo" command="login">使用Github登录</el-dropdown-item>
            <div v-else>
              <el-dropdown-item command="profile">个人中心</el-dropdown-item>
              <el-dropdown-item command="logout">退出登录</el-dropdown-item>
            </div>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <el-menu-item index="/" class="menu-item">
      <el-icon size="24">
        <House />
      </el-icon>
    </el-menu-item>
    <el-menu-item index="/search" class="menu-item">
      <el-icon size="24">
        <Search />
      </el-icon>
    </el-menu-item>
    <el-menu-item index="/monitor" class="menu-item">
      <el-icon size="24">
        <Odometer />
      </el-icon>
    </el-menu-item>
    <!-- TODO: 添加loogking glass列表相关页面和接口 -->
    <el-menu-item index="/looking-glass" class="menu-item">
      <el-icon size="24">
        <Notebook />
      </el-icon>
    </el-menu-item>
    <el-sub-menu index="4">
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
        whois查询
      </el-menu-item>
    </el-sub-menu>
  </el-menu>
</template>

<style scoped>
.m-avatar-container {
  width: 100%;
  text-align: center;
  padding: 8px;
  box-sizing: border-box;
  /* cursor: pointer; */
}
.menu-item {
  display: flex;
  align-items: center;
  justify-content: center;
}
.menu-icon {
  margin: 0;
}
</style>
