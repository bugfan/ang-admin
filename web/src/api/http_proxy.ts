import { http } from "@/utils/http";
import { formatApiError } from "@/utils/apiError";

export type HttpProxyItem = {
  id?: number;
  name?: string;
  port?: string;
  hostname?: string;
  http?: boolean;
  tls?: boolean;
  h2?: boolean;
  hsts?: boolean;
  certificate?: string;
  proxy_headers?: string;
  compress?: boolean;
  rules?: string;
  real_ip?: string;
  tunnel_type?: string;
  tunnel_id?: string;
  tunnel_token?: string;
  dns_resolver?: string;
  location_json?: string;
  remark?: string;
  created_at?: string;
  updated_at?: string;
};

export type HttpProxyResult = {
  code: number;
  message: string;
  data?: {
    list: Array<HttpProxyItem>;
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

/** 获取 HTTP Proxy 列表 (GET /api/http-proxy) */
export const getHttpProxyList = async (params?: object) => {
  try {
    const res = await http.request<any>("get", "/api/http-proxy", { params });
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
      message: formatApiError(err, "http", "获取 HTTP 代理列表失败"),
      data: { list: [], total: 0, pageSize: 10, currentPage: 1 }
    };
  }
};

/** 新增 HTTP Proxy (POST /api/http-proxy) */
export const createHttpProxy = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/http-proxy", { data });
    if (
      res &&
      typeof res === "object" &&
      res.code !== undefined &&
      res.code !== 0
    ) {
      return { code: res.code, message: formatApiError(res, "http", "创建 HTTP 代理失败") };
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "http", "创建 HTTP 代理失败") };
  }
};

/** 修改 HTTP Proxy (PUT /api/http-proxy/:id) */
export const updateHttpProxy = async (data: any) => {
  try {
    const id = data?.id || data?.Id;
    const res = await http.request<any>("put", `/api/http-proxy/${id}`, {
      data
    });
    if (
      res &&
      typeof res === "object" &&
      res.code !== undefined &&
      res.code !== 0
    ) {
      return { code: res.code, message: formatApiError(res, "http", "更新 HTTP 代理失败") };
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "http", "更新 HTTP 代理失败") };
  }
};

/** 删除 HTTP Proxy (DELETE /api/http-proxy/:id) */
export const deleteHttpProxy = async (param: any) => {
  try {
    if (param?.ids && Array.isArray(param.ids)) {
      for (const id of param.ids) {
        await http.request<any>("delete", `/api/http-proxy/${id}`);
      }
    } else {
      const id = typeof param === "object" ? param.id || param.Id : param;
      await http.request<any>("delete", `/api/http-proxy/${id}`);
    }
    return { code: 0, message: "success" };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "http", "删除 HTTP 代理失败") };
  }
};
