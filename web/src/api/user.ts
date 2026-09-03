import { http } from "@/utils/http";

export type UserResult = {
  code: number;
  message: string;
  data: {
    /** 头像 */
    avatar: string;
    /** 用户名 */
    username: string;
    /** 昵称 */
    nickname: string;
    /** 当前登录用户的角色 */
    roles: Array<string>;
    /** 按钮级别权限 */
    permissions: Array<string>;
    /** `token` */
    accessToken: string;
    /** 用于调用刷新`accessToken`的接口时所需的`token` */
    refreshToken: string;
    /** `accessToken`的过期时间（格式'xxxx/xx/xx xx:xx:xx'） */
    expires: Date;
  };
};

export type RefreshTokenResult = {
  code: number;
  message: string;
  data: {
    /** `token` */
    accessToken: string;
    /** 用于调用刷新`accessToken`的接口时所需的`token` */
    refreshToken: string;
    /** `accessToken`的过期时间（格式'xxxx/xx/xx xx:xx:xx'） */
    expires: Date;
  };
};

export type UserInfo = {
  /** 头像 */
  avatar: string;
  /** 用户名 */
  username: string;
  /** 昵称 */
  nickname: string;
  /** 邮箱 */
  email: string;
  /** 联系电话 */
  phone: string;
  /** 简介 */
  description: string;
};

export type UserInfoResult = {
  code: number;
  message: string;
  data: UserInfo;
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

/** 登录 */
export const getLogin = (data?: object) => {
  return http.request<UserResult>("post", "/login", { data });
};

/** 注册 */
export const getRegister = (data?: object) => {
  return http.request<ResultTable>("post", "/register", { data });
};

/** 刷新`token` */
export const refreshTokenApi = (data?: object) => {
  return http.request<RefreshTokenResult>("post", "/refresh-token", { data });
};

/** 账户设置-个人信息 */
export const getMine = (data?: object) => {
  return http.request<UserInfoResult>("get", "/mine", { data });
};

/** 账户设置-个人安全日志 */
export const getMineLogs = (data?: object) => {
  return http.request<ResultTable>("get", "/mine-logs", { data });
};

/** 验证码 */
export const getCaptcha = () => {
  return http.request<any>("get", "/captcha");
};

// ================= 业务用户管理 (User CRUD) =================

import { formatApiError } from "@/utils/apiError";

export type UserItem = {
  id?: number;
  username?: string;
  password?: string;
  full_name?: string;
  email?: string;
  mobile?: string;
  source_type?: string;
  source_id?: number;
  group_ids?: string; // JSON array string e.g. [1, 2]
  status?: number; // 1: 启用, 0: 禁用
  expire_at?: string;
  remark?: string;
  created_at?: string;
  updated_at?: string;
};

export const getUserList = async (params?: object) => {
  try {
    const res = await http.request<any>("get", "/api/user", { params });
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
      message: formatApiError(err, "user", "获取用户列表失败"),
      data: { list: [], total: 0, pageSize: 10, currentPage: 1 }
    };
  }
};

export const createUser = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/user", { data });
    if (res && typeof res.code === 'number' && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "user", "创建用户失败")
    };
  }
};

export const updateUser = async (id: number, data?: object) => {
  try {
    const res = await http.request<any>("put", `/api/user/${id}`, { data });
    if (res && typeof res.code === 'number' && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "user", "更新用户失败")
    };
  }
};

export const deleteUser = async (id: number) => {
  try {
    const res = await http.request<any>("delete", `/api/user/${id}`);
    if (res && typeof res.code === 'number' && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "user", "删除用户失败")
    };
  }
};

