<script setup lang="ts">
import { ref, reactive, computed } from "vue";
import ReCol from "@/components/ReCol";
import { REGEXP_PWD } from "../utils/rule";
import { useI18n } from "vue-i18n";
import { useUserStoreHook } from "@/store/modules/user";

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "新增",
    username: "",
    password: "",
    repeatPassword: "",
    is_super_admin: false,
    description: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const { t } = useI18n();
const isCurrentUserSuperAdmin = computed(() => useUserStoreHook().is_super_admin);

const formRules = reactive({
  nickname: [{ required: true, message: "用户昵称为必填项", trigger: "blur" }],
  username: [{ required: true, message: "用户名称为必填项", trigger: "blur" }],
  password: [
    {
      validator: (rule, value, callback) => {
        if (!value) {
          callback(new Error("密码为必填项"));
        } else if (!REGEXP_PWD.test(value)) {
          callback(new Error(t("login.purePassWordRuleReg") || "密码格式应为8-18位数字、字母、符号的任意两种组合"));
        } else {
          callback();
        }
      },
      trigger: "blur"
    }
  ],
  repeatPassword: [
    {
      validator: (rule, value, callback) => {
        if (!value) {
          callback(new Error("请确认密码"));
        } else if (newFormInline.value.password && value !== newFormInline.value.password) {
          callback(new Error(t("login.purePassWordDifferentReg") || "两次密码不一致!"));
        } else {
          callback();
        }
      },
      trigger: "blur"
    }
  ]
});

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
    label-width="auto"
  >
    <el-row :gutter="30">
      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('admin.username')" prop="username">
          <el-input
            v-model="newFormInline.username"
            :disabled="!isCurrentUserSuperAdmin && newFormInline.title !== t('admin.addAdmin')"
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('admin.isSuperAdmin')" prop="is_super_admin">
          <el-checkbox
            v-model="newFormInline.is_super_admin"
            :disabled="!isCurrentUserSuperAdmin"
          >
            {{ t('admin.isSuperAdmin') }}
          </el-checkbox>
        </el-form-item>
      </re-col>

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('admin.password')" prop="password">
          <el-input
            v-model="newFormInline.password"
            clearable
            type="password"
          />
        </el-form-item>
      </re-col>

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('admin.repeatPassword')" prop="repeatPassword">
          <el-input
            v-model="newFormInline.repeatPassword"
            clearable
            type="password"
          />
        </el-form-item>
      </re-col>
      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('admin.description')" prop="description">
          <el-input
            v-model="newFormInline.description"
            type="textarea"
            clearable
          />
        </el-form-item>
      </re-col>
    </el-row>
  </el-form>
</template>
