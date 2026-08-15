<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import ReCol from "@/components/ReCol";
import { useI18n } from "vue-i18n";
import { getActiveTunnelConnections, getTunnelClientList } from "@/api/tunnel-client";
import { message } from "@/utils/message";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import RefreshLine from "~icons/ri/refresh-line";

const props = withDefaults(defineProps<{ formInline: any }>(), {
  formInline: () => ({
    title: "新增隧道节点",
    id: undefined,
    name: "",
    type: "tls",
    tunnel_id: "",
    token: "",
    remark: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const { t } = useI18n();

const activeConnOptions = ref<Array<any>>([]);
const activeLoading = ref(false);
const selectedActiveConn = ref("");

const formRules = reactive({
  name: [
    { required: true, message: () => t("tunnelClient.nameRequired"), trigger: "blur" }
  ],
  token: [
    { required: true, message: () => t("tunnelClient.selectNodeRequired"), trigger: "change" }
  ]
});

async function fetchActiveConnections() {
  activeLoading.value = true;
  const [resConns, resClients] = await Promise.all([
    getActiveTunnelConnections(),
    getTunnelClientList()
  ]);
  activeLoading.value = false;

  if (resConns?.code === 0 && Array.isArray(resConns.data)) {
    const allConns = resConns.data;
    const savedClients = Array.isArray(resClients?.data?.list)
      ? resClients.data.list
      : Array.isArray(resClients?.data)
      ? resClients.data
      : [];

    const mapped = allConns.map((item: any) => {
      const boundClient = savedClients.find(
        (sc: any) =>
          String(sc.tunnel_id) === String(item.tunnel_id) &&
          String(sc.token) === String(item.token)
      );
      return {
        ...item,
        isAlreadyBound: !!boundClient,
        boundName: boundClient ? boundClient.name : "",
        boundId: boundClient ? boundClient.id : undefined
      };
    });

    const targetTunnelId = newFormInline.value?.tunnel_id;
    if (targetTunnelId) {
      activeConnOptions.value = mapped.filter(
        item => String(item.tunnel_id) === String(targetTunnelId)
      );
    } else {
      activeConnOptions.value = mapped;
    }

    // If token exists in formInline, pre-select matched active connection
    if (newFormInline.value?.token) {
      const curToken = newFormInline.value.token;
      const matched = activeConnOptions.value.find(item => item.token === curToken);
      if (matched) {
        selectedActiveConn.value = matched.label;
      }
    }
  }
}

async function refreshActiveConnections() {
  await fetchActiveConnections();
  message(t("tunnelClient.refreshNodesSuccess"), { type: "success" });
}

function handleSelectActiveConn(val: string) {
  if (!val) {
    selectedActiveConn.value = "";
    return;
  }
  const matched = activeConnOptions.value.find(
    item => item.label === val || item.token === val
  );
  if (matched) {
    newFormInline.value.type = matched.type;
    newFormInline.value.tunnel_id = matched.tunnel_id;
    newFormInline.value.token = matched.token;

    const tokenStr = matched.token || "";
    const tokenSuffix = tokenStr.length > 8 ? tokenStr.slice(-6) : tokenStr;
    const defaultName = tokenSuffix
      ? `Node-${matched.type.toUpperCase()}-${matched.tunnel_id}-${tokenSuffix}`
      : `Node-${matched.type.toUpperCase()}-${matched.tunnel_id}`;

    if (!newFormInline.value.name || newFormInline.value.name.startsWith("Node-")) {
      newFormInline.value.name = defaultName;
    }
  }
}

onMounted(() => {
  fetchActiveConnections();
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
    label-width="70px"
    class="py-1 px-1"
  >
    <el-row :gutter="16">
      <!-- 1. 名称 (Name) - 置于最上方 -->
      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('tunnelClient.name')" prop="name">
          <el-input
            v-model="newFormInline.name"
            :placeholder="t('tunnelClient.nodeNamePlaceholder')"
            clearable
          />
        </el-form-item>
      </re-col>

      <!-- 2. 节点 (Node) 下拉框 (刷新图标直接位于下拉框右侧) -->
      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('tunnelClient.nodeRef')" prop="token">
          <div class="w-full">
            <!-- 仅当在【节点】下拉框中选中项目后才在正上方呈现的信息胶囊 -->
            <div
              v-if="selectedActiveConn && (newFormInline.type || newFormInline.tunnel_id || newFormInline.token)"
              class="mb-2 inline-flex flex-wrap items-center gap-2 p-1.5 bg-gray-50 dark:bg-gray-800/60 rounded-md border border-gray-100 dark:border-gray-700/60 text-xs"
            >
              <el-tag
                size="small"
                type="primary"
                effect="dark"
                class="font-bold rounded"
              >
                {{ (newFormInline.type || 'tls').toUpperCase() }}
              </el-tag>

              <el-tag
                v-if="newFormInline.tunnel_id"
                size="small"
                type="info"
                effect="plain"
                class="font-mono rounded"
              >
                Tunnel ID: {{ newFormInline.tunnel_id }}
              </el-tag>

              <span
                v-if="newFormInline.token"
                class="font-mono text-xs font-semibold text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/40 px-2 py-0.5 rounded border border-blue-200/60 dark:border-blue-800/60"
              >
                Token: {{ newFormInline.token }}
              </span>
            </div>

            <!-- 节点下拉框与右侧刷新图标 -->
            <div class="flex items-center space-x-2 w-full">
              <el-select
                v-model="selectedActiveConn"
                clearable
                filterable
                :loading="activeLoading"
                class="flex-1"
                :placeholder="t('tunnelClient.selectActivePlaceholder')"
                @change="handleSelectActiveConn"
              >
                <el-option
                  v-for="item in activeConnOptions"
                  :key="`${item.type}-${item.tunnel_id}-${item.token}-${item.remote_addr}`"
                  :label="item.label"
                  :value="item.label"
                  :disabled="item.isAlreadyBound && (!newFormInline.id || Number(newFormInline.id) !== Number(item.boundId))"
                  class="h-auto! py-2"
                >
                  <div class="flex flex-col space-y-1 w-full text-xs font-mono leading-normal">
                    <div class="flex items-center justify-between gap-2">
                      <span class="font-semibold text-gray-800 dark:text-gray-200 break-all">
                        Token: {{ item.token }}
                      </span>
                      <div class="flex items-center space-x-1 shrink-0">
                        <el-tag
                          v-if="item.isAlreadyBound"
                          size="small"
                          type="info"
                          class="rounded"
                        >
                          {{ t('tunnelClient.bound') }}: {{ item.boundName }}
                        </el-tag>
                          <el-tag
                            v-else
                            size="small"
                            type="warning"
                            effect="light"
                            class="rounded font-medium inline-flex items-center gap-1"
                          >
                            <IconifyIconOffline icon="ri:flash-line" /> {{ t('tunnel.unsavedTag') }}
                          </el-tag>
                        <el-tag
                          size="small"
                          type="primary"
                          class="font-bold"
                        >
                          {{ item.type.toUpperCase() }}
                        </el-tag>
                      </div>
                    </div>

                    <div class="flex items-center justify-between gap-2 text-gray-500 dark:text-gray-400 text-[11px]">
                      <span class="break-all">
                        Remote: {{ item.remote_addr }} <template v-if="item.sni">({{ item.sni }})</template>
                      </span>
                      <span class="shrink-0 text-gray-400">
                        Tunnel ID: {{ item.tunnel_id }}
                      </span>
                    </div>
                  </div>
                </el-option>
              </el-select>

              <!-- 右侧带箭头的圆形刷新图标按钮 -->
              <el-tooltip :content="t('tunnelClient.refreshNodes')" placement="top">
                <el-button
                  type="primary"
                  link
                  class="px-1! cursor-pointer"
                  :icon="useRenderIcon(RefreshLine)"
                  :loading="activeLoading"
                  @click="refreshActiveConnections"
                />
              </el-tooltip>
            </div>
          </div>
        </el-form-item>
      </re-col>

      <!-- 3. 备注 (Remark) -->
      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item :label="t('tunnelClient.remark')" prop="remark">
          <el-input
            v-model="newFormInline.remark"
            type="textarea"
            :rows="2"
            :placeholder="t('tunnelClient.remarkPlaceholder')"
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
