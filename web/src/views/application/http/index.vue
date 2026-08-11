<script setup lang="ts">
import { ref, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { getProxyList, createProxy, updateProxy, deleteProxy } from "@/api/proxy";
import FormDrawer from "./FormDrawer.vue";

defineOptions({
  name: "ApplicationHttp"
});

const tableData = ref([]);
const loading = ref(false);
const drawerVisible = ref(false);
const currentRow = ref({});

const loadData = async () => {
  loading.value = true;
  try {
    const res = await getProxyList({ type: "HTTP" });
    if (res.code === 200 || res.success) {
      tableData.value = res.data || res;
    } else {
      tableData.value = res;
    }
  } catch (error) {
    console.error(error);
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  loadData();
});

const handleAdd = () => {
  currentRow.value = {};
  drawerVisible.value = true;
};

const handleEdit = (row: any) => {
  currentRow.value = JSON.parse(JSON.stringify(row));
  drawerVisible.value = true;
};

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确认删除代理: ${row.name}?`, "警告", {
    type: "warning"
  }).then(async () => {
    try {
      await deleteProxy(row.id);
      ElMessage.success("删除成功");
      loadData();
    } catch (error) {
      ElMessage.error("删除失败");
    }
  });
};

const submitForm = async (submitData: any) => {
  try {
    if (submitData.id) {
      await updateProxy(submitData.id, submitData);
      ElMessage.success("更新成功");
    } else {
      await createProxy(submitData);
      ElMessage.success("创建成功");
    }
    drawerVisible.value = false;
    loadData();
  } catch (error) {
    ElMessage.error("保存失败");
  }
};
</script>

<template>
  <div class="main">
    <el-card shadow="never">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="font-medium">HTTP 代理配置</span>
          <el-button type="primary" @click="handleAdd">新增配置</el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" align="center" />
        <el-table-column prop="name" label="配置名称" width="150" />
        <el-table-column prop="hostname" label="域名 (Hostname)" />
        <el-table-column prop="port" label="端口" width="80" align="center" />
        <el-table-column label="TLS" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.tls ? 'success' : 'info'">{{ row.tls ? '开启' : '关闭' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="certificate" label="证书 ID" width="120" align="center" />
        <el-table-column label="操作" width="150" align="center" fixed="right">
          <template #default="scope">
            <el-button type="primary" link @click="handleEdit(scope.row)">编辑</el-button>
            <el-button type="danger" link @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <FormDrawer 
      v-model="drawerVisible" 
      :rowData="currentRow" 
      @submit="submitForm" 
    />
  </div>
</template>

<style scoped>
.main {
  padding: 16px;
}
</style>
