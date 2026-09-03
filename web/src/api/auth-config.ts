import { http } from "@/utils/http";

export interface AuthItem {
  id?: number;
  name?: string;
  auth_method_ids?: string;
  token_name?: string;
  portal_url?: string;
  token_expire?: number;
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

export const getAuthList = (params?: object) => {
  return http.request<AuthResult>("get", "/api/auth", { params });
};

export const createAuth = (data: object) => {
  return http.request<any>("post", "/api/auth", { data }).then(res => {
    if (res && typeof res.code === 'number' && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  });
};

export const updateAuth = (id: number, data: object) => {
  return http.request<any>("put", `/api/auth/\${id}`, { data }).then(res => {
    if (res && typeof res.code === 'number' && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  });
};

export const deleteAuth = (id: number) => {
  return http.request<any>("delete", `/api/auth/\${id}`).then(res => {
    if (res && typeof res.code === 'number' && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  });
};
