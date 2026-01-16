<script setup>
const { ipQuery, dataSourceOptions } = useTool();

const target = ref("");
const dataSource = ref("ipapi");
const loading = ref(false);
const result = ref(null);
const errorMessage = ref("");

const handleQuery = async () => {
  loading.value = true;
  errorMessage.value = "";
  result.value = null;

  try {
    result.value = await ipQuery(target.value, dataSource.value);
  } catch (error) {
    errorMessage.value = "查询IP信息时发生错误";
    return;
  } finally {
    loading.value = false;
  }
};

const handleReset = () => {
  target.value = "";
  dataSource.value = "ipapi";
  result.value = null;
  errorMessage.value = "";
};

const tableData = computed(() => {
  if (!result.value) return [];
  
  const data = [];
  const r = result.value;
  
  // 基本信息
  if (r.ip || r.data?.ip) {
    data.push({ category: '基本信息', field: 'IP地址', value: r.ip || r.data?.ip || '-' });
  }
  if (r.data?.hostname) {
    data.push({ category: '基本信息', field: '主机名', value: r.data.hostname });
  }
  if (r.rir) {
    data.push({ category: '基本信息', field: 'RIR', value: r.rir });
  }
  if (r.data?.is_anycast !== undefined) {
    data.push({ 
      category: '基本信息', 
      field: 'Anycast', 
      value: r.data.is_anycast ? '是' : '否',
      isTag: true,
      tagType: r.data.is_anycast ? 'success' : 'info'
    });
  }
  
  // 地理位置
  if (r.location?.country || r.data?.country) {
    data.push({ category: '地理位置', field: '国家', value: r.location?.country || r.data?.country || '-' });
  }
  if (r.location?.country_code || r.data?.country) {
    data.push({ category: '地理位置', field: '国家代码', value: r.location?.country_code || r.data?.country || '-' });
  }
  if (r.location?.state || r.data?.region) {
    data.push({ category: '地理位置', field: '州/省', value: r.location?.state || r.data?.region || '-' });
  }
  if (r.location?.city || r.data?.city) {
    data.push({ category: '地理位置', field: '城市', value: r.location?.city || r.data?.city || '-' });
  }
  if (r.location?.zip || r.data?.postal) {
    data.push({ category: '地理位置', field: '邮编', value: r.location?.zip || r.data?.postal });
  }
  if (r.location?.timezone || r.data?.timezone) {
    data.push({ category: '地理位置', field: '时区', value: r.location?.timezone || r.data?.timezone });
  }
  let coords = '-';
  if (r.location?.latitude && r.location?.longitude) {
    coords = `${r.location.latitude}, ${r.location.longitude}`;
  } else if (r.data?.loc) {
    coords = r.data.loc;
  }
  if (coords !== '-') {
    data.push({ category: '地理位置', field: '经纬度', value: coords });
  }
  
  // ASN信息
  if (r.asn?.asn || r.data?.asn?.asn) {
    data.push({ category: 'ASN信息', field: 'ASN', value: r.asn?.asn || r.data?.asn?.asn || '-' });
  }
  if (r.asn?.org || r.data?.asn?.name || r.data?.org) {
    data.push({ category: 'ASN信息', field: '组织', value: r.asn?.org || r.data?.asn?.name || r.data?.org || '-' });
  }
  if (r.asn?.domain || r.data?.asn?.domain) {
    data.push({ category: 'ASN信息', field: '域名', value: r.asn?.domain || r.data?.asn?.domain });
  }
  if (r.asn?.route || r.data?.asn?.route) {
    data.push({ category: 'ASN信息', field: '路由', value: r.asn?.route || r.data?.asn?.route || '-' });
  }
  if (r.asn?.type || r.data?.asn?.type) {
    data.push({ category: 'ASN信息', field: '类型', value: r.asn?.type || r.data?.asn?.type || '-' });
  }
  if (r.asn?.descr) {
    data.push({ category: 'ASN信息', field: '描述', value: r.asn.descr });
  }
  if (r.asn?.abuser_score) {
    data.push({ category: 'ASN信息', field: '滥用评分', value: r.asn.abuser_score });
  }
  
  // 公司信息
  if (r.company?.name || r.data?.company?.name) {
    data.push({ category: '公司信息', field: '名称', value: r.company?.name || r.data?.company?.name || '-' });
  }
  if (r.company?.domain || r.data?.company?.domain) {
    data.push({ category: '公司信息', field: '域名', value: r.company?.domain || r.data?.company?.domain || '-' });
  }
  if (r.company?.type || r.data?.company?.type) {
    data.push({ category: '公司信息', field: '类型', value: r.company?.type || r.data?.company?.type || '-' });
  }
  if (r.company?.network) {
    data.push({ category: '公司信息', field: '网络', value: r.company.network });
  }
  if (r.company?.abuser_score) {
    data.push({ category: '公司信息', field: '滥用评分', value: r.company.abuser_score });
  }
  
  // 数据中心
  if (r.datacenter?.datacenter) {
    data.push({ category: '数据中心', field: '数据中心', value: r.datacenter.datacenter });
  }
  if (r.datacenter?.domain) {
    data.push({ category: '数据中心', field: '域名', value: r.datacenter.domain });
  }
  if (r.datacenter?.network) {
    data.push({ category: '数据中心', field: '网络', value: r.datacenter.network });
  }
  
  // 滥用联系
  if (r.abuse?.name || r.data?.abuse?.name) {
    data.push({ category: '滥用联系', field: '名称', value: r.abuse?.name || r.data?.abuse?.name || '-' });
  }
  if (r.abuse?.email || r.data?.abuse?.email) {
    data.push({ category: '滥用联系', field: '邮箱', value: r.abuse?.email || r.data?.abuse?.email || '-' });
  }
  if (r.abuse?.phone || r.data?.abuse?.phone) {
    data.push({ category: '滥用联系', field: '电话', value: r.abuse?.phone || r.data?.abuse?.phone });
  }
  if (r.abuse?.address || r.data?.abuse?.address) {
    data.push({ category: '滥用联系', field: '地址', value: r.abuse?.address || r.data?.abuse?.address || '-' });
  }
  
  // 安全标识
  const isMobile = r.is_mobile || r.data?.is_mobile;
  data.push({ 
    category: '安全标识', 
    field: '移动网络', 
    value: isMobile ? '是' : '否',
    isTag: true,
    tagType: isMobile ? 'warning' : 'success'
  });
  
  const isDatacenter = r.is_datacenter || r.data?.is_hosting;
  data.push({ 
    category: '安全标识', 
    field: '数据中心', 
    value: isDatacenter ? '是' : '否',
    isTag: true,
    tagType: isDatacenter ? 'warning' : 'success'
  });
  
  const isProxy = r.is_proxy || r.data?.privacy?.proxy;
  data.push({ 
    category: '安全标识', 
    field: '代理', 
    value: isProxy ? '是' : '否',
    isTag: true,
    tagType: isProxy ? 'danger' : 'success'
  });
  
  const isVpn = r.is_vpn || r.data?.privacy?.vpn;
  data.push({ 
    category: '安全标识', 
    field: 'VPN', 
    value: isVpn ? '是' : '否',
    isTag: true,
    tagType: isVpn ? 'danger' : 'success'
  });
  
  const isTor = r.is_tor || r.data?.privacy?.tor;
  data.push({ 
    category: '安全标识', 
    field: 'Tor', 
    value: isTor ? '是' : '否',
    isTag: true,
    tagType: isTor ? 'danger' : 'success'
  });
  
  if (r.is_crawler !== undefined) {
    data.push({ 
      category: '安全标识', 
      field: '爬虫', 
      value: r.is_crawler ? '是' : '否',
      isTag: true,
      tagType: r.is_crawler ? 'warning' : 'success'
    });
  }
  
  if (r.is_satellite !== undefined || r.data?.is_satellite !== undefined) {
    const isSatellite = r.is_satellite || r.data?.is_satellite;
    data.push({ 
      category: '安全标识', 
      field: '卫星', 
      value: isSatellite ? '是' : '否',
      isTag: true,
      tagType: isSatellite ? 'warning' : 'success'
    });
  }
  
  if (r.is_abuser !== undefined) {
    data.push({ 
      category: '安全标识', 
      field: '滥用者', 
      value: r.is_abuser ? '是' : '否',
      isTag: true,
      tagType: r.is_abuser ? 'danger' : 'success'
    });
  }
  
  if (r.data?.is_anonymous !== undefined) {
    data.push({ 
      category: '安全标识', 
      field: '匿名', 
      value: r.data.is_anonymous ? '是' : '否',
      isTag: true,
      tagType: r.data.is_anonymous ? 'danger' : 'success'
    });
  }
  
  return data;
});

