import { http } from "@/utils/http";
import { formatApiError } from "@/utils/apiError";

export type TcpProxyItem = {
  id?: number;
  name?: string;
  address?: string;
  port?: string;
  rules?: string;
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

export type TcpResult = {
  code: number;
  message: string;
  data?: {
    list: Array<TcpProxyItem>;
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

/** 获取 TCP 列表 (GET /api/tcp-proxy) */
export const getTcpList = async (params?: object) => {
  try {
    const res = await http.request<any>("get", "/api/tcp-proxy", { params });
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
      message: formatApiError(err, "tcp", "获取 TCP 列表失败"),
      data: { list: [], total: 0, pageSize: 10, currentPage: 1 }
    };
  }
};

/** 新增 TCP (POST /api/tcp-proxy) */
export const createTcp = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/tcp-proxy", { data });
    if (
      res &&
      typeof res === "object" &&
      res.code !== undefined &&
      res.code !== 0
    ) {
      return { code: res.code, message: formatApiError(res, "tcp", "创建 TCP 失败") };
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "tcp", "创建 TCP 失败") };
  }
};

/** 修改 TCP (PUT /api/tcp-proxy/:id) */
export const updateTcp = async (data: any) => {
  try {
    const id = data?.id || data?.Id;
    const res = await http.request<any>("put", `/api/tcp-proxy/${id}`, {
      data
    });
    if (
      res &&
      typeof res === "object" &&
      res.code !== undefined &&
      res.code !== 0
    ) {
      return { code: res.code, message: formatApiError(res, "tcp", "更新 TCP 失败") };
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "tcp", "更新 TCP 失败") };
  }
};

/** 删除 TCP (DELETE /api/tcp-proxy/:id) */
export const deleteTcp = async (param: any) => {
  try {
    if (param?.ids && Array.isArray(param.ids)) {
      for (const id of param.ids) {
        await http.request<any>("delete", `/api/tcp-proxy/${id}`);
      }
    } else {
      const id = typeof param === "object" ? param.id || param.Id : param;
      await http.request<any>("delete", `/api/tcp-proxy/${id}`);
    }
    return { code: 0, message: "success" };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "tcp", "删除 TCP 失败") };
  }
};
