<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import { useAcme } from "./utils/hook";
import acmeEditForm from "./form/index.vue";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import dayjs from "dayjs";
import {
  saveAcmeConfig,
  issueAcmeCertByConfigId
} from "@/api/acme";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import BackIcon from "~icons/ep/back";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";
import MagicIcon from "~icons/ri/magic-line";

defineOptions({
  name: "AppCertAcme"
});

const { t } = useI18n();
const router = useRouter();
const searchFormRef = ref();
const tableRef = ref();
const createEditFormRef = ref();

// 页面视图模式: 'list' | 'new' | 'edit'
const showView = ref<"list" | "new" | "edit">("list");
const formInline = ref<any>({});
const saving = ref(false);

// 签发弹窗状态
const issuing = ref(false);
const showResultDialog = ref(false);
const issueResult = ref<any>(null);
const issueError = ref("");
const currentIssuingName = ref("");

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
} = useAcme(
  t,
  tableRef,
  row => handleIssueRow(row),
  row => handleEditPage(row)
);

function getDefaultFormInline() {
  return {
    title: t("acme.addConfig"),
    id: undefined,
    name: "",
    email: "",
    serverSelect: "https://acme-v02.api.letsencrypt.org/directory",
    customServerUrl: "",
    keyType: "EC256",
    challengeType: "DNS-01",
    dnsProvider: "tencentcloud",
    dnsEnvMap: {},
    domains: "",
    certId: "",
    disableCname: true,
    autoRenew: true,
    renewDays: 30
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
    title: `${t("acme.editConfig")} (${row.name})`,
    id: row.id,
    name: row.name,
    email: row.email,
    serverSelect,
    customServerUrl,
    keyType: row.key_type || "EC256",
    challengeType: row.challenge_type || "DNS-01",
    dnsProvider: row.dns_provider || "tencentcloud",
    dnsEnvMap: envMap,
    domains: row.domains || "",
    certId: row.cert_id || "",
    disableCname: row.disable_cname !== false,
    autoRenew: row.auto_renew !== false,
    renewDays: row.renew_days || 30
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

async function handleSaveSubmit(andIssue = false) {
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
          dns_provider: curData.dnsProvider,
          dns_env: JSON.stringify(curData.dnsEnvMap || {}),
          domains: curData.domains,
          cert_id: curData.certId,
          disable_cname: curData.disableCname,
          auto_renew: curData.autoRenew,
          renew_days: curData.renewDays
        };

        const res = await saveAcmeConfig(payload);
        if (res.code === 0) {
          message(
            `${curData.id ? t("acme.editConfig") : t("acme.addConfig")} ${t("cert.success", "成功")}`,
            { type: "success" }
          );
          showView.value = "list";
          onSearch();

          if (andIssue) {
            const targetId = curData.id || (res.data && res.data.id);
            if (targetId) {
              handleIssueRow({ id: targetId, name: curData.name });
            }
          }
        } else {
          message(res.message || "保存失败", { type: "error" });
        }
      } catch (e: any) {
        message(e.message || "提交失败", { type: "error" });
      } finally {
        saving.value = false;
      }
    }
  });
}

