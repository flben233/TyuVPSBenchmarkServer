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

function saveConnections(connections) {
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
