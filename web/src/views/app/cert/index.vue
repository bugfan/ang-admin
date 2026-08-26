<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { useI18n } from "vue-i18n";
import { useCert } from "./utils/hook";
import editForm from "./form/index.vue";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import JSZip from "jszip";
import { createCert, updateCert } from "@/api/certificate";
import { wsManager } from "@/utils/websocket";

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

let unbindWs: (() => void) | null = null;

onMounted(() => {
  unbindWs = wsManager.on("CERT_STATUS_CHANGE", (certData: any) => {
    if (!certData) return;
    const targetId = certData.id || certData.Id;
    const status = certData.acme_issue_status || certData.AcmeIssueStatus;
    const domain = certData.subject_cn || certData.cert_id || certData.CertId;

    if (targetId) {
      if (status !== "ISSUING") {
        issuingMap.value[targetId] = false;
      }
    }

    if (status === "SUCCESS") {
      message(`${domain ? `[${domain}] ` : ''}${t('cert.issueSuccess')}`, { type: "success" });
    } else if (status === "FAILED") {
      const errMsg = certData.acme_issue_error || certData.AcmeIssueError || '';
      message(`${domain ? `[${domain}] ` : ''}${t('cert.issueFailed')}${errMsg ? `: ${errMsg}` : ''}`, { type: "error" });
    }

    onSearch();
  });
});

onUnmounted(() => {
  if (unbindWs) {
    unbindWs();
    unbindWs = null;
  }
});

// View Mode: 'list' | 'new' | 'edit'
const showView = ref<"list" | "new" | "edit">("list");
const formInline = ref<any>({});
const saving = ref(false);
const issuingMap = ref<Record<number, boolean>>({});

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
    source: "MANUAL",
    key_content: "",
    cert_content: "",
    intermediate_cert: "",
    remark: "",
    acme_account_id: undefined,
    acme_use_cname: false,
    domains: "",
    auto_renew: true,
    renew_days: 30
  };
}

