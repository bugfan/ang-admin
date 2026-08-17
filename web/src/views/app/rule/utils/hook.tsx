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
      label: t("rule.name") || "规则组名称",
      minWidth: 180,
      cellRenderer: scope => {
        const name = scope.row.Name || scope.row.name || "-";
        return (
          <span class="font-semibold text-[var(--el-text-color-primary)]">{name}</span>
        );
      }
    },
    {
      label: "包含条目数",
      minWidth: 160,
      cellRenderer: scope => {
        const itemsStr = scope.row.Items || scope.row.items || "";
        let count = 0;
        try {
          const parsed = typeof itemsStr === "string" ? JSON.parse(itemsStr) : itemsStr;
          if (Array.isArray(parsed)) count = parsed.length;
        } catch (e) {}

        return (
          <el-tag size="small" type="primary" effect="light" class="font-mono font-medium">
            包含 {count} 条 Matcher+Action 条目
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
      message(`删除规则组 ID: ${targetId} 成功`, { type: "success" });
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
      message("批量删除规则组成功", { type: "success" });
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
    const defaultItems = JSON.stringify([
      {
        Matcher: { Name: "ip_matcher", Config: { Address: ["127.0.0.1"] } },
        Action: { Name: "reset_conn_action", Config: { Content: "Connection reset by rule" } }
      }
    ], null, 2);

    addDialog({
      title: `${title}`,
      props: {
        formInline: {
          title,
          id: row?.Id ?? row?.id ?? undefined,
          name: row?.Name ?? row?.name ?? "",
          items: row?.Items ?? row?.items ?? defaultItems,
          remark: row?.Remark ?? row?.remark ?? ""
        }
      },
      width: deviceDetection() ? "95%" : "840px",
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
