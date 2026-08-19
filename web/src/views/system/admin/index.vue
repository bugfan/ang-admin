<script setup lang="ts">
import { ref, computed } from "vue";
import { useI18n } from "vue-i18n";
import { useAdmin } from "./utils/hook";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { useUserStoreHook } from "@/store/modules/user";

import Upload from "~icons/ri/upload-line";
import Role from "~icons/ri/admin-line";
import Password from "~icons/ri/lock-password-line";
import More from "~icons/ep/more-filled";
import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import Refresh from "~icons/ep/refresh";
import AddFill from "~icons/ri/add-circle-line";

defineOptions({
  name: "SystemAdmin"
});

const { t } = useI18n();
const formRef = ref();
const tableRef = ref();
const isSuperAdmin = computed(() => useUserStoreHook().is_super_admin);

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
} = useAdmin(t, tableRef);
</script>

<template>
  <div :class="['flex', 'justify-between', deviceDetection() && 'flex-wrap']">
    <div class="w-full">
      <el-form
        v-if="isSuperAdmin"
        ref="formRef"
        :inline="true"
        :model="form"
        :size="deviceDetection() ? 'small' : 'default'"
        class="search-form bg-bg_color w-full px-4 pt-4 mb-2 rounded overflow-auto"
      >
        <el-form-item :label="t('admin.username')" prop="username">
          <el-input
            v-model="form.username"
            :placeholder="t('admin.searchPlaceholder')"
            clearable
            class="w-36! sm:w-44!"
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
            {{ t("admin.search") }}
          </el-button>
        </el-form-item>
      </el-form>

      <PureTableBar
        :title="t('menus.pureAdminManagement')"
        :columns="columns"
        @refresh="onSearch"
      >
        <template #buttons>
          <el-button
            v-if="isSuperAdmin"
            type="primary"
            :icon="useRenderIcon(AddFill)"
            @click="openDialog(t('admin.addAdmin'))"
          >
            {{ t("admin.addAdmin") }}
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
                {{ t("admin.selected") }} {{ selectedNum }}
                {{ t("admin.items") }}
              </span>
              <el-button type="primary" text @click="onSelectionCancel">
                {{ t("admin.cancelSelection") }}
              </el-button>
            </div>
            <el-popconfirm
              :title="t('admin.confirmDelete')"
              @confirm="onbatchDel"
            >
              <template #reference>
                <el-button type="danger" text class="mr-1!">
                  {{ t("admin.batchDelete") }}
                </el-button>
              </template>
            </el-popconfirm>
          </div>
          <pure-table
            ref="tableRef"
            :row-key="row => row.Id || row.id"
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
            <template #operation="{ row }">
              <el-button
                class="reset-margin"
                link
                type="primary"
                :size="size"
                :icon="useRenderIcon(EditPen)"
                @click="openDialog(t('admin.edit'), row)"
              >
                {{ t("admin.edit") }}
              </el-button>
              <el-popconfirm
                v-if="isSuperAdmin"
                :title="t('admin.confirmDelete')"
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
                    {{ t("admin.delete") }}
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

.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
    margin-right: 16px;
  }
  :deep(.el-form-item__label) {
    font-size: 13px;
    white-space: nowrap;
    font-weight: 500;
    padding-right: 8px;
    text-align: left;
    justify-content: flex-start;
  }
  :deep(.el-input__inner),
  :deep(.el-select__wrapper) {
    font-size: 12px;
  }
  :deep(.el-input__inner::placeholder),
  :deep(.el-select__placeholder) {
    font-size: 12px;
  }
}
</style>
