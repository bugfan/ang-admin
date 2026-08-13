import dayjs from "dayjs";
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
import { type Ref, h, ref, toRaw, reactive, onMounted } from "vue";

export function useTunnel(t: any, tableRef: Ref) {
  const form = reactive({
    type: "",
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
    layout: deviceDetection() ? "prev, pager, next" : "total, sizes, prev, pager, next, jumper"
  });

  const columns: TableColumnList = [
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
      minWidth: 140,
      cellRenderer: scope => {
        const tunnelType = scope.row.Type || scope.row.type || "";
        const tagType = tunnelType === "TLS-TUNNEL" ? "success" : "warning";
        return (
          <el-tag type={tagType} effect="plain">
            {tunnelType}
          </el-tag>
        );
      }
    },
    {
      label: t("tunnel.port"),
      prop: "Port",
      minWidth: 100,
      cellRenderer: scope => {
        const port = scope.row.Port || scope.row.port || "";
        return (
          <el-tag type="info" effect="light">
            {port}
          </el-tag>
        );
      }
    },
    {
      label: t("tunnel.sni"),
      prop: "SNI",
      minWidth: 140,
      formatter: (row) => row.SNI || row.sni
    },
    {
      label: t("tunnel.certificate"),
      prop: "Certificate",
      minWidth: 120,
      cellRenderer: scope => {
        const cert = scope.row.Certificate || scope.row.certificate || "";
        return cert ? (
          <el-tag type="primary" size="small" effect="plain">
            {cert}
          </el-tag>
        ) : (
          <span class="text-gray-400">-</span>
        );
      }
    },
    {
      label: t("tunnel.remark"),
      prop: "Remark",
      minWidth: 140,
      formatter: (row) => row.Remark || row.remark || "-"
    },
    {
      label: t("tunnel.createTime"),
      minWidth: 160,
      prop: "CreatedAt",
      formatter: (row) => {
        const timeVal = row.CreatedAt || row.created_at;
        return timeVal ? dayjs(timeVal).format("YYYY-MM-DD HH:mm:ss") : "-";
      }
    },
    {
      label: t("tunnel.operation"),
      fixed: "right",
      width: 160,
      slot: "operation"
    }
  ];

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
    const { code, data } = await getTunnelList(searchParams);
    if (code === 0 && data) {
      dataList.value = data.list || [];
      pagination.total = data.total || dataList.value.length;
      pagination.pageSize = data.pageSize || 10;
      pagination.currentPage = data.currentPage || 1;
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
      width: "46%",
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
