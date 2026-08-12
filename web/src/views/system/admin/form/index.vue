<script setup lang="ts">
import { ref } from "vue";
import ReCol from "@/components/ReCol";

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "新增",
    nickname: "",
    username: "",
    password: ""
  })
});

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
    label-width="82px"
  >
    <el-row :gutter="30">
      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="Nickname" prop="nickname">
          <el-input
            v-model="newFormInline.nickname"
            clearable
            placeholder="Please enter nickname"
          />
        </el-form-item>
      </re-col>
      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="Username" prop="username">
          <el-input
            v-model="newFormInline.username"
            clearable
            placeholder="Please enter username"
          />
        </el-form-item>
      </re-col>

      <re-col
        v-if="newFormInline.title === '新增'"
        :value="12"
        :xs="24"
        :sm="24"
      >
        <el-form-item label="Password" prop="password">
          <el-input
            v-model="newFormInline.password"
            clearable
            placeholder="Please enter password"
          />
        </el-form-item>
      </re-col>
    </el-row>
  </el-form>
</template>
