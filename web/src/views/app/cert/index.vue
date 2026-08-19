<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useCert } from "./utils/hook";
import editForm from "./form/index.vue";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import { createCert, updateCert } from "@/api/certificate";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import BackIcon from "~icons/ep/back";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";

defineOptions({
  name: "AppCertificate"
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
} = useCert(t, tableRef);

function getDefaultFormInline() {
  return {
    title: t("cert.addCert"),
    id: undefined,
    cert_id: "",
    type: "STD",
    key_content: "",
    cert_content: "",
    remark: ""
  };
}

function getFormInlineFromRow(row: any) {
  return {
    title: `${t("cert.editCert")} [ID: ${row?.Id || row?.id || row?.cert_id}]`,
    id: row?.Id ?? row?.id ?? undefined,
    cert_id: row?.CertId ?? row?.cert_id ?? "",
    type: row?.Type ?? row?.type ?? "STD",
    key_content: row?.KeyContent ?? row?.key_content ?? "",
    cert_content: row?.CertContent ?? row?.cert_content ?? "",
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
      saving.value = true;
      try {
        const curData = formInline.value;
        if (showView.value === "new") {
          const { code, message: msg } = await createCert(curData);
          if (code !== 0) {
            message(msg, { type: "error" });
            return;
          }
          message(`${t("cert.addCert")} ${t("cert.success", "成功")}`, {
            type: "success"
          });
        } else {
          const { code, message: msg } = await updateCert(curData);
          if (code !== 0) {
            message(msg, { type: "error" });
            return;
          }
          message(`${t("cert.editCert")} ${t("cert.success", "成功")}`, {
            type: "success"
          });
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
      <!-- 顶部的多条件搜索表单 -->
      <el-form
        ref="searchFormRef"
        :inline="true"
        :model="form"
        class="search-form bg-bg_color w-full px-3 sm:px-6 pt-3 pb-1 overflow-auto mb-3 rounded-xl border border-(--el-border-color-lighter) shadow-2xs"
      >
        <el-form-item :label="t('cert.certId')" prop="cert_id">
          <el-input
            v-model="form.cert_id"
            :placeholder="t('cert.searchCertIdPlaceholder')"
            clearable
            class="w-full sm:w-45!"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
        </el-form-item>

        <el-form-item :label="t('cert.type')" prop="type">
          <el-select
            v-model="form.type"
            :placeholder="t('cert.searchTypePlaceholder', '请选择类型')"
            clearable
            class="w-full sm:w-45!"
            @change="onSearch"
          >
            <el-option label="STD" value="STD" />
            <el-option label="GM" value="GM" />
            <el-option label="SELF-STD" value="SELF-STD" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :icon="useRenderIcon('ri:search-line')"
            :loading="loading"
            @click="onSearch"
          >
            {{ t("cert.search") }}
          </el-button>
          <el-button
            :icon="useRenderIcon('ri:refresh-line')"
            @click="resetForm(searchFormRef)"
          >
            {{ t("cert.reset") }}
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
            @click="handleAddPage"
          >
            {{ t("cert.addCert") }}
          </el-button>
        </template>
        <template v-slot="{ size, dynamicColumns }">
          <div
            v-if="selectedNum > 0"
            class="bg-(--el-color-primary-light-9) text-(--el-color-primary) border border-(--el-color-primary-light-7) px-4 py-2 rounded-lg text-sm mb-3 flex-bc"
          >
            <span
              >{{ t("cert.selected") }} {{ selectedNum }}
              {{ t("cert.items") }}</span
            >
            <div>
              <el-button
                type="primary"
                link
                size="small"
                @click="onSelectionCancel"
              >
                {{ t("cert.cancelSelection") }}
              </el-button>
              <el-popconfirm
                :title="t('cert.confirmDelete')"
                @confirm="onbatchDel"
              >
                <template #reference>
                  <el-button type="danger" link size="small">
                    {{ t("cert.batchDelete") }}
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
            <!-- 展开明细: 公钥/私钥文本结构与状态校验 -->
            <template #expand="{ row }">
              <div
                class="p-3 sm:p-4 bg-(--el-fill-color-light) rounded-xl m-1 sm:my-2 sm:mx-4 border border-(--el-border-color-lighter) space-y-3 text-xs"
              >
                <div class="grid grid-cols-1 md:grid-cols-2 gap-3 sm:gap-4">
                  <!-- 证书公钥 (Public Cert) -->
                  <div>
                    <div
                      class="font-bold text-(--el-text-color-primary) mb-1 flex-bc"
                    >
                      <span class="inline-flex items-center gap-1">
                        <IconifyIconOffline icon="ri:shield-keyhole-line" />
                        {{ t("cert.certContent") }} (Public Key / Certificate)
                      </span>
                      <el-tag
                        size="small"
                        type="success"
                        effect="plain"
                        class="font-mono"
                        >CRT / PEM</el-tag
                      >
                    </div>
                    <pre
                      class="p-2.5 bg-gray-900 text-green-400 rounded-lg font-mono text-[11px] overflow-auto max-h-40 leading-relaxed border border-gray-800"
                      >{{
                        row.cert_content ||
                        row.CertContent ||
                        t("cert.noContent")
                      }}</pre>
                  </div>

                  <!-- 证书私钥 (Private Key) -->
                  <div>
                    <div
                      class="font-bold text-(--el-text-color-primary) mb-1 flex-bc"
                    >
                      <span class="inline-flex items-center gap-1">
                        <IconifyIconOffline icon="ri:lock-line" />
                        {{ t("cert.keyContent") }} (Private Key)
                      </span>
                      <el-tag
                        size="small"
                        type="warning"
                        effect="plain"
                        class="font-mono"
                        >KEY / RSA</el-tag
                      >
                    </div>
                    <pre
                      class="p-2.5 bg-gray-900 text-amber-300 rounded-lg font-mono text-[11px] overflow-auto max-h-40 leading-relaxed border border-gray-800"
                      >{{
                        row.key_content || row.KeyContent || t("cert.noContent")
                      }}</pre>
                  </div>
                </div>
              </div>
            </template>

            <!-- 操作列 -->
            <template #operation="{ row }">
              <div class="flex-c space-x-1">
                <el-button
                  class="reset-margin"
                  link
                  type="primary"
                  :size="size"
                  :icon="useRenderIcon(EditPen)"
                  @click="handleEditPage(row)"
                >
                  {{ t("cert.edit") }}
                </el-button>
                <el-popconfirm
                  :title="t('cert.confirmDelete')"
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
                      {{ t("cert.delete") }}
                    </el-button>
                  </template>
                </el-popconfirm>
              </div>
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
        description="配置 TLS/HTTPS 站点所需的 SSL/TLS 证书及私钥内容，支持一键自动生成自签名证书"
        :backTitle="t('cert.backToList', '返回证书列表')"
        @back="handleCancelPage"
      >
        <template #actions>
          <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
            取消
          </el-button>
          <el-button
            type="primary"
            :loading="saving"
            :icon="useRenderIcon(CheckIcon)"
            @click="handleSaveSubmit"
          >
            保存
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
          取消
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          :icon="useRenderIcon(CheckIcon)"
          @click="handleSaveSubmit"
        >
          保存
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
