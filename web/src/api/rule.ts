import { http } from "@/utils/http";

export type RuleItem = {
  id?: number;
  name?: string;
  items?: string; // JSON string e.g. [{"Matcher":{...},"Action":{...}}]
  remark?: string;
  created_at?: string;
  updated_at?: string;
};

export type RuleResult = {
  code: number;
  message: string;
  data?: {
    list: Array<RuleItem>;
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

/** 获取 Rule 列表 (GET /api/rule) */
export const getRuleList = async (params?: object) => {
  try {
    const res = await http.request<any>("get", "/api/rule", { params });
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
    return { code: 0, message: "success", data: { list: [], total: 0, pageSize: 10, currentPage: 1 } };
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || "获取规则列表失败";
    return { code: 1, message: msg, data: { list: [], total: 0, pageSize: 10, currentPage: 1 } };
  }
};

/** 新增 Rule (POST /api/rule) */
export const createRule = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/rule", { data });
    if (res && typeof res === "object" && res.code !== undefined && res.code !== 0) {
      return res;
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || "创建规则失败";
    return { code: 1, message: msg };
  }
};

/** 修改 Rule (PUT /api/rule/:id) */
export const updateRule = async (data: any) => {
  try {
    const id = data?.id || data?.Id;
    const res = await http.request<any>("put", `/api/rule/${id}`, { data });
    if (res && typeof res === "object" && res.code !== undefined && res.code !== 0) {
      return res;
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || "更新规则失败";
    return { code: 1, message: msg };
  }
};

/** 删除 Rule (DELETE /api/rule/:id) */
export const deleteRule = async (param: any) => {
  try {
    if (param?.ids && Array.isArray(param.ids)) {
      for (const id of param.ids) {
        await http.request<any>("delete", `/api/rule/${id}`);
      }
    } else {
      const id = typeof param === "object" ? (param.id || param.Id) : param;
      await http.request<any>("delete", `/api/rule/${id}`);
    }
    return { code: 0, message: "success" };
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || "删除规则失败";
    return { code: 1, message: msg };
  }
};
