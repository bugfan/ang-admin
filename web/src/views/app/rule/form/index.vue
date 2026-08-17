<script setup lang="ts">
import { ref, reactive, watch, computed } from "vue";
import ReCol from "@/components/ReCol";
import { useI18n } from "vue-i18n";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import draggable from "vuedraggable";

import AddFill from "~icons/ri/add-line";
import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import ArrowUp from "~icons/ep/arrow-up";
import ArrowDown from "~icons/ep/arrow-down";
import DragIcon from "~icons/ri/drag-move-2-line";

interface RuleItemConfig {
  id: string;
  matcherName: string;
  ipAddressText: string;
  httpIpAddressText: string;
  actionName: string;
  resetContent: string;
}

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "新增规则",
    id: undefined,
    name: "",
    items: JSON.stringify([
      {
        Matcher: { Name: "ip_matcher", Config: { Address: ["127.0.0.1"] } },
        Action: { Name: "reset_conn_action", Config: { Content: "Connection reset by rule" } }
      }
    ], null, 2),
    remark: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const { t } = useI18n();

// Supported Matcher options
const matcherOptions = [
  {
    label: "TCP / L4 IP 匹配器 (ip_matcher)",
    value: "ip_matcher",
    tag: "L4 传输层",
    tagType: "primary"
  },
  {
    label: "HTTP Proxy IP 匹配器 (http_ip_matcher)",
    value: "http_ip_matcher",
    tag: "HTTP 应用层",
    tagType: "success"
  }
];

// Supported Action options
const actionOptions = [
  {
    label: "TCP 终止并重置连接 (reset_conn_action)",
    value: "reset_conn_action",
    tag: "L4 传输层",
    tagType: "danger"
  },
  {
    label: "HTTP 隐藏 Server 版本 (hide_version_action)",
    value: "hide_version_action",
    tag: "HTTP 应用层",
    tagType: "warning"
  }
];

// Item List
const itemList = ref<RuleItemConfig[]>([]);

// Sub-dialog state for editing/adding a single RuleItem
const itemDialogVisible = ref(false);
const editingIndex = ref<number | null>(null);
const itemForm = reactive<RuleItemConfig>({
  id: "",
  matcherName: "ip_matcher",
  ipAddressText: "127.0.0.1",
  httpIpAddressText: "192.168.1.0/24",
  actionName: "reset_conn_action",
  resetContent: "Connection reset by rule"
});

onInitForm();

function createDefaultItem(): RuleItemConfig {
  return {
    id: String(Date.now() + Math.random()),
    matcherName: "ip_matcher",
    ipAddressText: "127.0.0.1",
    httpIpAddressText: "192.168.1.0/24",
    actionName: "reset_conn_action",
    resetContent: "Connection reset by rule"
  };
}

function onInitForm() {
  itemList.value = [];
  try {
    if (newFormInline.value.items) {
      const parsed = typeof newFormInline.value.items === "string"
        ? JSON.parse(newFormInline.value.items)
        : newFormInline.value.items;

      if (Array.isArray(parsed) && parsed.length > 0) {
        for (const item of parsed) {
          const mObj = item?.Matcher || item?.matcher || {};
          const aObj = item?.Action || item?.action || {};
          const mName = mObj?.Name || mObj?.name || "ip_matcher";
          const mCfg = mObj?.Config || mObj?.config || {};
          const aName = aObj?.Name || aObj?.name || "reset_conn_action";
          const aCfg = aObj?.Config || aObj?.config || {};

          let ipText = "127.0.0.1";
          let httpIpText = "192.168.1.0/24";
          if (mName === "ip_matcher" && Array.isArray(mCfg.Address)) {
            ipText = mCfg.Address.join("\n");
          }
          if (mName === "http_ip_matcher" && Array.isArray(mCfg.Address)) {
            httpIpText = mCfg.Address.join("\n");
          }

          itemList.value.push({
            id: String(Math.random()),
            matcherName: mName === "http_ip_matcher" ? "http_ip_matcher" : "ip_matcher",
            ipAddressText: ipText,
            httpIpAddressText: httpIpText,
            actionName: aName === "hide_version_action" ? "hide_version_action" : "reset_conn_action",
            resetContent: aCfg.Content ?? "Connection reset by rule"
          });
        }
      }
    }
  } catch (e) {}

  if (itemList.value.length === 0) {
    itemList.value.push(createDefaultItem());
  }

  syncToFormJSON();
}

