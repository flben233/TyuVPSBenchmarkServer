import {useMessage} from "~/composables/useMessage.js";

let userInfo = ref(null);
let token = ref(null);
let isAdmin = ref(false);
let refreshingToken = false;
const failQueue = [];

async function doRequest(url, method, options = {}) {
  const { backendUrl } = useAppConfig();

  if (!token.value) {
    let err =  new Error("No auth token found. Please log in.");
    err.statusCode = 401;
    throw err;
  }

  const headers = {
    Authorization: `Bearer ${token.value}`,
    "Content-Type": "application/json",
    ...options.headers,
  };
  return $fetch(`${backendUrl}${url}`, {
    method,
    headers,
    ...options,
  });
}

async function ensureTokenRefresh() {
  if (refreshingToken) {
    return new Promise((resolve, reject) => {
      failQueue.push({ resolve, reject });
    });
  }
  const { refreshToken } = useAuth();
  await refreshToken();
  console.log("Token refreshed, retrying request...");
}

export async function requestWithAuth(url, method, options = {}) {
  try {
    return await doRequest(url, method, options);
  } catch (error) {
    if (error.statusCode === 401) {
      await ensureTokenRefresh();
      return await doRequest(url, method, options);
    }
  }
}

export async function fetchWithAuth(url, options = {}) {
  const { backendUrl } = useAppConfig();

  if (!token.value) {
    const err = new Error("No auth token found. Please log in.");
    err.statusCode = 401;
    throw err;
  }

  const headers = {
    Authorization: `Bearer ${token.value}`,
    ...options.headers,
  };
  if (options.body && typeof options.body === "string") {
    headers["Content-Type"] = "application/json";
  }

  const resp = await fetch(`${backendUrl}${url}`, { ...options, headers });

  if (resp.status === 401) {
    await ensureTokenRefresh();
    const retryHeaders = {
      Authorization: `Bearer ${token.value}`,
      ...options.headers,
    };
    if (options.body && typeof options.body === "string") {
      retryHeaders["Content-Type"] = "application/json";
    }
    return fetch(`${backendUrl}${url}`, { ...options, headers: retryHeaders });
  }

  return resp;
}

export function useAuth() {
  const { backendUrl } = useAppConfig();

  if (process.client) {
    const storedToken = localStorage.getItem("auth_token");
    if (storedToken) {
      token.value = storedToken;
      fetchUserInfo();
      checkAdmin();
      console.log("Restored token from localStorage:", token.value);
    }
  }

  async function login(code) {
    let resp = await $fetch(`${backendUrl}/auth/github/login`, {
      method: "GET",
      query: { code },
    });
    if (resp && resp.code === 0) {
      token.value = resp.data.token;
      await fetchUserInfo();
      await checkAdmin();
      localStorage.setItem("auth_token", token.value);
    }
  }

  async function refreshToken() {
    refreshingToken = true;
    try {
      let resp = await $fetch(`${backendUrl}/auth/refresh`, { method: "POST" });
      if (resp && resp.code === 0) {
        token.value = resp.data.token;
        await fetchUserInfo();
        await checkAdmin();
        localStorage.setItem("auth_token", token.value);
      } else {
        await logout();
      }
    } catch (e) {
      if (e.statusCode === 400 && e.data.code === -7) {
        return
      }
      await logout();
      useMessage().err("会话已过期，请重新登录");
    } finally {
      refreshingToken = false;
      failQueue.forEach(p => p.resolve());
      failQueue.length = 0;
    }
  }

  async function fetchUserInfo() {
    let resp = await requestWithAuth(`/auth/user`, "GET");
    
    if (resp && resp.code === 0) {
      userInfo.value = {
        id: resp.data.id,
        name: resp.data.name,
        avatarUrl: resp.data.avatar_url,
      };
      console.log("User info fetched:", userInfo.value);
    }
  }

  async function logout() {
    token.value = null;
    userInfo.value = null;
    isAdmin.value = false;
    localStorage.removeItem("auth_token");
  }

  async function checkAdmin() {
    try {
      let resp = await requestWithAuth(`/auth/admin`, "GET");
      console.log("Admin check response:", resp);
      if (resp && resp.code === 0) {
        isAdmin.value = true;
      }
    } catch (error) {
      isAdmin.value = false;
    }
  }

  return {
    userInfo,
    token,
    login,
    fetchUserInfo,
    logout,
    isAdmin,
    refreshToken
  };
}
