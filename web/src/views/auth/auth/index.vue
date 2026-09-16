<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useAuthMethod } from "./utils/hook";
import editForm from "./form/index.vue";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import { createAuth, updateAuth } from "@/api/auth-config";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";
import Search from "~icons/ep/search";
import Refresh from "~icons/ep/refresh";
import BackIcon from "~icons/ep/back";

defineOptions({
  name: "AppAuthMethod"
});

const { t } = useI18n();
const searchFormRef = ref();
const tableRef = ref();
const createEditFormRef = ref();

const showView = ref<"list" | "new" | "edit">("list");
const formInline = ref<any>({});
const saving = ref(false);

const {
  form,
  loading,
  columns,
  dataList,
  pagination,
  onSearch,
  resetForm,
  handleDelete,
  handleTestConnection
} = useAuthMethod(t, tableRef);

function getDefaultFormInline() {
  return {
    title: t("identity.addAuthConfig", "添加认证配置"),
    id: 0,
    name: "",
    auth_method_ids: "[]",
        portal_url: "",
        remark: ""
  };
}

function getFormInlineFromRow(row: any) {
  return {
    title: `${t("identity.editAuthConfig", "编辑认证配置")} [ID: ${row?.Id || row?.id}]`,
    id: row?.Id ?? row?.id,
    name: row?.Name ?? row?.name ?? "",
    auth_method_ids: row?.AuthMethodIds ?? row?.auth_method_ids ?? "[]",
        portal_url: row?.PortalUrl ?? row?.portal_url ?? "",
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
  const formRef = createEditFormRef.value.getRef();
  if (!formRef) return;

  formRef.validate(async (valid: boolean) => {
    if (valid) {
      saving.value = true;
      try {
        const payload: any = {
          id: formInline.value.id,
          name: formInline.value.name,
          auth_method_ids: formInline.value.auth_method_ids,
                    portal_url: formInline.value.portal_url,
                    remark: formInline.value.remark
        };
        let res;
        if (showView.value === "new") {
          res = await createAuth(payload);
        } else {
          res = await updateAuth(formInline.value.id, payload);
        }

        if (res && res.code === 0) {
          message(t("common.saveSuccess", "保存成功"), { type: "success" });
          showView.value = "list";
          onSearch();
        } else {
          message(res?.message || t("common.saveFailed", "保存失败"), {
            type: "error"
          });
        }
      } catch (err: any) {
        message(err?.message || t("common.saveFailed", "保存失败"), {
          type: "error"
        });
      } finally {
        saving.value = false;
      }
    }
  });
}
</script>

<template>
  <div class="main">
    <!-- 1. 列表视图 -->
    <template v-if="showView === 'list'">
      <el-form
        ref="searchFormRef"
        :inline="true"
        :model="form"
        class="search-form bg-bg_color w-full px-3 sm:px-6 pt-3 pb-1 overflow-auto mb-3 rounded-xl border border-(--el-border-color-lighter) shadow-2xs"
      >
        <el-form-item :label="t('identity.sourceName', '名称')" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="t('identity.sourceNamePlaceholder', '请输入名称')"
            clearable
            class="w-50!"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="useRenderIcon(Search)" :loading="loading" @click="onSearch">
            {{ t("common.search", "搜索") }}
          </el-button>
          <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(searchFormRef)">
            {{ t("common.reset", "重置") }}
          </el-button>
        </el-form-item>
      </el-form>

      <PureTableBar
        :title="t('identity.authSourceTitle', '认证列表')"
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
          <pure-table
            ref="tableRef"
            row-key="id"
            adaptive
            :adaptiveConfig="{ offsetBottom: 108 }"
            align-whole="center"
            table-layout="auto"
            :loading="loading"
            :size="size"
            :data="dataList"
            :columns="dynamicColumns"
            :pagination="pagination"
            :header-cell-style="{
              background: 'var(--el-fill-color-light)',
              color: 'var(--el-text-color-primary)',
              fontWeight: 'bold'
            }"
            @page-size-change="onSearch"
            @page-current-change="onSearch"
          >
            <template #operation="{ row }">
              <el-button
                v-if="(row.Type || row.type) !== 'local'"
                class="reset-margin"
                link
                type="primary"
                :size="size"
                :icon="useRenderIcon(ConnectionIcon)"
                @click="handleTestConnection(row)"
              >
                {{ t("identity.testConnection", "测试") }}
              </el-button>
              <el-button
                class="reset-margin"
                link
                type="primary"
                :size="size"
                :icon="useRenderIcon(EditPen)"
                @click="handleEditPage(row)"
              >
                {{ t("common.edit", "编辑") }}
              </el-button>
              <el-popconfirm
                :title="
                  t(
                    'identity.deleteSourceConfirm',
                    { name: row.Name || row.name },
                    '确认删除该认证吗？'
                  )
                "
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
                    {{ t("common.delete", "删除") }}
                  </el-button>
                </template>
              </el-popconfirm>
            </template>
          </pure-table>
        </template>
      </PureTableBar>
    </template>

    <!-- 2. 新增 / 编辑视图 -->
    <div
      v-else
      class="p-3 sm:p-5 bg-bg_color rounded-xl border border-(--el-border-color-lighter) shadow-2xs"
    >
      <PageHeader
        :title="formInline.title"
        :description="
          t(
            'identity.authSourceDesc',
            '配置系统的统一认证策略（SSO），支持串联组合多种认证方式形成双因子/多因子（MFA）认证流水线。'
          )
        "
        :backTitle="t('common.backToList', '返回列表')"
        @back="handleCancelPage"
      >
        <template #actions>
          <el-button
            :icon="useRenderIcon(CloseIcon)"
            @click="handleCancelPage"
          >
            {{ t("common.cancel", "取消") }}
          </el-button>
          <el-button
            type="primary"
            :loading="saving"
            :icon="useRenderIcon(CheckIcon)"
            @click="handleSaveSubmit"
          >
            {{ t("common.save", "保存") }}
          </el-button>
        </template>
      </PageHeader>

      <editForm ref="createEditFormRef" :formInline="formInline" />

      <!-- Bottom Actions -->
      <div
        class="flex items-center justify-end space-x-3 pt-4 mt-4 border-t border-(--el-border-color-lighter)"
      >
        <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
          {{ t("common.cancel", "取消") }}
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          :icon="useRenderIcon(CheckIcon)"
          @click="handleSaveSubmit"
        >
          {{ t("common.save", "保存") }}
        </el-button>
      </div>
    </div>
  </div>
</template>