function openAddItemDialog() {
  editingIndex.value = null;
  const def = createDefaultItem();
  itemForm.id = def.id;
  itemForm.matcherName = def.matcherName;
  itemForm.ipAddressText = def.ipAddressText;
  itemForm.httpIpAddressText = def.httpIpAddressText;
  itemForm.actionName = def.actionName;
  itemForm.resetContent = def.resetContent;
  itemDialogVisible.value = true;
}

function openEditItemDialog(index: number) {
  editingIndex.value = index;
  const cur = itemList.value[index];
  itemForm.id = cur.id;
  itemForm.matcherName = cur.matcherName;
  itemForm.ipAddressText = cur.ipAddressText;
  itemForm.httpIpAddressText = cur.httpIpAddressText;
  itemForm.actionName = cur.actionName;
  itemForm.resetContent = cur.resetContent;
  itemDialogVisible.value = true;
}

function saveItemFromDialog() {
  const target: RuleItemConfig = {
    id: itemForm.id || String(Math.random()),
    matcherName: itemForm.matcherName,
    ipAddressText: itemForm.ipAddressText,
    httpIpAddressText: itemForm.httpIpAddressText,
    actionName: itemForm.actionName,
    resetContent: itemForm.resetContent
  };

  if (editingIndex.value !== null && editingIndex.value >= 0) {
    itemList.value[editingIndex.value] = target;
  } else {
    itemList.value.push(target);
  }

  itemDialogVisible.value = false;
  syncToFormJSON();
}

function removeItem(idx: number) {
  if (itemList.value.length <= 1) {
    return;
  }
  itemList.value.splice(idx, 1);
  syncToFormJSON();
}

function moveItemUp(idx: number) {
  if (idx <= 0) return;
  const temp = itemList.value[idx];
  itemList.value[idx] = itemList.value[idx - 1];
  itemList.value[idx - 1] = temp;
  syncToFormJSON();
}

function moveItemDown(idx: number) {
  if (idx >= itemList.value.length - 1) return;
  const temp = itemList.value[idx];
  itemList.value[idx] = itemList.value[idx + 1];
  itemList.value[idx + 1] = temp;
  syncToFormJSON();
}

function syncToFormJSON() {
  const result: any[] = [];
  for (const item of itemList.value) {
    let matcherObj: any = {};
    if (item.matcherName === "http_ip_matcher") {
      const addrs = item.httpIpAddressText.split("\n").map(s => s.trim()).filter(Boolean);
      matcherObj = { Name: "http_ip_matcher", Config: { Address: addrs } };
    } else {
      const addrs = item.ipAddressText.split("\n").map(s => s.trim()).filter(Boolean);
      matcherObj = { Name: "ip_matcher", Config: { Address: addrs } };
    }

    let actionObj: any = {};
    if (item.actionName === "hide_version_action") {
      actionObj = { Name: "hide_version_action", Config: {} };
    } else {
      actionObj = { Name: "reset_conn_action", Config: { Content: item.resetContent } };
    }

    result.push({
      Matcher: matcherObj,
      Action: actionObj
    });
  }

  newFormInline.value.items = JSON.stringify(result, null, 2);
}

