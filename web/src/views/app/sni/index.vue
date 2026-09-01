<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useSniProxy } from "./utils/hook";
import editForm from "./form/index.vue";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import { createSni, updateSni } from "@/api/sni";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import BackIcon from "~icons/ep/back";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";

defineOptions({
  name: "AppSniProxy"
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
} = useSniProxy(t, tableRef);

function parseDnsList(rawDns: any): string[] {
  if (!rawDns) return [];
  try {
    if (typeof rawDns === "string") {
      if (rawDns.startsWith("[")) {
        return JSON.parse(rawDns);
      }
      return rawDns.split(/[\n,;]+/).map((s: string) => s.trim()).filter(Boolean);
    }
    if (Array.isArray(rawDns)) return rawDns;
  } catch {
    return [String(rawDns)];
  }
  return [];
}

function getDefaultFormInline() {
  return {
    id: undefined,
    name: "",
    sni: "",
    extra_sni: "",
    port: "443",
    rules: "[]",
    tunnel_type: "quic",
    tunnel_id: "",
    tunnel_token: "",
    dns_resolver: "",
    remark: ""
  };
}

function getFormInlineFromRow(row: any) {
  return {
    id: row.Id || row.id,
    name: row.Name || row.name,
    sni: row.SNI || row.sni,
    extra_sni: row.ExtraSNI || row.extra_sni || "",
    port: row.Port || row.port,
    rules: row.Rules || row.rules,
    tunnel_type: row.TunnelType || row.tunnel_type,
    tunnel_id: row.TunnelId || row.tunnel_id,
    tunnel_token: row.TunnelToken || row.tunnel_token,
    dns_resolver: row.DNSResolver || row.dns_resolver,
    remark: row.Remark || row.remark
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
  onSearch();
}

async function handleSaveSubmit() {
  const formRef = createEditFormRef.value?.getRef();
  if (!formRef) return;
  formRef.validate(async (valid: boolean) => {
    if (valid) {
      saving.value = true;
      try {
        const curData = formInline.value;
        if (showView.value === "new") {
          const { code, message: msg } = await createSni(curData);
          if (code !== 0) {
            message(msg || t("sni.saveFailed", "保存失败"), { type: "error" });
          } else {
            message(t("common.saveSuccess", "保存成功"), { type: "success" });
            handleCancelPage();
          }
        } else {
          const { code, message: msg } = await updateSni(curData);
          if (code !== 0) {
            message(msg || t("sni.saveFailed", "保存失败"), { type: "error" });
          } else {
            message(t("common.saveSuccess", "保存成功"), { type: "success" });
            handleCancelPage();
          }
        }
      } catch (e: any) {
        message(e.message || t("sni.saveFailed", "保存失败"), { type: "error" });
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
        <el-form-item :label="t('sni.name', '名称')" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="t('sni.namePlaceholder', '请输入名称')"
            clearable
            class="w-full sm:w-45!"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item :label="t('sni.sni', 'SNI')" prop="sni">
          <el-input
            v-model="form.sni"
            :placeholder="t('sni.sniPlaceholder', '请输入 SNI')"
            clearable
            class="w-full sm:w-45!"
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
            {{ t("sni.search", "搜索") }}
          </el-button>
          <el-button
            :icon="useRenderIcon('ri:refresh-line')"
            @click="resetForm(searchFormRef)"
          >
            {{ t("sni.reset", "重置") }}
          </el-button>
        </el-form-item>
      </el-form>

      <!-- Table Bar -->
      <PureTableBar
        :title="t('menus.pureSni', 'SNI')"
        :columns="columns"
        @refresh="onSearch"
      >
        <template #buttons>
          <el-button
            type="primary"
            :icon="useRenderIcon(AddFill)"
            @click="handleAddPage"
          >
            {{ t("sni.addSni", "添加") }}
          </el-button>
        </template>
        <template v-slot="{ size, dynamicColumns }">
          <div
            v-if="selectedNum > 0"
            class="bg-(--el-fill-color-light) border border-(--el-color-primary-light-8) text-(--el-color-primary) rounded-md p-2 flex items-center justify-between mb-2 text-sm"
          >
            <div class="flex items-center space-x-2">
              <span class="font-medium"
                >{{ t("sni.selected", "已选") }}
                <span class="font-bold mx-1">{{ selectedNum }}</span>
                {{ t("sni.items", "项") }}</span
              >
              <el-button
                type="primary"
                link
                size="small"
                @click="onSelectionCancel"
              >
                {{ t("sni.cancelSelection", "取消选择") }}
              </el-button>
            </div>
            <el-popconfirm
              :title="t('sni.batchDeleteConfirm', '确认批量删除选中的 SNI 代理？')"
              @confirm="onbatchDel"
            >
              <template #reference>
                <el-button type="danger" size="small" plain>
                  {{ t("sni.batchDelete", "批量删除") }}
                </el-button>
              </template>
            </el-popconfirm>
          </div>
          <pure-table
            ref="tableRef"
            row-key="id"
            align-whole="center"
            showOverflowTooltip
            table-layout="auto"
            :loading="loading"
            :size="size"
            adaptive
            :adaptiveConfig="{ offsetBottom: 108 }"
            :data="dataList"
            :columns="dynamicColumns"
            :pagination="pagination"
            :paginationSmall="size === 'small' ? true : false"
            :header-cell-style="{
              background: 'var(--el-fill-color-light)',
              color: 'var(--el-text-color-primary)'
            }"
            @selection-change="handleSelectionChange"
            @page-size-change="handleSizeChange"
            @page-current-change="handleCurrentChange"
          >
            <!-- Sub table expand -->
            

            <!-- Operation Column -->
            <template #operation="{ row }">
              <div class="flex items-center justify-center space-x-2 whitespace-nowrap">
                <el-button class="reset-margin" link type="primary" :size="size" :icon="useRenderIcon(EditPen)" @click="handleEditPage(row)">
                  {{ t("sni.edit", "编辑") }}
                </el-button>
                <el-popconfirm :title="t('sni.confirmDelete', '是否确认删除该 SNI 代理配置?')" @confirm="handleDelete(row)">
                  <template #reference>
                    <el-button class="reset-margin" link type="danger" :size="size" :icon="useRenderIcon(Delete)">
                      {{ t("sni.delete", "删除") }}
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
    <div v-else-if="showView === 'new' || showView === 'edit'" class="p-3 sm:p-5 bg-bg_color rounded-xl border border-(--el-border-color-lighter) shadow-2xs">
      <PageHeader
        :title="showView === 'new' ? t('sni.addSni') : t('sni.editSni') + ' (id: ' + (formInline.id || 'new') + ')'"
        :description="t('sni.headerDesc', '配置 SNI 代理监听端口、传输层规则与后端 Tunnel/DNS')"
        :backTitle="t('sni.backToList', '返回 SNI 列表')"
        @back="handleCancelPage"
      >
        <template #actions>
          <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
            {{ t("sni.cancel", "取消") }}
          </el-button>
          <el-button type="primary" :loading="saving" :icon="useRenderIcon(CheckIcon)" @click="handleSaveSubmit">
            {{ t("sni.save", "保存") }}
          </el-button>
        </template>
      </PageHeader>

      <editForm ref="createEditFormRef" :formInline="formInline" />

      <div class="flex items-center justify-end space-x-3 pt-4 mt-4 border-t border-(--el-border-color-lighter)">
        <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
          {{ t("sni.cancel", "取消") }}
        </el-button>
        <el-button type="primary" :loading="saving" :icon="useRenderIcon(CheckIcon)" @click="handleSaveSubmit">
          {{ t("sni.save", "保存") }}
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
