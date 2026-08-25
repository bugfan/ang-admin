<script setup lang="ts">
import { ref, reactive, computed } from "vue";
import { useI18n } from "vue-i18n";
import { deviceDetection } from "@pureadmin/utils";
import ReCol from "@/components/ReCol";

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "",
    id: undefined,
    name: "",
    email: "",
    serverSelect: "https://acme-v02.api.letsencrypt.org/directory",
    customServerUrl: "",
    keyType: "EC256",
    challengeType: "DNS-01",
    provider: "tencentcloud",
    dnsEnvMap: {} as Record<string, string>
  })
});

const { t } = useI18n();
const ruleFormRef = ref();
const newFormInline = ref(props.formInline);

// ACME 服务商
const acmeServers = [
  { label: "Let's Encrypt (生产环境 - 推荐)", value: "https://acme-v02.api.letsencrypt.org/directory" },
  { label: "Let's Encrypt (Staging 测试环境 - 无速率限制)", value: "https://acme-staging-v02.api.letsencrypt.org/directory" },
  { label: "ZeroSSL (需支持标准 ACME)", value: "https://acme.zerossl.com/v2/DV90" },
  { label: "自定义 Directory URL", value: "custom" }
];

// 密钥算法
const keyTypes = [
  { label: "ECDSA P-256 (EC256 - 推荐)", value: "EC256" },
  { label: "ECDSA P-384 (EC384)", value: "EC384" },
  { label: "RSA 2048", value: "RSA2048" },
  { label: "RSA 4096", value: "RSA4096" }
];

// DNS 云厂商及对应 API 环境变量模板
const acmeAccounts = [
  {
    name: "tencentcloud",
    label: "腾讯云 (TencentCloud - SecretId / SecretKey)",
    envKeys: [
      { key: "TENCENTCLOUD_SECRET_ID", label: "Secret ID", placeholder: "例如: AKID..." },
      { key: "TENCENTCLOUD_SECRET_KEY", label: "Secret Key", placeholder: "例如: secretKey..." }
    ],
    help: "请前往 腾讯云控制台 -> 访问管理 (CAM) -> API 密钥管理 获取 SecretId 与 SecretKey。"
  },
  {
    name: "alidns",
    label: "阿里云 (AliCloud DNS - AccessKey)",
    envKeys: [
      { key: "ALICLOUD_ACCESS_KEY", label: "Access Key ID", placeholder: "例如: LTAI5t..." },
      { key: "ALICLOUD_SECRET_KEY", label: "Access Key Secret", placeholder: "例如: 8n9x..." }
    ],
    help: "请前往 阿里云控制台 -> AccessKey 管理 获取具有 DNS 解析权限的 API 密钥。"
  },
  {
    name: "dnspod",
    label: "DNSPod 经典 Token (ID,Token)",
    envKeys: [
      { key: "DNSPOD_API_KEY", label: "DNSPod Token (ID,Token)", placeholder: "格式: ID,Token (例如: 12345,6789abcdef...)" }
    ],
    help: "请前往 DNSPod 控制台 -> 账号中心 -> API 密钥 申请 DNSPod Token。如使用的是腾讯云 CAM 密钥，请选择上方的【腾讯云】。"
  },
  {
    name: "cloudflare",
    label: "Cloudflare DNS",
    envKeys: [
      { key: "CLOUDFLARE_DNS_API_TOKEN", label: "API Token", placeholder: "例如: vW8x..." }
    ],
    help: "请前往 Cloudflare 控制台 -> My Profile -> API Tokens 创建具有 Zone.DNS 权限的 Token。"
  },
  {
    name: "huaweicloud",
    label: "华为云 (HuaweiCloud DNS)",
    envKeys: [
      { key: "HUAWEICLOUD_ACCESS_KEY_ID", label: "Access Key (AK)", placeholder: "例如: AK..." },
      { key: "HUAWEICLOUD_SECRET_ACCESS_KEY", label: "Secret Access Key (SK)", placeholder: "例如: SK..." }
    ],
    help: "请前往 华为云控制台 -> 我的凭证 -> Access Key 管理 创建凭证。"
  },
  {
    name: "route53",
    label: "Amazon AWS Route53",
    envKeys: [
      { key: "AWS_ACCESS_KEY_ID", label: "AWS Access Key ID", placeholder: "例如: AKIA..." },
      { key: "AWS_SECRET_ACCESS_KEY", label: "AWS Secret Access Key", placeholder: "例如: secret..." }
    ],
    help: "请前往 AWS IAM 控制台创建包含 Route53ChangeResourceRecordSets 权限的用户凭证。"
  },
  {
    name: "godaddy",
    label: "GoDaddy DNS",
    envKeys: [
      { key: "GODADDY_API_KEY", label: "API Key", placeholder: "例如: key..." },
      { key: "GODADDY_API_SECRET", label: "API Secret", placeholder: "例如: secret..." }
    ],
    help: "请前往 GoDaddy Developer Portal 获取 API Key 与 Secret。"
  },
  {
    name: "digitalocean",
    label: "DigitalOcean DNS",
    envKeys: [
      { key: "DIGITALOCEAN_TOKEN", label: "Personal Access Token", placeholder: "例如: dop_v1_..." }
    ],
    help: "请前往 DigitalOcean 控制台 -> API -> Personal Access Tokens 创建 Token。"
  }
];

