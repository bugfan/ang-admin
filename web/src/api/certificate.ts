import { http } from "@/utils/http";
import { formatApiError } from "@/utils/apiError";

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
    if (res && res.code !== undefined && res.code !== 0) {
      return { code: res.code, message: formatApiError(res, "cert", "create failed") };
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "cert", "create failed") };
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
    return { code: 1, message: formatApiError(err, "cert", "batch delete failed") };
  }
};

/** 一键签发证书 (POST /api/certificate/:id/issue) */
export const issueCert = async (id: string | number) => {
  try {
    // 设置 5 分钟超时，因为 ACME DNS 验证最长可能需要数分钟
    const res = await http.request<any>("post", `/api/certificate/${id}/issue`, { timeout: 300000 });
    return res;
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "cert", "issue failed") };
  }
};

/** 修改证书 (RESTful PUT /api/certificate/:id) */
export const updateCert = async (data: any) => {
  try {
    const id = data.id || data.Id;
    const res = await http.request<any>("put", `/api/certificate/${id}`, {
      data
    });
    if (res && res.code !== undefined && res.code !== 0) {
      return { code: res.code, message: formatApiError(res, "cert", "update failed") };
    }
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return { code: 1, message: formatApiError(err, "cert", "update failed") };
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
    return { code: 1, message: formatApiError(err, "cert", "delete failed") };
  }
};

/** 生成自签名证书 (POST /api/certificate/generate-self-signed) */
export const generateSelfSignedCert = async (data?: {
  common_name: string;
  dns_names?: string[];
  valid_days?: number;
}) => {
  try {
    const res = await http.request<any>(
      "post",
      "/api/certificate/generate-self-signed",
      { data }
    );
    return res;
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "cert", "generate self-signed cert failed")
    };
  }
};
