<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import ReCol from "@/components/ReCol";
import { getUserGroupList, type UserGroupItem } from "@/api/user-group";
import type { FormProps } from "../utils/types";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    id: undefined,
    username: "",
    password: "",
    full_name: "",
    email: "",
    mobile: "",
    group_ids: [],
    status: 1,
    expire_at: "",
    remark: ""
  })
});

const { t } = useI18n();
const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const groupOptions = ref<UserGroupItem[]>([]);


async function loadGroups() {
  try {
    const res = await getUserGroupList();
    groupOptions.value = res.data.list;
  } catch (e) {}
}

onMounted(() => {
  loadGroups();
  });

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :model="newFormInline"
    label-width="120px"
    label-position="top"
    class="space-y-4"
  >
    <el-card shadow="never" class="border-(--el-border-color-lighter)! rounded-xl">
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-teal-500 rounded-full" />
          <span class="font-bold text-(--el-text-color-primary) text-sm sm:text-base">
            {{ t("identity.baseInfo", "基本信息") }}
          </span>
        </div>
      </template>

      <el-row :gutter="16">
        <re-col :value="12" :xs="24">
          <el-form-item
            :label="t('identity.username', '用户名')"
            prop="username"
            :rules="[{ required: true, message: () => t('common.nameRequired', '用户名不能为空'), trigger: 'blur' }]"
          >
            <el-input
              v-model="newFormInline.username"
              :disabled="Boolean(newFormInline.id)"
              :placeholder="t('identity.usernamePlaceholder', '请输入登录用户名')"
              clearable
            />
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item
            :label="t('identity.password', '密码')"
            prop="password"
            :rules="!newFormInline.id ? [{ required: true, min: 6, message: () => t('identity.passwordPlaceholder', '请输入 6 位以上密码'), trigger: 'blur' }] : []"
          >
            <el-input
              v-model="newFormInline.password"
              type="password"
              show-password
              :placeholder="newFormInline.id ? t('identity.passwordEditPlaceholder', '留空表示保留原密码不变') : t('identity.passwordPlaceholder', '请输入 6 位以上密码')"
            />
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('identity.fullName', '姓名/昵称')">
            <el-input
              v-model="newFormInline.full_name"
              :placeholder="t('identity.fullNamePlaceholder', '请输入用户姓名')"
              clearable
            />
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('identity.belongGroups', '所属用户组')">
            <el-select
              v-model="newFormInline.group_ids"
              multiple
              clearable
              collapse-tags
              collapse-tags-tooltip
              class="w-full"
              :placeholder="t('identity.belongGroupsPlaceholder', '请选择所属用户组')"
            >
              <el-option
                v-for="g in groupOptions"
                :key="g.id"
                :label="g.name"
                :value="g.id"
              />
            </el-select>
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('identity.mobile', '手机号码')">
            <el-input
              v-model="newFormInline.mobile"
              :placeholder="t('identity.mobilePlaceholder', '请输入联系手机号')"
              clearable
            />
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('identity.email', '电子邮箱')">
            <el-input
              v-model="newFormInline.email"
              :placeholder="t('identity.emailPlaceholder', '请输入电子邮箱')"
              clearable
            />
          </el-form-item>
        </re-col>

        
        

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('identity.expireAt', '账号有效期')">
            <el-date-picker
              v-model="newFormInline.expire_at"
              type="datetime"
              value-format="YYYY-MM-DD HH:mm:ss"
              class="!w-full"
              :placeholder="t('identity.expireAtPlaceholder', '选择过期时间 (留空永不过期)')"
              clearable
            />
          </el-form-item>
        </re-col>

        <re-col :value="12" :xs="24">
          <el-form-item :label="t('identity.status', '账号状态')">
            <div class="pt-1">
              <el-switch
                v-model="newFormInline.status"
                :active-value="1"
                :inactive-value="0"
                :active-text="t('identity.statusActive', '启用')"
                :inactive-text="t('identity.statusDisabled', '禁用')"
              />
            </div>
          </el-form-item>
        </re-col>

        <re-col :value="24" :xs="24">
          <el-form-item :label="t('identity.remark', '备注')">
            <el-input
              v-model="newFormInline.remark"
              type="textarea"
              :rows="2"
              :placeholder="t('identity.remarkPlaceholder', '请输入备注信息')"
            />
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>
  </el-form>
</template>
