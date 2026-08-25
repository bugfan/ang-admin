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
      minWidth: 160,
      align: "center",
      cellRenderer: scope => {
        const name = scope.row.name || "-";
        return (
          <span class="font-semibold text-sm/snug text-(--el-text-color-primary) wrap-break-word">
            {name}
          </span>
        );
      }
    },
    {
      label: t("acmeAccount.provider"),
      prop: "provider",
      minWidth: 130,
      align: "center",
      cellRenderer: scope => {
        const provider = scope.row.provider || "";
        const label = providerLabels[provider] || provider.toUpperCase() || "-";
        return (
          <el-tag size="small" type="primary" effect="plain" class="font-medium">
            {label}
          </el-tag>
        );
      }
    },
    {
      label: t("acmeAccount.email"),
      prop: "email",
      minWidth: 160,
      align: "center",
      cellRenderer: scope => {
        const email = scope.row.email || "-";
        return <span class="text-xs text-(--el-text-color-regular)">{email}</span>;
      }
    },
    {
      label: t("acmeAccount.server"),
      prop: "directory_url",
      minWidth: 150,
      align: "center",
      cellRenderer: scope => {
        const url = scope.row.directory_url || "";
        let serverName = url;
        if (url.includes("letsencrypt.org/directory") && !url.includes("staging")) {
          serverName = "Let's Encrypt";
        } else if (url.includes("staging")) {
          serverName = "Let's Encrypt (Staging)";
        } else if (url.includes("zerossl.com")) {
          serverName = "ZeroSSL";
        } else if (!url) {
          serverName = "Let's Encrypt";
        }
        return (
          <el-tag size="small" type="info" effect="plain" class="font-mono">
            {serverName}
          </el-tag>
        );
      }
    },
    {
      label: t("acmeAccount.keyType"),
      prop: "key_type",
      minWidth: 110,
      align: "center",
      cellRenderer: scope => {
        const kt = scope.row.key_type || "EC256";
        return (
          <el-tag size="small" type="success" effect="plain" class="font-mono font-semibold">
            {kt}
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
