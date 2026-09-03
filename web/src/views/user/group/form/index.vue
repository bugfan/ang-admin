<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import ReCol from "@/components/ReCol";
import type { FormProps } from "../utils/types";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    id: undefined,
    name: "",
    description: "",
    is_default: false
  })
});

const { t } = useI18n();
const ruleFormRef = ref();
const newFormInline = ref(props.formInline);

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :model="newFormInline"
    label-width="120px"
    label-position="top"
    class="space-y-4"
  >
    <el-card shadow="never" class="border-(--el-border-color-lighter)! rounded-xl">
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-indigo-500 rounded-full" />
          <span class="font-bold text-(--el-text-color-primary) text-sm sm:text-base">
            {{ t("identity.baseInfo", "基本信息") }}
          </span>
        </div>
      </template>

      <el-row :gutter="16">
        <re-col :value="16" :xs="24">
          <el-form-item
            :label="t('identity.groupName', '用户组名称')"
            prop="name"
            :rules="[{ required: true, message: () => t('common.nameRequired', '名称不能为空'), trigger: 'blur' }]"
          >
            <el-input
              v-model="newFormInline.name"
              :placeholder="t('identity.groupNamePlaceholder', '如：研发部、外部访客、系统管理员')"
              clearable
            />
          </el-form-item>
        </re-col>

        <re-col :value="8" :xs="24">
          <el-form-item :label="t('identity.isDefaultGroup', '默认组')">
            <div class="pt-1">
              <el-switch v-model="newFormInline.is_default" />
              <div class="text-xs text-(--el-text-color-secondary) mt-1">
                {{ t("identity.isDefaultGroupTip", "外部认证用户首次登录且未指定组时自动归属此组") }}
              </div>
            </div>
          </el-form-item>
        </re-col>

        <re-col :value="24" :xs="24">
          <el-form-item :label="t('identity.groupDescLabel', '描述说明')">
            <el-input
              v-model="newFormInline.description"
              type="textarea"
              :rows="3"
              :placeholder="t('identity.groupDescPlaceholder', '简要说明该组的权限范围或用途')"
            />
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>
  </el-form>
</template>
