<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useTcpProxy } from "./utils/hook";
import editForm from "./form/index.vue";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import { createTcp, updateTcp } from "@/api/tcp";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import BackIcon from "~icons/ep/back";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";

defineOptions({
  name: "AppTcpProxy"
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
} = useTcpProxy(t, tableRef);

function parseUpstreamJSON(upstreamJsonStr: string) {
  let servers: Array<{ target: string; weight: number }> = [];
  try {
    if (upstreamJsonStr) {
      const parsed = JSON.parse(upstreamJsonStr);
      if (Array.isArray(parsed)) {
        servers = parsed.map((item: any) => ({
          target: item.target || item.Target || "",
          weight: Number(item.weight || item.Weight || 1)
        }));
      }
    }
  } catch (e) {
    servers = [];
  }
  return servers;
}

function getDefaultFormInline() {
  return {
    title: t("tcp.addTcp", "添加 TCP 代理"),
    id: undefined,
    name: "",
    address: "",
    port: "2443",
    rules: "[]",
    tunnel_type: "quic",
    tunnel_id: "",
    tunnel_token: "",
    upstream_method: "round_robin",
    upstream_servers: JSON.stringify([{ target: "127.0.0.1:9999", weight: 1 }]),
    remark: ""
  };
}

function getFormInlineFromRow(row: any) {
  return {
    title: `${t("tcp.editTcp", "编辑 TCP 代理")} (${row.Name || row.name || row.Port || row.port})`,
    id: row.Id || row.id,
    name: row.Name || row.name || "",
    address: row.Address || row.address || "",
    port: String(row.Port || row.port || ""),
    rules: row.Rules || row.rules || "[]",
    tunnel_type: row.TunnelType || row.tunnel_type || "quic",
    tunnel_id: row.TunnelId || row.tunnel_id || "",
    tunnel_token: row.TunnelToken || row.tunnel_token || "",
    upstream_method: row.UpstreamMethod || row.upstream_method || "round_robin",
    upstream_servers:
      row.UpstreamServers ||
      row.upstream_servers ||
      JSON.stringify([{ target: "127.0.0.1:9999", weight: 1 }]),
    remark: row.Remark || row.remark || ""
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
          const { code, message: msg } = await createTcp(curData);
          if (code !== 0) {
            message(msg, { type: "error" });
            return;
          }
          message(t("tcp.success", "添加成功"), { type: "success" });
        } else {
          const { code, message: msg } = await updateTcp(curData);
          if (code !== 0) {
            message(msg, { type: "error" });
            return;
          }
          message(t("tcp.success", "更新成功"), { type: "success" });
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
      <!-- 搜索栏 -->
      <el-form
        ref="searchFormRef"
        :inline="true"
        :model="form"
        class="search-form bg-bg_color w-full pl-8 pt-3 pb-2 overflow-auto"
      >
        <el-form-item :label="t('tcp.name', '名称')" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="t('tcp.namePlaceholder', '请输入 TCP 代理名称')"
            clearable
            class="w-full sm:w-45!"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item :label="t('tcp.port', '端口')" prop="port">
          <el-input
            v-model="form.port"
            :placeholder="t('tcp.searchPortPlaceholder', '请输入端口')"
            clearable
            class="w-full sm:w-40!"
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
            {{ t("tcp.search", "搜索") }}
          </el-button>
          <el-button
            :icon="useRenderIcon('ri:refresh-line')"
            @click="resetForm(searchFormRef)"
          >
            {{ t("tcp.reset", "重置") }}
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 表格及操作栏 -->
      <PureTableBar
        :title="t('menus.pureTcp', 'TCP')"
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
              >{{ t("tcp.selected", "已选") }} {{ selectedNum }}
              {{ t("tcp.items", "项") }}</span
            >
            <div>
              <el-button
                type="primary"
                link
                size="small"
                @click="onSelectionCancel"
              >
                {{ t("tcp.cancelSelection", "取消选择") }}
              </el-button>
              <el-popconfirm
                :title="t('tcp.confirmDelete', '是否确认删除选中的配置?')"
                @confirm="onbatchDel"
              >
                <template #reference>
                  <el-button type="danger" link size="small">
                    {{ t("tcp.batchDelete", "批量删除") }}
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
            <!-- Expand Row Slot: Detailed Inspection -->
            



            <!-- Operation Column -->
            <template #operation="{ row }">
              <div
                class="flex items-center justify-center space-x-2 whitespace-nowrap"
              >
                <el-button
                  class="reset-margin"
                  link
                  type="primary"
                  :size="size"
                  :icon="useRenderIcon(EditPen)"
                  @click="handleEditPage(row)"
                >
                  {{ t("tcp.edit", "编辑") }}
                </el-button>
                <el-popconfirm
                  :title="t('tcp.confirmDelete', '是否确认删除该 TCP 代理配置?')"
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
                      {{ t("tcp.delete", "删除") }}
                    </el-button>
                  </template>
                </el-popconfirm>
              </div>
            </template>
          </pure-table>
        </template>
      </PureTableBar>
    </div>

    <!-- Create / Edit Full Page Mode -->
    <div
      v-else-if="showView === 'new' || showView === 'edit'"
      class="p-3 sm:p-5 bg-bg_color rounded-xl border border-(--el-border-color-lighter) shadow-2xs"
    >
      <!-- Full Page Header Bar -->
      <PageHeader
        :title="showView === 'new' ? t('tcp.addTcp') : t('tcp.editTcp') + ' (id: ' + (formInline.id || 'new') + ')'"
        :description="t('tcp.headerDesc', '配置 TCP 代理监听端口、传输层中间件规则 (Rule) 与 Backend 上游服务器')"
        :backTitle="t('tcp.backToList', '返回 TCP 列表')"
        @back="handleCancelPage"
      >
        <template #actions>
          <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
            {{ t("tcp.cancel", "取消") }}
          </el-button>
          <el-button
            type="primary"
            :loading="saving"
            :icon="useRenderIcon(CheckIcon)"
            @click="handleSaveSubmit"
          >
            {{ t("tcp.save", "保存") }}
          </el-button>
        </template>
      </PageHeader>

      <!-- Form Component Embedded Directly (Full Width) -->
      <editForm ref="createEditFormRef" :formInline="formInline" />

      <!-- Bottom Action Bar -->
      <div
        class="flex items-center justify-end space-x-3 pt-4 mt-4 border-t border-(--el-border-color-lighter)"
      >
        <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
          {{ t("tcp.cancel", "取消") }}
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          :icon="useRenderIcon(CheckIcon)"
          @click="handleSaveSubmit"
        >
          {{ t("tcp.save", "保存") }}
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
