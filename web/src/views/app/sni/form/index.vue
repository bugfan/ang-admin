<script setup lang="ts">
import { ref, onMounted, computed, watch } from "vue";
import ReCol from "@/components/ReCol";
import { useI18n } from "vue-i18n";
import { getTunnelList } from "@/api/tunnel";
import { getRuleList } from "@/api/rule";
import { message } from "@/utils/message";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import AddFill from "~icons/ri/add-line";
import Delete from "~icons/ep/delete";
import Rank from "~icons/ep/rank";
import Sortable from "sortablejs";

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    id: undefined,
    name: "",
    sni: "",
    extra_sni: "",
    port: "443",
    rules: "[]",
    tunnel_type: "quic",
    tunnel_id: "",
    tunnel_token: "",
    dns_resolver: "",
    remark: ""
  })
});

export interface TunnelGroupOption {
  label: string;
  value: string;
  tunnel_id: string;
  tunnel_token: string;
  tunnel_type: string;
  cName?: string;
  isOnline?: boolean;
  disabled?: boolean;
}

export interface TunnelGroup {
  tunnel_id: string;
  groupLabel: string;
  options: TunnelGroupOption[];
}

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const { t } = useI18n();

// Rules multi-select & ordering (L4 rules only for SNI Proxy)
const availableRules = ref<
  Array<{ label: string; value: string; desc?: string }>
>([]);
const selectedRules = ref<string[]>([]);
const tunnelNodeGroups = ref<TunnelGroup[]>([]);
const selectedTunnelNodeKey = ref<string>("");

// --- ExtraSNI textarea binding ---
// extraSniText holds the raw textarea text (one pattern per line).
// It is kept in sync with newFormInline.value.extra_sni (JSON array string).
const extraSniText = ref<string>("");

/** Convert a JSON-encoded []string or newline list to textarea text. */
function extraSniFromJSON(raw: string): string {
  if (!raw) return "";
  try {
    const arr = JSON.parse(raw);
    if (Array.isArray(arr)) return arr.join("\n");
  } catch {
    // already plain text
  }
  return raw;
}

/** Convert textarea text (one pattern per line) to JSON array string. */
function extraSniToJSON(text: string): string {
  const lines = text
    .split("\n")
    .map(l => l.trim())
    .filter(l => l.length > 0);
  return lines.length > 0 ? JSON.stringify(lines) : "";
}

function syncExtraSniToForm() {
  newFormInline.value.extra_sni = extraSniToJSON(extraSniText.value);
}

function initExtraSniFromProps() {
  extraSniText.value = extraSniFromJSON(newFormInline.value.extra_sni || "");
}

let nextDnsId = 1;
interface DnsItem {
  id: number;
  value: string;
}
const dnsResolverList = ref<DnsItem[]>([]);

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
          "response_text_action",
          "modify_status_action",
          "replace_request_body_action",
          "replace_response_body_action",
          "replace_request_header_action",
          "replace_response_header_action",
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

function addDnsResolver() {
  if (dnsResolverList.value.some(item => item.value.trim() === "")) {
    message(t("common.dnsEmptyExists", "已存在默认解析(空项)，无需重复添加"), { type: "warning" });
    return;
  }
  dnsResolverList.value.push({
    id: ++nextDnsId,
    value: ""
  });
  syncDnsResolver();
}

function removeDnsResolver(idx: number) {
  dnsResolverList.value.splice(idx, 1);
  syncDnsResolver();
}

const dnsError = ref("");

function syncDnsResolver() {
  let hasEmpty = false;
  const activeList = [];
  const nonEmpties: string[] = [];
  for (const item of dnsResolverList.value) {
    const val = item.value.trim();
    if (val === "") {
      if (!hasEmpty) {
        hasEmpty = true;
        activeList.push(val);
      }
    } else {
      activeList.push(val);
      nonEmpties.push(val);
    }
  }
  newFormInline.value.dns_resolver = JSON.stringify(activeList);

  const duplicates = nonEmpties.filter((item, idx) => nonEmpties.indexOf(item) !== idx);
  if (duplicates.length > 0) {
    dnsError.value = t(
      "common.upstreamTargetDuplicate",
      { target: duplicates[0] },
      `上游列表中存在重复的目标服务器地址 [${duplicates[0]}]`
    );
  } else {
    dnsError.value = "";
  }
}

