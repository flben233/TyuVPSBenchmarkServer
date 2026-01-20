import AutoImport from "unplugin-auto-import/vite";
import Components from "unplugin-vue-components/vite";
import { ElementPlusResolver } from "unplugin-vue-components/resolvers";

console.log("Backend URL:", process.env.NUXT_BACKEND_URL);

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  app: {
    head: {
      title: 'Lolicon VPS - 云服务器评测', // 网站标题
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' }
      ],
      link: [
        { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }
      ]
    }
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
  nitro: {
    routeRules: {
      "/api/**": {
        proxy: `${process.env.NUXT_BACKEND_URL}/**`,
      },
    },
  },
  modules: ["@element-plus/nuxt", "@nuxtjs/sitemap", "@nuxtjs/robots"],
  // 手动全局引入 Element Plus 的 CSS
  css: ["element-plus/dist/index.css", "@/assets/main.css"],
  sitemap: {
    sources: ["/__sitemap__/urls"],
  },
  robots: {
    sitemap: ["/sitemap.xml"],
  }
});
