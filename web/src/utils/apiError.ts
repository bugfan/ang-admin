import { i18n } from "@/plugins/i18n";

/**
 * Universal error formatter that automatically looks up i18n translations
 * based on error_key and details from backend or local errors.
 */
export function formatApiError(
  resOrErr: any,
  defaultNamespace?: string,
  fallback = "Operation failed"
): string {
  const data = resOrErr?.response?.data || resOrErr?.data || resOrErr;
  if (!data) return fallback;

  const errorKey = data.error_key || data.errorKey;
  if (errorKey) {
    // 1. Try direct full key (e.g. "tunnelClient.tokenDuplicate")
    try {
      const direct = (i18n.global as any).t(errorKey, data.details || {});
      if (direct && direct !== errorKey) return direct;
    } catch {}

    // 2. Try namespaced key (e.g. "tunnelClient." + errorKey)
    if (defaultNamespace) {
      try {
        const fullKey = `${defaultNamespace}.${errorKey}`;
        const namespaced = (i18n.global as any).t(fullKey, data.details || {});
        if (namespaced && namespaced !== fullKey) return namespaced;
      } catch {}
    }
  }

  // 3. Fallback to message or default fallback
  return data.message || resOrErr?.message || fallback;
}
