import { reactive } from "vue";
import type { FormRules } from "element-plus";

/** 密码正则（密码格式应为8-18位数字、字母、符号的任意两种组合） */
export const REGEXP_PWD =
  /^(?![0-9]+$)(?![a-z]+$)(?![A-Z]+$)(?!([^(0-9a-zA-Z)]|[()])+$)(?!^.*[\u4E00-\u9FA5].*$)([^(0-9a-zA-Z)]|[()]|[a-z]|[A-Z]|[0-9]){8,18}$/;

/** 自定义表单规则校验 */
export const formRules = reactive(<FormRules>{
  nickname: [{ required: true, message: "用户昵称为必填项", trigger: "blur" }],
  username: [{ required: true, message: "用户名称为必填项", trigger: "blur" }],
  password: [
    { required: true, message: "用户密码为必填项", trigger: "blur" },
    {
      validator: (rule, value, callback) => {
        if (!value) {
          callback(new Error("用户密码为必填项"));
        } else if (!REGEXP_PWD.test(value)) {
          callback(new Error("密码格式应为8-18位数字、字母、符号的任意两种组合"));
        } else {
          callback();
        }
      },
      trigger: "blur"
    }
  ],
  repeatPassword: [
    { required: true, message: "请确认用户密码", trigger: "blur" }
  ]
});
