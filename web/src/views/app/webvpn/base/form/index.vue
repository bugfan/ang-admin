<script setup lang="ts">
import { ref, onMounted, watch, nextTick, computed } from "vue";
import { useI18n } from "vue-i18n";
import ReCol from "@/components/ReCol";
import { formRules } from "../utils/rule";
import { BaseFormProps } from "../utils/types";
import { getCertList } from "@/api/certificate";
import { getAuthList } from "@/api/auth-config";
import { getRuleList } from "@/api/rule";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import Rank from "~icons/ep/rank";
import Delete from "~icons/ep/delete";
import Sortable from "sortablejs";
import { message } from "@/utils/message";

const props = withDefaults(defineProps<BaseFormProps>(), {
  formInline: () => ({
    name: "",
    hostname: "",
    port: "443",
    tls: true,
    h2: true,
    certificate: "",
    auth_id: null,
    rules: "[]",
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
        t("webvpnBase.certRequiredForTls", "开启 TLS 必须选择关联证书")
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

interface PipelineRuleItem {
  id: string | number;
  type: "rule" | "auth";
  name: string;
}

const pipelineList = ref<PipelineRuleItem[]>([]);
const pipelineGroupRef = ref<HTMLElement | null>(null);
const selectedCandidate = ref<string>("");

const hasAuthInPipeline = computed(() =>
  pipelineList.value.some(p => p.type === "auth")
);

const certOptions = ref<Array<{ label: string; value: string }>>([]);
const authOptions = ref<Array<{ name: string; id: number }>>([]);
const availableRules = ref<Array<{ name: string; id: string }>>([]);

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

async function fetchAuthOptions() {
  try {
    const res = await getAuthList();
    if (res?.code === 0 && res?.data?.list) {
      authOptions.value = res.data.list.map((a: any) => ({
        name: a.Name || a.name,
        id: a.Id || a.id
      }));
    }
  } catch (e) {}
}

async function fetchRules() {
  try {
    const res = await getRuleList();
    if (res?.code === 0 && res?.data?.list) {
      availableRules.value = res.data.list.map((r: any) => ({
        name: r.name || r.Name,
        id: r.name || r.Name
      }));
    }
  } catch (e) {}
}

function initPipeline() {
  pipelineList.value = [];
  try {
    if (newFormInline.value.rules) {
      const parsed =
        typeof newFormInline.value.rules === "string"
          ? JSON.parse(newFormInline.value.rules)
          : newFormInline.value.rules;
      if (Array.isArray(parsed)) {
        pipelineList.value = parsed.map((item: any) => {
          if (typeof item === "string") {
            if (item.startsWith("auth:")) {
              const aid = Number(item.replace("auth:", ""));
              const aObj = authOptions.value.find(a => a.id === aid);
              return {
                type: "auth",
                id: aid,
                name: aObj ? aObj.name : `${t("webvpnBase.typeAuth", "认证流")} #${aid}`
              };
            } else {
              const rName = item.replace("rule:", "");
              return {
                type: "rule",
                id: rName,
                name: rName
              };
            }
          }
          let name = item.name;
          if (!name) {
            if (item.type === "auth") {
              const aObj = authOptions.value.find(a => a.id === Number(item.id));
              name = aObj ? aObj.name : `${t("webvpnBase.typeAuth", "认证流")} #${item.id}`;
            } else {
              name = String(item.id);
            }
          }
          return {
            type: item.type || "rule",
            id: item.id,
            name: name
          };
        });
      }
    }
  } catch (e) {}

  // 兼容逻辑：若历史数据未配 rules 但有 auth_id，自动补全为认证流节点
  if (!pipelineList.value.some(p => p.type === "auth") && newFormInline.value.auth_id) {
    const aid = Number(newFormInline.value.auth_id);
    const aObj = authOptions.value.find(a => a.id === aid);
    pipelineList.value.push({
      type: "auth",
      id: aid,
      name: aObj ? aObj.name : `${t("webvpnBase.typeAuth", "认证流")} #${aid}`
    });
  }

  syncPipelineToForm();
}

function syncPipelineToForm() {
  newFormInline.value.rules = JSON.stringify(
    pipelineList.value.map(item => ({
      type: item.type,
      id: item.id,
      name: item.name
    }))
  );
  const authItem = pipelineList.value.find(item => item.type === "auth");
  newFormInline.value.auth_id = authItem ? Number(authItem.id) : null;
}

function initSortablePipeline() {
  if (pipelineGroupRef.value) {
    Sortable.create(pipelineGroupRef.value, {
      handle: ".drag-handle",
      animation: 150,
      onEnd: (evt: any) => {
        const { oldIndex, newIndex } = evt;
        if (
          oldIndex !== undefined &&
          newIndex !== undefined &&
          oldIndex !== newIndex
        ) {
          const item = pipelineList.value.splice(oldIndex, 1)[0];
          pipelineList.value.splice(newIndex, 0, item);
          syncPipelineToForm();
        }
      }
    });
  }
}

function handleAddCandidate(val: string) {
  if (!val) return;
  if (val.startsWith("auth:")) {
    const aid = Number(val.replace("auth:", ""));
    if (pipelineList.value.some(p => p.type === "auth" && Number(p.id) === aid)) {
      message(t("webvpnBase.authAlreadyAdded", "该认证流已在流水线中"), {
        type: "warning"
      });
      selectedCandidate.value = "";
      return;
    }
    const aObj = authOptions.value.find(a => a.id === aid);
    pipelineList.value.push({
      type: "auth",
      id: aid,
      name: aObj ? aObj.name : `${t("webvpnBase.typeAuth", "认证流")} #${aid}`
    });
  } else if (val.startsWith("rule:")) {
    const rName = val.replace("rule:", "");
    if (pipelineList.value.some(p => p.type === "rule" && String(p.id) === rName)) {
      message(t("webvpnBase.ruleAlreadyAdded", "该规则已在流水线中"), {
        type: "warning"
      });
      selectedCandidate.value = "";
      return;
    }
    pipelineList.value.push({
      type: "rule",
      id: rName,
      name: rName
    });
  }
  selectedCandidate.value = "";
  syncPipelineToForm();
}

function removePipelineItem(index: number) {
  pipelineList.value.splice(index, 1);
  syncPipelineToForm();
}

watch(
  () => props.formInline,
  val => {
    newFormInline.value = val;
    initPipeline();
  },
  { deep: true }
);

onMounted(async () => {
  fetchCertificates();
  await Promise.all([fetchAuthOptions(), fetchRules()]);
  initPipeline();
  nextTick(() => {
    initSortablePipeline();
  });
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
            {{ t("webvpnBase.basicSection", "基本信息") }}
          </span>
        </div>
      </template>

      <el-row :gutter="24">
        <!-- 1. 名称 -->
        <re-col :value="24">
          <el-form-item
            :label="t('webvpnBase.name', '名称')"
            prop="name"
          >
            <el-input
              v-model="newFormInline.name"
              clearable
              :placeholder="
                t('webvpnBase.namePlaceholder', '如：主校区 WebVPN 网关')
              "
            />
          </el-form-item>
        </re-col>

        <!-- 2. 泛域名 -->
        <re-col :value="16" :xs="24" :sm="16">
          <el-form-item
            :label="t('webvpnBase.hostname', '泛域名')"
            prop="hostname"
          >
            <div class="flex flex-col w-full">
              <el-input
                v-model="newFormInline.hostname"
                clearable
                :placeholder="
                  t(
                    'webvpnBase.hostnamePlaceholder',
                    '如：*.webvpn.example.com'
                  )
                "
              />
              <p class="text-xs/relaxed text-gray-400 mt-2">
                {{
                  t(
                    "webvpnBase.hostnameHint",
                    "WebVPN 底座泛域名，必须以 *. 开头，例如 *.webvpn.example.com。"
                  )
                }}
              </p>
            </div>
          </el-form-item>
        </re-col>

        <!-- 3. 端口 -->
        <re-col :value="8" :xs="24" :sm="8">
          <el-form-item :label="t('webvpnBase.port', '端口')" prop="port">
            <el-input-number
              v-model="newFormInline.port"
              :min="0"
              :max="65536"
              controls-position="right"
              :placeholder="t('webvpnBase.portPlaceholder', '443')"
              class="!w-full"
            />
          </el-form-item>
        </re-col>

        <!-- 4. 安全协议 -->
        <re-col :value="24">
          <el-form-item
            :label="t('webvpnBase.protocol', '安全协议')"
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
            :label="t('webvpnBase.certificate', 'SSL 证书')"
            prop="certificate"
          >
            <div class="flex flex-col w-full">
              <el-select
                v-model="newFormInline.certificate"
                filterable
                clearable
                class="w-full"
                :placeholder="
                  t('webvpnBase.certPlaceholder', '选择匹配的通配符 SSL 证书')
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
                    "webvpnBase.certHint",
                    "请选择已在系统中颁发且涵盖该泛域名的通配符证书；留空时将尝试自动匹配。"
                  )
                }}
              </p>
            </div>
          </el-form-item>
        </re-col>

        <!-- 8. 启用 -->
        <re-col :value="24">
          <el-form-item
            :label="t('webvpnBase.status', '启用')"
            prop="status"
            :for="''"
          >
            <el-switch
              v-model="newFormInline.status"
              :active-value="1"
              :inactive-value="0"
              :active-text="t('webvpnBase.statusEnabled', '启用')"
              :inactive-text="t('webvpnBase.statusDisabled', '禁用')"
            />
          </el-form-item>
        </re-col>

        <!-- 9. 备注 -->
        <re-col :value="24">
          <el-form-item :label="t('webvpnBase.remark', '备注')" prop="remark">
            <el-input
              v-model="newFormInline.remark"
              type="textarea"
              :rows="2"
              :placeholder="
                t(
                  'webvpnBase.remarkPlaceholder',
                  '选填，关于该 WebVPN 网关基础域的说明'
                )
              "
            />
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>

    <!-- Section 2: 访问与安全策略 (规则流水线) -->
    <el-card
      shadow="never"
      class="border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <div class="w-1.5 h-4 bg-primary rounded-full" />
            <span
              class="font-bold text-(--el-text-color-primary) text-sm sm:text-base"
            >
              {{ t("webvpnBase.policySection", "访问与安全策略") }}
            </span>
          </div>
          <span class="text-xs text-gray-400">
            {{
              t(
                "webvpnBase.pipelineDesc",
                "自上而下依次执行，可按住左侧 ☰ 拖拽调整顺序"
              )
            }}
          </span>
        </div>
      </template>

      <el-row :gutter="24">
        <!-- 规则流水线 -->
        <re-col :value="24">
          <el-form-item :label="t('webvpnBase.pipelineTitle', '规则流水线')">
            <div class="w-full flex flex-col space-y-3">
              <!-- 添加操作栏 -->
              <div class="flex items-center space-x-2">
                <el-select
                  v-model="selectedCandidate"
                  filterable
                  clearable
                  class="flex-1"
                  :placeholder="
                    t(
                      'webvpnBase.selectRuleOrAuthPlaceholder',
                      '选择要加入流水线的规则或认证流...'
                    )
                  "
                  @change="handleAddCandidate"
                >
                  <el-option-group
                    :label="t('webvpnBase.groupAuth', '🔐 认证流 (SSO 鉴权)')"
                  >
                    <el-option
                      v-for="a in authOptions"
                      :key="'auth:' + a.id"
                      :label="a.name"
                      :value="'auth:' + a.id"
                      :disabled="
                        pipelineList.some(
                          p => p.type === 'auth' && Number(p.id) === a.id
                        )
                      "
                    />
                  </el-option-group>
                  <el-option-group
                    :label="t('webvpnBase.groupRule', '🛡️ 规则管理 (中间件)')"
                  >
                    <el-option
                      v-for="r in availableRules"
                      :key="'rule:' + r.name"
                      :label="r.name"
                      :value="'rule:' + r.name"
                      :disabled="
                        pipelineList.some(
                          p => p.type === 'rule' && String(p.id) === r.name
                        )
                      "
                    />
                  </el-option-group>
                </el-select>
              </div>

              <!-- 拖拽列表容器 -->
              <div
                ref="pipelineGroupRef"
                class="min-h-[64px] border border-dashed border-(--el-border-color) rounded-xl p-2.5 bg-(--el-fill-color-blank) flex flex-col space-y-2"
              >
                <div
                  v-for="(item, idx) in pipelineList"
                  :key="item.type + ':' + item.id"
                  class="flex items-center justify-between p-2.5 bg-(--el-fill-color-light) border border-(--el-border-color-lighter) rounded-lg hover:border-primary transition-all shadow-2xs select-none"
                >
                  <div class="flex items-center space-x-2.5">
                    <div
                      class="drag-handle cursor-grab active:cursor-grabbing text-gray-400 hover:text-primary p-1"
                    >
                      <component :is="useRenderIcon(Rank)" class="w-4 h-4" />
                    </div>
                    <el-tag
                      :type="item.type === 'auth' ? 'warning' : 'primary'"
                      size="small"
                      effect="plain"
                      class="font-mono font-medium"
                    >
                      {{
                        item.type === "auth"
                          ? t("webvpnBase.typeAuth", "认证流")
                          : t("webvpnBase.typeRule", "规则中间件")
                      }}
                    </el-tag>
                    <span
                      class="font-bold text-sm text-(--el-text-color-primary)"
                      >{{ item.name }}</span
                    >
                    <span
                      v-if="item.type === 'auth'"
                      class="text-xs text-gray-400 hidden sm:inline"
                    >
                      ({{
                        t(
                          "webvpnBase.authFlowItemHint",
                          "WebVPN 访问凭证核验与 SSO 门户"
                        )
                      }})
                    </span>
                  </div>
                  <div class="flex items-center space-x-1">
                    <el-button
                      link
                      type="danger"
                      size="small"
                      :icon="useRenderIcon(Delete)"
                      @click="removePipelineItem(idx)"
                    >
                      {{ t("common.delete", "删除") }}
                    </el-button>
                  </div>
                </div>

                <div
                  v-if="pipelineList.length === 0"
                  class="py-6 text-center text-gray-400 text-xs flex flex-col items-center justify-center space-y-1"
                >
                  <span>{{
                    t(
                      "webvpnBase.emptyPipeline",
                      "暂无规则流水线，请在上方选择并添加规则或认证流"
                    )
                  }}</span>
                </div>
              </div>

              <p class="text-xs/relaxed text-gray-400 mt-1">
                {{
                  t(
                    "webvpnBase.pipelineHint",
                    "请求到达基座后将自上而下依次执行挂载的中间件。如需身份验证，请在流水线中添加认证流；认证流之前的规则将在鉴权前执行（如 IP黑名单、DDoS限流防刷），之后的规则将在鉴权通过后执行。"
                  )
                }}
              </p>

              <el-alert
                v-if="!hasAuthInPipeline && (Number(newFormInline.protected_site_count || 0) > 0 || Number(newFormInline.site_count || 0) > 0)"
                type="warning"
                show-icon
                :closable="false"
                class="mt-2"
                :title="t('webvpnBase.noAuthWarningTitle', '提示：当前基座未配置认证流')"
                :description="
                  t(
                    'webvpnBase.noAuthWarningDesc',
                    '检测到该基座下已有受保护的站点。若规则流水线中未添加认证流，这些站点将因缺少登录入口而无法正常登录访问（显示 401 阻断）。'
                  )
                "
              />
            </div>
          </el-form-item>
        </re-col>

        <!-- 7. 兜底策略 -->
        <re-col :value="24">
          <el-form-item
            :label="t('webvpnBase.fallback', '未命中策略')"
            prop="fallback"
          >
            <div class="flex flex-col w-full">
              <el-select v-model="newFormInline.fallback" class="w-full">
                <el-option
                  value="404"
                  :label="
                    t(
                      'webvpnBase.fallback404',
                      '404 页面阻断（推荐，严防未知请求穿透）'
                    )
                  "
                />
                <el-option
                  value="login"
                  :label="
                    t('webvpnBase.fallbackLogin', '重定向至认证中心登录页')
                  "
                />
              </el-select>
              <p class="text-xs/relaxed text-gray-400 mt-2">
                {{
                  t(
                    "webvpnBase.fallbackHint",
                    "当外部请求的子域名未在站点列表中注册或已被停用时的安全防护策略。"
                  )
                }}
              </p>
            </div>
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>
  </el-form>
</template>
