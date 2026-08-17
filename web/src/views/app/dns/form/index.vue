<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from "vue";
import ReCol from "@/components/ReCol";
import { useI18n } from "vue-i18n";
import { getTunnelList } from "@/api/tunnel";
import { getRuleList } from "@/api/rule";
import { message } from "@/utils/message";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import AddFill from "~icons/ri/add-line";
import Delete from "~icons/ep/delete";
import ArrowUp from "~icons/ep/arrow-up";
import ArrowDown from "~icons/ep/arrow-down";

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "新增 DNS",
    id: undefined,
    address: "",
    port: "5656",
    rules: "[]",
    hosts_text: "",
    hosts_json: "",
    tunnel_type: "quic",
    tunnel_id: "",
    tunnel_token: "",
    upstream_method: "round_robin",
    upstream_servers: JSON.stringify([{ target: "8.8.8.8:53", weight: 1 }]),
    remark: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const { t } = useI18n();

// Rules multi-select & ordering (L4 rules only for DNS Proxy)
const availableRules = ref<Array<{ label: string; value: string; desc?: string }>>([]);

function isL4Rule(r: any): boolean {
  try {
    const itemsStr = r.Items || r.items;
    if (!itemsStr) return true;
    const items = typeof itemsStr === "string" ? JSON.parse(itemsStr) : itemsStr;
    if (!Array.isArray(items)) return true;

    for (const item of items) {
      const mObj = item?.Matcher || item?.matcher || {};
      const aObj = item?.Action || item?.action || {};

      const mName = mObj?.Name || mObj?.name || "";
      const aName = aObj?.Name || aObj?.name || "";

      // Exclude HTTP application layer matchers
      if (["http_ip_matcher", "url_matcher", "js_matcher"].includes(mName)) {
        return false;
      }
      // Exclude HTTP application layer actions
      if ([
        "hide_version_action", "auth_portal_action", "response_text_action",
        "modify_status_action", "forward_request_action", "replace_request_body_action",
        "replace_response_body_action", "replace_request_header_action",
        "replace_response_header_action", "auth_guard_action", "insert_data_action",
        "subdomain_webvpn_action"
      ].includes(aName)) {
        return false;
      }
    }

    return true;
  } catch (e) {
    return false;
  }
}

async function fetchCustomRules() {
  try {
    const res = await getRuleList();
    const rulesList: Array<{ label: string; value: string; desc?: string }> = [];

    if (res?.code === 0 && res?.data?.list) {
      for (const r of res.data.list) {
        if (!isL4Rule(r)) continue;
        const ruleName = r.Name || r.name || `Rule #${r.Id || r.id}`;
        rulesList.push({
          label: ruleName,
          value: ruleName,
          desc: r.Remark || r.remark || ""
        });
      }
    }

    availableRules.value = rulesList;
  } catch (e) {
    availableRules.value = [];
  }
}

const selectedRules = ref<string[]>([]);

try {
  if (typeof newFormInline.value.rules === "string" && newFormInline.value.rules) {
    const parsed = JSON.parse(newFormInline.value.rules);
    if (Array.isArray(parsed)) selectedRules.value = parsed;
  } else if (Array.isArray(newFormInline.value.rules)) {
    selectedRules.value = newFormInline.value.rules;
  }
} catch (e) {
  selectedRules.value = [];
}

function handleRulesChange(vals: string[]) {
  selectedRules.value = vals;
  syncRulesJSON();
}

function moveRuleUp(index: number) {
  if (index <= 0) return;
  const temp = selectedRules.value[index];
  selectedRules.value[index] = selectedRules.value[index - 1];
  selectedRules.value[index - 1] = temp;
  syncRulesJSON();
}

function moveRuleDown(index: number) {
  if (index >= selectedRules.value.length - 1) return;
  const temp = selectedRules.value[index];
  selectedRules.value[index] = selectedRules.value[index + 1];
  selectedRules.value[index + 1] = temp;
  syncRulesJSON();
}

function syncRulesJSON() {
  newFormInline.value.rules = JSON.stringify(selectedRules.value);
}

// Tunnel selection
const tunnelOptions = ref<Array<any>>([]);
const tunnelLoading = ref(false);
const selectedTunnelKey = ref("");

function formatTunnelLabel(tItem: any) {
  if (!tItem) return "";
  const type = (tItem.type || tItem.Type || "QUIC").toUpperCase();
  const port = tItem.port ?? tItem.Port ?? "";
  const sni = tItem.sni ?? tItem.SNI ?? "";
  const id = tItem.id ?? tItem.Id ?? "";
  const sniText = sni ? ` | SNI: ${sni}` : "";
  return `[${type}] 端口: ${port}${sniText} (ID: ${id})`;
}

