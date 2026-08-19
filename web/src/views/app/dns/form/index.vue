<script setup lang="ts">
import { deviceDetection } from "@pureadmin/utils";
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
    name: "",
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
const availableRules = ref<
  Array<{ label: string; value: string; desc?: string }>
>([]);

function isL4Rule(r: any): boolean {
  try {
    const itemsStr = r.Items || r.items;
    if (!itemsStr) return true;
    const items =
      typeof itemsStr === "string" ? JSON.parse(itemsStr) : itemsStr;
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
      if (
        [
          "hide_version_action",
          "auth_portal_action",
          "response_text_action",
          "modify_status_action",
          "forward_request_action",
          "replace_request_body_action",
          "replace_response_body_action",
          "replace_request_header_action",
          "replace_response_header_action",
          "auth_guard_action",
          "insert_data_action",
          "subdomain_webvpn_action"
        ].includes(aName)
      ) {
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
    const rulesList: Array<{ label: string; value: string; desc?: string }> =
      [];

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
  if (
    typeof newFormInline.value.rules === "string" &&
    newFormInline.value.rules
  ) {
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
export interface DnsTunnelOption {
  label: string;
  value: string;
  tunnel_id: string;
  tunnel_token: string;
  tunnel_type: string;
  cName?: string;
  isOnline?: boolean;
  disabled?: boolean;
}

export interface DnsTunnelGroup {
  tunnel_id: string;
  groupLabel: string;
  options: DnsTunnelOption[];
}

const tunnelNodeGroups = ref<DnsTunnelGroup[]>([]);
const tunnelLoading = ref(false);
const selectedTunnelNodeKey = ref("");

async function fetchTunnels() {
  tunnelLoading.value = true;
  try {
    const res = await getTunnelList();
    let list: any[] = [];
    if (Array.isArray(res?.data?.list)) list = res.data.list;
    else if (Array.isArray(res?.data)) list = res.data;
    else if (Array.isArray(res)) list = res;

    // Filter ONLY QUIC tunnels for DNS proxy (DNS runs over UDP / QUIC)
    list = list.filter((tItem: any) => {
      const rawType = (tItem.Type || tItem.type || "").toLowerCase();
      return rawType.includes("quic");
    });

    const groups: DnsTunnelGroup[] = [];
    list.forEach((tItem: any) => {
      const tidStr = String(tItem.Id || tItem.id);
      const tName = tItem.Name || tItem.name || "";
      const tPort = tItem.Port || tItem.port || "";
      const tSni = tItem.SNI || tItem.sni || "";
      const tType = "quic";

      const portLabel = `${t("tunnel.port", "端口")}: ${tPort}`;
      const groupLabel = `${tName ? "[" + tName + "] " : ""}Tunnel #${tidStr} (${tType.toUpperCase()} | ${portLabel}${tSni ? " | SNI: " + tSni : ""})`;

      const nodeOpts: DnsTunnelOption[] = [];
      const cNodes = tItem.client_nodes || tItem.ClientNodes || [];

      if (Array.isArray(cNodes) && cNodes.length > 0) {
        cNodes.forEach((c: any) => {
          const isOnline = c.IsOnline ?? c.is_online ?? false;
          const cName = c.Name || c.name || "Node";
          const token = c.Token || c.token || "";
          const statusText = isOnline
            ? t("tunnelClient.online", "在线")
            : t("tunnelClient.offline", "离线");
          const key = `${tidStr}|${token}`;
          const nodeLabel = `[${statusText}] ${cName}`;
          nodeOpts.push({
            label: nodeLabel,
            value: key,
            tunnel_id: tidStr,
            tunnel_token: token,
            tunnel_type: tType,
            cName,
            isOnline,
            disabled: false
          });
        });
      } else {
        nodeOpts.push({
          label: t("tunnel.noClients", "暂无节点"),
          value: `${tidStr}||no_nodes`,
          tunnel_id: tidStr,
          tunnel_token: "",
          tunnel_type: tType,
          cName: t("tunnel.noClients", "暂无节点"),
          disabled: true
        });
      }

      groups.push({
        tunnel_id: tidStr,
        groupLabel,
        options: nodeOpts
      });
    });

    tunnelNodeGroups.value = groups;
    syncSelectedTunnelNodeKey();
  } catch (e) {
  } finally {
    tunnelLoading.value = false;
  }
}

function syncSelectedTunnelNodeKey() {
  const tid = newFormInline.value.tunnel_id;
  const token = newFormInline.value.tunnel_token;
  if (!tid) {
    selectedTunnelNodeKey.value = "";
    return;
  }

  let allOpts: DnsTunnelOption[] = [];
  tunnelNodeGroups.value.forEach(g => {
    allOpts = allOpts.concat(g.options);
  });

  const match = allOpts.find(
    o =>
      !o.disabled &&
      String(o.tunnel_id) === String(tid) &&
      String(o.tunnel_token || "") === String(token || "")
  );

  if (match) {
    selectedTunnelNodeKey.value = match.value;
  } else {
    const groupMatch = tunnelNodeGroups.value.find(
      g => String(g.tunnel_id) === String(tid)
    );
    const fallbackKey = `${tid}|${token || ""}`;
    const tType = newFormInline.value.tunnel_type || "quic";
    const fallbackName = token
      ? `Node-${token.length > 6 ? token.slice(-6) : token}`
      : t("tunnelClient.nodeRef", "节点");
    const fallbackOption: DnsTunnelOption = {
      label: `[${t("tunnelClient.unsavedId", "未存库")}] ${fallbackName}`,
      value: fallbackKey,
      tunnel_id: String(tid),
      tunnel_token: token || "",
      tunnel_type: tType,
      cName: fallbackName,
      isOnline: false,
      disabled: false
    };

    if (groupMatch) {
      groupMatch.options.push(fallbackOption);
    } else {
      tunnelNodeGroups.value.push({
        tunnel_id: String(tid),
        groupLabel: `Tunnel #${tid} (${t("tunnel.invalidAssoc", "失效")})`,
        options: [fallbackOption]
      });
    }
    selectedTunnelNodeKey.value = fallbackKey;
  }
}

function handleTunnelNodeChange(val: string) {
  if (!val) {
    newFormInline.value.tunnel_id = "";
    newFormInline.value.tunnel_token = "";
    newFormInline.value.tunnel_type = "";
    return;
  }
  let allOpts: DnsTunnelOption[] = [];
  tunnelNodeGroups.value.forEach(g => {
    allOpts = allOpts.concat(g.options);
  });
  const match = allOpts.find(o => o.value === val);
  if (match && !match.disabled) {
    newFormInline.value.tunnel_id = match.tunnel_id;
    newFormInline.value.tunnel_token = match.tunnel_token;
    newFormInline.value.tunnel_type = match.tunnel_type;
  }
}

// Upstream servers table & strategy logic
type UpstreamItem = { target: string; weight: number };
const upstreamList = ref<UpstreamItem[]>([{ target: "8.8.8.8:53", weight: 1 }]);

try {
  if (
    typeof newFormInline.value.upstream_servers === "string" &&
    newFormInline.value.upstream_servers
  ) {
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
    if (!trimmed || trimmed.startsWith("#") || trimmed.startsWith("//"))
      continue;
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
  name: [
    {
      required: true,
      message: () => t("dns.nameRequired", "请输入名称"),
      trigger: "blur"
    }
  ],
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

function validateCustomBackend() {
  const hasTunnel = Boolean(newFormInline.value.tunnel_id);
  const validServers = upstreamList.value.filter(
    s => s.target && s.target.trim().length > 0
  );
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
          message(
            "请至少配置一个有效的 Upstream 上游服务器或选择 Tunnel 隧道！",
            { type: "warning" }
          );
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
    :label-position="deviceDetection() ? 'top' : 'right'"
    ref="ruleFormRef"
    :model="newFormInline"
    :rules="formRules"
    label-width="auto"
    class="dns-form p-1 sm:px-2 space-y-4"
  >
    <!-- 板块 1: 基础监听与规则 -->
    <el-card
      shadow="never"
      class="border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-blue-500 rounded-full" />
          <span
            class="font-bold text-sm sm:text-base text-(--el-text-color-primary)"
          >
            {{ t("dns.baseInfoTab") }}
          </span>
        </div>
      </template>

      <el-row :gutter="16">
        <re-col :value="24" :xs="24">
          <el-form-item :label="t('dns.name', '名称')" prop="name">
            <el-input
              v-model="newFormInline.name"
              :placeholder="t('dns.namePlaceholder', '请输入 DNS 名称')"
              clearable
            />
          </el-form-item>
        </re-col>

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

              <div
                class="text-xs/relaxed text-(--el-text-color-secondary) mt-1"
              >
                提示: DNS 代理属于传输层 (L4)
                服务，下拉列表仅展示在“规则”菜单中配置的传输层 (L4) 中间件规则
                (如 ip_matcher / reset_conn_action)，HTTP 应用层规则不适用于
                DNS。
              </div>

              <!-- 已选中规则的排序清单 -->
              <div
                v-if="selectedRules.length > 0"
                class="p-2.5 bg-(--el-fill-color-light) rounded-lg border border-(--el-border-color-lighter) text-xs"
              >
                <div
                  class="text-(--el-text-color-secondary) mb-1.5 font-medium flex-bc"
                >
                  <span class="inline-flex items-center gap-1">
                    <IconifyIconOffline icon="ri:settings-3-line" />
                    {{ t("dns.ruleOrderTip") }}
                  </span>
                  <span class="font-mono"
                    >共 {{ selectedRules.length }} 项</span
                  >
                </div>
                <div class="space-y-1">
                  <div
                    v-for="(ruleVal, idx) in selectedRules"
                    :key="ruleVal"
                    class="flex-bc p-2 bg-(--el-bg-color) rounded-lg border border-(--el-border-color-lighter)"
                  >
                    <span class="font-mono text-(--el-text-color-primary)">
                      <span class="text-gray-400 font-bold mr-1.5"
                        >#{{ idx + 1 }}</span
                      >
                      {{ ruleVal }}
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
    </el-card>

    <!-- 板块 2: Hosts 域名映射 -->
    <el-card
      shadow="never"
      class="border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex-bc flex-wrap gap-2">
          <div class="flex items-center space-x-2">
            <div class="w-1.5 h-4 bg-emerald-500 rounded-full" />
            <span
              class="font-bold text-sm sm:text-base text-(--el-text-color-primary)"
            >
              {{ t("dns.hostsTab") }}
            </span>
          </div>
          <div class="flex items-center space-x-2">
            <el-tag
              size="small"
              type="primary"
              effect="light"
              class="font-mono font-medium"
            >
              A: {{ hostsSummary.aCount }}
            </el-tag>
            <el-tag
              size="small"
              type="success"
              effect="light"
              class="font-mono font-medium"
            >
              AAAA: {{ hostsSummary.aaaaCount }}
            </el-tag>
          </div>
        </div>
      </template>

      <div class="space-y-3">
        <div
          class="text-xs text-(--el-text-color-secondary) font-medium inline-flex items-center gap-1"
        >
          <IconifyIconOffline
            icon="ri:information-line"
            class="text-blue-500"
          />
          {{ t("dns.hostsTip") }}
        </div>

        <el-input
          v-model="newFormInline.hosts_text"
          type="textarea"
          :rows="5"
          :placeholder="t('dns.hostsPlaceholder')"
          class="font-mono text-xs"
        />
      </div>
    </el-card>

    <!-- 板块 3: 上游配置 -->
    <el-card
      shadow="never"
      class="border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-purple-500 rounded-full" />
          <span
            class="font-bold text-sm sm:text-base text-(--el-text-color-primary)"
          >
            {{ t("dns.backendTab") }}
          </span>
        </div>
      </template>

      <div class="space-y-4">
        <!-- 1. Tunnel 隧道选择 -->
        <el-form-item :label="t('dns.tunnel')" prop="tunnel_id">
          <div class="w-full space-y-1">
            <el-select
              v-model="selectedTunnelNodeKey"
              clearable
              filterable
              :loading="tunnelLoading"
              class="w-full"
              :placeholder="t('dns.selectTunnelPlaceholder')"
              @change="handleTunnelNodeChange"
            >
              <el-option-group
                v-for="group in tunnelNodeGroups"
                :key="group.tunnel_id"
                :label="group.groupLabel"
              >
                <el-option
                  v-for="item in group.options"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                  :disabled="item.disabled"
                >
                  <div
                    v-if="!item.disabled"
                    class="flex items-center space-x-2 py-0.5 text-xs"
                  >
                    <el-tag
                      size="small"
                      :type="item.isOnline ? 'success' : 'info'"
                      effect="light"
                      class="font-medium"
                    >
                      {{
                        item.isOnline
                          ? t("tunnelClient.online", "在线")
                          : t("tunnelClient.offline", "离线")
                      }}
                    </el-tag>
                    <span
                      class="font-semibold text-(--el-text-color-primary) font-mono"
                      >{{ item.cName || "Node" }}</span
                    >
                  </div>
                  <div v-else class="text-xs text-gray-400 py-0.5">
                    {{ item.cName }}
                  </div>
                </el-option>
              </el-option-group>
            </el-select>
            <div
              class="text-xs text-(--el-text-color-secondary) font-medium inline-flex items-center gap-1"
            >
              <IconifyIconOffline
                icon="ri:information-line"
                class="text-blue-500"
              />
              {{ t("dns.tunnelTip") }}
            </div>
          </div>
        </el-form-item>

        <!-- 2. Upstream DNS 服务器列表配置 -->
        <div class="space-y-3 pt-3 border-t border-(--el-border-color-lighter)">
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
            <div class="flex-bc">
              <span class="text-xs font-medium text-(--el-text-color-primary)">
                {{ t("dns.upstreamServers") }}
              </span>
              <el-button
                type="primary"
                size="small"
                :icon="useRenderIcon(AddFill)"
                @click="addUpstreamRow"
              >
                {{ t("dns.addServer") }}
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
              <el-table-column :label="t('dns.targetServer')" min-width="180">
                <template #default="{ row }">
                  <el-input
                    v-model="row.target"
                    placeholder="8.8.8.8:53"
                    size="small"
                    class="font-mono"
                  />
                </template>
              </el-table-column>
              <el-table-column
                :label="t('dns.weight')"
                width="100"
                align="center"
              >
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
              <el-table-column
                :label="t('dns.operation')"
                width="65"
                align="center"
              >
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
    </el-card>
  </el-form>
</template>

<style scoped lang="scss">
.dns-form :deep(.el-form-item) {
  margin-bottom: 14px;
}
</style>
