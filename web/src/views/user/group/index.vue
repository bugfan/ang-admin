<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useUserGroup } from "./utils/hook";
import editForm from "./form/index.vue";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import { createUserGroup, updateUserGroup } from "@/api/user-group";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";
import BackIcon from "~icons/ep/back";

defineOptions({
  name: "AppUserGroup"
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
  handleDelete
} = useUserGroup(t, tableRef);

function getDefaultFormInline() {
  return {
    title: t("identity.addGroup", "添加用户组"),
    id: undefined,
    name: "",
    description: "",
    is_default: false
  };
}

function getFormInlineFromRow(row: any) {
  return {
    title: `${t("identity.editGroup", "编辑用户组")} [ID: ${row?.Id || row?.id}]`,
    id: row?.Id ?? row?.id,
    name: row?.Name ?? row?.name ?? "",
    description: row?.Description ?? row?.description ?? "",
    is_default: Boolean(row?.IsDefault ?? row?.is_default)
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
          description: formInline.value.description,
          is_default: formInline.value.is_default
        };
        let res;
        if (showView.value === "new") {
          res = await createUserGroup(payload);
        } else {
          res = await updateUserGroup(formInline.value.id, payload);
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
        <el-form-item :label="t('identity.groupName', '用户组名称')" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="t('identity.groupNamePlaceholder', '请输入组名称')"
            clearable
            class="!w-[220px]"
            @keyup.enter="onSearch"
          />
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
        :title="t('identity.groupTitle', '用户组列表')"
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
                :title="t('identity.deleteGroupConfirm', { name: row.Name || row.name }, '确认删除该用户组吗？')"
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
        :description="t('identity.groupDesc', '配置系统用户群组分类，用于统一权限管理与多协议授权')"
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
