import { http } from "@/utils/http";

/** 获取系统管理-用户管理列表 (RESTful GET /api/admin) */
export const getUserList = async (params?: object) => {
  const res = await http.request<any>("get", "/api/admin", { params });
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

/** 新增用户 (RESTful POST /api/admin) */
export const registerUser = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/admin", { data });
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: err?.message || "create failed" };
  }
};

/** 修改用户 (RESTful PUT /api/admin/:id) */
export const updateUser = async (data: any) => {
  try {
    const id = data.id || data.Id;
    const res = await http.request<any>("put", `/api/admin/${id}`, { data });
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: err?.message || "update failed" };
  }
};

/** 删除用户 (RESTful DELETE /api/admin/:id) */
export const deleteUser = async (param: any) => {
  try {
    if (param?.ids && Array.isArray(param.ids)) {
      for (const id of param.ids) {
        await http.request<any>("delete", `/api/admin/${id}`);
      }
    } else {
      const id = typeof param === "object" ? param.id : param;
      await http.request<any>("delete", `/api/admin/${id}`);
    }
    return { code: 0, message: "success" };
  } catch (err: any) {
    return { code: 1, message: err?.message || "delete failed" };
  }
};
