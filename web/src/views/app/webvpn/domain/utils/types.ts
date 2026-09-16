export interface DomainFormItemProps {
  id?: number;
  name: string;
  hostname: string;
  port: string;
  tls: boolean;
  h2: boolean;
  certificate: string;
  auth_id: number | null;
  fallback: string;
  status: number;
  remark: string;
}

export interface DomainFormProps {
  formInline: DomainFormItemProps;
}