function getMatcherSummary(item: RuleItemConfig): { name: string; tagType: string; summary: string } {
  if (item.matcherName === "http_ip_matcher") {
    const lines = item.httpIpAddressText.split("\n").filter(s => s.trim());
    return {
      name: "http_ip_matcher",
      tagType: "success",
      summary: lines.length > 0 ? `${lines[0]}${lines.length > 1 ? ` 等 ${lines.length} 项` : ""}` : "未配置 IP"
    };
  }
  const lines = item.ipAddressText.split("\n").filter(s => s.trim());
  return {
    name: "ip_matcher",
    tagType: "primary",
    summary: lines.length > 0 ? `${lines[0]}${lines.length > 1 ? ` 等 ${lines.length} 项` : ""}` : "未配置 IP"
  };
}

function getActionSummary(item: RuleItemConfig): { name: string; tagType: string; summary: string } {
  if (item.actionName === "hide_version_action") {
    return {
      name: "hide_version_action",
      tagType: "warning",
      summary: "隐藏 Server 版本号"
    };
  }
  return {
    name: "reset_conn_action",
    tagType: "danger",
    summary: item.resetContent ? `重写消息: ${item.resetContent}` : "终止连接"
  };
}

watch(itemList, () => syncToFormJSON(), { deep: true });

