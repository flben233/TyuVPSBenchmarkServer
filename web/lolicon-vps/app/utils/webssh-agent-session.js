const STORAGE_KEY = "webssh-agent-sessions";
const MAX_SIZE = 2 * 1024 * 1024;

function loadAll() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? JSON.parse(raw) : {};
  } catch {
    return {};
  }
}

function persistAll(sessions) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(sessions));
}

function estimateSize(sessions) {
  return new Blob([JSON.stringify(sessions)]).size;
}

export function getSessions() {
  return loadAll();
}

export function getSession(id) {
  return loadAll()[id] || null;
}

export function saveSession(session) {
  const sessions = loadAll();
  sessions[session.id] = { ...session, updatedAt: Date.now() };

  const entries = Object.values(sessions).sort((a, b) => a.updatedAt - b.updatedAt);
  while (estimateSize(sessions) > MAX_SIZE && entries.length > 1) {
    const oldest = entries.shift();
    delete sessions[oldest.id];
  }

  persistAll(sessions);
}

export function deleteSession(id) {
  const sessions = loadAll();
  delete sessions[id];
  persistAll(sessions);
}

export function createSessionId() {
  return `sess_${Date.now()}_${Math.random().toString(36).slice(2, 9)}`;
}

export function getSortedSessions() {
  const sessions = loadAll();
  return Object.values(sessions).sort((a, b) => b.updatedAt - a.updatedAt);
}
