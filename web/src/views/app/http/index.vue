<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useHttpProxy } from "./utils/hook";
import editForm from "./form/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import { createHttpProxy, updateHttpProxy } from "@/api/http_proxy";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import BackIcon from "~icons/ep/back";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";

defineOptions({
  name: "AppHttpProxy"
});

const { t } = useI18n();
const searchFormRef = ref();
const tableRef = ref();
const createEditFormRef = ref();

// View Mode: 'list' | 'new' | 'edit'
const showView = ref<"list" | "new" | "edit">("list");
const formInline = ref<any>({});
const saving = ref(false);

const {
  form,
  loading,
  columns,
  dataList,
  selectedNum,
  pagination,
  deviceDetection,
  onSearch,
  resetForm,
  onbatchDel,
  handleDelete,
  handleSizeChange,
  onSelectionCancel,
  handleCurrentChange,
  handleSelectionChange
} = useHttpProxy(t, tableRef);

function getDefaultFormInline() {
  return {
    title: t("http.addHttp"),
    id: undefined,
    name: "",
    port: "443",
    hostname: "*.local.i443.cn",
    http: true,
    tls: true,
    h2: true,
    hsts: false,
    certificate: "",
    proxy_headers: JSON.stringify(["X-Forwarded-For"]),
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
  };
}

