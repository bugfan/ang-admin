<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useDnsProxy } from "./utils/hook";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";

defineOptions({
  name: "AppDns"
});

const { t } = useI18n();
const formRef = ref();
const tableRef = ref();

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
  openDialog,

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

  // Group A (IPv4) domains by IP
  const aGroupMap: Record<string, string[]> = {};
  for (const [domain, ip] of Object.entries(aMap)) {
    if (!aGroupMap[ip]) aGroupMap[ip] = [];
    if (!aGroupMap[ip].includes(domain)) aGroupMap[ip].push(domain);
  }

  // Group AAAA (IPv6) domains by IP
  const aaaaGroupMap: Record<string, string[]> = {};
  for (const [domain, ip] of Object.entries(aaaaMap)) {
    if (!aaaaGroupMap[ip]) aaaaGroupMap[ip] = [];
    if (!aaaaGroupMap[ip].includes(domain)) aaaaGroupMap[ip].push(domain);
  }

  const aList: GroupedHost[] = Object.entries(aGroupMap).map(([ip, domains]) => ({
    ip,
    isIPv6: false,
    domains
  }));

  const aaaaList: GroupedHost[] = Object.entries(aaaaGroupMap).map(([ip, domains]) => ({
    ip,
    isIPv6: true,
    domains
  }));

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
</script>

