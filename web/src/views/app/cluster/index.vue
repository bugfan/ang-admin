<script setup lang="ts">
import { ref, reactive, onMounted, computed } from "vue";
import { useI18n } from "vue-i18n";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import { deviceDetection } from "@pureadmin/utils";
import editForm from "./form/index.vue";

import {
  getClusterNodeList,
  createClusterNode,
  updateClusterNode,
  deleteClusterNode,
  pingClusterNode,
  syncClusterNode,
  syncAllClusterNodes,
  ClusterNodeItem
} from "@/api/cluster_node";

import AddFill from "~icons/ri/add-circle-line";
import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import PulseIcon from "~icons/ri/pulse-line";
import SendPlaneIcon from "~icons/ri/send-plane-line";
import SearchIcon from "~icons/ri/search-line";
import RefreshIcon from "~icons/ri/refresh-line";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";

defineOptions({
  name: "AppCluster"
});

const { t } = useI18n();

const loading = ref(false);
const dataList = ref<ClusterNodeItem[]>([]);
const selectedNum = ref(0);
const selectedRows = ref<ClusterNodeItem[]>([]);

// View Mode: 'list' | 'new' | 'edit'
const showView = ref<"list" | "new" | "edit">("list");
const formInline = ref<any>({});
const saving = ref(false);
const createEditFormRef = ref();

// Search Form
const searchFormRef = ref();
const searchForm = reactive({
  name: ""
});

// Table Columns
const columns = computed(() => [
  { type: "selection", width: 55 },
  { label: "ID", prop: "id", width: 80, align: "center" },
  {
    label: t("cluster.nodeName", "节点名称"),
    prop: "name",
    slot: "name",
    minWidth: 160
  },
  {
    label: t("cluster.nodeAddr", "API 地址"),
    prop: "addr",
    slot: "addr",
    minWidth: 200
  },
  {
    label: t("cluster.status", "节点状态"),
    prop: "status",
    slot: "status",
    minWidth: 140
  },
  {
    label: t("cluster.lastPing", "最后心跳时间"),
    prop: "last_ping",
    slot: "last_ping",
    minWidth: 170
  },
  { label: t("cluster.remark", "备注"), prop: "remark", minWidth: 140 },
  {
    label: t("rule.operation", "操作"),
    slot: "operation",
    fixed: "right",
    minWidth: 300
  }
]);

function formatTime(val?: string) {
  if (!val) return "-";
  try {
    return new Date(val).toLocaleString();
  } catch (e) {
    return val;
  }
}

