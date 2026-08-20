<script setup lang="ts">
import { ref, reactive, watch } from "vue";
import ReCol from "@/components/ReCol";
import { useI18n } from "vue-i18n";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { deviceDetection } from "@pureadmin/utils";
import draggable from "vuedraggable";

import AddFill from "~icons/ri/add-line";
import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import ArrowUp from "~icons/ep/arrow-up";
import ArrowDown from "~icons/ep/arrow-down";
import DragIcon from "~icons/ri/drag-move-2-line";

interface RuleItemConfig {
  id: string;
  matcherName: string;
  ipAddressText: string;
  httpIpAddressText: string;
  urlMethod: string;
  urlPath: string;
  jsScript: string;
  actionName: string;
  resetContent: string;
  authGuardLoginUrl: string;
  forwardUrl: string;
  modifyStatusCode: number | string;
  replaceReqBodyMap: string;
  replaceRespBodyMap: string;
  replaceReqHeaderMap: string;
  replaceRespHeaderMap: string;
  respTextCode: string;
  respTextContent: string;
  respTextHeader: string;
  webvpnSites: string;
}

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "",
    id: undefined,
    name: "",
    items: JSON.stringify(
      [
        {
          Matcher: { Name: "ip_matcher", Config: { Address: ["127.0.0.1"] } },
          Action: {
            Name: "reset_conn_action",
            Config: { Content: "Connection reset by rule" }
          }
        }
      ],
      null,
      2
    ),
    remark: ""
  })
});

const { t } = useI18n();

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);

// Supported Matcher options
const matcherOptions = [
  {
    label: "Always True (always_true_matcher)",
    value: "always_true_matcher",
    tag: "L4",
    tagType: "primary"
  },
  {
    label: "TCP / L4 IP (ip_matcher)",
    value: "ip_matcher",
    tag: "L4",
    tagType: "primary"
  },
  {
    label: "TCP / L4 CIDR (cidr_matcher)",
    value: "cidr_matcher",
    tag: "L4",
    tagType: "primary"
  },
  {
    label: "HTTP Proxy IP (http_ip_matcher)",
    value: "http_ip_matcher",
    tag: "HTTP",
    tagType: "success"
  },
  {
    label: "HTTP URL (url_matcher)",
    value: "url_matcher",
    tag: "HTTP",
    tagType: "success"
  },
  {
    label: "HTTP JS Script (js_matcher)",
    value: "js_matcher",
    tag: "HTTP",
    tagType: "success"
  }
];

// Supported Action options
const actionOptions = [
  {
    label: "TCP Reset (reset_conn_action)",
    value: "reset_conn_action",
    tag: "L4",
    tagType: "danger"
  },
  {
    label: "HTTP Hide Version (hide_version_action)",
    value: "hide_version_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: "Auth Guard (auth_guard_action)",
    value: "auth_guard_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: "Auth Portal (auth_portal_action)",
    value: "auth_portal_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: "Forward Request (forward_request_action)",
    value: "forward_request_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: "Modify Status (modify_status_action)",
    value: "modify_status_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: "Replace Request Body (replace_request_body_action)",
    value: "replace_request_body_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: "Replace Response Body (replace_response_body_action)",
    value: "replace_response_body_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: "Replace Request Header (replace_request_header_action)",
    value: "replace_request_header_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: "Replace Response Header (replace_response_header_action)",
    value: "replace_response_header_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: "Response Text (response_text_action)",
    value: "response_text_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: "Subdomain WebVPN (subdomain_webvpn_action)",
    value: "subdomain_webvpn_action",
    tag: "HTTP",
    tagType: "warning"
  }
];

// Item List
const itemList = ref<RuleItemConfig[]>([]);

// Sub-dialog state for editing/adding a single RuleItem
const itemDialogVisible = ref(false);
const editingIndex = ref<number | null>(null);
const itemForm = reactive<RuleItemConfig>({
  id: "",
  matcherName: "ip_matcher",
  ipAddressText: "127.0.0.1",
  httpIpAddressText: "192.168.1.0/24",
  urlMethod: "ALL",
  urlPath: "/",
  jsScript: "return true;",
  actionName: "reset_conn_action",
  resetContent: "Connection reset by rule",
  authGuardLoginUrl: "/login",
  forwardUrl: "",
  modifyStatusCode: 403,
  replaceReqBodyMap: "{}",
  replaceRespBodyMap: "{}",
  replaceReqHeaderMap: "{}",
  replaceRespHeaderMap: "{}",
  respTextCode: "403",
  respTextContent: "Forbidden",
  respTextHeader: "{}",
  webvpnSites: "{}"
});