const currentAcmeAccountObj = computed(() => {
  const found = acmeAccounts.find(p => p.name === newFormInline.value.provider);
  return found || acmeAccounts[0];
});

const onProviderChange = () => {
  if (!newFormInline.value.dnsEnvMap) {
    newFormInline.value.dnsEnvMap = {};
  }
  const envKeys = currentAcmeAccountObj.value.envKeys;
  for (const item of envKeys) {
    if (newFormInline.value.dnsEnvMap[item.key] === undefined) {
      newFormInline.value.dnsEnvMap[item.key] = "";
    }
  }
};

const formRules = reactive({
  name: [{ required: true, message: () => t("acmeAccount.name") + " " + t("acmeAccount.required"), trigger: "blur" }]
});

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
    label-width="auto"
    class="acme-account-form p-1 sm:px-2"
  >
    <el-row :gutter="20">
      <!-- 基础配置 -->
      <re-col :value="24" :xs="24" :sm="24">
        <div class="font-bold text-sm text-(--el-text-color-primary) border-b border-(--el-border-color-lighter) pb-2 mb-4 flex items-center gap-1.5">
          <IconifyIconOffline icon="ri:settings-4-line" class="text-blue-500" />
          <span>{{ t("acmeAccount.basicConfig") }}</span>
        </div>
      </re-col>

      <re-col :value="12" :xs="24" :sm="12">
        <el-form-item :label="t('acmeAccount.name')" prop="name">
          <el-input
            v-model="newFormInline.name"
            :placeholder="t('acmeAccount.namePlaceholder')"
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="12">
        <el-form-item :label="t('acme.email')" prop="email">
          <el-input
            v-model="newFormInline.email"
            placeholder="例如: admin@yourdomain.com (用于 ACME 账户注册与到期预警)"
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="12">
        <el-form-item :label="t('acme.server')">
          <el-select v-model="newFormInline.serverSelect" class="w-full">
            <el-option
              v-for="item in acmeServers"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
          <el-input
            v-if="newFormInline.serverSelect === 'custom'"
            v-model="newFormInline.customServerUrl"
            placeholder="输入自定义 ACME Directory URL"
            class="mt-2"
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="12">
        <el-form-item :label="t('acme.keyType')">
          <el-select v-model="newFormInline.keyType" class="w-full">
            <el-option
              v-for="k in keyTypes"
              :key="k.value"
              :label="k.label"
              :value="k.value"
            />
          </el-select>
        </el-form-item>
      </re-col>


        <re-col :value="12" :xs="24" :sm="12">
          <el-form-item :label="t('acmeAccount.provider')">
            <el-select
              v-model="newFormInline.provider"
              class="w-full"
              @change="onProviderChange"
            >
              <el-option
                v-for="p in acmeAccounts"
                :key="p.name"
                :label="p.label"
                :value="p.name"
              />
            </el-select>
          </el-form-item>
        </re-col>

        <!-- API 凭证 -->
        <re-col :value="24" :xs="24" :sm="24">
          <div class="font-bold text-sm text-(--el-text-color-primary) border-b border-(--el-border-color-lighter) pb-2 mt-4 mb-4 flex items-center gap-1.5">
            <IconifyIconOffline icon="ri:shield-check-line" class="text-green-500" />
            <span>{{ t("acmeAccount.credentials") }}</span>
          </div>
        </re-col>

        <re-col :value="24" :xs="24" :sm="24">
          <div class="bg-(--el-fill-color-light) p-3 rounded-xl mb-3 text-xs text-(--el-text-color-secondary) border border-(--el-border-color-lighter) flex items-center gap-2">
            <IconifyIconOffline icon="ri:information-line" class="text-blue-500 shrink-0 text-base" />
            <span>{{ currentAcmeAccountObj.help }}</span>
          </div>
        </re-col>

        <re-col
          v-for="env in currentAcmeAccountObj.envKeys"
          :key="env.key"
          :value="12"
          :xs="24"
          :sm="12"
        >
          <el-form-item :label="env.label">
            <el-input
              v-model="newFormInline.dnsEnvMap[env.key]"
              type="password"
              show-password
              :placeholder="env.placeholder"
              clearable
            />
          </el-form-item>
        </re-col>
    </el-row>
  </el-form>
</template>

<style scoped>
.acme-account-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--el-text-color-regular);
  white-space: nowrap;
}
</style>
