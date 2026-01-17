let userInfo = ref(null);
let token = ref(null);
let isAdmin = ref(false);

export function useAuth() {
  const { backendUrl } = useAppConfig();

  if (process.client) {
    const storedToken = sessionStorage.getItem("auth_token");
    if (storedToken) {
      token.value = storedToken;
      fetchUserInfo();
      checkAdmin();
      console.log("Restored token from sessionStorage:", token.value);
    }
  }

  async function login(t) {
    token.value = t;
    if (token.value) {
      await fetchUserInfo();
      await checkAdmin();
      sessionStorage.setItem("auth_token", token.value);
    }
  }

  async function fetchUserInfo() {
    if (!token.value) {
      throw new Error("No token available");
    }
    let resp = await $fetch(`${backendUrl}/auth/user`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token.value}`,
      },
    });
    console.log(resp);
    
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
    sessionStorage.removeItem("auth_token");
    location.reload();
  }

  async function checkAdmin() {
    try {
      let resp = await $fetch(`${backendUrl}/auth/admin`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${token.value}`,
        },
      });
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
  };
}
