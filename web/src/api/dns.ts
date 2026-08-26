import { http } from "@/utils/http";
import { formatApiError } from "@/utils/apiError";

export type DnsProxyItem = {
  id?: number;
  address?: string;
  port?: string;
  rules?: string;
  hosts_text?: string;
  hosts_json?: string;
  backend_type?: string;
  tunnel_type?: string;
  tunnel_id?: string;
  tunnel_token?: string;
  upstream_method?: string;
  upstream_servers?: string;
  remark?: string;
  created_at?: string;
  updated_at?: string;
};

export type DnsResult = {
  code: number;
  message: string;
  data?: {
    list: Array<DnsProxyItem>;
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

/** 获取 DNS列表 (GET /api/dns-proxy) */
export const getDnsList = async (params?: object) => {
  try {
    const res = await http.request<any>("get", "/api/dns-proxy", { params });
    if (Array.isArray(res)) {
      return {
        code: 0,
        message: "success",
        data: {
          list: res,
          total: res.length,
          pageSize: 10,
          currentPage: 1
        }
      };
    }
    if (res && typeof res === "object" && res.list) {
      return { code: 0, message: "success", data: res };
    }
    return {
      code: 0,
      message: "success",
      data: { list: [], total: 0, pageSize: 10, currentPage: 1 }
    };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "dns", "获取 DNS 列表失败"),
      data: { list: [], total: 0, pageSize: 10, currentPage: 1 }
    };
  }
};

/** 新增 DNS (POST /api/dns-proxy) */
export const createDns = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/dns-proxy", { data });
    if (
      res &&
      typeof res === "object" &&
      res.code !== undefined &&
      res.code !== 0
    ) {
      return { code: res.code, message: formatApiError(res, "dns", "创建 DNS 失败") };
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "dns", "创建 DNS 失败") };
  }
};

/** 修改 DNS (PUT /api/dns-proxy/:id) */
export const updateDns = async (data: any) => {
  try {
    const id = data?.id || data?.Id;
    const res = await http.request<any>("put", `/api/dns-proxy/${id}`, {
      data
    });
    if (
      res &&
      typeof res === "object" &&
      res.code !== undefined &&
      res.code !== 0
    ) {
      return { code: res.code, message: formatApiError(res, "dns", "更新 DNS 失败") };
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "dns", "更新 DNS 失败") };
  }
};

/** 删除 DNS (DELETE /api/dns-proxy/:id) */
export const deleteDns = async (param: any) => {
  try {
    if (param?.ids && Array.isArray(param.ids)) {
      for (const id of param.ids) {
        await http.request<any>("delete", `/api/dns-proxy/${id}`);
      }
    } else {
      const id = typeof param === "object" ? param.id || param.Id : param;
      await http.request<any>("delete", `/api/dns-proxy/${id}`);
    }
    return { code: 0, message: "success" };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "dns", "删除 DNS 失败") };
  }
};
