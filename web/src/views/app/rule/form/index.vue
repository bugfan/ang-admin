<script setup lang="ts">
import { ref, reactive, watch, computed } from "vue";
import ReCol from "@/components/ReCol";
import { useI18n } from "vue-i18n";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { deviceDetection } from "@pureadmin/utils";
import draggable from "vuedraggable";
import { message } from "@/utils/message";

import AddFill from "~icons/ri/add-line";
import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import ArrowUp from "~icons/ep/arrow-up";
import ArrowDown from "~icons/ep/arrow-down";
import DragIcon from "~icons/ri/drag-move-2-line";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";
import InfoIcon from "~icons/ep/info-filled";
import PlusIcon from "~icons/ep/plus";

// ─── Types ──────────────────────────────────────────────────────────────────

interface KVPair {
  k: string;
  v: string;
}

interface WebvpnSiteEntry {
  id: string;
  siteKey: string;
  protected: boolean;
  loginURL: string;
  hosts: string;
  replace: KVPair[];
}

interface RuleItemConfig {
  id: string;
  // Matcher
  matcherName: string;
  ipAddressText: string;
  httpIpAddressText: string;
  urlMethod: string;
  urlPath: string;
  jsScript: string;
  // Action
  actionName: string;
  resetContent: string;
  modifyStatusCode: number;
  replaceReqBodyScheme: string;
  replaceReqBodyMap: KVPair[];
  replaceRespBodyScheme: string;
  replaceRespBodyMap: KVPair[];
  replaceReqHeaderMap: KVPair[];
  replaceRespHeaderMap: KVPair[];
  respTextCode: string;
  respTextContent: string;
  respTextHeader: KVPair[];
  insertContent: string;
  insertPosition: string;
  insertRegexp: string;
  webvpnLoginURL: string;
  webvpnSites: WebvpnSiteEntry[];
  // Auth Portal
  authPortalTitle: string;
  authPortalTokenName: string;
  authPortalTokenExpire: number;
  authPortalCookieDomain: string;
  // Auth Guard
  authGuardPortalURL: string;
}

// ─── Props ───────────────────────────────────────────────────────────────────

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "",
    id: undefined,
    name: "",
    items: "[]",
    remark: ""
  })
});

const { t } = useI18n();

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);

// ─── Options ─────────────────────────────────────────────────────────────────

const matcherOptions = computed(() => [
  {
    label: t("rule.matcherAlwaysTrue"),
    value: "always_true_matcher",
    tag: "L4",
    tagType: "primary"
  },
  {
    label: t("rule.matcherIp"),
    value: "ip_matcher",
    tag: "L4",
    tagType: "primary"
  },
  {
    label: t("rule.matcherCidr"),
    value: "cidr_matcher",
    tag: "L4",
    tagType: "primary"
  },
  {
    label: t("rule.matcherHttpIp"),
    value: "http_ip_matcher",
    tag: "HTTP",
    tagType: "success"
  },
  {
    label: t("rule.matcherUrl"),
    value: "url_matcher",
    tag: "HTTP",
    tagType: "success"
  },
  {
    label: t("rule.matcherJs"),
    value: "js_matcher",
    tag: "HTTP",
    tagType: "success"
  }
]);

const actionOptions = computed(() => [
  {
    label: t("rule.actionResetConn"),
    value: "reset_conn_action",
    tag: "L4",
    tagType: "danger"
  },
  {
    label: t("rule.actionHideVersion"),
    value: "hide_version_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: t("rule.actionModifyStatus"),
    value: "modify_status_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: t("rule.actionReplaceReqBody"),
    value: "replace_request_body_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: t("rule.actionReplaceRespBody"),
    value: "replace_response_body_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: t("rule.actionReplaceReqHeader"),
    value: "replace_request_header_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: t("rule.actionReplaceRespHeader"),
    value: "replace_response_header_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: t("rule.actionResponseText"),
    value: "response_text_action",
    tag: "HTTP",
    tagType: "warning"
  },
  {
    label: t("rule.actionInsertData"),
    value: "insert_data_action",
    tag: "HTTP",
    tagType: "info"
  },
  {
    label: t("rule.actionSubdomainWebvpn"),
    value: "subdomain_webvpn_action",
    tag: "HTTP",
    tagType: "info"
  },
  {
    label: t("rule.actionAuthPortal"),
    value: "auth_portal_action",
    tag: "HTTP",
    tagType: "success"
  },
  {
    label: t("rule.actionAuthGuard"),
    value: "auth_guard_action",
    tag: "HTTP",
    tagType: "success"
  }
]);

const HTTP_METHODS = [
  "ALL",
  "GET",
  "POST",
  "PUT",
  "DELETE",
  "HEAD",
  "OPTIONS",
  "PATCH",
  "TRACE",
  "CONNECT"
];

const HTTP_STATUS_CODES = [
  { label: "200 OK", value: "200" },
  { label: "301 Moved Permanently", value: "301" },
  { label: "302 Found", value: "302" },
  { label: "400 Bad Request", value: "400" },
  { label: "401 Unauthorized", value: "401" },
  { label: "403 Forbidden", value: "403" },
  { label: "404 Not Found", value: "404" },
  { label: "500 Internal Server Error", value: "500" },
  { label: "502 Bad Gateway", value: "502" },
  { label: "503 Service Unavailable", value: "503" }
];

const BODY_SCHEMES = [
  { label: "HTTP + WebSocket (ALL)", value: "ALL" },
  { label: "HTTP Only", value: "HTTP" },
  { label: "WebSocket Only", value: "WS" }
];

const INSERT_POSITIONS = computed(() => [
  { label: "定位之前 (before)", value: "before" },
  { label: "定位之后 (after)", value: "after" }
]);

const INSERT_REGEXP_OPTIONS = [
  { label: "<head>", value: "<head>" },
  { label: "</head>", value: "</head>" },
  { label: "<body>", value: "<body>" },
  { label: "</body>", value: "</body>" }
];

// ─── Item List ────────────────────────────────────────────────────────────────

const itemList = ref<RuleItemConfig[]>([]);

// ─── Inline editor state ──────────────────────────────────────────────────────

const editorVisible = ref(false);
const editingIndex = ref<number | null>(null);
const editorTitle = computed(() =>
  editingIndex.value !== null
    ? `${t("rule.editItem")} #${editingIndex.value + 1}`
    : t("rule.addItem")
);

const itemForm = reactive<RuleItemConfig>({
  id: "",
  matcherName: "ip_matcher",
  ipAddressText: "",
  httpIpAddressText: "",
  urlMethod: "ALL",
  urlPath: "/",
  jsScript: "return true;",
  actionName: "reset_conn_action",
  resetContent: "Connection reset by rule",
  modifyStatusCode: 403,
  replaceReqBodyScheme: "ALL",
  replaceReqBodyMap: [],
  replaceRespBodyScheme: "ALL",
  replaceRespBodyMap: [],
  replaceReqHeaderMap: [],
  replaceRespHeaderMap: [],
  respTextCode: "403",
  respTextContent: "",
  respTextHeader: [],
  insertContent: "",
  insertPosition: "before",
  insertRegexp: "</body>",
  webvpnLoginURL: "",
  webvpnSites: [],
  authPortalTitle: "统一身份认证",
  authPortalTokenName: "_angt",
  authPortalTokenExpire: 86400,
  authPortalCookieDomain: "",
  authGuardPortalURL: ""
});

// ─── Init & Conversion Helpers ─────────────────────────────────────────────

function mapToKV(obj: Record<string, string> | undefined): KVPair[] {
  if (!obj || typeof obj !== "object") return [];
  return Object.entries(obj).map(([k, v]) => ({ k, v: String(v) }));
}

function kvToMap(pairs: KVPair[]): Record<string, string> {
  const m: Record<string, string> = {};
  for (const { k, v } of pairs) {
    if (k && k.trim()) m[k.trim()] = v || "";
  }
  return m;
}

function parseSites(raw: any): WebvpnSiteEntry[] {
  if (!raw || typeof raw !== "object") return [];
  return Object.entries(raw).map(([siteKey, siteVal]: any) => {
    const hostKeys = Object.keys(siteVal?.host || {});
    const wildcardKeys = Object.keys(siteVal?.wildcard || {});
    const hosts = [...hostKeys, ...wildcardKeys].join("\n");
    return {
      id: String(Math.random()),
      siteKey,
      protected: !!siteVal?.protected,
      loginURL: siteVal?.loginURL || siteVal?.LoginURL || "",
      hosts,
      replace: mapToKV(siteVal?.replace)
    };
  });
}

