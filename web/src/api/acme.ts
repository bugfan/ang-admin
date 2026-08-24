import { http } from "@/utils/http";

export interface IssueAcmeCertParams {
  cert_id?: string;
  email: string;
  directory_url: string;
  key_type?: string;
  challenge_type: string;
  dns_provider?: string;
  dns_env_map?: Record<string, string>;
  domains: string[];
  save_cert?: boolean;
}

/** 触发 ACME 实时签发证书 */
export const issueAcmeCert = async (data: IssueAcmeCertParams) => {
  try {
    const res = await http.request<any>("post", "/api/certificate/acme-issue", {
      data
    });
    return res;
  } catch (err: any) {
    return {
      code: 1,
      message: err?.response?.data?.message || err?.message || "ACME 签发失败"
    };
  }
};

/** 获取配置模板列表 */
export const getAcmeConfigs = async (params?: object) => {
  const res = await http.request<any>("get", "/api/acme-config", { params });
  if (Array.isArray(res)) {
    return { code: 0, data: res };
  }
  return res;
};

/** 保存/更新 ACME 配置模板 */
export const saveAcmeConfig = async (data: any) => {
  try {
    if (data.id) {
      const res = await http.request<any>("put", `/api/acme-config/${data.id}`, { data });
      return { code: 0, data: res };
    } else {
      const res = await http.request<any>("post", "/api/acme-config", { data });
      return { code: 0, data: res };
    }
  } catch (err: any) {
    return { code: 1, message: err?.response?.data?.message || err?.message || "保存配置失败" };
  }
};

/** 删除 ACME 配置模板 */
export const deleteAcmeConfig = async (id: number | string) => {
  try {
    await http.request<any>("delete", `/api/acme-config/${id}`);
    return { code: 0 };
  } catch (err: any) {
    return { code: 1, message: err?.response?.data?.message || err?.message || "删除配置失败" };
  }
};

/** 基于配置 ID 触发实时签发 (长轮询) */
export const issueAcmeCertByConfigId = async (id: number | string) => {
  try {
    const res = await http.request<any>(
      "post",
      `/api/certificate/acme-issue-by-config/${id}`,
      {
        timeout: 180000 // 180秒超长超时，防止 DNS 传播导致前端超时断开
      }
    );
    return res;
  } catch (err: any) {
    return {
      code: 1,
      message: err?.response?.data?.message || err?.message || "签发失败"
    };
  }
};
