<script setup lang="ts">
import { deviceDetection } from "@pureadmin/utils";
import { ref, reactive, computed, onMounted } from "vue";
import ReCol from "@/components/ReCol";
import { useI18n } from "vue-i18n";
import { getCertList } from "@/api/certificate";
import { getRuleList } from "@/api/rule";
import { getTunnelList } from "@/api/tunnel";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import AddFill from "~icons/ri/add-line";
import Delete from "~icons/ep/delete";
import Rank from "~icons/ep/rank";
import WarningIcon from "~icons/ep/warning";
import Sortable from "sortablejs";

export interface LocationItem {
  Path: string;
  Upstream: {
    Type: string;
    Data: {
      Method?: string;
      Servers?: Array<{ Target: string; Weight: number }>;
      Dir?: string;
    };
  };
}

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

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "",
    id: undefined,
    name: "",
    port: 80,
    hostname: "",
    http: true,
    tls: true,
    h2: true,
    hsts: false,
    certificate: "",
    proxy_headers: JSON.stringify([]),
    compress: false,
    rules: JSON.stringify([]),
    real_ip: "",
    tunnel_type: "quic",
    tunnel_id: "",
    tunnel_token: "",
    dns_resolver: "dns1",
    location_json: JSON.stringify(
      [
        {
          Path: "/",
          Upstream: {
            Type: "proxy_pass",
            Data: {
              Method: "round_robin",
              Servers: [{ Target: "http://127.0.0.1:8080", Weight: 1 }]
            }
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

const httpFormRef = ref();
const newFormInline = ref(props.formInline);

// Common Forward / Proxy Header options
const commonHeaderOptions = [
  { label: "Host", value: "Host" },
  { label: "X-Forwarded-For", value: "X-Forwarded-For" },
  { label: "X-Forwarded-Proto", value: "X-Forwarded-Proto" },
  { label: "X-Forwarded-Host", value: "X-Forwarded-Host" },
  { label: "X-Real-IP", value: "X-Real-IP" },
  { label: "Connection", value: "Connection" },
  { label: "Upgrade", value: "Upgrade" },
  { label: "Keep-Alive", value: "Keep-Alive" },
  { label: "Proxy-Authenticate", value: "Proxy-Authenticate" },
  { label: "Proxy-Authorization", value: "Proxy-Authorization" },
  { label: "TE", value: "TE" },
  { label: "Trailers", value: "Trailers" },
  { label: "Transfer-Encoding", value: "Transfer-Encoding" }
];
const selectedProxyHeaders = ref<string[]>([]);

// Certificate options
const certOptions = ref<Array<{ label: string; value: string }>>([]);
// Tunnel options
const tunnelOptions = ref<
  Array<{ label: string; value: string; type: string }>
>([]);
const tunnelNodeGroups = ref<TunnelGroup[]>([]);
const selectedTunnelNodeKey = ref<string>("");
// Rule options
const availableRules = ref<Array<{ label: string; value: string }>>([]);
// Selected Rules array
const selectedRules = ref<string[]>([]);

// Location items array
const locationList = ref<LocationItem[]>([]);

interface DnsItem {
  id: string;
  value: string;
}
const dnsResolverList = ref<DnsItem[]>([]);
const dnsResolverGroupRef = ref();

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

function addDnsResolver() {
  dnsResolverList.value.push({
    id: Math.random().toString(36).substring(2, 9),
    value: ""
  });
  syncDnsResolver();
}

function removeDnsResolver(idx: number) {
  dnsResolverList.value.splice(idx, 1);
  syncDnsResolver();
}

function syncDnsResolver() {
  newFormInline.value.dns_resolver = JSON.stringify(
    dnsResolverList.value.map(item => item.value.trim()).filter(v => v !== "")
  );
}

onMounted(() => {
  fetchCertificates();
  fetchTunnels();
  fetchRules();
  initFormState();
  initSortableDns();
});

function initFormState() {
  newFormInline.value.port = Number(newFormInline.value.port) || 80;
  try {
    if (newFormInline.value.proxy_headers) {
      const parsed =
        typeof newFormInline.value.proxy_headers === "string"
          ? JSON.parse(newFormInline.value.proxy_headers)
          : newFormInline.value.proxy_headers;
      if (Array.isArray(parsed)) selectedProxyHeaders.value = parsed;
    }
  } catch (e) {}
  try {
    if (newFormInline.value.rules) {
      const parsed =
        typeof newFormInline.value.rules === "string"
          ? JSON.parse(newFormInline.value.rules)
          : newFormInline.value.rules;
      if (Array.isArray(parsed)) selectedRules.value = parsed;
    }
  } catch (e) {}

  try {
    if (newFormInline.value.dns_resolver) {
      let parsed = newFormInline.value.dns_resolver;
      if (typeof parsed === "string") {
        if (parsed.startsWith("[")) {
          parsed = JSON.parse(parsed);
        } else {
          parsed = parsed ? [parsed] : [];
        }
      }
      if (Array.isArray(parsed)) {
        dnsResolverList.value = parsed.map((v: string) => ({
          id: Math.random().toString(36).substring(2, 9),
          value: v
        }));
      }
    }
  } catch (e) {
    if (newFormInline.value.dns_resolver) {
      dnsResolverList.value = [
        {
          id: Math.random().toString(36).substring(2, 9),
          value: newFormInline.value.dns_resolver
        }
      ];
    }
  }
  if (dnsResolverList.value.length === 0) {
    dnsResolverList.value.push({
      id: Math.random().toString(36).substring(2, 9),
      value: ""
    });
  }

  try {
    if (newFormInline.value.location_json) {
      const parsed =
        typeof newFormInline.value.location_json === "string"
          ? JSON.parse(newFormInline.value.location_json)
          : newFormInline.value.location_json;
      if (Array.isArray(parsed)) {
        locationList.value = parsed.map((loc: any) => {
          const type = loc?.Upstream?.Type || "proxy_pass";
          if (type === "root" || type === "alias") {
            return {
              Path: loc.Path || "/",
              Upstream: {
                Type: type,
                Data: {
                  Dir: loc?.Upstream?.Data?.Dir || "./static"
                }
              }
            };
          }
          return {
            Path: loc.Path || "/",
            Upstream: {
              Type: "proxy_pass",
              Data: {
                Method: loc?.Upstream?.Data?.Method || "round_robin",
                Servers:
                  Array.isArray(loc?.Upstream?.Data?.Servers) &&
                  loc.Upstream.Data.Servers.length > 0
                    ? loc.Upstream.Data.Servers
                    : [{ Target: "http://127.0.0.1:8080", Weight: 1 }]
              }
            }
          };
        });
      }
    }
  } catch (e) {}

  if (locationList.value.length === 0) {
    locationList.value.push({
      Path: "/",
      Upstream: {
        Type: "proxy_pass",
        Data: {
          Method: "round_robin",
          Servers: [{ Target: "http://127.0.0.1:8080", Weight: 1 }]
        }
      }
    });
  }
  syncLocationJSON();
}

const certListRaw = ref<any[]>([]);

function matchDomain(hostname: string, certPattern: string): boolean {
  if (!hostname || !certPattern) return false;
  const host = hostname.trim().toLowerCase();
  const pattern = certPattern.trim().toLowerCase();

  if (host === pattern) return true;

  // Wildcard match (RFC 6125): *.example.com matches sub.example.com
  if (pattern.startsWith("*.")) {
    const baseDomain = pattern.slice(2);
    if (host === baseDomain) return false; // *.example.com does not match example.com itself
    if (host.endsWith("." + baseDomain)) {
      const prefix = host.slice(0, host.length - baseDomain.length - 1);
      // Wildcard matches exactly one level of subdomain (prefix contains no dot)
      if (prefix && !prefix.includes(".")) {
        return true;
      }
    }
  }

  return false;
}

const certMismatchWarning = computed(() => {
  if (
    !newFormInline.value.tls ||
    !newFormInline.value.certificate ||
    !newFormInline.value.hostname
  ) {
    return null;
  }

  const selectedCertId = newFormInline.value.certificate;
  const cert = certListRaw.value.find(
    (c: any) =>
      (c.CertId || c.cert_id) === selectedCertId ||
      String(c.Id || c.id) === selectedCertId ||
      `id-${c.Id || c.id}` === selectedCertId
  );

  if (!cert) return null;

  const coveredDomains: string[] = [];
  const cn = cert.SubjectCN || cert.subject_cn;
  if (cn) coveredDomains.push(cn);

  const rawSans = cert.SANs || cert.sans;
  if (Array.isArray(rawSans)) {
    coveredDomains.push(...rawSans);
  } else if (typeof rawSans === "string" && rawSans.trim()) {
    try {
      if (rawSans.startsWith("[")) {
        const parsed = JSON.parse(rawSans);
        if (Array.isArray(parsed)) coveredDomains.push(...parsed);
      } else {
        const parts = rawSans
          .split(/[\n,;]+/)
          .map(s => s.trim())
          .filter(Boolean);
        coveredDomains.push(...parts);
      }
    } catch {
      const parts = rawSans
        .split(/[\n,;]+/)
        .map(s => s.trim())
        .filter(Boolean);
      coveredDomains.push(...parts);
    }
  }

  // 去重
  const uniqueDomains = Array.from(
    new Set(coveredDomains.map(d => d.trim()))
  ).filter(Boolean);
  if (uniqueDomains.length === 0) return null;

  // 支持用户输入多个逗号/空格分隔的主机名
  const currentHostnames = newFormInline.value.hostname
    .split(/[\s,;]+/)
    .map((h: string) => h.trim())
    .filter(Boolean);

  if (currentHostnames.length === 0) return null;

  const unmatchedHosts = currentHostnames.filter(
    (host: string) =>
      !uniqueDomains.some(pattern => matchDomain(host, pattern))
  );

  if (unmatchedHosts.length > 0) {
    return {
      hostname: unmatchedHosts.join(", "),
      covered: uniqueDomains.join(", ")
    };
  }

  return null;
});

async function fetchCertificates() {
  try {
    const res = await getCertList();
    if (res?.code === 0 && res?.data?.list) {
      certListRaw.value = res.data.list;
      certOptions.value = res.data.list.map((c: any) => {
        const idVal = c.CertId || c.cert_id || `id-${c.Id || c.id}`;
        const cnVal = c.SubjectCN || c.subject_cn || c.Name || c.name || idVal;
        return {
          label: `${cnVal} (${idVal})`,
          value: idVal
        };
      });
    }
  } catch (e) {}
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

async function fetchRules() {
  try {
    const res = await getRuleList();
    if (res?.code === 0 && res?.data?.list) {
      availableRules.value = res.data.list.map((r: any) => ({
        label: r.Name || r.name || `Rule #${r.Id || r.id}`,
        value: r.Name || r.name || `Rule #${r.Id || r.id}`
      }));
    }
  } catch (e) {}
}

function handleProxyHeadersChange(vals: string[]) {
  newFormInline.value.proxy_headers = JSON.stringify(vals);
}

function handleRulesChange(vals: string[]) {
  newFormInline.value.rules = JSON.stringify(vals);
}

function handleTunnelChange(val: string) {
  const match = tunnelOptions.value.find(tItem => tItem.value === val);
  if (match) {
    newFormInline.value.tunnel_type = match.type;
  }
}

function addLocation() {
  locationList.value.push({
    Path:
      locationList.value.length === 0
        ? "/"
        : `/path-${locationList.value.length + 1}`,
    Upstream: {
      Type: "proxy_pass",
      Data: {
        Method: "round_robin",
        Servers: [{ Target: "http://127.0.0.1:8080", Weight: 1 }]
      }
    }
  });
  syncLocationJSON();
}

function handleLocationTypeChange(loc: LocationItem) {
  const type = loc.Upstream.Type;
  if (type === "proxy_pass") {
    if (!loc.Upstream.Data) loc.Upstream.Data = {};
    if (!loc.Upstream.Data.Method) loc.Upstream.Data.Method = "round_robin";
    if (
      !Array.isArray(loc.Upstream.Data.Servers) ||
      loc.Upstream.Data.Servers.length === 0
    ) {
      loc.Upstream.Data.Servers = [
        { Target: "http://127.0.0.1:8080", Weight: 1 }
      ];
    }
    delete loc.Upstream.Data.Dir;
  } else if (type === "root" || type === "alias") {
    if (!loc.Upstream.Data) loc.Upstream.Data = {};
    if (!loc.Upstream.Data.Dir) loc.Upstream.Data.Dir = "./static";
    delete loc.Upstream.Data.Method;
    delete loc.Upstream.Data.Servers;
  }
  syncLocationJSON();
}

function quickSetPathPrefix(loc: LocationItem, prefix: string) {
  let curPath = (loc.Path || "").trim();
  // Strip existing prefix if any (=, ~*, ~)
  curPath = curPath.replace(/^(=\/?|~\*?\/?|\/?)/, "");
  if (!curPath.startsWith("/")) {
    curPath = "/" + curPath;
  }
  if (prefix === "=") {
    loc.Path = "=" + curPath;
  } else if (prefix === "~") {
    loc.Path = "~" + curPath;
  } else if (prefix === "~*") {
    loc.Path = "~*" + curPath;
  } else {
    loc.Path = curPath;
  }
  syncLocationJSON();
}

function removeLocation(idx: number) {
  if (locationList.value.length <= 1) return;
  locationList.value.splice(idx, 1);
  syncLocationJSON();
}

function addUpstreamServer(locIdx: number) {
  const loc = locationList.value[locIdx];
  if (!loc || loc.Upstream.Type !== "proxy_pass") return;
  if (!loc.Upstream.Data) loc.Upstream.Data = {};
  if (!Array.isArray(loc.Upstream.Data.Servers)) loc.Upstream.Data.Servers = [];
  loc.Upstream.Data.Servers.push({
    Target: "http://127.0.0.1:8081",
    Weight: 1
  });
  syncLocationJSON();
}

function removeUpstreamServer(locIdx: number, serverIdx: number) {
  const loc = locationList.value[locIdx];
  if (!loc || loc.Upstream.Type !== "proxy_pass" || !loc.Upstream.Data?.Servers)
    return;
  if (loc.Upstream.Data.Servers.length <= 1) return;
  loc.Upstream.Data.Servers.splice(serverIdx, 1);
  syncLocationJSON();
}

function syncLocationJSON() {
  const cleanedLocations = locationList.value.map(loc => {
    const type = loc.Upstream?.Type || "proxy_pass";
    if (type === "root" || type === "alias") {
      return {
        Path: loc.Path,
        Upstream: {
          Type: type,
          Data: {
            Dir: loc.Upstream?.Data?.Dir || "./static"
          }
        }
      };
    }
    return {
      Path: loc.Path,
      Upstream: {
        Type: "proxy_pass",
        Data: {
          Method: loc.Upstream?.Data?.Method || "round_robin",
          Servers: (
            loc.Upstream?.Data?.Servers || [
              { Target: "http://127.0.0.1:8080", Weight: 1 }
            ]
          ).map(s => ({
            Target: s.Target,
            Weight: s.Weight || 1
          }))
        }
      }
    };
  });
  newFormInline.value.location_json = JSON.stringify(cleanedLocations, null, 2);
}

const validateCertificate = (_rule: any, _value: any, callback: any) => {
  if (newFormInline.value.tls && !newFormInline.value.certificate) {
    callback(
      new Error(
        t("http.certRequiredForTls", "开启 HTTPS (TLS) 必须选择关联证书")
      )
    );
  } else {
    callback();
  }
};

function handleTlsChange(val: boolean) {
  if (!val) {
    newFormInline.value.h2 = false;
    newFormInline.value.hsts = false;
  }
  if (httpFormRef.value) {
    httpFormRef.value.validateField("certificate", () => {});
  }
}

const formRules = reactive({
  name: [
    {
      required: true,
      message: () => t("http.nameRequired", "请输入应用名称"),
      trigger: "blur"
    }
  ],
  hostname: [
    { required: true, message: () => t("http.hostname"), trigger: "blur" }
  ],
  port: [{ required: true, message: () => t("http.port"), trigger: "blur" }],
  certificate: [{ validator: validateCertificate, trigger: ["change", "blur"] }]
});

function getRef() {
  return httpFormRef.value;
}

defineExpose({ getRef, syncLocationJSON });
</script>

<template>
  <el-form
    ref="httpFormRef"
    :label-position="deviceDetection() ? 'top' : 'right'"
    :model="newFormInline"
    :rules="formRules"
    label-width="auto"
    class="http-form p-1 sm:px-2"
  >
    <div class="space-y-4">
      <!-- Section 1: Front 基础接入属性 -->
      <el-card
        shadow="never"
        class="mb-4 border-(--el-border-color-lighter)! rounded-xl"
      >
        <template #header>
          <div class="flex items-center space-x-2">
            <div class="w-1.5 h-4 bg-(--el-color-primary) rounded-full" />
            <span
              class="font-bold text-(--el-text-color-primary) text-sm sm:text-base"
              >{{ t("http.frontSection") }}</span
            >
          </div>
        </template>
        <el-row :gutter="16">
          <re-col :value="24" :xs="24">
            <el-form-item :label="t('http.name')" prop="name">
              <el-input
                v-model="newFormInline.name"
                :placeholder="t('http.searchNamePlaceholder')"
                clearable
              />
            </el-form-item>
          </re-col>

          <re-col :value="17" :xs="24" :sm="17">
            <el-form-item :label="t('http.hostname')" prop="hostname">
              <el-input
                v-model="newFormInline.hostname"
                :placeholder="
                  t('http.hostnamePlaceholder', '例如 foo.com 或 *.example.com')
                "
                clearable
              />
              <div
                v-if="certMismatchWarning"
                class="text-xs text-amber-600 dark:text-amber-400 mt-1.5 flex items-center gap-1 font-medium"
              >
                <WarningIcon class="shrink-0 text-sm" />
                <span>{{ t("http.hostnameCertMismatchTip") }}</span>
              </div>
              <div v-else class="text-xs text-(--el-text-color-secondary) mt-1.5">
                {{ t("http.hostnameTip") }}
              </div>
            </el-form-item>
          </re-col>

          <re-col :value="7" :xs="24" :sm="7">
            <el-form-item :label="t('http.port')" prop="port">
              <el-input-number
                v-model="newFormInline.port"
                :min="1"
                :max="65535"
                :step="1"
                class="w-full!"
                controls-position="right"
                placeholder="80"
              />
            </el-form-item>
          </re-col>

          <re-col :value="24" :xs="24">
            <el-form-item
              :label="t('http.proxyHeaders', '转发头')"
              prop="proxy_headers"
            >
              <el-select
                v-model="selectedProxyHeaders"
                multiple
                filterable
                allow-create
                default-first-option
                clearable
                class="w-full"
                :placeholder="
                  t(
                    'http.proxyHeadersPlaceholder',
                    '请选择或输入转发头 (如 Host, X-Forwarded-For)'
                  )
                "
                @change="handleProxyHeadersChange"
              >
                <el-option
                  v-for="item in commonHeaderOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </re-col>

          <re-col :value="24" :xs="24">
            <el-form-item :label="t('http.frontSection')">
              <div
                class="p-3 rounded-lg border border-(--el-border-color-lighter) bg-(--el-fill-color-light) w-full space-y-3"
              >
                <div class="flex flex-wrap gap-4 sm:gap-6 items-center">
                  <el-checkbox
                    v-model="newFormInline.http"
                    :label="t('http.enableHttp')"
                  />
                  <el-checkbox
                    v-model="newFormInline.tls"
                    :label="t('http.enableTls')"
                    @change="handleTlsChange"
                  />
                  <el-switch
                    v-show="newFormInline.tls"
                    v-model="newFormInline.h2"
                    :active-text="t('http.http2')"
                    inactive-text=""
                  />
                  <el-switch
                    v-show="newFormInline.tls"
                    v-model="newFormInline.hsts"
                    :active-text="t('http.hsts')"
                    inactive-text=""
                  />
                </div>

                <div
                  v-show="newFormInline.tls"
                  class="pt-2 border-t border-(--el-border-color-lighter)"
                >
                  <el-form-item
                    :label="t('http.selectCert')"
                    label-width="110px"
                    class="mb-0!"
                    prop="certificate"
                  >
                    <el-select
                      v-model="newFormInline.certificate"
                      clearable
                      :placeholder="t('http.selectCertPlaceholder')"
                      class="w-full"
                      @change="() => httpFormRef?.validateField('certificate', () => {})"
                    >
                      <el-option
                        v-for="c in certOptions"
                        :key="c.value"
                        :label="c.label"
                        :value="c.value"
                      />
                    </el-select>

                    <!-- 证书域名不匹配黄色提示条 -->
                    <div
                      v-if="certMismatchWarning"
                      class="mt-2.5 p-2.5 bg-amber-50 dark:bg-amber-950/40 border border-amber-300 dark:border-amber-700/60 rounded-lg text-xs text-amber-700 dark:text-amber-300 flex items-start gap-2 leading-relaxed"
                    >
                      <WarningIcon class="shrink-0 text-amber-500 text-sm mt-0.5" />
                      <div>
                        <span class="font-bold">{{ t("http.certMismatchTitle") }}：</span>
                        <span>{{ t("http.certMismatchDesc", { hostname: certMismatchWarning.hostname, covered: certMismatchWarning.covered }) }}</span>
                      </div>
                    </div>
                  </el-form-item>
                </div>
              </div>
            </el-form-item>
          </re-col>

          <re-col :value="24" :xs="24">
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

      <!-- Section 2: Feature 特性设置 -->
      <el-card
        shadow="never"
        class="mb-4 border-(--el-border-color-lighter)! rounded-xl"
      >
        <template #header>
          <div class="flex items-center space-x-2">
            <div class="w-1.5 h-4 bg-emerald-500 rounded-full" />
            <span
              class="font-bold text-(--el-text-color-primary) text-sm sm:text-base"
              >{{ t("http.featureSection") }}</span
            >
          </div>
        </template>
        <div
          class="p-3.5 rounded-lg border border-(--el-border-color-lighter) bg-(--el-fill-color-light)"
        >
          <div
            class="flex flex-col sm:flex-row sm:items-center justify-between gap-3"
          >
            <div>
              <div class="font-bold text-sm text-(--el-text-color-primary)">
                {{ t("http.compress") }}
              </div>
              <div class="text-xs text-(--el-text-color-secondary) mt-1">
                {{ t("http.compressTip") }}
              </div>
            </div>
            <el-switch
              v-model="newFormInline.compress"
              class="shrink-0 self-start sm:self-auto"
            />
          </div>
        </div>
      </el-card>

      <!-- Section 3: Rule 中间件规则 -->
      <el-card
        shadow="never"
        class="mb-4 border-(--el-border-color-lighter)! rounded-xl"
      >
        <template #header>
          <div class="flex items-center space-x-2">
            <div class="w-1.5 h-4 bg-amber-500 rounded-full" />
            <span
              class="font-bold text-(--el-text-color-primary) text-sm sm:text-base"
              >{{ t("http.ruleSection") }}</span
            >
          </div>
        </template>
        <el-form-item :label="t('http.mountRules')" prop="rules">
          <el-select
            v-model="selectedRules"
            multiple
            clearable
            class="w-full"
            :placeholder="t('http.mountRulesPlaceholder')"
            @change="handleRulesChange"
          >
            <el-option
              v-for="r in availableRules"
              :key="r.value"
              :label="r.label"
              :value="r.value"
            />
          </el-select>
        </el-form-item>
        <div class="text-xs/relaxed text-(--el-text-color-secondary) mt-1">
          {{ t("http.mountRulesTip") }}
        </div>
      </el-card>

      <!-- Section 4: Backend 反向代理与路由 -->
      <el-card
        shadow="never"
        class="mb-4 border-(--el-border-color-lighter)! rounded-xl"
      >
        <template #header>
          <div class="flex items-center space-x-2">
            <div class="w-1.5 h-4 bg-purple-500 rounded-full" />
            <span
              class="font-bold text-(--el-text-color-primary) text-sm sm:text-base"
              >{{ t("http.backendSection") }}</span
            >
          </div>
        </template>
        <div class="space-y-4">
          <!-- General Backend Fields -->
          <el-row :gutter="16">
            <re-col :value="24" :xs="24">
              <el-form-item :label="t('http.assocTunnel', 'Tunnel')">
                <el-select
                  v-model="selectedTunnelNodeKey"
                  clearable
                  filterable
                  :placeholder="
                    t(
                      'http.assocTunnelPlaceholder',
                      '选择关联的 Tunnel 客户端节点'
                    )
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
                      <div v-else class="text-xs text-gray-400 py-0.5">
                        {{ item.cName }}
                      </div>
                    </el-option>
                  </el-option-group>
                </el-select>
              </el-form-item>
            </re-col>

            <re-col :value="24" :xs="24">
              <el-form-item :label="t('http.dnsResolver', 'DNS')">
                <div class="w-full flex flex-col space-y-2">
                  <div ref="dnsResolverGroupRef" class="w-full space-y-2">
                    <div
                      v-for="(dns, dIdx) in dnsResolverList"
                      :key="dns.id"
                      class="flex items-center space-x-2"
                    >
                      <el-input
                        v-model="dns.value"
                        :placeholder="
                          t(
                            'http.dnsResolverPlaceholder',
                            '8.8.8.8:53 (留空使用系统默认 DNS)'
                          )
                        "
                        clearable
                        @input="syncDnsResolver"
                      />
                      <el-button
                        type="danger"
                        link
                        :icon="useRenderIcon(Delete)"
                        @click="removeDnsResolver(dIdx)"
                      />
                      <el-tooltip
                        :content="t('http.dragToReorder', '拖动改变顺序')"
                        placement="top"
                      >
                        <el-button
                          type="info"
                          link
                          class="cursor-move drag-handle text-(--el-text-color-secondary)! hover:text-(--el-color-primary)!"
                          :icon="useRenderIcon(Rank)"
                        />
                      </el-tooltip>
                    </div>
                  </div>
                  <el-button
                    type="primary"
                    plain
                    size="small"
                    class="self-start"
                    :icon="useRenderIcon(AddFill)"
                    @click="addDnsResolver"
                  >
                    {{ t("http.addDns", "添加 DNS") }}
                  </el-button>
                </div>
              </el-form-item>
            </re-col>
          </el-row>

          <!-- Location Routers List -->
          <div class="border-t border-(--el-border-color-lighter) pt-4">
            <div
              class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-3"
            >
              <div>
                <span
                  class="font-bold text-sm text-(--el-text-color-primary)"
                  >{{ t("http.locationTitle") }}</span
                >
                <div class="text-xs text-(--el-text-color-secondary) mt-0.5">
                  {{ t("http.locationTip") }}
                </div>
              </div>
              <el-button
                type="primary"
                size="small"
                :icon="useRenderIcon(AddFill)"
                class="self-start sm:self-auto shrink-0"
                @click="addLocation"
              >
                {{ t("http.addLocation") }}
              </el-button>
            </div>

            <div class="space-y-4">
              <div
                v-for="(loc, lIdx) in locationList"
                :key="lIdx"
                class="p-3 sm:p-4 rounded-xl border border-(--el-border-color-lighter) bg-(--el-fill-color-light) transition-all"
              >
                <div
                  class="flex-bc mb-3 pb-2 border-b border-(--el-border-color-lighter)"
                >
                  <div class="flex items-center space-x-2">
                    <el-tag
                      size="small"
                      type="primary"
                      effect="dark"
                      class="font-bold font-mono"
                    >
                      Location #{{ lIdx + 1 }}
                    </el-tag>
                    <span class="text-xs text-(--el-text-color-secondary)"
                      >Path</span
                    >
                  </div>
                  <el-button
                    size="small"
                    link
                    type="danger"
                    :disabled="locationList.length <= 1"
                    :icon="useRenderIcon(Delete)"
                    @click="removeLocation(lIdx)"
                  >
                    {{ t("http.deleteLocation") }}
                  </el-button>
                </div>

                <el-row :gutter="12">
                  <re-col :value="10" :xs="24">
                    <el-form-item
                      :label="t('http.matchPath')"
                      class="mb-3!"
                      required
                    >
                      <div class="w-full space-y-1">
                        <el-input
                          v-model="loc.Path"
                          placeholder="/ or =/api or ~*\.(jpg|png)$"
                          @input="syncLocationJSON"
                        />
                        <div class="flex items-center gap-1 flex-wrap">
                          <span
                            class="text-[11px] text-(--el-text-color-secondary)"
                            >{{ t("http.matchMode") }}</span
                          >
                          <el-button
                            size="small"
                            link
                            type="primary"
                            @click="quickSetPathPrefix(loc, '/')"
                            >{{ t("http.prefixMatch") }}</el-button
                          >
                          <el-button
                            size="small"
                            link
                            type="success"
                            @click="quickSetPathPrefix(loc, '=')"
                            >{{ t("http.exactMatch") }}</el-button
                          >
                          <el-button
                            size="small"
                            link
                            type="warning"
                            @click="quickSetPathPrefix(loc, '~')"
                            >{{ t("http.regexMatch") }}</el-button
                          >
                          <el-button
                            size="small"
                            link
                            type="danger"
                            @click="quickSetPathPrefix(loc, '~*')"
                            >{{ t("http.regexIgnoreMatch") }}</el-button
                          >
                        </div>
                      </div>
                    </el-form-item>
                  </re-col>
                  <re-col :value="8" :xs="24">
                    <el-form-item :label="t('http.upstreamType')" class="mb-3!">
                      <el-select
                        v-model="loc.Upstream.Type"
                        class="w-full"
                        @change="handleLocationTypeChange(loc)"
                      >
                        <el-option
                          :label="t('http.proxyPass')"
                          value="proxy_pass"
                        />
                        <el-option :label="t('http.rootDir')" value="root" />
                        <el-option :label="t('http.aliasDir')" value="alias" />
                      </el-select>
                    </el-form-item>
                  </re-col>
                  <re-col
                    v-if="loc.Upstream.Type === 'proxy_pass'"
                    :value="6"
                    :xs="24"
                  >
                    <el-form-item :label="t('http.lbAlgorithm')" class="mb-3!">
                      <el-select
                        v-model="loc.Upstream.Data.Method"
                        class="w-full"
                        @change="syncLocationJSON"
                      >
                        <el-option
                          :label="t('http.roundRobin')"
                          value="round_robin"
                        />
                        <el-option
                          :label="t('http.weightRoundRobin')"
                          value="weight"
                        />
                        <el-option :label="t('http.ipHash')" value="ip_hash" />
                      </el-select>
                    </el-form-item>
                  </re-col>
                </el-row>

                <!-- Case A: Upstream Target Servers (proxy_pass) -->
                <div
                  v-if="loc.Upstream.Type === 'proxy_pass'"
                  class="mt-1 p-2.5 sm:p-3 bg-(--el-bg-color) rounded-lg border border-(--el-border-color-lighter) space-y-2"
                >
                  <div
                    class="flex-bc text-xs font-bold text-(--el-text-color-primary)"
                  >
                    <span>{{ t("http.targetUrl") }}</span>
                    <el-button
                      size="small"
                      link
                      type="primary"
                      @click="addUpstreamServer(lIdx)"
                      >{{ t("http.addBackendNode") }}</el-button
                    >
                  </div>

                  <div
                    v-for="(srv, sIdx) in loc.Upstream.Data.Servers || []"
                    :key="sIdx"
                    class="flex flex-wrap sm:flex-nowrap items-center gap-2 pb-2 sm:pb-0 border-b sm:border-b-0 border-(--el-border-color-lighter) last:border-b-0"
                  >
                    <el-input
                      v-model="srv.Target"
                      size="small"
                      placeholder="http://127.0.0.1:8080"
                      class="w-full sm:flex-1"
                      @input="syncLocationJSON"
                    />
                    <div
                      class="flex items-center space-x-2 shrink-0 self-end sm:self-auto"
                    >
                      <span class="text-xs text-(--el-text-color-secondary)">{{
                        t("http.weight")
                      }}</span>
                      <el-input-number
                        v-model="srv.Weight"
                        size="small"
                        :min="1"
                        :max="100"
                        class="w-24!"
                        @change="syncLocationJSON"
                      />
                      <el-button
                        size="small"
                        link
                        type="danger"
                        :disabled="
                          (loc.Upstream.Data.Servers || []).length <= 1
                        "
                        :icon="useRenderIcon(Delete)"
                        @click="removeUpstreamServer(lIdx, sIdx)"
                      />
                    </div>
                  </div>
                </div>

                <!-- Case B: Static Dir (root or alias) -->
                <div
                  v-else
                  class="mt-1 p-2.5 sm:p-3 bg-(--el-bg-color) rounded-lg border border-(--el-border-color-lighter) space-y-2"
                >
                  <el-form-item
                    :label="t('http.staticDir')"
                    class="mb-0!"
                    required
                  >
                    <el-input
                      v-model="loc.Upstream.Data.Dir"
                      :placeholder="t('http.staticDirPlaceholder')"
                      @input="syncLocationJSON"
                    />
                  </el-form-item>
                  <div class="text-[11px] text-(--el-text-color-secondary)">
                    <span v-if="loc.Upstream.Type === 'root'">{{
                      t("http.rootTip")
                    }}</span>
                    <span v-else>{{ t("http.aliasTip") }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-card>
    </div>
  </el-form>
</template>

<style scoped>
.http-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--el-text-color-regular);
  white-space: nowrap;
}
</style>
