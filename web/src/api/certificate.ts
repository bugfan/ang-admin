import { http } from "@/utils/http";
import { i18n } from "@/plugins/i18n";

/** Helper to extract server error message with i18n support */
const getErrorMessage = (err: any, fallback: string) => {
  const data = err?.response?.data;
  if (data?.error_key) {
    const i18nKey = `cert.${data.error_key}`;
    try {
      const translated = (i18n.global as any).t(i18nKey, data.details || {});
      if (translated && translated !== i18nKey) {
        return translated;
      }
    } catch {
      // fallback
    }
  }
  return data?.message || err?.message || fallback;
};

/** 获取证书列表 (RESTful GET /api/certificate) */
export const getCertList = async (params?: object) => {
  const res = await http.request<any>("get", "/api/certificate", { params });
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

/** 新增证书 (RESTful POST /api/certificate) */
export const createCert = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/certificate", { data });
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: getErrorMessage(err, "create failed") };
  }
};

/** 批量删除证书 (RESTful 暂不支持批量，只能逐个调用或后端支持) */
export const batchDeleteCert = async (data?: Array<any>) => {
  if (!data || data.length === 0) return { code: 0, message: "success" };
  try {
    for (const item of data) {
      await http.request("delete", `/api/certificate/${item.id}`);
    }
    return { code: 0, message: "success" };
  } catch (err: any) {
    return { code: 1, message: getErrorMessage(err, "batch delete failed") };
  }
};

/** 一键签发证书 (POST /api/certificate/:id/issue) */
export const issueCert = async (id: string | number) => {
  try {
    const res = await http.request<any>("post", `/api/certificate/${id}/issue`);
    return res;
  } catch (err: any) {
    return { code: 1, message: getErrorMessage(err, "issue failed") };
  }
};

/** 修改证书 (RESTful PUT /api/certificate/:id) */
export const updateCert = async (data: any) => {
  try {
    const id = data.id || data.Id;
    const res = await http.request<any>("put", `/api/certificate/${id}`, {
      data
    });
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: getErrorMessage(err, "update failed") };
  }
};

/** 删除证书 (RESTful DELETE /api/certificate/:id) */
export const deleteCert = async (param: any) => {
  try {
    if (param?.ids && Array.isArray(param.ids)) {
      for (const id of param.ids) {
        await http.request<any>("delete", `/api/certificate/${id}`);
      }
    } else {
      const id = typeof param === "object" ? param.id || param.Id : param;
      await http.request<any>("delete", `/api/certificate/${id}`);
    }
    return { code: 0, message: "success" };
  } catch (err: any) {
    return { code: 1, message: getErrorMessage(err, "delete failed") };
  }
};

/** 自动生成自签证书 (POST /api/certificate/generate) */
export const generateSelfSignedCert = async (data: {
  common_name: string;
  dns_names: string[];
  valid_days: number;
}) => {
  try {
    const res = await http.request<any>("post", "/api/certificate/generate", {
      data
    });
    return res;
  } catch (err: any) {
    return { code: 1, message: getErrorMessage(err, "generate failed") };
  }
};
