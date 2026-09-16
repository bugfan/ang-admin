export interface FormItemProps {
  id?: number;
  name: string;
  type: string; // local, cas, radius
  enabled: boolean;
  config_json: string;
  remark: string;
}

export interface FormProps {
  formInline: FormItemProps;
}
