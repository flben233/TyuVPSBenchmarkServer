import { defineEventHandler, proxyRequest } from "h3";

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig();
  const path = event.path.replace(/^\/api/, "");
  return proxyRequest(event, `${config.backendUrl}${path}`);
});
