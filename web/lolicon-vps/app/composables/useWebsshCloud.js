import {requestWithAuth} from "~/composables/useAuth.js";

export function useWebsshCloud() {
  async function uploadEncryptedData(encryptedData) {
    return await requestWithAuth("/webssh/sync/upload", "POST", {
      body: {encrypted_data: encryptedData},
    });
  }

  async function downloadEncryptedData() {
    return await requestWithAuth("/webssh/sync/download", "GET");
  }

  async function resetCloudData() {
    return await requestWithAuth("/webssh/sync/reset", "POST");
  }

  return {
    uploadEncryptedData,
    downloadEncryptedData,
    resetCloudData,
  };
}
