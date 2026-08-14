import "./reset.css";
import dayjs from "dayjs";
import editForm from "../form/index.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import { getKeyList, deviceDetection } from "@pureadmin/utils";
import { getUserList, registerUser, updateUser, deleteUser } from "@/api/system";
import { ElMessageBox } from "element-plus";
import { type Ref, h, ref, toRaw, computed, reactive, onMounted } from "vue";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { useUserStoreHook } from "@/store/modules/user";
import EditPen from "~icons/ep/edit-pen";
import Delete from "~icons/ep/delete";

export function useAdmin(t: any, tableRef: Ref) {
  const form = reactive({
    username: ""
  });
  const formRef = ref();
  const dataList = ref([]);
  const loading = ref(true);
  const selectedNum = ref(0);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true,
    size: deviceDetection() ? "small" : "default",
    layout: deviceDetection() ? "prev, pager, next" : "total, sizes, prev, pager, next, jumper"
  });
  const columns = computed<TableColumnList>(() => [
    {
      label: t('admin.selectionColumn'),
      type: "selection",
      fixed: "left",
      reserveSelection: true,
      hide: !useUserStoreHook().is_super_admin
    },
    {
      label: t('admin.id'),
      prop: "Id",
      width: 90
    },
    {
      label: t('admin.username'),
      prop: "Username",
      minWidth: 130
    },
    {
      label: t('admin.isSuperAdmin'),
      prop: "is_super_admin",
      minWidth: 100,
      cellRenderer: scope => {
        const isSuper = scope.row.is_super_admin ?? scope.row.IsSuperAdmin ?? false;
        return (
          <el-tag
            type={isSuper ? "success" : "info"}
            effect="plain"
          >
            {isSuper ? t('admin.isSuperAdmin') : t('admin.commonAdmin')}
          </el-tag>
        );
      }
    },
    {
      label: t('admin.createTime'),
      minWidth: 90,
      prop: "CreatedAt",
      formatter: (row) =>
        dayjs(row.CreatedAt || row.created_at).format("YYYY-MM-DD HH:mm:ss")
    },
    {
      label: t('admin.operation'),
      fixed: "right",
      width: 180,
      slot: "operation"
    }
  ]);

  async function handleDelete(row) {
    const targetId = row.Id || row.id;
    const { code, message: msg } = await deleteUser({ id: targetId });
    if (code === 0) {
      message(`${t('admin.delete')} ID: ${targetId} success`, { type: "success" });
      onSearch();
    } else {
      message(msg, { type: "error" });
    }
  }

  function handleSizeChange(val: number) {
    console.log(`${val} items per page`);
  }

  function handleCurrentChange(val: number) {
    console.log(`current page: ${val}`);
  }

  function handleSelectionChange(val) {
    selectedNum.value = val.length;
    tableRef.value.setAdaptive();
  }

  function onSelectionCancel() {
    selectedNum.value = 0;
    tableRef.value.getTableRef().clearSelection();
  }

  async function onbatchDel() {
    const curSelected = tableRef.value.getTableRef().getSelectionRows();
    const ids = curSelected.map(item => item.Id || item.id);
    const { code, message: msg } = await deleteUser({ ids });
    if (code === 0) {
      message(`${t('admin.batchDelete')} ID: ${ids.join(",")} success`, {
        type: "success"
      });
      tableRef.value.getTableRef().clearSelection();
      onSearch();
    } else {
      message(msg, { type: "error" });
    }
  }

  async function onSearch() {
    loading.value = true;
    // Using mock API getUserList for now, just filtering by username
    const { code, data } = await getUserList(toRaw(form));
    if (code === 0) {
      dataList.value = data.list;
      pagination.total = data.total;
      pagination.pageSize = data.pageSize;
      pagination.currentPage = data.currentPage;
    }
    setTimeout(() => {
      loading.value = false;
    }, 500);
  }

  const resetForm = formEl => {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  }

  function openDialog(title = "", row?: any) {
    addDialog({
      title: `${title}`,
      props: {
        formInline: {
          title,
          id: row?.Id ?? row?.id ?? undefined,
          username: row?.Username ?? row?.username ?? "",
          password: row?.Password ?? row?.password ?? "",
          repeatPassword: "",
          is_super_admin: row?.is_super_admin ?? row?.IsSuperAdmin ?? false,
          description: row?.Description ?? row?.description ?? ""
        }
      },
      width: "46%",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () => h(editForm, { ref: formRef, formInline: null }),
      beforeSure: (done, { options }) => {
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline;
        FormRef.validate(async valid => {
          if (valid) {
            if (title === t('admin.addAdmin')) {
              const { code, message: msg } = await registerUser(curData);
              if (code !== 0) {
                message(msg, { type: "error" });
                return;
              }
            } else {
              const { code, message: msg } = await updateUser(curData);
              if (code !== 0) {
                message(msg, { type: "error" });
                return;
              }
            }
            message(`${title} ${curData.username} success`, {
              type: "success"
            });
            done();
            onSearch();
          }
        });
      }
    });
  }

  onMounted(async () => {
    onSearch();
  });

  return {
    form,
    loading,
    columns,
    dataList,
    selectedNum,
    pagination,
    deviceDetection,
    onSearch,
    resetForm,
    onbatchDel,
    openDialog,
    handleDelete,
    handleSizeChange,
    onSelectionCancel,
    handleCurrentChange,
    handleSelectionChange
  };
}
