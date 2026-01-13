// TODO: 保存登录状态
export function useAuth() {
  const { backendUrl } = useAppConfig();
  let userInfo = ref(null);
  let token = ref(null);

  async function login(token) {
    token.value = token;
    if (token.value) {
      await fetchUserInfo();
    }
  }

  async function fetchUserInfo() {
    if (!token.value) {
      throw new Error("No token available");
    }
    let resp = await useFetch(`${backendUrl}/auth/user`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token.value}`,
      },
    });
    if (resp.data.value && resp.data.value.code === 0) {
      userInfo.value = {
        name: resp.data.value.data.name,
        avatarUrl: resp.data.value.data.avatar_url,
      };
    }
  }
  return {
    userInfo,
    token,
    login,
    fetchUserInfo,
  };
}
