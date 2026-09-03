export interface FormItemProps {
  id?: number;
  name: string;
   // local, cas, radius
  auth_method_ids: string;
  token_name: string;
  portal_url: string;
  token_expire: number;
  
  remark: string;
}

export interface FormProps {
  formInline: FormItemProps;
}
