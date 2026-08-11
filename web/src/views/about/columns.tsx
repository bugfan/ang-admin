export function useColumns() {
  const { pkg, lastBuildTime } = __APP_INFO__;
  const { version, engines } = pkg;
  const columns = [
    {
      label: "当前版本",
      minWidth: 100,
      cellRenderer: () => {
        return (
          <el-tag size="large" class="text-base!">
            {version}
          </el-tag>
        );
      }
    },
    {
      label: "最后编译时间",
      minWidth: 120,
      cellRenderer: () => {
        return (
          <el-tag size="large" class="text-base!">
            {lastBuildTime}
          </el-tag>
        );
      }
    },
    {
      label: "推荐 node 版本",
      minWidth: 140,
      cellRenderer: () => {
        return (
          <el-tag size="large" class="text-base!">
            {engines.node}
          </el-tag>
        );
      }
    },
    {
      label: "推荐 pnpm 版本",
      minWidth: 140,
      cellRenderer: () => {
        return (
          <el-tag size="large" class="text-base!">
            {engines.pnpm}
          </el-tag>
        );
      }
    }
  ];

  return {
    columns
  };
}