function sitesToRaw(sites: WebvpnSiteEntry[]): Record<string, any> {
  const result: Record<string, any> = {};
  for (const s of sites) {
    if (!s.siteKey.trim()) continue;
    const hostMap: Record<string, string> = {};
    const wildcardMap: Record<string, string> = {};
    for (const line of s.hosts.split("\n")) {
      const domain = line.trim();
      if (!domain) continue;
      if (domain.includes("*")) {
        wildcardMap[domain] = "";
      } else {
        hostMap[domain] = "";
      }
    }
    result[s.siteKey.trim()] = {
      protected: s.protected,
      loginURL: s.loginURL,
      host: hostMap,
      wildcard: wildcardMap,
      replace: kvToMap(s.replace)
    };
  }
  return result;
}

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
    resetContent: "Connection reset by rule",
    modifyStatusCode: 403,
    replaceReqBodyScheme: "ALL",
    replaceReqBodyMap: [],
    replaceRespBodyScheme: "ALL",
    replaceRespBodyMap: [],
    replaceReqHeaderMap: [],
    replaceRespHeaderMap: [],
    respTextCode: "403",
    respTextContent: "",
    respTextHeader: [],
    insertContent: "",
    insertPosition: "before",
    insertRegexp: "</body>",
    webvpnLoginURL: "",
    webvpnSites: [],
    authPortalTitle: "统一身份认证",
    authPortalTokenName: "_angt",
    authPortalTokenExpire: 86400,
    authPortalCookieDomain: "",
    authGuardPortalURL: ""
  };
}

function parseItemFromJSON(item: any): RuleItemConfig {
  const mObj = item?.Matcher || item?.matcher || {};
  const aObj = item?.Action || item?.action || {};
  const mName = mObj?.Name || mObj?.name || "ip_matcher";
  const mCfg = mObj?.Config || mObj?.config || {};
  const aName = aObj?.Name || aObj?.name || "reset_conn_action";
  const aCfg = aObj?.Config || aObj?.config || {};

  const base = createDefaultItem();

  // Matcher fields
  base.matcherName = mName;
  if (mName === "ip_matcher" && Array.isArray(mCfg.Address))
    base.ipAddressText = mCfg.Address.join("\n");
  else if (mName === "cidr_matcher" && Array.isArray(mCfg.CIDRs))
    base.ipAddressText = mCfg.CIDRs.join("\n");
  else if (mName === "http_ip_matcher" && Array.isArray(mCfg.Address))
    base.httpIpAddressText = mCfg.Address.join("\n");
  else if (mName === "url_matcher") {
    base.urlMethod = mCfg.Method || "ALL";
    base.urlPath = mCfg.Path || "/";
  } else if (mName === "js_matcher") {
    base.jsScript = mCfg.Script || "return true;";
  }

  // Action fields
  base.actionName = aName;
  if (aName === "reset_conn_action") base.resetContent = aCfg.Content || "";
  if (aName === "modify_status_action")
    base.modifyStatusCode = aCfg.Code || 403;
  if (aName === "replace_request_body_action") {
    base.replaceReqBodyScheme =
      aCfg.Scheme && aCfg.Scheme !== "" ? aCfg.Scheme : "ALL";
    base.replaceReqBodyMap = mapToKV(aCfg.Map);
  }
  if (aName === "replace_response_body_action") {
    base.replaceRespBodyScheme =
      aCfg.Scheme && aCfg.Scheme !== "" ? aCfg.Scheme : "ALL";
    base.replaceRespBodyMap = mapToKV(aCfg.Map);
  }
  if (aName === "replace_request_header_action")
    base.replaceReqHeaderMap = mapToKV(aCfg.Map);
  if (aName === "replace_response_header_action")
    base.replaceRespHeaderMap = mapToKV(aCfg.Map);
  if (aName === "response_text_action") {
    base.respTextCode = String(aCfg.Code || "403");
    base.respTextContent = aCfg.Content || "";
    base.respTextHeader = mapToKV(aCfg.Header);
  }
  if (aName === "insert_data_action") {
    base.insertContent = aCfg.Content || "";
    base.insertPosition = aCfg.Position || "before";
    base.insertRegexp = aCfg.Regexp || "</body>";
  }
  if (aName === "subdomain_webvpn_action") {
    base.webvpnLoginURL = aCfg.LoginURL || "";
    base.webvpnSites = parseSites(aCfg.Sites);
  }
  if (aName === "auth_portal_action") {
    base.authPortalTitle = aCfg.title || aCfg.Title || "统一身份认证";
    base.authPortalTokenName = aCfg.token_name || aCfg.TokenName || "_angt";
    base.authPortalTokenExpire = aCfg.token_expire || aCfg.TokenExpire || 86400;
    base.authPortalCookieDomain = aCfg.cookie_domain || aCfg.CookieDomain || "";
  }
  if (aName === "auth_guard_action") {
    base.authGuardPortalURL = aCfg.portal_url || aCfg.PortalURL || "";
  }

  return base;
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
          itemList.value.push(parseItemFromJSON(item));
        }
      }
    }
  } catch (e) {}
  syncToFormJSON();
}

onInitForm();

// ─── Sync to JSON ─────────────────────────────────────────────────────────────

function itemToJSON(item: RuleItemConfig): any {
  let matcherObj: any;
  switch (item.matcherName) {
    case "always_true_matcher":
      matcherObj = { Name: "always_true_matcher", Config: {} };
      break;
    case "http_ip_matcher":
      matcherObj = {
        Name: "http_ip_matcher",
        Config: {
          Address: item.httpIpAddressText
            .split("\n")
            .map(s => s.trim())
            .filter(Boolean)
        }
      };
      break;
    case "url_matcher":
      matcherObj = {
        Name: "url_matcher",
        Config: { Method: item.urlMethod, Path: item.urlPath }
      };
      break;
    case "js_matcher":
      matcherObj = { Name: "js_matcher", Config: { Script: item.jsScript } };
      break;
    case "cidr_matcher":
      matcherObj = {
        Name: "cidr_matcher",
        Config: {
          CIDRs: item.ipAddressText
            .split("\n")
            .map(s => s.trim())
            .filter(Boolean)
        }
      };
      break;
    default:
      matcherObj = {
        Name: "ip_matcher",
        Config: {
          Address: item.ipAddressText
            .split("\n")
            .map(s => s.trim())
            .filter(Boolean)
        }
      };
  }

  let actionObj: any;
  switch (item.actionName) {
    case "hide_version_action":
      actionObj = { Name: "hide_version_action", Config: {} };
      break;
    case "modify_status_action":
      actionObj = {
        Name: "modify_status_action",
        Config: { Code: Number(item.modifyStatusCode) }
      };
      break;
    case "replace_request_body_action":
      actionObj = {
        Name: "replace_request_body_action",
        Config: {
          Scheme:
            item.replaceReqBodyScheme === "ALL"
              ? ""
              : item.replaceReqBodyScheme,
          Map: kvToMap(item.replaceReqBodyMap)
        }
      };
      break;
    case "replace_response_body_action":
      actionObj = {
        Name: "replace_response_body_action",
        Config: {
          Scheme:
            item.replaceRespBodyScheme === "ALL"
              ? ""
              : item.replaceRespBodyScheme,
          Map: kvToMap(item.replaceRespBodyMap)
        }
      };
      break;
    case "replace_request_header_action":
      actionObj = {
        Name: "replace_request_header_action",
        Config: { Map: kvToMap(item.replaceReqHeaderMap) }
      };
      break;
    case "replace_response_header_action":
      actionObj = {
        Name: "replace_response_header_action",
        Config: { Map: kvToMap(item.replaceRespHeaderMap) }
      };
      break;
    case "response_text_action":
      actionObj = {
        Name: "response_text_action",
        Config: {
          Code: String(item.respTextCode),
          Content: item.respTextContent,
          Header: kvToMap(item.respTextHeader)
        }
      };
      break;
    case "insert_data_action":
      actionObj = {
        Name: "insert_data_action",
        Config: {
          Content: item.insertContent,
          Position: item.insertPosition,
          Regexp: item.insertRegexp
        }
      };
      break;
    case "subdomain_webvpn_action":
      actionObj = {
        Name: "subdomain_webvpn_action",
        Config: {
          LoginURL: item.webvpnLoginURL,
          Sites: sitesToRaw(item.webvpnSites)
        }
      };
      break;
    case "auth_portal_action":
      actionObj = {
        Name: "auth_portal_action",
        Config: {
          title: item.authPortalTitle || "统一身份认证",
          token_name: item.authPortalTokenName || "_angt",
          token_expire: Number(item.authPortalTokenExpire) || 86400,
          cookie_domain: item.authPortalCookieDomain || ""
        }
      };
      break;
    case "auth_guard_action":
      actionObj = {
        Name: "auth_guard_action",
        Config: {
          portal_url: item.authGuardPortalURL || ""
        }
      };
      break;
    default:
      actionObj = {
        Name: "reset_conn_action",
        Config: { Content: item.resetContent }
      };
  }

  return { Matcher: matcherObj, Action: actionObj };
}

