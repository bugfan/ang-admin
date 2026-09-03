import { reactive, ref, onMounted } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@/utils/message";
import {
  getUserList,
  deleteUser,
  updateUser,
  type UserItem
} from "@/api/user";
import { getUserGroupList, type UserGroupItem } from "@/api/user-group";

export function useUser(t: Function, tableRef: any) {
  const form = reactive({
    query: "",
    source_type: "",
    status: ""
  });
  const dataList = ref<UserItem[]>([]);
  const groupList = ref<UserGroupItem[]>([]);
  const groupMap = ref<Record<number, string>>({});
  const loading = ref(true);

  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  async function fetchGroups() {
    try {
      const res = await getUserGroupList();
      groupList.value = res.data.list;
      const gMap: Record<number, string> = {};
      res.data.list.forEach((g: any) => {
        const id = g.Id || g.id;
        gMap[id] = g.Name || g.name;
      });
      groupMap.value = gMap;
    } catch (e) {}
  }

  const columns: TableColumnList = [
    {
      label: "ID",
      align: "center",
      prop: "Id",
      width: 70,
      formatter: row => row.Id || row.id
    },
    {
      label: t("identity.username", "用户名"),
      align: "center",
      prop: "Username",
      minWidth: 130,
      cellRenderer: scope => {
        const row = scope.row;
        return (
          <span class="font-semibold text-sm text-(--el-text-color-primary)">
            {row.Username || row.username}
          </span>
        );
      }
    },
    {
      label: t("identity.fullName", "姓名/昵称"),
      align: "center",
      prop: "FullName",
      minWidth: 120,
      formatter: row => row.FullName || row.full_name || "-"
    },
    {
      label: t("identity.belongGroups", "所属用户组"),
      align: "center",
      prop: "GroupIds",
      minWidth: 160,
      cellRenderer: scope => {
        const row = scope.row;
        const gIdsRaw = row.GroupIds || row.group_ids || "[]";
        let gIds: number[] = [];
        try {
          gIds = typeof gIdsRaw === "string" ? JSON.parse(gIdsRaw) : gIdsRaw;
        } catch (e) {}

        if (!Array.isArray(gIds) || gIds.length === 0) {
          return <span class="text-gray-400 text-xs">-</span>;
        }

        return (
          <div class="flex flex-wrap justify-center gap-1">
            {gIds.map(gid => {
              const name = groupMap.value[gid] || `Group #${gid}`;
              return (
                <el-tag size="small" type="info" effect="light">
                  {name}
                </el-tag>
              );
            })}
          </div>
        );
      }
    },
    
    {
      label: t("identity.mobile", "手机号"),
      align: "center",
      prop: "Mobile",
      minWidth: 120,
      formatter: row => row.Mobile || row.mobile || "-"
    },
    {
      label: t("identity.status", "状态"),
      align: "center",
      prop: "Status",
      width: 90,
      align: "center",
      cellRenderer: scope => {
        const row = scope.row;
        const id = row.Id || row.id;
        return (
          <el-switch
            v-model={row.status}
            active-value={1}
            inactive-value={0}
            onChange={async (val: number) => {
              try {
                const res = await updateUser(id, { id, status: val, username: row.Username || row.username });
                if (res && res.code === 0) {
                  message(t("common.operationSuccess", "操作成功"), { type: "success" });
                } else {
                  row.status = val === 1 ? 0 : 1;
                  message(res?.message || t("common.operationFailed", "操作失败"), { type: "error" });
                }
              } catch (e) {
                row.status = val === 1 ? 0 : 1;
              }
            }}
          />
        );
      }
    },
    {
      label: t("identity.expireAt", "有效期"),
      align: "center",
      prop: "ExpireAt",
      minWidth: 150,
      formatter: row => {
        const val = row.ExpireAt || row.expire_at;
        return val ? val : "永久有效";
      }
    },
    {
      label: t("identity.operation", "操作"),
      align: "center",
      fixed: "right",
      width: 190,
      slot: "operation"
    }
  ];

  async function onSearch() {
    loading.value = true;
    try {
      const { data } = await getUserList({
        query: form.query,
        source_type: form.source_type,
        status: form.status
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
    form.query = "";
    form.source_type = "";
    form.status = "";
    onSearch();
  }

  async function handleDelete(row: any) {
    const id = row.Id || row.id;
    try {
      const res = await deleteUser(id);
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

  onMounted(async () => {
    await fetchGroups();
    onSearch();
  });

  return {
    form,
    loading,
    columns,
    dataList,
    groupList,
    groupMap,
    pagination,
    onSearch,
    resetForm,
    handleDelete
  };
}
