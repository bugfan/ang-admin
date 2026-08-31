import { http } from "@/utils/http";
import { formatApiError } from "@/utils/apiError";

export type UdpProxyItem = {
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

export type UdpResult = {
  code: number;
  message: string;
  data?: {
    list: Array<UdpProxyItem>;
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

/** 获取 TCP 列表 (GET /api/udp-proxy) */
export const getUdpList = async (params?: object) => {
  try {
    const res = await http.request<any>("get", "/api/udp-proxy", { params });
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
      message: formatApiError(err, "udp", "获取 TCP 列表失败"),
      data: { list: [], total: 0, pageSize: 10, currentPage: 1 }
    };
  }
};

/** 新增 TCP (POST /api/udp-proxy) */
export const createUdp = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/udp-proxy", { data });
    if (
      res &&
      typeof res === "object" &&
      res.code !== undefined &&
      res.code !== 0
    ) {
      return { code: res.code, message: formatApiError(res, "udp", "创建 TCP 失败") };
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "udp", "创建 TCP 失败") };
  }
};

/** 修改 TCP (PUT /api/udp-proxy/:id) */
export const updateUdp = async (data: any) => {
  try {
    const id = data?.id || data?.Id;
    const res = await http.request<any>("put", `/api/udp-proxy/${id}`, {
      data
    });
    if (
      res &&
      typeof res === "object" &&
      res.code !== undefined &&
      res.code !== 0
    ) {
      return { code: res.code, message: formatApiError(res, "udp", "更新 TCP 失败") };
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "udp", "更新 TCP 失败") };
  }
};

/** 删除 TCP (DELETE /api/udp-proxy/:id) */
export const deleteUdp = async (param: any) => {
  try {
    if (param?.ids && Array.isArray(param.ids)) {
      for (const id of param.ids) {
        await http.request<any>("delete", `/api/udp-proxy/${id}`);
      }
    } else {
      const id = typeof param === "object" ? param.id || param.Id : param;
      await http.request<any>("delete", `/api/udp-proxy/${id}`);
    }
    return { code: 0, message: "success" };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "udp", "删除 TCP 失败") };
  }
};
