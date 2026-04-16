const LLM_SETTINGS_KEY = "webssh_llm_settings";

export function getLLMSettings() {
  if (!process.client) {
    return {
      enabled: false,
      apiBase: "",
      apiKey: "",
      model: "",
    };
  }
  const raw = localStorage.getItem(LLM_SETTINGS_KEY);
  if (!raw) {
    return {
      enabled: false,
      apiBase: "",
      apiKey: "",
      model: "",
    };
  }
  try {
    const parsed = JSON.parse(raw);
    return {
      enabled: !!parsed.enabled,
      apiBase: parsed.apiBase || "",
      apiKey: parsed.apiKey || "",
      model: parsed.model || "",
    };
  } catch {
    return {
      enabled: false,
      apiBase: "",
      apiKey: "",
      model: "",
    };
  }
}

export function saveLLMSettings(settings) {
  if (!process.client) return;
  const normalized = {
    enabled: !!settings?.enabled,
    apiBase: settings?.apiBase || "",
    apiKey: settings?.apiKey || "",
    model: settings?.model || "",
  };
  localStorage.setItem(LLM_SETTINGS_KEY, JSON.stringify(normalized));
}
