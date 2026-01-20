import { defineEventHandler, proxyRequest } from "h3";

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig();
  console.log("Proxying request to backend:", config.backendUrl);
  const path = event.path.replace(/^\/api/, "");
  return proxyRequest(event, `${config.backendUrl}${path}`);
});