async function fetchData() {
  loading.value = true;
  try {
    const res = await getClusterNodeList({ name: searchForm.name });
    if (res.code === 0 && res.data) {
      dataList.value = res.data.list || [];
    } else {
      message(res.message || t("cluster.pingError", "获取节点列表失败"), {
        type: "error"
      });
    }
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  fetchData();
}

function resetSearchForm(formEl: any) {
  if (!formEl) return;
  formEl.resetFields();
  fetchData();
}

function handleSelectionChange(rows: ClusterNodeItem[]) {
  selectedRows.value = rows;
  selectedNum.value = rows.length;
}

function getDefaultFormInline() {
  return {
    title: t("cluster.addNode", "添加节点"),
    id: undefined,
    name: "",
    addr: "http://127.0.0.1:8081",
    secret: "",
    remark: ""
  };
}

function getFormInlineFromRow(row: ClusterNodeItem) {
  return {
    title: `${t("cluster.editNode", "编辑集群节点")} [ID: ${row.id}]`,
    id: row.id,
    name: row.name || "",
    addr: row.addr || "",
    secret: row.secret || "",
    remark: row.remark || ""
  };
}

function handleAddPage() {
  formInline.value = getDefaultFormInline();
  showView.value = "new";
}

function handleEditPage(row: ClusterNodeItem) {
  formInline.value = getFormInlineFromRow(row);
  showView.value = "edit";
}

function handleCancelPage() {
  showView.value = "list";
}

async function handleSaveSubmit() {
  if (!createEditFormRef.value) return;
  const FormRef = createEditFormRef.value.getRef();
  if (!FormRef) return;

  FormRef.validate(async (valid: boolean) => {
    if (valid) {
      saving.value = true;
      try {
        const curData = formInline.value;
        let res;
        if (showView.value === "new") {
          res = await createClusterNode(curData);
        } else {
          res = await updateClusterNode(curData);
        }

        if (res.code === 0) {
          message(t("common.success", "成功"), {
            type: "success"
          });
          showView.value = "list";
          fetchData();
        } else {
          message(res.message || t("cluster.submitFailed", "提交失败"), {
            type: "error"
          });
        }
      } finally {
        saving.value = false;
      }
    }
  });
}

async function handleDelete(row: ClusterNodeItem) {
  if (!row.id) return;
  const res = await deleteClusterNode(row.id);
  if (res.code === 0) {
    message(t("common.deleteSuccess", "删除成功"), { type: "success" });
    fetchData();
  } else {
    message(res.message || t("cluster.deleteFailed", "删除失败"), {
      type: "error"
    });
  }
}

async function handleBatchDelete() {
  if (selectedRows.value.length === 0) return;
  const ids = selectedRows.value.map(r => r.id).filter(Boolean);
  const res = await deleteClusterNode({ ids });
  if (res.code === 0) {
    message(t("common.batchDeleteSuccess", "批量删除成功"), { type: "success" });
    selectedNum.value = 0;
    selectedRows.value = [];
    fetchData();
  } else {
    message(res.message || t("cluster.deleteFailed", "删除失败"), {
      type: "error"
    });
  }
}

async function handlePing(row: ClusterNodeItem) {
  if (!row.id) return;
  const res = await pingClusterNode(row.id);
  if (res.code === 0) {
    message(t("cluster.pingSuccess", "节点通信及鉴权正常"), {
      type: "success"
    });
  } else {
    let errStr = res.message || "";
    if (errStr === "auth_failed") {
      errStr = t("cluster.testAuthFailed", "鉴权失败：密钥不正确");
    } else if (
      errStr.includes("timeout") ||
      errStr.includes("connection refused") ||
      errStr.includes("no such host")
    ) {
      errStr = t("cluster.testTimeout", "连接失败：网络不通或地址错误");
    } else if (errStr === "empty address") {
      errStr = t("cluster.nodeAddrRequired", "请输入地址");
    } else {
      errStr = errStr
        ? `${t("cluster.pingError", "节点离线或通信异常")} (${errStr})`
        : t("cluster.pingError", "节点离线或通信异常");
    }
    message(errStr, { type: "warning" });
  }
  fetchData();
}

async function handleSync(row: ClusterNodeItem) {
  if (!row.id) return;
  const res = await syncClusterNode(row.id);
  if (res.code === 0) {
    message(res.message || t("cluster.pushSuccess", "配置下发成功"), {
      type: "success"
    });
  } else {
    message(res.message || t("cluster.pushError", "配置下发失败"), {
      type: "error"
    });
  }
  fetchData();
}

async function handleSyncAll() {
  const res = await syncAllClusterNodes();
  if (res.code === 0) {
    message(res.message || t("cluster.syncAllSuccess", "同步配置指令已触发"), {
      type: "success"
    });
  } else {
    message(res.message || t("cluster.syncAllError", "同步配置失败"), {
      type: "error"
    });
  }
  fetchData();
}

onMounted(() => {
  fetchData();
});
</script>

<template>
  <div class="main">
    <!-- List View Mode -->
    <div v-if="showView === 'list'">
      <!-- Filter / Search Form -->
      <el-form
        ref="searchFormRef"
        :inline="true"
        :model="searchForm"
        class="search-form bg-bg_color w-full px-3 sm:px-6 pt-3 pb-1 overflow-auto mb-3 rounded-xl border border-(--el-border-color-lighter) shadow-2xs"
      >
        <el-form-item
          :label="t('cluster.nodeName', '节点名称') + '：'"
          prop="name"
        >
          <el-input
            v-model="searchForm.name"
            :placeholder="
              t('cluster.searchNamePlaceholder', '搜索节点名称或 API 地址')
            "
            clearable
            class="w-full sm:w-60!"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :icon="useRenderIcon(SearchIcon)"
            :loading="loading"
            @click="onSearch"
          >
            {{ t("rule.search", "搜索") }}
          </el-button>
          <el-button
            :icon="useRenderIcon(RefreshIcon)"
            @click="resetSearchForm(searchFormRef)"
          >
            {{ t("rule.reset", "重置") }}
          </el-button>
        </el-form-item>
      </el-form>

      <PureTableBar
        :title="t('cluster.title', '集群管理')"
        :columns="columns"
        @refresh="fetchData"
      >
        <template #buttons>
          <el-button
            type="primary"
            :icon="useRenderIcon(AddFill)"
            @click="handleAddPage"
          >
            {{ t("buttons.pureAdd", "添加") }}
          </el-button>
          <el-button
            type="primary"
            :icon="useRenderIcon(SendPlaneIcon)"
            @click="handleSyncAll"
          >
            {{ t("cluster.syncAll", "同步") }}
          </el-button>
        </template>

        <template v-slot="{ size, dynamicColumns }">
          <!-- Batch Delete Banner -->
          <div
            v-if="selectedNum > 0"
            class="bg-(--el-color-primary-light-9) text-(--el-color-primary) border border-(--el-color-primary-light-7) px-4 py-2 rounded-lg text-sm mb-3 flex-bc"
          >
            <span>
              {{ t("rule.selected", "已选择") }} {{ selectedNum }}
              {{ t("rule.items", "项") }}
            </span>
            <el-popconfirm
              :title="
                t('cluster.batchDeleteConfirm', '确认批量删除选中的集群节点？')
              "
              @confirm="handleBatchDelete"
            >
              <template #reference>
                <el-button
                  type="danger"
                  size="small"
                  link
                  :icon="useRenderIcon(Delete)"
                >
                  {{ t("rule.batchDelete", "批量删除") }}
                </el-button>
              </template>
            </el-popconfirm>
          </div>

          <pure-table
            v-loading="loading"
            row-key="id"
            adaptive
            :data="dataList"
            :columns="dynamicColumns"
            :size="size"
            border
            @selection-change="handleSelectionChange"
          >
            <template #name="{ row }">
              <span class="font-bold text-(--el-text-color-primary)">{{
                row.name
              }}</span>
            </template>

            <template #addr="{ row }">
              <code class="font-mono text-xs text-(--el-color-primary)">{{
                row.addr
              }}</code>
            </template>

            <template #status="{ row }">
              <el-tag
                v-if="row.status === 1"
                type="success"
                effect="dark"
                class="font-mono"
              >
                {{ t("cluster.online", "🟢 在线") }}
              </el-tag>
              <el-tag
                v-else-if="row.status === 2"
                type="warning"
                effect="dark"
                class="font-mono"
              >
                {{ t("cluster.unauthorized", "🟠 鉴权失败") }}
              </el-tag>
              <el-tag v-else type="danger" effect="dark" class="font-mono">
                {{ t("cluster.offline", "🔴 离线") }}
              </el-tag>
            </template>

            <template #last_ping="{ row }">
              <span class="font-mono text-xs text-(--el-text-color-regular)">
                {{ formatTime(row.last_ping) }}
              </span>
            </template>

            <template #operation="{ row }">
              <el-button
                size="small"
                type="primary"
                link
                :icon="useRenderIcon(PulseIcon)"
                @click="handlePing(row)"
              >
                {{ t("cluster.ping", "测试 (Ping)") }}
              </el-button>
              <el-button
                size="small"
                type="success"
                link
                :icon="useRenderIcon(SendPlaneIcon)"
                @click="handleSync(row)"
              >
                {{ t("cluster.pushConfig", "下发配置") }}
              </el-button>
              <el-button
                size="small"
                type="primary"
                link
                :icon="useRenderIcon(EditPen)"
                @click="handleEditPage(row)"
              >
                {{ t("rule.edit", "编辑") }}
              </el-button>
              <el-popconfirm
                :title="
                  t('cluster.confirmDelete', '是否确认删除该集群节点配置?')
                "
                @confirm="handleDelete(row)"
              >
                <template #reference>
                  <el-button
                    size="small"
                    type="danger"
                    link
                    :icon="useRenderIcon(Delete)"
                  >
                    {{ t("rule.delete", "删除") }}
                  </el-button>
                </template>
              </el-popconfirm>
            </template>
          </pure-table>
        </template>
      </PureTableBar>
    </div>

    <!-- Create / Edit Full Page View Mode -->
    <div
      v-else-if="showView === 'new' || showView === 'edit'"
      class="p-3 sm:p-5 bg-bg_color rounded-xl border border-(--el-border-color-lighter) shadow-2xs"
    >
      <!-- Full Page Header Bar -->
      <PageHeader
        :title="formInline.title"
        :description="t('cluster.pageDesc', '配置和管理集群节点信息')"
        :backTitle="t('cluster.backToList', '返回节点列表')"
        @back="handleCancelPage"
      >
        <template #actions>
          <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
            {{ t("rule.cancel", "取消") }}
          </el-button>
          <el-button
            type="primary"
            :loading="saving"
            :icon="useRenderIcon(CheckIcon)"
            @click="handleSaveSubmit"
          >
            {{ t("rule.save", "保存") }}
          </el-button>
        </template>
      </PageHeader>

      <!-- Form Component Embedded Directly -->
      <editForm ref="createEditFormRef" :formInline="formInline" />
    </div>
  </div>
</template>

<style scoped>
.main {
  padding: 16px;
}
</style>
