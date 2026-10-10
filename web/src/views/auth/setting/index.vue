<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { useI18n } from "vue-i18n";
import { message } from "@/utils/message";
import { getAuthSetting, updateAuthSetting } from "@/api/auth-setting";
import CheckIcon from "~icons/ep/check";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { deviceDetection } from "@pureadmin/utils";

defineOptions({
  name: "AppAuthSetting"
});

const { t } = useI18n();

const formRef = ref();
const form = ref({
  token_name: "_angt",
  token_expire: 86400,
  cookie_domain: "",
  enable_ticket: false,
  ticket_param_name: ""
});
const loading = ref(false);
const saving = ref(false);

const rules = computed(() => ({
  token_name: [
    { required: true, message: t("authSetting.nameRequired", "认证Cookie名称不能为空"), trigger: "blur" }
  ],
  token_expire: [
    { required: true, message: t("authSetting.expireRequired", "凭证过期时间不能为空"), trigger: "blur" }
  ]
}));

async function fetchData() {
  loading.value = true;
  try {
    const res = await getAuthSetting();
    if (res && res.data) {
      form.value.token_name = res.data.token_name || "_angt";
      form.value.token_expire = res.data.token_expire || 86400;
      form.value.cookie_domain = res.data.cookie_domain || "";
      form.value.enable_ticket = !!res.data.enable_ticket;
      form.value.ticket_param_name = res.data.ticket_param_name || "";
    }
  } catch (e) {
    // ignore
  } finally {
    loading.value = false;
  }
}

async function handleSave() {
  const isValid = await formRef.value?.validate();
  if (!isValid) return;

  saving.value = true;
  try {
    const res = await updateAuthSetting(form.value);
    if (res.code === 0) {
      message(t("common.operationSuccess", "操作成功"), { type: "success" });
    } else {
      message(res.message || t("common.operationFailed", "操作失败"), { type: "error" });
    }
  } catch (e) {
    // error handled by axios
  } finally {
    saving.value = false;
  }
}

onMounted(() => {
  fetchData();
});
</script>

<template>
  <div class="main-content p-4" v-loading="loading">
    <el-card shadow="never" class="border-(--el-border-color-lighter)! rounded-xl m-4">
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-primary rounded-full" />
          <span class="font-bold text-(--el-text-color-primary) text-base">
            {{ t("authSetting.title", "认证高级设置") }}
          </span>
        </div>
      </template>

      <div class="p-2 sm:p-6">
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          :label-position="deviceDetection() ? 'top' : 'right'"
          label-width="190px"
          class="w-full mt-4"
        >
          <el-form-item :label="t('authSetting.tokenName', '认证Cookie名称')" prop="token_name">
            <el-input
              v-model="form.token_name"
              :placeholder="t('authSetting.tokenNamePlaceholder', '默认：_angt')"
              clearable
            />
            <div class="w-full text-xs text-(--el-text-color-secondary) mt-1.5">
              {{ t("authSetting.tokenNameHint", "网关拦截并校验用户登录凭证时读取的 HTTP Cookie Key。如无冲突，建议保持默认。") }}
            </div>
          </el-form-item>

          <el-form-item :label="t('authSetting.cookieDomain', 'Cookie共享域名 (Domain)')" prop="cookie_domain" class="mt-8">
            <el-input
              v-model="form.cookie_domain"
              :placeholder="t('authSetting.cookieDomainPlaceholder', '例如：.example.com 或留空')"
              clearable
            />
            <div class="w-full text-xs text-(--el-text-color-secondary) mt-1.5">
              {{ t("authSetting.cookieDomainHint", "认证成功后下发 Cookie 的 domain 属性。若设置为根域名（如 .example.com），同一根域名下的所有子域名业务站点均能直接共享此认证凭据，实现自然免跳转单点登录。留空则仅在当前认证中心域名生效。") }}
            </div>
          </el-form-item>

          <el-form-item :label="t('authSetting.tokenExpire', '会话有效期 (秒)')" prop="token_expire" class="mt-8">
            <el-input-number
              v-model="form.token_expire"
              :min="0"
              :step="3600"
              class="!w-full sm:!w-64"
              controls-position="right"
            />
            <div class="w-full text-xs text-(--el-text-color-secondary) mt-1.5">
              {{ t("authSetting.tokenExpireHint", "凭证的超期时间。默认 86400 秒 (24 小时)。过期后用户将被强制重新走一遍登录流水线。") }}
            </div>
          </el-form-item>

          <el-form-item :label="t('authSetting.enableTicket', '跨顶级域票据传递')" prop="enable_ticket" class="mt-8">
            <el-switch
              v-model="form.enable_ticket"
              inline-prompt
              :active-text="t('common.open', '开启')"
              :inactive-text="t('common.close', '关闭')"
            />
            <div class="w-full text-xs text-(--el-text-color-secondary) mt-1.5">
              {{ t("authSetting.enableTicketHint", "仅当业务系统与认证中心处于不同的顶级域名（跨根域无法共享 Cookie Domain）时才需要开启。开启后，跳转到目标站点时将在 URL Query 中携带一次性票据。") }}
            </div>
          </el-form-item>

          <el-form-item
            v-if="form.enable_ticket"
            :label="t('authSetting.ticketParamName', 'SSO重定向参数名称')"
            prop="ticket_param_name"
            class="mt-8"
          >
            <el-input
              v-model="form.ticket_param_name"
              :placeholder="t('authSetting.ticketParamPlaceholder', '留空则与认证Cookie名称一致 (默认：_angt)')"
              clearable
            />
            <div class="w-full text-xs text-(--el-text-color-secondary) mt-1.5">
              {{ t("authSetting.ticketParamHint", "跨顶级域名跳转时在 URL Query 携带的临时票据参数名。留空则自动使用上述 Cookie 名称（默认 _angt）。") }}
            </div>
          </el-form-item>

          <el-form-item class="mt-10">
            <el-button
              type="primary"
              :loading="saving"
              :icon="useRenderIcon(CheckIcon)"
              @click="handleSave"
              class="px-8"
            >
              {{ t("common.save", "保存配置") }}
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
:deep(.el-form-item__label) {
  font-weight: 500;
}
</style>