function syncToFormJSON() {
  const result = itemList.value.map(itemToJSON);
  newFormInline.value.items = JSON.stringify(result, null, 2);
}

watch(itemList, () => syncToFormJSON(), { deep: true });

watch(
  () => itemForm.actionName,
  newAction => {
    if (
      newAction === "replace_request_body_action" &&
      itemForm.replaceReqBodyMap.length === 0
    ) {
      itemForm.replaceReqBodyMap.push({ k: "", v: "" });
    } else if (
      newAction === "replace_response_body_action" &&
      itemForm.replaceRespBodyMap.length === 0
    ) {
      itemForm.replaceRespBodyMap.push({ k: "", v: "" });
    } else if (
      newAction === "replace_request_header_action" &&
      itemForm.replaceReqHeaderMap.length === 0
    ) {
      itemForm.replaceReqHeaderMap.push({ k: "", v: "" });
    } else if (
      newAction === "replace_response_header_action" &&
      itemForm.replaceRespHeaderMap.length === 0
    ) {
      itemForm.replaceRespHeaderMap.push({ k: "", v: "" });
    } else if (newAction === "insert_data_action" && !itemForm.insertRegexp) {
      itemForm.insertRegexp = "</body>";
    }
  }
);

// ─── Editor Open/Close ────────────────────────────────────────────────────────

function ensureKVRow() {
  if (
    itemForm.actionName === "replace_request_body_action" &&
    itemForm.replaceReqBodyMap.length === 0
  ) {
    itemForm.replaceReqBodyMap.push({ k: "", v: "" });
  } else if (
    itemForm.actionName === "replace_response_body_action" &&
    itemForm.replaceRespBodyMap.length === 0
  ) {
    itemForm.replaceRespBodyMap.push({ k: "", v: "" });
  } else if (
    itemForm.actionName === "replace_request_header_action" &&
    itemForm.replaceReqHeaderMap.length === 0
  ) {
    itemForm.replaceReqHeaderMap.push({ k: "", v: "" });
  } else if (
    itemForm.actionName === "replace_response_header_action" &&
    itemForm.replaceRespHeaderMap.length === 0
  ) {
    itemForm.replaceRespHeaderMap.push({ k: "", v: "" });
  }
}

function copyItemToForm(src: RuleItemConfig) {
  Object.assign(itemForm, {
    ...src,
    replaceReqBodyMap: src.replaceReqBodyMap.map(p => ({ ...p })),
    replaceRespBodyMap: src.replaceRespBodyMap.map(p => ({ ...p })),
    replaceReqHeaderMap: src.replaceReqHeaderMap.map(p => ({ ...p })),
    replaceRespHeaderMap: src.replaceRespHeaderMap.map(p => ({ ...p })),
    respTextHeader: src.respTextHeader.map(p => ({ ...p })),
    webvpnSites: src.webvpnSites.map(s => ({
      ...s,
      replace: s.replace.map(p => ({ ...p }))
    }))
  });
  ensureKVRow();
}

function openAddItemEditor() {
  editingIndex.value = null;
  copyItemToForm(createDefaultItem());
  editorVisible.value = true;
}

function openEditItemEditor(index: number) {
  editingIndex.value = index;
  copyItemToForm(itemList.value[index]);
  editorVisible.value = true;
}

function closeEditor() {
  editorVisible.value = false;
}

// ─── Comprehensive Item Validation ───────────────────────────────────────────

function validateItemForm(): boolean {
  // 1. Validate Matcher
  if (itemForm.matcherName === "ip_matcher") {
    const validLines = itemForm.ipAddressText.split("\n").filter(s => s.trim());
    if (validLines.length === 0) {
      message(t("rule.valIpRequired"), { type: "warning" });
      return false;
    }
  } else if (itemForm.matcherName === "cidr_matcher") {
    const validLines = itemForm.ipAddressText.split("\n").filter(s => s.trim());
    if (validLines.length === 0) {
      message(t("rule.valCidrRequired"), { type: "warning" });
      return false;
    }
  } else if (itemForm.matcherName === "http_ip_matcher") {
    const validLines = itemForm.httpIpAddressText
      .split("\n")
      .filter(s => s.trim());
    if (validLines.length === 0) {
      message(t("rule.valHttpIpRequired"), { type: "warning" });
      return false;
    }
  } else if (itemForm.matcherName === "url_matcher") {
    if (!itemForm.urlPath || !itemForm.urlPath.trim()) {
      message(t("rule.valUrlPathRequired"), { type: "warning" });
      return false;
    }
  } else if (itemForm.matcherName === "js_matcher") {
    if (!itemForm.jsScript || !itemForm.jsScript.trim()) {
      message(t("rule.valJsScriptRequired"), { type: "warning" });
      return false;
    }
  }

  // 2. Validate Action
  if (itemForm.actionName === "modify_status_action") {
    const code = Number(itemForm.modifyStatusCode);
    if (isNaN(code) || code < 100 || code > 599) {
      message(t("rule.valModifyStatusInvalid"), { type: "warning" });
      return false;
    }
  } else if (itemForm.actionName === "replace_request_body_action") {
    const valid = itemForm.replaceReqBodyMap.some(p => p.k && p.k.trim());
    if (!valid) {
      message(t("rule.valReplaceReqBodyRequired"), { type: "warning" });
      return false;
    }
  } else if (itemForm.actionName === "replace_response_body_action") {
    const valid = itemForm.replaceRespBodyMap.some(p => p.k && p.k.trim());
    if (!valid) {
      message(t("rule.valReplaceRespBodyRequired"), { type: "warning" });
      return false;
    }
  } else if (itemForm.actionName === "replace_request_header_action") {
    const valid = itemForm.replaceReqHeaderMap.some(p => p.k && p.k.trim());
    if (!valid) {
      message(t("rule.valReplaceReqHeaderRequired"), { type: "warning" });
      return false;
    }
  } else if (itemForm.actionName === "replace_response_header_action") {
    const valid = itemForm.replaceRespHeaderMap.some(p => p.k && p.k.trim());
    if (!valid) {
      message(t("rule.valReplaceRespHeaderRequired"), { type: "warning" });
      return false;
    }
  } else if (itemForm.actionName === "response_text_action") {
    if (!itemForm.respTextCode || !String(itemForm.respTextCode).trim()) {
      message(t("rule.valRespTextCodeRequired"), { type: "warning" });
      return false;
    }
  } else if (itemForm.actionName === "insert_data_action") {
    if (!itemForm.insertRegexp || !itemForm.insertRegexp.trim()) {
      message(t("rule.valInsertRegexpRequired"), { type: "warning" });
      return false;
    }
    if (!itemForm.insertContent || !itemForm.insertContent.trim()) {
      message(t("rule.valInsertContentRequired"), { type: "warning" });
      return false;
    }
  } else if (itemForm.actionName === "subdomain_webvpn_action") {
    if (itemForm.webvpnSites.length === 0) {
      message(t("rule.valWebvpnSitesRequired"), { type: "warning" });
      return false;
    }
    for (let i = 0; i < itemForm.webvpnSites.length; i++) {
      const site = itemForm.webvpnSites[i];
      if (!site.siteKey || !site.siteKey.trim()) {
        message(t("rule.valWebvpnSiteKeyRequired"), { type: "warning" });
        return false;
      }
      const validHosts = site.hosts.split("\n").filter(s => s.trim());
      if (validHosts.length === 0) {
        message(t("rule.valWebvpnSiteHostsRequired"), { type: "warning" });
        return false;
      }
    }
  } else if (itemForm.actionName === "auth_guard_action") {
    if (!itemForm.authGuardPortalURL || !itemForm.authGuardPortalURL.trim()) {
      message(t("rule.valAuthGuardPortalUrlRequired"), { type: "warning" });
      return false;
    }
    const pUrl = itemForm.authGuardPortalURL.trim();
    if (!pUrl.startsWith("http://") && !pUrl.startsWith("https://")) {
      message(t("rule.valAuthGuardPortalUrlInvalid"), { type: "warning" });
      return false;
    }
  }

  return true;
}

