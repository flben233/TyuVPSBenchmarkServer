<script setup>
import Profile from '~/components/profile.vue';

const { getBackrouteTypes, getMediaNames, getVirtualizations, searchReports } =
  useSearch();

// Load filter options
const backrouteTypes = ref([]);
const mediaNames = ref([]);
const virtualizations = ref([]);
const loading = ref(false);
const activeName = ref("cm");
const mainTab = ref("search");

// Search form data
const searchForm = ref({
  name: "",
  virtualization: "",
  ipv6_support: null,
  disk_level: null,
  media_unlocks: [],
  cm_params: {
    back_route: "",
    latency: null,
    min_download: null,
    max_download: null,
    min_upload: null,
    max_upload: null,
  },
  ct_params: {
    back_route: "",
    latency: null,
    min_download: null,
    max_download: null,
    min_upload: null,
    max_upload: null,
  },
  cu_params: {
    back_route: "",
    latency: null,
    min_download: null,
    max_download: null,
    min_upload: null,
    max_upload: null,
  },
});

// Results
const reports = ref([]);
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);
const disabled = ref(false);
const hasSearched = ref(false);

// Load options on mount
onMounted(async () => {
  backrouteTypes.value = await getBackrouteTypes();
  mediaNames.value = await getMediaNames();
  virtualizations.value = await getVirtualizations();
});

// Search function
const handleSearch = async (resetPage = true) => {
  loading.value = true;
  disabled.value = true;
  hasSearched.value = true;
  if (resetPage) {
    page.value = 1;
  }

  // Build search params - only include non-empty values
  const params = {};

  if (searchForm.value.name) params.name = searchForm.value.name;
  if (searchForm.value.virtualization)
    params.virtualization = searchForm.value.virtualization;
  if (searchForm.value.ipv6_support !== null)
    params.ipv6_support = searchForm.value.ipv6_support;
  if (searchForm.value.disk_level !== null)
    params.disk_level = searchForm.value.disk_level;
  if (searchForm.value.media_unlocks?.length > 0)
    params.media_unlocks = searchForm.value.media_unlocks;

  // Add CM params if any field is filled
  const cmParams = {};
  if (searchForm.value.cm_params.back_route)
    cmParams.back_route = searchForm.value.cm_params.back_route;
  if (searchForm.value.cm_params.latency !== null)
    cmParams.latency = searchForm.value.cm_params.latency;
  if (searchForm.value.cm_params.min_download !== null)
    cmParams.min_download = searchForm.value.cm_params.min_download;
  if (searchForm.value.cm_params.max_download !== null)
    cmParams.max_download = searchForm.value.cm_params.max_download;
  if (searchForm.value.cm_params.min_upload !== null)
    cmParams.min_upload = searchForm.value.cm_params.min_upload;
  if (searchForm.value.cm_params.max_upload !== null)
    cmParams.max_upload = searchForm.value.cm_params.max_upload;
  if (Object.keys(cmParams).length > 0) params.cm_params = cmParams;

  // Add CT params if any field is filled
  const ctParams = {};
  if (searchForm.value.ct_params.back_route)
    ctParams.back_route = searchForm.value.ct_params.back_route;
  if (searchForm.value.ct_params.latency !== null)
    ctParams.latency = searchForm.value.ct_params.latency;
  if (searchForm.value.ct_params.min_download !== null)
    ctParams.min_download = searchForm.value.ct_params.min_download;
  if (searchForm.value.ct_params.max_download !== null)
    ctParams.max_download = searchForm.value.ct_params.max_download;
  if (searchForm.value.ct_params.min_upload !== null)
    ctParams.min_upload = searchForm.value.ct_params.min_upload;
  if (searchForm.value.ct_params.max_upload !== null)
    ctParams.max_upload = searchForm.value.ct_params.max_upload;
  if (Object.keys(ctParams).length > 0) params.ct_params = ctParams;

  // Add CU params if any field is filled
  const cuParams = {};
  if (searchForm.value.cu_params.back_route)
    cuParams.back_route = searchForm.value.cu_params.back_route;
  if (searchForm.value.cu_params.latency !== null)
    cuParams.latency = searchForm.value.cu_params.latency;
  if (searchForm.value.cu_params.min_download !== null)
    cuParams.min_download = searchForm.value.cu_params.min_download;
  if (searchForm.value.cu_params.max_download !== null)
    cuParams.max_download = searchForm.value.cu_params.max_download;
  if (searchForm.value.cu_params.min_upload !== null)
    cuParams.min_upload = searchForm.value.cu_params.min_upload;
  if (searchForm.value.cu_params.max_upload !== null)
    cuParams.max_upload = searchForm.value.cu_params.max_upload;
  if (Object.keys(cuParams).length > 0) params.cu_params = cuParams;

  const result = await searchReports(params, page.value, pageSize.value);
  reports.value = result.data || [];
  total.value = result.total || 0;

  loading.value = false;
  disabled.value = false;

  // Switch to results tab after search
  mainTab.value = "results";
};

