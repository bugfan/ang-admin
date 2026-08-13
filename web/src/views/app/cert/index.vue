<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useCert } from "./utils/hook";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import View from "~icons/ep/view";

defineOptions({
  name: "AppCert"
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
  openDetailDialog,

  handleDelete,
  handleSizeChange,
  onSelectionCancel,
  handleCurrentChange,
  handleSelectionChange
} = useCert(t, tableRef);
</script>

<template>
  <div :class="['flex', 'justify-between', deviceDetection() && 'flex-wrap']">
    <div class="w-full">
      <!-- 顶部的多条件搜索表单 -->
      <el-form
        ref="formRef"
        :inline="true"
        :model="form"
        label-width="80px"
        :size="deviceDetection() ? 'small' : 'default'"
        class="search-form bg-bg_color w-full pl-4 md:pl-8 pt-3 overflow-auto"
      >
        <el-form-item :label="t('cert.certId')" prop="cert_id">
          <el-input
            v-model="form.cert_id"
            :placeholder="t('cert.searchCertIdPlaceholder')"
            clearable
            class="w-45!"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
        </el-form-item>

        <el-form-item :label="t('cert.type')" prop="type">
          <el-select
            v-model="form.type"
            :placeholder="t('cert.searchTypePlaceholder')"
            clearable
            class="w-45!"
            @change="onSearch"
          >
            <el-option :label="t('cert.allTypes')" value="" />
            <el-option :label="t('cert.std')" value="STD" />
            <el-option :label="t('cert.gm')" value="GM" />
            <el-option :label="t('cert.selfStd')" value="SELF-STD" />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('cert.subjectCn')" prop="subject_cn">
          <el-input
            v-model="form.subject_cn"
            :placeholder="t('cert.searchCnPlaceholder')"
            clearable
            class="w-45!"
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
            {{ t('cert.search') }}
          </el-button>
          <el-button
            :icon="useRenderIcon('ri/refresh-line')"
            @click="resetForm(formRef)"
          >
            {{ t('cert.reset') }}
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 表格及操作栏 -->
      <PureTableBar
        :title="t('menus.pureCert')"
        :columns="columns"
        @refresh="onSearch"
      >
        <template #buttons>
          <el-button
            type="primary"
            :icon="useRenderIcon(AddFill)"
            @click="openDialog(t('cert.addCert'))"
          >
            {{ t('cert.addCert') }}
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
                {{ t('cert.selected') }} {{ selectedNum }} 项
              </span>
              <el-button type="primary" text @click="onSelectionCancel">
                {{ t('cert.cancelSelection') }}
              </el-button>
            </div>
            <el-popconfirm :title="t('cert.confirmDelete')" @confirm="onbatchDel">
              <template #reference>
                <el-button type="danger" text class="mr-1!">
                  {{ t('cert.batchDelete') }}
                </el-button>
              </template>
            </el-popconfirm>
          </div>
          <pure-table
            ref="tableRef"
            :row-key="(row) => row.Id || row.id"
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
              <div class="flex items-center justify-center whitespace-nowrap space-x-1">
                <el-button
                  class="reset-margin"
                  link
                  type="info"
                  :size="size"
                  :icon="useRenderIcon(View)"
                  @click="openDetailDialog(row)"
                >
                  {{ t('cert.viewDetail') }}
                </el-button>
                <el-button
                  class="reset-margin"
                  link
                  type="primary"
                  :size="size"
                  :icon="useRenderIcon(EditPen)"
                  @click="openDialog(t('cert.editCert'), row)"
                >
                  {{ t('cert.edit') }}
                </el-button>
                <el-popconfirm
                  :title="t('cert.confirmDelete')"
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
                      {{ t('cert.delete') }}
                    </el-button>
                  </template>
                </el-popconfirm>
              </div>
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

.main-content {
  margin: 24px 24px 0 !important;
}

.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}
</style>
