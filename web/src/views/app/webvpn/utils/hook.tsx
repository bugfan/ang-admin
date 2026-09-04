import { reactive, ref, onMounted } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@/utils/message";
import {
  getWebvpnList,
  deleteWebvpn,
  updateWebvpn,
  type WebvpnSiteItem
} from "@/api/webvpn";
import { getUserGroupList, type UserGroupItem } from "@/api/user-group";
import { getHttpProxyList } from "@/api/http_proxy";

export function useWebvpn(t: Function, tableRef: any) {
  const form = reactive({
    name: "",
    http_proxy_id: ""
  });
  const dataList = ref<WebvpnSiteItem[]>([]);
  const groupList = ref<UserGroupItem[]>([]);
  const groupMap = ref<Record<number, string>>({});
  const proxyList = ref<any[]>([]);
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

  async function fetchProxies() {
    try {
      const res = await getHttpProxyList();
      proxyList.value = res.data.list || [];
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
      label: t("webvpn.name", "应用名称"),
      align: "center",
      prop: "Name",
      minWidth: 140,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpn.name", "应用名称")}</span>
      ),
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
      label: t("webvpn.targetUrl", "目标真实地址"),
      align: "center",
      prop: "TargetURL",
      minWidth: 200,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpn.targetUrl", "目标真实地址")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const url = row.TargetURL || row.target_url;
        return (
          <span class="font-mono text-xs text-(--el-text-color-regular) bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded">
            {url}
          </span>
        );
      }
    },
    {
      label: t("webvpn.accessUrl", "实际访问地址"),
      align: "center",
      prop: "full_access_url",
      minWidth: 240,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpn.accessUrl", "实际访问地址")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const fullUrl = row.full_access_url || "";
        if (!fullUrl) {
          return <span class="text-gray-400 text-xs">-</span>;
        }
        return (
          <div class="flex items-center justify-center gap-1">
            <a
              href={fullUrl}
              target="_blank"
              rel="noopener noreferrer"
              class="text-blue-500 hover:text-blue-600 hover:underline font-mono text-xs"
            >
              {fullUrl}
            </a>
          </div>
        );
      }
    },
    {
      label: t("webvpn.boundSite", "所属泛域名站点"),
      align: "center",
      prop: "http_proxy_name",
      minWidth: 160,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpn.boundSite", "所属泛域名站点")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const siteName = row.http_proxy_name || row.HttpProxyName;
        const hostName = row.http_proxy_hostname || row.HttpProxyHostname;
        if (!siteName && !hostName) {
          return <span class="text-gray-400 text-xs">{t("webvpn.unknownSite", "未知站点")}</span>;
        }
        return (
          <div class="flex flex-col items-center">
            <span class="text-xs font-medium">{siteName || hostName}</span>
            {hostName && (
              <span class="text-[11px] text-gray-400 font-mono">{hostName}</span>
            )}
          </div>
        );
      }
    },
    {
      label: t("webvpn.accessMode", "访问模式"),
      align: "center",
      prop: "is_protected",
      minWidth: 130,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpn.accessMode", "访问模式")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const isProt = (row.IsProtected ?? row.is_protected ?? 1) === 1;
        return (
          <el-tag
            size="small"
            type={isProt ? "primary" : "success"}
            effect="plain"
          >
            {isProt
              ? t("webvpn.modeProtected", "受保护 (需登录)")
              : t("webvpn.modePublic", "公开应用 (免登录)")}
          </el-tag>
        );
      }
    },
    {
      label: t("webvpn.allowedGroups", "授权用户组"),
      align: "center",
      prop: "allowed_group_ids",
      minWidth: 150,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpn.allowedGroups", "授权用户组")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const isProt = (row.IsProtected ?? row.is_protected ?? 1) === 1;
        if (!isProt) {
          return (
            <span class="text-xs text-gray-400">
              {t("webvpn.publicGroupTip", "免登录开放")}
            </span>
          );
        }
        const gIdsRaw = row.AllowedGroupIds || row.allowed_group_ids || "[]";
        let gIds: number[] = [];
        try {
          gIds = typeof gIdsRaw === "string" ? JSON.parse(gIdsRaw) : gIdsRaw;
        } catch (e) {}

        if (!Array.isArray(gIds) || gIds.length === 0) {
          return (
            <el-tag size="small" type="success" effect="plain">
              {t("webvpn.allUsersAllowed", "全员开放")}
            </el-tag>
          );
        }

        return (
          <div class="flex flex-wrap justify-center gap-1">
            {gIds.map(gid => {
              const name = groupMap.value[gid] || `Group #${gid}`;
              return (
                <el-tag size="small" type="primary" effect="light">
                  {name}
                </el-tag>
              );
            })}
          </div>
        );
      }
    },
    {
      label: t("webvpn.status", "状态"),
      align: "center",
      prop: "Status",
      width: 90,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpn.status", "状态")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const isEnabled = (row.Status ?? row.status ?? 1) === 1;
        return (
          <el-switch
            modelValue={isEnabled}
            active-text=""
            inactive-text=""
            inline-prompt
            onChange={() => handleStatusChange(row)}
          />
        );
      }
    },
    {
      label: t("common.operations", "操作"),
      fixed: "right",
      width: 170,
      align: "center",
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("common.operations", "操作")}</span>
      ),
      slot: "operation"
    }
  ];

  async function handleStatusChange(row: any) {
    const id = row.Id || row.id;
    const currentStatus = (row.Status ?? row.status ?? 1) === 1 ? 1 : 0;
    const newStatus = currentStatus === 1 ? 0 : 1;
    try {
      const res = await updateWebvpn(id, {
        status: newStatus
      });
      if (res.code === 0) {
        row.status = newStatus;
        row.Status = newStatus;
        message(t("webvpn.statusUpdated", "状态更新成功"), { type: "success" });
      } else {
        message(res.message || t("common.failed", "操作失败"), { type: "error" });
      }
    } catch (e: any) {
      message(e.message || t("common.failed", "操作失败"), { type: "error" });
    }
  }

  async function handleDelete(row: any) {
    const id = row.Id || row.id;
    try {
      const res = await deleteWebvpn(id);
      if (res.code === 0) {
        message(t("webvpn.delSuccess", "删除成功"), { type: "success" });
        onSearch();
      } else {
        message(res.message || t("common.failed", "操作失败"), { type: "error" });
      }
    } catch (e: any) {
      message(e.message || t("common.failed", "操作失败"), { type: "error" });
    }
  }

  async function onSearch() {
    loading.value = true;
    try {
      const params: any = {};
      if (form.name) params.name = form.name;
      if (form.http_proxy_id) params.http_proxy_id = form.http_proxy_id;
      const res = await getWebvpnList(params);
      if (res.code === 0) {
        dataList.value = res.data.list || [];
        pagination.total = res.data.total || 0;
      }
    } finally {
      loading.value = false;
    }
  }

  const resetForm = formEl => {
    if (!formEl) return;
    formEl.resetFields();
    form.name = "";
    form.http_proxy_id = "";
    onSearch();
  };

  onMounted(async () => {
    await fetchGroups();
    await fetchProxies();
    await onSearch();
  });

  return {
    form,
    loading,
    columns,
    dataList,
    groupList,
    proxyList,
    pagination,
    onSearch,
    resetForm,
    handleDelete
  };
}
