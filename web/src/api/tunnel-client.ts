import { http } from "@/utils/http";
import { formatApiError } from "@/utils/apiError";

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
    if (res && res.code !== undefined && res.code !== 0) {
      return { code: res.code, message: formatApiError(res, "tunnelClient", "create failed") };
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "tunnelClient", "create failed")
    };
  }
};

/** 修改TunnelClient (RESTful PUT /api/tunnel-client/:id) */
export const updateTunnelClient = async (data: any) => {
  try {
    const id = data.id || data.Id;
    const res = await http.request<any>("put", `/api/tunnel-client/${id}`, {
      data
    });
    if (res && res.code !== undefined && res.code !== 0) {
      return { code: res.code, message: formatApiError(res, "tunnelClient", "update failed") };
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "tunnelClient", "update failed")
    };
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
      const id = typeof param === "object" ? param.id || param.Id : param;
      await http.request<any>("delete", `/api/tunnel-client/${id}`);
    }
    return { code: 0, message: "success" };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "tunnelClient", "delete failed")
    };
  }
};

/** 获取引擎内存中当前在线连接列表 (用于下拉选择) */
export const getActiveTunnelConnections = async () => {
  try {
    const res = await http.request<any>(
      "get",
      "/api/tunnel-client/active-connections"
    );
    return res;
  } catch {
    return { code: 0, message: "success", data: [] };
  }
};

/** 随机生成 10 位不重复的 Token (GET /api/tunnel-client/generate-token) */
export const generateTunnelClientToken = async () => {
  return await http.request<any>("get", "/api/tunnel-client/generate-token");
};
