<script setup lang="ts">
import { ref, reactive } from "vue";
import ReCol from "@/components/ReCol";
import { useI18n } from "vue-i18n";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { generateSelfSignedCert } from "@/api/certificate";
import { message } from "@/utils/message";

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "新增",
    id: undefined,
    cert_id: "",
    type: "STD",
    key_content: "",
    cert_content: "",
    remark: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const { t } = useI18n();

const showGenPanel = ref(false);
const genLoading = ref(false);
const genConfig = reactive({
  common_name: "",
  sans_str: "",
  valid_days: 365
});

const typeOptions = [
  { label: t("cert.std"), value: "STD" },
  { label: t("cert.gm"), value: "GM" },
  { label: t("cert.selfStd"), value: "SELF-STD" }
];

const formRules = reactive({
  cert_id: [
    { required: true, message: () => t("cert.certIdRequired"), trigger: "blur" }
  ],
  type: [
    { required: true, message: () => t("cert.typeRequired"), trigger: "change" }
  ],
  key_content: [
    { required: true, message: () => t("cert.keyRequired"), trigger: "blur" }
  ],
  cert_content: [
    { required: true, message: () => t("cert.certRequired"), trigger: "blur" }
  ]
});

async function handleGenerate() {
  const cn = genConfig.common_name.trim();
  if (!cn) {
    message(t("cert.genCnRequired"), { type: "warning" });
    return;
  }

  genLoading.value = true;
  const rawSans = genConfig.sans_str.trim();
  const dnsNames = rawSans
    ? rawSans.split(",").map(s => s.trim()).filter(Boolean)
    : [cn, "*." + cn];

  const res = await generateSelfSignedCert({
    common_name: cn,
    dns_names: dnsNames,
    valid_days: genConfig.valid_days
  });
  genLoading.value = false;

  if (res.code === 0 && res.data) {
    newFormInline.value.key_content = res.data.key_content;
    newFormInline.value.cert_content = res.data.cert_content;
    newFormInline.value.type = "SELF-STD";
    if (!newFormInline.value.cert_id) {
      newFormInline.value.cert_id = "id-" + Math.floor(100 + Math.random() * 900);
    }
    showGenPanel.value = false;
    message(t("cert.genSuccess"), { type: "success" });
  } else {
    message(res.message || "生成失败", { type: "error" });
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
    :model="newFormInline"
    :rules="formRules"
    label-width="auto"
  >
    <el-row :gutter="30">
      <re-col :value="24" :xs="24" :sm="24">
        <div class="mb-3 flex justify-between items-center">
          <span class="text-xs text-gray-500">可手动粘贴证书或使用右侧一键自动生成：</span>
          <el-button
            type="success"
            plain
            size="small"
            :icon="useRenderIcon('ri/magic-line')"
            @click="showGenPanel = !showGenPanel"
          >
            {{ t('cert.autoGenerate') }}
          </el-button>
        </div>

        <!-- 自动生成参数面板 -->
        <el-card
          v-if="showGenPanel"
          shadow="never"
          class="mb-4 bg-gray-50 dark:bg-gray-800 border-dashed border-green-300"
        >
          <template #header>
            <div class="flex items-center justify-between text-sm font-semibold">
              <span class="text-green-700 dark:text-green-400 font-bold">
                ⚡ {{ t('cert.autoGenerate') }}
              </span>
              <el-button link type="info" size="small" @click="showGenPanel = false">
                关闭面板
              </el-button>
            </div>
          </template>
          <el-form label-width="auto" size="small" class="pt-1">
            <el-form-item :label="t('cert.genCnLabel')">
              <el-input
                v-model="genConfig.common_name"
                placeholder="例如: foo.com"
                clearable
              />
            </el-form-item>
            <el-form-item :label="t('cert.genSansLabel')">
              <el-input
                v-model="genConfig.sans_str"
                placeholder="例如: *.foo.com, foo.com"
                clearable
              />
            </el-form-item>
            <el-form-item :label="t('cert.genValidityLabel')">
              <el-select v-model="genConfig.valid_days" class="w-full">
                <el-option :label="t('cert.month1')" :value="30" />
                <el-option :label="t('cert.month3')" :value="90" />
                <el-option :label="t('cert.month6')" :value="180" />
                <el-option :label="t('cert.year1')" :value="365" />
                <el-option :label="t('cert.year3')" :value="1095" />
                <el-option :label="t('cert.year10')" :value="3650" />
                <el-option :label="t('cert.year20')" :value="7300" />
              </el-select>
            </el-form-item>
            <div class="flex justify-end mt-2">
              <el-button
                type="success"
                size="small"
                :loading="genLoading"
                :icon="useRenderIcon('ri/check-line')"
                @click="handleGenerate"
              >
                {{ t('cert.genBtn') }}
              </el-button>
            </div>
          </el-form>
        </el-card>
      </re-col>

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('cert.certId')" prop="cert_id">
          <el-input
            v-model="newFormInline.cert_id"
            placeholder="例如: id-1, id-2"
            :disabled="!!newFormInline.id"
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('cert.type')" prop="type">
          <el-select
            v-model="newFormInline.type"
            class="w-full"
            placeholder="请选择证书类型"
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

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('cert.certContent')" prop="cert_content">
          <el-input
            v-model="newFormInline.cert_content"
            type="textarea"
            :rows="5"
            placeholder="-----BEGIN CERTIFICATE----- ... -----END CERTIFICATE-----"
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('cert.keyContent')" prop="key_content">
          <el-input
            v-model="newFormInline.key_content"
            type="textarea"
            :rows="5"
            placeholder="-----BEGIN PRIVATE KEY----- ... -----END PRIVATE KEY-----"
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('cert.remark')" prop="remark">
          <el-input
            v-model="newFormInline.remark"
            type="textarea"
            :rows="2"
            placeholder="选填"
            clearable
          />
        </el-form-item>
      </re-col>
    </el-row>
  </el-form>
</template>