// Reset form
const handleReset = () => {
  searchForm.value = {
    name: "",
    virtualization: "",
    ipv6_support: null,
    disk_level: null,
    media_unlocks: [],
    cm_params: {
      back_route: "",
      latency: null,
      min_download: null,
      max_download: null,
      min_upload: null,
      max_upload: null,
    },
    ct_params: {
      back_route: "",
      latency: null,
      min_download: null,
      max_download: null,
      min_upload: null,
      max_upload: null,
    },
    cu_params: {
      back_route: "",
      latency: null,
      min_download: null,
      max_download: null,
      min_upload: null,
      max_upload: null,
    },
  };
  reports.value = [];
  total.value = 0;
  hasSearched.value = false;
};

// Watch page changes
watch(page, async (newPage) => {
  if (!hasSearched.value) return;
  await handleSearch(false); // Don't reset page when pagination triggers
});

// Navigate to detail
const gotoDetail = (reportId) => {
  useRouter().push(`/report/${reportId}`);
};
</script>

<template>
  <div id="search-root">
    <el-row>
      <el-col :span="24" id="search-title"> 高级搜索 </el-col>
      <el-col :span="17">
        <el-tabs v-model="mainTab">
          <!-- Search Panel -->
          <el-tab-pane label="搜索条件" name="search">
            <div class="search-form-card">
              <el-form
                :model="searchForm"
                label-width="120px"
                label-position="left"
              >
                <!-- Basic Filters -->
                <div class="filter-section">
                  <div class="section-title">基本信息</div>
                  <el-form-item label="关键词">
                    <el-input
                      v-model="searchForm.name"
                      placeholder="输入VPS名称关键词"
                      clearable
                    />
                  </el-form-item>
                  <el-form-item label="虚拟化技术">
                    <el-select
                      v-model="searchForm.virtualization"
                      placeholder="选择虚拟化技术"
                      clearable
                    >
                      <el-option
                        v-for="virt in virtualizations"
                        :key="virt"
                        :label="virt"
                        :value="virt"
                      />
                    </el-select>
                  </el-form-item>
                  <el-form-item label="IPv6支持">
                    <el-select
                      v-model="searchForm.ipv6_support"
                      placeholder="是否支持IPv6"
                      clearable
                    >
                      <el-option label="支持" :value="true" />
                      <el-option label="不支持" :value="false" />
                    </el-select>
                  </el-form-item>
                  <el-form-item label="磁盘等级">
                    <el-select
                      v-model="searchForm.disk_level"
                      placeholder="选择磁盘等级"
                      clearable
                    >
                      <el-option label="0级 (<100MB/s)" :value="0" />
                      <el-option label="1级 (100-200MB/s)" :value="1" />
                      <el-option label="2级 (200-400MB/s)" :value="2" />
                      <el-option label="3级 (400-600MB/s)" :value="3" />
                      <el-option label="4级 (600-1000MB/s)" :value="4" />
                      <el-option label="5级 (>1000MB/s)" :value="5" />
                    </el-select>
                  </el-form-item>
                  <el-form-item label="媒体解锁">
                    <el-select
                      v-model="searchForm.media_unlocks"
                      placeholder="选择支持的流媒体"
                      multiple
                      clearable
                    >
                      <el-option
                        v-for="media in mediaNames"
                        :key="media"
                        :label="media"
                        :value="media"
                      />
                    </el-select>
                  </el-form-item>
                </div>

                <div class="section-title">运营商特定选项</div>
                <!-- CM Parameters -->
                <el-tabs v-model="activeName" id="isp-tabs">
                  <el-tab-pane
                    label="中国移动"
                    name="cm"
                    class="filter-section"
                  >
                    <el-form-item label="回程线路">
                      <el-select
                        v-model="searchForm.cm_params.back_route"
                        placeholder="选择回程线路"
                        clearable
                      >
                        <el-option
                          v-for="route in backrouteTypes"
                          :key="route"
                          :label="route"
                          :value="route"
                        />
                      </el-select>
                    </el-form-item>
                    <el-form-item label="延迟 (ms)">
                      <el-input-number
                        v-model="searchForm.cm_params.latency"
                        :min="0"
                        placeholder="最大延迟"
                      />
                    </el-form-item>
                    <el-row :gutter="16">
                      <el-col :span="12">
                        <el-form-item label="下载 (Mbps)">
                          <el-input-number
                            v-model="searchForm.cm_params.min_download"
                            :min="0"
                            placeholder="最低"
                            style="width: 100%"
                          />
                        </el-form-item>
                      </el-col>
                      <el-col :span="12">
                        <el-form-item label-width="60px" label="至">
                          <el-input-number
                            v-model="searchForm.cm_params.max_download"
                            :min="0"
                            placeholder="最高"
                            style="width: 100%"
                          />
                        </el-form-item>
                      </el-col>
                    </el-row>
                    <el-row :gutter="16">
                      <el-col :span="12">
                        <el-form-item label="上传 (Mbps)">
                          <el-input-number
                            v-model="searchForm.cm_params.min_upload"
                            :min="0"
                            placeholder="最低"
                            style="width: 100%"
                          />
                        </el-form-item>
                      </el-col>
                      <el-col :span="12">
                        <el-form-item label-width="60px" label="至">
                          <el-input-number
                            v-model="searchForm.cm_params.max_upload"
                            :min="0"
                            placeholder="最高"
                            style="width: 100%"
                          />
                        </el-form-item>
                      </el-col>
                    </el-row>
                  </el-tab-pane>

                  <!-- CT Parameters -->
                  <el-tab-pane
                    label="中国电信"
                    name="ct"
                    class="filter-section"
                  >
                    <el-form-item label="回程线路">
                      <el-select
                        v-model="searchForm.ct_params.back_route"
                        placeholder="选择回程线路"
                        clearable
                      >
                        <el-option
                          v-for="route in backrouteTypes"
                          :key="route"
                          :label="route"
                          :value="route"
                        />
                      </el-select>
                    </el-form-item>
                    <el-form-item label="延迟 (ms)">
                      <el-input-number
                        v-model="searchForm.ct_params.latency"
                        :min="0"
                        placeholder="最大延迟"
                      />
                    </el-form-item>
                    <el-row :gutter="16">
                      <el-col :span="12">
                        <el-form-item label="下载 (Mbps)">
                          <el-input-number
                            v-model="searchForm.ct_params.min_download"
                            :min="0"
                            placeholder="最低"
                            style="width: 100%"
                          />
                        </el-form-item>
                      </el-col>
                      <el-col :span="12">
                        <el-form-item label-width="60px" label="至">
                          <el-input-number
                            v-model="searchForm.ct_params.max_download"
                            :min="0"
                            placeholder="最高"
                            style="width: 100%"
                          />
                        </el-form-item>
                      </el-col>
                    </el-row>
                    <el-row :gutter="16">
                      <el-col :span="12">
                        <el-form-item label="上传 (Mbps)">
                          <el-input-number
                            v-model="searchForm.ct_params.min_upload"
                            :min="0"
                            placeholder="最低"
                            style="width: 100%"
                          />
                        </el-form-item>
                      </el-col>
                      <el-col :span="12">
                        <el-form-item label-width="60px" label="至">
                          <el-input-number
                            v-model="searchForm.ct_params.max_upload"
                            :min="0"
                            placeholder="最高"
                            style="width: 100%"
                          />
                        </el-form-item>
                      </el-col>
                    </el-row>
                  </el-tab-pane>

                  <!-- CU Parameters -->
                  <el-tab-pane
                    label="中国联通"
                    name="cu"
                    class="filter-section"
                  >
                    <el-form-item label="回程线路">
                      <el-select
                        v-model="searchForm.cu_params.back_route"
                        placeholder="选择回程线路"
                        clearable
                      >
                        <el-option
                          v-for="route in backrouteTypes"
                          :key="route"
                          :label="route"
                          :value="route"
                        />
                      </el-select>
                    </el-form-item>
                    <el-form-item label="延迟 (ms)">
                      <el-input-number
                        v-model="searchForm.cu_params.latency"
                        :min="0"
                        placeholder="最大延迟"
                      />
                    </el-form-item>
                    <el-row :gutter="16">
                      <el-col :span="12">
                        <el-form-item label="下载 (Mbps)">
                          <el-input-number
                            v-model="searchForm.cu_params.min_download"
                            :min="0"
                            placeholder="最低"
                            style="width: 100%"
                          />
                        </el-form-item>
                      </el-col>
                      <el-col :span="12">
                        <el-form-item label-width="60px" label="至">
                          <el-input-number
                            v-model="searchForm.cu_params.max_download"
                            :min="0"
                            placeholder="最高"
                            style="width: 100%"
                          />
                        </el-form-item>
                      </el-col>
                    </el-row>
                    <el-row :gutter="16">
                      <el-col :span="12">
                        <el-form-item label="上传 (Mbps)">
                          <el-input-number
                            v-model="searchForm.cu_params.min_upload"
                            :min="0"
                            placeholder="最低"
                            style="width: 100%"
                          />
                        </el-form-item>
                      </el-col>
                      <el-col :span="12">
                        <el-form-item label-width="60px" label="至">
                          <el-input-number
                            v-model="searchForm.cu_params.max_upload"
                            :min="0"
                            placeholder="最高"
                            style="width: 100%"
                          />
                        </el-form-item>
                      </el-col>
                    </el-row>
                  </el-tab-pane>
                </el-tabs>

                <!-- Actions -->
                <el-button
                  type="primary"
                  @click="handleSearch"
                  :loading="loading"
                  >搜索</el-button
                >
                <el-button @click="handleReset">重置</el-button>
              </el-form>
            </div>
          </el-tab-pane>

          <!-- Results Panel -->
          <el-tab-pane label="搜索结果" name="results" :disabled="!hasSearched">
            <div class="results-section">
              <div class="results-header">搜索结果 (共 {{ total }} 条)</div>
              <el-skeleton :rows="5" animated v-if="loading" />
              <div v-else-if="reports.length === 0" class="no-results">
                未找到符合条件的测试记录
              </div>
              <template v-else>
                <el-card
                  v-for="report in reports"
                  :key="report.id"
                  shadow="never"
                  class="report-item"
                  @click="gotoDetail(report.id)"
                >
                  <div class="report-item-header">{{ report.name }}</div>
                  <div>创建时间: {{ report.date }}</div>
                </el-card>
                <el-pagination
                  v-if="total > 0"
                  v-model:current-page="page"
                  :disabled="disabled"
                  :background="false"
                  layout="total, prev, pager, next, jumper"
                  :total="total"
                  :page-size="pageSize"
                />
              </template>
            </div>
          </el-tab-pane>
        </el-tabs>
      </el-col>

      <!-- Sidebar -->
      <el-col :span="6" :offset="1">
        <Profile>
          <div>
            <div style="font-weight: 600; color: #303133">搜索提示</div>
            <div class="hint-item">所有筛选条件都是可选的</div>
            <div class="hint-item">支持多条件组合搜索</div>
            <div class="hint-item">速度单位: Mbps</div>
            <div class="hint-item">延迟单位: 毫秒(ms)</div>
            <div class="hint-item">磁盘等级对应读写速度</div>
          </div>
        </Profile>
      </el-col>
    </el-row>
  </div>
</template>

<style>
#search-root {
  width: 100%;
  padding: 16px;
  box-sizing: border-box;
}

#search-title {
  font-size: 28px;
  font-weight: 300;
  margin-bottom: 16px;
}

.search-form-card {
  margin-bottom: 16px;
}

.filter-section {
  padding-bottom: 16px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
}

.results-section {
  margin-top: 16px;
}

.results-header {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
}

.no-results {
  text-align: center;
  padding: 40px;
  color: #909399;
  font-size: 14px;
}

.report-item {
  margin-bottom: 16px;
  cursor: pointer;
}

.report-item:hover {
  border-color: var(--el-color-primary);
}

.report-item-header {
  font-weight: 600;
  font-size: 18px;
  margin-bottom: 8px;
}

.hint-item {
  margin-top: 8px;
  font-size: 14px;
  color: #606266;
}

#isp-tabs .el-tabs__nav-wrap::after {
  background-color: transparent;
}
</style>
