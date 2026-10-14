export interface BaseFormItemProps {
  id?: number;
  name: string;
  hostname: string;
  port: string;
  tls: boolean;
  h2: boolean;
  certificate: string;
  auth_id: number | null;
  rules?: string;
  fallback: string;
  status: number;
  site_count?: number;
  protected_site_count?: number;
  remark: string;
}

export interface BaseFormProps {
  formInline: BaseFormItemProps;
}
