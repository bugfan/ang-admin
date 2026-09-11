<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import type { FormRules } from "element-plus";
import ReCol from "@/components/ReCol";
import PlusIcon from "~icons/ep/plus";
import Delete from "~icons/ep/delete";
import { getTunnelList } from "@/api/tunnel";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import type { FormProps } from "../utils/types";

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

const props = withDefaults(
  defineProps<FormProps & { groupList?: any[]; domainList?: any[] }>(),
  {
    groupList: () => [],
    domainList: () => []
  }
);

const { t } = useI18n();
const ruleFormRef = ref();
const newFormInline = ref(props.formInline);

if (!newFormInline.value.replaceList) {
  newFormInline.value.replaceList = [];
}

watch(
  () => props.formInline,
  val => {
    newFormInline.value = val;
    if (!newFormInline.value.replaceList) {
      newFormInline.value.replaceList = [];
    }
    syncSelectedTunnelNodeKey();
  },
  { deep: true }
);

function addReplaceRule() {
  if (!newFormInline.value.replaceList) {
    newFormInline.value.replaceList = [];
  }
  newFormInline.value.replaceList.push({ k: "", v: "" });
}

function removeReplaceRule(index: number) {
  newFormInline.value.replaceList?.splice(index, 1);
}

const tunnelNodeGroups = ref<TunnelGroup[]>([]);
const selectedTunnelNodeKey = ref<string>("");

onMounted(() => {
  fetchTunnels();
});

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
  } catch (e) {
    console.error("fetchTunnels error:", e);
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

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef, newFormInline });

const rules: FormRules = {
  name: [
    {
      required: true,
      validator: (rule, value, callback) => {
        if (!value || !value.trim()) {
          callback(new Error(t("webvpn.valNameRequired", "应用名称不能为空")));
        } else {
          callback();
        }
      },
      trigger: "blur"
    }
  ],
  domain_id: [
    {
      required: true,
      validator: (rule, value, callback) => {
        if (!value || value <= 0) {
          callback(
            new Error(
              t("webvpn.valDomainRequired", "必须选择所属 WebVPN 基础域")
            )
          );
        } else {
          callback();
        }
      },
      trigger: "change"
    }
  ],
  target_url: [
    {
      required: true,
      validator: (rule, value, callback) => {
        if (!value || !value.trim()) {
          callback(
            new Error(t("webvpn.valTargetUrlRequired", "站点地址不能为空"))
          );
        } else if (
          !value.startsWith("http://") &&
          !value.startsWith("https://")
        ) {
          callback(
            new Error(
              t(
                "webvpn.valTargetUrlInvalid",
                "必须是以 http:// 或 https:// 开头的合法完整 URL"
              )
            )
          );
        } else {
          callback();
        }
      },
      trigger: "blur"
    }
  ]
};

// Selected WebVPN Domain details
const selectedDomain = computed(() => {
  const sid =
    newFormInline.value.domain_id || newFormInline.value.http_proxy_id;
  if (!sid || !props.domainList) return null;
  return props.domainList.find((s: any) => (s.Id || s.id) === sid);
});

