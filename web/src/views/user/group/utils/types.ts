export interface FormItemProps {
  id?: number;
  name: string;
  description: string;
  is_default: boolean;
}

export interface FormProps {
  formInline: FormItemProps;
}
