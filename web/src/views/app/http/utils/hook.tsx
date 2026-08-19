import dayjs from "dayjs";
import editForm from "../form/index.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import { deviceDetection } from "@pureadmin/utils";
import {
  getHttpProxyList,
  createHttpProxy,
  updateHttpProxy,
  deleteHttpProxy,
  type HttpProxyItem
} from "@/api/http_proxy";
import { type Ref, h, ref, computed, toRaw, reactive, onMounted } from "vue";

export function useHttpProxy(t: any, tableRef: Ref) {
  const form = reactive({
    hostname: "",
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
      label: t("admin.selectionColumn", "勾选列"),
      type: "selection",
      fixed: "left",
      reserveSelection: true
    },
    {
      label: "#",
      prop: "Id",
      width: 70,
      align: "center",
      formatter: (row) => row.Id || row.id
    },
    {
      label: t("http.name", "名称"),
      prop: "Name",
      minWidth: 140,
      cellRenderer: scope => {
        const name = scope.row.Name || scope.row.name || "-";
        return (
          <span class="font-semibold text-[var(--el-text-color-primary)]">{name}</span>
        );
      }
    },
    {
      label: t("http.hostname", "Hostname"),
      prop: "Hostname",
      minWidth: 160,
      cellRenderer: scope => {
        const host = scope.row.Hostname || scope.row.hostname || "-";
        return (
          <span class="font-mono text-xs text-[var(--el-text-color-regular)]">{host}</span>
        );
      }
    },
    {
      label: t("http.port", "监听端口"),
      prop: "Port",
      width: 100,
      align: "center",
      cellRenderer: scope => {
        const port = scope.row.Port || scope.row.port || "80";
        return (
          <el-tag size="small" type="primary" effect="plain" class="font-mono font-bold">
            {port}
          </el-tag>
        );
      }
    },
    {
      label: t("http.backendSection"),
      minWidth: 260,
      cellRenderer: scope => {
        let locations: any[] = [];
        try {
          const locStr = scope.row.LocationJSON || scope.row.location_json;
          if (locStr) locations = typeof locStr === "string" ? JSON.parse(locStr) : locStr;
        } catch (e) {}

        if (!locations || locations.length === 0) {
          return <span class="text-[var(--el-text-color-placeholder)] text-xs">-</span>;
        }

        return (
          <div class="space-y-1 py-1">
            {locations.map((loc: any, idx: number) => {
              const path = loc.Path || "/";
              const uType = loc.Upstream?.Type || "proxy_pass";
              if (uType === "root" || uType === "alias") {
                const dir = loc.Upstream?.Data?.Dir || "./static";
                return (
                  <div key={idx} class="text-xs flex items-center gap-1.5 flex-wrap">
                    <el-tag size="small" type="info" effect="light" class="font-mono font-bold">
                      {path}
                    </el-tag>
                    <el-tag size="small" type={uType === "root" ? "success" : "warning"} effect="plain" class="font-bold font-mono">
                      {uType}
                    </el-tag>
                    <span class="font-mono text-xs text-[var(--el-text-color-regular)] bg-[var(--el-fill-color-light)] px-1.5 py-0.5 rounded border border-[var(--el-border-color-lighter)]">
                      ➔ {dir}
                    </span>
                  </div>
                );
              }

              const servers = loc.Upstream?.Data?.Servers || [];
              const method = loc.Upstream?.Data?.Method || "round_robin";
              return (
                <div key={idx} class="text-xs flex items-center gap-1.5 flex-wrap">
                  <el-tag size="small" type="info" effect="light" class="font-mono font-bold">
                    {path}
                  </el-tag>
                  <span class="text-[var(--el-text-color-secondary)] text-[11px]">({method}):</span>
                  {servers.map((srv: any, sIdx: number) => (
                    <span key={sIdx} class="font-mono font-medium text-[var(--el-text-color-primary)] bg-[var(--el-fill-color-light)] px-1.5 py-0.5 rounded border border-[var(--el-border-color-lighter)] text-[11px]">
                      {srv.Target} <span class="text-gray-400">({srv.Weight || 1})</span>
                    </span>
                  ))}
                </div>
              );
            })}
          </div>
        );
      }
    },
    {
      label: t("http.featureSection"),
      minWidth: 220,
      align: "center",
      cellRenderer: scope => {
        const row = scope.row;
        const isTls = row.TLS ?? row.tls;
        const isH2 = row.H2 ?? row.h2;
        const isHsts = row.HSTS ?? row.hsts;
        const isCompress = row.Compress ?? row.compress;
        const cert = row.Certificate || row.certificate;

        return (
          <div class="flex items-center justify-center gap-1.5 flex-wrap py-1">
            <el-tooltip content={isTls ? "HTTPS ON" : "HTTP OFF"} placement="top">
              <el-tag size="small" type={isTls ? "success" : "danger"} effect={isTls ? "dark" : "plain"} class="font-bold">
                {isTls ? "HTTPS" : "HTTP"}
              </el-tag>
            </el-tooltip>
            {isH2 && (
              <el-tooltip content="HTTP/2" placement="top">
                <el-tag size="small" type="warning" effect="light" class="font-bold">H2</el-tag>
              </el-tooltip>
            )}
            {isHsts && (
              <el-tooltip content="HSTS" placement="top">
                <el-tag size="small" type="danger" effect="light" class="font-bold">HSTS</el-tag>
              </el-tooltip>
            )}
            {isCompress && (
              <el-tooltip content="Gzip/Brotli Compress" placement="top">
                <el-tag size="small" type="primary" effect="light" class="font-bold">Compress</el-tag>
              </el-tooltip>
            )}
            {cert && (
              <el-tooltip content={`Cert: ${cert}`} placement="top">
                <el-tag size="small" type="info" effect="plain" class="font-mono">
                  Cert: {cert}
                </el-tag>
              </el-tooltip>
            )}
          </div>
        );
      }
    },
    {
      label: t("http.ruleSection"),
      minWidth: 110,
      align: "center",
      cellRenderer: scope => {
        const rulesStr = scope.row.Rules || scope.row.rules || "";
        let count = 0;
        try {
          const parsed = typeof rulesStr === "string" ? JSON.parse(rulesStr) : rulesStr;
          if (Array.isArray(parsed)) count = parsed.length;
        } catch (e) {}
        return (
          <el-tag size="small" type="primary" effect="light" class="font-mono font-bold">
            {count} Rules
          </el-tag>
        );
      }
    },
    {
      label: t("admin.createTime", "创建时间"),
      minWidth: 160,
      prop: "created_at",
      formatter: (row) => {
        const timeVal = row.created_at || row.CreatedAt;
        return timeVal && dayjs(timeVal).isValid() && dayjs(timeVal).year() > 1
          ? dayjs(timeVal).format("YYYY-MM-DD HH:mm:ss")
          : "-";
      }
    },
    {
      label: t("admin.operation", "操作"),
      fixed: "right",
      width: 160,
      slot: "operation"
    }
  ]);

  async function handleDelete(row: any) {
    const targetId = row.Id || row.id;
    const { code, message: msg } = await deleteHttpProxy({ id: targetId });
    if (code === 0) {
      message(`${t("http.delete", "删除")} ID: ${targetId} ${t("http.success", "成功")}`, { type: "success" });
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
    const { code, message: msg } = await deleteHttpProxy({ ids });
    if (code === 0) {
      message(`${t("http.batchDelete", "批量删除")} ${t("http.success", "成功")}`, { type: "success" });
      tableRef.value.getTableRef().clearSelection();
      onSearch();
    } else {
      message(msg, { type: "error" });
    }
  }

  async function onSearch() {
    loading.value = true;
    const searchParams = toRaw(form);
    const res = await getHttpProxyList(searchParams);

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
          port: row?.Port ?? row?.port ?? "443",
          hostname: row?.Hostname ?? row?.hostname ?? "",
          http: row?.HTTP ?? row?.http ?? true,
          tls: row?.TLS ?? row?.tls ?? true,
          h2: row?.H2 ?? row?.h2 ?? true,
          hsts: row?.HSTS ?? row?.hsts ?? false,
          certificate: row?.Certificate ?? row?.certificate ?? "",
          proxy_headers: row?.ProxyHeaders ?? row?.proxy_headers ?? JSON.stringify([]),
          compress: row?.Compress ?? row?.compress ?? false,
          rules: row?.Rules ?? row?.rules ?? JSON.stringify([]),
          real_ip: row?.RealIp ?? row?.real_ip ?? "",
          tunnel_type: row?.TunnelType ?? row?.tunnel_type ?? "",
          tunnel_id: row?.TunnelId ?? row?.tunnel_id ?? "",
          tunnel_token: row?.TunnelToken ?? row?.tunnel_token ?? "",
          dns_resolver: row?.DNSResolver ?? row?.dns_resolver ?? "",
          location_json: row?.LocationJSON ?? row?.location_json ?? JSON.stringify([
            {
              Path: "/",
              Upstream: {
                Type: "proxy_pass",
                Data: {
                  Method: "round_robin",
                  Servers: [{ Target: "http://127.0.0.1:8080", Weight: 1 }]
                }
              }
            }
          ], null, 2),
          remark: row?.Remark ?? row?.remark ?? ""
        }
      },
      width: deviceDetection() ? "95%" : "880px",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: ({ options }) =>
        h(editForm, { ref: formRef, formInline: options.props.formInline }),
      beforeSure: (done, { options }) => {
        const FormRef = formRef.value.getRef();
        if (formRef.value?.syncLocationJSON) {
          formRef.value.syncLocationJSON();
        }
        const curData = options.props.formInline;
        FormRef.validate(async (valid: boolean) => {
          if (valid) {
            if (title === t("http.addHttp")) {
              const { code, message: msg } = await createHttpProxy(curData);
              if (code !== 0) {
                message(msg, { type: "error" });
                return;
              }
            } else {
              const { code, message: msg } = await updateHttpProxy(curData);
              if (code !== 0) {
                message(msg, { type: "error" });
                return;
              }
            }
            message(`${title} ${t("http.success", "成功")}`, { type: "success" });
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