onInitForm();

function createDefaultItem(): RuleItemConfig {
  return {
    id: String(Date.now() + Math.random()),
    matcherName: "ip_matcher",
    ipAddressText: "",
    httpIpAddressText: "",
    urlMethod: "ALL",
    urlPath: "/",
    jsScript: "return true;",
    actionName: "reset_conn_action",
    resetContent: "",
    authGuardLoginUrl: "/login",
    forwardUrl: "",
    modifyStatusCode: 403,
    replaceReqBodyMap: "{}",
    replaceRespBodyMap: "{}",
    replaceReqHeaderMap: "{}",
    replaceRespHeaderMap: "{}",
    respTextCode: "403",
    respTextContent: "Forbidden",
    respTextHeader: "{}",
    webvpnSites: "{}"
  };
}

function onInitForm() {
  itemList.value = [];
  try {
    if (newFormInline.value.items) {
      const parsed =
        typeof newFormInline.value.items === "string"
          ? JSON.parse(newFormInline.value.items)
          : newFormInline.value.items;

      if (Array.isArray(parsed) && parsed.length > 0) {
        for (const item of parsed) {
          const mObj = item?.Matcher || item?.matcher || {};
          const aObj = item?.Action || item?.action || {};
          const mName = mObj?.Name || mObj?.name || "ip_matcher";
          const mCfg = mObj?.Config || mObj?.config || {};
          const aName = aObj?.Name || aObj?.name || "reset_conn_action";
          const aCfg = aObj?.Config || aObj?.config || {};

          let ipText = "127.0.0.1";
          let httpIpText = "192.168.1.0/24";
          let urlMethod = "ALL";
          let urlPath = "/";
          let jsScript = "return true;";

          if (mName === "ip_matcher" && Array.isArray(mCfg.Address)) {
            ipText = mCfg.Address.join("\n");
          } else if (mName === "cidr_matcher" && Array.isArray(mCfg.CIDRs)) {
            ipText = mCfg.CIDRs.join("\n");
          } else if (mName === "http_ip_matcher" && Array.isArray(mCfg.Address)) {
            httpIpText = mCfg.Address.join("\n");
          } else if (mName === "url_matcher") {
            urlMethod = mCfg.Method || "ALL";
            urlPath = mCfg.Path || "/";
          } else if (mName === "js_matcher") {
            jsScript = mCfg.Script || "";
          }

          let resetContent = "";
          let authGuardLoginUrl = "/login";
          let forwardUrl = "";
          let modifyStatusCode = 403;
          let replaceReqBodyMap = "{}";
          let replaceRespBodyMap = "{}";
          let replaceReqHeaderMap = "{}";
          let replaceRespHeaderMap = "{}";
          let respTextCode = "403";
          let respTextContent = "";
          let respTextHeader = "{}";
          let webvpnSites = "{}";

          if (aName === "reset_conn_action") resetContent = aCfg.Content || "";
          if (aName === "auth_guard_action") authGuardLoginUrl = aCfg.LoginURL || "";
          if (aName === "forward_request_action") forwardUrl = aCfg.Url || "";
          if (aName === "modify_status_action") modifyStatusCode = aCfg.Code || 403;
          if (aName === "replace_request_body_action") replaceReqBodyMap = aCfg.Map ? JSON.stringify(aCfg.Map, null, 2) : "{}";
          if (aName === "replace_response_body_action") replaceRespBodyMap = aCfg.Map ? JSON.stringify(aCfg.Map, null, 2) : "{}";
          if (aName === "replace_request_header_action") replaceReqHeaderMap = aCfg.Map ? JSON.stringify(aCfg.Map, null, 2) : "{}";
          if (aName === "replace_response_header_action") replaceRespHeaderMap = aCfg.Map ? JSON.stringify(aCfg.Map, null, 2) : "{}";
          if (aName === "response_text_action") {
             respTextCode = aCfg.Code || "403";
             respTextContent = aCfg.Content || "";
             respTextHeader = aCfg.Header ? JSON.stringify(aCfg.Header, null, 2) : "{}";
          }
          if (aName === "subdomain_webvpn_action") webvpnSites = aCfg.Sites ? JSON.stringify(aCfg.Sites, null, 2) : "{}";

          itemList.value.push({
            id: String(Math.random()),
            matcherName: mName,
            ipAddressText: ipText,
            httpIpAddressText: httpIpText,
            urlMethod,
            urlPath,
            jsScript,
            actionName: aName,
            resetContent,
            authGuardLoginUrl,
            forwardUrl,
            modifyStatusCode,
            replaceReqBodyMap,
            replaceRespBodyMap,
            replaceReqHeaderMap,
            replaceRespHeaderMap,
            respTextCode,
            respTextContent,
            respTextHeader,
            webvpnSites
          });
        }
      }
    }
  } catch (e) {}

  // if (itemList.value.length === 0) {
  //   itemList.value.push(createDefaultItem());
  // }

  syncToFormJSON();
}

