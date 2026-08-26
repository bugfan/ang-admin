<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useAcmeAccount } from "./utils/hook";
import acmeAccountEditForm from "./form/index.vue";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import { saveAcmeAccount } from "@/api/acme-account";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";

defineOptions({
  name: "AppAcmeAccount"
});

const { t } = useI18n();
const searchFormRef = ref();
const tableRef = ref();
const createEditFormRef = ref();

// 页面视图模式: 'list' | 'new' | 'edit'
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
} = useAcmeAccount(
  t,
  tableRef,
  row => handleEditPage(row)
);

function getDefaultFormInline() {
  return {
    title: t("acmeAccount.addProvider"),
    id: undefined,
    name: "",
    email: "",
    serverSelect: "https://acme-v02.api.letsencrypt.org/directory",
    customServerUrl: "",
    keyType: "EC256",
    challengeType: "DNS-01",
    provider: "tencentcloud",
    dnsEnvMap: {}
  };
}

function getFormInlineFromRow(row: any) {
  let serverSelect = "https://acme-v02.api.letsencrypt.org/directory";
  let customServerUrl = "";
  if (
    row.directory_url === "https://acme-v02.api.letsencrypt.org/directory" ||
    row.directory_url === "https://acme-staging-v02.api.letsencrypt.org/directory" ||
    row.directory_url === "https://acme.zerossl.com/v2/DV90"
  ) {
    serverSelect = row.directory_url;
  } else if (row.directory_url) {
    serverSelect = "custom";
    customServerUrl = row.directory_url;
  }

  let envMap: Record<string, string> = {};
  try {
    envMap = row.dns_env ? JSON.parse(row.dns_env) : {};
  } catch {
    envMap = {};
  }
  return {
    title: `${t("acmeAccount.editProvider")} (${row.name})`,
    id: row.id,
    name: row.name,
    email: row.email,
    serverSelect,
    customServerUrl,
    keyType: row.key_type || "EC256",
    challengeType: row.challenge_type || "DNS-01",
    provider: row.provider || "tencentcloud",
    dnsEnvMap: envMap
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
        const directoryUrl = curData.serverSelect === "custom" ? curData.customServerUrl : curData.serverSelect;
        const payload = {
          id: curData.id,
          name: curData.name,
          email: curData.email,
          directory_url: directoryUrl,
          key_type: curData.keyType,
          challenge_type: curData.challengeType,
          provider: curData.provider,
          dns_env: JSON.stringify(curData.dnsEnvMap || {})
        };

        const res = await saveAcmeAccount(payload);
        if (res.code === 0) {
          message(
            `${curData.id ? t("acmeAccount.editProvider") : t("acmeAccount.addProvider")} ${t("cert.success", "成功")}`,
            { type: "success" }
          );
          showView.value = "list";
          onSearch();
        } else {
          message(res.message || t("acmeAccount.saveFailed", "保存失败"), { type: "error" });
        }
      } catch (e: any) {
        message(e.message || t("acmeAccount.submitFailed", "提交失败"), { type: "error" });
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
        <el-form-item :label="t('acmeAccount.name')" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="t('acmeAccount.searchNamePlaceholder')"
            clearable
            class="w-full sm:w-45!"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
        </el-form-item>

        <el-form-item :label="t('acmeAccount.provider')" prop="provider">
          <el-select
            v-model="form.provider"
            :placeholder="t('acmeAccount.searchProviderPlaceholder')"
            clearable
            class="w-full sm:w-45!"
            @change="onSearch"
          >
            <el-option label="腾讯云 (TencentCloud)" value="tencentcloud" />
            <el-option label="阿里云 (AliCloud)" value="alidns" />
            <el-option label="DNSPod (经典 Token)" value="dnspod" />
            <el-option label="Cloudflare" value="cloudflare" />
            <el-option label="华为云 (HuaweiCloud)" value="huaweicloud" />
            <el-option label="AWS Route53" value="route53" />
            <el-option label="GoDaddy" value="godaddy" />
            <el-option label="DigitalOcean" value="digitalocean" />
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
        :title="t('menus.pureAcme')"
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
                :title="t('acmeAccount.deleteConfirm')"
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
                  :title="t('acmeAccount.deleteConfirm')"
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
        :description="t('acmeAccount.pageHeaderDesc')"
        :backTitle="t('acmeAccount.backToList')"
        @back="handleCancelPage"
      >
        <template #actions>
          <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
            {{ t("cert.cancel") }}
          </el-button>
          <el-button
            type="primary"
            :loading="saving"
            :icon="useRenderIcon(CheckIcon)"
            @click="handleSaveSubmit"
          >
            {{ t("cert.save") }}
          </el-button>
        </template>
      </PageHeader>

      <!-- Embedded Form Component -->
      <acmeAccountEditForm ref="createEditFormRef" :formInline="formInline" />

      <!-- Bottom Action Bar -->
      <div
        class="flex items-center justify-end space-x-3 pt-4 mt-4 border-t border-(--el-border-color-lighter)"
      >
        <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
          {{ t("cert.cancel") }}
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          :icon="useRenderIcon(CheckIcon)"
          @click="handleSaveSubmit"
        >
          {{ t("cert.save") }}
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
