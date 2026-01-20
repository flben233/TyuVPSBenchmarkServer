import AutoImport from "unplugin-auto-import/vite";
import Components from "unplugin-vue-components/vite";
import { ElementPlusResolver } from "unplugin-vue-components/resolvers";

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
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
    devProxy: {
      "/api": {
        target: "http://127.0.0.1:12345/api",
      },
    },
    routeRules: {
      "/api/**": {
        proxy: "http://127.0.0.1:12345/api/**",
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
