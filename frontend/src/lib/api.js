const defaultApiBase =
  typeof window !== "undefined"
    ? `${window.location.protocol}//${window.location.hostname}:9090/api`
    : "http://192.168.248.133:9090/api";

const envApiBase = import.meta.env.VITE_API_BASE_URL;

// Guard against stale builds where VITE_API_BASE_URL was set to localhost.
const normalizedEnvApiBase =
  typeof window !== "undefined" && typeof envApiBase === "string" && envApiBase.includes("localhost")
    ? envApiBase.replace("localhost", window.location.hostname)
    : envApiBase;

const API_BASE = normalizedEnvApiBase || defaultApiBase;

export async function api(path, { method = "GET", token, body } = {}) {
  const headers = { "Content-Type": "application/json" };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined
  });

  let data = {};
  try {
    data = await res.json();
  } catch (_e) {
    data = {};
  }

  if (!res.ok) {
    throw new Error(data.error || `Request failed (${res.status})`);
  }

  return data;
}

export async function apiWithMeta(path, { method = "GET", token, body } = {}) {
  const headers = { "Content-Type": "application/json" };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined
  });

  let data = {};
  try {
    data = await res.json();
  } catch (_e) {
    data = {};
  }

  if (!res.ok) {
    throw new Error(data.error || `Request failed (${res.status})`);
  }

  return { data, headers: res.headers };
}