function saveItemFromEditor() {
  if (!validateItemForm()) return;

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
    modifyStatusCode: itemForm.modifyStatusCode,
    replaceReqBodyScheme: itemForm.replaceReqBodyScheme,
    replaceReqBodyMap: itemForm.replaceReqBodyMap.map(p => ({ ...p })),
    replaceRespBodyScheme: itemForm.replaceRespBodyScheme,
    replaceRespBodyMap: itemForm.replaceRespBodyMap.map(p => ({ ...p })),
    replaceReqHeaderMap: itemForm.replaceReqHeaderMap.map(p => ({ ...p })),
    replaceRespHeaderMap: itemForm.replaceRespHeaderMap.map(p => ({ ...p })),
    respTextCode: itemForm.respTextCode,
    respTextContent: itemForm.respTextContent,
    respTextHeader: itemForm.respTextHeader.map(p => ({ ...p })),
    insertContent: itemForm.insertContent,
    insertPosition: itemForm.insertPosition,
    insertRegexp: itemForm.insertRegexp,
    webvpnLoginURL: itemForm.webvpnLoginURL,
    webvpnSites: itemForm.webvpnSites.map(s => ({
      ...s,
      replace: s.replace.map(p => ({ ...p }))
    })),
    authPortalTitle: itemForm.authPortalTitle,
    authPortalTokenName: itemForm.authPortalTokenName,
    authPortalTokenExpire: itemForm.authPortalTokenExpire,
    authPortalCookieDomain: itemForm.authPortalCookieDomain,
    authGuardPortalURL: itemForm.authGuardPortalURL
  };

  if (editingIndex.value !== null && editingIndex.value >= 0) {
    itemList.value[editingIndex.value] = target;
  } else {
    itemList.value.push(target);
  }

  editorVisible.value = false;
  syncToFormJSON();
}

// ─── List operations ──────────────────────────────────────────────────────────

function removeItem(idx: number) {
  itemList.value.splice(idx, 1);
  syncToFormJSON();
}

function moveItemUp(idx: number) {
  if (idx <= 0) return;
  [itemList.value[idx], itemList.value[idx - 1]] = [
    itemList.value[idx - 1],
    itemList.value[idx]
  ];
  syncToFormJSON();
}

function moveItemDown(idx: number) {
  if (idx >= itemList.value.length - 1) return;
  [itemList.value[idx], itemList.value[idx + 1]] = [
    itemList.value[idx + 1],
    itemList.value[idx]
  ];
  syncToFormJSON();
}

// ─── Summary helpers ──────────────────────────────────────────────────────────

function getMatcherSummary(item: RuleItemConfig) {
  const opt = matcherOptions.value.find(o => o.value === item.matcherName);
  const tagType = opt?.tagType ?? "info";
  const tag = opt?.tag ?? "";

  let summary = "";
  switch (item.matcherName) {
    case "always_true_matcher":
      summary = t("rule.matchAll");
      break;
    case "ip_matcher":
    case "cidr_matcher": {
      const lines = item.ipAddressText.split("\n").filter(s => s.trim());
      summary =
        lines.length > 0
          ? `${lines[0]}${lines.length > 1 ? ` ...(${lines.length})` : ""}`
          : t("rule.unconfiguredIp");
      break;
    }
    case "http_ip_matcher": {
      const lines = item.httpIpAddressText.split("\n").filter(s => s.trim());
      summary =
        lines.length > 0
          ? `${lines[0]}${lines.length > 1 ? ` ...(${lines.length})` : ""}`
          : t("rule.unconfiguredIp");
      break;
    }
    case "url_matcher":
      summary = `${item.urlMethod} ${item.urlPath}`;
      break;
    case "js_matcher":
      summary = t("rule.customJsScript");
      break;
    default:
      summary = item.matcherName;
  }
  return { tagType, tag, summary, name: item.matcherName };
}

function getActionSummary(item: RuleItemConfig) {
  const opt = actionOptions.value.find(o => o.value === item.actionName);
  const tagType = opt?.tagType ?? "info";
  const tag = opt?.tag ?? "";

  let summary = "";
  switch (item.actionName) {
    case "reset_conn_action":
      summary = item.resetContent
        ? `"${item.resetContent}"`
        : t("rule.tcpReset");
      break;
    case "hide_version_action":
      summary = t("rule.hideServerHeader");
      break;
    case "modify_status_action":
      summary = `HTTP ${item.modifyStatusCode}`;
      break;
    case "replace_request_body_action":
      summary = `(${item.replaceReqBodyMap.length})`;
      break;
    case "replace_response_body_action":
      summary = `(${item.replaceRespBodyMap.length})`;
      break;
    case "replace_request_header_action":
      summary = `(${item.replaceReqHeaderMap.length})`;
      break;
    case "replace_response_header_action":
      summary = `(${item.replaceRespHeaderMap.length})`;
      break;
    case "response_text_action":
      summary = `Code: ${item.respTextCode}`;
      break;
    case "insert_data_action":
      summary = `${item.insertPosition} ${item.insertRegexp}`;
      break;
    case "subdomain_webvpn_action":
      summary = `WebVPN (${item.webvpnSites.length})`;
      break;
    case "auth_portal_action":
      summary = item.authPortalTitle || "统一身份认证";
      break;
    case "auth_guard_action":
      summary = `Portal: ${item.authGuardPortalURL || "/login"}`;
      break;
    default:
      summary = item.actionName;
  }
  return { tagType, tag, summary, name: item.actionName };
}

// ─── KV Table Helpers ─────────────────────────────────────────────────────────

function addKV(list: KVPair[]) {
  list.push({ k: "", v: "" });
}
function removeKV(list: KVPair[], index: number) {
  list.splice(index, 1);
}

// ─── WebVPN site helpers ──────────────────────────────────────────────────────

function addWebvpnSite() {
  itemForm.webvpnSites.push({
    id: String(Math.random()),
    siteKey: "",
    protected: false,
    loginURL: "",
    hosts: "",
    replace: []
  });
}

function removeWebvpnSite(index: number) {
  itemForm.webvpnSites.splice(index, 1);
}

// ─── Form Validation ──────────────────────────────────────────────────────────

const rules = reactive({
  name: [
    {
      required: true,
      message: () => t("rule.valNameRequired"),
      trigger: "blur"
    }
  ]
});

function getRef() {
  return ruleFormRef.value;
}
defineExpose({ getRef });
</script>

