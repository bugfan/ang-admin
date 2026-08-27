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

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "新增 TCP",
    id: undefined,
    name: "",
    address: "",
    port: "2443",
    rules: "[]",
    tunnel_type: "quic",
    tunnel_id: "",
    tunnel_token: "",
    upstream_method: "round_robin",
    upstream_servers: JSON.stringify([{ target: "127.0.0.1:9999", weight: 1 }]),
    remark: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const { t } = useI18n();

// Rules multi-select & ordering (L4 rules only for TCP Proxy)
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

function initRulesFromProps() {
  try {
    const raw = newFormInline.value.rules;
    if (raw) {
      const parsed = typeof raw === "string" ? JSON.parse(raw) : raw;
      if (Array.isArray(parsed)) {
        selectedRules.value = parsed;
        return;
      }
    }
  } catch (e) {}
  selectedRules.value = [];
}

function handleRulesChange(val: string[]) {
  newFormInline.value.rules = JSON.stringify(val);
}

function removeRule(index: number) {
  selectedRules.value.splice(index, 1);
  handleRulesChange(selectedRules.value);
}

interface TunnelGroupOption {
  label: string;
  value: string;
  tunnel_id: string;
  tunnel_token: string;
  tunnel_type: string;
  cName: string;
  isOnline: boolean;
  disabled: boolean;
}

interface TunnelGroup {
  tunnel_id: string;
  groupLabel: string;
  options: TunnelGroupOption[];
}

// Available Tunnel Nodes Grouped
const tunnelNodeGroups = ref<TunnelGroup[]>([]);
const selectedTunnelNodeKey = ref<string>("");

