// 模拟后端动态生成路由
import { defineFakeRoute } from "vite-plugin-fake-server/client";
import { system, monitor, permission, frame, tabs } from "@/router/enums";

/**
 * roles：页面级别权限，这里模拟二种 "admin"、"common"
 * admin：管理员角色
 * common：普通角色
 */

const appRouter = {
  path: "/app",
  meta: {
    icon: "ri:apps-line",
    title: "menus.pureApp",
    rank: 5
  },
  children: [
    {
      path: "/app/cert",
      name: "AppCert",
      component: "app/cert/index",
      meta: {
        icon: "ri:shield-keyhole-line",
        title: "menus.pureCert",
        roles: ["admin", "common"]
      }
    },
    {
      path: "/app/tunnel",
      name: "AppTunnel",
      component: "app/tunnel/index",
      meta: {
        icon: "ri:route-line",
        title: "menus.pureTunnel",
        roles: ["admin", "common"]
      }
    },
    {
      path: "/app/dns",
      name: "AppDns",
      component: "app/dns/index",
      meta: {
        icon: "ri:global-line",
        title: "menus.pureDns",
        roles: ["admin", "common"]
      }
    }
  ]
};

const accountManagementRouter = {
  path: "/admin",
  meta: {
    icon: "ri:admin-line",
    title: "menus.pureAdminManagement",
    rank: system
  },
  children: [
    {
      path: "/admin/index",
      name: "SystemAdmin",
      component: "system/admin/index",
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
          appRouter,
          accountManagementRouter
        ]
      };
    }
  }
]);
