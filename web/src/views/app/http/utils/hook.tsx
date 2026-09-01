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
import { getTunnelList } from "@/api/tunnel";
import { type Ref, h, ref, computed, toRaw, reactive, onMounted } from "vue";

export function useHttpProxy(t: any, tableRef: Ref) {
  const form = reactive({
    keyword: ""
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
    layout: deviceDetection()
      ? "prev, pager, next"
      : "total, sizes, prev, pager, next, jumper"
  });

  const tunnelMap = ref<Record<string, string>>({});

  const columns = computed<TableColumnList>(() => [
    {
      label: t("admin.selectionColumn", "勾选列"),
      type: "selection",
      fixed: "left",
      reserveSelection: true
    },
    {
      label: t("cert.id", "ID"),
      prop: "id",
      width: 70,
      align: "center",
      formatter: row => row.id || row.Id
    },
    {
      label: t("http.name", "名称"),
      prop: "Name",
      minWidth: 120,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("http.name", "名称")}</span>
      ),
      cellRenderer: scope => {
        const name = scope.row.Name || scope.row.name || "-";
        return (
          <span class="font-semibold text-sm/snug text-(--el-text-color-primary) wrap-break-word inline-block  py-1">
            {name}
          </span>
        );
      }
    },
    {
      label: t("http.hostAndPort", "监听"),
      minWidth: 160,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("http.hostAndPort", "监听")}</span>
      ),
      cellRenderer: scope => {
        const host = scope.row.Hostname || scope.row.hostname || "-";
        const port = scope.row.Port || scope.row.port || "80";
        return (
          <div class="font-mono text-xs font-normal text-(--el-text-color-secondary) py-1 inline-flex items-center gap-0.5">
            <span>{host}</span>
            <span>:{port}</span>
          </div>
        );
      }
    },
    {
      label: t("http.featureSection"),
      minWidth: 160,
      align: "center",
      cellRenderer: scope => {
        const row = scope.row;
        const isHttp = row.HTTP ?? row.http;
        const isTls = row.TLS ?? row.tls;
        const isH2 = row.H2 ?? row.h2;
        const isHsts = row.HSTS ?? row.hsts;
        const isCompress = row.Compress ?? row.compress;
        const cert = row.Certificate || row.certificate;

        const featureTags: any[] = [];
        if (isHttp) {
          featureTags.push(
            <el-tag
              size="small"
              type="info"
              effect="plain"
              class="font-mono font-bold text-[11px]"
            >
              HTTP
            </el-tag>
          );
        }
        if (isTls) {
          const tlsTip = cert
            ? `${t("http.selectCert", "证书")}: ${cert}`
            : "HTTPS (TLS)";
          featureTags.push(
            <el-tooltip content={tlsTip} placement="top">
              <el-tag
                size="small"
                type="success"
                effect="light"
                class="font-mono font-bold text-[11px] cursor-pointer"
              >
                HTTPS
              </el-tag>
            </el-tooltip>
          );
        }
        if (!isHttp && !isTls) {
          featureTags.push(
            <el-tag
              size="small"
              type="info"
              effect="plain"
              class="font-mono text-[11px]"
            >
              -
            </el-tag>
          );
        }
        if (isH2) {
          const h2Tip = cert
            ? `${t("http.selectCert", "证书")}: ${cert}`
            : "HTTP/2 (H2)";
          featureTags.push(
            <el-tooltip content={h2Tip} placement="top">
              <el-tag
                size="small"
                type="primary"
                effect="plain"
                class="font-mono font-bold text-[11px] cursor-pointer"
              >
                H2
              </el-tag>
            </el-tooltip>
          );
        }
        if (isHsts) {
          featureTags.push(
            <el-tag
              size="small"
              type="warning"
              effect="plain"
              class="font-mono font-bold text-[11px]"
            >
              HSTS
            </el-tag>
          );
        }
        if (isCompress) {
          featureTags.push(
            <el-tag
              size="small"
              type="info"
              effect="light"
              class="font-mono font-medium text-[11px]"
            >
              Compress
            </el-tag>
          );
        }

        const tagChunks: any[][] = [];
        for (let i = 0; i < featureTags.length; i += 3) {
          tagChunks.push(featureTags.slice(i, i + 3));
        }

        return (
          <div class="space-y-1 py-1">
            {tagChunks.map((chunk: any[], cIdx: number) => (
              <div key={cIdx} class="flex-c   gap-1.5 flex-wrap">
                {chunk}
              </div>
            ))}
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
        let ruleList: string[] = [];
        try {
          if (rulesStr) ruleList = typeof rulesStr === "string" ? JSON.parse(rulesStr) : rulesStr;
        } catch (e) {
          ruleList = [];
        }
        if (!Array.isArray(ruleList)) ruleList = [];
        if (ruleList.length === 0) {
          return <span class="text-gray-400 text-xs">-</span>;
        }
        return (
          <div class="flex flex-wrap justify-center gap-1">
            {ruleList.map(r => (
              <el-tag
                key={r}
                size="small"
                type="info"
                effect="plain"
                class="font-mono"
              >
                {r}
              </el-tag>
            ))}
          </div>
        );
      }
    },
    {
      label: t("http.tunnel", "Tunnel"),
      minWidth: 150,
      align: "center",
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("http.tunnel", "Tunnel")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const tunnelId = row.TunnelId || row.tunnel_id;
        const tunnelToken = row.TunnelToken || row.tunnel_token || "";
        const mapKey = `${tunnelId}|${tunnelToken}`;
        const tName = tunnelMap.value[mapKey];
        let displayName = `${(row.TunnelType || row.tunnel_type || "TLS").toUpperCase()} ${tunnelId}`;
        if (tName) {
          displayName = tName;
        }

        return (
          <div class="flex justify-center items-center h-full w-full py-1">
            {tunnelId ? (
              <el-tag
                size="small"
                type="success"
                effect="light"
                class="font-mono whitespace-nowrap"
              >
                {displayName}
              </el-tag>
            ) : (
              <span class="text-(--el-text-color-placeholder) text-xs">-</span>
            )}
          </div>
        );
      }
    },
    {
      label: t("http.backendSection"),
      minWidth: 160,
      cellRenderer: scope => {
        let locations: any[] = [];
        try {
          const locStr = scope.row.LocationJSON || scope.row.location_json;
          if (locStr)
            locations =
              typeof locStr === "string" ? JSON.parse(locStr) : locStr;
        } catch (e) {}

        if (!locations || locations.length === 0) {
          return (
            <span class="text-(--el-text-color-placeholder) text-xs">-</span>
          );
        }

        return (
          <div class="space-y-1.5 py-1">
            {locations.map((loc: any, idx: number) => {
              const path = loc.Path || "/";
              const uType = loc.Upstream?.Type || "proxy_pass";

              if (uType === "root" || uType === "alias") {
                const dir = loc.Upstream?.Data?.Dir || "./static";
                return (
                  <div
                    key={idx}
                    class="p-1.5 rounded-lg border border-(--el-border-color-lighter) bg-(--el-fill-color-light) flex flex-wrap items-center gap-1.5"
                  >
                    <el-tag
                      size="small"
                      type="info"
                      effect="dark"
                      class="font-mono font-bold"
                    >
                      {path}
                    </el-tag>
                    <el-tag
                      size="small"
                      type={uType === "root" ? "success" : "warning"}
                      effect="plain"
                      class="font-bold font-mono"
                    >
                      {uType}
                    </el-tag>
                    <span class="font-mono text-xs text-(--el-text-color-regular) bg-(--el-bg-color) px-2 py-0.5 rounded border border-(--el-border-color-lighter) truncate max-w-50">
                      ➔ {dir}
                    </span>
                  </div>
                );
              }

              const servers: any[] = loc.Upstream?.Data?.Servers || [];
              const method = loc.Upstream?.Data?.Method || "round_robin";

              const serverChunks: any[][] = [];
              for (let i = 0; i < servers.length; i += 2) {
                serverChunks.push(servers.slice(i, i + 2));
              }

              return (
                <div
                  key={idx}
                  class="p-1.5 rounded-lg border border-(--el-border-color-lighter) bg-(--el-fill-color-light) space-y-1"
                >
                  <div class="flex items-center gap-1.5 flex-wrap">
                    <el-tag
                      size="small"
                      type="info"
                      effect="dark"
                      class="font-mono font-bold"
                    >
                      {path}
                    </el-tag>
                    <el-tag
                      size="small"
                      type="primary"
                      effect="plain"
                      class="font-bold font-mono"
                    >
                      proxy_pass
                    </el-tag>
                    <el-tag
                      size="small"
                      type="info"
                      effect="plain"
                      class="text-[11px] font-mono"
                    >
                      {method}
                    </el-tag>
                  </div>

                  <div class="space-y-1 pt-0.5">
                    {serverChunks.map((chunk: any[], cIdx: number) => (
                      <div
                        key={cIdx}
                        class="flex items-center gap-1.5 flex-wrap"
                      >
                        {chunk.map((srv: any, sIdx: number) => (
                          <span
                            key={sIdx}
                            class="inline-flex items-center gap-1 font-mono text-[11px] bg-(--el-bg-color) px-1.5 py-0.5 rounded border border-(--el-border-color-lighter) shrink-0"
                          >
                            <span class="font-medium text-(--el-text-color-primary)">
                              {srv.Target}
                            </span>
                            <span class="text-[10px] text-gray-400 font-normal">
                              ({srv.Weight || 1})
                            </span>
                          </span>
                        ))}
                      </div>
                    ))}
                  </div>
                </div>
              );
            })}
          </div>
        );
      }
    },
    {
      label: t("http.remark", "备注"),
      prop: "Remark",
      minWidth: 110,
      align: "center",
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("http.remark", "备注")}</span>
      ),
      cellRenderer: scope => {
        const remark = scope.row.Remark || scope.row.remark || "-";
        if (!remark || remark === "-") {
          return (
            <span class="text-xs text-(--el-text-color-placeholder)">-</span>
          );
        }
        return (
          <span class="text-xs/snug text-(--el-text-color-regular) wrap-break-word inline-block  py-1">
            {remark}
          </span>
        );
      }
    },
    {
      label: t("admin.createTime", "创建时间"),
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
      message(
        `${t("http.delete", "删除")} ID: ${targetId} ${t("common.success", "成功")}`,
        { type: "success" }
      );
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
      message(
        `${t("http.batchDelete", "批量删除")} ${t("common.success", "成功")}`,
        { type: "success" }
      );
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
          proxy_headers:
            row?.ProxyHeaders ?? row?.proxy_headers ?? JSON.stringify([]),
          compress: row?.Compress ?? row?.compress ?? false,
          rules: row?.Rules ?? row?.rules ?? JSON.stringify([]),
          real_ip: row?.RealIp ?? row?.real_ip ?? "",
          tunnel_type: row?.TunnelType ?? row?.tunnel_type ?? "",
          tunnel_id: row?.TunnelId ?? row?.tunnel_id ?? "",
          tunnel_token: row?.TunnelToken ?? row?.tunnel_token ?? "",
          dns_resolver: row?.DNSResolver ?? row?.dns_resolver ?? "",
          location_json:
            row?.LocationJSON ??
            row?.location_json ??
            JSON.stringify(
              [
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
              ],
              null,
              2
            ),
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
            message(`${title} ${t("common.success", "成功")}`, {
              type: "success"
            });
            done();
            onSearch();
          }
        });
      }
    });
  }

  async function fetchTunnelNames() {
    try {
      const res = await getTunnelList();
      let list: any[] = [];
      if (Array.isArray(res?.data?.list)) list = res.data.list;
      else if (Array.isArray(res?.data)) list = res.data;
      else if (Array.isArray(res)) list = res;

      const map: Record<string, string> = {};
      list.forEach((tItem: any) => {
        const tid = String(tItem.Id || tItem.id);
        const cNodes = tItem.client_nodes || tItem.ClientNodes || [];
        cNodes.forEach((c: any) => {
          const cName = c.Name || c.name || "";
          if (cName) {
            map[`${tid}|${c.token || c.Token}`] = cName;
          }
        });
      });
      tunnelMap.value = map;
    } catch (e) {}
  }

  onMounted(() => {
    fetchTunnelNames();
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
