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
      script: [
        { type: 'text/javascript', src: 'https://hm.baidu.com/hm.js?b812ec888c5b927ccd8fef3a34a5dc16' }
      ]
    },
  },
  vite: {
    plugins: [
      AutoImport({
        resolvers: [ElementPlusResolver({ importStyle: true })],
      }),
      Components({
        resolvers: [ElementPlusResolver({ importStyle: true })],
      }),
    ],
  },
  routeRules: {
    "/report/**": { cache: false },
    "/slide/**": { cache: false },
    "/inspector": { ssr: false },
    "/inspector/**": { ssr: false },
    "/tools/**": { prerender: true },
    "/center": { ssr: false },
    "/index": { cache: false },
    "/looking-glass": { cache: false },
    "/monitor": { ssr: false },
    "/search": { prerender: true },
    "/webssh": { ssr: false },
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
