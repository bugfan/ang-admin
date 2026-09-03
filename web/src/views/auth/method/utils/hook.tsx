import { reactive, ref, onMounted } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@/utils/message";
import {
  getAuthMethodList,
  deleteAuthMethod,
  updateAuthMethod,
  testAuthMethodConnection,
  type AuthMethodItem
} from "@/api/auth-method";

export function useAuthMethod(t: Function, tableRef: any) {
  const form = reactive({
    name: "",
    type: ""
  });
  const dataList = ref<AuthMethodItem[]>([]);
  const loading = ref(true);
  const selectedNum = ref(0);

  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  const columns: TableColumnList = [
    {
      label: "ID",
      align: "center",
      prop: "Id",
      width: 70,
      formatter: row => row.Id || row.id
    },
    {
      label: t("identity.sourceName", "认证方式名称"),
      align: "center",
      prop: "Name",
      minWidth: 150,
      cellRenderer: scope => {
        const row = scope.row;
        return (
          <span class="font-semibold text-sm text-(--el-text-color-primary)">
            {row.Name || row.name}
          </span>
        );
      }
    },
    {
      label: t("identity.sourceType", "认证类型"),
      align: "center",
      prop: "Type",
      minWidth: 140,
      cellRenderer: scope => {
        const type = String(scope.row.Type || scope.row.type || "").toLowerCase();
        let tagType: "" | "success" | "warning" | "info" | "danger" | "primary" = "info";
        let label = type;
        if (type === "local") {
          tagType = "primary";
          label = t("identity.sourceTypeLocal", "本地用户 (Local)");
        } else if (type === "cas") {
          tagType = "success";
          label = t("identity.sourceTypeCas", "CAS 单点 (v2/v3)");
        } else if (type === "radius") {
          tagType = "warning";
          label = t("identity.sourceTypeRadius", "RADIUS 认证");
        }
        return (
          <el-tag type={tagType} effect="light" class="font-medium">
            {label}
          </el-tag>
        );
      }
    },
    {
      label: t("identity.priority", "优先级"),
      align: "center",
      prop: "Priority",
      width: 90,
      align: "center",
      formatter: row => row.Priority ?? row.priority ?? 0
    },
    {
      label: t("identity.status", "状态"),
      align: "center",
      prop: "Enabled",
      width: 100,
      align: "center",
      cellRenderer: scope => {
        const row = scope.row;
        const id = row.Id || row.id;
        return (
          <el-switch
            v-model={row.enabled}
            active-value={true}
            inactive-value={false}
            onChange={async (val: boolean) => {
              try {
                const res = await updateAuthMethod(id, { id, enabled: val, name: row.Name || row.name, type: row.Type || row.type });
                if (res && res.code === 0) {
                  message(t("common.operationSuccess", "操作成功"), { type: "success" });
                } else {
                  row.enabled = !val;
                  message(res?.message || t("common.operationFailed", "操作失败"), { type: "error" });
                }
              } catch (e) {
                row.enabled = !val;
              }
            }}
          />
        );
      }
    },
    
    {
      label: t("identity.remark", "备注"),
      align: "center",
      prop: "Remark",
      minWidth: 130,
      formatter: row => row.Remark || row.remark || "-"
    },
    {
      label: t("identity.updateTime", "更新时间"),
      align: "center",
      prop: "UpdatedAt",
      minWidth: 160,
      formatter: row => {
        const tVal = row.UpdatedAt || row.updated_at;
        return tVal ? String(tVal).replace("T", " ").substring(0, 19) : "-";
      }
    },
    {
      label: t("identity.operation", "操作"),
      align: "center",
      fixed: "right",
      width: 200,
      slot: "operation"
    }
  ];

  async function onSearch() {
    loading.value = true;
    try {
      const { data } = await getAuthMethodList({
        name: form.name,
        type: form.type
      });
      dataList.value = data.list;
      pagination.total = data.total;
    } catch (e) {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  function resetForm(formEl: any) {
    if (!formEl) return;
    formEl.resetFields();
    form.name = "";
    form.type = "";
    onSearch();
  }

  async function handleDelete(row: any) {
    const id = row.Id || row.id;
    try {
      const res = await deleteAuthMethod(id);
      if (res && res.code === 0) {
        message(t("common.deleteSuccess", "删除成功"), { type: "success" });
        onSearch();
      } else {
        message(res?.message || t("common.operationFailed", "删除失败"), { type: "error" });
      }
    } catch (e) {
      // ignore
    }
  }

  async function handleTestConnection(row: any) {
    const type = row.Type || row.type;
    const config_json = row.ConfigJSON || row.config_json || "{}";
    message("正在测试连通性...", { type: "info" });
    try {
      const res = await testAuthMethodConnection({ type, config_json });
      if (res && res.code === 0) {
        message(res.message || t("identity.testSuccess", "连通性测试通过"), { type: "success" });
      } else {
        message(res?.message || t("identity.testFailed", "连通性测试失败"), { type: "error" });
      }
    } catch (e: any) {
      message(e?.message || t("identity.testFailed", "连通性测试失败"), { type: "error" });
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
    onSearch,
    resetForm,
    handleDelete,
    handleTestConnection
  };
}
