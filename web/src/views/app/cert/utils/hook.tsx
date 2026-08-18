import dayjs from "dayjs";
import editForm from "../form/index.vue";
import { message } from "@/utils/message";
import { addDialog, closeDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import { deviceDetection } from "@pureadmin/utils";
import {
  getCertList,
  createCert,
  updateCert,
  deleteCert
} from "@/api/certificate";
import { type Ref, h, ref, computed, toRaw, reactive, onMounted } from "vue";

export function useCert(t: any, tableRef: Ref) {
  const form = reactive({
    cert_id: "",
    type: "",
    subject_cn: ""
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
      label: t("cert.selectionColumn"),
      type: "selection",
      fixed: "left",
      reserveSelection: true
    },
    {
      label: t("cert.id"),
      prop: "Id",
      width: 70,
      formatter: (row) => row.Id || row.id
    },
    {
      label: t("cert.certId"),
      prop: "CertId",
      minWidth: 100,
      cellRenderer: scope => {
        const certId = scope.row.CertId || scope.row.cert_id || "";
        return (
          <el-tag type="primary" effect="plain" class="font-bold">
            {certId}
          </el-tag>
        );
      }
    },
    {
      label: t("cert.type"),
      prop: "Type",
      minWidth: 120,
      cellRenderer: scope => {
        const certType = scope.row.Type || scope.row.type || "";
        let tagType: "primary" | "success" | "warning" | "info" | "danger" = "primary";
        let labelText = certType;
        if (certType === "STD") {
          tagType = "success";
          labelText = t("cert.std");
        } else if (certType === "GM") {
          tagType = "warning";
          labelText = t("cert.gm");
        } else if (certType === "SELF-STD") {
          tagType = "info";
          labelText = t("cert.selfStd");
        }
        return (
          <el-tag type={tagType} effect="plain">
            {labelText}
          </el-tag>
        );
      }
    },
    {
      label: t("cert.subjectCn"),
      prop: "SubjectCN",
      minWidth: 140,
      formatter: (row) => row.SubjectCN || row.subject_cn || "-"
    },
    {
      label: t("cert.sans"),
      prop: "SANs",
      minWidth: 180,
      cellRenderer: scope => {
        const sansStr = scope.row.SANs || scope.row.sans || "";
        if (!sansStr) return <span class="text-gray-400">-</span>;
        const sanArr = sansStr.split(",").map((s: string) => s.trim());
        return (
          <div class="flex flex-wrap gap-1">
            {sanArr.slice(0, 3).map((item: string) => (
              <el-tag size="small" type="info" effect="light">
                {item}
              </el-tag>
            ))}
            {sanArr.length > 3 && (
              <el-tooltip content={sansStr} placement="top">
                <el-tag size="small" type="info" effect="plain">
                  +{sanArr.length - 3}
                </el-tag>
              </el-tooltip>
            )}
          </div>
        );
      }
    },
    {
      label: t("cert.validity"),
      minWidth: 210,
      cellRenderer: scope => {
        const notBefore = scope.row.NotBefore || scope.row.not_before;
        const notAfter = scope.row.NotAfter || scope.row.not_after;
        if (!notBefore || !notAfter) return <span class="text-gray-400">-</span>;
        const startDay = dayjs(notBefore);
        const endDay = dayjs(notAfter);
        if (!startDay.isValid() || !endDay.isValid() || startDay.year() <= 1 || endDay.year() <= 1) {
          return <span class="text-gray-400">-</span>;
        }
        const startStr = startDay.format("YYYY-MM-DD");
        const endStr = endDay.format("YYYY-MM-DD");
        const isExpired = dayjs().isAfter(endDay);
        return (
          <div class="flex flex-col text-xs">
            <span>{startStr} ~ {endStr}</span>
            {isExpired ? (
              <span class="text-red-500 font-semibold">{t("cert.expired")}</span>
            ) : (
              <span class="text-green-600 font-semibold">{t("cert.valid")}</span>
            )}
          </div>
        );
      }
    },
    {
      label: t("cert.remark"),
      prop: "Remark",
      minWidth: 120,
      formatter: (row) => row.Remark || row.remark || "-"
    },
    {
      label: t("cert.createTime"),
      minWidth: 160,
      prop: "created_at",
      formatter: (row) => {
        const timeVal = row.created_at || row.CreatedAt;
        return timeVal && dayjs(timeVal).isValid() && dayjs(timeVal).year() > 1
          ? dayjs(timeVal).format("YYYY-MM-DD HH:mm:ss")
          : "-";
      }
    },
    {
      label: t("cert.operation"),
      fixed: deviceDetection() ? undefined : "right",
      minWidth: 200,
      slot: "operation"
    }
  ]);

  async function handleDelete(row: any) {
    const targetId = row.Id || row.id;
    const { code, message: msg } = await deleteCert({ id: targetId });
    if (code === 0) {
      message(`${t("cert.delete")} ID: ${targetId} ${t("cert.success", "成功")}`, { type: "success" });
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
    const { code, message: msg } = await deleteCert({ ids });
    if (code === 0) {
      message(`${t("cert.batchDelete")} ${t("cert.success", "成功")}`, { type: "success" });
      tableRef.value.getTableRef().clearSelection();
      onSearch();
    } else {
      message(msg, { type: "error" });
    }
  }

  async function onSearch() {
    loading.value = true;
    const searchParams = toRaw(form);
    const { code, data } = await getCertList(searchParams);
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
          cert_id: row?.CertId ?? row?.cert_id ?? "",
          type: row?.Type ?? row?.type ?? "STD",
          key_content: row?.KeyContent ?? row?.key_content ?? "",
          cert_content: row?.CertContent ?? row?.cert_content ?? "",
          remark: row?.Remark ?? row?.remark ?? ""
        }
      },
      width: "50%",
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
            if (title === t("cert.addCert")) {
              const { code, message: msg } = await createCert(curData);
              if (code !== 0) {
                message(msg, { type: "error" });
                return;
              }
            } else {
              const { code, message: msg } = await updateCert(curData);
              if (code !== 0) {
                message(msg, { type: "error" });
                return;
              }
            }
            message(`${title} ${t("cert.success", "成功")}`, { type: "success" });
            done();
            onSearch();
          }
        });
      }
    });
  }

  function formatDialogDate(dateVal: any) {
    if (!dateVal) return "-";
    const d = dayjs(dateVal);
    return d.isValid() && d.year() > 1 ? d.format("YYYY-MM-DD HH:mm:ss") : "-";
  }

  function openDetailDialog(row: any) {
    addDialog({
      title: `${t("cert.certDetail")} (${row.CertId || row.cert_id})`,
      width: deviceDetection() ? "92%" : "520px",
      footerRenderer: ({ options, index }) => (
        <el-button type="primary" onClick={() => closeDialog(options, index)}>
          {t("cert.close")}
        </el-button>
      ),
      contentRenderer: () => (
        <div class="space-y-3 text-sm p-2">
          <div class="flex border-b pb-2">
            <span class="w-32 text-gray-500 font-medium">{t("cert.certId")}:</span>
            <span class="font-semibold text-blue-600">{row.CertId || row.cert_id}</span>
          </div>
          <div class="flex border-b pb-2">
            <span class="w-32 text-gray-500 font-medium">{t("cert.type")}:</span>
            <span>{row.Type || row.type}</span>
          </div>
          <div class="flex border-b pb-2">
            <span class="w-32 text-gray-500 font-medium">{t("cert.subjectCn")}:</span>
            <span class="font-mono">{row.SubjectCN || row.subject_cn || "-"}</span>
          </div>
          <div class="flex border-b pb-2">
            <span class="w-32 text-gray-500 font-medium">{t("cert.sans")}:</span>
            <span class="font-mono break-all">{row.SANs || row.sans || "-"}</span>
          </div>
          <div class="flex border-b pb-2">
            <span class="w-32 text-gray-500 font-medium">{t("cert.issuer")}:</span>
            <span>{row.Issuer || row.issuer || "-"}</span>
          </div>
          <div class="flex border-b pb-2">
            <span class="w-32 text-gray-500 font-medium">{t("cert.serialNumber")}:</span>
            <span class="font-mono text-xs break-all">{row.SerialNumber || row.serial_number || "-"}</span>
          </div>
          <div class="flex border-b pb-2">
            <span class="w-32 text-gray-500 font-medium">{t("cert.notBefore")}:</span>
            <span>{formatDialogDate(row.NotBefore || row.not_before)}</span>
          </div>
          <div class="flex border-b pb-2">
            <span class="w-32 text-gray-500 font-medium">{t("cert.notAfter")}:</span>
            <span>{formatDialogDate(row.NotAfter || row.not_after)}</span>
          </div>
          <div class="flex">
            <span class="w-32 text-gray-500 font-medium">{t("cert.remark")}:</span>
            <span>{row.Remark || row.remark || "-"}</span>
          </div>
        </div>
      )
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
    openDetailDialog,
    handleDelete,
    handleSizeChange,
    onSelectionCancel,
    handleCurrentChange,
    handleSelectionChange
  };
}
