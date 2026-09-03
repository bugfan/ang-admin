<script setup lang="ts">
import { ref, reactive, watch, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import ReCol from "@/components/ReCol";
import { message } from "@/utils/message";
import { testAuthMethodConnection } from "@/api/auth-method";
import type { FormProps } from "../utils/types";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    id: undefined,
    name: "",
    type: "local",
    enabled: true,
    priority: 1,
    config_json: "{}",
    remark: ""
  })
});

const { t } = useI18n();
const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const testing = ref(false);

// Local Specific
const localConfig = reactive({
  allow_self_register: false,
  password_min_len: 6
});

// CAS Specific
const casConfig = reactive({
  base_url: "",
  version: 2,
  login_path: "/login",
  validate_path: "/serviceValidate",
  skip_tls_verify: false
});

// RADIUS Specific
const radiusConfig = reactive({
  host: "",
  port: 1812,
  secret: "",
  nas_ip: "",
  auth_protocol: "PAP"
});

function initConfigFromProps() {
  try {
    const raw = newFormInline.value.config_json;
    if (raw) {
      const parsed = typeof raw === "string" ? JSON.parse(raw) : raw;
      if (newFormInline.value.type === "local") {
        Object.assign(localConfig, parsed);
      } else if (newFormInline.value.type === "cas") {
        Object.assign(casConfig, parsed);
      } else if (newFormInline.value.type === "radius") {
        Object.assign(radiusConfig, parsed);
      }
    }
  } catch (e) {}
}

function syncConfigToForm() {
  if (newFormInline.value.type === "local") {
    newFormInline.value.config_json = JSON.stringify(localConfig);
  } else if (newFormInline.value.type === "cas") {
    newFormInline.value.config_json = JSON.stringify(casConfig);
  } else if (newFormInline.value.type === "radius") {
    newFormInline.value.config_json = JSON.stringify(radiusConfig);
  }
}

watch(
  () => [localConfig, casConfig, radiusConfig, newFormInline.value.type],
  () => {
    syncConfigToForm();
  },
  { deep: true }
);

async function handleTest() {
  syncConfigToForm();
  testing.value = true;
  try {
    const res = await testAuthMethodConnection({
      type: newFormInline.value.type,
      config_json: newFormInline.value.config_json
    });
    if (res && res.code === 0) {
      message(res.message || t("identity.testSuccess", "连通性测试通过"), { type: "success" });
    } else {
      message(res?.message || t("identity.testFailed", "连通性测试失败"), { type: "error" });
    }
  } catch (e: any) {
    message(e?.message || t("identity.testFailed", "连通性测试失败"), { type: "error" });
  } finally {
    testing.value = false;
  }
}

onMounted(() => {
  initConfigFromProps();
});

function getRef() {
  syncConfigToForm();
  return ruleFormRef.value;
}

