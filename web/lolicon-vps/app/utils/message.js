import { ElMessage } from 'element-plus'

export function warn(message) {
  ElMessage({
    message,
    type: "warning",
    plain: true
  });
}

export function success(message) {
  ElMessage({
    message,
    type: "success",
    plain: true
  });
}

export function error(message) {
  ElMessage({
    message,
    type: "error",
    plain: true
  });
}