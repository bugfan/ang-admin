import dayjs from "dayjs";
import editForm from "../form/index.vue";
import clientEditForm from "../../tunnel-client/form/index.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import { deviceDetection } from "@pureadmin/utils";
import {
  getTunnelList,
  createTunnel,
  updateTunnel,
  deleteTunnel
} from "@/api/tunnel";
import {
  createTunnelClient,
  updateTunnelClient,
  deleteTunnelClient
} from "@/api/tunnel-client";
import { getCertList } from "@/api/certificate";
import { type Ref, h, ref, computed, toRaw, reactive, onMounted } from "vue";

export function useTunnel(t: any, tableRef: Ref) {
  const form = reactive({
    type: "",
    sni: "",
    port: ""
  });
  const formRef = ref();
  const clientFormRef = ref();
  const dataList = ref([]);
  const loading = ref(true);
  const selectedNum = ref(0);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true,
    size: deviceDetection() ? "small" : "default",
    layout: deviceDetection() ? "prev, pager, next" : "total, sizes, prev, pager, next, jumper"
  });

  const validCertSet = ref<Set<string>>(new Set());

  const columns = computed<TableColumnList>(() => [
    {
      type: "expand",
      slot: "expand"
    },
    {
      label: t("tunnel.selectionColumn"),
      type: "selection",
      fixed: "left",
      reserveSelection: true
    },
    {
      label: t("tunnel.id"),
      prop: "Id",
      width: 80,
      formatter: (row) => row.Id || row.id
    },
    {
      label: t("tunnel.type"),
      prop: "Type",
      minWidth: 130,
      cellRenderer: scope => {
        const tunnelType = scope.row.Type || scope.row.type || "";
        const tagType = "primary";
        return (
          <el-tag type={tagType} effect="plain" class="font-bold">
            {tunnelType}
          </el-tag>
        );
      }
    },
    {
      label: t("tunnel.port"),
      prop: "Port",
      minWidth: 90,
      cellRenderer: scope => {
        const port = scope.row.Port || scope.row.port || "";
        return (
          <el-tag type="info" effect="light" class="font-mono">
            {port}
          </el-tag>
        );
      }
    },
    {
      label: t("tunnel.sni"),
      prop: "SNI",
      minWidth: 130,
      formatter: (row) => row.SNI || row.sni
    },
    {
      label: t("tunnel.certificate"),
      prop: "Certificate",
      minWidth: 130,
      cellRenderer: scope => {
        const cert = scope.row.Certificate || scope.row.certificate || "";
        if (!cert) return <span class="text-gray-400">-</span>;
        const isValid = validCertSet.value.has(cert);
        return isValid ? (
          <el-tag type="info" effect="light" class="font-mono">
            {cert}
          </el-tag>
        ) : (
          <el-tooltip content="引用的证书已被删除，请编辑重选" placement="top">
            <el-tag type="danger" size="small" effect="light" class="font-bold">
              {cert} (关联失效)
            </el-tag>
          </el-tooltip>
        );
      }
    },
    {
      label: t("tunnel.clientNodes"),
      minWidth: 190,
      slot: "clientNodes"
    },
    {
      label: t("tunnel.remark"),
      prop: "Remark",
      minWidth: 130,
      formatter: (row) => row.Remark || row.remark || "-"
    },
    {
      label: t("tunnel.createTime"),
      minWidth: 160,
      prop: "CreatedAt",
      formatter: (row) => {
        const timeVal = row.CreatedAt || row.created_at;
        return timeVal && dayjs(timeVal).isValid() && dayjs(timeVal).year() > 1
          ? dayjs(timeVal).format("YYYY-MM-DD HH:mm:ss")
          : "-";
      }
    },
    {
      label: t("tunnel.operation"),
      fixed: "right",
      width: 160,
      slot: "operation"
    }
  ]);

  async function handleDelete(row: any) {
    const targetId = row.Id || row.id;
    const { code, message: msg } = await deleteTunnel({ id: targetId });
    if (code === 0) {
      message(`${t("tunnel.delete")} ID: ${targetId} success`, { type: "success" });
      onSearch();
    } else {
      message(msg, { type: "error" });
    }
  }

  async function handleDeleteClient(clientRow: any) {
    const targetId = clientRow.Id || clientRow.id;
    const { code, message: msg } = await deleteTunnelClient({ id: targetId });
    if (code === 0) {
      message(`${t("tunnel.delete")} ID: ${targetId} success`, { type: "success" });
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
    const { code, message: msg } = await deleteTunnel({ ids });
    if (code === 0) {
      message(`${t("tunnel.batchDelete")} success`, { type: "success" });
      tableRef.value.getTableRef().clearSelection();
      onSearch();
    } else {
      message(msg, { type: "error" });
    }
  }

  async function onSearch() {
    loading.value = true;
    const searchParams = toRaw(form);
    const [tunnelRes, certRes] = await Promise.all([
      getTunnelList(searchParams),
      getCertList()
    ]);

    if (certRes?.code === 0 && certRes?.data?.list) {
      validCertSet.value = new Set(certRes.data.list.map((c: any) => c.CertId || c.cert_id));
    }

    if (tunnelRes?.code === 0 && tunnelRes?.data) {
      dataList.value = tunnelRes.data.list || [];
      pagination.total = tunnelRes.data.total || dataList.value.length;
      pagination.pageSize = tunnelRes.data.pageSize || 10;
      pagination.currentPage = tunnelRes.data.currentPage || 1;
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
          type: row?.Type ?? row?.type ?? "TLS-TUNNEL",
          port: row?.Port ?? row?.port ?? "",
          sni: row?.SNI ?? row?.sni ?? "",
          certificate: row?.Certificate ?? row?.certificate ?? "",
          remark: row?.Remark ?? row?.remark ?? ""
        }
      },
      width: deviceDetection() ? "92%" : "520px",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () => h(editForm, { ref: formRef, formInline: null }),
      beforeSure: (done, { options }) => {
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline;
        FormRef.validate(async (valid: boolean) => {
          if (valid) {
            if (title === t("tunnel.addTunnel")) {
              const { code, message: msg } = await createTunnel(curData);
              if (code !== 0) {
                message(msg, { type: "error" });
                return;
              }
            } else {
              const { code, message: msg } = await updateTunnel(curData);
              if (code !== 0) {
                message(msg, { type: "error" });
                return;
              }
            }
            message(`${title} success`, { type: "success" });
            done();
            onSearch();
          }
        });
      }
    });
  }

  function openClientDialog(title = "", tunnelRow?: any, clientRow?: any) {
    const srvType = tunnelRow?.Type || tunnelRow?.type || "TLS-TUNNEL";
    const normType = srvType.toLowerCase().includes("tls") ? "tls" : "quic";
    const tid = tunnelRow ? String(tunnelRow.Id || tunnelRow.id || "") : "";

    const isUnsaved = !clientRow || !clientRow.is_saved || !(clientRow.Id || clientRow.id);
    const clientId = isUnsaved ? undefined : (clientRow.Id || clientRow.id);
    const tokenVal = clientRow?.Token || clientRow?.token || "";
    const tokenSuffix = tokenVal.length > 8 ? tokenVal.slice(-6) : tokenVal;

    let defaultName = clientRow?.Name || clientRow?.name || "";
    if (isUnsaved && !defaultName && tokenVal) {
      defaultName = `Node-${normType.toUpperCase()}-${tid}-${tokenSuffix}`;
    }

    addDialog({
      title: `${title}`,
      props: {
        formInline: {
          title,
          id: clientId,
          name: defaultName,
          type: (clientRow?.Type ?? clientRow?.type ?? normType).toLowerCase(),
          tunnel_id: clientRow?.TunnelId ?? clientRow?.tunnel_id ?? tid,
          token: tokenVal,
          remark: clientRow?.Remark ?? clientRow?.remark ?? ""
        }
      },
      width: deviceDetection() ? "92%" : "520px",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: ({ options }) =>
        h(clientEditForm, { ref: clientFormRef, formInline: options.props.formInline }),
      beforeSure: (done, { options }) => {
        const FormRef = clientFormRef.value.getRef();
        const curData = options.props.formInline;
        FormRef.validate(async (valid: boolean) => {
          if (valid) {
            if (!curData.id) {
              const { code, message: msg } = await createTunnelClient(curData);
              if (code !== 0) {
                message(msg, { type: "error" });
                return;
              }
            } else {
              const { code, message: msg } = await updateTunnelClient(curData);
              if (code !== 0) {
                message(msg, { type: "error" });
                return;
              }
            }
            message(`${title} success`, { type: "success" });
            done();
            onSearch();
          }
        });
      }
    });
  }

  const refreshingRowId = ref<any>(null);

  async function refreshSingleTunnel(row: any) {
    const targetId = row.Id || row.id;
    refreshingRowId.value = targetId;
    const res = await getTunnelList({ id: targetId });
    refreshingRowId.value = null;
    if (res?.code === 0 && Array.isArray(res.data?.list) && res.data.list.length > 0) {
      const updated = res.data.list.find((item: any) => String(item.Id || item.id) === String(targetId)) || res.data.list[0];
      row.client_nodes = updated.client_nodes || [];
      row.total_count = updated.total_count || 0;
      row.online_count = updated.online_count || 0;
      row.unsaved_count = updated.unsaved_count || 0;
      message(`刷新隧道 ID: ${targetId} 节点列表成功`, { type: "success" });
    } else {
      onSearch();
    }
  }

  onMounted(() => {
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
    refreshingRowId,
    refreshSingleTunnel,
    onSearch,
    resetForm,
    onbatchDel,
    openDialog,
    openClientDialog,
    handleDelete,
    handleDeleteClient,
    handleSizeChange,
    onSelectionCancel,
    handleCurrentChange,
    handleSelectionChange
  };
}
