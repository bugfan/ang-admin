<script setup lang="ts">
import { deviceDetection } from "@pureadmin/utils";
import { ref, reactive, onMounted } from "vue";
import ReCol from "@/components/ReCol";
import { useI18n } from "vue-i18n";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { generateSelfSignedCert } from "@/api/certificate";
import { getAcmeAccounts } from "@/api/acme-account";
import { message } from "@/utils/message";
import { useRouter } from "vue-router";
import { closeAllDialog } from "@/components/ReDialog";

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "新增",
    id: undefined,
    cert_id: "",
    type: "STD",
    source: "MANUAL",
    key_content: "",
    cert_content: "",
    intermediate_cert: "",
    remark: "",
    acme_account_id: undefined,
    domains: "",
    auto_renew: true,
    renew_days: 30
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const { t } = useI18n();
const router = useRouter();

function goToAcmeAccount() {
  closeAllDialog();
  router.push("/cert/acme");
}

const showGenPanel = ref(false);
const genLoading = ref(false);
const genConfig = reactive({
  common_name: "",
  sans_str: "",
  valid_days: 365
});

const typeOptions = [
  { label: t("cert.std"), value: "STD" },
  { label: t("cert.gm"), value: "GM" }
];

const acmeAccountList = ref<any[]>([]);
onMounted(async () => {
  const res = await getAcmeAccounts();
  if (res.code === 0 && res.data) {
    acmeAccountList.value = res.data;
  }
});

