import { http } from "@/utils/http";
import { formatApiError } from "@/utils/apiError";

export type AuthMethodItem = {
  id?: number;
  name?: string;
  type?: string; // local, cas, radius
  enabled?: boolean;
  priority?: number;
  config_json?: string;
  remark?: string;
  user_count?: number;
  created_at?: string;
  updated_at?: string;
};

export const getAuthMethodList = async (params?: object) => {
  try {
    const res = await http.request<any>("get", "/api/auth-method", { params });
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
      message: formatApiError(err, "authMethod", "获取认证方式列表失败"),
      data: { list: [], total: 0, pageSize: 10, currentPage: 1 }
    };
  }
};

export const createAuthMethod = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/auth-method", { data });
    if (res && typeof res.code === 'number' && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "authMethod", "创建认证方式失败")
    };
  }
};

export const updateAuthMethod = async (id: number, data?: object) => {
  try {
    const res = await http.request<any>("put", `/api/auth-method/${id}`, { data });
    if (res && typeof res.code === 'number' && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "authMethod", "更新认证方式失败")
    };
  }
};

export const deleteAuthMethod = async (id: number) => {
  try {
    const res = await http.request<any>("delete", `/api/auth-method/${id}`);
    if (res && typeof res.code === 'number' && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "authMethod", "删除认证方式失败")
    };
  }
};

export const testAuthMethodConnection = async (data: { type: string; config_json: string }) => {
  try {
    const res = await http.request<any>("post", "/api/auth-method/test", { data });
    if (res && typeof res.code === 'number' && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "authMethod", "连通性测试失败")
    };
  }
};
