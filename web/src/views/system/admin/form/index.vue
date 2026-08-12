<script setup lang="ts">
import { ref } from "vue";
import ReCol from "@/components/ReCol";
import { formRules } from "../utils/rule";
import { useI18n } from "vue-i18n";

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "新增",
    username: "",
    password: "",
    repeatPassword: "",
    description: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const { t } = useI18n();

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
            clearable
          />
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