// Filter tunnel options for DNS mode (currently QUIC only; extensible for future DoH features)
const allowedTunnelTypes = ["quic"];
const filteredTunnelOptions = computed(() => {
  return tunnelOptions.value.filter(tItem => {
    const rawType = (tItem.type || tItem.Type || "").toLowerCase();
    return allowedTunnelTypes.some(allowed => rawType.includes(allowed));
  });
});

async function fetchTunnels() {
  tunnelLoading.value = true;
  const res = await getTunnelList();
  tunnelLoading.value = false;
  if (res?.code === 0 && Array.isArray(res?.data?.list)) {
    tunnelOptions.value = res.data.list;

    if (newFormInline.value.tunnel_id) {
      const matched = tunnelOptions.value.find(
        t => String(t.id || t.Id) === String(newFormInline.value.tunnel_id)
      );
      if (matched) {
        selectedTunnelKey.value = String(matched.id || matched.Id);
      }
    }
  }
}

function handleSelectTunnel(val: string) {
  if (!val) {
    newFormInline.value.tunnel_id = "";
    newFormInline.value.tunnel_type = "";
    return;
  }
  const matched = tunnelOptions.value.find(
    t => String(t.id || t.Id) === String(val)
  );
  if (matched) {
    const rawType = matched.type || matched.Type || "quic";
    newFormInline.value.tunnel_type = rawType.toLowerCase().includes("tls") ? "tls" : "quic";
    newFormInline.value.tunnel_id = String(matched.id || matched.Id);
  }
}

// Upstream servers table & strategy logic
type UpstreamItem = { target: string; weight: number };
const upstreamList = ref<UpstreamItem[]>([
  { target: "8.8.8.8:53", weight: 1 }
]);

try {
  if (typeof newFormInline.value.upstream_servers === "string" && newFormInline.value.upstream_servers) {
    const parsed = JSON.parse(newFormInline.value.upstream_servers);
    if (Array.isArray(parsed)) upstreamList.value = parsed;
  } else if (Array.isArray(newFormInline.value.upstream_servers)) {
    upstreamList.value = newFormInline.value.upstream_servers;
  }
} catch (e) {
  upstreamList.value = [{ target: "8.8.8.8:53", weight: 1 }];
}

// Check strategy mode
const isWeightDisabled = computed(() => {
  return newFormInline.value.upstream_method !== "weight";
});

function handleMethodChange(val: string) {
  if (val === "round_robin" || val === "ip_hash") {
    // Reset all row weights to 1 when strategy is round_robin or ip_hash
    upstreamList.value.forEach(row => {
      row.weight = 1;
    });
  }
  syncUpstreamJSON();
}

function addUpstreamRow() {
  upstreamList.value.push({ target: "8.8.8.8:53", weight: 1 });
  syncUpstreamJSON();
}

function removeUpstreamRow(index: number) {
  upstreamList.value.splice(index, 1);
  syncUpstreamJSON();
}

function syncUpstreamJSON() {
  newFormInline.value.upstream_servers = JSON.stringify(upstreamList.value);
}

watch(
  upstreamList,
  () => {
    syncUpstreamJSON();
  },
  { deep: true }
);

// Live Hosts parsing statistics
const hostsSummary = computed(() => {
  const text = newFormInline.value.hosts_text || "";
  let aCount = 0;
  let aaaaCount = 0;
  const lines = text.split("\n");
  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#") || trimmed.startsWith("//")) continue;
    const parts = trimmed.split(/\s+/);
    if (parts.length >= 2) {
      const ip = parts[0];
      const domains = parts.slice(1);
      if (ip.includes(":")) {
        aaaaCount += domains.length;
      } else {
        aCount += domains.length;
      }
    }
  }
  return { aCount, aaaaCount };
});

const formRules = reactive({
  port: [
    { required: true, message: () => t("dns.portRequired"), trigger: "blur" },
    {
      validator: (rule: any, value: any, callback: any) => {
        const num = Number(value);
        if (!value || isNaN(num) || num < 1 || num > 65535) {
          callback(new Error(t("dns.portFormatError")));
        } else {
          callback();
        }
      },
      trigger: "blur"
    }
  ]
});

// Custom validation for upstream / tunnel requirement before submit
function validateCustomBackend() {
  const hasTunnel = Boolean(newFormInline.value.tunnel_id);
  const validServers = upstreamList.value.filter(s => s.target && s.target.trim().length > 0);
  if (!hasTunnel && validServers.length === 0) {
    return false;
  }
  return true;
}

onMounted(() => {
  fetchTunnels();
  fetchCustomRules();
  syncRulesJSON();
  syncUpstreamJSON();
});

