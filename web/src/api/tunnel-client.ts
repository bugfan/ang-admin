import { http } from "@/utils/http";

/** 获取TunnelClient配置列表 (RESTful GET /api/tunnel-client) */
export const getTunnelClientList = async (params?: object) => {
  const res = await http.request<any>("get", "/api/tunnel-client", { params });
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

/** 新增TunnelClient (RESTful POST /api/tunnel-client) */
export const createTunnelClient = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/tunnel-client", { data });
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: err?.response?.data?.message || err?.message || "create failed" };
  }
};

/** 修改TunnelClient (RESTful PUT /api/tunnel-client/:id) */
export const updateTunnelClient = async (data: any) => {
  try {
    const id = data.id || data.Id;
    const res = await http.request<any>("put", `/api/tunnel-client/${id}`, { data });
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: err?.response?.data?.message || err?.message || "update failed" };
  }
};

/** 删除TunnelClient (RESTful DELETE /api/tunnel-client/:id) */
export const deleteTunnelClient = async (param: any) => {
  try {
    if (param?.ids && Array.isArray(param.ids)) {
      for (const id of param.ids) {
        await http.request<any>("delete", `/api/tunnel-client/${id}`);
      }
    } else {
      const id = typeof param === "object" ? (param.id || param.Id) : param;
      await http.request<any>("delete", `/api/tunnel-client/${id}`);
    }
    return { code: 0, message: "success" };
  } catch (err: any) {
    return { code: 1, message: err?.response?.data?.message || err?.message || "delete failed" };
  }
};

/** 获取 ang 引擎内存中当前在线连接列表 (用于下拉选择) */
export const getActiveTunnelConnections = async () => {
  try {
    const res = await http.request<any>("get", "/api/tunnel-client/active-connections");
    return res;
  } catch (err: any) {
    return { code: 0, message: "success", data: [] };
  }
};
