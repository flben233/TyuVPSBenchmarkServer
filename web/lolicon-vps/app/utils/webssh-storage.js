const STORAGE_KEY = "webssh_connections";

export function getConnections() {
  if (!process.client) return [];
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return [];
  try {
    return JSON.parse(raw);
  } catch {
    return [];
  }
}

export function saveConnections(connections) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(connections));
}

export function saveConnection(connection) {
  const connections = getConnections();
  const idx = connections.findIndex((c) => c.id === connection.id);
  if (idx >= 0) {
    connections[idx] = connection;
  } else {
    connections.push(connection);
  }
  saveConnections(connections);
}

export function deleteConnection(id) {
  const connections = getConnections().filter((c) => c.id !== id);
  saveConnections(connections);
}

export function generateId() {
  return crypto.randomUUID();
}

const PASSWORD_KEY = "webssh_password";

export function savePassword(password) {
  if (!process.client) return;
  if (password) {
    localStorage.setItem(PASSWORD_KEY, password);
  } else {
    localStorage.removeItem(PASSWORD_KEY);
  }
}

export function getPassword() {
  if (!process.client) return null;
  return localStorage.getItem(PASSWORD_KEY);
}

export function clearPassword() {
  if (!process.client) return;
  localStorage.removeItem(PASSWORD_KEY);
}

export function hasPassword() {
  return !!getPassword();
}
