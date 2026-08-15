<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { useRouter } from "vue-router";
import ReCol from "@/components/ReCol";
import { useI18n } from "vue-i18n";
import { getCertList } from "@/api/certificate";
import { closeAllDialog } from "@/components/ReDialog";

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "新增",
    id: undefined,
    type: "TLS-TUNNEL",
    port: "",
    sni: "",
    certificate: "",
    remark: ""
  })
});

const router = useRouter();
const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const { t } = useI18n();

const certOptions = ref<Array<{ label: string; value: string; cn: string }>>([]);
const certLoading = ref(false);

const typeOptions = [
  { label: "TLS-TUNNEL", value: "TLS-TUNNEL" },
  { label: "QUIC-TUNNEL", value: "QUIC-TUNNEL" }
];

const formRules = reactive({
  type: [
    { required: true, message: () => t("tunnel.typeRequired"), trigger: "change" }
  ],
  port: [
    { required: true, message: () => t("tunnel.portRequired"), trigger: "blur" },
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

async function fetchCertificates(query = "") {
  certLoading.value = true;
  const { code, data } = await getCertList({ cert_id: query });
  certLoading.value = false;
  if (code === 0 && data?.list) {
    certOptions.value = data.list.map((item: any) => ({
      label: item.CertId || item.cert_id,
      value: item.CertId || item.cert_id,
      cn: item.SubjectCN || item.subject_cn || ""
    }));

    // If current selected certificate value exists but not in searched list, append it
    const curCert = newFormInline.value.certificate;
    if (curCert && !certOptions.value.some(c => c.value === curCert)) {
      certOptions.value.unshift({
        label: curCert,
        value: curCert,
        cn: ""
      });
    }

    // Auto-sync SNI if certificate is set but SNI is empty
    if (newFormInline.value.certificate) {
      handleCertChange(newFormInline.value.certificate);
    }
  }
}

function handleCertChange(val: string) {
  if (!val) {
    // If cert is cleared, clear SNI
    newFormInline.value.sni = "";
    return;
  }
  const matched = certOptions.value.find(item => item.value === val);
  if (matched && matched.cn) {
    newFormInline.value.sni = matched.cn;
  }
}

function goToCertPage() {
  closeAllDialog();
  router.push("/app/cert");
}

onMounted(() => {
  fetchCertificates();
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

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('tunnel.certificate')" prop="certificate">
          <div class="w-full">
            <!-- 选中证书后自动在上方展示关联的 SNI (Common Name) -->
            <div
              v-if="newFormInline.sni"
              class="mb-2 flex items-center text-xs text-gray-600 dark:text-gray-300"
            >
              <span class="mr-2 font-medium">{{ t('tunnel.associatedSni') }}</span>
              <el-tag size="small" type="success" effect="plain" class="font-mono font-bold">
                {{ newFormInline.sni }}
              </el-tag>
            </div>

            <div class="flex items-center w-full space-x-2">
              <el-select
                v-model="newFormInline.certificate"
                filterable
                clearable
                remote
                :remote-method="fetchCertificates"
                :loading="certLoading"
                class="flex-1"
                :placeholder="t('tunnel.selectCertPlaceholder')"
                @change="handleCertChange"
              >
                <el-option
                  v-for="item in certOptions"
                  :key="item.value"
                  :label="item.value"
                  :value="item.value"
                >
                  <div class="flex justify-between items-center w-full pr-2">
                    <span class="font-medium">{{ item.value }}</span>
                    <span v-if="item.cn" class="text-xs text-gray-400 font-mono ml-4">{{ item.cn }}</span>
                  </div>
                </el-option>
              </el-select>

              <el-button
                v-if="certOptions.length === 0 && !certLoading"
                type="primary"
                link
                class="whitespace-nowrap font-medium"
                @click="goToCertPage"
              >
                {{ t('tunnel.goCertPage') }} &gt;
              </el-button>
            </div>
          </div>
        </el-form-item>
      </re-col>

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('tunnel.port')" prop="port">
          <el-input
            v-model="newFormInline.port"
            placeholder="443"
            clearable
          />
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
