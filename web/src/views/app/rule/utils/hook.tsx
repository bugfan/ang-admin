import dayjs from "dayjs";
import editForm from "../form/index.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import { deviceDetection } from "@pureadmin/utils";
import {
  getRuleList,
  createRule,
  updateRule,
  deleteRule,
  type RuleItem
} from "@/api/rule";
import { type Ref, h, ref, computed, toRaw, reactive, onMounted } from "vue";

export function useRule(t: any, tableRef: Ref) {
  const form = reactive({
    name: ""
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
      type: "expand",
      slot: "expand"
    },
    {
      label: t("rule.selectionColumn") || "勾选列",
      type: "selection",
      fixed: "left",
      reserveSelection: true
    },
    {
      label: t("rule.id") || "ID",
      prop: "Id",
      width: 80,
      formatter: (row) => row.Id || row.id
    },
    {
      label: t("rule.name") || "规则名称",
      minWidth: 160,
      cellRenderer: scope => {
        const name = scope.row.Name || scope.row.name || "-";
        return (
          <span class="font-semibold text-[var(--el-text-color-primary)]">{name}</span>
        );
      }
    },
    {
      label: t("rule.matcher") || "匹配器 (Matcher)",
      minWidth: 220,
      cellRenderer: scope => {
        const mStr = scope.row.Matcher || scope.row.matcher || "";
        let mName = "-";
        try {
          const parsed = typeof mStr === "string" ? JSON.parse(mStr) : mStr;
          mName = parsed?.Name || parsed?.name || "-";
        } catch (e) {
          mName = mStr;
        }

        let tagType: "primary" | "success" | "warning" | "info" = "info";
        let labelText = mName;

        if (mName === "ip_matcher") {
          tagType = "primary";
          labelText = "ip_matcher (L4 IP)";
        } else if (mName === "http_ip_matcher") {
          tagType = "success";
          labelText = "http_ip_matcher (HTTP IP)";
        }

        return (
          <el-tag size="small" type={tagType} effect="light" class="font-mono font-medium">
            {labelText}
          </el-tag>
        );
      }
    },
    {
      label: t("rule.action") || "动作 (Action)",
      minWidth: 220,
      cellRenderer: scope => {
        const aStr = scope.row.Action || scope.row.action || "";
        let aName = "-";
        try {
          const parsed = typeof aStr === "string" ? JSON.parse(aStr) : aStr;
          aName = parsed?.Name || parsed?.name || "-";
        } catch (e) {
          aName = aStr;
        }

        let tagType: "danger" | "warning" | "info" | "success" = "info";
        let labelText = aName;

        if (aName === "reset_conn_action") {
          tagType = "danger";
          labelText = "reset_conn_action (L4 TCP重置)";
        } else if (aName === "hide_version_action") {
          tagType = "warning";
          labelText = "hide_version_action (HTTP隐藏版本)";
        }

        return (
          <el-tag size="small" type={tagType} effect="light" class="font-mono font-medium">
            {labelText}
          </el-tag>
        );
      }
    },
    {
      label: t("rule.remark") || "备注",
      prop: "Remark",
      minWidth: 140,
      formatter: (row) => row.Remark || row.remark || "-"
    },
    {
      label: t("rule.createTime") || "创建时间",
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
      label: t("rule.operation") || "操作",
      fixed: "right",
      width: 160,
      slot: "operation"
    }
  ]);

  async function handleDelete(row: any) {
    const targetId = row.Id || row.id;
    const { code, message: msg } = await deleteRule({ id: targetId });
    if (code === 0) {
      message(`删除规则 ID: ${targetId} 成功`, { type: "success" });
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
    const { code, message: msg } = await deleteRule({ ids });
    if (code === 0) {
      message("批量删除规则成功", { type: "success" });
      tableRef.value.getTableRef().clearSelection();
      onSearch();
    } else {
      message(msg, { type: "error" });
    }
  }

  async function onSearch() {
    loading.value = true;
    const searchParams = toRaw(form);
    const res = await getRuleList(searchParams);

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
          matcher: row?.Matcher ?? row?.matcher ?? JSON.stringify({ Name: "ip_matcher", Config: { Address: ["127.0.0.1"] } }),
          action: row?.Action ?? row?.action ?? JSON.stringify({ Name: "reset_conn_action", Config: { Content: "Connection reset by rule" } }),
          remark: row?.Remark ?? row?.remark ?? ""
        }
      },
      width: deviceDetection() ? "95%" : "780px",
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
            if (title === (t("rule.addRule") || "新增规则")) {
              const { code, message: msg } = await createRule(curData);
              if (code !== 0) {
                message(msg, { type: "error" });
                return;
              }
            } else {
              const { code, message: msg } = await updateRule(curData);
              if (code !== 0) {
                message(msg, { type: "error" });
                return;
              }
            }
            message(`${title} 成功`, { type: "success" });
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
