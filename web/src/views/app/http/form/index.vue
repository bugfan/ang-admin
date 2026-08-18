<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import ReCol from "@/components/ReCol";
import { useI18n } from "vue-i18n";
import { getCertList } from "@/api/certificate";
import { getRuleList } from "@/api/rule";
import { getTunnelList } from "@/api/tunnel";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import AddFill from "~icons/ri/add-line";
import Delete from "~icons/ep/delete";

export interface LocationItem {
  Path: string;
  Upstream: {
    Type: string;
    Data: {
      Method: string;
      Servers: Array<{ Target: string; Weight: number }>;
    };
  };
}

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "",
    id: undefined,
    name: "",
    port: "443",
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
    location_json: JSON.stringify([
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
    ], null, 2),
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
const tunnelOptions = ref<Array<{ label: string; value: string; type: string }>>([]);
// Rule options
const availableRules = ref<Array<{ label: string; value: string }>>([]);
// Selected Rules array
const selectedRules = ref<string[]>([]);
// Location items array
const locationList = ref<LocationItem[]>([]);

onMounted(() => {
  fetchCertificates();
  fetchTunnels();
  fetchRules();
  initFormState();
});

function initFormState() {
  try {
    if (newFormInline.value.proxy_headers) {
      const parsed = typeof newFormInline.value.proxy_headers === "string"
        ? JSON.parse(newFormInline.value.proxy_headers)
        : newFormInline.value.proxy_headers;
      if (Array.isArray(parsed)) selectedProxyHeaders.value = parsed;
    }
  } catch (e) {}
  try {
    if (newFormInline.value.rules) {
      const parsed = typeof newFormInline.value.rules === "string"
        ? JSON.parse(newFormInline.value.rules)
        : newFormInline.value.rules;
      if (Array.isArray(parsed)) selectedRules.value = parsed;
    }
  } catch (e) {}

  try {
    if (newFormInline.value.location_json) {
      const parsed = typeof newFormInline.value.location_json === "string"
        ? JSON.parse(newFormInline.value.location_json)
        : newFormInline.value.location_json;
      if (Array.isArray(parsed)) locationList.value = parsed;
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
}

async function fetchCertificates() {
  try {
    const res = await getCertList();
    if (res?.code === 0 && res?.data?.list) {
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
    if (res?.code === 0 && res?.data?.list) {
      tunnelOptions.value = res.data.list.map((tItem: any) => {
        const idVal = String(tItem.Id || tItem.id);
        const tType = tItem.Type || tItem.type || "TLS";
        const sni = tItem.SNI || tItem.sni || "";
        const port = tItem.Port || tItem.port || "";
        return {
          label: `Tunnel #${idVal} (${tType} | Port: ${port} ${sni ? '| SNI: ' + sni : ''})`,
          value: idVal,
          type: tType.toLowerCase().includes("quic") ? "quic" : "tls"
        };
      });
    }
  } catch (e) {}
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
    Path: `/path-${locationList.value.length + 1}`,
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

function removeLocation(idx: number) {
  if (locationList.value.length <= 1) return;
  locationList.value.splice(idx, 1);
  syncLocationJSON();
}

function addUpstreamServer(locIdx: number) {
  locationList.value[locIdx].Upstream.Data.Servers.push({ Target: "http://127.0.0.1:8081", Weight: 1 });
  syncLocationJSON();
}

function removeUpstreamServer(locIdx: number, serverIdx: number) {
  if (locationList.value[locIdx].Upstream.Data.Servers.length <= 1) return;
  locationList.value[locIdx].Upstream.Data.Servers.splice(serverIdx, 1);
  syncLocationJSON();
}

function syncLocationJSON() {
  newFormInline.value.location_json = JSON.stringify(locationList.value, null, 2);
}

const formRules = reactive({
  name: [{ required: true, message: () => t("http.nameRequired", "请输入应用名称"), trigger: "blur" }],
  hostname: [{ required: true, message: () => t("http.hostname"), trigger: "blur" }],
  port: [{ required: true, message: () => t("http.port"), trigger: "blur" }]
});

function getRef() {
  return httpFormRef.value;
}

defineExpose({ getRef, syncLocationJSON });
</script>

<template>
  <el-form
    ref="httpFormRef"
    :model="newFormInline"
    :rules="formRules"
    label-width="120px"
    class="http-form px-1 sm:px-2 py-1"
  >
    <div class="space-y-4">
      <!-- Section 1: Front 基础接入属性 -->
      <el-card shadow="never" class="mb-4 !border-[var(--el-border-color-lighter)] rounded-xl">
        <template #header>
          <div class="flex items-center space-x-2">
            <div class="w-1.5 h-4 bg-[var(--el-color-primary)] rounded-full"></div>
            <span class="font-bold text-[var(--el-text-color-primary)] text-sm sm:text-base">{{ t("http.frontSection") }}</span>
          </div>
        </template>
        <el-row :gutter="16">
          <re-col :value="12" :xs="24">
            <el-form-item :label="t('http.name')" prop="name">
              <el-input
                v-model="newFormInline.name"
                :placeholder="t('http.searchNamePlaceholder')"
                clearable
              />
            </el-form-item>
          </re-col>

          <re-col :value="12" :xs="24">
            <el-form-item :label="t('http.port')" prop="port">
              <el-input
                v-model="newFormInline.port"
                placeholder="443 / 80"
                clearable
              />
            </el-form-item>
          </re-col>

          <re-col :value="24" :xs="24">
            <el-form-item :label="t('http.hostname')" prop="hostname">
              <el-input
                v-model="newFormInline.hostname"
                :placeholder="t('http.hostnamePlaceholder', '例如 foo.com 或 *.example.com')"
                clearable
              />
              <div class="text-xs text-[var(--el-text-color-secondary)] mt-1.5">
                {{ t("http.hostnameTip") }}
              </div>
            </el-form-item>
          </re-col>

          <re-col :value="24" :xs="24">
            <el-form-item :label="t('http.proxyHeaders', '转发头')" prop="proxy_headers">
              <el-select
                v-model="selectedProxyHeaders"
                multiple
                filterable
                allow-create
                default-first-option
                clearable
                class="w-full"
                :placeholder="t('http.proxyHeadersPlaceholder', '请选择或输入转发头 (如 Host, X-Forwarded-For)')"
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
              <div class="p-3 rounded-lg border border-[var(--el-border-color-lighter)] bg-[var(--el-fill-color-light)] w-full space-y-3">
                <div class="flex flex-wrap gap-4 sm:gap-6 items-center">
                  <el-checkbox v-model="newFormInline.http" :label="t('http.enableHttp')" />
                  <el-checkbox v-model="newFormInline.tls" :label="t('http.enableTls')" />
                  <el-switch v-show="newFormInline.tls" v-model="newFormInline.h2" :active-text="t('http.http2')" inactive-text="" />
                  <el-switch v-show="newFormInline.tls" v-model="newFormInline.hsts" :active-text="t('http.hsts')" inactive-text="" />
                </div>

                <div v-show="newFormInline.tls" class="pt-2 border-t border-[var(--el-border-color-lighter)]">
                  <el-form-item :label="t('http.selectCert')" label-width="110px" class="!mb-0">
                    <el-select
                      v-model="newFormInline.certificate"
                      clearable
                      :placeholder="t('http.selectCertPlaceholder')"
                      class="w-full"
                    >
                      <el-option
                        v-for="c in certOptions"
                        :key="c.value"
                        :label="c.label"
                        :value="c.value"
                      />
                    </el-select>
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
      <el-card shadow="never" class="mb-4 !border-[var(--el-border-color-lighter)] rounded-xl">
        <template #header>
          <div class="flex items-center space-x-2">
            <div class="w-1.5 h-4 bg-emerald-500 rounded-full"></div>
            <span class="font-bold text-[var(--el-text-color-primary)] text-sm sm:text-base">{{ t("http.featureSection") }}</span>
          </div>
        </template>
        <div class="p-3.5 rounded-lg border border-[var(--el-border-color-lighter)] bg-[var(--el-fill-color-light)]">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <div class="font-bold text-sm text-[var(--el-text-color-primary)]">{{ t("http.compress") }}</div>
              <div class="text-xs text-[var(--el-text-color-secondary)] mt-1">
                {{ t("http.compressTip") }}
              </div>
            </div>
            <el-switch v-model="newFormInline.compress" class="shrink-0 self-start sm:self-auto" />
          </div>
        </div>
      </el-card>

      <!-- Section 3: Rule 中间件规则 -->
      <el-card shadow="never" class="mb-4 !border-[var(--el-border-color-lighter)] rounded-xl">
        <template #header>
          <div class="flex items-center space-x-2">
            <div class="w-1.5 h-4 bg-amber-500 rounded-full"></div>
            <span class="font-bold text-[var(--el-text-color-primary)] text-sm sm:text-base">{{ t("http.ruleSection") }}</span>
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
        <div class="text-xs text-[var(--el-text-color-secondary)] leading-relaxed mt-1">
          {{ t("http.mountRulesTip") }}
        </div>
      </el-card>

      <!-- Section 4: Backend 反向代理与路由 -->
      <el-card shadow="never" class="mb-4 !border-[var(--el-border-color-lighter)] rounded-xl">
        <template #header>
          <div class="flex items-center space-x-2">
            <div class="w-1.5 h-4 bg-purple-500 rounded-full"></div>
            <span class="font-bold text-[var(--el-text-color-primary)] text-sm sm:text-base">{{ t("http.backendSection") }}</span>
          </div>
        </template>
        <div class="space-y-4">
          <!-- General Backend Fields -->
          <el-row :gutter="16">
            <re-col :value="12" :xs="24">
              <el-form-item :label="t('http.assocTunnel')">
                <el-select
                  v-model="newFormInline.tunnel_id"
                  clearable
                  :placeholder="t('http.assocTunnelPlaceholder')"
                  class="w-full"
                  @change="handleTunnelChange"
                >
                  <el-option
                    v-for="tItem in tunnelOptions"
                    :key="tItem.value"
                    :label="tItem.label"
                    :value="tItem.value"
                  />
                </el-select>
              </el-form-item>
            </re-col>

            <re-col :value="12" :xs="24">
              <el-form-item :label="t('http.dnsResolver')">
                <el-input
                  v-model="newFormInline.dns_resolver"
                  placeholder="dns1"
                  clearable
                />
              </el-form-item>
            </re-col>

            <re-col :value="12" :xs="24">
              <el-form-item :label="t('http.tunnelToken')">
                <el-input
                  v-model="newFormInline.tunnel_token"
                  placeholder="Token"
                  clearable
                />
              </el-form-item>
            </re-col>
          </el-row>

          <!-- Location Routers List -->
          <div class="border-t border-[var(--el-border-color-lighter)] pt-4">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-3">
              <div>
                <span class="font-bold text-sm text-[var(--el-text-color-primary)]">{{ t("http.locationTitle") }}</span>
                <div class="text-xs text-[var(--el-text-color-secondary)] mt-0.5">
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
                class="p-3 sm:p-4 rounded-xl border border-[var(--el-border-color-lighter)] bg-[var(--el-fill-color-light)] transition-all"
              >
                <div class="flex items-center justify-between mb-3 pb-2 border-b border-[var(--el-border-color-lighter)]">
                  <div class="flex items-center space-x-2">
                    <el-tag size="small" type="primary" effect="dark" class="font-bold font-mono">
                      Location #{{ lIdx + 1 }}
                    </el-tag>
                    <span class="text-xs text-[var(--el-text-color-secondary)]">Path</span>
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
                  <re-col :value="12" :xs="24">
                    <el-form-item :label="t('http.matchPath')" class="!mb-3" required>
                      <el-input
                        v-model="loc.Path"
                        placeholder="/ or /api"
                        @input="syncLocationJSON"
                      />
                    </el-form-item>
                  </re-col>
                  <re-col :value="12" :xs="24">
                    <el-form-item :label="t('http.lbAlgorithm')" class="!mb-3">
                      <el-select
                        v-model="loc.Upstream.Data.Method"
                        class="w-full"
                        @change="syncLocationJSON"
                      >
                        <el-option :label="t('http.roundRobin')" value="round_robin" />
                        <el-option :label="t('http.weightRoundRobin')" value="weight" />
                        <el-option :label="t('http.ipHash')" value="ip_hash" />
                      </el-select>
                    </el-form-item>
                  </re-col>
                </el-row>

                <!-- Upstream Target Servers -->
                <div class="mt-1 p-2.5 sm:p-3 bg-[var(--el-bg-color)] rounded-lg border border-[var(--el-border-color-lighter)] space-y-2">
                  <div class="flex items-center justify-between text-xs font-bold text-[var(--el-text-color-primary)]">
                    <span>{{ t("http.targetUrl") }}</span>
                    <el-button size="small" link type="primary" @click="addUpstreamServer(lIdx)">{{ t("http.addBackendNode") }}</el-button>
                  </div>

                  <div
                    v-for="(srv, sIdx) in loc.Upstream.Data.Servers"
                    :key="sIdx"
                    class="flex flex-wrap sm:flex-nowrap items-center gap-2 pb-2 sm:pb-0 border-b sm:border-b-0 border-[var(--el-border-color-lighter)] last:border-b-0"
                  >
                    <el-input
                      v-model="srv.Target"
                      size="small"
                      placeholder="http://127.0.0.1:8080"
                      class="w-full sm:flex-1"
                      @input="syncLocationJSON"
                    />
                    <div class="flex items-center space-x-2 shrink-0 self-end sm:self-auto">
                      <span class="text-xs text-[var(--el-text-color-secondary)]">{{ t("http.weight") }}</span>
                      <el-input-number
                        v-model="srv.Weight"
                        size="small"
                        :min="1"
                        :max="100"
                        class="!w-24"
                        @change="syncLocationJSON"
                      />
                      <el-button
                        size="small"
                        link
                        type="danger"
                        :disabled="loc.Upstream.Data.Servers.length <= 1"
                        :icon="useRenderIcon(Delete)"
                        @click="removeUpstreamServer(lIdx, sIdx)"
                      />
                    </div>
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
}

@media (max-width: 640px) {
  .http-form :deep(.el-form-item) {
    flex-direction: column;
    align-items: flex-start;
  }
  .http-form :deep(.el-form-item__label) {
    justify-content: flex-start;
    margin-bottom: 4px;
    width: 100% !important;
  }
}
</style>
