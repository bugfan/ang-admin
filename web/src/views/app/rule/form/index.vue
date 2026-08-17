<script setup lang="ts">
import { ref, reactive, watch, computed } from "vue";
import ReCol from "@/components/ReCol";
import { useI18n } from "vue-i18n";

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "新增规则",
    id: undefined,
    name: "",
    matcher: JSON.stringify({ Name: "ip_matcher", Config: { Address: ["127.0.0.1"] } }, null, 2),
    action: JSON.stringify({ Name: "reset_conn_action", Config: { Content: "Connection reset by rule" } }, null, 2),
    remark: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const { t } = useI18n();

// Supported Matcher options
const matcherOptions = [
  {
    label: "TCP / L4 IP 匹配器 (ip_matcher)",
    value: "ip_matcher",
    tag: "L4 传输层",
    tagType: "primary",
    desc: "匹配 TCP/UDP 客户端 IP 地址、CIDR 网段或 IP 范围"
  },
  {
    label: "HTTP Proxy IP 匹配器 (http_ip_matcher)",
    value: "http_ip_matcher",
    tag: "HTTP 应用层",
    tagType: "success",
    desc: "匹配 HTTP 请求客户端 IP 地址、CIDR 网段或 IP 范围"
  }
];

// Supported Action options
const actionOptions = [
  {
    label: "TCP 终止并重置连接 (reset_conn_action)",
    value: "reset_conn_action",
    tag: "L4 传输层",
    tagType: "danger",
    desc: "终止 TCP 会话连接，并在关闭前可选写入提示响应内容"
  },
  {
    label: "HTTP 隐藏 Server 版本 (hide_version_action)",
    value: "hide_version_action",
    tag: "HTTP 应用层",
    tagType: "warning",
    desc: "清除或重写 HTTP 响应头中的 Server 软件版本号"
  }
];

// Reactive states for Matcher Config
const selectedMatcher = ref("ip_matcher");
const matcherConfig = reactive<{
  ipAddress: string[];
  httpIpAddress: string[];
}>({
  ipAddress: ["127.0.0.1"],
  httpIpAddress: ["192.168.1.0/24"]
});

// Reactive states for Action Config
const selectedAction = ref("reset_conn_action");
const actionConfig = reactive<{
  resetContent: string;
}>({
  resetContent: "Connection reset by rule"
});

// Multiline text helpers for IP lists
const ipAddressText = computed({
  get: () => matcherConfig.ipAddress.join("\n"),
  set: (val: string) => {
    matcherConfig.ipAddress = val
      .split("\n")
      .map(s => s.trim())
      .filter(Boolean);
    syncToFormJSON();
  }
});

const httpIpAddressText = computed({
  get: () => matcherConfig.httpIpAddress.join("\n"),
  set: (val: string) => {
    matcherConfig.httpIpAddress = val
      .split("\n")
      .map(s => s.trim())
      .filter(Boolean);
    syncToFormJSON();
  }
});

onInitForm();

function onInitForm() {
  try {
    if (newFormInline.value.matcher) {
      const mObj = typeof newFormInline.value.matcher === "string" 
        ? JSON.parse(newFormInline.value.matcher) 
        : newFormInline.value.matcher;
      
      const mName = mObj?.Name || mObj?.name;
      const mCfg = mObj?.Config || mObj?.config || {};
      
      if (mName === "http_ip_matcher") {
        selectedMatcher.value = "http_ip_matcher";
        if (Array.isArray(mCfg.Address)) matcherConfig.httpIpAddress = mCfg.Address;
      } else {
        selectedMatcher.value = "ip_matcher";
        if (Array.isArray(mCfg.Address)) matcherConfig.ipAddress = mCfg.Address;
      }
    }
  } catch (e) {
    selectedMatcher.value = "ip_matcher";
  }

  try {
    if (newFormInline.value.action) {
      const aObj = typeof newFormInline.value.action === "string"
        ? JSON.parse(newFormInline.value.action)
        : newFormInline.value.action;
      
      const aName = aObj?.Name || aObj?.name;
      const aCfg = aObj?.Config || aObj?.config || {};

      if (aName === "hide_version_action") {
        selectedAction.value = "hide_version_action";
      } else {
        selectedAction.value = "reset_conn_action";
        if (aCfg.Content !== undefined) actionConfig.resetContent = aCfg.Content;
      }
    }
  } catch (e) {
    selectedAction.value = "reset_conn_action";
  }

  syncToFormJSON();
}