function getFormInlineFromRow(row: any) {
  return {
    title: `${t("http.editHttp")} [ID: ${row?.Id || row?.id}]`,
    id: row?.Id ?? row?.id ?? undefined,
    name: row?.Name ?? row?.name ?? "",
    port: row?.Port ?? row?.port ?? "443",
    hostname: row?.Hostname ?? row?.hostname ?? "*.local.i443.cn",
    http: row?.HTTP ?? row?.http ?? true,
    tls: row?.TLS ?? row?.tls ?? true,
    h2: row?.H2 ?? row?.h2 ?? true,
    hsts: row?.HSTS ?? row?.hsts ?? false,
    certificate: row?.Certificate ?? row?.certificate ?? "",
    proxy_headers: row?.ProxyHeaders ?? row?.proxy_headers ?? JSON.stringify(["X-Forwarded-For"]),
    compress: row?.Compress ?? row?.compress ?? false,
    rules: row?.Rules ?? row?.rules ?? JSON.stringify([]),
    real_ip: row?.RealIp ?? row?.real_ip ?? "",
    tunnel_type: row?.TunnelType ?? row?.tunnel_type ?? "quic",
    tunnel_id: row?.TunnelId ?? row?.tunnel_id ?? "",
    tunnel_token: row?.TunnelToken ?? row?.tunnel_token ?? "",
    dns_resolver: row?.DNSResolver ?? row?.dns_resolver ?? "dns1",
    location_json: row?.LocationJSON ?? row?.location_json ?? JSON.stringify([
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
    remark: row?.Remark ?? row?.remark ?? ""
  };
}

function handleAddPage() {
  formInline.value = getDefaultFormInline();
  showView.value = "new";
}

function handleEditPage(row: any) {
  formInline.value = getFormInlineFromRow(row);
  showView.value = "edit";
}

function handleCancelPage() {
  showView.value = "list";
}

async function handleSaveSubmit() {
  if (!createEditFormRef.value) return;

  if (createEditFormRef.value.syncLocationJSON) {
    createEditFormRef.value.syncLocationJSON();
  }

  const FormRef = createEditFormRef.value.getRef();
  if (!FormRef) return;

  FormRef.validate(async (valid: boolean) => {
    if (valid) {
      saving.value = true;
      try {
        const curData = formInline.value;
        if (showView.value === "new") {
          const { code, message: msg } = await createHttpProxy(curData);
          if (code !== 0) {
            message(msg, { type: "error" });
            return;
          }
          message(t("http.addHttp") + " " + t("http.success", "成功"), { type: "success" });
        } else {
          const { code, message: msg } = await updateHttpProxy(curData);
          if (code !== 0) {
            message(msg, { type: "error" });
            return;
          }
          message(t("http.editHttp") + " " + t("http.success", "成功"), { type: "success" });
        }
        showView.value = "list";
        onSearch();
      } catch (e: any) {
        message(e.message || "提交失败", { type: "error" });
      } finally {
        saving.value = false;
      }
    }
  });
}

function formatJSON(row: any) {
  try {
    let proxyHeaders = [];
    if (row.ProxyHeaders || row.proxy_headers) {
      const ph = row.ProxyHeaders || row.proxy_headers;
      proxyHeaders = typeof ph === "string" ? JSON.parse(ph) : ph;
    }
    let rules = [];
    if (row.Rules || row.rules) {
      const r = row.Rules || row.rules;
      rules = typeof r === "string" ? JSON.parse(r) : r;
    }
    let locations = [];
    if (row.LocationJSON || row.location_json) {
      const loc = row.LocationJSON || row.location_json;
      locations = typeof loc === "string" ? JSON.parse(loc) : loc;
    }

    let backendTunnel: any = null;
    if (row.TunnelId || row.tunnel_id) {
      backendTunnel = {
        Type: row.TunnelType || row.tunnel_type || "quic",
        ID: row.TunnelId || row.tunnel_id,
        Token: row.TunnelToken || row.tunnel_token || ""
      };
    }

    const angServerHttpConfig = {
      Front: {
        Port: row.Port || row.port || "443",
        Hostname: row.Hostname || row.hostname || "",
        HTTP: row.HTTP ?? row.http ?? true,
        TLS: row.TLS ?? row.tls ?? true,
        H2: row.H2 ?? row.h2 ?? true,
        HSTS: row.HSTS ?? row.hsts ?? false,
        Certificate: row.Certificate || row.certificate || "",
        ProxyHeaders: proxyHeaders
      },
      Feature: {
        Compress: row.Compress ?? row.compress ?? false
      },
      Rule: rules,
      Backend: {
        RealIp: row.RealIp || row.real_ip || "",
        Tunnel: backendTunnel,
        DNSResolver: row.DNSResolver || row.dns_resolver || "dns1",
        Location: locations
      }
    };

    return JSON.stringify(angServerHttpConfig, null, 2);
  } catch (e) {
    return "-";
  }
}
</script>

<template>
  <div class="main">
    <!-- List View -->
    <div v-if="showView === 'list'">
      <el-form
        ref="searchFormRef"
        :inline="true"
        :model="form"
        class="search-form bg-bg_color w-full px-3 sm:px-6 pt-3 pb-1 overflow-auto mb-3 rounded-xl border border-[var(--el-border-color-lighter)] shadow-2xs"
      >
        <el-form-item :label="t('http.hostname') + '：'" prop="hostname">
          <el-input
            v-model="form.hostname"
            :placeholder="t('http.searchHostnamePlaceholder')"
            clearable
            class="w-full sm:!w-[200px]"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item :label="t('http.name') + '：'" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="t('http.searchNamePlaceholder')"
            clearable
            class="w-full sm:!w-[200px]"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :icon="useRenderIcon('ri:search-line')"
            :loading="loading"
            @click="onSearch"
          >
            {{ t("http.search", "搜索") }}
          </el-button>
          <el-button
            :icon="useRenderIcon('ri:refresh-line')"
            @click="resetForm(searchFormRef)"
          >
            {{ t("http.reset", "重置") }}
          </el-button>
        </el-form-item>
      </el-form>

      <PureTableBar
        :title="t('http.title')"
        :columns="columns"
        @refresh="onSearch"
      >
        <template #buttons>
          <el-button
            type="primary"
            :icon="useRenderIcon(AddFill)"
            @click="handleAddPage"
          >
            {{ t("http.addHttp") }}
          </el-button>
        </template>
        <template v-slot="{ size, dynamicColumns }">
          <div
            v-if="selectedNum > 0"
            class="bg-[var(--el-color-primary-light-9)] text-[var(--el-color-primary)] border border-[var(--el-color-primary-light-7)] px-4 py-2 rounded-lg text-sm mb-3 flex items-center justify-between"
          >
            <span>{{ t("http.selected", { count: selectedNum }) }}</span>
            <div>
              <el-button type="primary" link size="small" @click="onSelectionCancel">
                {{ t("http.cancelSelection") }}
              </el-button>
              <el-popconfirm
                :title="t('http.batchDeleteConfirm')"
                @confirm="onbatchDel"
              >
                <template #reference>
                  <el-button type="danger" link size="small">
                    {{ t("http.batchDelete", "批量删除") }}
                  </el-button>
                </template>
              </el-popconfirm>
            </div>
          </div>

          <pure-table
            ref="tableRef"
            row-key="id"
            align-whole="center"
            table-layout="auto"
            :loading="loading"
            :size="size"
            :adaptive="true"
            :data="dataList"
            :columns="dynamicColumns"
            :pagination="pagination"
            :paginationSmall="size === 'small'"
            :header-cell-style="{
              background: 'var(--el-fill-color-light)',
              color: 'var(--el-text-color-primary)',
              fontWeight: 'bold'
            }"
            @selection-change="handleSelectionChange"
            @page-size-change="handleSizeChange"
            @page-current-change="handleCurrentChange"
          >
            <!-- Expand Row Slot: Theme-Adaptive JSON Preview -->
            <template #expand="{ row }">
              <div class="p-3 sm:p-4 bg-[var(--el-fill-color-light)] rounded-xl m-1 sm:m-2 border border-[var(--el-border-color-lighter)]">
                <div class="text-xs font-bold text-[var(--el-text-color-regular)] mb-2 flex items-center justify-between flex-wrap gap-1">
                  <div class="flex items-center space-x-2">
                    <div class="w-2 h-2 bg-[var(--el-color-primary)] rounded-full"></div>
                    <span>ang HTTP JSON [ID: {{ row.Id || row.id }}]</span>
                  </div>
                  <span class="text-[var(--el-text-color-secondary)] font-mono text-[11px]">HTTP Server JSON</span>
                </div>
                <div class="bg-[var(--el-bg-color)] p-3 rounded-lg border border-[var(--el-border-color-lighter)]">
                  <div class="flex items-center justify-between mb-2 pb-1.5 border-b border-[var(--el-border-color-lighter)]">
                    <span class="text-xs font-bold text-[var(--el-color-primary)]">Front + Feature + Rule + Backend</span>
                    <el-tag size="small" type="primary" effect="plain" class="font-mono">JSON Config</el-tag>
                  </div>
                  <el-scrollbar max-height="220px" class="item-scrollbar pr-1">
                    <pre class="text-xs text-[var(--el-text-color-primary)] font-mono whitespace-pre-wrap break-all leading-relaxed">{{ formatJSON(row) }}</pre>
                  </el-scrollbar>
                </div>
              </div>
            </template>

            <!-- Operation Slot -->
            <template #operation="{ row }">
              <el-button
                class="reset-margin"
                link
                type="primary"
                :size="size"
                :icon="useRenderIcon(EditPen)"
                @click="handleEditPage(row)"
              >
                {{ t("http.edit", "编辑") }}
              </el-button>
              <el-popconfirm
                :title="t('http.deleteConfirm')"
                @confirm="handleDelete(row)"
              >
                <template #reference>
                  <el-button
                    class="reset-margin"
                    link
                    type="danger"
                    :size="size"
                    :icon="useRenderIcon(Delete)"
                  >
                    {{ t("http.delete", "删除") }}
                  </el-button>
                </template>
              </el-popconfirm>
            </template>
          </pure-table>
        </template>
      </PureTableBar>
    </div>

    <!-- Create / Edit Full Page View -->
    <div v-else-if="showView === 'new' || showView === 'edit'" class="p-3 sm:p-5 bg-bg_color rounded-xl border border-[var(--el-border-color-lighter)] shadow-2xs">
      <!-- Full Page Header Bar -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-4 mb-4 border-b border-[var(--el-border-color-lighter)]">
        <div class="flex items-center space-x-3">
          <el-button
            circle
            :icon="useRenderIcon(BackIcon)"
            :title="t('http.backToList', '返回应用列表')"
            @click="handleCancelPage"
          />
          <div>
            <h2 class="text-base sm:text-lg font-bold text-[var(--el-text-color-primary)]">
              {{ formInline.title }}
            </h2>
            <div class="text-xs text-[var(--el-text-color-secondary)] mt-0.5">
              {{ t('http.headerDesc') }}
            </div>
          </div>
        </div>

        <div class="flex items-center space-x-2 sm:space-x-3 shrink-0 self-end sm:self-auto">
          <el-button
            :icon="useRenderIcon(CloseIcon)"
            @click="handleCancelPage"
          >
            {{ t("http.cancel") }}
          </el-button>
          <el-button
            type="primary"
            :loading="saving"
            :icon="useRenderIcon(CheckIcon)"
            @click="handleSaveSubmit"
          >
            {{ t("http.save") }}
          </el-button>
        </div>
      </div>

      <!-- Main Form View embedded directly in full-page container -->
      <editForm ref="createEditFormRef" :formInline="formInline" />

      <!-- Bottom Action Bar -->
      <div class="flex items-center justify-end space-x-3 pt-4 mt-4 border-t border-[var(--el-border-color-lighter)]">
        <el-button
          :icon="useRenderIcon(CloseIcon)"
          @click="handleCancelPage"
        >
          {{ t("http.cancel") }}
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          :icon="useRenderIcon(CheckIcon)"
          @click="handleSaveSubmit"
        >
          {{ t("http.save") }}
        </el-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.search-form :deep(.el-form-item) {
  margin-bottom: 12px;
}

@media (max-width: 640px) {
  .search-form :deep(.el-form-item) {
    margin-right: 0;
    width: 100%;
  }
}
</style>
