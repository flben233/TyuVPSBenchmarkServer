export function useAdmin() {
  const { backendUrl } = useAppConfig();

  async function listUsers(token) {
    try {
      const resp = await $fetch(`${backendUrl}/auth/admin/users`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (resp && resp.code === 0) {
        return resp.data || [];
      }
      return [];
    } catch (error) {
      console.error("Failed to list users:", error);
      return [];
    }
  }

  async function updateUser(token, user) {
    try {
      const resp = await $fetch(`${backendUrl}/auth/admin/user/update`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(user),
      });

      if (resp && resp.code === 0) {
        return { success: true };
      }
      return { success: false, message: resp?.message || "Failed to update user" };
    } catch (error) {
      console.error("Failed to update user:", error);
      return { success: false, message: error.message };
    }
  }

  async function deleteUser(token, id) {
    try {
      const resp = await $fetch(`${backendUrl}/auth/admin/user/delete`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ id }),
      });

      if (resp && resp.code === 0) {
        return { success: true };
      }
      return { success: false, message: resp?.message || "Failed to delete user" };
    } catch (error) {
      console.error("Failed to delete user:", error);
      return { success: false, message: error.message };
    }
  }

  async function listUserGroups(token) {
    try {
      const resp = await $fetch(`${backendUrl}/auth/admin/groups`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (resp && resp.code === 0) {
        return resp.data || [];
      }
      return [];
    } catch (error) {
      console.error("Failed to list user groups:", error);
      return [];
    }
  }

  async function createUserGroup(token, group) {
    try {
      const resp = await $fetch(`${backendUrl}/auth/admin/group/create`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(group),
      });

      if (resp && resp.code === 0) {
        return { success: true };
      }
      return { success: false, message: resp?.message || "Failed to create group" };
    } catch (error) {
      console.error("Failed to create user group:", error);
      return { success: false, message: error.message };
    }
  }

  async function updateUserGroup(token, group) {
    try {
      const resp = await $fetch(`${backendUrl}/auth/admin/group/update`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(group),
      });

      if (resp && resp.code === 0) {
        return { success: true };
      }
      return { success: false, message: resp?.message || "Failed to update group" };
    } catch (error) {
      console.error("Failed to update user group:", error);
      return { success: false, message: error.message };
    }
  }

  async function deleteUserGroup(token, id) {
    try {
      const resp = await $fetch(`${backendUrl}/auth/admin/group/delete`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ id }),
      });

      if (resp && resp.code === 0) {
        return { success: true };
      }
      return { success: false, message: resp?.message || "Failed to delete group" };
    } catch (error) {
      console.error("Failed to delete user group:", error);
      return { success: false, message: error.message };
    }
  }

  return {
    listUsers,
    updateUser,
    deleteUser,
    listUserGroups,
    createUserGroup,
    updateUserGroup,
    deleteUserGroup,
  };
}
