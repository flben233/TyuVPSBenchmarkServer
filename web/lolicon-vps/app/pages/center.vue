<script setup>
useHead({
  title: '个人中心 - Lolicon VPS',
  meta: [
    { name: 'description', content: '管理您的监控主机、Looking Glass记录和VPS测试报告，查看审核状态，支持管理员审核功能' },
    { name: 'keywords', content: '个人中心,用户中心,监控管理,Looking Glass管理,报告管理,VPS管理' },
    { property: 'og:title', content: '个人中心 - Lolicon VPS' },
    { property: 'og:description', content: '管理您的监控主机、Looking Glass记录和测试报告' },
    { property: 'og:type', content: 'website' }
  ]
});

import { ref } from "vue";

const { token, isAdmin } = useAuth();
const activeTab = ref("monitor");
const monitorRef = ref(null);

const monitorHosts = computed(() => {
  return monitorRef.value?.hosts || [];
});
</script>

<template>
  <el-row id="center-root">
    <el-col :span="24" id="content-area">
      <h1 class="page-title">个人中心</h1>
      <div v-if="!token" class="login-prompt">
        <el-empty description="请登录以访问个人中心">
          <el-button type="primary" @click="$router.push('/login')"
            >前往登录</el-button
          >
        </el-empty>
      </div>

      <div v-else>
        <el-tabs v-model="activeTab" class="center-tabs">
          <!-- Monitor Hosts Tab -->
          <el-tab-pane label="监控主机" name="monitor">
            <center-monitor ref="monitorRef" :token="token" :is-admin="isAdmin" />
          </el-tab-pane>

          <!-- Looking Glass Tab -->
          <el-tab-pane label="Looking Glass" name="lookingglass">
            <center-looking-glass :token="token" :is-admin="isAdmin" />
          </el-tab-pane>

          <!-- Admin: Report Management Tab -->
          <el-tab-pane label="报告管理" name="report" v-if="isAdmin">
            <report-uploader :token="token" :hosts="monitorHosts" />
          </el-tab-pane>

          <!-- Admin: User Management Tab -->
          <el-tab-pane label="用户管理" name="users" v-if="isAdmin">
            <center-users :token="token" />
          </el-tab-pane>

          <!-- Admin: User Group Management Tab -->
          <el-tab-pane label="用户组管理" name="groups" v-if="isAdmin">
            <center-groups :token="token" />
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-col>
  </el-row>
</template>

<style scoped>
#center-root {
  padding: 16px;
  box-sizing: border-box;
  overflow-x: hidden;
}

#content-area {
  margin: 0 auto;
}

.page-title {
  font-size: 32px;
  font-weight: 300;
  margin: 0 0 24px 0;
  color: #303133;
}

.login-prompt {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}

.center-tabs {
  margin-top: 16px;
}
</style>