const spanMethod = ({ row, column, rowIndex, columnIndex }) => {
  if (columnIndex === 0) { // Category column
    const currentCategory = tableData.value[rowIndex].category;
    let rowspan = 1;
    
    // Count how many consecutive rows have the same category
    for (let i = rowIndex + 1; i < tableData.value.length; i++) {
      if (tableData.value[i].category === currentCategory) {
        rowspan++;
      } else {
        break;
      }
    }
    
    // Check if this is the first row of the category
    if (rowIndex === 0 || tableData.value[rowIndex - 1].category !== currentCategory) {
      return {
        rowspan: rowspan,
        colspan: 1
      };
    } else {
      // Hide this cell as it's already merged above
      return {
        rowspan: 0,
        colspan: 0
      };
    }
  }
};
</script>

<template>
  <div id="tool-root">
    <el-row>
      <el-col :span="24" id="tool-title"> IP查询 </el-col>
      <el-col :span="17">
        <div class="tool-form-card">
          <el-form label-width="120px" label-position="left">
            <el-form-item label="查询目标">
              <el-input
                v-model="target"
                placeholder="输入IP地址或域名 (留空为本机IP)"
                clearable
                @keyup.enter="handleQuery"
              />
            </el-form-item>
            <el-form-item label="数据源">
              <el-select v-model="dataSource" placeholder="选择数据源">
                <el-option
                  v-for="option in dataSourceOptions"
                  :key="option.value"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleQuery" :loading="loading">
                查询
              </el-button>
              <el-button @click="handleReset">重置</el-button>
            </el-form-item>
          </el-form>
        </div>

        <div v-if="errorMessage" class="error-message">
          <el-alert type="error" :closable="false">
            {{ errorMessage }}
          </el-alert>
        </div>

        <div v-if="result" class="result-section">
          <div class="result-header">查询结果</div>
          
          <el-table :data="tableData" border style="width: 100%" :show-header="true" :span-method="spanMethod">
            <el-table-column prop="category" label="类别" width="120" />
            <el-table-column prop="field" label="字段" width="180" />
            <el-table-column prop="value" label="值" />
          </el-table>

          <!-- Raw JSON (collapsible) -->
          <el-card shadow="never" class="result-card">
            <el-collapse>
              <el-collapse-item title="查看完整JSON响应" name="1">
                <pre class="result-content">{{ JSON.stringify(result, null, 2) }}</pre>
              </el-collapse-item>
            </el-collapse>
          </el-card>
        </div>
      </el-col>

      <el-col :span="6" :offset="1">
        <Profile>
          <div>
            <div style="font-weight: 600; color: #303133">关于IP查询</div>
            <div class="hint-item">查询IP地址信息</div>
            <div class="hint-item">不输入为本机IP</div>
            <div class="hint-item">提供多种数据源</div>
            <div class="hint-item">获取地理位置和ISP信息</div>
          </div>
        </Profile>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
#tool-root {
  width: 100%;
  padding: 16px;
  box-sizing: border-box;
}

#tool-title {
  font-size: 28px;
  font-weight: 300;
  margin-bottom: 16px;
}

.tool-form-card {
  margin-bottom: 16px;
}

.error-message {
  margin-bottom: 16px;
}

.result-section {
  margin-top: 16px;
}

.result-header {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
}

.result-card {
  margin-top: 16px;
}

.card-header {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.result-content {
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: "Courier New", monospace;
  font-size: 13px;
  color: #303133;
  max-height: 400px;
  overflow-y: auto;
}

.hint-item {
  margin-top: 8px;
  font-size: 14px;
  color: #606266;
}

#tool-root {
  overflow-y: auto;
}
</style>
