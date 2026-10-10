export interface FormItemProps {
  id?: number;
  username: string;
  password?: string;
  full_name: string;
  email: string;
  mobile: string;
  source_type: string;
  source_id: number;
  group_ids: number[];
  status: number;
  expire_at: string;
  radius_username?: string;
  cas_username?: string;
  oauth_openid?: string;
  identities?: string;
  remark: string;
}

export interface FormProps {
  formInline: FormItemProps;
}