// 触发基于配置的立即签发
async function handleIssueRow(row: any) {
  currentIssuingName.value = row.name || `ID #${row.id}`;
  issuing.value = true;
  issueResult.value = null;
  issueError.value = "";
  showResultDialog.value = true;

  try {
    const res = await issueAcmeCertByConfigId(row.id);
    if (res.code === 0 && res.data) {
      issueResult.value = res.data;
      message(`【${currentIssuingName.value}】${t("acme.issueSuccess")}`, {
        type: "success"
      });
      onSearch();
    } else {
      issueError.value = res.message || t("acme.issueFailed");
      message(issueError.value, { type: "error" });
      onSearch();
    }
  } catch (err: any) {
    issueError.value = err.message || "请求超时或服务端处理异常";
    message(issueError.value, { type: "error" });
    onSearch();
  } finally {
    issuing.value = false;
  }
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
        <el-form-item :label="t('acme.configName')" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="t('acme.searchNamePlaceholder')"
            clearable
            class="w-full sm:w-45!"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
        </el-form-item>

        <el-form-item :label="t('acme.domains')" prop="domains">
          <el-input
            v-model="form.domains"
            :placeholder="t('acme.searchDomainsPlaceholder')"
            clearable
            class="w-full sm:w-45!"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
        </el-form-item>

        <el-form-item :label="t('acme.dnsProvider')" prop="dns_provider">
          <el-select
            v-model="form.dns_provider"
            :placeholder="t('acme.searchProviderPlaceholder')"
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
        :title="t('menus.pureCertAcme')"
        :columns="columns"
        @refresh="onSearch"
      >
        <template #buttons>
          <el-button
            type="primary"
            :icon="useRenderIcon(AddFill)"
            @click="handleAddPage"
          >
            {{ t("acme.addConfig") }}
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
                :title="t('acme.deleteConfirm')"
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
                  :loading="issuing && currentIssuingName === row.name"
                  :icon="useRenderIcon(MagicIcon)"
                  @click="handleIssueRow(row)"
                >
                  {{ t("acme.issueNow") }}
                </el-button>

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
                  :title="t('acme.deleteConfirm')"
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
        :description="t('acme.pageHeaderDesc')"
        :backTitle="t('acme.backToList')"
        @back="handleCancelPage"
      >
        <template #actions>
          <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
            {{ t("common.cancel", "取消") }}
          </el-button>
          <el-button
            :loading="saving"
            :icon="useRenderIcon(CheckIcon)"
            @click="handleSaveSubmit(false)"
          >
            {{ t("common.save", "保存") }}
          </el-button>
          <el-button
            type="primary"
            :loading="saving"
            :icon="useRenderIcon(MagicIcon)"
            @click="handleSaveSubmit(true)"
          >
            保存并立即签发
          </el-button>
        </template>
      </PageHeader>

      <!-- Embedded Form Component -->
      <acmeEditForm ref="createEditFormRef" :formInline="formInline" />

      <!-- Bottom Action Bar -->
      <div
        class="flex items-center justify-end space-x-3 pt-4 mt-4 border-t border-(--el-border-color-lighter)"
      >
        <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
          {{ t("common.cancel", "取消") }}
        </el-button>
        <el-button
          :loading="saving"
          :icon="useRenderIcon(CheckIcon)"
          @click="handleSaveSubmit(false)"
        >
          {{ t("common.save", "保存") }}
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          :icon="useRenderIcon(MagicIcon)"
          @click="handleSaveSubmit(true)"
        >
          保存并立即签发
        </el-button>
      </div>
    </div>

    <!-- 签发结果与实时进度弹窗 -->
    <el-dialog
      v-model="showResultDialog"
      :title="`${t('acme.issuingTitle')} - ${currentIssuingName}`"
      width="680px"
      :close-on-click-modal="!issuing"
      :close-on-press-escape="!issuing"
    >
      <div v-if="issuing" class="py-10 text-center space-y-4">
        <IconifyIconOffline icon="ri:loader-4-line" class="animate-spin text-5xl text-(--el-color-primary) mx-auto" />
        <div class="text-base font-bold text-(--el-text-color-primary)">
          {{ t("acme.issuingDesc") }}
        </div>
        <div class="text-xs text-(--el-text-color-secondary) max-w-md mx-auto leading-relaxed">
          正在在云厂商 DNS 自动设置 TXT 校验记录，并等待全国公共 DNS 节点同步生效。请耐心等待。
        </div>
        <div class="flex justify-center gap-2 pt-2">
          <el-tag size="small" type="info">① CA 账户注册</el-tag>
          <el-tag size="small" type="primary">② 注入 DNS TXT 记录</el-tag>
          <el-tag size="small" type="warning">③ 等待传播校验</el-tag>
          <el-tag size="small" type="success">④ 签发并全网入库</el-tag>
        </div>
      </div>

      <div v-else-if="issueResult">
        <el-result icon="success" :title="t('acme.issueSuccess')" sub-title="证书已生成并自动同步至证书库与集群所有节点">
          <template #extra>
            <div class="text-left bg-(--el-fill-color-light) p-4 rounded-xl text-xs space-y-2.5 mb-5 border border-(--el-border-color-lighter)">
              <div class="flex"><span class="w-32 font-bold text-(--el-text-color-secondary)">{{ t("acme.certId") }}:</span> <span class="font-mono font-semibold text-blue-600">{{ issueResult.cert_id }}</span></div>
              <div class="flex"><span class="w-32 font-bold text-(--el-text-color-secondary)">Common Name:</span> <span class="font-mono font-semibold">{{ issueResult.domain }}</span></div>
              <div class="flex"><span class="w-32 font-bold text-(--el-text-color-secondary)">{{ t("cert.issuer") }}:</span> <span>{{ issueResult.issuer }}</span></div>
              <div class="flex"><span class="w-32 font-bold text-(--el-text-color-secondary)">{{ t("cert.notBefore") }}:</span> <span>{{ dayjs(issueResult.not_before).format("YYYY-MM-DD HH:mm:ss") }}</span></div>
              <div class="flex"><span class="w-32 font-bold text-(--el-text-color-secondary)">{{ t("cert.notAfter") }}:</span> <span class="text-green-600 font-semibold">{{ dayjs(issueResult.not_after).format("YYYY-MM-DD HH:mm:ss") }}</span></div>
              <div class="flex"><span class="w-32 font-bold text-(--el-text-color-secondary)">{{ t("cert.sans") }}:</span> <span class="font-mono break-all">{{ issueResult.sans }}</span></div>
            </div>

            <div class="flex justify-center gap-3">
              <el-button @click="showResultDialog = false">
                {{ t("cert.close") }}
              </el-button>
              <el-button type="primary" @click="router.push('/cert/index')">
                {{ t("acme.viewInCertList") }}
              </el-button>
            </div>
          </template>
        </el-result>
      </div>

      <div v-else-if="issueError">
        <el-result icon="error" :title="t('acme.issueFailed')" :sub-title="issueError">
          <template #extra>
            <div class="flex justify-center gap-3">
              <el-button @click="showResultDialog = false">
                {{ t("cert.close") }}
              </el-button>
              <el-button type="primary" @click="handleIssueRow({ id: formInline.id || '', name: currentIssuingName })">
                重试签发
              </el-button>
            </div>
          </template>
        </el-result>
      </div>
    </el-dialog>
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