<template>
  <div class="rule-form-wrap">
    <el-form
      ref="ruleFormRef"
      :label-position="deviceDetection() ? 'top' : 'right'"
      :model="newFormInline"
      :rules="rules"
      label-width="auto"
      class="p-1 sm:px-2"
    >
      <!-- ── Section 1: Basic Info ─────────────────────────────────────── -->
      <el-card shadow="never" class="mb-4 section-card">
        <template #header>
          <div class="card-header-inner">
            <div class="header-stripe" />
            <span class="header-title">{{ t("rule.basicAttr") }}</span>
          </div>
        </template>
        <el-row :gutter="20">
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

      <!-- ── Section 2: Item List ───────────────────────────────────────── -->
      <el-card shadow="never" class="mb-4 section-card">
        <template #header>
          <div class="flex-bc gap-2 flex-wrap">
            <div class="card-header-inner flex-wrap">
              <div class="header-stripe" />
              <span class="header-title">{{ t("rule.itemListTitle") }}</span>
              <el-tag
                size="small"
                type="primary"
                effect="plain"
                class="font-mono ml-1 sm:ml-2"
              >
                {{ t("rule.totalItems", { count: itemList.length }) }}
              </el-tag>
              <span
                class="text-xs text-(--el-text-color-placeholder) ml-1 hidden sm:inline"
              >
                {{ t("rule.dragTip") }}
              </span>
            </div>
            <el-button
              type="primary"
              size="small"
              :icon="useRenderIcon(AddFill)"
              @click="openAddItemEditor"
            >
              {{ t("rule.addItem") }}
            </el-button>
          </div>
        </template>

        <div v-if="itemList.length === 0" class="empty-state">
          <component
            :is="useRenderIcon(InfoIcon)"
            class="size-8 text-(--el-text-color-placeholder) mb-2"
          />
          <p class="text-sm text-(--el-text-color-placeholder)">
            {{ t("rule.noItems") }}
          </p>
        </div>

        <draggable
          v-else
          v-model="itemList"
          item-key="id"
          handle=".drag-handle"
          ghost-class="drag-ghost"
          class="space-y-2"
          @end="syncToFormJSON"
        >
          <template #item="{ element, index }">
            <div class="item-row">
              <!-- Left: drag + index -->
              <div class="item-row-left">
                <span class="drag-handle" :title="t('rule.dragTip')">
                  <component :is="useRenderIcon(DragIcon)" class="size-4" />
                </span>
                <el-tag
                  size="small"
                  type="info"
                  effect="plain"
                  class="font-mono font-bold w-7 sm:w-8 text-center shrink-0"
                >
                  {{ index + 1 }}
                </el-tag>
              </div>

              <!-- Middle: summary -->
              <div class="item-row-summary">
                <div class="summary-chip">
                  <el-tag
                    size="small"
                    :type="getMatcherSummary(element).tagType as any"
                    effect="light"
                    class="font-mono shrink-0"
                  >
                    {{ getMatcherSummary(element).tag }}
                  </el-tag>
                  <span class="summary-label">{{
                    getMatcherSummary(element).name
                  }}</span>
                  <span class="summary-value">{{
                    getMatcherSummary(element).summary
                  }}</span>
                </div>
                <div class="summary-arrow hidden sm:inline">→</div>
                <div class="summary-chip">
                  <el-tag
                    size="small"
                    :type="getActionSummary(element).tagType as any"
                    effect="light"
                    class="font-mono shrink-0"
                  >
                    {{ getActionSummary(element).tag }}
                  </el-tag>
                  <span class="summary-label">{{
                    getActionSummary(element).name
                  }}</span>
                  <span class="summary-value">{{
                    getActionSummary(element).summary
                  }}</span>
                </div>
              </div>

              <!-- Right: actions -->
              <div class="item-row-actions">
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
                  type="primary"
                  link
                  :icon="useRenderIcon(EditPen)"
                  @click="openEditItemEditor(index)"
                >
                  {{ t("rule.edit") }}
                </el-button>
                <el-popconfirm
                  :title="t('rule.confirmRemoveItem')"
                  @confirm="removeItem(index)"
                >
                  <template #reference>
                    <el-button
                      size="small"
                      type="danger"
                      link
                      :icon="useRenderIcon(Delete)"
                    >
                      {{ t("rule.delete") }}
                    </el-button>
                  </template>
                </el-popconfirm>
              </div>
            </div>
          </template>
        </draggable>
      </el-card>

      <!-- ── Section 3: JSON Preview ───────────────────────────────────── -->
      <div class="json-preview-box">
        <div class="flex-bc mb-1.5">
          <span
            class="text-xs font-semibold text-(--el-text-color-secondary)"
            >{{ t("rule.jsonPreview") }}</span
          >
          <el-tag size="small" type="info" effect="plain" class="font-mono"
            >Items Array</el-tag
          >
        </div>
        <el-scrollbar max-height="160px">
          <pre class="json-preview-code">{{ newFormInline.items }}</pre>
        </el-scrollbar>
      </div>
    </el-form>

    <!-- ═══════════════════════════════════════════════════════════════════
         Inline Editor Panel (right-side drawer overlay, 100% full width on mobile)
         ══════════════════════════════════════════════════════════════════ -->
    <transition name="editor-slide">
      <div v-if="editorVisible" class="editor-overlay">
        <div class="editor-backdrop" @click="closeEditor" />

        <div class="editor-panel">
          <!-- Panel Header -->
          <div class="editor-panel-header">
            <span class="editor-panel-title">{{ editorTitle }}</span>
            <div class="flex items-center gap-2">
              <el-button
                type="primary"
                size="small"
                :icon="useRenderIcon(CheckIcon)"
                @click="saveItemFromEditor"
              >
                {{ t("rule.saveItem") }}
              </el-button>
              <el-button
                size="small"
                :icon="useRenderIcon(CloseIcon)"
                @click="closeEditor"
              >
                {{ t("rule.cancel") }}
              </el-button>
            </div>
          </div>

          <!-- Panel Body -->
          <el-scrollbar class="editor-panel-body" always>
            <div class="editor-panel-content space-y-5">
              <!-- ── Matcher Section ────────────────────────────────────── -->
              <div class="config-section">
                <div class="config-section-header">
                  <span class="config-section-title">{{
                    t("rule.matcherConfigTitle")
                  }}</span>
                  <el-tag size="small" type="primary" effect="light"
                    >Matcher</el-tag
                  >
                </div>

                <div class="field-row">
                  <label class="field-label required">{{
                    t("rule.matcherType")
                  }}</label>
                  <el-select
                    v-model="itemForm.matcherName"
                    class="flex-1 w-full"
                  >
                    <el-option
                      v-for="m in matcherOptions"
                      :key="m.value"
                      :value="m.value"
                      :label="m.label"
                    >
                      <div class="flex items-center gap-2">
                        <el-tag
                          size="small"
                          :type="m.tagType as any"
                          effect="plain"
                          class="font-mono"
                          >{{ m.tag }}</el-tag
                        >
                        <span>{{ m.label }}</span>
                      </div>
                    </el-option>
                  </el-select>
                </div>

                <!-- always_true_matcher -->
                <el-alert
                  v-if="itemForm.matcherName === 'always_true_matcher'"
                  :title="t('rule.matcherAlwaysTrueTip')"
                  type="info"
                  :closable="false"
                  show-icon
                  class="mt-3"
                />

                <!-- ip_matcher / cidr_matcher -->
                <div
                  v-if="
                    itemForm.matcherName === 'ip_matcher' ||
                    itemForm.matcherName === 'cidr_matcher'
                  "
                  class="mt-3 space-y-2"
                >
                  <div class="field-row">
                    <label class="field-label required">
                      {{
                        itemForm.matcherName === "ip_matcher"
                          ? t("rule.ipAddressList")
                          : t("rule.cidrAddressList")
                      }}
                    </label>
                    <div class="flex-1 w-full">
                      <el-input
                        v-model="itemForm.ipAddressText"
                        type="textarea"
                        :rows="4"
                        :placeholder="
                          itemForm.matcherName === 'ip_matcher'
                            ? '127.0.0.1\n10.0.0.1-10.0.0.100'
                            : '192.168.1.0/24\n10.0.0.0/8'
                        "
                        class="font-mono text-sm"
                      />
                      <p class="field-hint">
                        {{
                          itemForm.matcherName === "ip_matcher"
                            ? t("rule.matcherIpTip")
                            : t("rule.matcherCidrTip")
                        }}
                      </p>
                    </div>
                  </div>
                </div>

                <!-- http_ip_matcher -->
                <div
                  v-if="itemForm.matcherName === 'http_ip_matcher'"
                  class="mt-3"
                >
                  <div class="field-row">
                    <label class="field-label required">{{
                      t("rule.httpIpAddressList")
                    }}</label>
                    <div class="flex-1 w-full">
                      <el-input
                        v-model="itemForm.httpIpAddressText"
                        type="textarea"
                        :rows="4"
                        placeholder="127.0.0.1&#10;192.168.1.0/24"
                        class="font-mono text-sm"
                      />
                      <p class="field-hint">{{ t("rule.matcherHttpIpTip") }}</p>
                    </div>
                  </div>
                </div>

                <!-- url_matcher -->
                <div
                  v-if="itemForm.matcherName === 'url_matcher'"
                  class="mt-3 space-y-3"
                >
                  <div class="field-row">
                    <label class="field-label required">{{
                      t("rule.urlMethod")
                    }}</label>
                    <el-select
                      v-model="itemForm.urlMethod"
                      class="flex-1 w-full"
                    >
                      <el-option
                        v-for="m in HTTP_METHODS"
                        :key="m"
                        :label="m"
                        :value="m"
                      />
                    </el-select>
                  </div>
                  <div class="field-row">
                    <label class="field-label required">{{
                      t("rule.urlPath")
                    }}</label>
                    <div class="flex-1 w-full">
                      <el-input
                        v-model="itemForm.urlPath"
                        placeholder="/api/user"
                      />
                      <p class="field-hint">{{ t("rule.matcherUrlTip") }}</p>
                    </div>
                  </div>
                </div>

                <!-- js_matcher -->
                <div v-if="itemForm.matcherName === 'js_matcher'" class="mt-3">
                  <div class="field-row">
                    <label class="field-label required">{{
                      t("rule.jsScript")
                    }}</label>
                    <div class="flex-1 w-full">
                      <el-input
                        v-model="itemForm.jsScript"
                        type="textarea"
                        :rows="6"
                        placeholder="return true;"
                        class="font-mono text-sm"
                      />
                      <p class="field-hint">{{ t("rule.matcherJsTip") }}</p>
                    </div>
                  </div>
                </div>
              </div>

              <!-- ── Action Section ─────────────────────────────────────── -->
              <div class="config-section">
                <div class="config-section-header">
                  <span class="config-section-title">{{
                    t("rule.actionConfigTitle")
                  }}</span>
                  <el-tag size="small" type="success" effect="light"
                    >Action</el-tag
                  >
                </div>

                <div class="field-row">
                  <label class="field-label required">{{
                    t("rule.actionType")
                  }}</label>
                  <el-select
                    v-model="itemForm.actionName"
                    class="flex-1 w-full"
                  >
                    <el-option
                      v-for="a in actionOptions"
                      :key="a.value"
                      :value="a.value"
                      :label="a.label"
                    >
                      <div class="flex items-center gap-2">
                        <el-tag
                          size="small"
                          :type="a.tagType as any"
                          effect="plain"
                          class="font-mono"
                          >{{ a.tag }}</el-tag
                        >
                        <span>{{ a.label }}</span>
                      </div>
                    </el-option>
                  </el-select>
                </div>

                <!-- reset_conn_action -->
                <div
                  v-if="itemForm.actionName === 'reset_conn_action'"
                  class="action-config-body"
                >
                  <div class="field-row">
                    <label class="field-label">{{
                      t("rule.resetContent")
                    }}</label>
                    <div class="flex-1 w-full">
                      <el-input
                        v-model="itemForm.resetContent"
                        placeholder="Connection reset by rule"
                      />
                      <p class="field-hint">{{ t("rule.resetContentTip") }}</p>
                    </div>
                  </div>
                </div>

                <!-- hide_version_action -->
                <div
                  v-if="itemForm.actionName === 'hide_version_action'"
                  class="action-config-body"
                >
                  <el-alert
                    :title="t('rule.actionHideVersionAlert')"
                    type="info"
                    :closable="false"
                    show-icon
                  />
                </div>

                <!-- modify_status_action -->
                <div
                  v-if="itemForm.actionName === 'modify_status_action'"
                  class="action-config-body"
                >
                  <div class="field-row">
                    <label class="field-label required"
                      >{{ t("rule.modifyStatusCode") }} <code>Code</code></label
                    >
                    <div class="flex-1 w-full">
                      <el-input-number
                        v-model="itemForm.modifyStatusCode"
                        :min="100"
                        :max="599"
                        class="w-full"
                        controls-position="right"
                      />
                      <p class="field-hint">{{ t("rule.modifyStatusTip") }}</p>
                    </div>
                  </div>
                </div>

                <!-- replace_request_body_action -->
                <div
                  v-if="itemForm.actionName === 'replace_request_body_action'"
                  class="action-config-body space-y-4"
                >
                  <div class="field-row">
                    <label class="field-label"
                      >{{ t("rule.scheme") }} <code>Scheme</code></label
                    >
                    <el-select
                      v-model="itemForm.replaceReqBodyScheme"
                      class="flex-1 w-full"
                    >
                      <el-option
                        v-for="s in BODY_SCHEMES"
                        :key="s.value"
                        :label="s.label"
                        :value="s.value"
                      />
                    </el-select>
                  </div>
                  <div>
                    <div class="kv-table-header">
                      <span class="kv-col-label">{{ t("rule.regexKey") }}</span>
                      <span class="kv-col-label">{{
                        t("rule.replaceValue")
                      }}</span>
                      <span class="w-8" />
                    </div>
                    <div
                      v-for="(pair, i) in itemForm.replaceReqBodyMap"
                      :key="i"
                      class="kv-row"
                    >
                      <el-input
                        v-model="pair.k"
                        placeholder="foo|bar"
                        class="font-mono"
                        size="small"
                      />
                      <el-input v-model="pair.v" placeholder="" size="small" />
                      <el-button
                        type="danger"
                        link
                        size="small"
                        :icon="useRenderIcon(Delete)"
                        @click="removeKV(itemForm.replaceReqBodyMap, i)"
                      />
                    </div>
                    <el-button
                      type="primary"
                      plain
                      size="small"
                      :icon="useRenderIcon(PlusIcon)"
                      class="mt-2"
                      @click="addKV(itemForm.replaceReqBodyMap)"
                    >
                      {{ t("rule.addReplaceRule") }}
                    </el-button>
                    <p class="field-hint mt-1">
                      {{ t("rule.replaceReqBodyHint") }}
                    </p>
                  </div>
                </div>

                <!-- replace_response_body_action -->
                <div
                  v-if="itemForm.actionName === 'replace_response_body_action'"
                  class="action-config-body space-y-4"
                >
                  <div class="field-row">
                    <label class="field-label"
                      >{{ t("rule.scheme") }} <code>Scheme</code></label
                    >
                    <el-select
                      v-model="itemForm.replaceRespBodyScheme"
                      class="flex-1 w-full"
                    >
                      <el-option
                        v-for="s in BODY_SCHEMES"
                        :key="s.value"
                        :label="s.label"
                        :value="s.value"
                      />
                    </el-select>
                  </div>
                  <div>
                    <div class="kv-table-header">
                      <span class="kv-col-label">{{ t("rule.regexKey") }}</span>
                      <span class="kv-col-label">{{
                        t("rule.replaceValue")
                      }}</span>
                      <span class="w-8" />
                    </div>
                    <div
                      v-for="(pair, i) in itemForm.replaceRespBodyMap"
                      :key="i"
                      class="kv-row"
                    >
                      <el-input
                        v-model="pair.k"
                        placeholder="foo|bar"
                        class="font-mono"
                        size="small"
                      />
                      <el-input v-model="pair.v" placeholder="" size="small" />
                      <el-button
                        type="danger"
                        link
                        size="small"
                        :icon="useRenderIcon(Delete)"
                        @click="removeKV(itemForm.replaceRespBodyMap, i)"
                      />
                    </div>
                    <el-button
                      type="primary"
                      plain
                      size="small"
                      :icon="useRenderIcon(PlusIcon)"
                      class="mt-2"
                      @click="addKV(itemForm.replaceRespBodyMap)"
                    >
                      {{ t("rule.addReplaceRule") }}
                    </el-button>
                    <p class="field-hint mt-1">
                      {{ t("rule.replaceRespBodyHint") }}
                    </p>
                  </div>
                </div>

                <!-- replace_request_header_action / replace_response_header_action -->
                <div
                  v-if="
                    itemForm.actionName === 'replace_request_header_action' ||
                    itemForm.actionName === 'replace_response_header_action'
                  "
                  class="action-config-body space-y-3"
                >
                  <div class="kv-table-header">
                    <span class="kv-col-label">{{
                      t("rule.headerRegexKey")
                    }}</span>
                    <span class="kv-col-label">{{
                      t("rule.headerReplaceValue")
                    }}</span>
                    <span class="w-8" />
                  </div>
                  <template
                    v-if="
                      itemForm.actionName === 'replace_request_header_action'
                    "
                  >
                    <div
                      v-for="(pair, i) in itemForm.replaceReqHeaderMap"
                      :key="i"
                      class="kv-row"
                    >
                      <el-input
                        v-model="pair.k"
                        placeholder="Bearer.*"
                        class="font-mono"
                        size="small"
                      />
                      <el-input v-model="pair.v" placeholder="" size="small" />
                      <el-button
                        type="danger"
                        link
                        size="small"
                        :icon="useRenderIcon(Delete)"
                        @click="removeKV(itemForm.replaceReqHeaderMap, i)"
                      />
                    </div>
                    <el-button
                      type="primary"
                      plain
                      size="small"
                      :icon="useRenderIcon(PlusIcon)"
                      @click="addKV(itemForm.replaceReqHeaderMap)"
                    >
                      {{ t("rule.addReplaceRule") }}
                    </el-button>
                  </template>
                  <template v-else>
                    <div
                      v-for="(pair, i) in itemForm.replaceRespHeaderMap"
                      :key="i"
                      class="kv-row"
                    >
                      <el-input
                        v-model="pair.k"
                        placeholder="Server.*"
                        class="font-mono"
                        size="small"
                      />
                      <el-input v-model="pair.v" placeholder="" size="small" />
                      <el-button
                        type="danger"
                        link
                        size="small"
                        :icon="useRenderIcon(Delete)"
                        @click="removeKV(itemForm.replaceRespHeaderMap, i)"
                      />
                    </div>
                    <el-button
                      type="primary"
                      plain
                      size="small"
                      :icon="useRenderIcon(PlusIcon)"
                      @click="addKV(itemForm.replaceRespHeaderMap)"
                    >
                      {{ t("rule.addReplaceRule") }}
                    </el-button>
                  </template>
                  <p class="field-hint">{{ t("rule.headerReplaceHint") }}</p>
                </div>

                <!-- response_text_action -->
                <div
                  v-if="itemForm.actionName === 'response_text_action'"
                  class="action-config-body space-y-4"
                >
                  <div class="field-row">
                    <label class="field-label required"
                      >{{ t("rule.respTextCode") }} <code>Code</code></label
                    >
                    <el-select
                      v-model="itemForm.respTextCode"
                      filterable
                      allow-create
                      default-first-option
                      class="flex-1 w-full"
                    >
                      <el-option
                        v-for="s in HTTP_STATUS_CODES"
                        :key="s.value"
                        :label="s.label"
                        :value="s.value"
                      />
                    </el-select>
                  </div>
                  <div class="field-row">
                    <label class="field-label"
                      >{{ t("rule.respTextHeader") }} <code>Header</code></label
                    >
                    <div class="flex-1 w-full space-y-2">
                      <div class="kv-table-header">
                        <span class="kv-col-label">Content-Type</span>
                        <span class="kv-col-label">Value</span>
                        <span class="w-8" />
                      </div>
                      <div
                        v-for="(h, i) in itemForm.respTextHeader"
                        :key="i"
                        class="kv-row"
                      >
                        <el-input
                          v-model="h.k"
                          placeholder="Content-Type"
                          size="small"
                        />
                        <el-input
                          v-model="h.v"
                          placeholder="text/html; charset=utf-8"
                          size="small"
                        />
                        <el-button
                          type="danger"
                          link
                          size="small"
                          :icon="useRenderIcon(Delete)"
                          @click="removeKV(itemForm.respTextHeader, i)"
                        />
                      </div>
                      <el-button
                        type="primary"
                        plain
                        size="small"
                        :icon="useRenderIcon(PlusIcon)"
                        @click="addKV(itemForm.respTextHeader)"
                      >
                        {{ t("rule.addResponseHeader") }}
                      </el-button>
                    </div>
                  </div>
                  <div class="field-row">
                    <label class="field-label"
                      >{{ t("rule.respTextContent") }}
                      <code>Content</code></label
                    >
                    <el-input
                      v-model="itemForm.respTextContent"
                      type="textarea"
                      :rows="4"
                      :placeholder="t('rule.respTextContentPlaceholder')"
                      class="flex-1 w-full"
                    />
                  </div>
                </div>

                <!-- insert_data_action -->
                <div
                  v-if="itemForm.actionName === 'insert_data_action'"
                  class="action-config-body space-y-4"
                >
                  <div class="field-row">
                    <label class="field-label required"
                      >{{ t("rule.insertRegexp") }} <code>Regexp</code></label
                    >
                    <div class="flex-1 w-full">
                      <el-select
                        v-model="itemForm.insertRegexp"
                        filterable
                        allow-create
                        default-first-option
                        :placeholder="t('rule.insertRegexpPlaceholder')"
                        class="w-full font-mono"
                      >
                        <el-option
                          v-for="r in INSERT_REGEXP_OPTIONS"
                          :key="r.value"
                          :label="r.label"
                          :value="r.value"
                        />
                      </el-select>
                      <p class="field-hint">{{ t("rule.insertRegexpHint") }}</p>
                    </div>
                  </div>
                  <div class="field-row">
                    <label class="field-label required"
                      >{{ t("rule.insertPosition") }}
                      <code>Position</code></label
                    >
                    <el-select
                      v-model="itemForm.insertPosition"
                      class="flex-1 w-full"
                    >
                      <el-option
                        v-for="p in INSERT_POSITIONS"
                        :key="p.value"
                        :label="p.label"
                        :value="p.value"
                      />
                    </el-select>
                  </div>
                  <div class="field-row">
                    <label class="field-label required"
                      >{{ t("rule.insertContent") }} <code>Content</code></label
                    >
                    <el-input
                      v-model="itemForm.insertContent"
                      type="textarea"
                      :rows="4"
                      :placeholder="t('rule.insertContentPlaceholder')"
                      class="flex-1 w-full"
                    />
                  </div>
                </div>

                <!-- subdomain_webvpn_action -->
                <div
                  v-if="itemForm.actionName === 'subdomain_webvpn_action'"
                  class="action-config-body space-y-4"
                >
                  <div class="field-row">
                    <label class="field-label"
                      >{{ t("rule.webvpnLoginUrl") }}
                      <code>LoginURL</code></label
                    >
                    <div class="flex-1 w-full">
                      <el-input
                        v-model="itemForm.webvpnLoginURL"
                        placeholder="/login"
                      />
                      <p class="field-hint">
                        {{ t("rule.webvpnLoginUrlHint") }}
                      </p>
                    </div>
                  </div>

                  <div>
                    <div class="flex-bc mb-3">
                      <span
                        class="text-sm font-semibold text-(--el-text-color-primary)"
                      >
                        {{ t("rule.webvpnSitesTitle") }}
                        <code class="text-xs font-normal">Sites</code>
                      </span>
                      <el-button
                        type="primary"
                        plain
                        size="small"
                        :icon="useRenderIcon(PlusIcon)"
                        @click="addWebvpnSite"
                      >
                        {{ t("rule.addWebvpnSite") }}
                      </el-button>
                    </div>

                    <div
                      v-if="itemForm.webvpnSites.length === 0"
                      class="text-center py-4 text-xs text-(--el-text-color-placeholder)"
                    >
                      {{ t("rule.noWebvpnSites") }}
                    </div>

                    <div
                      v-for="(site, si) in itemForm.webvpnSites"
                      :key="site.id"
                      class="webvpn-site-card"
                    >
                      <div class="webvpn-site-header">
                        <span class="text-sm font-semibold"
                          >Site #{{ si + 1 }}</span
                        >
                        <el-button
                          type="danger"
                          plain
                          size="small"
                          :icon="useRenderIcon(Delete)"
                          @click="removeWebvpnSite(si)"
                        >
                          {{ t("rule.removeSite") }}
                        </el-button>
                      </div>

                      <div class="space-y-3 mt-3">
                        <div class="field-row">
                          <label class="field-label required">{{
                            t("rule.siteKey")
                          }}</label>
                          <el-input
                            v-model="site.siteKey"
                            :placeholder="t('rule.siteKeyPlaceholder')"
                            class="flex-1 font-mono"
                            size="small"
                          />
                        </div>
                        <div class="field-row">
                          <label class="field-label">{{
                            t("rule.siteProtected")
                          }}</label>
                          <el-switch
                            v-model="site.protected"
                            active-text="Yes"
                            inactive-text="No"
                          />
                        </div>
                        <div class="field-row">
                          <label class="field-label"
                            >{{ t("rule.loginUrl") }}
                            <code>loginURL</code></label
                          >
                          <el-input
                            v-model="site.loginURL"
                            placeholder="/login"
                            class="flex-1"
                            size="small"
                          />
                        </div>

                        <!-- Hosts -->
                        <div class="webvpn-sub-section">
                          <div class="mb-2">
                            <span
                              class="text-xs font-semibold text-(--el-text-color-regular)"
                            >
                              {{ t("rule.siteHostsTitle") }}
                              <code class="font-normal">host / wildcard</code>
                            </span>
                          </div>
                          <el-input
                            v-model="site.hosts"
                            type="textarea"
                            :rows="4"
                            :placeholder="t('rule.siteHostsPlaceholder')"
                            class="font-mono text-xs"
                          />
                          <p
                            class="text-xs/relaxed text-(--el-text-color-placeholder) mt-1.5"
                          >
                            {{ t("rule.siteHostsHint") }}
                          </p>
                        </div>

                        <!-- Replace mapping -->
                        <div class="webvpn-sub-section">
                          <div class="flex-bc mb-2">
                            <span
                              class="text-xs font-semibold text-(--el-text-color-regular)"
                              >{{ t("rule.siteReplaceTitle") }}
                              <code class="font-normal">replace</code></span
                            >
                            <el-button
                              type="primary"
                              link
                              size="small"
                              :icon="useRenderIcon(PlusIcon)"
                              @click="addKV(site.replace)"
                              >{{ t("rule.edit") }}</el-button
                            >
                          </div>
                          <div class="kv-table-header">
                            <span class="kv-col-label">Regex</span>
                            <span class="kv-col-label">Value</span>
                            <span class="w-8" />
                          </div>
                          <div
                            v-for="(pair, pi) in site.replace"
                            :key="pi"
                            class="kv-row"
                          >
                            <el-input
                              v-model="pair.k"
                              placeholder="Regex"
                              class="font-mono"
                              size="small"
                            />
                            <el-input
                              v-model="pair.v"
                              placeholder="Value"
                              size="small"
                            />
                            <el-button
                              type="danger"
                              link
                              size="small"
                              :icon="useRenderIcon(Delete)"
                              @click="removeKV(site.replace, pi)"
                            />
                          </div>
                          <p
                            v-if="site.replace.length === 0"
                            class="text-xs text-(--el-text-color-placeholder) text-center py-2"
                          >
                            {{ t("rule.noReplaceRules") }}
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- auth_portal_action -->
                <div
                  v-if="itemForm.actionName === 'auth_portal_action'"
                  class="action-config-body space-y-4"
                >
                  <el-alert
                    :title="t('rule.actionAuthPortalAlert')"
                    type="success"
                    :closable="false"
                    show-icon
                    class="mb-3"
                  />
                  <div class="field-row">
                    <label class="field-label">{{ t("rule.portalTitle") }}</label>
                    <div class="flex-1 w-full">
                      <el-input
                        v-model="itemForm.authPortalTitle"
                        placeholder="统一身份认证"
                      />
                    </div>
                  </div>
                  <div class="field-row">
                    <label class="field-label">{{ t("rule.portalTokenExpire") }}</label>
                    <div class="flex-1 w-full">
                      <el-input-number
                        v-model="itemForm.authPortalTokenExpire"
                        :min="60"
                        :step="3600"
                        class="w-full"
                      />
                    </div>
                  </div>
                  <div class="field-row">
                    <label class="field-label">{{ t("rule.portalCookieDomain") }}</label>
                    <div class="flex-1 w-full">
                      <el-input
                        v-model="itemForm.authPortalCookieDomain"
                        placeholder="例如 .example.com (留空为当前域名)"
                      />
                      <p class="field-hint">{{ t("rule.portalCookieDomainTip") }}</p>
                    </div>
                  </div>
                </div>

                <!-- auth_guard_action -->
                <div
                  v-if="itemForm.actionName === 'auth_guard_action'"
                  class="action-config-body space-y-4"
                >
                  <el-alert
                    :title="t('rule.actionAuthGuardAlert')"
                    type="info"
                    :closable="false"
                    show-icon
                    class="mb-3"
                  />
                  <div class="field-row">
                    <label class="field-label required">{{ t("rule.guardPortalUrl") }}</label>
                    <div class="flex-1 w-full">
                      <el-input
                        v-model="itemForm.authGuardPortalURL"
                        :placeholder="t('rule.guardPortalUrlPlaceholder')"
                      />
                      <p class="field-hint">{{ t("rule.guardPortalUrlTip") }}</p>
                    </div>
                  </div>
                </div>
              </div>
              <!-- ── End Action Section ──────────────────────────────────── -->
            </div>
          </el-scrollbar>

          <!-- Panel Footer -->
          <div class="editor-panel-footer">
            <el-button :icon="useRenderIcon(CloseIcon)" @click="closeEditor">{{
              t("rule.cancel")
            }}</el-button>
            <el-button
              type="primary"
              :icon="useRenderIcon(CheckIcon)"
              @click="saveItemFromEditor"
            >
              {{ t("rule.saveItem") }}
            </el-button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped>
