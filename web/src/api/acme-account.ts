import { http } from "@/utils/http";

/** 获取 ACME 签发账号列表 */
export const getAcmeAccounts = async (params?: object) => {
  const res = await http.request<any>("get", "/api/acme-account", { params });
  if (!res) return { code: 0, data: [] };
  if (Array.isArray(res)) {
    return { code: 0, data: res };
  }
  return res;
};

/** 保存/更新 ACME 签发账号 */
export const saveAcmeAccount = async (data: any) => {
  try {
    if (data.id) {
      const res = await http.request<any>("put", `/api/acme-account/${data.id}`, { data });
      return { code: 0, data: res };
    } else {
      const res = await http.request<any>("post", "/api/acme-account", { data });
      return { code: 0, data: res };
    }
  } catch (err: any) {
    return { code: 1, message: err?.response?.data?.message || err?.message || "保存失败" };
  }
};

/** 删除 ACME 签发账号 */
export const deleteAcmeAccount = async (id: number | string) => {
  try {
    await http.request<any>("delete", `/api/acme-account/${id}`);
    return { code: 0 };
  } catch (err: any) {
    return { code: 1, message: err?.response?.data?.message || err?.message || "删除失败" };
  }
};
