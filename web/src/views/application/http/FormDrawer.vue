<script setup lang="ts">
import { ref, reactive, watch, computed, onMounted, onUnmounted } from "vue";
import { ElMessage } from "element-plus";

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  rowData: { type: Object, default: () => ({}) }
});

const emit = defineEmits(["update:modelValue", "submit"]);

const visible = ref(false);
const windowWidth = ref(window.innerWidth);

// 响应式抽屉宽度
const drawerSize = computed(() => {
  return windowWidth.value < 768 ? "100%" : "60%";
});

const handleResize = () => {
  windowWidth.value = window.innerWidth;
};

onMounted(() => {
  window.addEventListener("resize", handleResize);
});
onUnmounted(() => {
  window.removeEventListener("resize", handleResize);
});

const formData = reactive({
  id: "",
  name: "",
  type: "HTTP",
  status: 1,
  hostname: "",
  port: 443,
  tls: true,
  certificate: "",
  config: {
    Front: {
      ProxyHeaders: [],
      HTTP: true,
      H2: true,
      HSTS: false,
      Protected: false
    },
    Feature: {
      Compress: false
    },
    Rule: [],
    Backend: {
      RealIp: "",
      Tunnel: "",
      DNSResolver: "",
      Location: []
    }
  }
});

const portOptions = [
  { value: 80, label: "80 (HTTP)" },
  { value: 443, label: "443 (HTTPS)" },
  { value: 8080, label: "8080 (Alt HTTP)" },
  { value: 8443, label: "8443 (Alt HTTPS)" }
];

const headerOptions = ["Host", "X-Forwarded-For", "X-Real-IP", "Upgrade", "Connection"];

const matcherOptions = [
  { value: "always_true_matcher", label: "无条件匹配 (Always True)" },
  { value: "url_matcher", label: "URL 匹配 (URL Matcher)" },
  { value: "js_matcher", label: "JS 脚本匹配 (JS Matcher)" }
];

const actionOptions = [
  { value: "response_text_action", label: "自定义响应" },
  { value: "hide_version_action", label: "隐藏版本号" },
  { value: "replace_response_body_action", label: "替换响应Body" },
  { value: "replace_request_body_action", label: "替换请求Body" },
  { value: "replace_request_header_action", label: "替换请求头" },
  { value: "replace_response_header_action", label: "替换响应头" },
  { value: "modify_status_action", label: "修改状态码" }
];

watch(() => props.modelValue, (val) => {
  visible.value = val;
  if (val) {
    formData.id = props.rowData.id || "";
    formData.name = props.rowData.name || "";
    formData.hostname = props.rowData.hostname || "";
    formData.port = props.rowData.port || 443;
    formData.tls = props.rowData.tls ?? true;
    formData.certificate = props.rowData.certificate || "";
    formData.status = props.rowData.status ?? 1;

    try {
      if (props.rowData.config_json) {
        const parsed = JSON.parse(props.rowData.config_json);
        formData.config = Object.assign(formData.config, parsed);
      }
    } catch (e) {
      console.warn("Parse config_json error", e);
    }
  }
});

watch(() => visible.value, (val) => {
  emit("update:modelValue", val);
});

const addRule = () => {
  formData.config.Rule.push({
    Matcher: { Name: "always_true_matcher", Config: {} },
    Action: { Name: "hide_version_action", Config: {} }
  });
};
const removeRule = (index: number) => {
  formData.config.Rule.splice(index, 1);
};
const moveRuleUp = (index: number) => {
  if (index > 0) {
    const temp = formData.config.Rule[index - 1];
    formData.config.Rule[index - 1] = formData.config.Rule[index];
    formData.config.Rule[index] = temp;
  }
};
const moveRuleDown = (index: number) => {
  if (index < formData.config.Rule.length - 1) {
    const temp = formData.config.Rule[index + 1];
    formData.config.Rule[index + 1] = formData.config.Rule[index];
    formData.config.Rule[index] = temp;
  }
};

