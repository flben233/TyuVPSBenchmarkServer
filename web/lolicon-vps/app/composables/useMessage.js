export function useMessage() {
  function warn(message) {
    ElMessage({
      message,
      type: "warning",
      plain: true
    });
  }

  function success(message) {
    ElMessage({
      message,
      type: "success",
      plain: true
    });
  }

  function err(message) {
    ElMessage({
      message,
      type: "error",
      plain: true
    });
  }

  return {
    warn,
    success,
    err
  };
}