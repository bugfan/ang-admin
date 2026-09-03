<script setup lang="ts">
import { ref, onMounted } from "vue";
import { FormProps } from "../utils/types";
import { useI18n } from "vue-i18n";
import { getAuthMethodList, type AuthMethodItem } from "@/api/auth-method";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "新增认证",
    id: 0,
    name: "",
    auth_method_ids: "[]",
    token_name: "ANG_TOKEN",
    portal_url: "",
    token_expire: 86400,
    remark: ""
  })
});


const { t } = useI18n();
const formRules = {
  name: [{ required: true, message: t("identity.authConfigNamePlaceholder", "请输入认证名称"), trigger: "blur" }]
};

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const authMethodOptions = ref<AuthMethodItem[]>([]);
const selectedMethods = ref<number[]>([]);

onMounted(async () => {
  try {
    const res = await getAuthMethodList();
    if (res && res.code === 0 && res.data && res.data.list) {
      authMethodOptions.value = res.data.list;
    }
    selectedMethods.value = JSON.parse(newFormInline.value.auth_method_ids || "[]");
  } catch (e) {
    selectedMethods.value = [];
  }
});

function handleMethodChange(val: number[]) {
  newFormInline.value.auth_method_ids = JSON.stringify(val);
}

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :model="newFormInline"
    :rules="formRules"
    label-position="top"
  >
    <el-row :gutter="30">
      <re-col :value="12" :xs="24">
        <el-form-item :label="t('identity.authConfigName', '认证名称')" prop="name">
          <el-input
            v-model="newFormInline.name"
            clearable
            :placeholder="t('identity.authConfigNamePlaceholder', '如：企业内网认证策略')"
          />
        </el-form-item>
      </re-col>
      
      <re-col :value="12" :xs="24">
        <el-form-item :label="t('identity.tokenName', '凭证名称')" prop="token_name">
          <el-input
            v-model="newFormInline.token_name"
            clearable
            :placeholder="t('identity.tokenNamePlaceholder', '如：ANG_TOKEN')"
          />
        </el-form-item>
      </re-col>

      <re-col :value="24" :xs="24">
        <el-form-item :label="t('identity.portalUrl', '登录入口(Portal)')" prop="portal_url">
          <el-input
            v-model="newFormInline.portal_url"
            clearable
            :placeholder="t('identity.portalUrlPlaceholder', '如：https://auth.example.com/login 或 /ang-portal')"
          />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24">
        <el-form-item :label="t('identity.bindAuthMethods', '绑定的认证源')">
          <el-select
            v-model="selectedMethods"
            multiple
            clearable
            class="w-full"
            :placeholder="t('identity.bindAuthMethodsPlaceholder', '请选择认证源(可多选)')"
            @change="handleMethodChange"
          >
            <el-option
              v-for="a in authMethodOptions"
              :key="a.id"
              :label="`${a.name} (${a.type})`"
              :value="a.id"
            />
          </el-select>
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24">
        <el-form-item :label="t('identity.tokenExpire', '凭证过期时间(秒)')" prop="token_expire">
          <el-input-number
            v-model="newFormInline.token_expire"
            :min="0"
            :step="3600"
            class="!w-full"
            controls-position="right"
          />
        </el-form-item>
      </re-col>

      <re-col :value="24" :xs="24">
        <el-form-item :label="t('identity.remark', '备注')" prop="remark">
          <el-input
            v-model="newFormInline.remark"
            type="textarea"
            :rows="3"
            clearable
            :placeholder="t('identity.remarkPlaceholder', '请输入备注信息')"
          />
        </el-form-item>
      </re-col>
    </el-row>
  </el-form>
</template>
