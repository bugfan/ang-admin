import { http } from "@/utils/http";

export interface AuthItem {
  id?: number;
  name?: string;
  auth_method_ids?: string;
    portal_url?: string;
    remark?: string;
  created_at?: string;
}

export type AuthResult = {
  code: number;
  message: string;
  data: {
    list: Array<AuthItem>;
    total: number;
  };
};

export const getAuthList = async (params?: object) => {
  try {
    const res = await http.request<any>("get", "/api/auth", { params });
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
      message: "获取列表失败",
      data: { list: [], total: 0, pageSize: 10, currentPage: 1 }
    };
  }
};

export const createAuth = (data: object) => {
  return http.request<any>("post", "/api/auth", { data }).then(res => {
    if (res && typeof res.code === "number" && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  });
};

export const updateAuth = (id: number, data: object) => {
  return http.request<any>("put", `/api/auth/${id}`, { data }).then(res => {
    if (res && typeof res.code === "number" && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  });
};

export const deleteAuth = (id: number) => {
  return http.request<any>("delete", `/api/auth/${id}`).then(res => {
    if (res && typeof res.code === "number" && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  });
};