const rules = reactive({
  name: [{ required: true, message: "请输入规则组名称", trigger: "blur" }]
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
    :rules="rules"
    label-width="110px"
    class="rule-form px-1 py-1"
  >
    <!-- Section 1: Basic Info -->
    <el-card shadow="never" class="mb-4 !border-[var(--el-border-color-lighter)] rounded-lg">
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1 h-3.5 bg-[var(--el-color-primary)] rounded-full"></div>
          <span class="font-bold text-[var(--el-text-color-primary)] text-sm">1. 规则组基本属性</span>
        </div>
      </template>
      <el-row :gutter="20">
        <re-col :value="12" :xs="24">
          <el-form-item label="规则组名称" prop="name">
            <el-input
              v-model="newFormInline.name"
              placeholder="请输入规则组名称 (如: 内网安全与访问控制策略)"
              clearable
            />
          </el-form-item>
        </re-col>
        <re-col :value="12" :xs="24">
          <el-form-item label="备注说明" prop="remark">
            <el-input
              v-model="newFormInline.remark"
              placeholder="选填"
              clearable
            />
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>

    <!-- Section 2: Compact Item List with Scrollbar & Drag-and-Drop -->
    <el-card shadow="never" class="mb-4 !border-[var(--el-border-color-lighter)] rounded-lg">
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <div class="w-1 h-3.5 bg-[var(--el-color-primary)] rounded-full"></div>
            <span class="font-bold text-[var(--el-text-color-primary)] text-sm">2. 规则条目清单</span>
            <el-tag size="small" type="primary" effect="plain" class="font-mono">共 {{ itemList.length }} 条</el-tag>
            <span class="text-xs text-[var(--el-text-color-secondary)] hidden sm:inline-block">(按住拖拽或点击微调图标排序)</span>
          </div>
          <el-button
            type="primary"
            size="small"
            :icon="useRenderIcon(AddFill)"
            @click="openAddItemDialog"
          >
            添加规则条目
          </el-button>
        </div>
      </template>

      <!-- Sleek Custom Scrollbar Container -->
      <el-scrollbar max-height="280px" class="item-scrollbar pr-1">
        <draggable
          v-model="itemList"
          item-key="id"
          handle=".drag-handle"
          ghost-class="opacity-40"
          class="space-y-2 pr-1"
          @end="syncToFormJSON"
        >
          <template #item="{ element, index }">
            <div
              class="flex items-center justify-between p-3 rounded-lg border border-[var(--el-border-color-lighter)] bg-[var(--el-bg-color)] hover:border-[var(--el-color-primary-light-5)] transition-all shadow-2xs"
            >
              <!-- Drag Handle & Index Badge -->
              <div class="flex items-center space-x-3">
                <span class="drag-handle cursor-move text-[var(--el-text-color-placeholder)] hover:text-[var(--el-color-primary)] p-1 rounded" title="按住拖拽调整顺序">
                  <component :is="useRenderIcon(DragIcon)" class="w-4 h-4" />
                </span>
                <el-tag size="small" type="info" effect="plain" class="font-mono font-bold">
                  #{{ index + 1 }}
                </el-tag>
              </div>

              <!-- Matcher & Action Summaries -->
              <div class="flex-1 grid grid-cols-1 sm:grid-cols-2 gap-3 px-4">
                <!-- Matcher Badge & Text -->
                <div class="flex items-center space-x-2 overflow-hidden">
                  <el-tag size="small" :type="getMatcherSummary(element).tagType as any" effect="light" class="font-mono shrink-0">
                    {{ getMatcherSummary(element).name }}
                  </el-tag>
                  <span class="text-xs text-[var(--el-text-color-regular)] truncate font-mono">
                    {{ getMatcherSummary(element).summary }}
                  </span>
                </div>

                <!-- Action Badge & Text -->
                <div class="flex items-center space-x-2 overflow-hidden">
                  <el-tag size="small" :type="getActionSummary(element).tagType as any" effect="light" class="font-mono shrink-0">
                    {{ getActionSummary(element).name }}
                  </el-tag>
                  <span class="text-xs text-[var(--el-text-color-regular)] truncate">
                    {{ getActionSummary(element).summary }}
                  </span>
                </div>
              </div>

              <!-- Item Row Operations -->
              <div class="flex items-center space-x-1 shrink-0">
                <el-button
                  size="small"
                  link
                  :disabled="index === 0"
                  :icon="useRenderIcon(ArrowUp)"
                  @click="moveItemUp(index)"
                />
                <el-button
                  size="small"
                  link
                  :disabled="index === itemList.length - 1"
                  :icon="useRenderIcon(ArrowDown)"
                  @click="moveItemDown(index)"
                />
                <el-button
                  size="small"
                  link
                  type="primary"
                  :icon="useRenderIcon(EditPen)"
                  @click="openEditItemDialog(index)"
                >
                  编辑
                </el-button>
                <el-button
                  size="small"
                  link
                  type="danger"
                  :disabled="itemList.length <= 1"
                  :icon="useRenderIcon(Delete)"
                  @click="removeItem(index)"
                >
                  删除
                </el-button>
              </div>
            </div>
          </template>
        </draggable>

        <div v-if="itemList.length === 0" class="text-center py-6 text-xs text-[var(--el-text-color-placeholder)]">
          暂无条目，请点击右上角“添加规则条目”
        </div>
      </el-scrollbar>
    </el-card>

    <!-- Theme-aware JSON Preview Box -->
    <div class="rounded-lg border border-[var(--el-border-color-lighter)] bg-[var(--el-fill-color-light)] p-3">
      <div class="flex items-center justify-between mb-2">
        <span class="text-xs font-semibold text-[var(--el-text-color-secondary)]">规则组 JSON 序列化数据结构预览 (Items 数组)</span>
        <el-tag size="small" type="info" effect="plain" class="font-mono">JSON Preview</el-tag>
      </div>
      <div class="bg-[var(--el-bg-color)] p-2.5 rounded border border-[var(--el-border-color-lighter)]">
        <el-scrollbar max-height="150px" class="item-scrollbar pr-1">
          <pre class="text-[11px] text-[var(--el-text-color-primary)] font-mono whitespace-pre-wrap break-all leading-relaxed">{{ newFormInline.items }}</pre>
        </el-scrollbar>
      </div>
    </div>

    <!-- Nested Sub-Dialog for Adding / Editing a Single Rule Item -->
    <el-dialog
      v-model="itemDialogVisible"
      :title="editingIndex !== null ? `编辑规则条目 (#${editingIndex + 1})` : '添加规则条目'"
      width="640px"
      append-to-body
      destroy-on-close
      class="item-config-dialog"
    >
      <div class="space-y-4 py-1">
        <!-- Matcher Card -->
        <div class="p-3.5 rounded-lg border border-[var(--el-border-color-lighter)] bg-[var(--el-fill-color-light)]">
          <div class="text-xs font-bold text-[var(--el-color-primary)] mb-3 flex items-center justify-between">
            <span class="text-sm">1. 匹配条件配置 (Matcher)</span>
            <el-tag size="small" type="primary" effect="plain">Matcher</el-tag>
          </div>

          <el-form-item label="选择 Matcher" label-width="110px">
            <el-select v-model="itemForm.matcherName" class="w-full">
              <el-option
                v-for="m in matcherOptions"
                :key="m.value"
                :label="m.label"
                :value="m.value"
              />
            </el-select>
          </el-form-item>

          <div v-if="itemForm.matcherName === 'ip_matcher'" class="mt-2 pl-[110px]">
            <el-input
              v-model="itemForm.ipAddressText"
              type="textarea"
              :rows="3"
              placeholder="每行一个 IP、CIDR 或 Range，例如:&#10;127.0.0.1&#10;192.168.1.0/24&#10;10.0.0.1-10.0.0.100"
            />
            <div class="text-xs text-[var(--el-text-color-secondary)] mt-1">
              提示: 用于 TCP/UDP 传输层匹配。支持精确 IP (1.2.3.4)、CIDR (192.168.1.0/24) 或范围 (10.0.0.1-10.0.0.50)
            </div>
          </div>

          <div v-if="itemForm.matcherName === 'http_ip_matcher'" class="mt-2 pl-[110px]">
            <el-input
              v-model="itemForm.httpIpAddressText"
              type="textarea"
              :rows="3"
              placeholder="每行一个 IP、CIDR 或 Range，例如:&#10;127.0.0.1&#10;192.168.1.0/24"
            />
            <div class="text-xs text-[var(--el-text-color-secondary)] mt-1">
              提示: 用于 HTTP 代理应用层匹配。在 ang 中注册为 http_ip_matcher
            </div>
          </div>
        </div>

        <!-- Action Card -->
        <div class="p-3.5 rounded-lg border border-[var(--el-border-color-lighter)] bg-[var(--el-fill-color-light)]">
          <div class="text-xs font-bold text-[var(--el-color-success)] mb-3 flex items-center justify-between">
            <span class="text-sm">2. 触发动作配置 (Action)</span>
            <el-tag size="small" type="success" effect="plain">Action</el-tag>
          </div>

          <el-form-item label="选择 Action" label-width="110px">
            <el-select v-model="itemForm.actionName" class="w-full">
              <el-option
                v-for="a in actionOptions"
                :key="a.value"
                :label="a.label"
                :value="a.value"
              />
            </el-select>
          </el-form-item>

          <div v-if="itemForm.actionName === 'reset_conn_action'" class="mt-2 pl-[110px]">
            <el-input
              v-model="itemForm.resetContent"
              placeholder="关闭 TCP 连接前向客户端写入的可选提示文本"
            />
            <div class="text-xs text-[var(--el-text-color-secondary)] mt-1">
              提示: 适用于 TCP 会话，终止连接前可选发送提示消息
            </div>
          </div>

          <div v-if="itemForm.actionName === 'hide_version_action'" class="mt-2">
            <el-alert
              title="HTTP 隐藏版本号动作 (hide_version_action) 无需配置额外参数，生效后会自动将 Server 响应头重写为 '***'。"
              type="info"
              :closable="false"
              show-icon
            />
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end space-x-2">
          <el-button @click="itemDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="saveItemFromDialog">确认保存条目</el-button>
        </div>
      </template>
    </el-dialog>
  </el-form>
</template>

<style scoped>
.rule-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--el-text-color-regular);
}

.item-scrollbar :deep(.el-scrollbar__bar.is-vertical) {
  width: 5px;
}
.item-scrollbar :deep(.el-scrollbar__thumb) {
  background-color: var(--el-border-color);
  border-radius: 4px;
  opacity: 0.6;
}
.item-scrollbar :deep(.el-scrollbar__thumb:hover) {
  background-color: var(--el-color-primary);
  opacity: 0.85;
}
</style>
