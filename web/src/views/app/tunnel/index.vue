<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useTunnel } from "./utils/hook";
import editForm from "./form/index.vue";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import { createTunnel, updateTunnel } from "@/api/tunnel";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import BackIcon from "~icons/ep/back";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";

defineOptions({
  name: "AppTunnel"
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
  refreshingRowId,
  refreshSingleTunnel,
  onSearch,
  resetForm,
  onbatchDel,
  openClientDialog,
  handleDeleteClient,
  handleDelete,
  handleSizeChange,
  onSelectionCancel,
  handleCurrentChange,
  handleSelectionChange
} = useTunnel(t, tableRef);

function getClientNodes(row: any): any[] {
  if (!row) return [];
  if (Array.isArray(row.client_nodes)) return row.client_nodes;
  if (Array.isArray(row.ClientNodes)) return row.ClientNodes;
  if (row.clients_json || row.ClientsJSON) {
    try {
      const parsed = JSON.parse(row.clients_json || row.ClientsJSON);
      if (Array.isArray(parsed)) return parsed;
    } catch (e) {}
  }
  return [];
}

function getDefaultFormInline() {
  return {
    title: t("tunnel.addTunnel"),
    id: undefined,
    name: "",
    type: "TLS-TUNNEL",
    port: "",
    sni: "",
    certificate: "",
    remark: ""
  };
}

function getFormInlineFromRow(row: any) {
  return {
    title: `${t("tunnel.editTunnel")} [ID: ${row?.Id || row?.id}]`,
    id: row?.Id ?? row?.id ?? undefined,
    name: row?.Name ?? row?.name ?? "",
    type: row?.Type ?? row?.type ?? "TLS-TUNNEL",
    port: row?.Port ?? row?.port ?? "",
    sni: row?.SNI ?? row?.sni ?? "",
    certificate: row?.Certificate ?? row?.certificate ?? "",
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
  const FormRef = createEditFormRef.value.getRef();
  if (!FormRef) return;

  FormRef.validate(async (valid: boolean) => {
    if (valid) {
      saving.value = true;
      try {
        const curData = formInline.value;
        if (showView.value === "new") {
          const { code, message: msg } = await createTunnel(curData);
          if (code !== 0) {
            message(msg, { type: "error" });
            return;
          }
          message(`${t("tunnel.addTunnel")} ${t("tunnel.success", "成功")}`, { type: "success" });
        } else {
          const { code, message: msg } = await updateTunnel(curData);
          if (code !== 0) {
            message(msg, { type: "error" });
            return;
          }
          message(`${t("tunnel.editTunnel")} ${t("tunnel.success", "成功")}`, { type: "success" });
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
        class="search-form bg-bg_color w-full px-3 sm:px-6 pt-3 pb-1 overflow-auto mb-3 rounded-xl border border-[var(--el-border-color-lighter)] shadow-2xs"
      >
        <el-form-item :label="t('tunnel.port')" prop="port">
          <el-input
            v-model="form.port"
            :placeholder="t('tunnel.searchPortPlaceholder')"
            clearable
            class="w-full sm:!w-[180px]"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
        </el-form-item>

        <el-form-item :label="t('tunnel.type')" prop="type">
          <el-select
            v-model="form.type"
            :placeholder="t('tunnel.searchTypePlaceholder', '请选择类型')"
            clearable
            class="w-full sm:!w-[180px]"
            @change="onSearch"
          >
            <el-option label="TLS-TUNNEL" value="TLS-TUNNEL" />
            <el-option label="QUIC-TUNNEL" value="QUIC-TUNNEL" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :icon="useRenderIcon('ri:search-line')"
            :loading="loading"
            @click="onSearch"
          >
            {{ t('tunnel.search') }}
          </el-button>
          <el-button
            :icon="useRenderIcon('ri:refresh-line')"
            @click="resetForm(searchFormRef)"
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
            @click="handleAddPage"
          >
            {{ t('tunnel.addTunnel') }}
          </el-button>
        </template>
        <template v-slot="{ size, dynamicColumns }">
          <div
            v-if="selectedNum > 0"
            class="bg-[var(--el-color-primary-light-9)] text-[var(--el-color-primary)] border border-[var(--el-color-primary-light-7)] px-4 py-2 rounded-lg text-sm mb-3 flex items-center justify-between"
          >
            <span>{{ t('tunnel.selected') }} {{ selectedNum }} {{ t('tunnel.items') }}</span>
            <div>
              <el-button type="primary" link size="small" @click="onSelectionCancel">
                {{ t('tunnel.cancelSelection') }}
              </el-button>
              <el-popconfirm :title="t('tunnel.confirmDelete')" @confirm="onbatchDel">
                <template #reference>
                  <el-button type="danger" link size="small">
                    {{ t('tunnel.batchDelete') }}
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
            <!-- 1. 动态挂载客户端节点 (Client Nodes) Popover 浮层列 -->
            <template #clientNodes="{ row }">
              <el-popover placement="top" :width="380" trigger="hover">
                <template #reference>
                  <div class="inline-flex items-center gap-1.5 cursor-pointer">
                    <el-tag
                      v-if="getClientNodes(row).length > 0"
                      type="success"
                      effect="light"
                      size="small"
                      class="font-mono font-bold"
                    >
                      <IconifyIconOffline icon="ri:router-line" class="mr-1" />
                      {{ getClientNodes(row).length }} Node(s)
                    </el-tag>
                    <el-tag v-else type="info" size="small" effect="plain" class="text-gray-400">
                      {{ t('tunnel.noClients') }}
                    </el-tag>
                  </div>
                </template>

                <!-- Popover 浮层内容 -->
                <div class="p-1 text-xs">
                  <div class="font-bold border-b pb-1 mb-2 flex justify-between items-center">
                    <span class="inline-flex items-center gap-1">
                      <IconifyIconOffline icon="ri:router-line" />
                      {{ t('tunnel.clientNodes') }}
                    </span>
                    <span class="text-gray-400 font-mono">Tunnel #{{ row.Id || row.id }}</span>
                  </div>

                  <div
                    v-if="getClientNodes(row).length === 0"
                    class="text-gray-400 text-center py-2"
                  >
                    {{ t('tunnel.noClientsConfig') }}
                  </div>

                  <div v-else class="space-y-1.5 max-h-60 overflow-auto">
                    <div
                      v-for="(c, idx) in getClientNodes(row)"
                      :key="idx"
                      class="p-2 bg-[var(--el-fill-color-light)] rounded font-mono border border-[var(--el-border-color-lighter)] text-xs space-y-1"
                    >
                      <div class="flex justify-between items-center">
                        <span class="font-bold text-blue-600 dark:text-blue-400 cursor-pointer hover:underline" @click="openClientDialog(c.is_saved || c.id ? t('tunnelClient.editClient') : t('tunnelClient.addClient'), row, c)">
                          {{ c.name || c.id || 'Node' }}
                        </span>
                        <div class="flex items-center gap-1.5">
                          <el-tag size="small" :type="c.is_online ? 'success' : 'info'" effect="light">
                            {{ c.is_online ? t('tunnelClient.online') : t('tunnelClient.offline') }}
                          </el-tag>
                          <el-button
                            type="primary"
                            link
                            size="small"
                            class="!p-0 text-xs"
                            @click="openClientDialog(c.is_saved || c.id ? t('tunnelClient.editClient') : t('tunnelClient.addClient'), row, c)"
                          >
                            {{ c.is_saved || c.id ? t('tunnel.edit') : t('tunnelClient.saveAsNode') }}
                          </el-button>
                        </div>
                      </div>
                      <div class="text-[11px] text-gray-500 truncate">Token: {{ c.token || '-' }}</div>
                      <div v-if="c.remote_addr" class="text-[11px] text-gray-400">{{ t('tunnelClient.remoteAddr') }}: {{ c.remote_addr }}</div>
                    </div>
                  </div>
                </div>
              </el-popover>
            </template>

            <!-- 2. 折叠展开明细行 -->
            <template #expand="{ row }">
              <div class="p-3 sm:p-4 bg-[var(--el-fill-color-light)] rounded-xl my-1 sm:my-2 mx-1 sm:mx-4 border border-[var(--el-border-color-lighter)] space-y-3 text-xs">
                <div class="font-bold text-[var(--el-text-color-primary)] flex items-center justify-between flex-wrap gap-2">
                  <div class="flex items-center space-x-2">
                    <IconifyIconOffline icon="ri:router-line" />
                    <span>{{ t('tunnel.clientNodesDetail') }}</span>
                    <el-tag size="small" type="info" effect="plain" class="font-mono">
                      Total: {{ getClientNodes(row).length }}
                    </el-tag>
                  </div>
                  <div class="flex items-center gap-2">
                    <el-button
                      type="primary"
                      size="small"
                      :icon="useRenderIcon(AddFill)"
                      @click="openClientDialog(t('tunnelClient.addClient'), row)"
                    >
                      {{ t('tunnelClient.addClient') }}
                    </el-button>
                    <el-button
                      size="small"
                      :icon="useRenderIcon('ri:refresh-line')"
                      :loading="refreshingRowId === (row.Id || row.id)"
                      @click="refreshSingleTunnel(row)"
                    >
                      {{ t('tunnelClient.refreshNodes') }}
                    </el-button>
                  </div>
                </div>

                <div v-if="getClientNodes(row).length > 0">
                  <el-table
                    :data="getClientNodes(row)"
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
                    <el-table-column prop="name" :label="t('tunnelClient.name')" min-width="140">
                      <template #default="{ row: c }">
                        <span
                          class="font-mono font-semibold text-blue-600 dark:text-blue-400 cursor-pointer hover:underline"
                          @click="openClientDialog(c.is_saved || c.id ? t('tunnelClient.editClient') : t('tunnelClient.addClient'), row, c)"
                        >
                          {{ c.name || c.id || 'Client Node' }}
                        </span>
                      </template>
                    </el-table-column>
                    <el-table-column prop="token" :label="t('tunnelClient.token')" min-width="180">
                      <template #default="{ row: c }">
                        <span class="font-mono text-gray-600 dark:text-gray-300">
                          {{ c.token || '-' }}
                        </span>
                      </template>
                    </el-table-column>
                    <el-table-column prop="remote_addr" :label="t('tunnelClient.remoteAddr')" min-width="150">
                      <template #default="{ row: c }">
                        <span class="font-mono text-gray-500">
                          {{ c.remote_addr || '-' }}
                        </span>
                      </template>
                    </el-table-column>
                    <el-table-column prop="remark" :label="t('tunnel.remark')" min-width="120">
                      <template #default="{ row: c }">
                        <span class="text-gray-500">
                          {{ c.remark || '-' }}
                        </span>
                      </template>
                    </el-table-column>
                    <el-table-column prop="is_online" :label="t('tunnelClient.status')" width="140" align="center">
                      <template #default="{ row: c }">
                        <el-tag size="small" :type="c.is_online ? 'success' : 'info'" effect="light">
                          {{ c.is_online ? t('tunnelClient.online') : t('tunnelClient.offline') }}
                        </el-tag>
                      </template>
                    </el-table-column>
                    <el-table-column :label="t('tunnel.operation')" width="140" align="center" fixed="right">
                      <template #default="{ row: c }">
                        <div class="flex items-center justify-center space-x-2">
                          <el-button
                            v-if="c.is_saved || c.id"
                            link
                            type="primary"
                            size="small"
                            :icon="useRenderIcon(EditPen)"
                            @click="openClientDialog(t('tunnelClient.editClient'), row, c)"
                          >
                            {{ t('tunnel.edit') }}
                          </el-button>
                          <el-button
                            v-else
                            link
                            type="success"
                            size="small"
                            :icon="useRenderIcon(AddFill)"
                            @click="openClientDialog(t('tunnelClient.addClient'), row, c)"
                          >
                            {{ t('tunnelClient.saveAsNode') }}
                          </el-button>

                          <el-popconfirm
                            v-if="c.is_saved || c.id"
                            :title="t('tunnelClient.confirmDeleteNode')"
                            @confirm="handleDeleteClient(c)"
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
                        </div>
                      </template>
                    </el-table-column>
                  </el-table>
                </div>
                <div v-else class="text-gray-400 text-xs py-2 text-center border border-dashed rounded border-gray-300">
                  {{ t('tunnel.noClientsConfig') }}
                </div>
              </div>
            </template>

            <!-- 操作列 -->
            <template #operation="{ row }">
              <div class="flex items-center justify-center space-x-1">
                <el-button
                  class="reset-margin"
                  link
                  type="primary"
                  :size="size"
                  :icon="useRenderIcon(EditPen)"
                  @click="handleEditPage(row)"
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
                      type="danger"
                      :size="size"
                      :icon="useRenderIcon(Delete)"
                    >
                      {{ t('tunnel.delete') }}
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
    <div v-else-if="showView === 'new' || showView === 'edit'" class="p-3 sm:p-5 bg-bg_color rounded-xl border border-[var(--el-border-color-lighter)] shadow-2xs">
      <!-- Full Page Header Bar -->
      <PageHeader
        :title="formInline.title"
        :description="t('tunnel.pageDesc')"
        :backTitle="t('tunnel.backToList')"
        @back="handleCancelPage"
      >
        <template #actions>
          <el-button
            :icon="useRenderIcon(CloseIcon)"
            @click="handleCancelPage"
          >
            {{ t('tunnel.cancel') }}
          </el-button>
          <el-button
            type="primary"
            :loading="saving"
            :icon="useRenderIcon(CheckIcon)"
            @click="handleSaveSubmit"
          >
            {{ t('tunnel.save') }}
          </el-button>
        </template>
      </PageHeader>

      <!-- Form Component Embedded Directly -->
      <editForm ref="createEditFormRef" :formInline="formInline" />

      <!-- Bottom Action Bar -->
      <div class="flex items-center justify-end space-x-3 pt-4 mt-4 border-t border-[var(--el-border-color-lighter)]">
        <el-button
          :icon="useRenderIcon(CloseIcon)"
          @click="handleCancelPage"
        >
          {{ t('tunnel.cancel') }}
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          :icon="useRenderIcon(CheckIcon)"
          @click="handleSaveSubmit"
        >
          {{ t('tunnel.save') }}
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
