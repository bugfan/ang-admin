import { http } from "@/utils/http";
import { formatApiError } from "@/utils/apiError";

export type UserGroupItem = {
  id?: number;
  name?: string;
  description?: string;
  is_default?: boolean;
  user_count?: number;
  created_at?: string;
  updated_at?: string;
};

export const getUserGroupList = async (params?: object) => {
  try {
    const res = await http.request<any>("get", "/api/user-group", { params });
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
      message: formatApiError(err, "userGroup", "获取用户组列表失败"),
      data: { list: [], total: 0, pageSize: 10, currentPage: 1 }
    };
  }
};

export const createUserGroup = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/user-group", { data });
    if (res && typeof res.code === 'number' && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "userGroup", "创建用户组失败")
    };
  }
};

export const updateUserGroup = async (id: number, data?: object) => {
  try {
    const res = await http.request<any>("put", `/api/user-group/${id}`, { data });
    if (res && typeof res.code === 'number' && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "userGroup", "更新用户组失败")
    };
  }
};

export const deleteUserGroup = async (id: number) => {
  try {
    const res = await http.request<any>("delete", `/api/user-group/${id}`);
    if (res && typeof res.code === 'number' && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "userGroup", "删除用户组失败")
    };
  }
};
