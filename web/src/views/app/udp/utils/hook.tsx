import dayjs from "dayjs";
import { message } from "@/utils/message";
import type { PaginationProps } from "@pureadmin/table";
import { deviceDetection } from "@pureadmin/utils";
import {
  getUdpList,
  createUdp,
  updateUdp,
  deleteUdp,
  type UdpProxyItem
} from "@/api/udp";
import { getTunnelList } from "@/api/tunnel";
import { type Ref, ref, computed, reactive, onMounted } from "vue";

export function useUdpProxy(t: any, tableRef: Ref) {
  const form = reactive({
    name: "",
    port: ""
  });
  const formRef = ref();
  const dataList = ref([]);
  const loading = ref(true);
  const selectedNum = ref(0);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true,
    size: deviceDetection() ? "small" : "default",
    layout: deviceDetection()
      ? "prev, pager, next"
      : "total, sizes, prev, pager, next, jumper"
  });

  const tunnelMap = ref<Record<string, string>>({});

  const columns = computed<TableColumnList>(() => [
    
    {
      label: t("udp.selectionColumn", "勾选列"),
      type: "selection",
      fixed: "left",
      reserveSelection: true
    },
    {
      label: "ID",
      prop: "Id",
      width: 80,
      formatter: row => row.Id || row.id
    },
    {
      label: t("udp.name", "名称"),
      prop: "Name",
      minWidth: 140,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("udp.name", "名称")}</span>
      ),
      cellRenderer: scope => {
        const name = scope.row.Name || scope.row.name || "-";
        return (
          <span class="font-semibold text-sm/snug text-(--el-text-color-primary) wrap-break-word inline-block py-1">
            {name}
          </span>
        );
      }
    },
    {
      label: t("udp.listenAddressPort", "监听地址与端口"),
      minWidth: 170,
      cellRenderer: scope => {
        const addr = scope.row.Address || scope.row.address || "0.0.0.0";
        const port = scope.row.Port || scope.row.port || "";
        return (
          <el-tag
            type="primary"
            effect="light"
            class="font-mono font-bold whitespace-nowrap"
          >
            {addr}:{port}
          </el-tag>
        );
      }
    },
    {
      label: t("udp.rules", "规则"),
      align: "center",
      minWidth: 160,
      cellRenderer: scope => {
        const rulesStr = scope.row.Rules || scope.row.rules || "";
        let ruleList: string[] = [];
        try {
          if (rulesStr) ruleList = JSON.parse(rulesStr);
        } catch (e) {
          ruleList = [];
        }
        if (!ruleList || ruleList.length === 0) {
          return <span class="text-gray-400 text-xs">-</span>;
        }
        return (
          <div class="flex flex-wrap justify-center gap-1">
            {ruleList.map(r => (
              <el-tag
                key={r}
                size="small"
                type="info"
                effect="plain"
                class="font-mono"
              >
                {r}
              </el-tag>
            ))}
          </div>
        );
      }
    },
    {
      label: t("udp.tunnel", "Tunnel"),
      minWidth: 150,
      align: "center",
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("udp.tunnel", "Tunnel")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const tunnelId = row.TunnelId || row.tunnel_id;
        const tunnelToken = row.TunnelToken || row.tunnel_token || "";
        const mapKey = `${tunnelId}|${tunnelToken}`;
        const tName = tunnelMap.value[mapKey];
        let displayName = `${(row.TunnelType || row.tunnel_type || "TLS").toUpperCase()} ${tunnelId}`;
        if (tName) {
          displayName = tName;
        }

        return (
          <div class="flex justify-center items-center h-full w-full py-1">
            {tunnelId ? (
              <el-tag
                size="small"
                type="success"
                effect="light"
                class="font-mono whitespace-nowrap"
              >
                {displayName}
              </el-tag>
            ) : (
              <span class="text-(--el-text-color-placeholder) text-xs">-</span>
            )}
          </div>
        );
      }
    },
    {
      label: t("udp.backend", "上游"),
      minWidth: 200,
      cellRenderer: scope => {
        const row = scope.row;
        const serversStr = row.UpstreamServers || row.upstream_servers || "";
        let servers: Array<{ target: string; weight: number }> = [];
        try {
          if (serversStr) {
            const parsed =
              typeof serversStr === "string"
                ? JSON.parse(serversStr)
                : serversStr;
            if (Array.isArray(parsed)) {
              servers = parsed
                .map((item: any) => ({
                  target: item.target || item.Target || "",
                  weight: Number(item.weight || item.Weight || 1)
                }))
                .filter((item: any) => item.target);
            }
          }
        } catch (e) {}

        if (!servers || servers.length === 0) {
          return (
            <span class="text-(--el-text-color-placeholder) text-xs">-</span>
          );
        }

        const method =
          row.UpstreamMethod || row.upstream_method || "round_robin";

        const serverChunks: any[][] = [];
        for (let i = 0; i < servers.length; i += 2) {
          serverChunks.push(servers.slice(i, i + 2));
        }

        return (
          <div class="p-1.5 rounded-lg border border-(--el-border-color-lighter) bg-(--el-fill-color-light) space-y-1 py-1">
            <div class="flex items-center gap-1.5 flex-wrap">
              <el-tag
                size="small"
                type="info"
                effect="plain"
                class="text-[11px] font-mono font-bold"
              >
                {method}
              </el-tag>
              <span class="text-xs text-gray-400">
                ({servers.length})
              </span>
            </div>
            <div class="space-y-1 pt-0.5">
              {serverChunks.map((chunk, cIdx) => (
                <div key={cIdx} class="flex items-center gap-1.5 flex-wrap">
                  {chunk.map((srv, sIdx) => (
                    <span
                      key={sIdx}
                      class="inline-flex items-center gap-1 font-mono text-[11px] bg-(--el-bg-color) px-1.5 py-0.5 rounded border border-(--el-border-color-lighter) shrink-0"
                    >
                      <span class="font-medium text-(--el-text-color-primary)">
                        {srv.target}
                      </span>
                      <span class="text-[10px] text-gray-400 font-normal">
                        ({srv.weight})
                      </span>
                    </span>
                  ))}
                </div>
              ))}
            </div>
          </div>
        );
      }
    },
    {
      label: t("udp.remark", "备注"),
      prop: "Remark",
      minWidth: 120,
      align: "center",
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("udp.remark", "备注")}</span>
      ),
      cellRenderer: scope => {
        const remark = scope.row.Remark || scope.row.remark || "-";
        if (!remark || remark === "-") {
          return (
            <span class="text-xs text-(--el-text-color-placeholder)">-</span>
          );
        }
        return (
          <span class="text-xs/snug text-(--el-text-color-regular) wrap-break-word inline-block py-1">
            {remark}
          </span>
        );
      }
    },
    {
      label: t("udp.createTime", "创建时间"),
      prop: "CreatedAt",
      minWidth: 160,
      formatter: ({ CreatedAt, created_at }) =>
        CreatedAt || created_at
          ? dayjs(CreatedAt || created_at).format("YYYY-MM-DD HH:mm:ss")
          : "-"
    },
    {
      label: t("udp.operation", "操作"),
      fixed: "right",
      width: 160,
      slot: "operation"
    }
  ]);

  async function onSearch() {
    loading.value = true;
    try {
      const { data } = await getUdpList({
        name: form.name,
        port: form.port
      });
      dataList.value = data?.list || [];
      pagination.total = data?.total || dataList.value.length;
    } catch (e: any) {
      message(e.message || "获取 UDP 列表失败", { type: "error" });
    } finally {
      loading.value = false;
    }
  }

  const resetForm = (formEl: any) => {
    if (!formEl) return;
    formEl.resetFields();
    form.name = "";
    form.port = "";
    onSearch();
  };

  async function handleDelete(row: any) {
    try {
      const targetId = row.Id || row.id;
      const { code, message: msg } = await deleteUdp(targetId);
      if (code === 0) {
        message(t("udp.deleteSuccess", "删除成功"), { type: "success" });
        onSearch();
      } else {
        message(msg || "删除失败", { type: "error" });
      }
    } catch (e: any) {
      message(e.message || "删除失败", { type: "error" });
    }
  }

  function handleSelectionChange(val: any[]) {
    selectedNum.value = val.length;
  }

  function onSelectionCancel() {
    selectedNum.value = 0;
    tableRef.value?.getTableRef()?.clearSelection();
  }

  async function onbatchDel() {
    const rows = tableRef.value.getTableRef().getSelectionRows();
    if (!rows || rows.length === 0) return;
    const ids = rows.map((r: any) => r.Id || r.id);
    try {
      const { code, message: msg } = await deleteUdp({ ids });
      if (code === 0) {
        message(t("udp.batchDeleteSuccess", "批量删除成功"), {
          type: "success"
        });
        tableRef.value?.getTableRef()?.clearSelection();
        selectedNum.value = 0;
        onSearch();
      } else {
        message(msg || "批量删除失败", { type: "error" });
      }
    } catch (e: any) {
      message(e.message || "批量删除失败", { type: "error" });
    }
  }

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
    pagination.currentPage = 1;
    onSearch();
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val;
  }

  async function fetchTunnelNames() {
    try {
      const res = await getTunnelList();
      let list: any[] = [];
      if (Array.isArray(res?.data?.list)) list = res.data.list;
      else if (Array.isArray(res?.data)) list = res.data;
      else if (Array.isArray(res)) list = res;

      const map: Record<string, string> = {};
      list.forEach((tItem: any) => {
        const tid = String(tItem.Id || tItem.id);
        const cNodes = tItem.client_nodes || tItem.ClientNodes || [];
        cNodes.forEach((c: any) => {
          const cName = c.Name || c.name || "";
          if (cName) {
            map[`${tid}|${c.token || c.Token}`] = cName;
          }
        });
      });
      tunnelMap.value = map;
    } catch (e) {}
  }

  onMounted(() => {
    fetchTunnelNames();
    onSearch();
  });

  return {
    form,
    loading,
    columns,
    dataList,
    selectedNum,
    pagination,
    onSearch,
    resetForm,
    onbatchDel,
    handleDelete,
    handleSizeChange,
    onSelectionCancel,
    handleCurrentChange,
    handleSelectionChange
  };
}