function openAddItemDialog() {
  editingIndex.value = null;
  const def = createDefaultItem();
  itemForm.id = def.id;
  itemForm.matcherName = def.matcherName;
  itemForm.ipAddressText = def.ipAddressText;
  itemForm.httpIpAddressText = def.httpIpAddressText;
  itemForm.actionName = def.actionName;
  itemForm.resetContent = def.resetContent;
  itemDialogVisible.value = true;
}

function openEditItemDialog(index: number) {
  editingIndex.value = index;
  const cur = itemList.value[index];
  itemForm.id = cur.id;
  itemForm.matcherName = cur.matcherName;
  itemForm.ipAddressText = cur.ipAddressText;
  itemForm.httpIpAddressText = cur.httpIpAddressText;
  itemForm.actionName = cur.actionName;
  itemForm.resetContent = cur.resetContent;
  itemDialogVisible.value = true;
}

function saveItemFromDialog() {
  const target: RuleItemConfig = {
    id: itemForm.id || String(Math.random()),
    matcherName: itemForm.matcherName,
    ipAddressText: itemForm.ipAddressText,
    httpIpAddressText: itemForm.httpIpAddressText,
    urlMethod: itemForm.urlMethod,
    urlPath: itemForm.urlPath,
    jsScript: itemForm.jsScript,
    actionName: itemForm.actionName,
    resetContent: itemForm.resetContent,
    authGuardLoginUrl: itemForm.authGuardLoginUrl,
    forwardUrl: itemForm.forwardUrl,
    modifyStatusCode: itemForm.modifyStatusCode,
    replaceReqBodyMap: itemForm.replaceReqBodyMap,
    replaceRespBodyMap: itemForm.replaceRespBodyMap,
    replaceReqHeaderMap: itemForm.replaceReqHeaderMap,
    replaceRespHeaderMap: itemForm.replaceRespHeaderMap,
    respTextCode: itemForm.respTextCode,
    respTextContent: itemForm.respTextContent,
    respTextHeader: itemForm.respTextHeader,
    webvpnSites: itemForm.webvpnSites
  };

  if (editingIndex.value !== null && editingIndex.value >= 0) {
    itemList.value[editingIndex.value] = target;
  } else {
    itemList.value.push(target);
  }

  itemDialogVisible.value = false;
  syncToFormJSON();
}

function removeItem(idx: number) {
  itemList.value.splice(idx, 1);
  syncToFormJSON();
}

function moveItemUp(idx: number) {
  if (idx <= 0) return;
  const temp = itemList.value[idx];
  itemList.value[idx] = itemList.value[idx - 1];
  itemList.value[idx - 1] = temp;
  syncToFormJSON();
}

function moveItemDown(idx: number) {
  if (idx >= itemList.value.length - 1) return;
  const temp = itemList.value[idx];
  itemList.value[idx] = itemList.value[idx + 1];
  itemList.value[idx + 1] = temp;
  syncToFormJSON();
}

