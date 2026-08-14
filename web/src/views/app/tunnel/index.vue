<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useTunnel } from "./utils/hook";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import RefreshLine from "~icons/ri/refresh-line";

defineOptions({
  name: "AppTunnel"
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
  refreshingRowId,
  refreshSingleTunnel,
  onSearch,
  resetForm,
  onbatchDel,
  openDialog,
  openClientDialog,

  handleDelete,
  handleDeleteClient,
  handleSizeChange,
  onSelectionCancel,
  handleCurrentChange,
  handleSelectionChange
} = useTunnel(t, tableRef);

function toggleRowExpand(row: any, forceExpand?: boolean) {
  if (tableRef.value?.getTableRef) {
    const rawTable = tableRef.value.getTableRef();
    if (rawTable && typeof rawTable.toggleRowExpansion === "function") {
      rawTable.toggleRowExpansion(row, forceExpand);
    }
  }
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
        label-width="80px"
        :size="deviceDetection() ? 'small' : 'default'"
        class="search-form bg-bg_color w-full pl-4 md:pl-8 pt-3 overflow-auto"
      >
        <el-form-item :label="t('tunnel.type')" prop="type">
          <el-select
            v-model="form.type"
            :placeholder="t('tunnel.searchTypePlaceholder')"
            clearable
            class="w-45!"
            @change="onSearch"
          >
            <el-option :label="t('tunnel.allTypes')" value="" />
            <el-option label="TLS-TUNNEL" value="TLS-TUNNEL" />
            <el-option label="QUIC-TUNNEL" value="QUIC-TUNNEL" />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('tunnel.port')" prop="port">
          <el-input
            v-model="form.port"
            :placeholder="t('tunnel.searchPortPlaceholder')"
            clearable
            class="w-45!"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
        </el-form-item>

        <el-form-item :label="t('tunnel.sni')" prop="sni">
          <el-input
            v-model="form.sni"
            :placeholder="t('tunnel.searchSniPlaceholder')"
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
            {{ t('tunnel.search') }}
          </el-button>
          <el-button
            :icon="useRenderIcon('ri/refresh-line')"
            @click="resetForm(formRef)"
          >
            {{ t('tunnel.reset') }}
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 表格及操作栏 -->
      <PureTableBar
        :title="t('menus.pureTunnel')"
        :columns="columns"
        @refresh="onSearch"
      >
        <template #buttons>
          <el-button
            type="primary"
            :icon="useRenderIcon(AddFill)"
            @click="openDialog(t('tunnel.addTunnel'))"
          >
            {{ t('tunnel.addTunnel') }}
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
                {{ t('tunnel.selected') }} {{ selectedNum }} 项
              </span>
              <el-button type="primary" text @click="onSelectionCancel">
                {{ t('tunnel.cancelSelection') }}
              </el-button>
            </div>
            <el-popconfirm :title="t('tunnel.confirmDelete')" @confirm="onbatchDel">
              <template #reference>
                <el-button type="danger" text class="mr-1!">
                  {{ t('tunnel.batchDelete') }}
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
            <!-- 鼠标悬浮查看关联节点 Popover 列 (点击节点或 Popover 可直接自动展开节点列表) -->
            <template #clientNodes="{ row }">
              <el-popover placement="top" :width="380" trigger="hover">
                <template #reference>
                  <el-tooltip :content="t('tunnel.clickToExpand')" placement="top">
                    <div
                      class="inline-flex items-center gap-1.5 cursor-pointer flex-wrap hover:opacity-80 transition-opacity"
                      @click="toggleRowExpand(row)"
                    >
                      <el-tag
                        v-if="row.unsaved_count > 0"
                        type="warning"
                        effect="light"
                        class="font-medium animate-pulse border-amber-300"
                      >
                        ⚡ {{ row.unsaved_count }} {{ t('tunnel.unsavedCount') }}
                      </el-tag>
                      <el-tag
                        v-if="(row.online_count - row.unsaved_count) > 0"
                        type="success"
                        effect="light"
                        class="font-medium"
                      >
                        🟢 {{ row.online_count - row.unsaved_count }} {{ t('tunnel.onlineCount') }} / {{ row.total_count - row.unsaved_count }} {{ t('tunnel.totalNodes') }}
                      </el-tag>
                      <el-tag
                        v-else-if="row.total_count > 0 && row.unsaved_count === 0"
                        type="info"
                        effect="light"
                        class="font-medium text-gray-500"
                      >
                        🔴 0 {{ t('tunnel.onlineCount') }} / {{ row.total_count }} {{ t('tunnel.totalNodes') }}
                      </el-tag>
                      <el-tag v-else-if="row.total_count === 0 && row.unsaved_count === 0" type="info" effect="plain" class="text-gray-400">
                        ⚪ {{ t('tunnel.noNodes') }}
                      </el-tag>
                    </div>
                  </el-tooltip>
                </template>

                <!-- Popover 浮层内容 (点击也可展开下侧节点明细) -->
                <div class="p-1 cursor-pointer" @click="toggleRowExpand(row, true)">
                  <div class="font-bold text-sm mb-2 border-b pb-1.5 flex justify-between items-center">
                    <span class="flex items-center gap-1">🔗 {{ t('tunnel.nodeDetail') }}</span>
                    <span class="text-xs text-blue-600 dark:text-blue-400 font-normal hover:underline">
                      {{ t('tunnel.clickToExpand') }} ⬇
                    </span>
                  </div>

                  <div v-if="row.unsaved_count > 0" class="mb-2 p-1.5 bg-amber-50 dark:bg-amber-950/30 border border-amber-200/80 dark:border-amber-800/50 rounded text-amber-700 dark:text-amber-300 text-[11px] leading-tight">
                    {{ t('tunnel.unsavedBannerTip', { count: row.unsaved_count }) }}
                  </div>

                  <div
                    v-if="!row.client_nodes || row.client_nodes.length === 0"
                    class="text-gray-400 text-xs py-3 text-center"
                  >
                    {{ t('tunnel.noNodes') }}
                  </div>
                  <div v-else class="space-y-2 max-h-60 overflow-auto pr-1">
                    <div
                      v-for="c in row.client_nodes"
                      :key="c.token"
                      :class="[
                        'p-2 rounded border text-xs',
                        !c.is_saved
                          ? 'bg-amber-50/50 dark:bg-amber-950/20 border-amber-200 dark:border-amber-800/60'
                          : 'bg-gray-50 dark:bg-gray-700/50 border-gray-100 dark:border-gray-700'
                      ]"
                    >
                      <div class="flex justify-between items-center font-medium">
                        <span :class="!c.is_saved ? 'text-amber-800 dark:text-amber-300 font-semibold' : 'text-gray-800 dark:text-gray-200'">
                          {{ c.name || (c.is_saved ? '-' : t('tunnelClient.unsavedNode')) }}
                        </span>
                        <div class="flex items-center space-x-1">
                          <el-tag v-if="!c.is_saved" type="warning" size="small" effect="light">
                            ⚡ {{ t('tunnel.unsavedOnline') }}
                          </el-tag>
                          <el-tag v-else-if="c.is_online" type="success" size="small" effect="plain">
                            🟢 {{ t('tunnelClient.online') }}
                          </el-tag>
                          <el-tag v-else type="info" size="small" effect="plain" class="text-gray-400">
                            🔴 {{ t('tunnelClient.offline') }}
                          </el-tag>
                        </div>
                      </div>
                      <div class="text-gray-500 dark:text-gray-400 font-mono mt-1 flex justify-between items-center">
                        <span>Token: <span class="font-mono text-gray-800 dark:text-gray-200">{{ c.token }}</span></span>
                        <el-button
                          v-if="!c.is_saved"
                          type="warning"
                          size="small"
                          link
                          class="p-0! text-[11px]"
                          @click.stop="openClientDialog(t('tunnelClient.addClient'), row, c)"
                        >
                          + {{ t('tunnelClient.saveAsNode') }}
                        </el-button>
                      </div>
                      <div v-if="c.is_online" class="text-emerald-600 dark:emerald-400 font-mono mt-0.5">
                        Remote: {{ c.remote_addr }}
                      </div>
                    </div>
                  </div>
                </div>
              </el-popover>
            </template>

            <!-- 折叠展开行：下侧展示该 Tunnel 关联的隧道节点表格 -->
            <template #expand="{ row }">
              <div class="p-3 sm:p-4 bg-gray-50/80 dark:bg-gray-800/60 rounded-md my-2 mx-2 sm:mx-4 border border-gray-200/60 dark:border-gray-700">
                <!-- 顶部未保存节点提醒 Banner -->
                <div
                  v-if="(row.unsaved_count || 0) > 0"
                  class="mb-3 p-2 bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800/60 rounded-md flex items-center justify-between text-xs text-amber-800 dark:text-amber-200 flex-wrap gap-2"
                >
                  <div class="flex items-center space-x-1.5 font-medium">
                    <span class="text-base leading-none">⚡</span>
                    <span>{{ t('tunnel.unsavedBannerTip', { count: row.unsaved_count }) }}</span>
                  </div>
                </div>

                <div class="flex flex-wrap sm:flex-nowrap justify-between items-center gap-2 mb-3">
                  <div class="flex items-center space-x-2 flex-wrap gap-y-1">
                    <span class="font-bold text-sm text-gray-800 dark:text-gray-200">
                      🔗 {{ t('tunnel.clientNodes') }}
                    </span>
                    <el-tag
                      v-if="(row.unsaved_count || 0) > 0"
                      size="small"
                      type="warning"
                      effect="light"
                      class="font-medium animate-pulse"
                    >
                      ⚡ {{ row.unsaved_count }} {{ t('tunnel.unsavedCount') }}
                    </el-tag>
                    <el-tag
                      size="small"
                      :type="((row.online_count || 0) - (row.unsaved_count || 0)) > 0 ? 'success' : 'info'"
                      effect="light"
                      class="font-medium"
                    >
                      🟢 {{ (row.online_count || 0) - (row.unsaved_count || 0) }} {{ t('tunnel.onlineCount') }} / {{ (row.total_count || 0) - (row.unsaved_count || 0) }} {{ t('tunnel.totalNodes') }}
                    </el-tag>
                  </div>
                  <div class="flex items-center space-x-2">
                    <el-button
                      type="primary"
                      size="small"
                      :icon="useRenderIcon(AddFill)"
                      @click="openClientDialog(t('tunnelClient.addClient'), row)"
                    >
                      {{ t('tunnel.addNode') }}
                    </el-button>

                    <el-tooltip :content="t('tunnelClient.refreshNodes')" placement="top">
                      <el-button
                        type="primary"
                        link
                        class="px-1! cursor-pointer"
                        :icon="useRenderIcon(RefreshLine)"
                        :loading="refreshingRowId === (row.Id || row.id)"
                        @click="refreshSingleTunnel(row)"
                      />
                    </el-tooltip>
                  </div>
                </div>

                <!-- 隧道节点嵌套表格 -->
                <el-table
                  :data="row.client_nodes || []"
                  border
                  size="small"
                  class="w-full"
                  :header-cell-style="{ background: 'var(--el-fill-color)', color: 'var(--el-text-color-primary)' }"
                >
                  <el-table-column label="ID" width="90" align="center">
                    <template #default="{ row: client }">
                      <el-tag v-if="client.is_saved && (client.id || client.Id)" type="info" size="small" effect="plain" class="font-mono">
                        #{{ client.id || client.Id }}
                      </el-tag>
                      <el-tag v-else type="warning" size="small" effect="light" class="font-mono">
                        {{ t('tunnelClient.unsavedId') }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="name" :label="t('tunnelClient.name')" min-width="140">
                    <template #default="{ row: client }">
                      <span v-if="client.is_saved" class="font-medium text-gray-800 dark:text-gray-200">
                        {{ client.name }}
                      </span>
                      <span v-else class="font-semibold text-amber-600 dark:text-amber-400 flex items-center gap-1">
                        <span class="text-xs">⚡</span> {{ client.name || t('tunnelClient.unsavedNode') }}
                      </span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="token" label="Token" min-width="120">
                    <template #default="{ row: client }">
                      <span class="font-mono text-blue-600 dark:text-blue-400">
                        {{ client.token }}
                      </span>
                    </template>
                  </el-table-column>
                  <el-table-column :label="t('tunnelClient.status')" min-width="180">
                    <template #default="{ row: client }">
                      <el-tag v-if="!client.is_saved" type="warning" size="small" effect="light" class="font-medium border-amber-300">
                        ⚡ {{ t('tunnel.unsavedOnline') }} ({{ client.remote_addr }})
                      </el-tag>
                      <el-tag v-else-if="client.is_online" type="success" size="small" effect="light" class="font-medium">
                        🟢 {{ t('tunnelClient.online') }} ({{ client.remote_addr }})
                      </el-tag>
                      <el-tag v-else type="info" size="small" effect="plain" class="text-gray-400">
                        🔴 {{ t('tunnelClient.offline') }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column :label="t('tunnelClient.remark')" min-width="160">
                    <template #default="{ row: client }">
                      <span class="text-gray-500 text-xs">
                        {{ client.remark || '-' }}
                      </span>
                    </template>
                  </el-table-column>
                  <el-table-column :label="t('tunnel.operation')" min-width="150" align="center" fixed="right">
                    <template #default="{ row: client }">
                      <div class="flex items-center justify-center space-x-1 whitespace-nowrap">
                        <el-button
                          v-if="!client.is_saved"
                          type="warning"
                          size="small"
                          :icon="useRenderIcon(AddFill)"
                          @click="openClientDialog(t('tunnelClient.addClient'), row, client)"
                        >
                          {{ t('tunnelClient.saveAsNode') }}
                        </el-button>
                        <template v-else>
                          <el-button
                            link
                            type="primary"
                            size="small"
                            :icon="useRenderIcon(EditPen)"
                            @click="openClientDialog(t('tunnelClient.editClient'), row, client)"
                          >
                            {{ t('tunnel.edit') }}
                          </el-button>
                          <el-popconfirm
                            :title="t('tunnel.confirmDelete')"
                            @confirm="handleDeleteClient(client)"
                          >
                            <template #reference>
                              <el-button
                                link
                                type="danger"
                                size="small"
                                :icon="useRenderIcon(Delete)"
                              >
                                {{ t('tunnel.delete') }}
                              </el-button>
                            </template>
                          </el-popconfirm>
                        </template>
                      </div>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </template>

            <!-- 操作列 -->
            <template #operation="{ row }">
              <el-button
                class="reset-margin"
                link
                type="primary"
                :size="size"
                :icon="useRenderIcon(EditPen)"
                @click="openDialog(t('tunnel.editTunnel'), row)"
              >
                {{ t('tunnel.edit') }}
              </el-button>
              <el-popconfirm
                :title="t('tunnel.confirmDelete')"
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
                    {{ t('tunnel.delete') }}
                  </el-button>
                </template>
              </el-popconfirm>
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

.main-content {
  margin: 24px 24px 0 !important;
}

.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}
</style>
