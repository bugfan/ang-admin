<script setup lang="ts">
import { deviceDetection } from "@pureadmin/utils";
import { ref, reactive } from "vue";
import ReCol from "@/components/ReCol";
import { useI18n } from "vue-i18n";

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "新增节点",
    id: undefined,
    name: "",
    addr: "http://127.0.0.1:8081",
    secret: "",
    remark: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const { t } = useI18n();

const formRules = reactive({
  name: [
    { required: true, message: () => t("cluster.nodeNameRequired", "请输入名称"), trigger: "blur" }
  ],
  addr: [
    { required: true, message: () => t("cluster.nodeAddrRequired", "请输入地址"), trigger: "blur" }
  ],
  secret: [
    { required: true, message: () => t("cluster.nodeSecretRequired", "请输入节点密钥"), trigger: "blur" }
  ]
});

import { message } from "@/utils/message";
import { verifyClusterNode } from "@/api/cluster_node";

const testing = ref(false);

async function handleTestConnection() {
  if (!newFormInline.value.addr) {
    message(t("cluster.nodeAddrRequired", "请输入地址"), { type: "warning" });
    return;
  }
  if (!newFormInline.value.secret) {
    message(t("cluster.nodeSecretRequired", "请输入节点密钥"), { type: "warning" });
    return;
  }
  
  testing.value = true;
  try {
    const res = await verifyClusterNode({
      addr: newFormInline.value.addr,
      secret: newFormInline.value.secret
    });
    if (res.code === 0) {
      message(t("cluster.testSuccess", "连接并鉴权成功 (Authorized)"), { type: "success" });
    } else {
      let errStr = res.message || "";
      if (errStr === "auth_failed") {
        errStr = t("cluster.testAuthFailed", "鉴权失败：密钥不正确");
      } else if (errStr.includes("timeout") || errStr.includes("connection refused") || errStr.includes("no such host")) {
        errStr = t("cluster.testTimeout", "连接失败：网络不通或地址错误");
      } else if (errStr === "empty address") {
        errStr = t("cluster.nodeAddrRequired", "请输入地址");
      }
      message(errStr || t("cluster.testFailed", "连通性或鉴权测试失败"), { type: "error" });
    }
  } finally {
    testing.value = false;
  }
}

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    :label-position="deviceDetection() ? 'top' : 'right'"
    ref="ruleFormRef"
    :model="newFormInline"
    :rules="formRules"
    label-width="120px"
    class="max-w-4xl"
  >
    <el-row :gutter="30">
      <re-col :value="24">
        <el-form-item :label="t('cluster.formName', '名称')" prop="name">
          <el-input
            v-model="newFormInline.name"
            clearable
            :placeholder="t('cluster.nodeNamePlaceholder', '如 华东主节点 或 Local-Node')"
          />
        </el-form-item>
      </re-col>
      <re-col :value="24">
        <el-form-item :label="t('cluster.formAddr', '地址')" prop="addr">
          <el-input
            v-model="newFormInline.addr"
            clearable
            placeholder="http://127.0.0.1:8081"
            class="font-mono"
          />
          <p class="text-xs text-(--el-text-color-placeholder) mt-1 w-full">
            {{ t("cluster.nodeAddrHint", "API 通信地址及端口（默认端口 :8081）") }}
          </p>
        </el-form-item>
      </re-col>
      
      <re-col :value="24">
        <el-form-item :label="t('cluster.nodeSecret', '密钥')" prop="secret">
          <div class="flex gap-2 w-full">
            <el-input
              v-model="newFormInline.secret"
              clearable
              show-password
              :placeholder="t('cluster.nodeSecretPlaceholder', '请输入节点鉴权密钥')"
              class="flex-1"
            />
            <el-button 
              type="primary" 
              plain 
              :loading="testing" 
              @click="handleTestConnection"
            >
              {{ t('cluster.testConnection', '测试连通性') }}
            </el-button>
          </div>
        </el-form-item>
      </re-col>
      
      <re-col :value="24">
        <el-form-item :label="t('cluster.remark', '备注')" prop="remark">
          <el-input
            v-model="newFormInline.remark"
            type="textarea"
            :rows="2"
            clearable
            :placeholder="t('rule.remarkPlaceholder', '选填')"
          />
        </el-form-item>
      </re-col>
    </el-row>
  </el-form>
</template>