/* ── Root wrapper ─────────────────────────────────────────────────── */
.rule-form-wrap {
  position: relative;
}

/* ── Section cards ────────────────────────────────────────────────── */
.section-card {
  border: 1px solid var(--el-border-color-lighter) !important;
  border-radius: 12px !important;
}

.card-header-inner {
  display: flex;
  align-items: center;
  gap: 8px;
}
.header-stripe {
  width: 4px;
  height: 16px;
  background: var(--el-color-primary);
  border-radius: 2px;
  flex-shrink: 0;
}
.header-title {
  font-weight: 700;
  font-size: 14px;
  color: var(--el-text-color-primary);
}

/* ── Empty state ──────────────────────────────────────────────────── */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 28px 0;
}

/* ── Item row ─────────────────────────────────────────────────────── */
.item-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
  transition: border-color 0.18s;
  flex-wrap: wrap;
}
.item-row:hover {
  border-color: var(--el-color-primary-light-5);
}

.item-row-left {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.drag-handle {
  cursor: grab;
  color: var(--el-text-color-placeholder);
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  transition: color 0.15s;
}
.drag-handle:hover {
  color: var(--el-color-primary);
}
.drag-handle:active {
  cursor: grabbing;
}

.item-row-summary {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  flex-wrap: wrap;
}

.summary-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  overflow: hidden;
}