function getFormInlineFromRow(row: any) {
  return {
    title: `${t("cert.editCert")} [ID: ${row?.Id || row?.id || row?.cert_id}]`,
    id: row?.Id ?? row?.id ?? undefined,
    cert_id: row?.CertId ?? row?.cert_id ?? "",
    type: row?.Type ?? row?.type ?? "STD",
    source: row?.Source ?? row?.source ?? "MANUAL",
    key_content: row?.KeyContent ?? row?.key_content ?? "",
    cert_content: row?.CertContent ?? row?.cert_content ?? "",
    intermediate_cert: row?.IntermediateCert ?? row?.intermediate_cert ?? "",
    remark: row?.Remark ?? row?.remark ?? "",
    acme_account_id: row?.AcmeAccountId ?? row?.acme_account_id ?? undefined,
    acme_use_cname: row?.AcmeUseCname ?? row?.acme_use_cname ?? false,
    domains: row?.Domains ?? row?.domains ?? "",
    auto_renew: row?.AutoRenew ?? row?.auto_renew ?? true,
    renew_days: row?.RenewDays ?? row?.renew_days ?? 30
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

async function handleIssue(row: any) {
  const targetId = row.id || row.Id;
  issuingMap.value[targetId] = true;
  try {
    const { issueCert } = await import("@/api/certificate");
    const res = await issueCert(targetId);
    if (res.code === 0) {
      message(t('cert.issueTaskSubmitted', '签发任务已提交后台执行'), { type: "info" });
      onSearch(); // 立即刷新列表状态为签发中
    } else {
      message(res.message || t('cert.issueFailed'), { type: "error" });
      issuingMap.value[targetId] = false;
      onSearch();
    }
  } catch (error: any) {
    message(error?.message || t('cert.issueFailed'), { type: "error" });
    issuingMap.value[targetId] = false;
    onSearch();
  }
}

function handleDownload(row: any) {
  const name = row.cert_id || row.CertId || "cert";
  const keyContent = row.key_content || row.KeyContent;
  const certContent = row.cert_content || row.CertContent;
  const intermediateContent = row.intermediate_cert || row.IntermediateCert;
  
  if (!keyContent && !certContent) {
    message(t("cert.noContentToDownload", "该证书尚未生成内容，无法下载"), { type: "warning" });
    return;
  }

  const zip = new JSZip();
  if (certContent) {
    zip.file(`${name}.pem`, certContent);
  }
  if (keyContent) {
    zip.file(`${name}.key`, keyContent);
  }
  if (intermediateContent) {
    zip.file(`${name}_ca.pem`, intermediateContent);
    // 同时生成带中间证书的完整证书链 fullchain
    const fullchain = `${certContent.trim()}\n\n${intermediateContent.trim()}\n`;
    zip.file(`${name}_fullchain.pem`, fullchain);
  }

  zip.generateAsync({ type: "blob" }).then((blob) => {
    const url = URL.createObjectURL(blob);
    const element = document.createElement('a');
    element.setAttribute('href', url);
    element.setAttribute('download', `${name}_cert.zip`);
    element.style.display = 'none';
    document.body.appendChild(element);
    element.click();
    document.body.removeChild(element);
    URL.revokeObjectURL(url);
  });
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
          const { code, message: msg, data } = await createCert(curData);
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
        message(e.message || t("cert.submitFailed", "提交失败"), { type: "error" });
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
          </el-select>
        </el-form-item>

        <el-form-item :label="t('cert.source', '来源')" prop="source">
          <el-select
            v-model="form.source"
            :placeholder="t('cert.allSource')"
            clearable
            class="w-full sm:w-45!"
            @change="onSearch"
          >
            <el-option :label="t('cert.sourceAcme')" value="ACME" />
            <el-option :label="t('cert.sourceManual')" value="MANUAL" />
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
            <!-- 展开明细: 公钥/中间证书/私钥文本结构与状态校验 -->
            <template #expand="{ row }">
              <div
                class="p-3 sm:p-4 bg-(--el-fill-color-light) rounded-xl m-1 sm:my-2 sm:mx-4 border border-(--el-border-color-lighter) space-y-3 text-xs"
              >
                <div class="grid grid-cols-1 md:grid-cols-3 gap-3 sm:gap-4">
                  <!-- 证书公钥 (Public Cert) -->
                  <div>
                    <div
                      class="font-bold text-(--el-text-color-primary) mb-1 flex-bc"
                    >
                      <span class="inline-flex items-center gap-1">
                        <IconifyIconOffline icon="ri:shield-keyhole-line" />
                        {{ t("cert.certContent") }} (Public Key)
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

                  <!-- 中间证书 / CA 证书 (Intermediate / CA Cert) -->
                  <div>
                    <div
                      class="font-bold text-(--el-text-color-primary) mb-1 flex-bc"
                    >
                      <span class="inline-flex items-center gap-1">
                        <IconifyIconOffline icon="ri:shield-check-line" />
                        {{ t("cert.intermediateCert", "中间证书 / CA") }}
                      </span>
                      <el-tag
                        size="small"
                        type="info"
                        effect="plain"
                        class="font-mono"
                        >CA / CHAIN</el-tag
                      >
                    </div>
                    <pre
                      class="p-2.5 bg-gray-900 text-cyan-300 rounded-lg font-mono text-[11px] overflow-auto max-h-40 leading-relaxed border border-gray-800"
                      >{{
                        row.intermediate_cert ||
                        row.IntermediateCert ||
                        t("cert.noContent", "无独立中间证书")
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

            <!-- 签发状态列 -->
            <template #issue_status="{ row }">
              <div v-if="row.source === 'ACME' || row.Source === 'ACME'" class="flex items-center justify-center">
                <span v-if="row.acme_issue_status === 'ISSUING' || row.AcmeIssueStatus === 'ISSUING'" class="text-blue-500 text-xs font-semibold">{{ $t('cert.issuing') }}...</span>
                <el-tooltip
                  v-else-if="row.acme_issue_status === 'FAILED' || row.AcmeIssueStatus === 'FAILED'"
                  effect="dark"
                  :content="row.acme_issue_error || row.AcmeIssueError || $t('cert.issueFailed')"
                  placement="top"
                >
                  <span class="text-red-500 text-xs font-semibold cursor-help">{{ $t('cert.issueFailed') }}</span>
                </el-tooltip>
                <span
                  v-else-if="row.acme_issue_status === 'SUCCESS' || row.AcmeIssueStatus === 'SUCCESS'"
                  class="text-green-600 text-xs font-semibold"
                >{{ $t('cert.issueSuccess') }}</span>
                <span v-else class="text-gray-400 text-xs">{{ $t('cert.notIssued') }}</span>
              </div>
              <span v-else class="text-gray-400 text-xs">-</span>
            </template>

            <!-- 选项列 -->
            <template #options="{ row }">
              <div class="flex items-center justify-center gap-2">
                <template v-if="row.source === 'ACME' || row.Source === 'ACME'">
                  <el-button
                    v-if="row.acme_issue_status === 'ISSUING' || row.AcmeIssueStatus === 'ISSUING'"
                    class="reset-margin shrink-0"
                    link
                    type="success"
                    :size="size"
                    loading
                  >
                    {{ $t('cert.issuing') }}
                  </el-button>
                  <el-popconfirm
                    v-else
                    :title="$t('cert.issueConfirmTitle')"
                    @confirm="handleIssue(row)"
                  >
                    <template #reference>
                      <el-button
                        class="reset-margin shrink-0"
                        link
                        type="success"
                        :size="size"
                        :loading="issuingMap[row.id || row.Id]"
                      >
                        {{ issuingMap[row.id || row.Id] ? $t('cert.issuing') : $t('cert.issueBtn') }}
                      </el-button>
                    </template>
                  </el-popconfirm>
                </template>

                <el-button
                  class="reset-margin shrink-0"
                  link
                  type="primary"
                  :size="size"
                  @click="handleDownload(row)"
                >
                  {{ $t('cert.downloadBtn') }}
                </el-button>
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
        :description="t('cert.pageHeaderDesc')"
        :backTitle="t('cert.backToList')"
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

      <!-- Form Component Embedded Directly -->
      <editForm ref="createEditFormRef" :formInline="formInline" />

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
