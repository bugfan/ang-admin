import dayjs from "dayjs";
import editForm from "../form/index.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import { deviceDetection } from "@pureadmin/utils";
import {
  getTunnelClientList,
  createTunnelClient,
  updateTunnelClient,
  deleteTunnelClient
} from "@/api/tunnel-client";
import { type Ref, h, ref, computed, toRaw, reactive, onMounted } from "vue";

export function useTunnelClient(t: any, tableRef: Ref) {
  const form = reactive({
    name: "",
    type: "",
    tunnel_id: "",
    token: ""
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
    layout: deviceDetection() ? "prev, pager, next" : "total, sizes, prev, pager, next, jumper"
  });

  const columns = computed<TableColumnList>(() => [
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
      label: t("tunnelClient.name"),
      prop: "Name",
      minWidth: 120,
      headerRenderer: () => <span class="whitespace-nowrap">{t("tunnelClient.name")}</span>,
      cellRenderer: scope => {
        const name = scope.row.Name || scope.row.name || "-";
        return (
          <span class="font-semibold text-sm text-[var(--el-text-color-primary)] break-words inline-block leading-snug py-1">
            {name}
          </span>
        );
      }
    },
    {
      label: t("tunnelClient.type"),
      prop: "Type",
      minWidth: 100,
      cellRenderer: scope => {
        const clientType = (scope.row.Type || scope.row.type || "").toLowerCase();
        const tagType = "primary";
        return (
          <el-tag type={tagType} effect="plain" class="font-bold">
            {clientType.toUpperCase()}
          </el-tag>
        );
      }
    },
    {
      label: t("tunnelClient.tunnelId"),
      prop: "TunnelId",
      minWidth: 110,
      cellRenderer: scope => {
        const tid = scope.row.TunnelId || scope.row.tunnel_id || "";
        return (
          <el-tag type="info" effect="light" class="font-mono">
            ID: {tid}
          </el-tag>
        );
      }
    },
    {
      label: t("tunnelClient.token"),
      prop: "Token",
      minWidth: 120,
      cellRenderer: scope => {
        const token = scope.row.Token || scope.row.token || "";
        return (
          <span class="font-mono font-medium text-blue-600 dark:text-blue-400">
            {token}
          </span>
        );
      }
    },
    {
      label: t("tunnelClient.status"),
      minWidth: 160,
      cellRenderer: scope => {
        const isOnline = scope.row.IsOnline ?? scope.row.is_online ?? false;
        const remoteAddr = scope.row.RemoteAddr || scope.row.remote_addr || "";
        return isOnline ? (
          <el-tooltip content={`Remote: ${remoteAddr}`} placement="top">
            <el-tag type="success" effect="light" class="font-medium inline-flex items-center">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 inline-block mr-1"></span>
              {t("tunnelClient.online")} ({remoteAddr})
            </el-tag>
          </el-tooltip>
        ) : (
          <el-tag type="info" effect="plain" class="text-gray-400 inline-flex items-center">
            <span class="w-1.5 h-1.5 rounded-full bg-rose-500 inline-block mr-1"></span>
            {t("tunnelClient.offline")}
          </el-tag>
        );
      }
    },
    {
      label: t("tunnelClient.remark"),
      prop: "Remark",
      minWidth: 110,
      align: "center",
      headerRenderer: () => <span class="whitespace-nowrap">{t("tunnelClient.remark")}</span>,
      cellRenderer: scope => {
        const remark = scope.row.Remark || scope.row.remark || "-";
        if (!remark || remark === "-") {
          return <span class="text-xs text-[var(--el-text-color-placeholder)]">-</span>;
        }
        return (
          <span class="text-xs text-[var(--el-text-color-regular)] break-words inline-block leading-snug py-1">
            {remark}
          </span>
        );
      }
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
    const { code, message: msg } = await deleteTunnelClient({ ids });
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
    const res = await getTunnelClientList(searchParams);

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
          name: row?.Name ?? row?.name ?? "",
          type: (row?.Type ?? row?.type ?? "tls").toLowerCase(),
          tunnel_id: row?.TunnelId ?? row?.tunnel_id ?? "",
          token: row?.Token ?? row?.token ?? "",
          remark: row?.Remark ?? row?.remark ?? ""
        }
      },
      width: "46%",
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
            if (title === t("tunnelClient.addClient")) {
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
