import {requestWithAuth} from "~/composables/useAuth.js";

export function useAdmin() {
  async function listUsers(token) {
    try {
      const resp = await requestWithAuth(`/auth/admin/users`, "GET");
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
      const resp = await requestWithAuth(`/auth/admin/user/update`, "POST",{
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
      const resp = await requestWithAuth(`/auth/admin/user/delete`, "POST", {
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
      const resp = await requestWithAuth(`/auth/admin/groups`, "GET");

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
      const resp = await requestWithAuth(`/auth/admin/group/create`, "POST", {
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
      const resp = await requestWithAuth(`/auth/admin/group/update`, "POST", {
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
      const resp = await requestWithAuth(`/auth/admin/group/delete`, "POST", {
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
