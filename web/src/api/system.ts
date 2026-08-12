import { http } from "@/utils/http";

type Result = {
  code: number;
  message: string;
  data?: Array<any>;
};

type ResultTable = {
  code: number;
  message: string;
  data?: {
    /** 列表数据 */
    list: Array<any>;
    /** 总条目数 */
    total?: number;
    /** 每页显示条目个数 */
    pageSize?: number;
    /** 当前页数 */
    currentPage?: number;
  };
};

/** 获取系统管理-用户管理列表 */
export const getUserList = (data?: object) => {
  return http.request<ResultTable>("post", "/api/users", { data });
};

/** 注册用户 */
export const registerUser = (data?: object) => {
  return http.request<Result>("post", "/api/register", { data });
};

/** 修改用户 */
export const updateUser = (data?: object) => {
  return http.request<Result>("post", "/api/updateUser", { data });
};

/** 删除用户 */
export const deleteUser = (data?: object) => {
  return http.request<Result>("post", "/api/deleteUser", { data });
};
