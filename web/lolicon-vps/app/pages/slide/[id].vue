<script setup>
import { ref, computed, onMounted, onUnmounted } from "vue";

const route = useRoute();
const reportId = route.params.id;
const report = ref(null);
const currentSlideIndex = ref(0);
const speedTestType = ref(0);

// Fetch report data on client side
const { getReportDetails } = useReport();
const data = await getReportDetails(reportId);
if (data && data.data) {
  report.value = data.data;
}

const speedTestLabels = ["大陆三网多线程", "大陆三网单线程", "国际方向多线程"];

// Define slides based on the order in AGENTS.md
const slides = computed(() => {
  if (!report.value) return [];

  const slideList = [];

  // Slide 1: ITDOG - 全国Ping
  if (report.value.itdog?.ping) {
    slideList.push({
      title: "全国Ping",
      type: "itdog-ping",
      data: report.value.itdog.ping,
    });
  }

  // Slides 2-4: ITDOG - 去程路由1, 2, 3
  if (report.value.itdog?.route?.length) {
    report.value.itdog.route.forEach((routeImg, index) => {
      slideList.push({
        title: `去程路由${index + 1}`,
        type: "itdog-route",
        data: routeImg,
      });
    });
  }

  // Slide 5: ECS - 路由类型
  if (report.value.ecs?.trace?.types) {
    slideList.push({
      title: "路由类型",
      type: "ecs-route-types",
      data: report.value.ecs.trace.types,
    });
  }

  // Slide 6: ECS - 回程路由
  if (report.value.ecs?.trace?.back_route) {
    slideList.push({
      title: "回程路由",
      type: "ecs-back-route",
      data: report.value.ecs.trace.back_route,
    });
  }

  // Slide 7: 速度测试
  if (report.value.spdtest?.length) {
    slideList.push({
      title: "速度测试",
      type: "speed-test",
      data: report.value.spdtest,
    });
  }

  // Slide 8: ECS - 系统信息
  if (report.value.ecs?.info) {
    slideList.push({
      title: "系统信息",
      type: "ecs-info",
      data: report.value.ecs.info,
    });
  }

  // Slide 9: 性能测试页
  if (
    report.value.ecs?.cpu ||
    report.value.ecs?.mem ||
    report.value.ecs?.disk ||
    report.value.disk
  ) {
    slideList.push({
      title: "性能测试",
      type: "performance",
      data: {
        cpu: report.value.ecs?.cpu,
        mem: report.value.ecs?.mem,
        disk: report.value.ecs?.disk,
        diskTest: report.value.disk,
      },
    });
  }

  // Slide 10: IP 质量检测
  if (report.value.ipquality) {
    slideList.push({
      title: "IP 质量检测",
      type: "ip-quality",
      data: report.value.ipquality,
    });
  }

  // Slide 11: 流媒体解锁测试
  if (report.value.media) {
    slideList.push({
      title: "流媒体解锁测试",
      type: "media-unlock",
      data: report.value.media,
    });
  }

  return slideList;
});

const currentSlide = computed(() => slides.value[currentSlideIndex.value]);

const goToNextSlide = () => {
  if (currentSlideIndex.value < slides.value.length - 1) {
    currentSlideIndex.value++;
  }
};

const goToPrevSlide = () => {
  if (currentSlideIndex.value > 0) {
    currentSlideIndex.value--;
  }
};

const handleKeydown = (event) => {
  if (event.key === "ArrowRight") {
    goToNextSlide();
  } else if (event.key === "ArrowLeft") {
    goToPrevSlide();
  }
};

onMounted(() => {
  window.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleKeydown);
});

const getMediaStatusColor = (status) => {
  if (
    status.includes("Banned") ||
    status.includes("Risky") ||
    status.includes("Only") ||
    status.includes("No") ||
    status.includes("Failed")
  ) {
    return "var(--el-color-danger)";
  }
  return "var(--el-color-success)";
};

const diskLabels = ["测试项目", "读速度 (MB/s)", "写速度 (MB/s)"];
</script>

