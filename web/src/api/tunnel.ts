import { http } from "@/utils/http";
import { i18n } from "@/plugins/i18n";

/** Helper to extract server error message with i18n support */
const getErrorMessage = (err: any, fallback: string) => {
  const data = err?.response?.data;
  if (data?.error_key) {
    const i18nKey = `tunnel.${data.error_key}`;
    try {
      const translated = (i18n.global as any).t(i18nKey, data.details || {});
      if (translated && translated !== i18nKey) {
        return translated;
      }
    } catch {
      // fallback
    }
  }
  return data?.message || err?.message || fallback;
};

/** 获取Tunnel列表 (RESTful GET /api/tunnel) */
export const getTunnelList = async (params?: object) => {
  const res = await http.request<any>("get", "/api/tunnel", { params });
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
  return res;
};

/** 新增Tunnel (RESTful POST /api/tunnel) */
export const createTunnel = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/tunnel", { data });
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: getErrorMessage(err, "create failed") };
  }
};

/** 修改Tunnel (RESTful PUT /api/tunnel/:id) */
export const updateTunnel = async (data: any) => {
  try {
    const id = data.id || data.Id;
    const res = await http.request<any>("put", `/api/tunnel/${id}`, { data });
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: getErrorMessage(err, "update failed") };
  }
};

/** 删除Tunnel (RESTful DELETE /api/tunnel/:id) */
export const deleteTunnel = async (param: any) => {
  try {
    if (param?.ids && Array.isArray(param.ids)) {
      for (const id of param.ids) {
        await http.request<any>("delete", `/api/tunnel/${id}`);
      }
    } else {
      const id = typeof param === "object" ? param.id || param.Id : param;
      await http.request<any>("delete", `/api/tunnel/${id}`);
    }
    return { code: 0, message: "success" };
  } catch (err: any) {
    return { code: 1, message: getErrorMessage(err, "delete failed") };
  }
};
