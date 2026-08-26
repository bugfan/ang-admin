<script setup lang="ts">
import { deviceDetection } from "@pureadmin/utils";
import { ref, reactive } from "vue";
import ReCol from "@/components/ReCol";
import { useI18n } from "vue-i18n";

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "新增",
    id: undefined,
    name: "",
    type: "TLS-TUNNEL",
    port: "",
    remark: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const { t } = useI18n();

const typeOptions = [
  { label: "TLS", value: "TLS-TUNNEL" },
  { label: "QUIC", value: "QUIC-TUNNEL" }
];

const formRules = reactive({
  name: [
    {
      required: true,
      message: () => t("tunnel.nameRequired", "请输入名称"),
      trigger: "blur"
    }
  ],
  type: [
    {
      required: true,
      message: () => t("tunnel.typeRequired"),
      trigger: "change"
    }
  ],
  port: [
    {
      required: true,
      message: () => t("tunnel.portRequired"),
      trigger: "blur"
    },
    {
      validator: (rule: any, value: string, callback: any) => {
        if (!value) {
          callback(new Error(t("tunnel.portRequired")));
        } else {
          const num = Number(value);
          if (isNaN(num) || num < 1 || num > 65535 || !Number.isInteger(num)) {
            callback(new Error(t("tunnel.portFormatError")));
          } else {
            callback();
          }
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
    :label-position="deviceDetection() ? 'top' : 'right'"
    :model="newFormInline"
    :rules="formRules"
    label-width="auto"
    class="tunnel-form p-1 sm:px-2"
  >
    <el-row :gutter="16">
      <re-col :value="24" :xs="24">
        <el-form-item :label="t('tunnel.name', '名称')" prop="name">
          <el-input
            v-model="newFormInline.name"
            :placeholder="t('tunnel.namePlaceholder', '请输入 Tunnel 名称')"
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="12">
        <el-form-item :label="t('tunnel.type')" prop="type">
          <el-select
            v-model="newFormInline.type"
            class="w-full"
            :placeholder="t('tunnel.typeRequired')"
          >
            <el-option
              v-for="item in typeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="12">
        <el-form-item :label="t('tunnel.port')" prop="port">
          <el-input v-model="newFormInline.port" placeholder="443" clearable />
        </el-form-item>
      </re-col>

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('tunnel.remark')" prop="remark">
          <el-input
            v-model="newFormInline.remark"
            type="textarea"
            :rows="3"
            :placeholder="t('tunnel.remarkPlaceholder')"
            clearable
          />
        </el-form-item>
      </re-col>
    </el-row>
  </el-form>
</template>

<style scoped>
.tunnel-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--el-text-color-regular);
  white-space: nowrap;
}
</style>