function getRef() {
  return {
    validate: (callback: (valid: boolean) => void) => {
      ruleFormRef.value.validate((valid: boolean) => {
        if (!valid) {
          callback(false);
          return;
        }
        if (!validateCustomBackend()) {
          message("请至少配置一个有效的 Upstream 上游服务器或选择 Tunnel 隧道！", { type: "warning" });
          callback(false);
          return;
        }
        callback(true);
      });
    }
  };
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :model="newFormInline"
    :rules="formRules"
    label-width="auto"
    class="py-2 px-3 space-y-6 max-h-[75vh] overflow-y-auto"
  >
    <!-- 板块 1: 基础监听与规则 -->
    <div class="space-y-4 border-b border-gray-100 dark:border-gray-700/80 pb-5">
      <div class="flex items-center space-x-2">
        <div class="w-1.5 h-4 bg-blue-500 rounded-full"></div>
        <span class="font-bold text-sm text-gray-800 dark:text-gray-200">
          {{ t('dns.baseInfoTab') }}
        </span>
      </div>

      <el-row :gutter="16">
        <re-col :value="12" :xs="24" :sm="12">
          <el-form-item :label="t('dns.port')" prop="port">
            <el-input
              v-model="newFormInline.port"
              placeholder="5656"
              clearable
            />
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24" :sm="12">
          <el-form-item :label="t('dns.address')" prop="address">
            <el-input
              v-model="newFormInline.address"
              placeholder="0.0.0.0"
              clearable
            />
          </el-form-item>
        </re-col>

        <!-- 中间件规则多选与排序 -->
        <re-col :value="24" :xs="24" :sm="24">
          <el-form-item :label="t('dns.rules')" prop="rules">
            <div class="w-full space-y-2">
              <el-select
                v-model="selectedRules"
                multiple
                clearable
                class="w-full"
                :placeholder="t('dns.selectRulesPlaceholder')"
                @change="handleRulesChange"
              >
                <el-option
                  v-for="r in availableRules"
                  :key="r.value"
                  :label="r.label"
                  :value="r.value"
                />
              </el-select>

              <div class="text-[11px] text-[var(--el-text-color-secondary)] mt-1 leading-relaxed">
                提示: DNS 代理属于传输层 (L4) 服务，下拉列表仅展示在“规则”菜单中配置的传输层 (L4) 中间件规则 (如 ip_matcher / reset_conn_action)，HTTP 应用层规则 (如 http_ip_matcher / hide_version_action) 不适用于 DNS。
              </div>

              <!-- 已选中规则的排序清单 -->
              <div v-if="selectedRules.length > 0" class="p-2.5 bg-gray-50 dark:bg-gray-800/60 rounded border border-gray-100 dark:border-gray-700/60 text-xs">
                <div class="text-gray-500 dark:text-gray-400 mb-1.5 font-medium flex justify-between items-center">
                  <span class="inline-flex items-center gap-1">
                    <IconifyIconOffline icon="ri:settings-3-line" />
                    {{ t('dns.ruleOrderTip') }}
                  </span>
                  <span class="font-mono text-gray-400">共 {{ selectedRules.length }} 项</span>
                </div>
                <div class="space-y-1">
                  <div
                    v-for="(ruleVal, idx) in selectedRules"
                    :key="ruleVal"
                    class="flex items-center justify-between p-2 bg-white dark:bg-gray-800 rounded border border-gray-200/80 dark:border-gray-600"
                  >
                    <span class="font-mono text-gray-800 dark:text-gray-200">
                      <span class="text-gray-400 font-bold mr-1.5">#{{ idx + 1 }}</span> {{ ruleVal }}
                    </span>
                    <div class="flex items-center space-x-1">
                      <el-button
                        size="small"
                        link
                        :disabled="idx === 0"
                        :icon="useRenderIcon(ArrowUp)"
                        @click="moveRuleUp(idx)"
                      />
                      <el-button
                        size="small"
                        link
                        :disabled="idx === selectedRules.length - 1"
                        :icon="useRenderIcon(ArrowDown)"
                        @click="moveRuleDown(idx)"
                      />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </el-form-item>
        </re-col>

        <!-- 备注 -->
        <re-col :value="24" :xs="24" :sm="24">
          <el-form-item :label="t('dns.remark')" prop="remark">
            <el-input
              v-model="newFormInline.remark"
              type="textarea"
              :rows="2"
              :placeholder="t('dns.remarkPlaceholder')"
              clearable
            />
          </el-form-item>
        </re-col>
      </el-row>
    </div>

    <!-- 板块 2: Hosts 域名映射 -->
    <div class="space-y-3 border-b border-gray-100 dark:border-gray-700/80 pb-5">
      <div class="flex items-center justify-between flex-wrap gap-2">
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-emerald-500 rounded-full"></div>
          <span class="font-bold text-sm text-gray-800 dark:text-gray-200">
            {{ t('dns.hostsTab') }}
          </span>
        </div>
        <div class="flex items-center space-x-2">
          <el-tag size="small" type="primary" effect="light" class="font-mono font-medium">
            A: {{ hostsSummary.aCount }}
          </el-tag>
          <el-tag size="small" type="success" effect="light" class="font-mono font-medium">
            AAAA: {{ hostsSummary.aaaaCount }}
          </el-tag>
        </div>
      </div>

      <div class="text-xs text-gray-500 dark:text-gray-400 font-medium inline-flex items-center gap-1">
        <IconifyIconOffline icon="ri:information-line" class="text-blue-500" />
        {{ t('dns.hostsTip') }}
      </div>

      <el-input
        v-model="newFormInline.hosts_text"
        type="textarea"
        :rows="5"
        :placeholder="t('dns.hostsPlaceholder')"
        class="font-mono text-xs"
      />
    </div>

    <!-- 板块 3: 上游配置 -->
    <div class="space-y-4">
      <div class="flex items-center space-x-2">
        <div class="w-1.5 h-4 bg-purple-500 rounded-full"></div>
        <span class="font-bold text-sm text-gray-800 dark:text-gray-200">
          {{ t('dns.backendTab') }}
        </span>
      </div>

      <!-- 1. Tunnel 隧道选择 -->
      <div class="space-y-2">
        <el-form-item :label="t('dns.tunnel')" prop="tunnel_id">
          <div class="w-full space-y-1">
            <el-select
              v-model="selectedTunnelKey"
              clearable
              filterable
              :loading="tunnelLoading"
              class="w-full"
              :placeholder="t('dns.selectTunnelPlaceholder')"
              @change="handleSelectTunnel"
            >
              <el-option
                v-for="tItem in filteredTunnelOptions"
                :key="tItem.id || tItem.Id"
                :label="formatTunnelLabel(tItem)"
                :value="String(tItem.id || tItem.Id)"
              />
            </el-select>
            <div class="text-xs text-gray-500 dark:text-gray-400 font-medium inline-flex items-center gap-1">
              <IconifyIconOffline icon="ri:information-line" class="text-blue-500" />
              {{ t('dns.tunnelTip') }}
            </div>
          </div>
        </el-form-item>
      </div>

      <!-- 2. Upstream DNS 服务器列表配置 -->
      <div class="space-y-3 pt-2 border-t border-gray-100 dark:border-gray-700/60">
        <el-form-item :label="t('dns.upstreamMethod')" prop="upstream_method">
          <el-select
            v-model="newFormInline.upstream_method"
            class="w-full"
            @change="handleMethodChange"
          >
            <el-option label="轮询 (round_robin)" value="round_robin" />
            <el-option label="权重 (weight)" value="weight" />
            <el-option label="IP Hash (ip_hash)" value="ip_hash" />
          </el-select>
        </el-form-item>

        <!-- 上游 DNS 服务器列表表格 -->
        <div class="space-y-2">
          <div class="flex justify-between items-center">
            <span class="text-xs font-medium text-gray-700 dark:text-gray-300">
              {{ t('dns.upstreamServers') }}
            </span>
            <el-button
              type="primary"
              size="small"
              :icon="useRenderIcon(AddFill)"
              @click="addUpstreamRow"
            >
              {{ t('dns.addServer') }}
            </el-button>
          </div>

          <el-table
            :data="upstreamList"
            border
            size="small"
            class="w-full text-xs"
            :empty-text="t('dns.noUpstreamServers')"
          >
            <el-table-column label="#" width="50" align="center">
              <template #default="{ $index }">
                <span class="font-mono text-gray-400">{{ $index + 1 }}</span>
              </template>
            </el-table-column>
            <el-table-column :label="t('dns.targetServer')" min-width="240">
              <template #default="{ row }">
                <el-input
                  v-model="row.target"
                  placeholder="8.8.8.8:53 / https://dns.google/dns-query"
                  size="small"
                  class="font-mono"
                />
              </template>
            </el-table-column>
            <el-table-column :label="t('dns.weight')" width="110" align="center">
              <template #default="{ row }">
                <el-input-number
                  v-model="row.weight"
                  :min="1"
                  :max="100"
                  :disabled="isWeightDisabled"
                  size="small"
                  controls-position="right"
                  class="w-full!"
                />
              </template>
            </el-table-column>
            <el-table-column :label="t('dns.operation')" width="70" align="center">
              <template #default="{ $index }">
                <el-button
                  type="danger"
                  link
                  size="small"
                  :icon="useRenderIcon(Delete)"
                  @click="removeUpstreamRow($index)"
                />
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </div>
  </el-form>
</template>

<style scoped lang="scss">
:deep(.el-form-item) {
  margin-bottom: 14px;
}
:deep(.el-form-item__label) {
  white-space: nowrap;
}
</style>