async function fetchTunnelNodes() {
  try {
    const res = await getTunnelList();
    let list: any[] = [];
    if (res?.code === 0 && res?.data?.list) {
      list = res.data.list;
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
  } catch (e) {
    tunnelNodeGroups.value = [];
  }
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

// Upstream Servers Table State
type UpstreamRow = {
  target: string;
  weight: number;
};

const upstreamList = ref<UpstreamRow[]>([]);

function initUpstreamFromProps() {
  try {
    const raw = newFormInline.value.upstream_servers;
    if (raw) {
      const parsed = typeof raw === "string" ? JSON.parse(raw) : raw;
      if (Array.isArray(parsed)) {
        upstreamList.value = parsed.map((item: any) => ({
          target: item.target || item.Target || "",
          weight: Number(item.weight || item.Weight || 1)
        }));
        return;
      }
    }
  } catch (e) {}
  upstreamList.value = [{ target: "127.0.0.1:9999", weight: 1 }];
}

function syncUpstreamToForm() {
  newFormInline.value.upstream_servers = JSON.stringify(upstreamList.value);
}

function addUpstreamRow() {
  upstreamList.value.push({ target: "", weight: 1 });
  syncUpstreamToForm();
}

function removeUpstreamRow(index: number) {
  upstreamList.value.splice(index, 1);
  syncUpstreamToForm();
}

const rules = reactive({
  port: [
    {
      required: true,
      validator: (rule: any, value: string, callback: any) => {
        if (!value || value.trim() === "") {
          callback(new Error(t("tcp.portRequired", "请输入监听端口")));
        } else {
          const num = Number(value);
          if (isNaN(num) || num < 1 || num > 65535) {
            callback(
              new Error(t("tcp.portFormatError", "端口必须为有效数字(1-65535)"))
            );
          } else {
            callback();
          }
        }
      },
      trigger: "blur"
    }
  ]
});

function getRef() {
  syncUpstreamToForm();
  return ruleFormRef.value;
}

defineExpose({ getRef });

onMounted(() => {
  initRulesFromProps();
  initUpstreamFromProps();
  fetchCustomRules();
  fetchTunnelNodes();
});

watch(
  () => props.formInline,
  val => {
    newFormInline.value = val;
    initRulesFromProps();
    initUpstreamFromProps();
  },
  { deep: true }
);
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :label-position="deviceDetection() ? 'top' : 'right'"
    :model="newFormInline"
    :rules="rules"
    label-width="120px"
    class="tcp-form p-1 sm:px-2 space-y-4"
  >
    <!-- Section 1: Basic Information -->
    <el-card
      shadow="never"
      class="mb-4 border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-primary rounded-full" />
          <span
            class="font-bold text-(--el-text-color-primary) text-sm sm:text-base"
          >
            {{ t("tcp.baseInfoSection", "基本信息") }}
          </span>
        </div>
      </template>

      <el-row :gutter="16">
        <re-col :value="12" :xs="24">
          <el-form-item :label="t('tcp.name', '名称')" prop="name">
            <el-input
              v-model="newFormInline.name"
              clearable
              :placeholder="t('tcp.namePlaceholder', '请输入 TCP 代理名称')"
            />
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('tcp.port', '监听端口')" prop="port">
            <el-input
              v-model="newFormInline.port"
              clearable
              :placeholder="t('tcp.portFormatError', '如 2443')"
            />
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('tcp.address', '绑定地址')" prop="address">
            <el-input
              v-model="newFormInline.address"
              clearable
              :placeholder="
                t('tcp.addressPlaceholder', '如 0.0.0.0 (留空默认监听所有网卡)')
              "
            />
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('tcp.remark', '备注')" prop="remark">
            <el-input
              v-model="newFormInline.remark"
              clearable
              :placeholder="
                t('tcp.remarkPlaceholder', '请输入备注信息 (选填)')
              "
            />
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>

    <!-- Section 2: Rules (规则) -->
    <el-card
      shadow="never"
      class="mb-4 border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-amber-500 rounded-full" />
          <span
            class="font-bold text-(--el-text-color-primary) text-sm sm:text-base"
          >
            {{ t("tcp.rulesSection", "规则") }}
          </span>
        </div>
      </template>

      <el-form-item :label="t('tcp.rules', '中间件')" prop="rules">
        <el-select
          v-model="selectedRules"
          multiple
          filterable
          clearable
          collapse-tags
          collapse-tags-tooltip
          class="w-full"
          :placeholder="t('tcp.selectRulesPlaceholder', '请选择中间件')"
          @change="handleRulesChange"
        >
          <el-option
            v-for="ruleItem in availableRules"
            :key="ruleItem.value"
            :label="ruleItem.label"
            :value="ruleItem.value"
          >
            <div class="flex items-center justify-between w-full">
              <span class="font-mono text-sm">{{ ruleItem.label }}</span>
              <span class="text-xs text-gray-400 ml-4">{{ ruleItem.desc }}</span>
            </div>
          </el-option>
        </el-select>
      </el-form-item>

      <div class="text-xs/relaxed text-(--el-text-color-secondary) mt-1">
        {{
          t(
            "tcp.rulesTip",
            "提示: 下拉列表仅展示在“规则”菜单中配置的传输层 (L4) 中间件规则 (如 ip_matcher / reset_conn_action)。"
          )
        }}
      </div>

      <!-- Selected rule tag ordering visualization -->
      <div
        v-if="selectedRules && selectedRules.length > 0"
        class="mt-3 p-3 bg-(--el-fill-color-light) rounded-lg border border-(--el-border-color-lighter)"
      >
        <div class="text-xs font-semibold text-gray-500 mb-2">
          {{
            t(
              "tcp.ruleOrderTip",
              "中间件规则按列表从上到下先后顺序依次执行"
            )
          }}:
        </div>
        <div class="flex flex-wrap gap-2">
          <el-tag
            v-for="(r, idx) in selectedRules"
            :key="r"
            closable
            size="default"
            type="success"
            effect="light"
            class="font-mono text-sm py-1 px-2.5 flex items-center"
            @close="removeRule(idx)"
          >
            <span class="opacity-60 text-xs mr-1 font-sans">#{{ idx + 1 }}</span>
            {{ r }}
          </el-tag>
        </div>
      </div>
    </el-card>

    <!-- Section 3: Upstream (上游) -->
    <el-card
      shadow="never"
      class="mb-4 border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-purple-500 rounded-full" />
          <span
            class="font-bold text-(--el-text-color-primary) text-sm sm:text-base"
          >
            {{ t("tcp.backendSection", "上游") }}
          </span>
        </div>
      </template>

      <div class="space-y-4">
        <!-- 1. Tunnel Selector (Top part) -->
        <el-row :gutter="16">
          <re-col :value="24" :xs="24">
            <el-form-item :label="t('tcp.tunnel', 'Tunnel')">
              <el-select
                v-model="selectedTunnelNodeKey"
                clearable
                filterable
                class="w-full"
                :placeholder="
                  t(
                    'tcp.selectTunnelPlaceholder',
                    '选择关联的 Tunnel 客户端节点'
                  )
                "
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
              <div class="text-xs text-(--el-text-color-secondary) mt-1">
                {{
                  t(
                    "tcp.tunnelTip",
                    "可选配置。若选择隧道节点，TCP 流量将通过该 Tunnel 客户端进行代理转发。"
                  )
                }}
              </div>
            </el-form-item>
          </re-col>
        </el-row>

        <!-- 2. Target Server Address & Load Balance (Bottom part) -->
        <div class="border-t border-(--el-border-color-lighter) pt-4">
          <div
            class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-3"
          >
            <div>
              <span class="font-bold text-sm text-(--el-text-color-primary)">
                {{ t("tcp.upstreamServers", "目标服务器地址") }}
              </span>
            </div>
            <el-button
              type="primary"
              size="small"
              :icon="useRenderIcon(AddFill)"
              class="self-start sm:self-auto shrink-0"
              @click="addUpstreamRow"
            >
              {{ t("tcp.addTarget", "添加服务器") }}
            </el-button>
          </div>

          <el-row :gutter="16" class="mb-3">
            <re-col :value="12" :xs="24">
              <el-form-item :label="t('tcp.upstreamMethod', '负载均衡')">
                <el-select
                  v-model="newFormInline.upstream_method"
                  class="w-full"
                >
                  <el-option label="轮询 (Round Robin)" value="round_robin" />
                  <el-option label="权重 (Weight)" value="weight" />
                  <el-option label="IP 哈希 (IP Hash)" value="ip_hash" />
                </el-select>
              </el-form-item>
            </re-col>
          </el-row>

          <div
            class="border border-(--el-border-color-lighter) rounded-lg overflow-hidden"
          >
            <el-table
              :data="upstreamList"
              size="small"
              :empty-text="
                t(
                  'tcp.noUpstreamServers',
                  '暂无服务器，请点击右上角“添加服务器”'
                )
              "
              class="w-full"
            >
              <el-table-column
                type="index"
                label="#"
                width="50"
                align="center"
              />
              <el-table-column
                :label="t('tcp.target', '目标地址 (Host:Port)')"
                min-width="220"
              >
                <template #default="{ row }">
                  <el-input
                    v-model="row.target"
                    size="small"
                    :placeholder="
                      t(
                        'tcp.targetPlaceholder',
                        '如 127.0.0.1:9999 或 10.0.0.5:3306'
                      )
                    "
                    @change="syncUpstreamToForm"
                  />
                </template>
              </el-table-column>
              <el-table-column
                :label="t('tcp.weight', '权重')"
                width="140"
                align="center"
              >
                <template #default="{ row }">
                  <el-input-number
                    v-model="row.weight"
                    :min="1"
                    :max="1000"
                    size="small"
                    class="w-full"
                    @change="syncUpstreamToForm"
                  />
                </template>
              </el-table-column>
              <el-table-column
                :label="t('tcp.operation', '操作')"
                width="80"
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
