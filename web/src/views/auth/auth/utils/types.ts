export interface FormItemProps {
  id?: number;
  name: string;
  title: string;
  // local, cas, radius
  auth_method_ids: string;
  portal_url: string;
  remark: string;
}

export interface FormProps {
  formInline: FormItemProps;
}