.summary-label {
  font-size: 11px;
  font-family: var(--el-font-family-mono, monospace);
  color: var(--el-text-color-secondary);
  white-space: nowrap;
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
}
.summary-value {
  font-size: 12px;
  color: var(--el-text-color-regular);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 180px;
}
.summary-arrow {
  color: var(--el-text-color-placeholder);
  font-size: 14px;
  flex-shrink: 0;
}

.item-row-actions {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
}

/* drag ghost style */
:deep(.drag-ghost) {
  opacity: 0.4;
  background: var(--el-color-primary-light-9) !important;
}

/* ── JSON Preview ─────────────────────────────────────────────────── */
.json-preview-box {
  border-radius: 10px;
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-light);
  padding: 10px 12px;
}
.json-preview-code {
  font-size: 11px;
  font-family: var(--el-font-family-mono, monospace);
  color: var(--el-text-color-primary);
  white-space: pre-wrap;
  word-break: break-all;
  line-height: 1.6;
  margin: 0;
}

/* ── Editor Overlay ───────────────────────────────────────────────── */
.editor-overlay {
  position: fixed;
  inset: 0;
  z-index: 1001;
  display: flex;
  justify-content: flex-end;
}

.editor-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
}

.editor-panel {
  position: relative;
  width: min(580px, 92vw);
  height: 100%;
  background: var(--el-bg-color);
  display: flex;
  flex-direction: column;
  box-shadow: -4px 0 24px rgba(0, 0, 0, 0.15);
}