function syncToFormJSON() {
  const result: any[] = [];
  for (const item of itemList.value) {
    let matcherObj: any = {};
    if (item.matcherName === "http_ip_matcher") {
      matcherObj = { Name: "http_ip_matcher", Config: { Address: item.httpIpAddressText.split("\n").map(s => s.trim()).filter(Boolean) } };
    } else if (item.matcherName === "url_matcher") {
      matcherObj = { Name: "url_matcher", Config: { Method: item.urlMethod, Path: item.urlPath } };
    } else if (item.matcherName === "js_matcher") {
      matcherObj = { Name: "js_matcher", Config: { Script: item.jsScript } };
    } else if (item.matcherName === "cidr_matcher") {
      matcherObj = { Name: "cidr_matcher", Config: { CIDRs: item.ipAddressText.split("\n").map(s => s.trim()).filter(Boolean) } };
    } else if (item.matcherName === "always_true_matcher") {
      matcherObj = { Name: "always_true_matcher", Config: {} };
    } else {
      matcherObj = { Name: "ip_matcher", Config: { Address: item.ipAddressText.split("\n").map(s => s.trim()).filter(Boolean) } };
    }

    let actionObj: any = {};
    const tryParse = (str: string) => { try { return JSON.parse(str); } catch(e) { return {}; } };

    if (item.actionName === "hide_version_action") actionObj = { Name: "hide_version_action", Config: {} };
    else if (item.actionName === "auth_portal_action") actionObj = { Name: "auth_portal_action", Config: {} };
    else if (item.actionName === "auth_guard_action") actionObj = { Name: "auth_guard_action", Config: { LoginURL: item.authGuardLoginUrl } };
    else if (item.actionName === "forward_request_action") actionObj = { Name: "forward_request_action", Config: { Url: item.forwardUrl, Forward: true } };
    else if (item.actionName === "modify_status_action") actionObj = { Name: "modify_status_action", Config: { Code: Number(item.modifyStatusCode) } };
    else if (item.actionName === "replace_request_body_action") actionObj = { Name: "replace_request_body_action", Config: { Map: tryParse(item.replaceReqBodyMap) } };
    else if (item.actionName === "replace_response_body_action") actionObj = { Name: "replace_response_body_action", Config: { Map: tryParse(item.replaceRespBodyMap) } };
    else if (item.actionName === "replace_request_header_action") actionObj = { Name: "replace_request_header_action", Config: { Map: tryParse(item.replaceReqHeaderMap) } };
    else if (item.actionName === "replace_response_header_action") actionObj = { Name: "replace_response_header_action", Config: { Map: tryParse(item.replaceRespHeaderMap) } };
    else if (item.actionName === "response_text_action") actionObj = { Name: "response_text_action", Config: { Code: String(item.respTextCode), Content: item.respTextContent, Header: tryParse(item.respTextHeader) } };
    else if (item.actionName === "subdomain_webvpn_action") actionObj = { Name: "subdomain_webvpn_action", Config: { Sites: tryParse(item.webvpnSites) } };
    else actionObj = { Name: "reset_conn_action", Config: { Content: item.resetContent } };

    result.push({ Matcher: matcherObj, Action: actionObj });
  }
  newFormInline.value.items = JSON.stringify(result, null, 2);
}

function getMatcherSummary(item: RuleItemConfig): {
  name: string;
  tagType: string;
  summary: string;
} {
  const matchOpt = matcherOptions.find(o => o.value === item.matcherName);
  const tagType = matchOpt ? matchOpt.tagType : "info";

  if (item.matcherName === "always_true_matcher") {
    return { name: "always_true_matcher", tagType, summary: "Match All" };
  } else if (item.matcherName === "http_ip_matcher") {
    const lines = item.httpIpAddressText.split("\n").filter(s => s.trim());
    return { name: "http_ip_matcher", tagType, summary: lines.length > 0 ? `${lines[0]}${lines.length > 1 ? ` (${lines.length})` : ""}` : "No IP" };
  } else if (item.matcherName === "url_matcher") {
    return { name: "url_matcher", tagType, summary: `${item.urlMethod} ${item.urlPath}` };
  } else if (item.matcherName === "js_matcher") {
    return { name: "js_matcher", tagType, summary: "Custom Script" };
  }
  const lines = item.ipAddressText.split("\n").filter(s => s.trim());
  return { name: item.matcherName, tagType, summary: lines.length > 0 ? `${lines[0]}${lines.length > 1 ? ` (${lines.length})` : ""}` : "No Config" };
}

