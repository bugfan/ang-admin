<script setup lang="ts">
import { ref, onMounted } from "vue";
import { deviceDetection } from "@pureadmin/utils";
import ReCol from "@/components/ReCol";
import { FormProps } from "../utils/types";
import { useI18n } from "vue-i18n";
import { getAuthMethodList, type AuthMethodItem } from "@/api/auth-method";
import draggable from "vuedraggable";
import { Rank } from "@element-plus/icons-vue";
import Delete from "~icons/ep/delete";
import Plus from "~icons/ep/plus";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "添加认证配置",
    id: 0,
    name: "",
    auth_method_ids: "[]",
    token_name: "_angt",
    portal_url: "",
    token_expire: 86400,
    remark: ""
  })
});

const { t } = useI18n();
const formRules = {
  name: [
    {
      required: true,
      message: t("identity.authConfigNamePlaceholder", "请输入认证名称"),
      trigger: "blur"
    }
  ]
};

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const authMethodOptions = ref<AuthMethodItem[]>([]);
const selectedMethods = ref<number[]>([]);

onMounted(async () => {
  try {
    const res = await getAuthMethodList();
    if (res && res.code === 0 && res.data && res.data.list) {
      authMethodOptions.value = res.data.list;
    }
    selectedMethods.value = JSON.parse(
      newFormInline.value.auth_method_ids || "[]"
    );
  } catch (e) {
    selectedMethods.value = [];
  }
});

function syncMethodChange() {
  newFormInline.value.auth_method_ids = JSON.stringify(selectedMethods.value);
}

function handleAddMethod() {
  selectedMethods.value.push(undefined as any);
  syncMethodChange();
}

function handleRemoveMethod(index: number) {
  selectedMethods.value.splice(index, 1);
  syncMethodChange();
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
    :label-position="deviceDetection() ? 'top' : 'right'"
    label-width="auto"
    class="auth-form p-1 sm:px-2 space-y-4"
  >
    <!-- Section 1: 基本信息 -->
    <el-card
      shadow="never"
      class="border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-primary rounded-full" />
          <span class="font-bold text-(--el-text-color-primary) text-sm sm:text-base">
            {{ t("identity.baseInfo", "基本信息") }}
          </span>
        </div>
      </template>

      <el-row :gutter="16">
        <re-col :value="12" :xs="24">
          <el-form-item :label="t('identity.authConfigName', '名称')" prop="name">
            <el-input
              v-model="newFormInline.name"
              clearable
              :placeholder="t('identity.authConfigNamePlaceholder', '如：企业内网认证策略')"
            />
          </el-form-item>
        </re-col>
        
        <re-col :value="12" :xs="24">
          <el-form-item :label="t('identity.remark', '备注')" prop="remark">
            <el-input
              v-model="newFormInline.remark"
              clearable
              :placeholder="t('identity.remarkPlaceholder', '请输入备注信息')"
            />
          </el-form-item>
        </re-col>

        <re-col :value="24" :xs="24">
          <el-form-item :label="t('identity.portalUrl', '登录入口')" prop="portal_url">
            <el-input
              v-model="newFormInline.portal_url"
              clearable
              :placeholder="t('identity.portalUrlPlaceholder', '输入 Portal 跳转地址，如：https://auth.example.com/login')"
            />
            <div class="w-full text-xs text-(--el-text-color-secondary) mt-1.5">
              提示：未登录用户访问受保护站点时，将 302 重定向至此登录入口进行拦截。
            </div>
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>

    <!-- Section 2: 多因子认证流 -->
    <el-card
      shadow="never"
      class="border-(--el-border-color-lighter)! rounded-xl"
    >
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-success rounded-full" />
          <span class="font-bold text-(--el-text-color-primary) text-sm sm:text-base">
            {{ t("identity.authFlow", "多因子认证流 (MFA)") }}
          </span>
        </div>
      </template>

      <div class="w-full bg-gray-50 dark:bg-gray-800/40 rounded-md p-4 border border-(--el-border-color-lighter)">
        <draggable
          v-model="selectedMethods"
          :item-key="(item, index) => index"
          handle=".drag-handle"
          ghost-class="ghost"
          class="space-y-3 mb-4"
          @change="syncMethodChange"
        >
          <template #item="{ element, index }">
            <div class="item-row">
              <!-- Left: drag + index -->
              <div class="item-row-left">
                <span class="drag-handle" title="按住拖拽以排序">
                  <el-icon class="text-lg"><Rank /></el-icon>
                </span>
                <el-tag
                  size="small"
                  :type="index === 0 ? 'primary' : 'info'"
                  effect="plain"
                  class="font-mono font-bold w-7 sm:w-8 text-center shrink-0"
                >
                  {{ index + 1 }}
                </el-tag>
              </div>

              <!-- Middle: selector -->
              <div class="item-row-summary">
                <el-select
                  v-model="selectedMethods[index]"
                  class="w-full"
                  placeholder="请选择此步骤的认证方式"
                  @change="syncMethodChange"
                >
                  <el-option
                    v-for="a in authMethodOptions"
                    :key="a.id"
                    :label="a.name"
                    :value="a.id"
                  >
                    <span style="float: left">{{ a.name }}</span>
                    <span style="float: right; color: var(--el-text-color-secondary); font-size: 12px">
                      {{ a.type }}
                    </span>
                  </el-option>
                </el-select>
              </div>

              <!-- Right: actions -->
              <div class="item-row-actions">
                <el-button
                  type="danger"
                  link
                  :icon="useRenderIcon(Delete)"
                  @click="handleRemoveMethod(index)"
                >
                  移除
                </el-button>
              </div>
            </div>
          </template>
        </draggable>
        
        <el-button type="primary" plain :icon="useRenderIcon(Plus)" @click="handleAddMethod" class="w-full border-dashed! py-5">
          {{ t('identity.addAuthMethod', '添加认证步骤') }}
        </el-button>

        <div class="text-xs text-gray-500 dark:text-gray-400 mt-3 flex items-start gap-1">
          <i class="el-icon-info mt-0.5"></i> 
          <span>{{ t('identity.dragHint', '拖拽可调整验证顺序。系统将严格按照列表从上到下的顺序要求用户逐一完成认证（支持串联组装任意多因子验证）。') }}</span>
        </div>
      </div>
    </el-card>

    </el-form>
</template>

<style scoped>
/* ── Item row ─────────────────────────────────────────────────────── */
.item-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
  transition: border-color 0.18s;
  flex-wrap: wrap;
}
.item-row:hover {
  border-color: var(--el-color-primary-light-5);
}

.item-row-left {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.drag-handle {
  cursor: grab;
  color: var(--el-text-color-placeholder);
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  transition: color 0.15s;
}
.drag-handle:hover {
  color: var(--el-color-primary);
}
.drag-handle:active {
  cursor: grabbing;
}

.item-row-summary {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  flex-wrap: wrap;
}

.item-row-actions {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
}

/* drag ghost style */
:deep(.drag-ghost) {
  opacity: 0.4;
  background: var(--el-color-primary-light-9) !important;
}

@media (max-width: 640px) {
  .item-row {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
  .item-row-left {
    justify-content: space-between;
  }
  .item-row-actions {
    justify-content: flex-end;
    border-top: 1px dashed var(--el-border-color-lighter);
    padding-top: 6px;
  }
}
</style>
