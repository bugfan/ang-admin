import editForm from "../form/index.vue";
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
import { getCertList } from "@/api/certificate";
import { type Ref, h, ref, computed, toRaw, reactive, onMounted } from "vue";

export function useTunnel(t: any, tableRef: Ref) {
  const form = reactive({
    name: "",
    type: "",
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

  const validCertSet = ref<Set<string>>(new Set());

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
      formatter: row => row.Id || row.id
    },
    {
      label: t("tunnel.name", "名称"),
      prop: "Name",
      minWidth: 120,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("tunnel.name", "名称")}</span>
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
      label: t("tunnel.type"),
      prop: "Type",
      minWidth: 130,
      cellRenderer: scope => {
        const rawType = scope.row.Type || scope.row.type || "";
        const displayType = rawType.toUpperCase().includes("QUIC") ? "QUIC" : "TLS";
        const tagType = "primary";
        return (
          <el-tag type={tagType} effect="plain" class="font-bold">
            {displayType}
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
      label: t("tunnel.remark"),
      prop: "Remark",
      minWidth: 110,
      align: "center",
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("tunnel.remark")}</span>
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
      label: t("tunnel.operation"),
      fixed: "right",
      width: 140,
      slot: "operation"
    }
  ]);

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
    onSearch();
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val;
    onSearch();
  }

  function handleSelectionChange(val: any[]) {
    selectedNum.value = val.length;
  }

  function onSelectionCancel() {
    selectedNum.value = 0;
    tableRef.value.getTableRef().clearSelection();
  }

  async function handleDelete(row: any) {
    const id = row.Id || row.id;
    const { code, message: msg } = await deleteTunnel({ ids: [id] });
    if (code === 0) {
      message(`${t("tunnel.deleteTunnel")} ${t("tunnel.success", "成功")}`, {
        type: "success"
      });
      onSearch();
    } else {
      message(msg, { type: "error" });
    }
  }

  async function onbatchDel() {
    const curSelected = tableRef.value.getTableRef().getSelectionRows();
    const ids = curSelected.map((item: any) => item.Id || item.id);
    const { code, message: msg } = await deleteTunnel({ ids });
    if (code === 0) {
      message(`${t("tunnel.batchDelete")} ${t("tunnel.success", "成功")}`, {
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
    const [tunnelRes, certRes] = await Promise.all([
      getTunnelList(searchParams),
      getCertList()
    ]);

    if (certRes?.code === 0 && certRes?.data?.list) {
      validCertSet.value = new Set(
        certRes.data.list.map((c: any) => c.CertId || c.cert_id)
      );
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
            message(`${title} ${t("tunnel.success", "成功")}`, {
              type: "success"
            });
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
