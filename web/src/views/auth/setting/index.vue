<script setup lang="ts">
import { ref, onMounted } from "vue";
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
  token_expire: 86400
});
const loading = ref(false);
const saving = ref(false);

const rules = {
  token_name: [
    { required: true, message: "认证Cookie名称不能为空", trigger: "blur" }
  ],
  token_expire: [
    { required: true, message: "凭证过期时间不能为空", trigger: "blur" }
  ]
};

async function fetchData() {
  loading.value = true;
  try {
    const res = await getAuthSetting();
    if (res && res.data) {
      form.value.token_name = res.data.token_name || "_angt";
      form.value.token_expire = res.data.token_expire || 86400;
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
            设置
          </span>
        </div>
      </template>

      <div class="p-2 sm:p-6">
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          :label-position="deviceDetection() ? 'top' : 'right'"
          label-width="180px"
          class="w-full mt-4"
        >
          <el-form-item label="认证Cookie名称" prop="token_name">
            <el-input
              v-model="form.token_name"
              placeholder="默认：_angt"
              clearable
            />
            <div class="w-full text-xs text-(--el-text-color-secondary) mt-1.5">
              网关拦截并校验用户登录凭证时读取的 HTTP Cookie Key。如无冲突，建议保持默认。
            </div>
          </el-form-item>

          <el-form-item label="会话有效期 (秒)" prop="token_expire" class="mt-8">
            <el-input-number
              v-model="form.token_expire"
              :min="0"
              :step="3600"
              class="!w-full sm:!w-64"
              controls-position="right"
            />
            <div class="w-full text-xs text-(--el-text-color-secondary) mt-1.5">
              凭证的超期时间。默认 86400 秒 (24 小时)。过期后用户将被强制重新走一遍登录流水线。
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