<template>
  <div :class="['flex', 'justify-between', deviceDetection() && 'flex-wrap']">
    <div class="w-full">
      <!-- 顶部的多条件搜索表单 -->
      <el-form
        ref="formRef"
        :inline="true"
        :model="form"
        label-width="auto"
        :size="deviceDetection() ? 'small' : 'default'"
        class="search-form bg-bg_color w-full pl-4 md:pl-8 pt-3 overflow-auto"
      >
        <el-form-item :label="t('dns.port')" prop="port">
          <el-input
            v-model="form.port"
            :placeholder="t('dns.searchPortPlaceholder')"
            clearable
            class="w-45!"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
        </el-form-item>

        <el-form-item :label="t('dns.address')" prop="address">
          <el-input
            v-model="form.address"
            :placeholder="t('dns.searchAddressPlaceholder')"
            clearable
            class="w-45!"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :icon="useRenderIcon('ri/search-line')"
            :loading="loading"
            @click="onSearch"
          >
            {{ t('dns.search') }}
          </el-button>
          <el-button
            :icon="useRenderIcon('ri/refresh-line')"
            @click="resetForm(formRef)"
          >
            {{ t('dns.reset') }}
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
            @click="openDialog(t('dns.addDns'))"
          >
            {{ t('dns.addDns') }}
          </el-button>
        </template>
        <template v-slot="{ size, dynamicColumns }">
          <div
            v-if="selectedNum > 0"
            v-motion-fade
            class="bg-(--el-fill-color-light) w-full h-11.5 mb-2 pl-4 flex items-center"
          >
            <div class="flex-auto">
              <span
                style="font-size: var(--el-font-size-base)"
                class="text-[rgba(42,46,54,0.5)] dark:text-[rgba(220,220,242,0.5)]"
              >
                {{ t('dns.selected') }} {{ selectedNum }} {{ t('dns.items') }}
              </span>
              <el-button type="primary" text @click="onSelectionCancel">
                {{ t('dns.cancelSelection') }}
              </el-button>
            </div>
            <el-popconfirm :title="t('dns.confirmDelete')" @confirm="onbatchDel">
              <template #reference>
                <el-button type="danger" text class="mr-1!">
                  {{ t('dns.batchDelete') }}
                </el-button>
              </template>
            </el-popconfirm>
          </div>

          <pure-table
            ref="tableRef"
            :row-key="(row) => row.Id || row.id"
            adaptive
            :adaptiveConfig="{ offsetBottom: 108 }"
            align-whole="center"
            table-layout="auto"
            :loading="loading"
            :size="size"
            :data="dataList"
            :columns="dynamicColumns"
            :pagination="{ ...pagination, size }"
            :header-cell-style="{
              background: 'var(--el-fill-color-light)',
              color: 'var(--el-text-color-primary)'
            }"
            @selection-change="handleSelectionChange"
            @page-size-change="handleSizeChange"
            @page-current-change="handleCurrentChange"
          >
            <!-- 1. Hosts 域名映射列 (带 Popover 悬浮明细) -->
            <template #hosts="{ row }">
              <el-popover placement="top" :width="360" trigger="hover">
                <template #reference>
                  <div class="inline-flex items-center gap-1.5 cursor-pointer">
                    <template v-if="parseHostsJSON(row.hosts_json || row.HostsJSON).aCount > 0 || parseHostsJSON(row.hosts_json || row.HostsJSON).aaaaCount > 0">
                      <el-tag size="small" type="primary" effect="light" class="font-mono">
                        A: {{ parseHostsJSON(row.hosts_json || row.HostsJSON).aCount }}
                      </el-tag>
                      <el-tag size="small" type="success" effect="light" class="font-mono">
                        AAAA: {{ parseHostsJSON(row.hosts_json || row.HostsJSON).aaaaCount }}
                      </el-tag>
                    </template>
                    <el-tag v-else type="info" size="small" effect="plain" class="text-gray-400">
                      {{ t('dns.noHostsConfig') }}
                    </el-tag>
                  </div>
                </template>

                <!-- Popover 浮层内容 -->
                <div class="p-1 text-xs">
                  <div class="font-bold border-b pb-1 mb-2 flex justify-between items-center">
                    <span class="inline-flex items-center gap-1">
                      <IconifyIconOffline icon="ri:file-list-3-line" />
                      {{ t('dns.hostsList') }}
                    </span>
                    <span class="text-gray-400 font-mono">ID: {{ row.Id || row.id }}</span>
                  </div>

                  <div
                    v-if="parseHostsJSON(row.hosts_json || row.HostsJSON).aCount === 0 && parseHostsJSON(row.hosts_json || row.HostsJSON).aaaaCount === 0"
                    class="text-gray-400 text-center py-2"
                  >
                    {{ t('dns.noHostsConfig') }}
                  </div>

                  <div v-else class="space-y-2 max-h-60 overflow-auto">
                    <!-- A 记录 (IPv4) -->
                    <div v-if="parseHostsJSON(row.hosts_json || row.HostsJSON).aList.length > 0" class="space-y-1">
                      <div class="font-semibold text-blue-600 dark:text-blue-400 text-[11px]">{{ t('dns.aRecord') }}</div>
                      <div
                        v-for="item in parseHostsJSON(row.hosts_json || row.HostsJSON).aList"
                        :key="item.ip"
                        class="p-1.5 bg-gray-50 dark:bg-gray-700/50 rounded font-mono flex items-center justify-between gap-2 border border-gray-100 dark:border-gray-600 text-xs overflow-hidden"
                      >
                        <span class="font-semibold text-blue-600 dark:text-blue-400 shrink-0">{{ item.ip }}</span>
                        <span class="text-gray-700 dark:text-gray-200 font-medium truncate text-right" :title="item.domains.join(' ')">
                          {{ formatDomains(item.domains) }}
                        </span>
                      </div>
                    </div>

                    <!-- AAAA 记录 (IPv6) -->
                    <div v-if="parseHostsJSON(row.hosts_json || row.HostsJSON).aaaaList.length > 0" class="space-y-1">
                      <div class="font-semibold text-emerald-600 dark:text-emerald-400 text-[11px]">{{ t('dns.aaaaRecord') }}</div>
                      <div
                        v-for="item in parseHostsJSON(row.hosts_json || row.HostsJSON).aaaaList"
                        :key="item.ip"
                        class="p-1.5 bg-gray-50 dark:bg-gray-700/50 rounded font-mono flex items-center justify-between gap-2 border border-gray-100 dark:border-gray-600 text-xs overflow-hidden"
                      >
                        <span class="font-semibold text-emerald-600 dark:text-emerald-400 shrink-0">{{ item.ip }}</span>
                        <span class="text-gray-700 dark:text-gray-200 font-medium truncate text-right" :title="item.domains.join(' ')">
                          {{ formatDomains(item.domains) }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </el-popover>
            </template>

            <!-- 2. 上游 (Backend) 列 (带 Popover 悬浮明细) -->
            <template #backend="{ row }">
              <el-popover placement="top" :width="380" trigger="hover">
                <template #reference>
                  <div class="inline-flex items-center gap-1.5 flex-wrap cursor-pointer">
                    <!-- Tunnel 标签 -->
                    <el-tag
                      v-if="row.tunnel_id || row.TunnelId"
                      type="primary"
                      effect="light"
                      size="small"
                      class="font-mono font-medium inline-flex items-center gap-1"
                    >
                      <IconifyIconOffline icon="ri:route-line" />
                      Tunnel #{{ row.tunnel_id || row.TunnelId }} ({{ (row.tunnel_type || row.TunnelType || 'quic').toUpperCase() }})
                    </el-tag>

                    <!-- Upstream 标签 -->
                    <el-tag
                      v-if="parseUpstreamServers(row.upstream_servers || row.UpstreamServers).length > 0"
                      type="success"
                      effect="light"
                      size="small"
                      class="font-mono font-medium inline-flex items-center gap-1"
                    >
                      <IconifyIconOffline icon="ri:global-line" />
                      Upstream ({{ parseUpstreamServers(row.upstream_servers || row.UpstreamServers).length }} / {{ row.upstream_method || row.UpstreamMethod || 'round_robin' }})
                    </el-tag>

                    <el-tag
                      v-if="!(row.tunnel_id || row.TunnelId) && parseUpstreamServers(row.upstream_servers || row.UpstreamServers).length === 0"
                      type="info"
                      size="small"
                      effect="plain"
                      class="text-gray-400"
                    >
                      {{ t('dns.noUpstream') }}
                    </el-tag>
                  </div>
                </template>

                <!-- Popover 浮层内容 -->
                <div class="p-1 text-xs">
                  <div class="font-bold border-b pb-1 mb-2 flex justify-between items-center">
                    <span class="inline-flex items-center gap-1">
                      <IconifyIconOffline icon="ri:global-line" />
                      {{ t('dns.upstreamConfig') }}
                    </span>
                    <span class="text-gray-400 font-mono">ID: {{ row.Id || row.id }}</span>
                  </div>

                  <div
                    v-if="!(row.tunnel_id || row.TunnelId) && parseUpstreamServers(row.upstream_servers || row.UpstreamServers).length === 0"
                    class="text-gray-400 text-center py-2"
                  >
                    {{ t('dns.noUpstreamConfig') }}
                  </div>

                  <div v-else class="space-y-2.5 max-h-60 overflow-auto">
                    <!-- Tunnel 信息 -->
                    <div v-if="row.tunnel_id || row.TunnelId" class="p-2 bg-blue-50/60 dark:bg-blue-900/20 rounded border border-blue-100 dark:border-blue-800/40">
                      <div class="font-semibold text-blue-600 dark:text-blue-400 text-[11px] mb-1 inline-flex items-center gap-1">
                        <IconifyIconOffline icon="ri:route-line" />
                        {{ t('dns.tunnelConfig') }}
                      </div>
                      <div class="font-mono text-gray-700 dark:text-gray-300 space-y-0.5">
                        <div><span class="text-gray-400">Tunnel ID:</span> #{{ row.tunnel_id || row.TunnelId }}</div>
                        <div><span class="text-gray-400">Type:</span> {{ (row.tunnel_type || row.TunnelType || 'quic').toUpperCase() }}</div>
                      </div>
                    </div>

                    <!-- Upstream 服务器列表 -->
                    <div v-if="parseUpstreamServers(row.upstream_servers || row.UpstreamServers).length > 0" class="space-y-1">
                      <div class="font-semibold text-emerald-600 dark:text-emerald-400 text-[11px] flex justify-between items-center">
                        <span class="inline-flex items-center gap-1">
                          <IconifyIconOffline icon="ri:server-line" />
                          {{ t('dns.upstreamServersTitle') }} ({{ t('dns.strategy') }}: {{ row.upstream_method || row.UpstreamMethod || 'round_robin' }})
                        </span>
                        <span class="text-gray-400 font-mono">Total: {{ parseUpstreamServers(row.upstream_servers || row.UpstreamServers).length }}</span>
                      </div>
                      <div
                        v-for="(s, idx) in parseUpstreamServers(row.upstream_servers || row.UpstreamServers)"
                        :key="idx"
                        class="p-1.5 bg-gray-50 dark:bg-gray-700/50 rounded font-mono flex justify-between items-center border border-gray-100 dark:border-gray-600"
                      >
                        <span class="text-blue-600 dark:text-blue-400 font-medium truncate mr-2">{{ s.target }}</span>
                        <el-tag size="small" type="info" effect="plain" class="font-mono scale-90 origin-right">
                          W: {{ s.weight }}
                        </el-tag>
                      </div>
                    </div>
                  </div>
                </div>
              </el-popover>
            </template>

            <!-- 3. 折叠展开明细行 (Hosts 配置在前，Upstream 服务器在后) -->
            <template #expand="{ row }">
              <div class="p-3 sm:p-4 bg-gray-50/80 dark:bg-gray-800/60 rounded-md my-2 mx-2 sm:mx-4 border border-gray-200/60 dark:border-gray-700 space-y-3 text-xs">
                <!-- Raw Hosts 文本明细 (置于最前) -->
                <div v-if="row.hosts_text || parseHostsJSON(row.hosts_json || row.HostsJSON).aCount > 0 || parseHostsJSON(row.hosts_json || row.HostsJSON).aaaaCount > 0" class="space-y-1.5">
                  <div class="font-bold text-gray-800 dark:text-gray-200 flex items-center justify-between">
                    <span class="inline-flex items-center gap-1">
                      <IconifyIconOffline icon="ri:file-text-line" />
                      {{ t('dns.hostsTab') }}:
                    </span>
                    <span class="text-gray-400 font-mono">
                      A: {{ parseHostsJSON(row.hosts_json || row.HostsJSON).aCount }} / AAAA: {{ parseHostsJSON(row.hosts_json || row.HostsJSON).aaaaCount }}
                    </span>
                  </div>
                  <pre v-if="row.hosts_text" class="p-2 bg-gray-900 text-gray-100 rounded text-xs font-mono overflow-auto max-h-40 leading-relaxed">{{ row.hosts_text }}</pre>
                  <div v-else class="text-gray-400 text-xs">{{ t('dns.noHostsConfig') }}</div>
                </div>

                <!-- Upstream 上游服务器明细 (置于 Hosts 之后) -->
                <div v-if="parseUpstreamServers(row.upstream_servers || row.UpstreamServers).length > 0" class="space-y-1.5">
                  <div class="font-bold text-gray-800 dark:text-gray-200 flex items-center justify-between">
                    <span class="inline-flex items-center gap-1">
                      <IconifyIconOffline icon="ri:global-line" />
                      {{ t('dns.upstreamServersTitle') }} ({{ t('dns.strategy') }}: {{ row.upstream_method || row.UpstreamMethod || 'round_robin' }})
                    </span>
                    <span class="text-gray-400 font-mono">Total: {{ parseUpstreamServers(row.upstream_servers || row.UpstreamServers).length }}</span>
                  </div>
                  <el-table
                    :data="parseUpstreamServers(row.upstream_servers || row.UpstreamServers)"
                    border
                    size="small"
                    class="w-full"
                    :header-cell-style="{ background: 'var(--el-fill-color)', color: 'var(--el-text-color-primary)' }"
                  >
                    <el-table-column label="#" width="50" align="center">
                      <template #default="{ $index }">
                        <span class="font-mono text-gray-400">{{ $index + 1 }}</span>
                      </template>
                    </el-table-column>
                    <el-table-column prop="target" label="服务器 Target (IP/Port 或 DoH)" min-width="240">
                      <template #default="{ row: s }">
                        <span class="font-mono font-semibold text-blue-600 dark:text-blue-400">
                          {{ s.target }}
                        </span>
                      </template>
                    </el-table-column>
                    <el-table-column prop="weight" label="权重 Weight" width="100" align="center">
                      <template #default="{ row: s }">
                        <el-tag size="small" type="info" effect="plain" class="font-mono">
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
              <div class="flex items-center justify-center whitespace-nowrap space-x-1">
                <el-button
                  class="reset-margin"
                  link
                  type="primary"
                  :size="size"
                  :icon="useRenderIcon(EditPen)"
                  @click="openDialog(t('dns.editDns'), row)"
                >
                  {{ t('dns.edit') }}
                </el-button>
                <el-popconfirm
                  :title="t('dns.confirmDelete')"
                  @confirm="handleDelete(row)"
                >
                  <template #reference>
                    <el-button
                      class="reset-margin"
                      link
                      type="primary"
                      :size="size"
                      :icon="useRenderIcon(Delete)"
                    >
                      {{ t('dns.delete') }}
                    </el-button>
                  </template>
                </el-popconfirm>
              </div>
            </template>
          </pure-table>
        </template>
      </PureTableBar>
    </div>
  </div>
</template>

<style lang="scss" scoped>
:deep(.el-dropdown-menu__item i) {
  margin: 0;
}

:deep(.el-button:focus-visible) {
  outline: none;
}

.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
  :deep(.el-form-item__label) {
    white-space: nowrap;
  }
}
</style>
