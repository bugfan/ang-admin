import { http } from "@/utils/http";

export type ProxyConfig = {
  id?: string | number;
  name: string;
  type: string;
  hostname: string;
  port: number;
  tls: boolean;
  certificate: string;
  config_json: string;
  status: number;
};

export const getProxyList = (params?: object) => {
  return http.request<any>("get", "http://localhost:8080/api/app_proxy", { params });
};

export const createProxy = (data: ProxyConfig) => {
  return http.request<any>("post", "http://localhost:8080/api/app_proxy", { data });
};

export const updateProxy = (id: string | number, data: ProxyConfig) => {
  return http.request<any>("put", `http://localhost:8080/api/app_proxy/${id}`, { data });
};

export const deleteProxy = (id: string | number) => {
  return http.request<any>("delete", `http://localhost:8080/api/app_proxy/${id}`);
};
