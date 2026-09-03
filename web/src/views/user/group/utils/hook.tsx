import { reactive, ref, onMounted } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@/utils/message";
import {
  getUserGroupList,
  deleteUserGroup,
  type UserGroupItem
} from "@/api/user-group";

export function useUserGroup(t: Function, tableRef: any) {
  const form = reactive({
    name: ""
  });
  const dataList = ref<UserGroupItem[]>([]);
  const loading = ref(true);

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
      label: t("identity.groupName", "用户组名称"),
      align: "center",
      prop: "Name",
      minWidth: 160,
      cellRenderer: scope => {
        const row = scope.row;
        const isDefault = Boolean(row.IsDefault ?? row.is_default);
        return (
          <div class="flex justify-center items-center space-x-2">
            <span class="font-semibold text-sm text-(--el-text-color-primary)">
              {row.Name || row.name}
            </span>
            {isDefault && (
              <el-tag size="small" type="success" effect="light" class="font-medium">
                {t("identity.defaultGroupTag", "默认")}
              </el-tag>
            )}
          </div>
        );
      }
    },
    {
      label: t("identity.groupDescLabel", "描述说明"),
      align: "center",
      prop: "Description",
      minWidth: 200,
      formatter: row => row.Description || row.description || "-"
    },
    {
      label: t("identity.userCount", "组内成员数"),
      align: "center",
      prop: "user_count",
      width: 120,
      align: "center",
      cellRenderer: scope => {
        const count = scope.row.user_count ?? 0;
        return (
          <el-tag size="small" type="info" effect="plain" class="font-mono">
            {count}
          </el-tag>
        );
      }
    },
    {
      label: t("identity.createTime", "创建时间"),
      align: "center",
      prop: "CreatedAt",
      minWidth: 160,
      formatter: row => {
        const tVal = row.CreatedAt || row.created_at;
        return tVal ? String(tVal).replace("T", " ").substring(0, 19) : "-";
      }
    },
    {
      label: t("identity.operation", "操作"),
      align: "center",
      fixed: "right",
      width: 150,
      slot: "operation"
    }
  ];

  async function onSearch() {
    loading.value = true;
    try {
      const { data } = await getUserGroupList({
        name: form.name
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
    onSearch();
  }

  async function handleDelete(row: any) {
    const id = row.Id || row.id;
    try {
      const res = await deleteUserGroup(id);
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
    pagination,
    onSearch,
    resetForm,
    handleDelete
  };
}
