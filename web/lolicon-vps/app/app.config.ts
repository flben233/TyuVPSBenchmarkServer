export default defineAppConfig({
  backendUrl: "/api",  // 这个定义app的后端路径前缀，设置为/app的话会走nuxt的server/api进行代理，runtimeConfig的backendUrl定义的是nuxt转发的目标地址
  // clientId: "Ov23lincx6LjHXADmQ94",
  clientId: "Ov23limxDDoGO9of9P4m",
  backendWsUrl: "/api",  // WebSocket直连后端URL，开发时设为空则自动推导，生产环境可设为 wss://your-domain.com/api
});