// Auto-derive WebVPN Prefix and Full Access Address in real-time
const derivedInfo = computed(() => {
  const target = (newFormInline.value.target_url || "").trim();
  const domain = selectedDomain.value;
  if (!target) {
    return { prefix: "", fullUrl: "" };
  }

  try {
    const u = new URL(target);
    const targetHost = u.hostname;
    let targetPort = u.port;
    const isHttps = u.protocol === "https:";
    const schemePrefix = isHttps ? "s-" : "c-";
    if (!targetPort) {
      targetPort = isHttps ? "443" : "80";
    }

    // Escape any '-' to '--', then replace '.' with '-'
    const escaped = targetHost.replace(/-/g, "--");
    const dashed = escaped.replace(/\./g, "-");
    const prefix = `${schemePrefix}${dashed}-${targetPort}`;

    if (!domain) {
      return { prefix, fullUrl: "" };
    }

    const domainHostname = domain.Hostname || domain.hostname || "";
    const rootDomain = domainHostname.replace(/^\*\./, "");
    if (!rootDomain) {
      return { prefix, fullUrl: "" };
    }

    const isTLS = domain.TLS ?? domain.tls ?? domain.H2 ?? domain.h2 ?? true;
    const scheme = isTLS ? "https://" : "http://";
    const port = domain.Port || domain.port || "443";
    const portSuffix =
      port !== "80" && port !== "443" && port !== "" ? `:${port}` : "";

    const fullUrl = `${scheme}${prefix}.${rootDomain}${portSuffix}`;
    return { prefix, fullUrl };
  } catch (e) {
    return { prefix: "", fullUrl: "" };
  }
});
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :model="newFormInline"
    :rules="rules"
    label-width="140px"
    class="space-y-6"
  >
    <!-- Section 1: 基本信息 -->
    <el-card
      shadow="never"
      class="border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-primary rounded-full" />
          <span
            class="font-bold text-(--el-text-color-primary) text-sm sm:text-base"
          >
            {{ t("webvpn.basicSection", "基本信息") }}
          </span>
        </div>
      </template>

      <el-row :gutter="24">
        <!-- 1. 应用名称 -->
        <re-col :value="24">
          <el-form-item :label="t('webvpn.name', '名称')" prop="name">
            <el-input
              v-model="newFormInline.name"
              clearable
              :placeholder="
                t('webvpn.namePlaceholder', '如：中国知网、内部 OA 办公系统')
              "
            />
          </el-form-item>
        </re-col>

        <!-- 2. 所属基础域 -->
        <re-col :value="24">
          <el-form-item
            :label="t('webvpn.domain', '所属基础域')"
            prop="domain_id"
          >
            <el-select
              v-model="newFormInline.domain_id"
              :placeholder="
                t('webvpn.domainPlaceholder', '选择已配置的 WebVPN 基础域网关')
              "
              class="w-full"
              filterable
            >
              <el-option
                v-for="item in domainList"
                :key="item.Id || item.id"
                :label="`${item.Name || item.name} (${item.Hostname || item.hostname})`"
                :value="item.Id || item.id"
              >
                <div class="flex-bc w-full">
                  <span>{{ item.Name || item.name }}</span>
                  <div class="flex items-center gap-2">
                    <span class="text-xs text-gray-400 font-mono">{{
                      item.Hostname || item.hostname
                    }}</span>
                    <el-tag size="small" type="primary" effect="light">
                      {{ item.Port || item.port || "443" }}
                    </el-tag>
                  </div>
                </div>
              </el-option>
            </el-select>
          </el-form-item>
        </re-col>

        <!-- 3. 站点地址 -->
        <re-col :value="24">
          <el-form-item
            :label="t('webvpn.targetUrl', '站点地址')"
            prop="target_url"
          >
            <el-input
              v-model="newFormInline.target_url"
              clearable
              :placeholder="
                t(
                  'webvpn.targetUrlPlaceholder',
                  '例如：https://www.cnki.net 或 http://192.168.1.100:8080'
                )
              "
            />
            <p class="text-xs/relaxed text-gray-400 mt-2">
              {{
                t(
                  "webvpn.targetUrlHint",
                  "内部业务系统的真实 URL，输入后将实时自动推导 WebVPN 替换地址。"
                )
              }}
            </p>
          </el-form-item>
        </re-col>

        <!-- 4. 访问地址 (只读自动生成，放在站点地址下面) -->
        <re-col :value="24">
          <el-form-item :label="t('webvpn.accessUrl', '访问地址')">
            <el-input
              :model-value="derivedInfo.fullUrl"
              disabled
              :placeholder="
                t(
                  'webvpn.accessUrlPlaceholder',
                  '根据目标地址与所属泛域名站点自动生成'
                )
              "
              class="font-mono"
            >
              <template #suffix>
                <el-tag
                  v-if="derivedInfo.fullUrl"
                  size="small"
                  type="info"
                  effect="plain"
                >
                  {{ t("webvpn.autoGenerated", "自动生成") }}
                </el-tag>
              </template>
            </el-input>
            <p class="text-xs/relaxed text-gray-400 mt-2">
              {{
                t(
                  "webvpn.accessUrlHint",
                  "用户在外网免客户端访问该系统时使用的实际 URL（系统自动编排锁定，不可手动更改）。"
                )
              }}
            </p>
          </el-form-item>
        </re-col>

        <!-- 5. 公开访问 (Protected) -->
        <re-col :value="24">
          <el-form-item
            :label="t('webvpn.accessMode', '公开访问')"
            prop="is_protected"
            :for="''"
          >
            <div class="flex flex-col items-start">
              <el-switch
                v-model="newFormInline.is_protected"
                :active-value="0"
                :inactive-value="1"
                :active-text="t('webvpn.modePublic', '公开')"
                :inactive-text="t('webvpn.modeProtected', '不公开')"
              />
              <p class="text-xs/relaxed text-gray-400 mt-2">
                {{
                  t(
                    "webvpn.accessModeHint",
                    "公开后外网可直接免登录访问；不公开则必须登录且具备用户组权限方可访问。"
                  )
                }}
              </p>
            </div>
          </el-form-item>
        </re-col>

        <!-- 6. 用户组 (仅在受保护/不公开模式下展示，紧随公开访问开关) -->
        <re-col v-if="newFormInline.is_protected === 1" :value="24">
          <el-form-item
            :label="t('webvpn.allowedGroups', '用户组')"
            prop="group_ids"
          >
            <el-select
              v-model="newFormInline.group_ids"
              multiple
              collapse-tags
              collapse-tags-tooltip
              :placeholder="
                t('webvpn.groupsPlaceholder', '全部已登录用户可访问（留空）')
              "
              class="w-full"
            >
              <el-option
                v-for="item in groupList"
                :key="item.Id || item.id"
                :label="item.Name || item.name"
                :value="item.Id || item.id"
              />
            </el-select>
            <p class="text-xs/relaxed text-gray-400 mt-2">
              {{
                t(
                  "webvpn.groupsHint",
                  "留空代表对所有已登录用户开放；选择特定组后，仅有权限的组才可访问。"
                )
              }}
            </p>
          </el-form-item>
        </re-col>

        <!-- 7. 启用 -->
        <re-col :value="24">
          <el-form-item
            :label="t('webvpn.status', '启用')"
            prop="status"
            :for="''"
          >
            <el-switch
              v-model="newFormInline.status"
              :active-value="1"
              :inactive-value="0"
              :active-text="t('webvpn.statusEnabled', '启用')"
              :inactive-text="t('webvpn.statusDisabled', '禁用')"
            />
          </el-form-item>
        </re-col>

        <!-- 8. 备注 -->
        <re-col :value="24">
          <el-form-item :label="t('webvpn.remark', '备注')" prop="remark">
            <el-input
              v-model="newFormInline.remark"
              type="textarea"
              :rows="2"
              :placeholder="
                t(
                  'webvpn.remarkPlaceholder',
                  '选填，关于该 WebVPN 应用的详细说明'
                )
              "
            />
          </el-form-item>
        </re-col>
      </el-row>

      <!-- Tunnel Section -->
      <el-divider class="my-5 border-(--el-border-color-lighter)!" />
      <el-row :gutter="16">
        <re-col :value="12" :xs="24">
          <el-form-item :label="t('http.assocTunnel', 'Tunnel')">
            <el-select
              v-model="selectedTunnelNodeKey"
              clearable
              filterable
              :placeholder="
                t('http.assocTunnelPlaceholder', '选择关联的 Tunnel 客户端节点')
              "
              class="w-full"
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
                  <div v-else class="text-xs text-gray-400">
                    {{ item.label }}
                  </div>
                </el-option>
              </el-option-group>
            </el-select>
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>

    <!-- Section 2: 关联地址 / 扩展替换 -->
    <el-card
      shadow="never"
      class="border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-primary rounded-full" />
          <span
            class="font-bold text-(--el-text-color-primary) text-sm sm:text-base"
          >
            {{ t("webvpn.advancedSection", "关联地址 / 扩展替换") }}
          </span>
        </div>
      </template>

      <!-- 1. 关联地址 -->
      <el-form-item :label="t('webvpn.hostsLabel', '关联地址')" prop="hosts">
        <el-input
          v-model="newFormInline.hosts"
          type="textarea"
          :rows="5"
          :placeholder="
            t(
              'webvpn.hostsPlaceholder',
              '每行一个域名，例如：\ncnki.net\napi.cnki.net\n*.static.cnki.net'
            )
          "
          class="font-mono text-xs"
        />
        <p class="text-xs/relaxed text-gray-400 mt-2">
          {{
            t(
              "webvpn.hostsHint",
              "每行一个域名。主目标域名之外如果该系统还引用了其他域名或静态资源域名，可填写在此处。规则：含 * 的行自动归入 wildcard 通配映射，其余行自动归入 host 映射，引擎自动计算 WebVPN 子域名动态替换。"
            )
          }}
        </p>
      </el-form-item>

      <el-divider class="my-5 border-(--el-border-color-lighter)!" />

      <!-- 2. 扩展替换 -->
      <el-form-item
        :label="t('webvpn.replaceSection', '扩展替换')"
        prop="replace"
      >
        <div class="w-full space-y-3">
          <div class="flex-bc min-h-8">
            <p class="text-xs/relaxed text-gray-400">
              {{
                t(
                  "webvpn.replaceDesc",
                  "支持在响应报文（HTML、JS、CSS 等）中将指定的文本或字符实时替换为目标内容。"
                )
              }}
            </p>
            <el-button
              type="primary"
              plain
              size="small"
              :icon="useRenderIcon(PlusIcon)"
              @click="addReplaceRule"
            >
              {{ t("webvpn.addReplaceRule", "添加") }}
            </el-button>
          </div>

          <div
            v-if="
              !newFormInline.replaceList ||
              newFormInline.replaceList.length === 0
            "
            class="text-center py-4 text-xs text-(--el-text-color-placeholder) bg-gray-50 dark:bg-gray-800/40 rounded-lg border border-dashed border-gray-200 dark:border-gray-700"
          >
            {{
              t(
                "webvpn.noReplaceRules",
                "暂无内容替换规则，如需修改页面响应内容可点击上方按钮添加"
              )
            }}
          </div>

          <div v-else class="space-y-2">
            <div
              class="grid grid-cols-12 gap-2 text-xs font-semibold text-gray-500 px-1 mb-1"
            >
              <div class="col-span-5">
                {{ t("webvpn.replaceOriginal", "原字符串 / 匹配内容") }}
              </div>
              <div class="col-span-6">
                {{ t("webvpn.replaceTarget", "替换为目标内容") }}
              </div>
              <div class="col-span-1 text-center">
                {{ t("common.operations", "操作") }}
              </div>
            </div>
            <div
              v-for="(item, index) in newFormInline.replaceList"
              :key="index"
              class="grid grid-cols-12 gap-2 items-center"
            >
              <div class="col-span-5">
                <el-input
                  v-model="item.k"
                  :placeholder="
                    t('webvpn.replaceOriginal', '原字符串 / 匹配内容')
                  "
                  class="font-mono text-xs"
                  clearable
                />
              </div>
              <div class="col-span-6">
                <el-input
                  v-model="item.v"
                  :placeholder="t('webvpn.replaceTarget', '替换为目标内容')"
                  class="font-mono text-xs"
                  clearable
                />
              </div>
              <div class="col-span-1 text-center">
                <el-button
                  type="danger"
                  link
                  size="small"
                  :icon="useRenderIcon(Delete)"
                  @click="removeReplaceRule(index)"
                />
              </div>
            </div>
          </div>
        </div>
      </el-form-item>
    </el-card>
  </el-form>
</template>
