const Layout = () => import("@/layout/index.vue");

export default {
  path: "/application",
  name: "Application",
  component: Layout,
  redirect: "/application/http",
  meta: {
    icon: "ep/menu",
    title: "应用",
    rank: 10
  },
  children: [
    {
      path: "/application/http",
      name: "ApplicationHttp",
      component: () => import("@/views/application/http/index.vue"),
      meta: {
        title: "HTTP 代理",
        showParent: true
      }
    }
  ]
} satisfies RouteConfigsTable;
