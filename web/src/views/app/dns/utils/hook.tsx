import dayjs from "dayjs";
import editForm from "../form/index.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import { deviceDetection } from "@pureadmin/utils";
import {
  getDnsList,
  createDns,
  updateDns,
  deleteDns,
  type DnsProxyItem
} from "@/api/dns";
import { getTunnelClientList } from "@/api/tunnel-client";
import { type Ref, h, ref, computed, toRaw, reactive, onMounted } from "vue";

export function useDnsProxy(t: any, tableRef: Ref) {
  const form = reactive({
    address: "",
    port: ""
  });
  const formRef = ref();
  const tunnelMap = ref<Record<string, any>>({});
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

  const columns = computed<TableColumnList>(() => [
    {
      type: "expand",
      slot: "expand"
    },
    {
      label: t("dns.selectionColumn") || t("tunnel.selectionColumn"),
      type: "selection",
      fixed: "left",
      reserveSelection: true
    },
    {
      label: t("dns.id"),
      prop: "Id",
      width: 80,
      formatter: row => row.Id || row.id
    },
    {
      label: t("dns.name", "名称"),
      prop: "Name",
      minWidth: 120,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("dns.name", "名称")}</span>
      ),
      cellRenderer: scope => {
        const name = scope.row.Name || scope.row.name || "-";
        return (
          <span class="font-semibold text-sm/snug text-(--el-text-color-primary) wrap-break-word inline-block  py-1">
            {name}
          </span>
        );
      }
    },
    {
      label: t("dns.listenAddressPort"),
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
      label: t("dns.hosts"),
      minWidth: 180,
      slot: "hosts"
    },
    {
      label: t("dns.rules"),
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
      label: t("dns.tunnel", "Tunnel"),
      minWidth: 150,
      align: "center",
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("dns.tunnel", "Tunnel")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const tunnelId = String(row.TunnelId ?? row.tunnel_id ?? "").trim();
        const tunnelToken = String(row.TunnelToken ?? row.tunnel_token ?? "").trim();
        const tunnelType = String(row.TunnelType ?? row.tunnel_type ?? "TLS").toUpperCase();

        if (!tunnelId && !tunnelToken) {
          return (
            <div class="flex justify-center items-center h-full w-full py-1">
              <span class="text-(--el-text-color-placeholder) text-xs">-</span>
            </div>
          );
        }

        const mapKey = `${tunnelId}|${tunnelToken}`;
        const tInfo = tunnelMap.value[tunnelToken] || tunnelMap.value[mapKey] || tunnelMap.value[tunnelId];

        const displayName = tInfo?.name || `${tunnelType} ${tunnelId || tunnelToken}`;
        const isOnline = Boolean(tInfo?.isOnline);

        return (
          <div class="flex justify-center items-center h-full w-full py-1">
            <span
              class={[
                "font-mono font-bold text-sm whitespace-nowrap",
                isOnline ? "text-(--el-color-success)" : "text-gray-400"
              ]}
            >
              {displayName}
            </span>
          </div>
        );
      }
    },
    {
      label: t("dns.backend"),
      minWidth: 200,
      slot: "backend"
    },
    {
      label: t("dns.remark"),
      prop: "Remark",
      minWidth: 110,
      align: "center",
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("dns.remark")}</span>
      ),
      cellRenderer: scope => {
        const remark = scope.row.Remark || scope.row.remark || "-";
        if (!remark || remark === "-") {
          return (
            <span class="text-xs text-(--el-text-color-placeholder)">-</span>
          );
        }
        return (
          <span class="text-xs/snug text-(--el-text-color-regular) wrap-break-word inline-block  py-1">
            {remark}
          </span>
        );
      }
    },
    {
      label: t("dns.createTime"),
      minWidth: 160,
      prop: "created_at",
      formatter: row => {
        const timeVal = row.created_at || row.CreatedAt;
        return timeVal && dayjs(timeVal).isValid() && dayjs(timeVal).year() > 1
          ? dayjs(timeVal).format("YYYY-MM-DD HH:mm:ss")
          : "-";
      }
    },
    {
      label: t("dns.operation"),
      fixed: "right",
      width: 160,
      slot: "operation"
    }
  ]);

  async function handleDelete(row: any) {
    const targetId = row.Id || row.id;
    const { code, message: msg } = await deleteDns({ id: targetId });
    if (code === 0) {
      message(
        `${t("dns.delete")} ID: ${targetId} ${t("common.success", "成功")}`,
        { type: "success" }
      );
      onSearch();
    } else {
      message(msg, { type: "error" });
    }
  }

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val;
  }

  function handleSelectionChange(val: any[]) {
    selectedNum.value = val.length;
    if (tableRef.value) {
      tableRef.value.setAdaptive();
    }
  }

  function onSelectionCancel() {
    selectedNum.value = 0;
    if (tableRef.value?.getTableRef) {
      tableRef.value.getTableRef().clearSelection();
    }
  }

  async function onbatchDel() {
    const curSelected = tableRef.value.getTableRef().getSelectionRows();
    const ids = curSelected.map((item: any) => item.Id || item.id);
    const { code, message: msg } = await deleteDns({ ids });
    if (code === 0) {
      message(`${t("dns.batchDelete")} ${t("common.success", "成功")}`, {
        type: "success"
      });
      tableRef.value.getTableRef().clearSelection();
      onSearch();
    } else {
      message(msg, { type: "error" });
    }
  }

  async function onSearch() {
    loading.value = true;
    const searchParams = toRaw(form);
    const res = await getDnsList(searchParams);

    if (res?.code === 0 && res?.data) {
      dataList.value = res.data.list || [];
      pagination.total = res.data.total || dataList.value.length;
      pagination.pageSize = res.data.pageSize || 10;
      pagination.currentPage = res.data.currentPage || 1;
    }
    setTimeout(() => {
      loading.value = false;
    }, 300);
  }

  const resetForm = (formEl: any) => {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  };

  function openDialog(title = "", row?: any) {
    addDialog({
      title: `${title}`,
      props: {
        formInline: {
          title,
          id: row?.Id ?? row?.id ?? undefined,
          address: row?.Address ?? row?.address ?? "",
          port: row?.Port ?? row?.port ?? "5656",
          rules: row?.Rules ?? row?.rules ?? "[]",
          hosts_text: row?.HostsText ?? row?.hosts_text ?? "",
          hosts_json: row?.HostsJSON ?? row?.hosts_json ?? "",
          backend_type: row?.BackendType ?? row?.backend_type ?? "upstream",
          tunnel_type: (
            row?.TunnelType ??
            row?.tunnel_type ??
            "quic"
          ).toLowerCase(),
          tunnel_id: row?.TunnelId ?? row?.tunnel_id ?? "",
          tunnel_token: row?.TunnelToken ?? row?.tunnel_token ?? "",
          upstream_method:
            row?.UpstreamMethod ?? row?.upstream_method ?? "round_robin",
          upstream_servers:
            row?.UpstreamServers ??
            row?.upstream_servers ??
            JSON.stringify([{ target: "8.8.8.8:53", weight: 1 }]),
          remark: row?.Remark ?? row?.remark ?? ""
        }
      },
      width: deviceDetection() ? "95%" : "860px",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: ({ options }) =>
        h(editForm, { ref: formRef, formInline: options.props.formInline }),
      beforeSure: (done, { options }) => {
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline;
        FormRef.validate(async (valid: boolean) => {
          if (valid) {
            if (title === t("dns.addDns")) {
              const { code, message: msg } = await createDns(curData);
              if (code !== 0) {
                message(msg, { type: "error" });
                return;
              }
            } else {
              const { code, message: msg } = await updateDns(curData);
              if (code !== 0) {
                message(msg, { type: "error" });
                return;
              }
            }
            message(`${title} ${t("common.success", "成功")}`, {
              type: "success"
            });
            done();
            onSearch();
          }
        });
      }
    });
  }

    async function fetchTunnelNames() {
    try {
      const clientRes = await getTunnelClientList();
      let cList: any[] = [];
      if (Array.isArray(clientRes?.data?.list)) cList = clientRes.data.list;
      else if (Array.isArray(clientRes?.data)) cList = clientRes.data;
      else if (Array.isArray(clientRes?.list)) cList = clientRes.list;
      else if (Array.isArray(clientRes)) cList = clientRes;

      const map: Record<string, { name: string; isOnline: boolean }> = {};
      cList.forEach((c: any) => {
        const tid = String(c.TunnelId ?? c.tunnel_id ?? "").trim();
        const tkn = String(c.Token ?? c.token ?? "").trim();
        const cName = String(c.Name ?? c.name ?? "").trim();
        const isOnline = Boolean(c.IsOnline ?? c.is_online);

        if (cName) {
          const info = { name: cName, isOnline };
          if (tkn) {
            map[tkn] = info;
            if (tid) {
              map[`${tid}|${tkn}`] = info;
            }
          }
          if (tid && !map[tid]) {
            map[tid] = info;
          }
        }
      });

      tunnelMap.value = map;
      if (Array.isArray(dataList.value)) {
        dataList.value = [...dataList.value];
      }
    } catch (e) {
      console.error("fetchTunnelNames failed:", e);
    }
  }

  onMounted(async () => {
    await fetchTunnelNames();
    onSearch();
  });

  return {
    form,
    loading,
    columns,
    dataList,
    selectedNum,
    pagination,
    deviceDetection,
    onSearch,
    resetForm,
    onbatchDel,
    openDialog,
    handleDelete,
    handleSizeChange,
    onSelectionCancel,
    handleCurrentChange,
    handleSelectionChange
  };
}
