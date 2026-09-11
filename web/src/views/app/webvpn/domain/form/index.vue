<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { useI18n } from "vue-i18n";
import ReCol from "@/components/ReCol";
import { formRules } from "../utils/rule";
import { DomainFormProps } from "../utils/types";
import { getCertList } from "@/api/certificate";

const props = withDefaults(defineProps<DomainFormProps>(), {
  formInline: () => ({
    name: "",
    hostname: "",
    port: "443",
    tls: true,
    h2: true,
    certificate: "",
    login_url: "",
    fallback: "404",
    status: 1,
    remark: ""
  })
});

const { t } = useI18n();

const ruleFormRef = ref();

function handleTlsChange(val: boolean) {
  if (!val) {
    newFormInline.value.h2 = false;
  }
  if (ruleFormRef.value) {
    ruleFormRef.value.validateField("certificate");
  }
}


const validateCertificate = (_rule: any, _value: any, callback: any) => {
  if (newFormInline.value.tls && !newFormInline.value.certificate) {
    callback(
      new Error(
        t("webvpnDomain.certRequiredForTls", "开启 TLS 必须选择关联证书")
      )
    );
  } else {
    callback();
  }
};

const localRules = {
  ...formRules,
  certificate: [{ validator: validateCertificate, trigger: ["change", "blur"] }]
};

const newFormInline = ref(props.formInline);

watch(
  () => props.formInline,
  val => {
    newFormInline.value = val;
  },
  { deep: true }
);
const certOptions = ref<Array<{ label: string; value: string }>>([]);

async function fetchCertificates() {
  try {
    const res = await getCertList();
    if (res?.code === 0 && res?.data?.list) {
      certOptions.value = res.data.list.map((c: any) => {
        const idVal = c.CertId || c.cert_id || `id-${c.Id || c.id}`;
        const cnVal = c.SubjectCN || c.subject_cn || c.Name || c.name || idVal;
        return {
          label: `${cnVal} (${idVal})`,
          value: idVal
        };
      });
    }
  } catch (e) {}
}

