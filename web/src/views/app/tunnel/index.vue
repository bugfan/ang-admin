<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useTunnel } from "./utils/hook";
import editForm from "./form/index.vue";
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
  onSearch,
  resetForm,
  onbatchDel,
  handleDelete,
  handleSizeChange,
  onSelectionCancel,
  handleCurrentChange,
  handleSelectionChange
} = useTunnel(t, tableRef);

function parseClientsJSON(clientsJsonStr: string) {
  try {
    if (clientsJsonStr) {
      const parsed = JSON.parse(clientsJsonStr);
      if (Array.isArray(parsed)) return parsed;
    }
  } catch (e) {
    // ignore
  }
  return [];
}

function getDefaultFormInline() {
  return {
    title: t("tunnel.addTunnel"),
    id: undefined,
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
          message(`${t("tunnel.addTunnel")} ${t("tunnel.success")}`, { type: "success" });
        } else {
          const { code, message: msg } = await updateTunnel(curData);
          if (code !== 0) {
            message(msg, { type: "error" });
            return;
          }
          message(`${t("tunnel.editTunnel")} ${t("tunnel.success")}`, { type: "success" });
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
            :placeholder="t('tunnel.selectTypePlaceholder')"
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
            <template #clients="{ row }">
              <el-popover placement="top" :width="380" trigger="hover">
                <template #reference>
                  <div class="inline-flex items-center gap-1.5 cursor-pointer">
                    <el-tag
                      v-if="parseClientsJSON(row.clients_json || row.ClientsJSON).length > 0"
                      type="success"
                      effect="light"
                      size="small"
                      class="font-mono font-bold"
                    >
                      <IconifyIconOffline icon="ri:router-line" class="mr-1" />
                      {{ parseClientsJSON(row.clients_json || row.ClientsJSON).length }} Client(s)
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
                    v-if="parseClientsJSON(row.clients_json || row.ClientsJSON).length === 0"
                    class="text-gray-400 text-center py-2"
                  >
                    {{ t('tunnel.noClientsConfig') }}
                  </div>

                  <div v-else class="space-y-1.5 max-h-60 overflow-auto">
                    <div
                      v-for="(c, idx) in parseClientsJSON(row.clients_json || row.ClientsJSON)"
                      :key="idx"
                      class="p-2 bg-[var(--el-fill-color-light)] rounded font-mono border border-[var(--el-border-color-lighter)] text-xs space-y-1"
                    >
                      <div class="flex justify-between items-center">
                        <span class="font-bold text-blue-600 dark:text-blue-400">{{ c.name || c.id || 'Node' }}</span>
                        <el-tag size="small" :type="c.status === 'online' ? 'success' : 'info'" effect="light">
                          {{ c.status || 'active' }}
                        </el-tag>
                      </div>
                      <div class="text-[11px] text-gray-500 truncate">Token: {{ c.token || '-' }}</div>
                    </div>
                  </div>
                </div>
              </el-popover>
            </template>

            <!-- 2. 折叠展开明细行 -->
            <template #expand="{ row }">
              <div class="p-3 sm:p-4 bg-[var(--el-fill-color-light)] rounded-xl my-1 sm:my-2 mx-1 sm:mx-4 border border-[var(--el-border-color-lighter)] space-y-3 text-xs">
                <div class="font-bold text-[var(--el-text-color-primary)] flex items-center justify-between flex-wrap gap-1">
                  <span class="inline-flex items-center gap-1">
                    <IconifyIconOffline icon="ri:router-line" />
                    {{ t('tunnel.clientNodes') }} 明细列表:
                  </span>
                  <span class="text-gray-400 font-mono">
                    Total Nodes: {{ parseClientsJSON(row.clients_json || row.ClientsJSON).length }}
                  </span>
                </div>

                <div v-if="parseClientsJSON(row.clients_json || row.ClientsJSON).length > 0">
                  <el-table
                    :data="parseClientsJSON(row.clients_json || row.ClientsJSON)"
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
                    <el-table-column prop="name" label="节点名称 Name" min-width="140">
                      <template #default="{ row: c }">
                        <span class="font-mono font-semibold text-blue-600 dark:text-blue-400">
                          {{ c.name || c.id || 'Client Node' }}
                        </span>
                      </template>
                    </el-table-column>
                    <el-table-column prop="token" label="节点鉴权 Token" min-width="180">
                      <template #default="{ row: c }">
                        <span class="font-mono text-gray-600 dark:text-gray-300">
                          {{ c.token || '-' }}
                        </span>
                      </template>
                    </el-table-column>
                    <el-table-column prop="status" label="状态 Status" width="100" align="center">
                      <template #default="{ row: c }">
                        <el-tag size="small" :type="c.status === 'online' ? 'success' : 'info'" effect="light">
                          {{ c.status || 'active' }}
                        </el-tag>
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
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-4 mb-4 border-b border-[var(--el-border-color-lighter)]">
        <div class="flex items-center space-x-3">
          <el-button
            circle
            :icon="useRenderIcon(BackIcon)"
            title="返回隧道列表"
            @click="handleCancelPage"
          />
          <div>
            <h2 class="text-base sm:text-lg font-bold text-[var(--el-text-color-primary)]">
              {{ formInline.title }}
            </h2>
            <div class="text-xs text-[var(--el-text-color-secondary)] mt-0.5">
              配置内网穿透 Tunnel 隧道监听端口、传输协议 (TLS/QUIC) 及 SNI 证书匹配
            </div>
          </div>
        </div>

        <div class="flex items-center space-x-2 sm:space-x-3 shrink-0 self-end sm:self-auto">
          <el-button
            :icon="useRenderIcon(CloseIcon)"
            @click="handleCancelPage"
          >
            取消
          </el-button>
          <el-button
            type="primary"
            :loading="saving"
            :icon="useRenderIcon(CheckIcon)"
            @click="handleSaveSubmit"
          >
            保存
          </el-button>
        </div>
      </div>

      <!-- Form Component Embedded Directly -->
      <editForm ref="createEditFormRef" :formInline="formInline" />

      <!-- Bottom Action Bar -->
      <div class="flex items-center justify-end space-x-3 pt-4 mt-4 border-t border-[var(--el-border-color-lighter)]">
        <el-button
          :icon="useRenderIcon(CloseIcon)"
          @click="handleCancelPage"
        >
          取消
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          :icon="useRenderIcon(CheckIcon)"
          @click="handleSaveSubmit"
        >
          保存
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