.editor-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}
.editor-panel-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}

.editor-panel-body {
  flex: 1;
  overflow: hidden;
}
.editor-panel-content {
  padding: 20px;
}

.editor-panel-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 20px;
  border-top: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

/* ── Config sections inside editor ──────────────────────────────────── */
.config-section {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  padding: 16px;
  background: var(--el-fill-color-blank);
}

.config-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.config-section-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}

.action-config-body {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px dashed var(--el-border-color-lighter);
}

/* ── Field rows ────────────────────────────────────────────────────── */
.field-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.field-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-regular);
  white-space: nowrap;
  padding-top: 6px;
  min-width: 140px;
  flex-shrink: 0;
}
.field-label.required::before {
  content: "* ";
  color: var(--el-color-danger);
}

.field-hint {
  margin: 4px 0 0;
  font-size: 11px;
  color: var(--el-text-color-placeholder);
  line-height: 1.5;
}
.field-hint code {
  background: var(--el-fill-color);
  padding: 0 3px;
  border-radius: 3px;
  font-family: var(--el-font-family-mono, monospace);
}

/* ── KV table ──────────────────────────────────────────────────────── */
.kv-table-header {
  display: grid;
  grid-template-columns: 1fr 1fr 2rem;
  gap: 6px;
  padding: 0 0 4px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  margin-bottom: 6px;
}
.kv-col-label {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  font-weight: 500;
}
.kv-row {
  display: grid;
  grid-template-columns: 1fr 1fr 2rem;
  gap: 6px;
  align-items: center;
  margin-bottom: 6px;
}

/* ── WebVPN Site Card ──────────────────────────────────────────────── */
.webvpn-site-card {
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  padding: 14px;
  background: var(--el-fill-color-light);
  margin-bottom: 12px;
}
.webvpn-site-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.webvpn-sub-section {
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  padding: 10px;
}

/* ── Mobile Responsiveness ─────────────────────────────────────────── */
@media (max-width: 640px) {
  .editor-panel {
    width: 100vw !important;
    max-width: 100vw !important;
  }
  .editor-panel-content {
    padding: 12px;
  }
  .field-row {
    flex-direction: column;
    align-items: stretch;
    gap: 4px;
  }
  .field-label {
    min-width: auto;
    padding-top: 0;
    margin-bottom: 2px;
  }
  .item-row {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
  .item-row-left {
    justify-content: space-between;
  }
  .item-row-actions {
    justify-content: flex-end;
    border-top: 1px dashed var(--el-border-color-lighter);
    padding-top: 6px;
  }
}

/* ── Slide transition ──────────────────────────────────────────────── */
.editor-slide-enter-active,
.editor-slide-leave-active {
  transition: opacity 0.22s ease;
}
.editor-slide-enter-active .editor-panel,
.editor-slide-leave-active .editor-panel {
  transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}
.editor-slide-enter-from,
.editor-slide-leave-to {
  opacity: 0;
}
.editor-slide-enter-from .editor-panel,
.editor-slide-leave-to .editor-panel {
  transform: translateX(100%);
}

:deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--el-text-color-regular);
}
</style>
