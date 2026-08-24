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
  redirect: "/cert/index",
  meta: {
    icon: "ri:apps-line",
    title: "menus.pureApp",
    rank: 1
  },
  children: [
    {
      path: "/app/index",
      name: "AppMain",
      redirect: "/cert/index",
      meta: {
        title: "menus.pureApp",
        roles: ["admin", "common"]
      }
    }
  ]
};

const certRouter = {
  path: "/cert",
  meta: {
    icon: "ri:lock-line",
    title: "menus.pureCert",
    rank: 2
  },
  children: [
    {
      path: "/cert/index",
      name: "AppCert",
      component: "app/cert/index",
      meta: {
        icon: "ri:lock-line",
        title: "menus.pureCert",
        roles: ["admin", "common"]
      }
    }
  ]
};

const tunnelRouter = {
  path: "/tunnel",
  name: "AppTunnelParent",
  meta: {
    icon: "ri:route-line",
    title: "menus.pureTunnel",
    rank: 3,
    roles: ["admin", "common"]
  },
  children: [
    {
      path: "/tunnel/server",
      name: "AppTunnel",
      component: "app/tunnel/index",
      meta: {
        title: "menus.pureTunnelServer",
        roles: ["admin", "common"]
      }
    },
    {
      path: "/tunnel/client",
      name: "AppTunnelClient",
      component: "app/tunnel-client/index",
      meta: {
        title: "menus.pureTunnelClient",
        roles: ["admin", "common"]
      }
    }
  ]
};

const ruleRouter = {
  path: "/rule",
  meta: {
    icon: "ri:filter-3-line",
    title: "menus.pureRule",
    rank: 4
  },
  children: [
    {
      path: "/rule/index",
      name: "AppRule",
      component: "app/rule/index",
      meta: {
        icon: "ri:filter-3-line",
        title: "menus.pureRule",
        roles: ["admin", "common"]
      }
    }
  ]
};

const dnsRouter = {
  path: "/dns",
  meta: {
    icon: "ri:global-line",
    title: "menus.pureDns",
    rank: 5
  },
  children: [
    {
      path: "/dns/index",
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

const httpRouter = {
  path: "/http",
  meta: {
    icon: "ri:links-line",
    title: "menus.pureHttpProxy",
    rank: 6
  },
  children: [
    {
      path: "/http/index",
      name: "AppHttpProxy",
      component: "app/http/index",
      meta: {
        icon: "ri:links-line",
        title: "menus.pureHttpProxy",
        roles: ["admin", "common"]
      }
    }
  ]
};

const clusterRouter = {
  path: "/cluster",
  meta: {
    icon: "ri:server-line",
    title: "menus.pureCluster",
    rank: 7
  },
  children: [
    {
      path: "/cluster/index",
      name: "AppCluster",
      component: "app/cluster/index",
      meta: {
        icon: "ri:server-line",
        title: "menus.pureCluster",
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
          certRouter,
          tunnelRouter,
          ruleRouter,
          dnsRouter,
          httpRouter,
          clusterRouter,
          accountManagementRouter
        ]
      };
    }
  }
]);
