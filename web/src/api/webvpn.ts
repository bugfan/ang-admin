import { http } from "@/utils/http";
import { formatApiError } from "@/utils/apiError";

export type WebvpnSiteItem = {
  id?: number;
  name?: string;
  http_proxy_id?: number;
  http_proxy_name?: string;
  http_proxy_hostname?: string;
  target_url?: string;
  prefix?: string;
  hosts?: string;
  allowed_group_ids?: string; // JSON array string e.g. "[1, 2]"
  status?: number; // 1: enabled, 0: disabled
  full_access_url?: string;
  remark?: string;
  created_at?: string;
  updated_at?: string;
};

export const getWebvpnList = async (params?: object) => {
  try {
    const res = await http.request<any>("get", "/api/webvpn-site", { params });
    const list = Array.isArray(res) ? res : (res?.data || res?.list || []);
    return {
      code: 0,
      message: "success",
      data: {
        list,
        total: list.length,
        pageSize: 10,
        currentPage: 1
      }
    };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "webvpn", "获取WebVPN应用列表失败"),
      data: { list: [], total: 0, pageSize: 10, currentPage: 1 }
    };
  }
};

export const createWebvpn = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/webvpn-site", { data });
    if (res && typeof res.code === "number" && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "webvpn", "创建WebVPN应用失败")
    };
  }
};

export const updateWebvpn = async (id: number, data?: object) => {
  try {
    const res = await http.request<any>("put", `/api/webvpn-site/${id}`, {
      data
    });
    if (res && typeof res.code === "number" && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "webvpn", "更新WebVPN应用失败")
    };
  }
};

export const deleteWebvpn = async (id: number) => {
  try {
    const res = await http.request<any>("delete", `/api/webvpn-site/${id}`);
    if (res && typeof res.code === "number" && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "webvpn", "删除WebVPN应用失败")
    };
  }
};
