<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRule } from "./utils/hook";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";

defineOptions({
  name: "AppRule"
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
} = useRule(t, tableRef);

function formatJSON(jsonStr: string) {
  if (!jsonStr) return "-";
  try {
    const obj = typeof jsonStr === "string" ? JSON.parse(jsonStr) : jsonStr;
    return JSON.stringify(obj, null, 2);
  } catch (e) {
    return jsonStr;
  }
}
</script>

<template>
  <div class="main">
    <el-form
      ref="formRef"
      :inline="true"
      :model="form"
      class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto mb-2 rounded-md"
    >
      <el-form-item label="规则名称：" prop="name">
        <el-input
          v-model="form.name"
          placeholder="请输入规则名称"
          clearable
          class="!w-[200px]"
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
          {{ t("rule.search") || "搜索" }}
        </el-button>
        <el-button
          :icon="useRenderIcon('ri:refresh-line')"
          @click="resetForm(formRef)"
        >
          {{ t("rule.reset") || "重置" }}
        </el-button>
      </el-form-item>
    </el-form>

    <PureTableBar
      title="规则管理 (Matcher & Action 匹配控制链)"
      :columns="columns"
      @refresh="onSearch"
    >
      <template #buttons>
        <el-button
          type="primary"
          :icon="useRenderIcon(AddFill)"
          @click="openDialog(t('rule.addRule') || '新增规则')"
        >
          {{ t("rule.addRule") || "新增规则" }}
        </el-button>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <div
          v-if="selectedNum > 0"
          class="bg-[var(--el-color-primary-light-9)] text-[var(--el-color-primary)] border border-[var(--el-color-primary-light-7)] px-3 py-1.5 rounded text-sm mb-2 flex items-center justify-between"
        >
          <span>
            {{ t("rule.selected") || "已选" }} {{ selectedNum }} 项
          </span>
          <div>
            <el-button type="primary" link size="small" @click="onSelectionCancel">
              {{ t("rule.cancelSelection") || "取消选择" }}
            </el-button>
            <el-popconfirm
              title="确认批量删除选中的规则？"
              @confirm="onbatchDel"
            >
              <template #reference>
                <el-button type="danger" link size="small">
                  {{ t("rule.batchDelete") || "批量删除" }}
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
            color: 'var(--el-text-color-primary)'
          }"
          @selection-change="handleSelectionChange"
          @page-size-change="handleSizeChange"
          @page-current-change="handleCurrentChange"
        >
          <!-- Expand Row Slot: Theme-Adaptive Layout -->
          <template #expand="{ row }">
            <div class="p-3 bg-[var(--el-fill-color-light)] rounded-lg m-2 border border-[var(--el-border-color-lighter)]">
              <div class="text-xs font-semibold text-[var(--el-text-color-regular)] mb-2 flex items-center justify-between">
                <div class="flex items-center space-x-2">
                  <div class="w-1.5 h-1.5 bg-[var(--el-color-primary)] rounded-full"></div>
                  <span>规则 JSON 详细配置 [ID: {{ row.Id || row.id }}]</span>
                </div>
                <span class="text-[var(--el-text-color-secondary)] font-mono text-[11px]">Storage JSON string</span>
              </div>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div class="bg-[var(--el-bg-color)] p-3 rounded-md border border-[var(--el-border-color-lighter)]">
                  <div class="flex items-center justify-between mb-1.5 pb-1 border-b border-[var(--el-border-color-lighter)]">
                    <span class="text-xs font-bold text-[var(--el-color-primary)]">Matcher (匹配条件)</span>
                    <el-tag size="small" type="primary" effect="plain" class="font-mono">JSON</el-tag>
                  </div>
                  <pre class="text-xs text-[var(--el-text-color-primary)] font-mono whitespace-pre-wrap break-all max-h-40 overflow-y-auto leading-relaxed">{{ formatJSON(row.Matcher || row.matcher) }}</pre>
                </div>
                <div class="bg-[var(--el-bg-color)] p-3 rounded-md border border-[var(--el-border-color-lighter)]">
                  <div class="flex items-center justify-between mb-1.5 pb-1 border-b border-[var(--el-border-color-lighter)]">
                    <span class="text-xs font-bold text-[var(--el-color-success)]">Action (触发动作)</span>
                    <el-tag size="small" type="success" effect="plain" class="font-mono">JSON</el-tag>
                  </div>
                  <pre class="text-xs text-[var(--el-text-color-primary)] font-mono whitespace-pre-wrap break-all max-h-40 overflow-y-auto leading-relaxed">{{ formatJSON(row.Action || row.action) }}</pre>
                </div>
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
              @click="openDialog(t('rule.editRule') || '编辑规则', row)"
            >
              {{ t("rule.edit") || "编辑" }}
            </el-button>
            <el-popconfirm
              :title="t('rule.confirmDelete') || '是否确认删除该规则配置?'"
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
                  {{ t("rule.delete") || "删除" }}
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </pure-table>
      </template>
    </PureTableBar>
  </div>
</template>

<style scoped>
.search-form :deep(.el-form-item) {
  margin-bottom: 12px;
}
</style>
