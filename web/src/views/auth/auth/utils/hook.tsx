import { reactive, ref, onMounted } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@/utils/message";
import {
  getAuthList,
  deleteAuth,
  updateAuth,
} from "@/api/auth-config";

export function useAuthMethod(t: Function, tableRef: any) {
  const form = reactive({
    name: "",
    type: ""
  });
  const loading = ref(true);
  const selectedNum = ref(0);
  const dataList = ref<any[]>([]);

  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

    const columns: TableColumnList = [
    {
      label: "ID",
      prop: "Id",
      width: 80,
      align: "center",
      cellRenderer: scope => scope.row.Id || scope.row.id
    },
    {
      label: t("identity.authConfigName", "认证名称"),
      align: "center",
      prop: "Name",
      minWidth: 160,
      cellRenderer: scope => (
        <span class="font-semibold text-sm text-(--el-text-color-primary)">
          {scope.row.Name || scope.row.name}
        </span>
      )
    },
    {
      label: t("identity.tokenName", "凭证名称"),
      align: "center",
      prop: "TokenName",
      minWidth: 120,
      cellRenderer: scope => scope.row.TokenName || scope.row.token_name
    },
    {
      label: t("identity.tokenExpire", "凭证过期时间(秒)"),
      align: "center",
      prop: "TokenExpire",
      minWidth: 140,
      cellRenderer: scope => scope.row.TokenExpire || scope.row.token_expire
    },
    {
      label: t("identity.portalUrl", "登录入口(Portal)"),
      align: "center",
      prop: "PortalUrl",
      minWidth: 200,
      cellRenderer: scope => scope.row.PortalUrl || scope.row.portal_url || "-"
    },
    {
      label: t("identity.remark", "备注"),
      align: "center",
      prop: "Remark",
      minWidth: 150,
      cellRenderer: scope => scope.row.Remark || scope.row.remark || "-"
    },
    {
      label: t("common.createdAt", "创建时间"),
      align: "center",
      prop: "CreatedAt",
      minWidth: 160,
      cellRenderer: scope => {
        const time = scope.row.CreatedAt || scope.row.created_at;
        return time ? time.replace("T", " ").split("+")[0] : "-";
      }
    },
    {
      label: t("common.operation", "操作"),
      align: "center",
      fixed: "right",
      width: 140,
      slot: "operation"
    }
  ];

  async function onSearch() {
    loading.value = true;
    try {
      const { data } = await getAuthList({
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
      const res = await deleteAuth(id);
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
    handleDelete
  };
}
