import { http } from "@/utils/http";
import { formatApiError } from "@/utils/apiError";

export type ClusterNodeItem = {
  id?: number;
  name?: string;
  addr?: string;
  secret?: string;
  status?: number; // 1: online, 0: offline
  last_ping?: string;
  remark?: string;
  created_at?: string;
  updated_at?: string;
};

/** 获取节点列表 (GET /api/cluster-node) */
export const getClusterNodeList = async (params?: object) => {
  try {
    const res = await http.request<any>("get", "/api/cluster-node", { params });
    if (Array.isArray(res)) {
      return {
        code: 0,
        message: "success",
        data: { list: res, total: res.length }
      };
    }
    if (res && typeof res === "object" && res.list) {
      return { code: 0, message: "success", data: res };
    }
    return { code: 0, message: "success", data: { list: [], total: 0 } };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "cluster", "获取节点列表失败"), data: { list: [], total: 0 } };
  }
};

/** 新增节点 (POST /api/cluster-node) */
export const createClusterNode = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/cluster-node", { data });
    if (
      res &&
      typeof res === "object" &&
      res.code !== undefined &&
      res.code !== 0
    ) {
      return { code: res.code, message: formatApiError(res, "cluster", "创建节点失败") };
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "cluster", "创建节点失败") };
  }
};

/** 修改节点 (PUT /api/cluster-node/:id) */
export const updateClusterNode = async (data: any) => {
  try {
    const id = data?.id || data?.Id;
    const res = await http.request<any>("put", `/api/cluster-node/${id}`, {
      data
    });
    if (
      res &&
      typeof res === "object" &&
      res.code !== undefined &&
      res.code !== 0
    ) {
      return { code: res.code, message: formatApiError(res, "cluster", "更新节点失败") };
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "cluster", "更新节点失败") };
  }
};

/** 删除节点 (DELETE /api/cluster-node/:id) */
export const deleteClusterNode = async (param: any) => {
  try {
    if (param?.ids && Array.isArray(param.ids)) {
      for (const id of param.ids) {
        await http.request<any>("delete", `/api/cluster-node/${id}`);
      }
    } else {
      const id = typeof param === "object" ? param.id || param.Id : param;
      await http.request<any>("delete", `/api/cluster-node/${id}`);
    }
    return { code: 0, message: "success" };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "cluster", "删除节点失败") };
  }
};

/** 测试节点连接 (POST /api/cluster-node/:id/ping) */
export const pingClusterNode = async (id: number) => {
  try {
    const res = await http.request<any>("post", `/api/cluster-node/${id}/ping`);
    return res;
  } catch (err: any) {
    const msg =
      err?.response?.data?.message || err?.message || "测试节点连接失败";
    return { code: 1, message: msg };
  }
};

/** 同步配置至指定节点 (POST /api/cluster-node/:id/sync) */
export const syncClusterNode = async (id: number) => {
  try {
    const res = await http.request<any>("post", `/api/cluster-node/${id}/sync`);
    return res;
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || "下发配置失败";
    return { code: 1, message: msg };
  }
};

/** 全集群同步 (POST /api/cluster-node/sync-all) */
export const syncAllClusterNodes = async () => {
  try {
    const res = await http.request<any>("post", "/api/cluster-node/sync-all");
    return res;
  } catch (err: any) {
    const msg =
      err?.response?.data?.message || err?.message || "全集群同步失败";
    return { code: 1, message: msg };
  }
};

/** 查询节点隧道概览 (GET /api/cluster-node/:id/tunnel) */
export const getClusterNodeTunnel = async (id: number) => {
  try {
    const res = await http.request<any>(
      "get",
      `/api/cluster-node/${id}/tunnel`
    );
    return res;
  } catch (err: any) {
    const msg =
      err?.response?.data?.message || err?.message || "查询节点隧道失败";
    return { code: 1, message: msg };
  }
};

/** 测试节点连通性和密钥 (POST /api/cluster-node/verify) */
export const verifyClusterNode = async (data: {
  addr: string;
  secret: string;
}) => {
  try {
    const res = await http.request<any>("post", "/api/cluster-node/verify", {
      data
    });
    return res;
  } catch (err: any) {
    const msg =
      err?.response?.data?.message || err?.message || "连通性测试失败";
    return { code: 1, message: msg };
  }
};
