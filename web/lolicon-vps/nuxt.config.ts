import AutoImport from "unplugin-auto-import/vite";
import Components from "unplugin-vue-components/vite";
import { ElementPlusResolver } from "unplugin-vue-components/resolvers";

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  runtimeConfig: {
    backendUrl: "",
  },
  app: {
    head: {
      title: "Lolicon VPS - 云服务器评测", // 网站标题
      meta: [
        { charset: "utf-8" },
        { name: "viewport", content: "width=device-width, initial-scale=1" },
        {
          name: "description",
          content:
            "VPS云服务器性能评测平台，提供VPS测速、服务器监控、IP查询、路由追踪、WHOIS查询等工具，支持三网回程线路测试与流媒体解锁检测",
        },
        {
          name: "keywords",
          content:
            "VPS测评,云服务器评测,VPS测速,服务器性能测试,VPS benchmark,主机监控,IP查询,路由追踪,traceroute,WHOIS查询,Looking Glass,回程线路,三网测速,流媒体解锁,Netflix解锁,中国电信,中国移动,中国联通,KVM,虚拟化,磁盘IO测试,网络延迟,带宽测试",
        },
      ],
      link: [{ rel: "icon", type: "image/x-icon", href: "/favicon.ico" }],
    },
  },
  vite: {
    plugins: [
      AutoImport({
        resolvers: [ElementPlusResolver({ importStyle: false })],
      }),
      Components({
        resolvers: [ElementPlusResolver({ importStyle: false })],
      }),
    ],
  },
  routeRules: {
    "/report/**": { swr: 31536000 },
    "/slide/**": { swr: 31536000 },
    "/tools/**": { prerender: true },
    "/center": { prerender: true },
    "/index": { cache: { maxAge: 300 } },
    "/looking-glass": { cache: { maxAge: 300 } },
    "/monitor": { prerender: true },
    "/search": { prerender: true },
  },
  modules: ["@element-plus/nuxt", "@nuxtjs/sitemap", "@nuxtjs/robots"],
  // 手动全局引入 Element Plus 的 CSS
  css: ["element-plus/dist/index.css", "@/assets/main.css"],
  sitemap: {
    sources: ["/__sitemap__/urls"],
  },
  robots: {
    sitemap: ["/sitemap.xml"],
  },
});
