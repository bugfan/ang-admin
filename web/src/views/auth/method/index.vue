<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useAuthMethod } from "./utils/hook";
import editForm from "./form/index.vue";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import { createAuthMethod, updateAuthMethod } from "@/api/auth-method";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import ConnectionIcon from "~icons/ep/connection";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";
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
    title: t("identity.addSource", "添加认证方式"),
    id: undefined,
    name: "",
    type: "local",
    enabled: true,
    priority: 1,
    config_json: "{}",
    remark: "",
};
}

function getFormInlineFromRow(row: any) {
  return {
    title: `${t("identity.editSource", "编辑认证方式")} [ID: ${row?.Id || row?.id}]`,
    id: row?.Id ?? row?.id,
    name: row?.Name ?? row?.name ?? "",
    type: row?.Type ?? row?.type ?? "local",
    enabled: Boolean(row?.Enabled ?? row?.enabled),
    priority: row?.Priority ?? row?.priority ?? 1,
    config_json: row?.ConfigJSON ?? row?.config_json ?? "{}",
    remark: row?.Remark ?? row?.remark ?? "",
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
          type: formInline.value.type,
          enabled: formInline.value.enabled,
          priority: formInline.value.priority,
          config_json: formInline.value.config_json,
          remark: formInline.value.remark
        };
        let res;
        if (showView.value === "new") {
          res = await createAuthMethod(payload);
        } else {
          res = await updateAuthMethod(formInline.value.id, payload);
        }

        if (res && res.code === 0) {
          message(t("common.saveSuccess", "保存成功"), { type: "success" });
          showView.value = "list";
          onSearch();
        } else {
          message(res?.message || t("common.saveFailed", "保存失败"), { type: "error" });
        }
      } catch (err: any) {
        message(err?.message || t("common.saveFailed", "保存失败"), { type: "error" });
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
        class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto"
      >
        <el-form-item :label="t('identity.sourceName', '认证方式名称')" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="t('identity.sourceNamePlaceholder', '请输入名称')"
            clearable
            class="!w-[200px]"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item :label="t('identity.sourceType', '认证类型')" prop="type">
          <el-select
            v-model="form.type"
            clearable
            placeholder="全部类型"
            class="!w-[160px]"
            @change="onSearch"
          >
            <el-option label="本地用户 (Local)" value="local" />
            <el-option label="CAS (v2/v3)" value="cas" />
            <el-option label="RADIUS" value="radius" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSearch">
            {{ t("common.search", "搜索") }}
          </el-button>
          <el-button @click="resetForm(searchFormRef)">
            {{ t("common.reset", "重置") }}
          </el-button>
        </el-form-item>
      </el-form>

      <PureTableBar
        :title="t('identity.authSourceTitle', '认证方式列表')"
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
                :title="t('identity.deleteSourceConfirm', { name: row.Name || row.name }, '确认删除该认证方式吗？')"
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
    <template v-else>
      <PageHeader
        :title="formInline.title"
        :description="t('identity.authSourceDesc', '配置系统身份认证方式，支持本地账号、CAS 单点登录与 RADIUS 认证服务')"
        @back="handleCancelPage"
      >
        <template #actions>
          <div class="flex items-center space-x-2">
            <el-button :icon="useRenderIcon(BackIcon)" @click="handleCancelPage">
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
        </template>
      </PageHeader>

      <div class="bg-bg_color p-4 rounded-xl border border-(--el-border-color-lighter)">
        <editForm ref="createEditFormRef" :formInline="formInline" />
      </div>
    </template>
  </div>
</template>