<template>
  <div id="slide-container">
    <div v-if="!report" class="loading">
      <el-empty description="加载中..." />
    </div>

    <div v-else-if="slides.length === 0" class="loading">
      <el-empty description="无可用幻灯片" />
    </div>

    <div v-else class="slide-wrapper">
      <!-- Navigation Header -->
      <div class="slide-header">
        <h1>{{ report.title || "测试报告" }}</h1>
        <!-- <div class="slide-nav">
          <el-button @click="goToPrevSlide" :disabled="currentSlideIndex === 0">
            ← 上一页
          </el-button>
          <span class="slide-counter">
            {{ currentSlideIndex + 1 }} / {{ slides.length }}
          </span>
          <el-button
            @click="goToNextSlide"
            :disabled="currentSlideIndex === slides.length - 1"
          >
            下一页 →
          </el-button>
        </div> -->
      </div>

      <!-- Slide Content -->
      <div class="slide-content">
        <div class="slide-title">{{ currentSlide.title }}</div>

        <!-- ITDOG Ping -->
        <div v-if="currentSlide.type === 'itdog-ping'" class="slide-body">
          <img :src="currentSlide.data" alt="Ping Result" class="slide-image" />
        </div>

        <!-- ITDOG Route -->
        <div v-else-if="currentSlide.type === 'itdog-route'" class="slide-body">
          <img
            :src="currentSlide.data"
            alt="Route Result"
            class="slide-image"
          />
        </div>

        <!-- ECS Route Types -->
        <div
          v-else-if="currentSlide.type === 'ecs-route-types'"
          class="slide-body"
        >
          <div class="card-grid">
            <el-card
              v-for="(value, key) in currentSlide.data"
              :key="key"
              shadow="hover"
              :class="{
                'vip-route':
                  value.includes('精品线路') || value.includes('优质线路'),
              }"
            >
              <h3>{{ key }}</h3>
              <p>{{ value }}</p>
            </el-card>
          </div>
        </div>

        <!-- ECS Back Route -->
        <div
          v-else-if="currentSlide.type === 'ecs-back-route'"
          class="slide-body"
        >
          <pre class="code-block">{{ currentSlide.data }}</pre>
        </div>

        <!-- Speed Test -->
        <div v-else-if="currentSlide.type === 'speed-test'" class="slide-body">
          <div class="speed-test-controls">
            <el-button
              v-for="(label, index) in speedTestLabels"
              :key="index"
              :text="speedTestType !== index"
              bg
              :type="speedTestType === index ? 'primary' : 'default'"
              @click="speedTestType = index"
            >
              {{ label }}
            </el-button>
          </div>
          <div
            v-if="currentSlide.data[speedTestType]"
            class="speed-test-content"
          >
            <h3 v-if="currentSlide.data[speedTestType].time">
              测试时间: {{ currentSlide.data[speedTestType].time }}
            </h3>
            <el-table
              :data="currentSlide.data[speedTestType].results"
              border
              stripe
            >
              <el-table-column prop="spot" label="测试点" min-width="150" />
              <el-table-column
                prop="download"
                label="下载速度 (Mbps)"
                min-width="150"
              >
                <template #default="{ row }">
                  {{ row.download?.toFixed(2) || "N/A" }}
                </template>
              </el-table-column>
              <el-table-column
                prop="upload"
                label="上传速度 (Mbps)"
                min-width="150"
              >
                <template #default="{ row }">
                  {{ row.upload?.toFixed(2) || "N/A" }}
                </template>
              </el-table-column>
              <el-table-column prop="latency" label="延迟 (ms)" min-width="120">
                <template #default="{ row }">
                  {{ row.latency?.toFixed(2) || "N/A" }}
                </template>
              </el-table-column>
              <el-table-column prop="jitter" label="抖动 (ms)" min-width="120">
                <template #default="{ row }">
                  {{ row.jitter?.toFixed(2) || "N/A" }}
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>

        <!-- ECS Info -->
        <div v-else-if="currentSlide.type === 'ecs-info'" class="slide-body">
          <el-descriptions :column="2" border>
            <el-descriptions-item
              v-for="(value, key) in currentSlide.data"
              :key="key"
              :label="key"
            >
              {{ value }}
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- Performance -->
        <div v-else-if="currentSlide.type === 'performance'" class="slide-body">
          <!-- CPU Performance -->
          <div v-if="currentSlide.data.cpu" class="subsection">
            <h3>CPU 性能</h3>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-statistic
                  title="单核得分"
                  :value="currentSlide.data.cpu.single || 0"
                />
              </el-col>
              <el-col :span="12">
                <el-statistic
                  title="多核得分"
                  :value="currentSlide.data.cpu.multi || 0"
                />
              </el-col>
            </el-row>
          </div>

          <!-- Memory Performance -->
          <div v-if="currentSlide.data.mem" class="subsection">
            <h3>内存性能</h3>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-statistic
                  title="读取速度 (MB/s)"
                  :value="currentSlide.data.mem.read || 0"
                  :precision="2"
                />
              </el-col>
              <el-col :span="12">
                <el-statistic
                  title="写入速度 (MB/s)"
                  :value="currentSlide.data.mem.write || 0"
                  :precision="2"
                />
              </el-col>
            </el-row>
          </div>

          <!-- Disk Performance -->
          <div v-if="currentSlide.data.disk" class="subsection">
            <h3>磁盘性能</h3>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="顺序读取">
                {{ currentSlide.data.disk.seq_read }}
              </el-descriptions-item>
              <el-descriptions-item label="顺序写入">
                {{ currentSlide.data.disk.seq_write }}
              </el-descriptions-item>
            </el-descriptions>
          </div>

          <!-- Disk Test -->
          <div v-if="currentSlide.data.diskTest" class="subsection">
            <h3>磁盘测试</h3>
            <el-table :data="currentSlide.data.diskTest.data" border stripe>
              <el-table-column
                v-for="(col, index) in currentSlide.data.diskTest.data[0] || []"
                :key="index"
                :label="diskLabels[index]"
                min-width="100"
              >
                <template #default="{ row }">
                  {{ row[index] }}
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>

        <!-- IP Quality -->
        <div
          v-else-if="currentSlide.type === 'ip-quality'"
          class="slide-body ip-quality-slide"
        >
          <!-- Head Info -->
          <div v-if="currentSlide.data.Head" class="subsection">
            <h3>检测信息</h3>
            <el-descriptions :column="2" border>
              <el-descriptions-item label-width="150px" label="IP">{{
                currentSlide.data.Head.IP
              }}</el-descriptions-item>
              <el-descriptions-item label-width="150px" label="检测时间">{{
                currentSlide.data.Head.Time
              }}</el-descriptions-item>
            </el-descriptions>
          </div>

          <!-- Geographic Info -->
          <div
            v-if="currentSlide.data.Info"
            class="subsection"
          >
            <h3>地理位置信息</h3>
            <el-descriptions label-width="150px" :column="2" border>
              <el-descriptions-item label="ASN">{{
                currentSlide.data.Info.ASN
              }}</el-descriptions-item>
              <el-descriptions-item label-width="150px" label="组织">{{
                currentSlide.data.Info.Organization
              }}</el-descriptions-item>
              <el-descriptions-item label-width="150px" label="国家/地区">
                {{ currentSlide.data.Info.Region?.Name }} ({{
                  currentSlide.data.Info.Region?.Code
                }})
              </el-descriptions-item>
              <el-descriptions-item label-width="150px" label="城市">{{
                currentSlide.data.Info.City?.Name
              }}</el-descriptions-item>
            </el-descriptions>
          </div>

          <!-- Risk Scores -->
          <div
            v-if="currentSlide.data.Score"
            class="subsection"
          >
            <h3>风险评分</h3>
            <el-descriptions :column="3" border>
              <el-descriptions-item
              label-width="150px"
                v-for="(value, key) in currentSlide.data.Score"
                :key="key"
                :label="key"
              >
                {{ value }}
              </el-descriptions-item>
            </el-descriptions>
          </div>

          <div v-if="report.ipquality.Type" class="subsection">
            <el-row :gutter="16">
              <el-col :span="12" v-if="report.ipquality.Type.Company">
                <h4>公司类型</h4>
                <el-descriptions :column="1" border size="small">
                  <el-descriptions-item label-width="150px"
                    v-for="(value, key) in report.ipquality.Type.Company"
                    :key="key"
                    :label="key"
                  >
                    {{ value }}
                  </el-descriptions-item>
                </el-descriptions>
              </el-col>
              <el-col :span="12" v-if="report.ipquality.Type.Usage">
                <h4>使用类型</h4>
                <el-descriptions :column="1" border size="small">
                  <el-descriptions-item label-width="150px"
                    v-for="(value, key) in report.ipquality.Type.Usage"
                    :key="key"
                    :label="key"
                  >
                    {{ value }}
                  </el-descriptions-item>
                </el-descriptions>
              </el-col>
            </el-row>
          </div>
        </div>

        <!-- Media Unlock -->
        <div
          v-else-if="currentSlide.type === 'media-unlock'"
          class="slide-body media-unlock-slide"
        >
          <div v-if="currentSlide.data.ipv4?.length" class="subsection">
            <h3>IPv4 解锁情况</h3>
            <div class="card-grid media-unlock-grid">
              <el-card
                shadow="never"
                v-for="(block, index) in currentSlide.data.ipv4"
                :key="'ipv4-' + index"
              >
                <h4>{{ block.region }}</h4>
                <div
                  v-for="(item, idx) in block.results"
                  :key="idx"
                  class="media-unlock-body"
                >
                  <div>{{ item.media }}</div>
                  <div :style="{ color: getMediaStatusColor(item.unlock) }">
                    {{ item.unlock }}
                  </div>
                </div>
              </el-card>
            </div>
          </div>

          <div v-if="currentSlide.data.ipv6?.length" class="subsection">
            <h3>IPv6 解锁情况</h3>
            <div class="card-grid media-unlock-grid">
              <el-card
                shadow="never"
                v-for="(block, index) in currentSlide.data.ipv6"
                :key="'ipv6-' + index"
              >
                <h4>{{ block.region }}</h4>
                <div
                  v-for="(item, idx) in block.results"
                  :key="idx"
                  class="media-unlock-body"
                >
                  <div>{{ item.media }}</div>
                  <div :style="{ color: getMediaStatusColor(item.unlock) }">
                    {{ item.unlock }}
                  </div>
                </div>
              </el-card>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
