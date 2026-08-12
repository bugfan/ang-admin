// 模拟后端动态生成路由
import { defineFakeRoute } from "vite-plugin-fake-server/client";
import { system, monitor, permission, frame, tabs } from "@/router/enums";

/**
 * roles：页面级别权限，这里模拟二种 "admin"、"common"
 * admin：管理员角色
 * common：普通角色
 */

const accountManagementRouter = {
  path: "/system",
  meta: {
    icon: "ri:group-line",
    title: "menus.pureAccountManagement",
    rank: system
  },
  children: [
    {
      path: "/system/admin/index",
      name: "SystemAdmin",
      meta: {
        icon: "ri:admin-line",
        title: "menus.pureAdminManagement",
        roles: ["admin"]
      }
    }
  ]
};

export default defineFakeRoute([
  {
    url: "/get-async-routes",
    method: "get",
    response: () => {
      return {
        code: 0,
        message: "操作成功",
        data: [
          accountManagementRouter
        ]
      };
    }
  }
]);