defineExpose({ getRef, handleTest });
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :model="newFormInline"
    label-width="140px"
    label-position="top"
    class="space-y-4"
  >
    <!-- 基本设置 -->
    <el-card shadow="never" class="border-(--el-border-color-lighter)! rounded-xl">
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-blue-500 rounded-full" />
          <span class="font-bold text-(--el-text-color-primary) text-sm sm:text-base">
            {{ t("identity.baseInfo", "基本信息") }}
          </span>
        </div>
      </template>

      <el-row :gutter="16">
        <re-col :value="12" :xs="24">
          <el-form-item
            :label="t('identity.sourceName', '认证方式名称')"
            prop="name"
            :rules="[{ required: true, message: () => t('common.nameRequired', '名称不能为空'), trigger: 'blur' }]"
          >
            <el-input
              v-model="newFormInline.name"
              :placeholder="t('identity.sourceNamePlaceholder', '如：本地账号认证、校园 CAS、企业 RADIUS')"
              clearable
            />
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('identity.sourceType', '认证类型')" prop="type">
            <el-select
              v-model="newFormInline.type"
              class="w-full"
            >
              <el-option
                :label="t('identity.sourceTypeLocal', '本地用户 (Local)')"
                value="local"
              />
              <el-option
                :label="t('identity.sourceTypeCas', 'CAS 单点登录 (CAS v2/v3)')"
                value="cas"
              />
              <el-option
                :label="t('identity.sourceTypeRadius', 'RADIUS 认证')"
                value="radius"
              />
            </el-select>
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('identity.priority', '优先级')">
            <el-input-number
              v-model="newFormInline.priority"
              :min="0"
              :max="100"
              class="!w-full"
            />
            <div class="text-xs text-(--el-text-color-secondary) mt-1">
              {{ t("identity.priorityTip", "数值越小优先级越高，用于客户端展示排序") }}
            </div>
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('identity.status', '启用状态')">
            <div class="pt-1">
              <el-switch
                v-model="newFormInline.enabled"
                :active-text="t('identity.statusActive', '启用')"
                :inactive-text="t('identity.statusDisabled', '禁用')"
              />
            </div>
          </el-form-item>
        </re-col>

        <re-col :value="24" :xs="24">
          <el-form-item :label="t('identity.remark', '备注')">
            <el-input
              v-model="newFormInline.remark"
              type="textarea"
              :rows="2"
              :placeholder="t('identity.remarkPlaceholder', '请输入备注信息')"
            />
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>

    <!-- CAS 配置项 -->
    <el-card
      v-if="newFormInline.type === 'cas'"
      shadow="never"
      class="border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <div class="w-1.5 h-4 bg-emerald-500 rounded-full" />
            <span class="font-bold text-(--el-text-color-primary) text-sm sm:text-base">
              CAS 协议参数
            </span>
          </div>
          <el-button
            type="primary"
            plain
            size="small"
            :loading="testing"
            @click="handleTest"
          >
            {{ t("identity.testConnection", "测试连通性") }}
          </el-button>
        </div>
      </template>

      <el-row :gutter="16">
        <re-col :value="16" :xs="24">
          <el-form-item :label="t('identity.casBaseUrl', 'CAS 服务基地址')">
            <el-input
              v-model="casConfig.base_url"
              :placeholder="t('identity.casBaseUrlPlaceholder', '如 https://sso.example.com/cas')"
              clearable
            />
          </el-form-item>
        </re-col>

        <re-col :value="8" :xs="24">
          <el-form-item :label="t('identity.casVersion', 'CAS 协议版本')">
            <el-select v-model="casConfig.version" class="w-full">
              <el-option label="CAS 2.0" :value="2" />
              <el-option label="CAS 3.0" :value="3" />
            </el-select>
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('identity.casLoginPath', '登录跳转路径')">
            <el-input v-model="casConfig.login_path" placeholder="/login" clearable />
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('identity.casValidatePath', 'Ticket 校验路径')">
            <el-input v-model="casConfig.validate_path" placeholder="/serviceValidate" clearable />
          </el-form-item>
        </re-col>

        <re-col :value="24" :xs="24">
          <el-form-item :label="t('identity.casSkipTls', '跳过 TLS 证书验证')">
            <el-switch v-model="casConfig.skip_tls_verify" />
            <span class="text-xs text-(--el-text-color-secondary) ml-2">
              {{ t("identity.casSkipTlsTip", "仅用于自签名或测试环境证书，生产环境建议关闭") }}
            </span>
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>

    <!-- RADIUS 配置项 -->
    <el-card
      v-if="newFormInline.type === 'radius'"
      shadow="never"
      class="border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <div class="w-1.5 h-4 bg-amber-500 rounded-full" />
            <span class="font-bold text-(--el-text-color-primary) text-sm sm:text-base">
              RADIUS 服务器参数
            </span>
          </div>
          <el-button
            type="primary"
            plain
            size="small"
            :loading="testing"
            @click="handleTest"
          >
            {{ t("identity.testConnection", "测试连通性") }}
          </el-button>
        </div>
      </template>

      <el-row :gutter="16">
        <re-col :value="16" :xs="24">
          <el-form-item :label="t('identity.radiusHost', '服务器地址 (Host)')">
            <el-input
              v-model="radiusConfig.host"
              :placeholder="t('identity.radiusHostPlaceholder', '如 192.168.1.100 或 radius.company.local')"
              clearable
            />
          </el-form-item>
        </re-col>

        <re-col :value="8" :xs="24">
          <el-form-item :label="t('identity.radiusPort', '服务端口 (Port)')">
            <el-input-number v-model="radiusConfig.port" :min="1" :max="65535" class="!w-full" />
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('identity.radiusSecret', '共享密钥 (Secret)')">
            <el-input
              v-model="radiusConfig.secret"
              type="password"
              show-password
              :placeholder="t('identity.radiusSecretPlaceholder', '请输入 RADIUS 共享密钥')"
            />
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('identity.radiusProtocol', '认证协议')">
            <el-select v-model="radiusConfig.auth_protocol" class="w-full">
              <el-option label="PAP (Password Authentication Protocol)" value="PAP" />
              <el-option label="CHAP (Challenge Handshake)" value="CHAP" />
            </el-select>
          </el-form-item>
        </re-col>

        <re-col :value="24" :xs="24">
          <el-form-item :label="t('identity.radiusNasIp', 'NAS-IP 地址')">
            <el-input
              v-model="radiusConfig.nas_ip"
              :placeholder="t('identity.radiusNasIpPlaceholder', '选填，如 10.0.0.1')"
              clearable
            />
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>
  </el-form>
</template>
