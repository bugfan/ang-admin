<script setup lang="ts">
import { deviceDetection } from "@pureadmin/utils";
import { ref, reactive } from "vue";
import { useI18n } from "vue-i18n";
import ReCol from "@/components/ReCol";
import { generateTunnelClientToken } from "@/api/tunnel-client";
import { message } from "@/utils/message";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "新增",
    id: undefined,
    name: "",
    token: "",
    remark: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const { t } = useI18n();

const formRules = reactive({
  name: [
    {
      required: true,
      message: () => t("tunnelClient.nameRequired", "请输入节点名称"),
      trigger: "blur"
    }
  ],
  token: [
    {
      required: true,
      message: () => t("tunnelClient.tokenRequired", "请输入 Token"),
      trigger: "blur"
    }
  ]
});

const generatingToken = ref(false);

async function handleGenerateToken() {
  generatingToken.value = true;
  try {
    const res = await generateTunnelClientToken();
    if (res.code === 0 && res.data?.token) {
      newFormInline.value.token = res.data.token;
      ruleFormRef.value?.validateField("token", () => {});
      message(t("tunnelClient.generateTokenSuccess", "Token 生成成功"), {
        type: "success"
      });
    } else {
      message(res.message || "生成 Token 失败", { type: "error" });
    }
  } catch (err: any) {
    message(err?.message || "生成 Token 失败", { type: "error" });
  } finally {
    generatingToken.value = false;
  }
}

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
    label-width="120px"
    class="p-1"
  >
    <el-row :gutter="16">
      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('tunnelClient.name', '名称')" prop="name">
          <el-input
            v-model="newFormInline.name"
            :placeholder="
              t('tunnelClient.nodeNamePlaceholder', '请输入节点名称')
            "
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('tunnelClient.token', 'Token')" prop="token">
          <el-input
            v-model="newFormInline.token"
            :placeholder="
              t('tunnelClient.tokenPlaceholder', '请输入或生成 Token 凭证')
            "
            clearable
          >
            <template #append>
              <el-button
                :loading="generatingToken"
                :icon="useRenderIcon('ri:magic-line')"
                @click="handleGenerateToken"
              >
                {{ t("tunnelClient.generateToken", "随机生成") }}
              </el-button>
            </template>
          </el-input>
        </el-form-item>
      </re-col>

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('tunnelClient.remark', '备注')" prop="remark">
          <el-input
            v-model="newFormInline.remark"
            type="textarea"
            :rows="2"
            :placeholder="t('tunnelClient.remarkPlaceholder', '请输入备注')"
            clearable
          />
        </el-form-item>
      </re-col>
    </el-row>
  </el-form>
</template>

<style scoped lang="scss">
:deep(.el-form-item) {
  margin-bottom: 18px;
}
</style>
