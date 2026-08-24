import dayjs from "dayjs";
import { message } from "@/utils/message";
import type { PaginationProps } from "@pureadmin/table";
import { deviceDetection } from "@pureadmin/utils";
import { getAcmeConfigs, deleteAcmeConfig, saveAcmeConfig } from "@/api/acme";
import { type Ref, ref, computed, toRaw, reactive, onMounted } from "vue";

export function useAcme(t: any, tableRef: Ref, emitIssue: (row: any) => void, emitEdit: (row: any) => void) {
  const form = reactive({
    name: "",
    email: "",
    domains: "",
    dns_provider: ""
  });
  const formRef = ref();
  const dataList = ref<any[]>([]);
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

  const formatDomains = (domainsStr: string) => {
    if (!domainsStr) return [];
    return domainsStr
      .split(/[\n,;]+/)
      .map(d => d.trim())
      .filter(Boolean);
  };

  const columns = computed<TableColumnList>(() => [
    {
      label: t("cert.selectionColumn"),
      type: "selection",
      fixed: "left",
      reserveSelection: true
    },
    {
      label: t("cert.id"),
      prop: "id",
      width: 70,
      formatter: row => row.id || row.Id
    },
    {
      label: t("acme.configName"),
      prop: "name",
      minWidth: 160,
      cellRenderer: scope => {
        const name = scope.row.name || "-";
        const email = scope.row.email || "";
        return (
          <div class="flex flex-col text-left py-1">
            <span class="font-semibold text-sm/snug text-(--el-text-color-primary) wrap-break-word">
              {name}
            </span>
            {email && (
              <span class="text-xs text-(--el-text-color-secondary) mt-0.5">
                {email}
              </span>
            )}
          </div>
        );
      }
    },
    {
      label: t("acme.domains"),
      prop: "domains",
      minWidth: 180,
      cellRenderer: scope => {
        const domains = formatDomains(scope.row.domains || "");
        if (domains.length === 0) return <span class="text-gray-400">-</span>;
        return (
          <div class="flex flex-wrap gap-1">
            {domains.slice(0, 3).map((d: string) => (
              <el-tag size="small" type="info" effect="plain" class="font-mono font-medium">
                {d}
              </el-tag>
            ))}
            {domains.length > 3 && (
              <el-tooltip content={scope.row.domains} placement="top">
                <el-tag size="small" type="info" effect="plain">
                  +{domains.length - 3}
                </el-tag>
              </el-tooltip>
            )}
          </div>
        );
      }
    },
    {
      label: t("acme.dnsProvider"),
      prop: "dns_provider",
      minWidth: 140,
      cellRenderer: scope => {
        const provider = scope.row.dns_provider || "tencentcloud";
        const challenge = scope.row.challenge_type || "DNS-01";
        return (
          <div class="flex flex-col text-xs py-1">
            <span class="font-semibold text-(--el-text-color-primary)">
              {provider.toUpperCase()}
            </span>
            <span class="text-(--el-text-color-secondary)">
              {challenge}
            </span>
          </div>
        );
      }
    },
    {
      label: t("acme.autoRenew"),
      prop: "auto_renew",
      minWidth: 160,
      cellRenderer: scope => {
        const isAuto = scope.row.auto_renew;
        const days = scope.row.renew_days || 30;
        return (
          <div class="flex items-center gap-2">
            <el-switch
              v-model={scope.row.auto_renew}
              size="small"
              onChange={() => handleToggleAutoRenew(scope.row)}
            />
            <span class={`text-xs ${isAuto ? "text-green-600 font-medium" : "text-gray-400"}`}>
              {isAuto ? `到期前 ${days} 天` : "未启用"}
            </span>
          </div>
        );
      }
    },
    {
      label: t("acme.lastStatus"),
      prop: "last_status",
      minWidth: 130,
      cellRenderer: scope => {
        const status = scope.row.last_status;
        const err = scope.row.last_error;
        if (status === "SUCCESS") {
          return (
            <el-tag size="small" type="success" effect="light" class="font-medium">
              {t("acme.statusSuccess")}
            </el-tag>
          );
        } else if (status === "FAILED") {
          return (
            <el-tooltip content={err || t("acme.statusFailed")} placement="top">
              <el-tag size="small" type="danger" effect="light" class="font-medium cursor-pointer">
                {t("acme.statusFailed")}
              </el-tag>
            </el-tooltip>
          );
        }
        return (
          <el-tag size="small" type="info" effect="plain">
            {t("acme.statusPending")}
          </el-tag>
        );
      }
    },
    {
      label: t("acme.lastIssuedAt"),
      minWidth: 160,
      prop: "last_issued_at",
      formatter: row => {
        const timeVal = row.last_issued_at || row.LastIssuedAt;
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

  async function handleToggleAutoRenew(row: any) {
    try {
      const res = await saveAcmeConfig({
        id: row.id,
        name: row.name,
        email: row.email,
        directory_url: row.directory_url,
        key_type: row.key_type,
        challenge_type: row.challenge_type,
        dns_provider: row.dns_provider,
        dns_env: row.dns_env,
        domains: row.domains,
        cert_id: row.cert_id,
        disable_cname: row.disable_cname,
        auto_renew: row.auto_renew,
        renew_days: row.renew_days
      });
      if (res.code === 0) {
        message(row.auto_renew ? "已启用自动续签" : "已停用自动续签", { type: "success" });
      } else {
        row.auto_renew = !row.auto_renew;
        message(res.message || "更新状态失败", { type: "error" });
      }
    } catch {
      row.auto_renew = !row.auto_renew;
    }
  }

  async function handleDelete(row: any) {
    const targetId = row.id || row.Id;
    const { code, message: msg } = await deleteAcmeConfig(targetId);
    if (code === 0) {
      message(`${t("cert.delete")} ID: ${targetId} ${t("cert.success", "成功")}`, {
        type: "success"
      });
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
    for (const item of curSelected) {
      await deleteAcmeConfig(item.id || item.Id);
    }
    message(`${t("cert.batchDelete")} ${t("cert.success", "成功")}`, {
      type: "success"
    });
    tableRef.value.getTableRef().clearSelection();
    onSearch();
  }

  async function onSearch() {
    loading.value = true;
    const searchParams = toRaw(form);
    const res = await getAcmeConfigs(searchParams);
    if (res.code === 0 && Array.isArray(res.data)) {
      dataList.value = res.data;
      pagination.total = res.data.length;
    } else if (Array.isArray(res)) {
      dataList.value = res;
      pagination.total = res.length;
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
    handleDelete,
    handleSizeChange,
    onSelectionCancel,
    handleCurrentChange,
    handleSelectionChange
  };
}
