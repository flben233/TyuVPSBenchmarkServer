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

export async function requestWithAuth(url, method, options = {}) {
  try {
    return await doRequest(url, method, options);
  } catch (error) {
    if (error.statusCode === 401) {
      if (refreshingToken) {
        return new Promise((resolve, reject) => {
          failQueue.push({ resolve, reject });
        }).then(() => doRequest(url, method, options));
      }
      const { refreshToken } = useAuth();
      await refreshToken();
      console.log("Token refreshed, retrying request...");
      return await doRequest(url, method, options);
    }
  }
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