const formRules = reactive({
  cert_id: [
    { required: true, message: () => t("cert.certIdRequired"), trigger: "blur" }
  ],
  type: [
    { required: true, message: () => t("cert.typeRequired"), trigger: "change" }
  ],
  key_content: [
    {
      required: false,
      validator: (rule: any, value: string, callback: any) => {
        if (newFormInline.value.source === "MANUAL" && !value) {
          callback(new Error(t("cert.keyRequired")));
        } else {
          callback();
        }
      },
      trigger: "blur"
    }
  ],
  cert_content: [
    {
      required: false,
      validator: (rule: any, value: string, callback: any) => {
        if (newFormInline.value.source === "MANUAL" && !value) {
          callback(new Error(t("cert.certRequired")));
        } else {
          callback();
        }
      },
      trigger: "blur"
    }
  ],
  acme_account_id: [
    {
      required: false,
      validator: (rule: any, value: any, callback: any) => {
        if (newFormInline.value.source === "ACME" && !value) {
          callback(new Error(t("acme.acmeAccountRequired")));
        } else {
          callback();
        }
      },
      trigger: "change"
    }
  ],
  domains: [
    {
      required: false,
      validator: (rule: any, value: string, callback: any) => {
        if (newFormInline.value.source === "ACME") {
          if (!value || !value.trim()) {
            callback(new Error(t("acme.domainsRequired", "请输入待签发域名")));
            return;
          }
          const domains = value.split(',').map(d => d.trim()).filter(Boolean);
          if (domains.length === 0) {
            callback(new Error(t("acme.domainsRequired", "请输入待签发域名")));
            return;
          }
          const domainRegex = /^(\*\.)?([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$/;
          for (const d of domains) {
            if (!domainRegex.test(d)) {
              callback(new Error(`域名格式不正确: ${d}`));
              return;
            }
          }
          callback();
        } else {
          callback();
        }
      },
      trigger: "blur"
    }
  ]
});

async function handleGenerate() {
  const cn = genConfig.common_name.trim();
  if (!cn) {
    message(t("cert.genCnRequired"), { type: "warning" });
    return;
  }

  genLoading.value = true;
  const rawSans = genConfig.sans_str.trim();
  const dnsNames = rawSans
    ? rawSans
        .split(",")
        .map(s => s.trim())
        .filter(Boolean)
    : [cn, "*." + cn];

  const res = await generateSelfSignedCert({
    common_name: cn,
    dns_names: dnsNames,
    valid_days: genConfig.valid_days
  });
  genLoading.value = false;

  if (res.code === 0 && res.data) {
    newFormInline.value.key_content = res.data.key_content;
    newFormInline.value.cert_content = res.data.cert_content;
    newFormInline.value.intermediate_cert = res.data.intermediate_cert || "";
    newFormInline.value.type = "STD";
    if (!newFormInline.value.cert_id) {
      newFormInline.value.cert_id =
        "id-" + Math.floor(100 + Math.random() * 900);
    }
    showGenPanel.value = false;
    message(t("cert.genSuccess"), { type: "success" });
  } else {
    message(res.message || "生成失败", { type: "error" });
  }
}

function handleSourceChange(val: string) {
  if (val === "ACME") {
    newFormInline.value.type = "STD";
  }
}

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :label-position="deviceDetection() ? 'top' : 'right'"
    :model="newFormInline"
    :rules="formRules"
    label-width="130px"
    class="cert-form p-1 sm:px-2"
  >
    <el-row :gutter="16">
      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('cert.source')">
          <el-radio-group v-model="newFormInline.source" @change="handleSourceChange">
            <el-radio value="MANUAL">{{ t("cert.sourceManual") }}</el-radio>
            <el-radio value="ACME">{{ t("cert.sourceAcme") }}</el-radio>
          </el-radio-group>
        </el-form-item>
      </re-col>

      <re-col :value="24" :xs="24" :sm="24" v-if="newFormInline.source === 'MANUAL'">
        <div
          class="mb-3 flex flex-col sm:flex-row sm:items-center justify-between gap-2"
        >
          <span class="text-xs text-(--el-text-color-secondary)">{{
            t("cert.pasteOrGenTip")
          }}</span>
          <el-button
            type="success"
            plain
            size="small"
            :icon="useRenderIcon('ri/magic-line')"
            class="self-start sm:self-auto shrink-0"
            @click="showGenPanel = !showGenPanel"
          >
            {{ t("cert.autoGenerate") }}
          </el-button>
        </div>

        <!-- 自动生成参数面板 -->
        <el-form-item v-if="showGenPanel" label=" " class="mb-4">
          <el-card
            shadow="never"
            class="w-full bg-gray-50 dark:bg-gray-800 border-dashed border-green-300 rounded-xl border-2!"
          >
            <template #header>
              <div class="flex-bc text-sm font-semibold">
                <span
                  class="text-green-700 dark:text-green-400 font-bold inline-flex items-center gap-1"
                >
                  <IconifyIconOffline icon="ri:magic-line" />
                  {{ t("cert.autoGenerate") }}
                </span>
                <el-button
                  link
                  type="info"
                  size="small"
                  @click="showGenPanel = false"
                >
                  {{ t("cert.closePanel") }}
                </el-button>
              </div>
            </template>
            <el-form
              label-position="top"
              label-width="auto"
              size="small"
              class="pt-1"
            >
              <el-form-item :label="t('cert.genCnLabel')">
                <el-input
                  v-model="genConfig.common_name"
                  :placeholder="t('cert.genCnPlaceholder')"
                  clearable
                />
              </el-form-item>
              <el-form-item :label="t('cert.genSansLabel')">
                <el-input
                  v-model="genConfig.sans_str"
                  :placeholder="t('cert.genSansPlaceholder')"
                  clearable
                />
              </el-form-item>
              <el-form-item :label="t('cert.genValidityLabel')">
                <el-select v-model="genConfig.valid_days" class="w-full">
                  <el-option :label="t('cert.month1')" :value="30" />
                  <el-option :label="t('cert.month3')" :value="90" />
                  <el-option :label="t('cert.month6')" :value="180" />
                  <el-option :label="t('cert.year1')" :value="365" />
                  <el-option :label="t('cert.year3')" :value="1095" />
                  <el-option :label="t('cert.year10')" :value="3650" />
                  <el-option :label="t('cert.year20')" :value="7300" />
                </el-select>
              </el-form-item>
              <div class="text-right">
                <el-button
                  type="success"
                  size="small"
                  :loading="genLoading"
                  @click="handleGenerate"
                >
                  {{ t("cert.genBtn") }}
                </el-button>
              </div>
            </el-form>
          </el-card>
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="12">
        <el-form-item :label="t('cert.certId')" prop="cert_id">
          <el-input
            v-model="newFormInline.cert_id"
            :placeholder="t('cert.certIdPlaceholder')"
            :disabled="!!newFormInline.id"
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col v-if="newFormInline.source === 'MANUAL'" :value="12" :xs="24" :sm="12">
        <el-form-item :label="t('cert.type')" prop="type">
          <el-select
            v-model="newFormInline.type"
            class="w-full"
            :placeholder="t('cert.typeRequired')"
          >
            <el-option
              v-for="item in typeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </re-col>

      <template v-if="newFormInline.source === 'MANUAL'">
        <re-col :value="24" :xs="24" :sm="24">
          <el-form-item :label="t('cert.certContent')" prop="cert_content" :required="true">
            <el-input
              v-model="newFormInline.cert_content"
              type="textarea"
              :rows="5"
              placeholder="-----BEGIN CERTIFICATE----- ... -----END CERTIFICATE-----"
              clearable
              class="font-mono text-xs"
            />
          </el-form-item>
        </re-col>

        <re-col :value="24" :xs="24" :sm="24">
          <el-form-item :label="t('cert.keyContent')" prop="key_content" :required="true">
            <el-input
              v-model="newFormInline.key_content"
              type="textarea"
              :rows="5"
              placeholder="-----BEGIN PRIVATE KEY----- ... -----END PRIVATE KEY-----"
              clearable
              class="font-mono text-xs"
            />
          </el-form-item>
        </re-col>

        <re-col :value="24" :xs="24" :sm="24">
          <el-form-item :label="t('cert.intermediateCert', '中间证书 / CA')" prop="intermediate_cert">
            <el-input
              v-model="newFormInline.intermediate_cert"
              type="textarea"
              :rows="4"
              :placeholder="t('cert.intermediateCertPlaceholder')"
              clearable
              class="font-mono text-xs"
            />
          </el-form-item>
        </re-col>
      </template>

      <template v-else-if="newFormInline.source === 'ACME'">
        <re-col :value="12" :xs="24" :sm="12">
          <el-form-item :label="t('acme.selectAcmeAccount')" prop="acme_account_id" :required="true">
            <div class="flex items-center w-full gap-2">
              <el-select
                v-model="newFormInline.acme_account_id"
                class="flex-1"
                :placeholder="t('acme.selectAcmeAccountPlaceholder')"
              >
                <el-option
                  v-for="dp in acmeAccountList"
                  :key="dp.id"
                  :label="`${dp.name} (${dp.provider?.toUpperCase()})`"
                  :value="dp.id"
                />
              </el-select>
              <el-link
                type="primary"
                :underline="false"
                class="!text-xs font-normal whitespace-nowrap shrink-0"
                @click="goToAcmeAccount"
              >
                {{ t('acme.noneClickToAdd') }}
              </el-link>
            </div>
          </el-form-item>
        </re-col>

        <re-col :value="24" :xs="24" :sm="24">
          <el-form-item :label="t('acme.domains', '签发域名')" prop="domains" :required="true">
            <el-input
              v-model="newFormInline.domains"
              type="textarea"
              :rows="3"
              placeholder="例如: example.com,*.example.com"
              clearable
              class="font-mono text-xs"
            />
          </el-form-item>
        </re-col>

        <re-col :value="24" :xs="24" :sm="24">
          <el-form-item label="CNAME 验证">
            <div class="flex items-center gap-2">
              <el-switch v-model="newFormInline.acme_use_cname" />
              <span class="text-xs text-(--el-text-color-secondary)">
                如果签发域名的 _acme-challenge 记录被 CNAME 到了其他域名，请开启此选项
              </span>
            </div>
          </el-form-item>
        </re-col>

        <re-col :value="24" :xs="24" :sm="24">
          <el-form-item :label="t('acme.autoRenew', '自动续签')">
            <div class="flex flex-wrap items-center gap-4">
              <el-switch v-model="newFormInline.auto_renew" active-text="启用自动续签" />
              <div v-if="newFormInline.auto_renew" class="flex items-center gap-1.5 text-xs text-(--el-text-color-regular)">
                <span>证书到期前</span>
                <el-input-number
                  v-model="newFormInline.renew_days"
                  :min="1"
                  :max="60"
                  size="small"
                  class="w-24!"
                />
                <span>天触发自动签发与重载</span>
              </div>
            </div>
          </el-form-item>
        </re-col>
      </template>

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('cert.remark')" prop="remark">
          <el-input
            v-model="newFormInline.remark"
            type="textarea"
            :rows="2"
            :placeholder="t('cert.remarkPlaceholder')"
            clearable
          />
        </el-form-item>
      </re-col>
    </el-row>
  </el-form>
</template>

<style scoped>
.cert-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--el-text-color-regular);
  white-space: nowrap;
}
</style>
