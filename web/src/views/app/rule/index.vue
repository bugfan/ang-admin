<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRule } from "./utils/hook";
import editForm from "./form/index.vue";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import { createRule, updateRule } from "@/api/rule";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import BackIcon from "~icons/ep/back";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";

defineOptions({
  name: "AppRule"
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

function getDefaultFormInline() {
  return {
    title: t("rule.addRule"),
    id: undefined,
    name: "",
    items: "[]",
    remark: ""
  };
}

function getFormInlineFromRow(row: any) {
  return {
    title: `${t("rule.editRule")} [ID: ${row?.Id || row?.id}]`,
    id: row?.Id ?? row?.id ?? undefined,
    name: row?.Name ?? row?.name ?? "",
    items: row?.Items ?? row?.items ?? JSON.stringify([]),
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
      if (!formInline.value.items || formInline.value.items === "[]") {
        message(t("rule.itemsRequired", "请至少添加一条规则配置"), {
          type: "error"
        });
        return;
      }
      saving.value = true;
      try {
        const curData = formInline.value;
        if (showView.value === "new") {
          const { code, message: msg } = await createRule(curData);
          if (code !== 0) {
            message(msg, { type: "error" });
            return;
          }
          message(t("rule.addRule") + " " + t("rule.success", "成功"), {
            type: "success"
          });
        } else {
          const { code, message: msg } = await updateRule(curData);
          if (code !== 0) {
            message(msg, { type: "error" });
            return;
          }
          message(t("rule.editRule") + " " + t("rule.success", "成功"), {
            type: "success"
          });
        }
        showView.value = "list";
        onSearch();
      } catch (e: any) {
        message(e.message || t("rule.submitFailed", "提交失败"), { type: "error" });
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
      <el-form
        ref="searchFormRef"
        :inline="true"
        :model="form"
        class="search-form bg-bg_color w-full px-3 sm:px-6 pt-3 pb-1 overflow-auto mb-3 rounded-xl border border-(--el-border-color-lighter) shadow-2xs"
      >
        <el-form-item :label="t('rule.name', '名称') + '：'" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="t('rule.ruleGroupSearchPlaceholder')"
            clearable
            class="w-full sm:w-50!"
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
            {{ t("rule.search") }}
          </el-button>
          <el-button
            :icon="useRenderIcon('ri:refresh-line')"
            @click="resetForm(searchFormRef)"
          >
            {{ t("rule.reset") }}
          </el-button>
        </el-form-item>
      </el-form>

      <PureTableBar
        :title="t('rule.ruleGroupTitle')"
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
            <span>
              {{ t("rule.selected") }} {{ selectedNum }} {{ t("rule.items") }}
            </span>
            <div>
              <el-button
                type="primary"
                link
                size="small"
                @click="onSelectionCancel"
              >
                {{ t("rule.cancelSelection") }}
              </el-button>
              <el-popconfirm
                :title="t('rule.confirmDelete')"
                @confirm="onbatchDel"
              >
                <template #reference>
                  <el-button type="danger" link size="small">
                    {{ t("rule.batchDelete") }}
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
            <!-- Expand Row Slot: Theme-Adaptive Layout for Rule Set Items -->
            <template #expand="{ row }">
              <div
                class="p-3 sm:p-4 bg-(--el-fill-color-light) rounded-xl m-1 sm:m-2 border border-(--el-border-color-lighter)"
              >
                <div
                  class="text-xs font-bold text-(--el-text-color-regular) mb-2 flex-bc flex-wrap gap-1"
                >
                  <div class="flex items-center space-x-2">
                    <div class="size-2 bg-(--el-color-primary) rounded-full" />
                    <span
                      >{{ t("rule.ruleGroupTitle") }} Items JSON [ID:
                      {{ row.Id || row.id }}]</span
                    >
                  </div>
                  <span
                    class="text-(--el-text-color-secondary) font-mono text-[11px]"
                    >Storage JSON string</span
                  >
                </div>
                <div
                  class="bg-(--el-bg-color) p-3 rounded-lg border border-(--el-border-color-lighter)"
                >
                  <div
                    class="flex-bc mb-2 pb-1.5 border-b border-(--el-border-color-lighter)"
                  >
                    <span class="text-xs font-bold text-(--el-color-primary)"
                      >Items (Matcher+Action)</span
                    >
                    <el-tag
                      size="small"
                      type="primary"
                      effect="plain"
                      class="font-mono"
                      >JSON Array</el-tag
                    >
                  </div>
                  <el-scrollbar max-height="220px" class="item-scrollbar pr-1">
                    <pre
                      class="text-xs/relaxed text-(--el-text-color-primary) font-mono whitespace-pre-wrap break-all"
                      >{{ formatJSON(row.Items || row.items) }}</pre>
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
                {{ t("rule.edit") }}
              </el-button>
              <el-popconfirm
                :title="t('rule.confirmDelete')"
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
                    {{ t("rule.delete") }}
                  </el-button>
                </template>
              </el-popconfirm>
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
        :title="formInline.title"
        :description="t('rule.headerDesc')"
        :backTitle="t('rule.backToList')"
        @back="handleCancelPage"
      >
        <template #actions>
          <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
            {{ t("rule.cancel") }}
          </el-button>
          <el-button
            type="primary"
            :loading="saving"
            :icon="useRenderIcon(CheckIcon)"
            @click="handleSaveSubmit"
          >
            {{ t("rule.save") }}
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
          {{ t("rule.cancel") }}
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          :icon="useRenderIcon(CheckIcon)"
          @click="handleSaveSubmit"
        >
          {{ t("rule.save") }}
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