#slide-container {
  width: 100vw;
  height: 100vh;
  background-color: #f5f7fa;
  overflow: hidden;
}

.loading {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
}

.slide-wrapper {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.slide-header {
  background-color: #fff;
  padding: 16px 24px;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.slide-header h1 {
  margin: 0;
  font-size: 16px;
  font-weight: 300;
  color: #303133;
}

.slide-nav {
  display: flex;
  gap: 16px;
  align-items: center;
}

.slide-counter {
  font-size: 14px;
  color: #606266;
  min-width: 80px;
  text-align: center;
}

.slide-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background-color: #fff;
  margin: 16px;
  border-radius: 8px;
}

.slide-title {
  font-size: 16px;
  font-weight: 600;
  color: #606266;
  margin: 8px 0 24px 0;
}

.slide-body {
  margin-top: 24px;
}

.slide-image {
  max-width: 100%;
  height: auto;
  display: block;
  margin: 0 auto;
}

.code-block {
  background-color: #f5f7fa;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 16px;
  font-family: "Courier New", Courier, monospace;
  font-size: 14px;
  white-space: pre-wrap;
  word-wrap: break-word;
  color: #303133;
  margin: 0;
  overflow-y: auto;
  max-height: 70vh;
}

.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
  gap: 16px;
}

.card-grid .el-card h3 {
  margin: 0 0 8px 0;
  font-size: 16px;
  color: #303133;
}

.card-grid .el-card p {
  margin: 0;
  font-size: 14px;
  color: #606266;
}

.vip-route {
  color: var(--el-color-success);
  background-color: var(--el-color-success-light-8);
}

.speed-test-controls {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.speed-test-content h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
  color: #606266;
}

.subsection {
  margin-bottom: 32px;
}

.subsection:last-child {
  margin-bottom: 0;
}

.subsection h3 {
  font-size: 20px;
  font-weight: 600;
  margin: 0 0 16px 0;
  color: #303133;
}

.media-unlock-grid :deep(.el-card__body) {
  padding: 16px;
}

.media-unlock-grid h4 {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.media-unlock-body {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  margin: 8px 0;
  padding: 8px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.media-unlock-body div {
  font-size: 14px;
}

:deep(.el-statistic__head) {
  font-size: 16px;
  color: #909399;
}

:deep(.el-statistic__content) {
  font-size: 32px;
  font-weight: 600;
  color: #303133;
}

.ip-quality-slide .subsection,
.media-unlock-slide .subsection {
  margin-bottom: 24px;
}
</style>