function buildMatcherJSONObj() {
  if (selectedMatcher.value === "http_ip_matcher") {
    return {
      Name: "http_ip_matcher",
      Config: { Address: matcherConfig.httpIpAddress }
    };
  }
  return {
    Name: "ip_matcher",
    Config: { Address: matcherConfig.ipAddress }
  };
}

function buildActionJSONObj() {
  if (selectedAction.value === "hide_version_action") {
    return {
      Name: "hide_version_action",
      Config: {}
    };
  }
  return {
    Name: "reset_conn_action",
    Config: { Content: actionConfig.resetContent }
  };
}

function syncToFormJSON() {
  const mObj = buildMatcherJSONObj();
  const aObj = buildActionJSONObj();
  newFormInline.value.matcher = JSON.stringify(mObj);
  newFormInline.value.action = JSON.stringify(aObj);
}

watch([selectedMatcher, matcherConfig], () => syncToFormJSON(), { deep: true });
watch([selectedAction, actionConfig], () => syncToFormJSON(), { deep: true });

const rules = reactive({
  name: [{ required: true, message: "请输入规则名称", trigger: "blur" }]
});

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :model="newFormInline"
    :rules="rules"
    label-width="110px"
    class="rule-form px-1 py-1"
  >
    <!-- Section 1: Basic Info -->
    <el-card shadow="never" class="mb-4 !border-[var(--el-border-color-lighter)] rounded-lg">
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1 h-3.5 bg-[var(--el-color-primary)] rounded-full"></div>
          <span class="font-bold text-[var(--el-text-color-primary)] text-sm">1. 规则基本属性</span>
        </div>
      </template>
      <el-row :gutter="20">
        <re-col :value="12" :xs="24">
          <el-form-item label="规则名称" prop="name">
            <el-input
              v-model="newFormInline.name"
              placeholder="请输入规则名称 (如: 阻断恶意IP访问)"
              clearable
            />
          </el-form-item>
        </re-col>
        <re-col :value="12" :xs="24">
          <el-form-item label="备注说明" prop="remark">
            <el-input
              v-model="newFormInline.remark"
              placeholder="选填"
              clearable
            />
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>

    <!-- Section 2: Matcher Selection -->
    <el-card shadow="never" class="mb-4 !border-[var(--el-border-color-lighter)] rounded-lg">
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <div class="w-1 h-3.5 bg-[var(--el-color-primary)] rounded-full"></div>
            <span class="font-bold text-[var(--el-text-color-primary)] text-sm">2. 匹配条件配置 (Matcher)</span>
          </div>
          <el-tag size="small" type="primary" effect="plain">Matcher</el-tag>
        </div>
      </template>

      <el-form-item label="选择 Matcher">
        <el-select
          v-model="selectedMatcher"
          placeholder="请选择 Matcher"
          class="w-full"
        >
          <el-option
            v-for="item in matcherOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          >
            <div class="flex items-center justify-between py-1">
              <div>
                <span class="font-medium text-[var(--el-text-color-primary)]">{{ item.label }}</span>
                <div class="text-xs text-[var(--el-text-color-secondary)]">{{ item.desc }}</div>
              </div>
              <el-tag :type="item.tagType as any" size="small" effect="light" class="ml-2">
                {{ item.tag }}
              </el-tag>
            </div>
          </el-option>
        </el-select>
      </el-form-item>

      <!-- Matcher Parameter Forms -->
      <div v-if="selectedMatcher === 'ip_matcher'" class="mt-3 p-3 rounded-md bg-[var(--el-fill-color-light)] border border-[var(--el-border-color-lighter)]">
        <el-form-item label="L4 IP 地址列表">
          <el-input
            v-model="ipAddressText"
            type="textarea"
            :rows="3"
            placeholder="每行一个 IP、CIDR 或 Range，例如:&#10;127.0.0.1&#10;192.168.1.0/24&#10;10.0.0.1-10.0.0.100"
          />
        </el-form-item>
        <div class="text-xs text-[var(--el-text-color-secondary)] pl-[110px]">
          提示: 用于 TCP/UDP 传输层匹配。支持精确 IP (1.2.3.4)、CIDR (192.168.1.0/24) 或范围 (10.0.0.1-10.0.0.50)
        </div>
      </div>

      <div v-if="selectedMatcher === 'http_ip_matcher'" class="mt-3 p-3 rounded-md bg-[var(--el-fill-color-light)] border border-[var(--el-border-color-lighter)]">
        <el-form-item label="HTTP IP 规则列表">
          <el-input
            v-model="httpIpAddressText"
            type="textarea"
            :rows="3"
            placeholder="每行一个 IP、CIDR 或 Range，例如:&#10;127.0.0.1&#10;192.168.1.0/24&#10;10.0.0.1-10.0.0.100"
          />
        </el-form-item>
        <div class="text-xs text-[var(--el-text-color-secondary)] pl-[110px]">
          提示: 用于 HTTP 代理应用层匹配。已在 ang 中注册为 http_ip_matcher
        </div>
      </div>
    </el-card>

    <!-- Section 3: Action Selection -->
    <el-card shadow="never" class="mb-4 !border-[var(--el-border-color-lighter)] rounded-lg">
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <div class="w-1 h-3.5 bg-[var(--el-color-success)] rounded-full"></div>
            <span class="font-bold text-[var(--el-text-color-primary)] text-sm">3. 触发动作配置 (Action)</span>
          </div>
          <el-tag size="small" type="success" effect="plain">Action</el-tag>
        </div>
      </template>

      <el-form-item label="选择 Action">
        <el-select
          v-model="selectedAction"
          placeholder="请选择 Action"
          class="w-full"
        >
          <el-option
            v-for="item in actionOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          >
            <div class="flex items-center justify-between py-1">
              <div>
                <span class="font-medium text-[var(--el-text-color-primary)]">{{ item.label }}</span>
                <div class="text-xs text-[var(--el-text-color-secondary)]">{{ item.desc }}</div>
              </div>
              <el-tag :type="item.tagType as any" size="small" effect="light" class="ml-2">
                {{ item.tag }}
              </el-tag>
            </div>
          </el-option>
        </el-select>
      </el-form-item>

      <!-- Action Parameter Forms -->
      <div v-if="selectedAction === 'reset_conn_action'" class="mt-3 p-3 rounded-md bg-[var(--el-fill-color-light)] border border-[var(--el-border-color-lighter)]">
        <el-form-item label="重置响应消息">
          <el-input
            v-model="actionConfig.resetContent"
            placeholder="关闭 TCP 连接前向客户端写入的可选提示文本"
          />
        </el-form-item>
        <div class="text-xs text-[var(--el-text-color-secondary)] pl-[110px]">
          提示: 适用于 TCP 会话 (reset_conn_action)，终止连接前可选发送提示信息
        </div>
      </div>

      <div v-if="selectedAction === 'hide_version_action'" class="mt-3 p-3 rounded-md bg-[var(--el-fill-color-light)] border border-[var(--el-border-color-lighter)]">
        <el-alert
          title="HTTP 隐藏版本号动作 (hide_version_action) 无需配置额外参数，生效后会自动将 Server 响应头重写为 '***'。"
          type="info"
          :closable="false"
          show-icon
        />
      </div>
    </el-card>

    <!-- Theme-aware JSON Preview Box -->
    <div class="rounded-lg border border-[var(--el-border-color-lighter)] bg-[var(--el-fill-color-light)] p-3">
      <div class="flex items-center justify-between mb-2">
        <span class="text-xs font-semibold text-[var(--el-text-color-secondary)]">实时序列化 JSON 存库数据结构预览</span>
        <el-tag size="small" type="info" effect="plain" class="font-mono">JSON Preview</el-tag>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div class="bg-[var(--el-bg-color)] p-2.5 rounded border border-[var(--el-border-color-lighter)]">
          <div class="text-xs font-bold text-[var(--el-color-primary)] mb-1">Matcher JSON:</div>
          <pre class="text-[11px] text-[var(--el-text-color-primary)] font-mono whitespace-pre-wrap break-all max-h-28 overflow-y-auto leading-relaxed">{{ newFormInline.matcher }}</pre>
        </div>
        <div class="bg-[var(--el-bg-color)] p-2.5 rounded border border-[var(--el-border-color-lighter)]">
          <div class="text-xs font-bold text-[var(--el-color-success)] mb-1">Action JSON:</div>
          <pre class="text-[11px] text-[var(--el-text-color-primary)] font-mono whitespace-pre-wrap break-all max-h-28 overflow-y-auto leading-relaxed">{{ newFormInline.action }}</pre>
        </div>
      </div>
    </div>
  </el-form>
</template>

<style scoped>
.rule-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--el-text-color-regular);
}
</style>
