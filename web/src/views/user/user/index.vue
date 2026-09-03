<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useUser } from "./utils/hook";
import editForm from "./form/index.vue";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import { createUser, updateUser } from "@/api/user";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";
import BackIcon from "~icons/ep/back";

defineOptions({
  name: "AppUser"
});

const { t } = useI18n();
const searchFormRef = ref();
const tableRef = ref();
const createEditFormRef = ref();

const showView = ref<"list" | "new" | "edit">("list");
const formInline = ref<any>({});
const saving = ref(false);

// Reset password dialog state
const resetPwdDialogVisible = ref(false);
const resetPwdUserId = ref<number | undefined>();
const resetPwdUsername = ref("");
const newPasswordValue = ref("");
const resetPwdLoading = ref(false);

const {
  form,
  loading,
  columns,
  dataList,
  pagination,
  onSearch,
  resetForm,
  handleDelete
} = useUser(t, tableRef);

function getDefaultFormInline() {
  return {
    title: t("identity.addUser", "添加用户"),
    id: undefined,
    username: "",
    password: "",
    full_name: "",
    email: "",
    mobile: "",
    group_ids: [],
    status: 1,
    expire_at: "",
    remark: ""
  };
}

function getFormInlineFromRow(row: any) {
  let gIds: number[] = [];
  try {
    const raw = row?.GroupIds ?? row?.group_ids ?? "[]";
    gIds = typeof raw === "string" ? JSON.parse(raw) : raw;
  } catch (e) {}

  return {
    title: `${t("identity.editUser", "编辑用户")} [${row?.Username || row?.username}]`,
    id: row?.Id ?? row?.id,
    username: row?.Username ?? row?.username ?? "",
    password: "",
    full_name: row?.FullName ?? row?.full_name ?? "",
    email: row?.Email ?? row?.email ?? "",
    mobile: row?.Mobile ?? row?.mobile ?? "",
    group_ids: Array.isArray(gIds) ? gIds : [],
    status: row?.Status ?? row?.status ?? 1,
    expire_at: row?.ExpireAt ?? row?.expire_at ?? "",
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

function openResetPwdDialog(row: any) {
  resetPwdUserId.value = row?.Id ?? row?.id;
  resetPwdUsername.value = row?.Username ?? row?.username ?? "";
  newPasswordValue.value = "";
  resetPwdDialogVisible.value = true;
}

async function handleConfirmResetPwd() {
  if (!newPasswordValue.value || newPasswordValue.value.length < 6) {
    message(t("identity.passwordPlaceholder", "请输入 6 位以上密码"), { type: "warning" });
    return;
  }
  if (!resetPwdUserId.value) return;

  resetPwdLoading.value = true;
  try {
    const res = await updateUser(resetPwdUserId.value, {
      password: newPasswordValue.value
    });
    if (res && res.code === 0) {
      message(t("common.operationSuccess", "密码重置成功"), { type: "success" });
      resetPwdDialogVisible.value = false;
    } else {
      message(res?.message || t("common.operationFailed", "密码重置失败"), { type: "error" });
    }
  } catch (e: any) {
    message(e?.message || t("common.operationFailed", "操作失败"), { type: "error" });
  } finally {
    resetPwdLoading.value = false;
  }
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
          username: formInline.value.username,
          full_name: formInline.value.full_name,
          email: formInline.value.email,
          mobile: formInline.value.mobile,
          group_ids: JSON.stringify(formInline.value.group_ids || []),
          status: formInline.value.status,
          expire_at: formInline.value.expire_at,
          remark: formInline.value.remark
        };
        if (formInline.value.password) {
          payload.password = formInline.value.password;
        }

        let res;
        if (showView.value === "new") {
          res = await createUser(payload);
        } else {
          res = await updateUser(formInline.value.id, payload);
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
        <el-form-item :label="t('identity.username', '用户名/姓名')" prop="query">
          <el-input
            v-model="form.query"
            placeholder="搜索用户名/姓名/手机"
            clearable
            class="!w-[200px]"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item :label="t('identity.status', '状态')" prop="status">
          <el-select
            v-model="form.status"
            clearable
            placeholder="全部状态"
            class="!w-[130px]"
            @change="onSearch"
          >
            <el-option label="启用" value="1" />
            <el-option label="禁用" value="0" />
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
        :title="t('identity.userTitle', '用户列表')"
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
                :title="t('identity.deleteUserConfirm', { name: row.Username || row.username }, '确认删除该用户吗？')"
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
        :description="t('identity.userDesc', '管理具备访问权限的系统用户，支持本地密码管理与归属组配置')"
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
