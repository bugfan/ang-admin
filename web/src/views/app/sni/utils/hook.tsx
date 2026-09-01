import dayjs from "dayjs";
import { message } from "@/utils/message";
import type { PaginationProps } from "@pureadmin/table";
import { deviceDetection } from "@pureadmin/utils";
import {
  getSniList,
  createSni,
  updateSni,
  deleteSni,
  type SniProxyItem
} from "@/api/sni";
import { getTunnelList } from "@/api/tunnel";
import { type Ref, ref, computed, reactive, onMounted } from "vue";

export function useSniProxy(t: any, tableRef: Ref) {
  const form = reactive({
    name: "",
    sni: "",
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
      label: t("sni.selectionColumn", "勾选列"),
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
      label: t("sni.name", "名称"),
      prop: "Name",
      minWidth: 140,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("sni.name", "名称")}</span>
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
      label: t("sni.sni", "SNI与端口"),
      minWidth: 170,
      cellRenderer: scope => {
        const sni = scope.row.SNI || scope.row.sni || "*";
        const port = scope.row.Port || scope.row.port || "";
        return (
          <el-tag
            type="primary"
            effect="light"
            class="font-mono font-bold whitespace-nowrap"
          >
            {sni}:{port}
          </el-tag>
        );
      }
    },
    {
      label: t("sni.rules", "规则"),
      minWidth: 160,
      align: "center",
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("sni.rules", "规则")}</span>
      ),
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
                class="font-mono whitespace-nowrap"
              >
                {r}
              </el-tag>
            ))}
          </div>
        );
      }
    },
    {
      label: t("sni.tunnel", "Tunnel"),
      minWidth: 150,
      align: "center",
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("sni.tunnel", "Tunnel")}</span>
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
      label: t("sni.dnsResolve", "解析"),
      minWidth: 200,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("sni.dnsResolve", "解析")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const rawDns = row.DNSResolver || row.dns_resolver;
        let dnsList: string[] = [];
        try {
          if (rawDns) {
            if (typeof rawDns === "string" && rawDns.startsWith("[")) {
              dnsList = JSON.parse(rawDns);
            } else if (Array.isArray(rawDns)) {
              dnsList = rawDns;
            } else if (typeof rawDns === "string") {
              dnsList = rawDns.split(/[\n,;]+/).map((s: string) => s.trim()).filter(Boolean);
            }
          }
        } catch {
          if (rawDns) dnsList = [String(rawDns)];
        }

        return (
          <div class="p-1.5 rounded-lg border border-(--el-border-color-lighter) bg-(--el-fill-color-light) space-y-1 py-1 text-left">
            {dnsList.length > 0 ? (
              <div class="flex items-center gap-1.5 flex-wrap">
                {dnsList.map((dns, idx) => (
                  <span
                    key={idx}
                    class="inline-flex items-center font-mono text-[11px] bg-(--el-bg-color) px-1.5 py-0.5 rounded border border-(--el-border-color-lighter) shrink-0 whitespace-nowrap"
                  >
                    <span class="font-medium text-(--el-text-color-primary)">{dns || "Default"}</span>
                  </span>
                ))}
              </div>
            ) : (
              <span class="inline-flex items-center font-mono text-[11px] bg-(--el-bg-color) px-1.5 py-0.5 rounded border border-(--el-border-color-lighter) shrink-0 whitespace-nowrap text-gray-400">
                Default
              </span>
            )}
          </div>
        );
      }
    },
    {
      label: t("sni.remark", "备注"),
      prop: "Remark",
      minWidth: 120,
      align: "center",
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("sni.remark", "备注")}</span>
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
      label: t("sni.createTime", "创建时间"),
      prop: "CreatedAt",
      minWidth: 160,
      formatter: ({ CreatedAt, created_at }) =>
        CreatedAt || created_at
          ? dayjs(CreatedAt || created_at).format("YYYY-MM-DD HH:mm:ss")
          : "-"
    },
    {
      label: t("sni.operation", "操作"),
      fixed: "right",
      width: 160,
      slot: "operation"
    }
  ]);

  async function onSearch() {
    loading.value = true;
    try {
      const { data } = await getSniList({
        name: form.name,
        sni: form.sni,
        port: form.port
      });
      dataList.value = data?.list || [];
      pagination.total = data?.total || dataList.value.length;
    } catch (e: any) {
      message(e.message || "获取 SNI 列表失败", { type: "error" });
    } finally {
      loading.value = false;
    }
  }

  const resetForm = (formEl: any) => {
    if (!formEl) return;
    formEl.resetFields();
    form.name = "";
    form.sni = "";
    form.port = "";
    onSearch();
  };

  async function handleDelete(row: any) {
    try {
      const targetId = row.Id || row.id;
      const { code, message: msg } = await deleteSni(targetId);
      if (code === 0) {
        message(t("common.success", "成功"), { type: "success" });
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
      const { code, message: msg } = await deleteSni({ ids });
      if (code === 0) {
        message(t("common.batchDeleteSuccess", "批量删除成功"), {
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
