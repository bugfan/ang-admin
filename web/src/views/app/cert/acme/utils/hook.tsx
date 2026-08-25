import dayjs from "dayjs";
import { message } from "@/utils/message";
import type { PaginationProps } from "@pureadmin/table";
import { deviceDetection } from "@pureadmin/utils";
import { getAcmeAccounts, deleteAcmeAccount } from "@/api/acme-account";
import { type Ref, ref, computed, toRaw, reactive, onMounted } from "vue";

export function useAcmeAccount(t: any, tableRef: Ref, emitEdit: (row: any) => void) {
  const form = reactive({
    name: "",
    provider: ""
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

  const providerLabels: Record<string, string> = {
    tencentcloud: "腾讯云",
    alidns: "阿里云",
    dnspod: "DNSPod",
    cloudflare: "Cloudflare",
    huaweicloud: "华为云",
    route53: "AWS Route53",
    godaddy: "GoDaddy",
    digitalocean: "DigitalOcean"
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
      label: t("acmeAccount.name"),
      prop: "name",
      minWidth: 180,
      cellRenderer: scope => {
        const name = scope.row.name || "-";
        const provider = scope.row.provider || "";
        const label = providerLabels[provider] || provider;
        return (
          <div class="flex flex-col text-left py-1">
            <span class="font-semibold text-sm/snug text-(--el-text-color-primary) wrap-break-word">
              {name}
            </span>
            {label && (
              <span class="text-xs text-(--el-text-color-secondary) mt-0.5">
                {label}
              </span>
            )}
          </div>
        );
      }
    },
    {
      label: t("acmeAccount.provider"),
      prop: "provider",
      minWidth: 140,
      cellRenderer: scope => {
        const provider = scope.row.provider || "-";
        return (
          <el-tag size="small" type="primary" effect="plain" class="font-mono font-medium">
            {provider.toUpperCase()}
          </el-tag>
        );
      }
    },
    {
      label: t("acmeAccount.disableCname"),
      prop: "disable_cname",
      minWidth: 130,
      cellRenderer: scope => {
        return (
          <el-tag size="small" type={scope.row.disable_cname ? "success" : "info"} effect="plain">
            {scope.row.disable_cname ? t("acmeAccount.enabled") : t("acmeAccount.disabled")}
          </el-tag>
        );
      }
    },
    {
      label: t("cert.createTime"),
      minWidth: 160,
      prop: "created_at",
      formatter: row => {
        const timeVal = row.created_at || row.CreatedAt;
        return timeVal && dayjs(timeVal).isValid() && dayjs(timeVal).year() > 1
          ? dayjs(timeVal).format("YYYY-MM-DD HH:mm:ss")
          : "-";
      }
    },
    {
      label: t("cert.operation"),
      fixed: deviceDetection() ? undefined : "right",
      minWidth: 150,
      slot: "operation"
    }
  ]);

  async function handleDelete(row: any) {
    const targetId = row.id || row.Id;
    const { code, message: msg } = await deleteAcmeAccount(targetId);
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
      await deleteAcmeAccount(item.id || item.Id);
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
    const res = await getAcmeAccounts(searchParams);
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