const addLocation = () => {
  formData.config.Backend.Location.push({
    Path: "/",
    Upstream: {
      Type: "proxy_pass",
      Data: {
        Method: "weight",
        Servers: [{ Target: "https://", Weight: 1 }]
      }
    }
  });
};
const removeLocation = (index: number) => {
  formData.config.Backend.Location.splice(index, 1);
};
const addServer = (locationIndex: number) => {
  formData.config.Backend.Location[locationIndex].Upstream.Data.Servers.push({ Target: "", Weight: 1 });
};
const removeServer = (locationIndex: number, serverIndex: number) => {
  formData.config.Backend.Location[locationIndex].Upstream.Data.Servers.splice(serverIndex, 1);
};

const handleSave = () => {
  const submitData = {
    id: formData.id,
    name: formData.name,
    type: "HTTP",
    hostname: formData.hostname,
    port: formData.port,
    tls: formData.tls,
    certificate: formData.certificate,
    status: formData.status,
    config_json: JSON.stringify(formData.config)
  };
  emit("submit", submitData);
};
</script>

<template>
  <el-drawer v-model="visible" :title="formData.id ? '编辑 HTTP 代理' : '新增 HTTP 代理'" :size="drawerSize" class="responsive-drawer">
    <!-- label-position="top" 对移动端更友好，避免文字换行挤压 -->
    <el-form :model="formData" label-position="top" class="complex-form">
      
      <!-- 基础与前端设置 -->
      <fieldset class="border border-dashed border-gray-300 rounded-lg p-4 mb-6">
        <legend class="px-2 font-bold text-blue-600">基础设置 (Front & Feature)</legend>
        <el-row :gutter="16">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="配置名称">
              <el-input v-model="formData.name" placeholder="请输入配置名称" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="域名 (Hostname)">
              <el-input v-model="formData.hostname" placeholder="支持通配符, 如: *.example.com" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="监听端口">
              <el-select-v2 v-model="formData.port" :options="portOptions" filterable allow-create placeholder="选择或输入" class="w-full" />
            </el-form-item>
          </el-col>
          
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="代理请求头 (ProxyHeaders)">
              <el-select-v2 v-model="formData.config.Front.ProxyHeaders" :options="headerOptions.map(h => ({value: h, label: h}))" filterable allow-create multiple placeholder="选择或输入" class="w-full" />
            </el-form-item>
          </el-col>
          
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="启用状态">
              <el-switch v-model="formData.status" :active-value="1" :inactive-value="0" />
            </el-form-item>
          </el-col>

          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="访问保护 (Protected)">
              <el-switch v-model="formData.config.Front.Protected" />
            </el-form-item>
          </el-col>

          <!-- TLS 与 HTTP 特性 -->
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="开启 TLS">
              <el-switch v-model="formData.tls" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8" v-if="formData.tls">
            <el-form-item label="证书 ID">
              <el-input v-model="formData.certificate" placeholder="请输入证书 ID" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="开启 HTTP/2">
              <el-switch v-model="formData.config.Front.H2" />
            </el-form-item>
          </el-col>

          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="数据压缩 (Compress)">
              <el-checkbox v-model="formData.config.Feature.Compress">启用</el-checkbox>
            </el-form-item>
          </el-col>
        </el-row>
      </fieldset>

      <!-- 规则设置 (Rule) -->
      <fieldset class="border border-dashed border-gray-300 rounded-lg p-4 mb-6">
        <legend class="px-2 font-bold text-green-600">规则设置 (Rule)</legend>
        
        <div class="mb-4 flex justify-between items-center flex-wrap gap-2">
          <span class="text-gray-500 text-xs sm:text-sm">支持多规则组合，按先后顺序匹配优先级</span>
          <el-button type="success" plain @click="addRule" icon="Plus" size="small">追加规则</el-button>
        </div>
        
        <div v-for="(rule, index) in formData.config.Rule" :key="index" class="bg-gray-50 rounded p-3 sm:p-4 mb-4 border border-gray-200">
          <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-3 pb-2 border-b border-gray-200 gap-2">
            <span class="font-bold text-gray-700">规则 #{{ index + 1 }}</span>
            <div class="flex gap-1">
              <el-button size="small" @click="moveRuleUp(index)" :disabled="index === 0" icon="Top" title="上移"></el-button>
              <el-button size="small" @click="moveRuleDown(index)" :disabled="index === formData.config.Rule.length - 1" icon="Bottom" title="下移"></el-button>
              <el-button size="small" type="danger" @click="removeRule(index)" icon="Delete" title="删除"></el-button>
            </div>
          </div>
          
          <el-row :gutter="16">
            <!-- 匹配器部分 -->
            <el-col :xs="24" :md="12" class="mb-4 md:mb-0">
              <div class="font-bold text-indigo-600 mb-2 border-l-2 border-indigo-500 pl-2">匹配条件 (Matcher)</div>
              <el-form-item label="匹配类型">
                <el-select v-model="rule.Matcher.Name" class="w-full">
                  <el-option v-for="opt in matcherOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
                </el-select>
              </el-form-item>
              
              <el-row :gutter="8" v-if="rule.Matcher.Name === 'url_matcher'">
                <el-col :span="10">
                  <el-form-item label="Method">
                    <el-input v-model="rule.Matcher.Config.Method" placeholder="ALL" />
                  </el-form-item>
                </el-col>
                <el-col :span="14">
                  <el-form-item label="Path">
                    <el-input v-model="rule.Matcher.Config.Path" placeholder="如: /api" />
                  </el-form-item>
                </el-col>
              </el-row>
              
              <el-form-item v-if="rule.Matcher.Name === 'js_matcher'" label="执行脚本 (Script)">
                <el-input v-model="rule.Matcher.Config.Script" type="textarea" placeholder="JS 匹配脚本" />
              </el-form-item>
            </el-col>
            
            <!-- 动作部分 -->
            <el-col :xs="24" :md="12">
              <div class="font-bold text-orange-600 mb-2 border-l-2 border-orange-500 pl-2">执行动作 (Action)</div>
              <el-form-item label="动作类型">
                <el-select v-model="rule.Action.Name" class="w-full">
                  <el-option v-for="opt in actionOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
                </el-select>
              </el-form-item>
              
              <el-row :gutter="8" v-if="rule.Action.Name === 'response_text_action'">
                <el-col :span="8">
                  <el-form-item label="Status">
                    <el-input v-model="rule.Action.Config.Code" placeholder="200" />
                  </el-form-item>
                </el-col>
                <el-col :span="16">
                  <el-form-item label="响应内容 (Content)">
                    <el-input v-model="rule.Action.Config.Content" placeholder="自定义内容" />
                  </el-form-item>
                </el-col>
              </el-row>
              
              <el-form-item v-if="rule.Action.Name === 'modify_status_action'" label="目标状态码">
                <el-input-number v-model="rule.Action.Config.Code" />
              </el-form-item>
              
              <el-form-item v-if="rule.Action.Name.includes('replace_')" label="替换映射 (Map JSON)">
                <el-input v-model="rule.Action.Config.Map" type="textarea" :rows="2" placeholder="{ &quot;old&quot;: &quot;new&quot; }" />
              </el-form-item>
            </el-col>
          </el-row>
        </div>
      </fieldset>

      <!-- 上游设置 (Backend) -->
      <fieldset class="border border-dashed border-gray-300 rounded-lg p-4 mb-6">
        <legend class="px-2 font-bold text-purple-600">上游设置 (Backend)</legend>
        
        <el-row :gutter="16">
          <el-col :xs="24" :sm="8">
            <el-form-item label="真实IP传递 (RealIp)">
              <el-input v-model="formData.config.Backend.RealIp" placeholder="Header Name" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="8">
            <el-form-item label="内网穿透 (Tunnel)">
              <el-select v-model="formData.config.Backend.Tunnel" clearable placeholder="留空则不使用" class="w-full">
                <el-option label="Quic Tunnel 1" value="tunnel_1" />
                <el-option label="TCP Tunnel 2" value="tunnel_2" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="8">
            <el-form-item label="DNS 解析器">
              <el-select v-model="formData.config.Backend.DNSResolver" clearable placeholder="系统默认" class="w-full">
                <el-option label="内部解析 dns1" value="dns1" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <div class="mt-4 mb-2 flex justify-between items-center flex-wrap gap-2">
          <span class="font-bold text-gray-700">路由匹配节点 (Location)</span>
          <el-button type="primary" plain @click="addLocation" icon="Plus" size="small">添加路由</el-button>
        </div>
        
        <div v-for="(loc, lIdx) in formData.config.Backend.Location" :key="lIdx" class="bg-gray-100 rounded p-3 sm:p-4 mb-4 border-l-4 border-purple-400">
          <div class="flex justify-between items-center mb-3 pb-2 border-b border-gray-300">
            <span class="font-medium">路由 #{{ lIdx + 1 }}</span>
            <el-button type="danger" link @click="removeLocation(lIdx)" icon="Close">移除此路由</el-button>
          </div>
          
          <el-row :gutter="16">
            <el-col :xs="24" :sm="8">
              <el-form-item label="匹配路径 (Path)">
                <el-input v-model="loc.Path" placeholder="如: /" />
              </el-form-item>
            </el-col>
            <el-col :xs="12" :sm="8">
              <el-form-item label="上游类型">
                <el-select v-model="loc.Upstream.Type" class="w-full">
                  <el-option label="反代 (proxy_pass)" value="proxy_pass" />
                  <el-option label="静态 (static_dir)" value="static_dir" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :xs="12" :sm="8" v-if="loc.Upstream.Type === 'proxy_pass'">
              <el-form-item label="负载策略">
                <el-select v-model="loc.Upstream.Data.Method" class="w-full">
                  <el-option label="权重 (weight)" value="weight" />
                  <el-option label="轮询 (roundrobin)" value="roundrobin" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>

          <!-- 节点列表 -->
          <div v-if="loc.Upstream.Type === 'proxy_pass'" class="bg-white p-3 rounded mt-2 border border-gray-200">
            <div class="text-sm font-medium mb-2 text-gray-500">后端目标节点列表 (Servers)</div>
            <div v-for="(server, sIdx) in loc.Upstream.Data.Servers" :key="sIdx" class="flex gap-2 mb-2 items-center">
              <el-input v-model="server.Target" placeholder="Target, 如 https://..." class="flex-grow" />
              <el-input-number v-model="server.Weight" :min="1" :controls="false" placeholder="权重" style="width: 80px" />
              <el-button type="danger" plain icon="Delete" @click="removeServer(lIdx, sIdx)" />
            </div>
            <el-button type="primary" link icon="Plus" @click="addServer(lIdx)">新增目标</el-button>
          </div>
          
          <div v-if="loc.Upstream.Type === 'static_dir'" class="bg-white p-3 rounded mt-2 border border-gray-200">
            <el-form-item label="本地目录路径" class="mb-0">
              <el-input v-model="loc.Upstream.Data.Dir" placeholder="绝对或相对路径" />
            </el-form-item>
          </div>
        </div>
      </fieldset>

    </el-form>

    <template #footer>
      <div class="flex justify-end pt-4">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存全部配置</el-button>
      </div>
    </template>
  </el-drawer>
</template>

<style scoped>
.complex-form .el-form-item {
  margin-bottom: 12px;
}
:deep(.el-drawer__body) {
  padding-bottom: 80px; /* 留出底部操作区空间 */
}
</style>