function initDnsResolverFromProps() {
  try {
    if (newFormInline.value.dns_resolver) {
      let parsed = newFormInline.value.dns_resolver;
      if (typeof parsed === "string") {
        if (parsed.startsWith("[")) {
          parsed = JSON.parse(parsed);
        } else {
          parsed = parsed.split(/[\n,;]+/).map((s: string) => s.trim()).filter(Boolean);
        }
      }
      if (Array.isArray(parsed) && parsed.length > 0) {
        dnsResolverList.value = parsed.map((v: string) => ({
          id: ++nextDnsId,
          value: v
        }));
        syncDnsResolver();
        return;
      }
    }
  } catch (e) {}

  dnsResolverList.value = [
    {
      id: ++nextDnsId,
      value: ""
    }
  ];
  syncDnsResolver();
}

function getRef() {
  syncDnsResolver();
  syncExtraSniToForm();
  return {
    validate: (callback: (valid: boolean) => void) => {
      ruleFormRef.value.validate((valid: boolean) => {
        const dnsVals = dnsResolverList.value.map(d => d.value.trim()).filter(Boolean);
        const dupDns = dnsVals.filter((item, idx) => dnsVals.indexOf(item) !== idx);
        if (dupDns.length > 0) {
          const errMsg = t(
            "common.upstreamTargetDuplicate",
            { target: dupDns[0] },
            `上游列表中存在重复的目标服务器地址 [${dupDns[0]}]`
          );
          dnsError.value = errMsg;
          message(errMsg, { type: "warning" });
          callback(false);
          return;
        }
        dnsError.value = "";
        callback(valid);
      });
    }
  };
}

defineExpose({ getRef });

const dnsResolverGroupRef = ref();

onMounted(() => {
  initRulesFromProps();
  initDnsResolverFromProps();
  initExtraSniFromProps();
  fetchCustomRules();
  fetchTunnels();
  initSortableDns();
});

function initSortableDns() {
  if (dnsResolverGroupRef.value) {
    Sortable.create(dnsResolverGroupRef.value, {
      handle: ".drag-handle",
      animation: 150,
      onEnd: (evt: any) => {
        const { oldIndex, newIndex } = evt;
        if (
          oldIndex !== undefined &&
          newIndex !== undefined &&
          oldIndex !== newIndex
        ) {
          const item = dnsResolverList.value.splice(oldIndex, 1)[0];
          dnsResolverList.value.splice(newIndex, 0, item);
          syncDnsResolver();
        }
      }
    });
  }
}

watch(
  () => props.formInline.id,
  () => {
    newFormInline.value = props.formInline;
    initRulesFromProps();
    syncSelectedTunnelNodeKey();
    initDnsResolverFromProps();
    initExtraSniFromProps();
  }
);

function initRulesFromProps() {
  if (newFormInline.value.rules) {
    try {
      const raw = newFormInline.value.rules;
      const parsed = typeof raw === "string" ? JSON.parse(raw) : raw;
      if (Array.isArray(parsed)) {
        selectedRules.value = parsed;
        return;
      }
    } catch {
      selectedRules.value = [];
    }
  }
  selectedRules.value = [];
}