function getActionSummary(item: RuleItemConfig): {
  name: string;
  tagType: string;
  summary: string;
} {
  const actOpt = actionOptions.find(o => o.value === item.actionName);
  const tagType = actOpt ? actOpt.tagType : "info";

  let summary = "";
  if (item.actionName === "reset_conn_action") summary = item.resetContent ? `Msg: ${item.resetContent}` : "Reset Connection";
  else if (item.actionName === "hide_version_action") summary = "Hide Server Version";
  else if (item.actionName === "auth_guard_action") summary = `Guard URL: ${item.authGuardLoginUrl}`;
  else if (item.actionName === "auth_portal_action") summary = "SSO Portal";
  else if (item.actionName === "forward_request_action") summary = `To: ${item.forwardUrl}`;
  else if (item.actionName === "modify_status_action") summary = `Status: ${item.modifyStatusCode}`;
  else if (item.actionName === "replace_request_body_action") summary = "Req Body Map";
  else if (item.actionName === "replace_response_body_action") summary = "Resp Body Map";
  else if (item.actionName === "replace_request_header_action") summary = "Req Header Map";
  else if (item.actionName === "replace_response_header_action") summary = "Resp Header Map";
  else if (item.actionName === "response_text_action") summary = `Code: ${item.respTextCode}`;
  else if (item.actionName === "subdomain_webvpn_action") summary = "WebVPN Sites";

  return {
    name: item.actionName,
    tagType,
    summary
  };
}

watch(itemList, () => syncToFormJSON(), { deep: true });

