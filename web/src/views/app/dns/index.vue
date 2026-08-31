<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useDnsProxy } from "./utils/hook";
import editForm from "./form/index.vue";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import { createDns, updateDns } from "@/api/dns";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import BackIcon from "~icons/ep/back";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";

defineOptions({
  name: "AppDns"
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
  onSearch,
  resetForm,
  onbatchDel,
  handleDelete,
  handleSizeChange,
  onSelectionCancel,
  handleCurrentChange,
  handleSelectionChange
} = useDnsProxy(t, tableRef);

type GroupedHost = {
  ip: string;
  isIPv6: boolean;
  domains: string[];
};

function parseHostsJSON(hostsJsonStr: string) {
  let aMap: Record<string, string> = {};
  let aaaaMap: Record<string, string> = {};
  try {
    if (hostsJsonStr) {
      const parsed = JSON.parse(hostsJsonStr);
      if (parsed.A) aMap = parsed.A;
      if (parsed.AAAA) aaaaMap = parsed.AAAA;
    }
  } catch (e) {
    // ignore
  }

  const aGroupMap: Record<string, string[]> = {};
  for (const [domain, ip] of Object.entries(aMap)) {
    if (!aGroupMap[ip]) aGroupMap[ip] = [];
    if (!aGroupMap[ip].includes(domain)) aGroupMap[ip].push(domain);
  }

  const aaaaGroupMap: Record<string, string[]> = {};
  for (const [domain, ip] of Object.entries(aaaaMap)) {
    if (!aaaaGroupMap[ip]) aaaaGroupMap[ip] = [];
    if (!aaaaGroupMap[ip].includes(domain)) aaaaGroupMap[ip].push(domain);
  }

  const aList: GroupedHost[] = Object.entries(aGroupMap).map(
    ([ip, domains]) => ({
      ip,
      isIPv6: false,
      domains
    })
  );

  const aaaaList: GroupedHost[] = Object.entries(aaaaGroupMap).map(
    ([ip, domains]) => ({
      ip,
      isIPv6: true,
      domains
    })
  );

  return {
    aCount: Object.keys(aMap).length,
    aaaaCount: Object.keys(aaaaMap).length,
    aMap,
    aaaaMap,
    aList,
    aaaaList
  };
}

function formatDomains(domains: string[]): string {
  if (!domains || domains.length === 0) return "";
  if (domains.length <= 2) {
    return domains.join(" ");
  }
  return `${domains[0]} ${domains[1]} ...`;
}

function parseUpstreamServers(serversStr: string) {
  try {
    if (serversStr) {
      const parsed = JSON.parse(serversStr);
      if (Array.isArray(parsed)) return parsed;
    }
  } catch (e) {
    // ignore
  }
  return [];
}

function getDefaultFormInline() {
  return {
    title: t("dns.addDns"),
    id: undefined,
    name: "",
    address: "0.0.0.0",
    port: "53",
    rules: JSON.stringify([]),
    tunnel_type: "quic",
    tunnel_id: "",
    tunnel_token: "",
    upstream_method: "round_robin",
    upstream_servers: JSON.stringify([{ target: "1.1.1.1:53", weight: 1 }]),
    hosts_text: "127.0.0.1 localhost\n::1 localhost",
    hosts_json: JSON.stringify({
      A: { localhost: "127.0.0.1" },
      AAAA: { localhost: "::1" }
    }),
    remark: ""
  };
}

function getFormInlineFromRow(row: any) {
  return {
    title: `${t("dns.editDns")} [ID: ${row?.Id || row?.id}]`,
    id: row?.Id ?? row?.id ?? undefined,
    name: row?.Name ?? row?.name ?? "",
    address: row?.Address ?? row?.address ?? "0.0.0.0",
    port: row?.Port ?? row?.port ?? "53",
    rules: row?.Rules ?? row?.rules ?? JSON.stringify([]),
    tunnel_type: row?.TunnelType ?? row?.tunnel_type ?? "quic",
    tunnel_id: row?.TunnelId ?? row?.tunnel_id ?? "",
    tunnel_token: row?.TunnelToken ?? row?.tunnel_token ?? "",
    upstream_method:
      row?.UpstreamMethod ?? row?.upstream_method ?? "round_robin",
    upstream_servers:
      row?.UpstreamServers ??
      row?.upstream_servers ??
      JSON.stringify([{ target: "1.1.1.1:53", weight: 1 }]),
    hosts_text: row?.HostsText ?? row?.hosts_text ?? "",
    hosts_json:
      row?.HostsJSON ?? row?.hosts_json ?? JSON.stringify({ A: {}, AAAA: {} }),
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
  if (createEditFormRef.value.parseHostsTextToJSON) {
    createEditFormRef.value.parseHostsTextToJSON();
  }
  const FormRef = createEditFormRef.value.getRef();
  if (!FormRef) return;

  FormRef.validate(async (valid: boolean) => {
    if (valid) {
      saving.value = true;
      try {
        const curData = formInline.value;
        if (showView.value === "new") {
          const { code, message: msg } = await createDns(curData);
          if (code !== 0) {
            message(msg, { type: "error" });
            return;
          }
          message(`${t("dns.addDns")} ${t("dns.success", "成功")}`, {
            type: "success"
          });
        } else {
          const { code, message: msg } = await updateDns(curData);
          if (code !== 0) {
            message(msg, { type: "error" });
            return;
          }
          message(`${t("dns.editDns")} ${t("dns.success", "成功")}`, {
            type: "success"
          });
        }
        showView.value = "list";
        onSearch();
      } catch (e: any) {
        message(e.message || t("dns.submitFailed", "提交失败"), { type: "error" });
      } finally {
        saving.value = false;
      }
    }
  });
}
</script>

<template>
  <div class="main">
    <!-- List View Mode -->
    <div v-if="showView === 'list'">
      <!-- 顶部的多条件搜索表单 -->
      <el-form
        ref="searchFormRef"
        :inline="true"
        :model="form"
        class="search-form bg-bg_color w-full px-3 sm:px-6 pt-3 pb-1 overflow-auto mb-3 rounded-xl border border-(--el-border-color-lighter) shadow-2xs"
      >
        <el-form-item :label="t('dns.port')" prop="port">
          <el-input
            v-model="form.port"
            :placeholder="t('dns.searchPortPlaceholder')"
            clearable
            class="w-full sm:w-45!"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
        </el-form-item>

        <el-form-item :label="t('dns.address')" prop="address">
          <el-input
            v-model="form.address"
            :placeholder="t('dns.searchAddressPlaceholder')"
            clearable
            class="w-full sm:w-45!"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :icon="useRenderIcon('ri:search-line')"
            :loading="loading"
            @click="onSearch"
          >
            {{ t("dns.search") }}
          </el-button>
          <el-button
            :icon="useRenderIcon('ri:refresh-line')"
            @click="resetForm(searchFormRef)"
          >
            {{ t("dns.reset") }}
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 表格及操作栏 -->
      <PureTableBar
        :title="t('menus.pureDns')"
        :columns="columns"
        @refresh="onSearch"
      >
        <template #buttons>
          <el-button
            type="primary"
            :icon="useRenderIcon(AddFill)"
            @click="handleAddPage"
          >
            {{ t("buttons.pureAdd", "添加") }}
          </el-button>
        </template>
        <template v-slot="{ size, dynamicColumns }">
          <div
            v-if="selectedNum > 0"
            class="bg-(--el-color-primary-light-9) text-(--el-color-primary) border border-(--el-color-primary-light-7) px-4 py-2 rounded-lg text-sm mb-3 flex-bc"
          >
            <span
              >{{ t("dns.selected") }} {{ selectedNum }}
              {{ t("dns.items") }}</span
            >
            <div>
              <el-button
                type="primary"
                link
                size="small"
                @click="onSelectionCancel"
              >
                {{ t("dns.cancelSelection") }}
              </el-button>
              <el-popconfirm
                :title="t('dns.confirmDelete')"
                @confirm="onbatchDel"
              >
                <template #reference>
                  <el-button type="danger" link size="small">
                    {{ t("dns.batchDelete") }}
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
            <!-- 1. Hosts 域名映射列 -->
            <template #hosts="{ row }">
              <el-popover placement="top" :width="360" trigger="hover">
                <template #reference>
                  <div class="inline-flex items-center gap-1.5 cursor-pointer">
                    <template
                      v-if="
                        parseHostsJSON(row.hosts_json || row.HostsJSON).aCount >
                          0 ||
                        parseHostsJSON(row.hosts_json || row.HostsJSON)
                          .aaaaCount > 0
                      "
                    >
                      <el-tag
                        size="small"
                        type="primary"
                        effect="light"
                        class="font-mono font-bold"
                      >
                        A:
                        {{
                          parseHostsJSON(row.hosts_json || row.HostsJSON).aCount
                        }}
                      </el-tag>
                      <el-tag
                        size="small"
                        type="success"
                        effect="light"
                        class="font-mono font-bold"
                      >
                        AAAA:
                        {{
                          parseHostsJSON(row.hosts_json || row.HostsJSON)
                            .aaaaCount
                        }}
                      </el-tag>
                    </template>
                    <el-tag
                      v-else
                      type="info"
                      size="small"
                      effect="plain"
                      class="text-gray-400"
                    >
                      {{ t("dns.noHostsConfig") }}
                    </el-tag>
                  </div>
                </template>

                <!-- Popover 浮层内容 -->
                <div class="p-1 text-xs">
                  <div class="font-bold border-b pb-1 mb-2 flex-bc">
                    <span class="inline-flex items-center gap-1">
                      <IconifyIconOffline icon="ri:file-list-3-line" />
                      {{ t("dns.hostsList") }}
                    </span>
                    <span class="text-gray-400 font-mono"
                      >ID: {{ row.Id || row.id }}</span
                    >
                  </div>

                  <div
                    v-if="
                      parseHostsJSON(row.hosts_json || row.HostsJSON).aCount ===
                        0 &&
                      parseHostsJSON(row.hosts_json || row.HostsJSON)
                        .aaaaCount === 0
                    "
                    class="text-gray-400 text-center py-2"
                  >
                    {{ t("dns.noHostsConfig") }}
                  </div>

                  <div v-else class="space-y-2 max-h-60 overflow-auto">
                    <!-- A 记录 (IPv4) -->
                    <div
                      v-if="
                        parseHostsJSON(row.hosts_json || row.HostsJSON).aList
                          .length > 0
                      "
                      class="space-y-1"
                    >
                      <div
                        class="font-semibold text-blue-600 dark:text-blue-400 text-[11px]"
                      >
                        {{ t("dns.aRecord") }}
                      </div>
                      <div
                        v-for="item in parseHostsJSON(
                          row.hosts_json || row.HostsJSON
                        ).aList"
                        :key="item.ip"
                        class="p-1.5 bg-(--el-fill-color-light) rounded font-mono flex-bc gap-2 border border-(--el-border-color-lighter) text-xs overflow-hidden"
                      >
                        <span
                          class="font-semibold text-blue-600 dark:text-blue-400 shrink-0"
                          >{{ item.ip }}</span
                        >
                        <span
                          class="text-(--el-text-color-primary) font-medium truncate text-right"
                          :title="item.domains.join(' ')"
                        >
                          {{ formatDomains(item.domains) }}
                        </span>
                      </div>
                    </div>

                    <!-- AAAA 记录 (IPv6) -->
                    <div
                      v-if="
                        parseHostsJSON(row.hosts_json || row.HostsJSON).aaaaList
                          .length > 0
                      "
                      class="space-y-1"
                    >
                      <div
                        class="font-semibold text-emerald-600 dark:text-emerald-400 text-[11px]"
                      >
                        {{ t("dns.aaaaRecord") }}
                      </div>
                      <div
                        v-for="item in parseHostsJSON(
                          row.hosts_json || row.HostsJSON
                        ).aaaaList"
                        :key="item.ip"
                        class="p-1.5 bg-(--el-fill-color-light) rounded font-mono flex-bc gap-2 border border-(--el-border-color-lighter) text-xs overflow-hidden"
                      >
                        <span
                          class="font-semibold text-emerald-600 dark:text-emerald-400 shrink-0"
                          >{{ item.ip }}</span
                        >
                        <span
                          class="text-(--el-text-color-primary) font-medium truncate text-right"
                          :title="item.domains.join(' ')"
                        >
                          {{ formatDomains(item.domains) }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </el-popover>
            </template>

            <!-- 2. 上游 (Backend) 列 -->
            <template #backend="{ row }">
              <el-popover placement="top" :width="380" trigger="hover">
                <template #reference>
                  <div
                    class="inline-flex items-center gap-1.5 flex-wrap cursor-pointer"
                  >
                    <el-tag
                      v-if="row.tunnel_id || row.TunnelId"
                      type="primary"
                      effect="light"
                      size="small"
                      class="font-mono font-medium inline-flex items-center gap-1"
                    >
                      <IconifyIconOffline icon="ri:route-line" />
                      Tunnel #{{ row.tunnel_id || row.TunnelId }} ({{
                        (
                          row.tunnel_type ||
                          row.TunnelType ||
                          "quic"
                        ).toUpperCase()
                      }})
                    </el-tag>

                    <el-tag
                      v-if="
                        parseUpstreamServers(
                          row.upstream_servers || row.UpstreamServers
                        ).length > 0
                      "
                      type="success"
                      effect="light"
                      size="small"
                      class="font-mono font-medium inline-flex items-center gap-1"
                    >
                      <IconifyIconOffline icon="ri:global-line" />
                      Upstream ({{
                        parseUpstreamServers(
                          row.upstream_servers || row.UpstreamServers
                        ).length
                      }}
                      /
                      {{
                        row.upstream_method ||
                        row.UpstreamMethod ||
                        "round_robin"
                      }})
                    </el-tag>

                    <el-tag
                      v-if="
                        !(row.tunnel_id || row.TunnelId) &&
                        parseUpstreamServers(
                          row.upstream_servers || row.UpstreamServers
                        ).length === 0
                      "
                      type="info"
                      size="small"
                      effect="plain"
                      class="text-gray-400"
                    >
                      {{ t("dns.noUpstream") }}
                    </el-tag>
                  </div>
                </template>

                <!-- Popover 浮层内容 -->
                <div class="p-1 text-xs">
                  <div class="font-bold border-b pb-1 mb-2 flex-bc">
                    <span class="inline-flex items-center gap-1">
                      <IconifyIconOffline icon="ri:global-line" />
                      {{ t("dns.upstreamConfig") }}
                    </span>
                    <span class="text-gray-400 font-mono"
                      >ID: {{ row.Id || row.id }}</span
                    >
                  </div>

                  <div
                    v-if="
                      !(row.tunnel_id || row.TunnelId) &&
                      parseUpstreamServers(
                        row.upstream_servers || row.UpstreamServers
                      ).length === 0
                    "
                    class="text-gray-400 text-center py-2"
                  >
                    {{ t("dns.noUpstreamConfig") }}
                  </div>

                  <div v-else class="space-y-2.5 max-h-60 overflow-auto">
                    <div
                      v-if="row.tunnel_id || row.TunnelId"
                      class="p-2 bg-blue-50/60 dark:bg-blue-900/20 rounded border border-blue-100 dark:border-blue-800/40"
                    >
                      <div
                        class="font-semibold text-blue-600 dark:text-blue-400 text-[11px] mb-1 inline-flex items-center gap-1"
                      >
                        <IconifyIconOffline icon="ri:route-line" />
                        {{ t("dns.tunnelConfig") }}
                      </div>
                      <div
                        class="font-mono text-gray-700 dark:text-gray-300 space-y-0.5"
                      >
                        <div>
                          <span class="text-gray-400">Tunnel ID:</span> #{{
                            row.tunnel_id || row.TunnelId
                          }}
                        </div>
                        <div>
                          <span class="text-gray-400">Type:</span>
                          {{
                            (
                              row.tunnel_type ||
                              row.TunnelType ||
                              "quic"
                            ).toUpperCase()
                          }}
                        </div>
                      </div>
                    </div>

                    <div
                      v-if="
                        parseUpstreamServers(
                          row.upstream_servers || row.UpstreamServers
                        ).length > 0
                      "
                      class="space-y-1"
                    >
                      <div
                        class="font-semibold text-emerald-600 dark:text-emerald-400 text-[11px] flex-bc"
                      >
                        <span class="inline-flex items-center gap-1">
                          <IconifyIconOffline icon="ri:server-line" />
                          {{ t("dns.upstreamServersTitle") }} ({{
                            t("dns.strategy")
                          }}:
                          {{
                            row.upstream_method ||
                            row.UpstreamMethod ||
                            "round_robin"
                          }})
                        </span>
                        <span class="text-gray-400 font-mono"
                          >Total:
                          {{
                            parseUpstreamServers(
                              row.upstream_servers || row.UpstreamServers
                            ).length
                          }}</span
                        >
                      </div>
                      <div
                        v-for="(s, idx) in parseUpstreamServers(
                          row.upstream_servers || row.UpstreamServers
                        )"
                        :key="idx"
                        class="p-1.5 bg-(--el-fill-color-light) rounded font-mono flex-bc border border-(--el-border-color-lighter)"
                      >
                        <span
                          class="text-blue-600 dark:text-blue-400 font-medium truncate mr-2"
                          >{{ s.target }}</span
                        >
                        <el-tag
                          size="small"
                          type="info"
                          effect="plain"
                          class="font-mono scale-90 origin-right"
                        >
                          W: {{ s.weight }}
                        </el-tag>
                      </div>
                    </div>
                  </div>
                </div>
              </el-popover>
            </template>

            <!-- 3. 折叠展开明细行 -->
            <template #expand="{ row }">
              <div
                class="p-3 sm:p-4 bg-(--el-fill-color-light) rounded-xl m-1 sm:my-2 sm:mx-4 border border-(--el-border-color-lighter) space-y-3 text-xs"
              >
                <div
                  v-if="
                    row.hosts_text ||
                    parseHostsJSON(row.hosts_json || row.HostsJSON).aCount >
                      0 ||
                    parseHostsJSON(row.hosts_json || row.HostsJSON).aaaaCount >
                      0
                  "
                  class="space-y-1.5"
                >
                  <div
                    class="font-bold text-(--el-text-color-primary) flex-bc flex-wrap gap-1"
                  >
                    <span class="inline-flex items-center gap-1">
                      <IconifyIconOffline icon="ri:file-text-line" />
                      {{ t("dns.hostsTab") }}:
                    </span>
                    <span class="text-gray-400 font-mono">
                      A:
                      {{
                        parseHostsJSON(row.hosts_json || row.HostsJSON).aCount
                      }}
                      / AAAA:
                      {{
                        parseHostsJSON(row.hosts_json || row.HostsJSON)
                          .aaaaCount
                      }}
                    </span>
                  </div>
                  <pre
                    v-if="row.hosts_text"
                    class="p-3 bg-gray-900 text-gray-100 rounded-lg text-xs/relaxed font-mono overflow-auto max-h-40"
                    >{{ row.hosts_text }}</pre>
                  <div v-else class="text-gray-400 text-xs">
                    {{ t("dns.noHostsConfig") }}
                  </div>
                </div>

                <div
                  v-if="
                    parseUpstreamServers(
                      row.upstream_servers || row.UpstreamServers
                    ).length > 0
                  "
                  class="space-y-1.5"
                >
                  <div
                    class="font-bold text-(--el-text-color-primary) flex-bc flex-wrap gap-1"
                  >
                    <span class="inline-flex items-center gap-1">
                      <IconifyIconOffline icon="ri:global-line" />
                      {{ t("dns.upstreamServersTitle") }} ({{
                        t("dns.strategy")
                      }}:
                      {{
                        row.upstream_method ||
                        row.UpstreamMethod ||
                        "round_robin"
                      }})
                    </span>
                    <span class="text-gray-400 font-mono"
                      >Total:
                      {{
                        parseUpstreamServers(
                          row.upstream_servers || row.UpstreamServers
                        ).length
                      }}</span
                    >
                  </div>
                  <el-table
                    :data="
                      parseUpstreamServers(
                        row.upstream_servers || row.UpstreamServers
                      )
                    "
                    border
                    size="small"
                    class="w-full"
                    :header-cell-style="{
                      background: 'var(--el-fill-color)',
                      color: 'var(--el-text-color-primary)'
                    }"
                  >
                    <el-table-column label="#" width="50" align="center">
                      <template #default="{ $index }">
                        <span class="font-mono text-gray-400">{{
                          $index + 1
                        }}</span>
                      </template>
                    </el-table-column>
                    <el-table-column
                      prop="target"
                      label="服务器 Target (IP/Port 或 DoH)"
                      min-width="180"
                    >
                      <template #default="{ row: s }">
                        <span
                          class="font-mono font-semibold text-blue-600 dark:text-blue-400"
                        >
                          {{ s.target }}
                        </span>
                      </template>
                    </el-table-column>
                    <el-table-column
                      prop="weight"
                      label="权重 Weight"
                      width="90"
                      align="center"
                    >
                      <template #default="{ row: s }">
                        <el-tag
                          size="small"
                          type="info"
                          effect="plain"
                          class="font-mono"
                        >
                          {{ s.weight }}
                        </el-tag>
                      </template>
                    </el-table-column>
                  </el-table>
                </div>
              </div>
            </template>

            <!-- 操作列 -->
            <template #operation="{ row }">
              <div class="flex-c space-x-1">
                <el-button
                  class="reset-margin"
                  link
                  type="primary"
                  :size="size"
                  :icon="useRenderIcon(EditPen)"
                  @click="handleEditPage(row)"
                >
                  {{ t("dns.edit") }}
                </el-button>
                <el-popconfirm
                  :title="t('dns.confirmDelete')"
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
                      {{ t("dns.delete") }}
                    </el-button>
                  </template>
                </el-popconfirm>
              </div>
            </template>
          </pure-table>
        </template>
      </PureTableBar>
    </div>

    <!-- Create / Edit Full Page View Mode -->
    <div
      v-else-if="showView === 'new' || showView === 'edit'"
      class="p-3 sm:p-5 bg-bg_color rounded-xl border border-(--el-border-color-lighter) shadow-2xs"
    >
      <!-- Full Page Header Bar -->
      <PageHeader
        :title="showView === 'new' ? t('dns.addDns') : t('dns.editDns') + ' (id: ' + (formInline.id || 'new') + ')'"
        :description="t('dns.headerDesc', '配置 DNS 代理监听端口、传输层规则过滤 (Rule)、本地 Hosts 静态解析与 Backend 上游服务器')"
        :backTitle="t('dns.backToList', '返回 DNS 列表')"
        @back="handleCancelPage"
      >
        <template #actions>
          <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
            {{ t("dns.cancel", "取消") }}
          </el-button>
          <el-button
            type="primary"
            :loading="saving"
            :icon="useRenderIcon(CheckIcon)"
            @click="handleSaveSubmit"
          >
            {{ t("dns.save", "保存") }}
          </el-button>
        </template>
      </PageHeader>

      <!-- Form Component Embedded Directly -->
      <editForm ref="createEditFormRef" :formInline="formInline" />

      <!-- Bottom Action Bar -->
      <div
        class="flex items-center justify-end space-x-3 pt-4 mt-4 border-t border-(--el-border-color-lighter)"
      >
        <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
          {{ t("dns.cancel", "取消") }}
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          :icon="useRenderIcon(CheckIcon)"
          @click="handleSaveSubmit"
        >
          {{ t("dns.save", "保存") }}
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