async function fetchTunnels() {
  try {
    const res = await getTunnelList();
    let list: any[] = [];
    if (Array.isArray(res?.data?.list)) {
      list = res.data.list;
    } else if (Array.isArray(res?.data)) {
      list = res.data;
    } else if (Array.isArray(res)) {
      list = res;
    }

    const groups: TunnelGroup[] = [];
    list.forEach((tItem: any) => {
      const tidStr = String(tItem.Id || tItem.id);
      const tName = tItem.Name || tItem.name || "";
      const tPort = tItem.Port || tItem.port || "";
      const tType = (tItem.Type || tItem.type || "TLS")
        .toLowerCase()
        .includes("quic")
        ? "quic"
        : "tls";

      const portLabel = `${t("tunnel.port", "端口")}: ${tPort}`;
      const groupLabel = `${tName ? "[" + tName + "] " : ""}Tunnel #${tidStr} (${tType.toUpperCase()} | ${portLabel})`;

      const nodeOpts: TunnelGroupOption[] = [];
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
  } catch (e) {}
}

function syncSelectedTunnelNodeKey() {
  const tid = newFormInline.value.tunnel_id;
  const token = newFormInline.value.tunnel_token;
  if (!tid) {
    selectedTunnelNodeKey.value = "";
    return;
  }

  let allOpts: TunnelGroupOption[] = [];
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
    const tType = newFormInline.value.tunnel_type || "tls";
    const fallbackName = token
      ? `Node-${token.length > 6 ? token.slice(-6) : token}`
      : t("tunnelClient.nodeRef", "节点");
    const fallbackOption: TunnelGroupOption = {
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

function handleRulesChange(val: string[]) {
  newFormInline.value.rules = JSON.stringify(val);
}

function removeRule(index: number) {
  selectedRules.value.splice(index, 1);
  handleRulesChange(selectedRules.value);
}

function handleTunnelNodeChange(val: string) {
  if (!val) {
    newFormInline.value.tunnel_id = "";
    newFormInline.value.tunnel_token = "";
    newFormInline.value.tunnel_type = "";
    return;
  }
  let allOpts: TunnelGroupOption[] = [];
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

const formRules = computed(() => ({
  sni: [
    {
      required: true,
      message: t("sni.sniRequired", "请输入主 SNI 作为转发目标"),
      trigger: "blur"
    }
  ],
  port: [
    {
      required: true,
      message: t("sni.portRequired", "请输入监听端口"),
      trigger: "blur"
    },
    {
      pattern: /^([1-9][0-9]{0,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$/,
      message: t("sni.portFormatError", "端口必须为有效数字 (1-65535)"),
      trigger: "blur"
    }
  ]
}));
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :model="newFormInline"
    :rules="formRules"
    label-width="140px"
    label-position="right"
    class="mt-2"
  >
    <el-card shadow="never" class="mb-4 border-(--el-border-color-lighter)! rounded-xl">
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-primary rounded-full" />
          <span class="font-bold text-(--el-text-color-primary) text-sm sm:text-base">
            {{ t("sni.baseInfoSection", "基本信息") }}
          </span>
        </div>
      </template>

      <el-row :gutter="16">
        <re-col :value="12" :xs="24">
          <el-form-item :label="t('sni.name', '名称')" prop="name">
            <el-input
              v-model="newFormInline.name"
              clearable
              :placeholder="t('sni.namePlaceholder', '选填，不填将自动生成')"
            />
          </el-form-item>
        </re-col>
        <re-col :value="12" :xs="24">
          <el-form-item :label="t('sni.remark', '备注')">
            <el-input
              v-model="newFormInline.remark"
              clearable
              :placeholder="t('sni.remarkPlaceholder', '请输入备注信息 (选填)')"
            />
          </el-form-item>
        </re-col>
      </el-row>
      <el-row :gutter="16">
        <re-col :value="12" :xs="24">
          <el-form-item :label="t('sni.sni', 'SNI')" prop="sni">
            <el-input
              v-model="newFormInline.sni"
              clearable
              :placeholder="t('sni.sniPlaceholder', '如 *.example.com')"
            />
          </el-form-item>
        </re-col>
        <re-col :value="12" :xs="24">
          <el-form-item :label="t('sni.port', '端口')" prop="port">
            <el-input
              v-model="newFormInline.port"
              clearable
              :placeholder="t('sni.portPlaceholder', '如 443')"
            />
          </el-form-item>
        </re-col>
      </el-row>
      <!-- Extra SNI patterns (one per line) -->
      <el-row :gutter="16">
        <re-col :value="24" :xs="24">
          <el-form-item :label="t('sni.extraSni', '关联地址')">
            <div class="w-full flex flex-col space-y-1">
              <el-input
                v-model="extraSniText"
                type="textarea"
                :rows="4"
                :placeholder="t('sni.extraSniPlaceholder', '每行一个 SNI 地址，支持通配符，如:\n*.example.com\napi.example.com\n*.other.org')"
                @input="syncExtraSniToForm"
              />
              <div class="text-xs text-(--el-text-color-secondary)">
                {{ t("sni.extraSniTip", "可在此填写多个关联域名（支持精确匹配和通配符，如 *.example.com）。匹配时将优先精确匹配主 SNI，再精确匹配关联地址，最后通配匹配。") }}
              </div>
            </div>
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>

    <el-card shadow="never" class="mb-4 border-(--el-border-color-lighter)! rounded-xl">
      <template #header>
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
          <div class="flex items-center space-x-2">
            <div class="w-1.5 h-4 bg-orange-500 rounded-full" />
            <span class="font-bold text-(--el-text-color-primary) text-sm sm:text-base">
              {{ t("sni.rulesSection", "规则") }}
            </span>
          </div>
        </div>
      </template>

      <el-form-item :label="t('sni.rules', '中间件')">
        <div class="w-full flex flex-col space-y-1">
          <el-select
            v-model="selectedRules"
            multiple
            filterable
            clearable
            class="w-full"
            :placeholder="t('sni.selectRulesPlaceholder', '请选择中间件')"
            @change="handleRulesChange"
          >
            <el-option
              v-for="item in availableRules"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            >
              <div class="flex justify-between items-center">
                <span>{{ item.label }}</span>
                <span class="text-xs text-gray-400 max-w-[200px] truncate ml-4">{{ item.desc }}</span>
              </div>
            </el-option>
          </el-select>
          <div class="text-xs text-(--el-text-color-secondary)">
            {{
              t(
                "sni.rulesTip",
                "提示: 下拉列表仅展示在“规则”菜单中配置的传输层 (L4) 中间件规则 (如 ip_matcher / reset_conn_action)。"
              )
            }}
          </div>
        </div>
      </el-form-item>

    </el-card>

    <el-card shadow="never" class="mb-4 border-(--el-border-color-lighter)! rounded-xl">
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-purple-500 rounded-full" />
          <span class="font-bold text-(--el-text-color-primary) text-sm sm:text-base">
            {{ t("sni.backendSection", "上游") }}
          </span>
        </div>
      </template>

      <div class="space-y-4">
        <!-- 1. Tunnel Selector -->
        <el-row :gutter="16">
          <re-col :value="24" :xs="24">
            <el-form-item :label="t('sni.tunnel', 'Tunnel')">
              <el-select
                v-model="selectedTunnelNodeKey"
                clearable
                filterable
                class="w-full"
                :placeholder="t('sni.selectTunnelPlaceholder', '选择关联的 Tunnel 客户端节点')"
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
                    <div v-if="!item.disabled" class="flex items-center space-x-2 py-0.5 text-xs">
                      <el-tag size="small" :type="item.isOnline ? 'success' : 'info'" effect="light" class="font-medium">
                        {{ item.isOnline ? t("tunnelClient.online", "在线") : t("tunnelClient.offline", "离线") }}
                      </el-tag>
                      <span class="font-semibold text-(--el-text-color-primary) font-mono">{{ item.cName || "Node" }}</span>
                    </div>
                    <div v-else class="text-xs text-gray-400 py-0.5">
                      {{ item.cName }}
                    </div>
                  </el-option>
                </el-option-group>
              </el-select>
              <div class="text-xs text-(--el-text-color-secondary) mt-1">
                {{ t("sni.tunnelTip", "可选配置。若选择隧道节点，流量将通过该 Tunnel 客户端进行代理转发。") }}
              </div>
            </el-form-item>
          </re-col>
        </el-row>

        <!-- 2. Multi DNS Resolvers -->
        <el-row :gutter="16">
          <re-col :value="24" :xs="24">
            <el-form-item :label="t('sni.dnsResolver', 'DNS')">
              <div class="w-full flex flex-col space-y-2">
                <div ref="dnsResolverGroupRef" class="w-full space-y-2">
                  <div
                    v-for="(dns, dIdx) in dnsResolverList"
                    :key="dns.id"
                    class="flex items-center space-x-2"
                  >
                    <el-input
                      v-model="dns.value"
                      :placeholder="t('sni.dnsResolverPlaceholder', '支持标准 DNS 与 DoH，如 8.8.8.8:53 或 https://dns.google/dns-query (留空使用系统默认 DNS)')"
                      clearable
                      @input="syncDnsResolver"
                    />
                    <el-button
                      type="danger"
                      link
                      :icon="useRenderIcon(Delete)"
                      @click="removeDnsResolver(dIdx)"
                    />
                    <el-tooltip :content="t('sni.dragToReorder', '拖动改变顺序')" placement="top">
                      <el-button
                        type="info"
                        link
                        class="cursor-move drag-handle text-(--el-text-color-secondary)! hover:text-(--el-color-primary)!"
                        :icon="useRenderIcon(Rank)"
                      />
                    </el-tooltip>
                  </div>
                </div>
                <div class="flex items-center justify-between w-full">
                  <el-button
                    type="primary"
                    plain
                    size="small"
                    class="self-start"
                    :icon="useRenderIcon(AddFill)"
                    @click="addDnsResolver"
                  >
                    {{ t("sni.addDnsResolver", "添加 DNS") }}
                  </el-button>
                </div>
                <div v-if="dnsError" class="text-xs text-red-500 mt-1 ml-0.5">
                  {{ dnsError }}
                </div>
                <div class="text-xs text-(--el-text-color-secondary) mt-1">
                  {{ t("sni.dnsResolverTip", "当流量未走 Tunnel 隧道转发时，系统将按顺序优先使用上方的 DNS 服务器解析 SNI 目标域名，仅在当前服务器解析失败或异常时自动向下尝试（类似于 /etc/resolv.conf 故障转移机制）。") }}
                </div>
              </div>
            </el-form-item>
          </re-col>
        </el-row>
      </div>
    </el-card>
  </el-form>
</template>