const rules = reactive({
  name: [
    { required: true, message: () => t("rule.nameRequired"), trigger: "blur" }
  ]
});

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    :label-position="deviceDetection() ? 'top' : 'right'"
    ref="ruleFormRef"
    :model="newFormInline"
    :rules="rules"
    label-width="auto"
    class="rule-form p-1 sm:px-2"
  >
    <!-- Section 1: Basic Info -->
    <el-card
      shadow="never"
      class="mb-4 border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-(--el-color-primary) rounded-full" />
          <span
            class="font-bold text-(--el-text-color-primary) text-sm sm:text-base"
            >{{ t("rule.basicAttr") }}</span
          >
        </div>
      </template>
      <el-row :gutter="16">
        <re-col :value="12" :xs="24">
          <el-form-item :label="t('rule.name')" prop="name">
            <el-input
              v-model="newFormInline.name"
              :placeholder="t('rule.nameRequired')"
              clearable
            />
          </el-form-item>
        </re-col>
        <re-col :value="12" :xs="24">
          <el-form-item :label="t('rule.remark')" prop="remark">
            <el-input
              v-model="newFormInline.remark"
              :placeholder="t('rule.remarkPlaceholder')"
              clearable
            />
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>

    <!-- Section 2: Compact Item List with Scrollbar & Drag-and-Drop -->
    <el-card
      shadow="never"
      class="mb-4 border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div
          class="flex flex-col sm:flex-row sm:items-center justify-between gap-2"
        >
          <div class="flex items-center space-x-2 flex-wrap">
            <div class="w-1.5 h-4 bg-(--el-color-primary) rounded-full" />
            <span
              class="font-bold text-(--el-text-color-primary) text-sm sm:text-base"
              >{{ t("rule.itemListTitle") }}</span
            >
            <el-tag
              size="small"
              type="primary"
              effect="plain"
              class="font-mono"
              >{{ t("rule.totalItems", { count: itemList.length }) }}</el-tag
            >
            <span
              class="text-xs text-(--el-text-color-secondary) hidden sm:inline-block"
              >{{ t("rule.dragTip") }}</span
            >
          </div>
          <el-button
            type="primary"
            size="small"
            :icon="useRenderIcon(AddFill)"
            class="self-start sm:self-auto shrink-0"
            @click="openAddItemDialog"
          >
            {{ t("rule.addItem") }}
          </el-button>
        </div>
      </template>

      <!-- Sleek Custom Scrollbar Container -->
      <el-scrollbar max-height="320px" class="item-scrollbar pr-1">
        <draggable
          v-model="itemList"
          item-key="id"
          handle=".drag-handle"
          ghost-class="opacity-40"
          class="space-y-2.5 pr-1"
          @end="syncToFormJSON"
        >
          <template #item="{ element, index }">
            <div
              class="flex flex-col sm:flex-row sm:items-center justify-between p-3 gap-2.5 rounded-xl border border-(--el-border-color-lighter) bg-(--el-bg-color) hover:border-(--el-color-primary-light-5) transition-all shadow-2xs"
            >
              <!-- Drag Handle & Index Badge -->
              <div class="flex items-center space-x-3">
                <span
                  class="drag-handle cursor-move text-(--el-text-color-placeholder) hover:text-(--el-color-primary) p-1 rounded"
                  title="Drag to reorder"
                >
                  <component :is="useRenderIcon(DragIcon)" class="size-4" />
                </span>
                <el-tag
                  size="small"
                  type="info"
                  effect="plain"
                  class="font-mono font-bold"
                >
                  #{{ index + 1 }}
                </el-tag>
              </div>

              <!-- Matcher & Action Summaries -->
              <div
                class="flex-1 grid grid-cols-1 sm:grid-cols-2 gap-2 sm:gap-3 px-1 sm:px-3"
              >
                <!-- Matcher Badge & Text -->
                <div class="flex items-center space-x-2 overflow-hidden">
                  <el-tag
                    size="small"
                    :type="getMatcherSummary(element).tagType as any"
                    effect="light"
                    class="font-mono shrink-0"
                  >
                    {{ getMatcherSummary(element).name }}
                  </el-tag>
                  <span
                    class="text-xs text-(--el-text-color-regular) truncate font-mono"
                  >
                    {{ getMatcherSummary(element).summary }}
                  </span>
                </div>

                <!-- Action Badge & Text -->
                <div class="flex items-center space-x-2 overflow-hidden">
                  <el-tag
                    size="small"
                    :type="getActionSummary(element).tagType as any"
                    effect="light"
                    class="font-mono shrink-0"
                  >
                    {{ getActionSummary(element).name }}
                  </el-tag>
                  <span class="text-xs text-(--el-text-color-regular) truncate">
                    {{ getActionSummary(element).summary }}
                  </span>
                </div>
              </div>

              <!-- Item Row Operations -->
              <div
                class="flex items-center space-x-1 shrink-0 self-end sm:self-auto border-t sm:border-t-0 border-(--el-border-color-lighter) pt-2 sm:pt-0 w-full sm:w-auto justify-end"
              >
                <el-button
                  size="small"
                  link
                  :disabled="index === 0"
                  :icon="useRenderIcon(ArrowUp)"
                  @click="moveItemUp(index)"
                />
                <el-button
                  size="small"
                  link
                  :disabled="index === itemList.length - 1"
                  :icon="useRenderIcon(ArrowDown)"
                  @click="moveItemDown(index)"
                />
                <el-button
                  size="small"
                  link
                  type="primary"
                  :icon="useRenderIcon(EditPen)"
                  @click="openEditItemDialog(index)"
                >
                  {{ t("rule.edit") }}
                </el-button>
                <el-button
                  size="small"
                  link
                  type="danger"
                  
                  :icon="useRenderIcon(Delete)"
                  @click="removeItem(index)"
                >
                  {{ t("rule.delete") }}
                </el-button>
              </div>
            </div>
          </template>
        </draggable>

        <div
          v-if="itemList.length === 0"
          class="text-center py-6 text-xs text-(--el-text-color-placeholder)"
        >
          {{ t("rule.noItems") }}
        </div>
      </el-scrollbar>
    </el-card>

    <!-- Theme-aware JSON Preview Box -->
    <div
      class="rounded-xl border border-(--el-border-color-lighter) bg-(--el-fill-color-light) p-3"
    >
      <div class="flex-bc mb-2">
        <span class="text-xs font-semibold text-(--el-text-color-secondary)">{{
          t("rule.jsonPreview")
        }}</span>
        <el-tag size="small" type="info" effect="plain" class="font-mono"
          >JSON Preview</el-tag
        >
      </div>
      <div
        class="bg-(--el-bg-color) p-2.5 rounded-lg border border-(--el-border-color-lighter)"
      >
        <el-scrollbar max-height="150px" class="item-scrollbar pr-1">
          <pre
            class="text-[11px] text-(--el-text-color-primary) font-mono whitespace-pre-wrap break-all leading-relaxed"
            >{{ newFormInline.items }}</pre>
        </el-scrollbar>
      </div>
    </div>

    <!-- Nested Sub-Dialog for Adding / Editing a Single Rule Item -->
    <el-dialog
      v-model="itemDialogVisible"
      :title="
        editingIndex !== null
          ? `${t('rule.editItem')} (#${editingIndex + 1})`
          : t('rule.addItem')
      "
      :width="deviceDetection() ? '92%' : '640px'"
      append-to-body
      destroy-on-close
      class="item-config-dialog"
    >
      <div class="space-y-4 py-1">
        <!-- Matcher Card -->
        <div
          class="p-3.5 rounded-lg border border-(--el-border-color-lighter) bg-(--el-fill-color-light)"
        >
          <div class="text-xs font-bold text-(--el-color-primary) mb-3 flex-bc">
            <span class="text-sm">{{ t("rule.matcherConfigTitle") }}</span>
            <el-tag size="small" type="primary" effect="plain">Matcher</el-tag>
          </div>

          <el-form-item :label="t('rule.selectMatcher')">
            <el-select v-model="itemForm.matcherName" class="w-full">
              <el-option
                v-for="m in matcherOptions"
                :key="m.value"
                :label="m.label"
                :value="m.value"
              />
            </el-select>
          </el-form-item>

          <div v-if="itemForm.matcherName === 'always_true_matcher'" class="mt-2">
            <el-alert :title="t('rule.matcherAlwaysTrueTip', '无条件匹配所有流量')" type="info" :closable="false" show-icon />
          </div>

          <div v-if="itemForm.matcherName === 'ip_matcher' || itemForm.matcherName === 'cidr_matcher'" class="mt-2">
            <el-input
              v-model="itemForm.ipAddressText"
              type="textarea"
              :rows="3"
              placeholder="127.0.0.1\n192.168.1.0/24\n10.0.0.1-10.0.0.100"
            />
            <div class="text-xs text-(--el-text-color-secondary) mt-1">
              {{ t("rule.matcherIpTip", "多行输入，支持单IP、网段或区间") }}
            </div>
          </div>

          <div v-if="itemForm.matcherName === 'http_ip_matcher'" class="mt-2">
            <el-input
              v-model="itemForm.httpIpAddressText"
              type="textarea"
              :rows="3"
              placeholder="127.0.0.1\n192.168.1.0/24"
            />
            <div class="text-xs text-(--el-text-color-secondary) mt-1">
              {{ t("rule.matcherHttpIpTip", "多行输入，支持单IP或CIDR") }}
            </div>
          </div>

          <div v-if="itemForm.matcherName === 'url_matcher'" class="mt-2 space-y-2">
            <el-select v-model="itemForm.urlMethod" class="w-full" placeholder="Method">
              <el-option label="ALL" value="ALL" />
              <el-option label="GET" value="GET" />
              <el-option label="POST" value="POST" />
              <el-option label="PUT" value="PUT" />
              <el-option label="DELETE" value="DELETE" />
              <el-option label="HEAD" value="HEAD" />
              <el-option label="OPTIONS" value="OPTIONS" />
              <el-option label="PATCH" value="PATCH" />
              <el-option label="TRACE" value="TRACE" />
              <el-option label="CONNECT" value="CONNECT" />
            </el-select>
            <el-input v-model="itemForm.urlPath" placeholder="URL Path (e.g., /api/ or ~^/img/)" />
            <div class="text-xs text-(--el-text-color-secondary) mt-1">
              {{ t("rule.matcherUrlTip", "支持前缀匹配，~开头表示正则匹配") }}
            </div>
          </div>

          <div v-if="itemForm.matcherName === 'js_matcher'" class="mt-2">
            <el-input v-model="itemForm.jsScript" type="textarea" :rows="4" placeholder="return true;" class="font-mono text-xs" />
            <div class="text-xs text-(--el-text-color-secondary) mt-1">
              {{ t("rule.matcherJsTip", "使用 get_url(), get_header(key) 进行判断，需返回 boolean") }}
            </div>
          </div>
        </div>

        <!-- Action Card -->
        <div
          class="p-3.5 rounded-lg border border-(--el-border-color-lighter) bg-(--el-fill-color-light)"
        >
          <div class="text-xs font-bold text-(--el-color-success) mb-3 flex-bc">
            <span class="text-sm">{{ t("rule.actionConfigTitle") }}</span>
            <el-tag size="small" type="success" effect="plain">Action</el-tag>
          </div>

          <el-form-item :label="t('rule.selectAction')">
            <el-select v-model="itemForm.actionName" class="w-full">
              <el-option
                v-for="a in actionOptions"
                :key="a.value"
                :label="a.label"
                :value="a.value"
              />
            </el-select>
          </el-form-item>

          <div v-if="itemForm.actionName === 'reset_conn_action'" class="mt-2">
            <el-input
              v-model="itemForm.resetContent"
              placeholder="Connection reset by rule"
            />
            <div class="text-xs text-(--el-text-color-secondary) mt-1">
              {{ t("rule.actionResetTip", "TCP直接阻断并发送此文本") }}
            </div>
          </div>

          <div v-if="itemForm.actionName === 'hide_version_action'" class="mt-2">
            <el-alert
              :title="t('rule.actionHideVersionAlert', '将隐藏 HTTP 响应头中的 Server 等服务器版本信息')"
              type="info"
              :closable="false"
              show-icon
            />
          </div>

          <div v-if="itemForm.actionName === 'auth_guard_action'" class="mt-2">
            <el-input v-model="itemForm.authGuardLoginUrl" placeholder="/login" />
            <div class="text-xs text-(--el-text-color-secondary) mt-1">
              {{ t("rule.actionAuthGuardTip", "配置拦截时的登录跳转地址") }}
            </div>
          </div>

          <div v-if="itemForm.actionName === 'auth_portal_action'" class="mt-2">
            <el-alert :title="t('rule.actionAuthPortalAlert', '提供内置的 SSO Portal 登录面板与接口')" type="info" :closable="false" show-icon />
          </div>

          <div v-if="itemForm.actionName === 'forward_request_action'" class="mt-2">
            <el-input v-model="itemForm.forwardUrl" placeholder="http://other-backend.com" />
            <div class="text-xs text-(--el-text-color-secondary) mt-1">
              {{ t("rule.actionForwardTip", "命中时直接将流量重写转发到此地址") }}
            </div>
          </div>

          <div v-if="itemForm.actionName === 'modify_status_action'" class="mt-2">
            <el-input-number v-model="itemForm.modifyStatusCode" :min="100" :max="599" class="w-full" />
            <div class="text-xs text-(--el-text-color-secondary) mt-1">
              {{ t("rule.actionModifyStatusTip", "强制覆盖 HTTP 响应状态码") }}
            </div>
          </div>

          <div v-if="['replace_request_body_action', 'replace_response_body_action', 'replace_request_header_action', 'replace_response_header_action'].includes(itemForm.actionName)" class="mt-2">
            <el-input v-model="itemForm.replaceReqBodyMap" type="textarea" :rows="4" class="font-mono text-xs" placeholder='{ "old": "new" }' v-if="itemForm.actionName === 'replace_request_body_action'" />
            <el-input v-model="itemForm.replaceRespBodyMap" type="textarea" :rows="4" class="font-mono text-xs" placeholder='{ "old": "new" }' v-if="itemForm.actionName === 'replace_response_body_action'" />
            <el-input v-model="itemForm.replaceReqHeaderMap" type="textarea" :rows="4" class="font-mono text-xs" placeholder='{ "old": "new" }' v-if="itemForm.actionName === 'replace_request_header_action'" />
            <el-input v-model="itemForm.replaceRespHeaderMap" type="textarea" :rows="4" class="font-mono text-xs" placeholder='{ "old": "new" }' v-if="itemForm.actionName === 'replace_response_header_action'" />
            <div class="text-xs text-(--el-text-color-secondary) mt-1">
              {{ t("rule.actionReplaceMapTip", "JSON格式 Map，Key为被替换的正则表达式，Value为替换后内容") }}
            </div>
          </div>

          <div v-if="itemForm.actionName === 'response_text_action'" class="mt-2 space-y-2">
            <el-input v-model="itemForm.respTextCode" placeholder="HTTP Status Code (e.g., 403)" />
            <el-input v-model="itemForm.respTextContent" type="textarea" :rows="2" placeholder="Response Content" />
            <el-input v-model="itemForm.respTextHeader" type="textarea" :rows="2" class="font-mono text-xs" placeholder='{ "Content-Type": "text/plain" }' />
          </div>

          <div v-if="itemForm.actionName === 'subdomain_webvpn_action'" class="mt-2">
            <el-input v-model="itemForm.webvpnSites" type="textarea" :rows="6" class="font-mono text-xs" placeholder='{ "site_1": { "Protected": true, "Host": { "real.com": "vpn-real.com" } } }' />
            <div class="text-xs text-(--el-text-color-secondary) mt-1">
              {{ t("rule.actionWebvpnTip", "高级WebVPN配置，格式请参考官方文档") }}
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end space-x-2">
          <el-button @click="itemDialogVisible = false">{{
            t("rule.cancel")
          }}</el-button>
          <el-button type="primary" @click="saveItemFromDialog">{{
            t("rule.saveItem")
          }}</el-button>
        </div>
      </template>
    </el-dialog>
  </el-form>
</template>

<style scoped>
.rule-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--el-text-color-regular);
  white-space: nowrap;
}
</style>
