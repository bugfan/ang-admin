import { http } from "@/utils/http";

export interface AuthSettingItem {
  id?: number;
  token_name?: string;
  token_expire?: number;
  created_at?: string;
  updated_at?: string;
}

export const getAuthSetting = () => {
  return http.request<any>("get", "/api/auth-setting/1").then(res => {
    // If bugfan/rest returns { code: 0, data: { ... } } or just the object
    if (res && res.data) {
      return { code: 0, data: res.data };
    }
    return { code: 0, data: res };
  });
};

export const updateAuthSetting = (data: object) => {
  return http.request<any>("put", "/api/auth-setting/1", { data }).then(res => {
    return { code: 0, message: "success", data: res };
  });
};