onMounted(() => {
  fetchCertificates();
});

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef, newFormInline });
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :model="newFormInline"
    :rules="localRules"
    label-width="140px"
    class="space-y-6"
  >
    <!-- Section 1: 基本信息 -->
    <el-card
      shadow="never"
      class="border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-primary rounded-full" />
          <span
            class="font-bold text-(--el-text-color-primary) text-sm sm:text-base"
          >
            {{ t("webvpnDomain.basicSection", "基本信息") }}
          </span>
        </div>
      </template>

      <el-row :gutter="24">
        <!-- 1. 名称 -->
        <re-col :value="24">
          <el-form-item
            :label="t('webvpnDomain.name', '名称')"
            prop="name"
          >
            <el-input
              v-model="newFormInline.name"
              clearable
              :placeholder="
                t('webvpnDomain.namePlaceholder', '如：主校区 WebVPN 网关')
              "
            />
          </el-form-item>
        </re-col>

        <!-- 2. 泛域名 -->
        <re-col :value="16" :xs="24" :sm="16">
          <el-form-item
            :label="t('webvpnDomain.hostname', '泛域名')"
            prop="hostname"
          >
            <div class="flex flex-col w-full">
              <el-input
                v-model="newFormInline.hostname"
                clearable
                :placeholder="
                  t(
                    'webvpnDomain.hostnamePlaceholder',
                    '如：*.webvpn.example.com'
                  )
                "
              />
              <p class="text-xs/relaxed text-gray-400 mt-2">
                {{
                  t(
                    "webvpnDomain.hostnameHint",
                    "WebVPN 底座泛域名，必须以 *. 开头，例如 *.webvpn.example.com。"
                  )
                }}
              </p>
            </div>
          </el-form-item>
        </re-col>

        <!-- 3. 端口 -->
        <re-col :value="8" :xs="24" :sm="8">
          <el-form-item :label="t('webvpnDomain.port', '端口')" prop="port">
            <el-input-number
              v-model="newFormInline.port"
              :min="0"
              :max="65536"
              controls-position="right"
              :placeholder="t('webvpnDomain.portPlaceholder', '443')"
              class="!w-full"
            />
          </el-form-item>
        </re-col>

        <!-- 4. 安全协议 -->
        <re-col :value="24">
          <el-form-item
            :label="t('webvpnDomain.protocol', '安全协议')"
            :for="''"
          >
            <div class="flex items-center gap-8">
              <div class="flex items-center gap-2">
                <span class="text-sm text-gray-600 dark:text-gray-300"
                  >TLS:</span
                >
                <el-switch v-model="newFormInline.tls" @change="handleTlsChange" />
              </div>
              <div class="flex items-center gap-2">
                <span class="text-sm text-gray-600 dark:text-gray-300"
                  >HTTP/2:</span
                >
                <el-switch v-model="newFormInline.h2" />
              </div>
            </div>
          </el-form-item>
        </re-col>

        <!-- 5. SSL 证书 -->
        <re-col :value="24">
          <el-form-item
            :label="t('webvpnDomain.certificate', 'SSL 证书')"
            prop="certificate"
          >
            <div class="flex flex-col w-full">
              <el-select
                v-model="newFormInline.certificate"
                filterable
                clearable
                class="w-full"
                :placeholder="
                  t('webvpnDomain.certPlaceholder', '选择匹配的通配符 SSL 证书')
                "
              >
                <el-option
                  v-for="c in certOptions"
                  :key="c.value"
                  :label="c.label"
                  :value="c.value"
                />
              </el-select>
              <p class="text-xs/relaxed text-gray-400 mt-2">
                {{
                  t(
                    "webvpnDomain.certHint",
                    "请选择已在系统中颁发且涵盖该泛域名的通配符证书；留空时将尝试自动匹配。"
                  )
                }}
              </p>
            </div>
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>

    <!-- Section 2: 访问与安全策略 -->
    <el-card
      shadow="never"
      class="border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-primary rounded-full" />
          <span
            class="font-bold text-(--el-text-color-primary) text-sm sm:text-base"
          >
            {{ t("webvpnDomain.policySection", "访问与安全策略") }}
          </span>
        </div>
      </template>

      <el-row :gutter="24">
        <!-- 6. 认证中心地址 -->
        <re-col :value="24">
          <el-form-item
            :label="t('webvpnDomain.loginUrl', '认证中心地址')"
            prop="login_url"
          >
            <div class="flex flex-col w-full">
              <el-input
                v-model="newFormInline.login_url"
                clearable
                :placeholder="
                  t(
                    'webvpnDomain.loginUrlPlaceholder',
                    '选填，如 https://auth.example.com'
                  )
                "
              />
              <p class="text-xs/relaxed text-gray-400 mt-2">
                {{
                  t(
                    "webvpnDomain.loginUrlHint",
                    "用户未登录时跳转的认证登录页面；留空时自动关联系统内已配置的认证中心。"
                  )
                }}
              </p>
            </div>
          </el-form-item>
        </re-col>

        <!-- 7. 兜底策略 -->
        <re-col :value="24">
          <el-form-item
            :label="t('webvpnDomain.fallback', '未命中策略')"
            prop="fallback"
          >
            <div class="flex flex-col w-full">
              <el-select v-model="newFormInline.fallback" class="w-full">
                <el-option
                  value="404"
                  :label="
                    t(
                      'webvpnDomain.fallback404',
                      '404 页面阻断（推荐，严防未知请求穿透）'
                    )
                  "
                />
                <el-option
                  value="login"
                  :label="
                    t('webvpnDomain.fallbackLogin', '重定向至认证中心登录页')
                  "
                />
              </el-select>
              <p class="text-xs/relaxed text-gray-400 mt-2">
                {{
                  t(
                    "webvpnDomain.fallbackHint",
                    "当外部请求的子域名未在站点列表中注册或已被停用时的安全防护策略。"
                  )
                }}
              </p>
            </div>
          </el-form-item>
        </re-col>

        <!-- 8. 启用 -->
        <re-col :value="24">
          <el-form-item
            :label="t('webvpnDomain.status', '启用')"
            prop="status"
            :for="''"
          >
            <el-switch
              v-model="newFormInline.status"
              :active-value="1"
              :inactive-value="0"
              :active-text="t('webvpnDomain.statusEnabled', '启用')"
              :inactive-text="t('webvpnDomain.statusDisabled', '禁用')"
            />
          </el-form-item>
        </re-col>

        <!-- 9. 备注 -->
        <re-col :value="24">
          <el-form-item :label="t('webvpnDomain.remark', '备注')" prop="remark">
            <el-input
              v-model="newFormInline.remark"
              type="textarea"
              :rows="2"
              :placeholder="
                t(
                  'webvpnDomain.remarkPlaceholder',
                  '选填，关于该 WebVPN 网关基础域的说明'
                )
              "
            />
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>
  </el-form>
</template>
