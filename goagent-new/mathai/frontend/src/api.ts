import type { Session, SessionSummary, TraceRecord } from "./types";

const tokenKey = "mathai-api-token";

export function getToken() { return localStorage.getItem(tokenKey) ?? ""; }
export function setToken(value: string) { localStorage.setItem(tokenKey, value.trim()); }

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = getToken();
  const response = await fetch(path, {
    ...init,
    headers: {
      ...(init.body ? { "Content-Type": "application/json" } : {}),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init.headers
    }
  });
  if (!response.ok) {
    const payload = await response.json().catch(() => ({ error: `HTTP ${response.status}` }));
    throw new Error(payload.error ?? `HTTP ${response.status}`);
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

export async function listSessions() {
  const data = await request<{ sessions: SessionSummary[] }>("/api/sessions");
  return data.sessions;
}

export function createSession(title = "新的数学探索") {
  return request<Session>("/api/sessions", { method: "POST", body: JSON.stringify({ title }) });
}

export function getSession(id: string) { return request<Session>(`/api/sessions/${encodeURIComponent(id)}`); }
export function deleteSession(id: string) { return request<void>(`/api/sessions/${encodeURIComponent(id)}`, { method: "DELETE" }); }

export function sendChat(id: string, message: string) {
  return request<{ session: Session; answer: string; traceId: string }>(`/api/sessions/${encodeURIComponent(id)}/chat`, {
    method: "POST", body: JSON.stringify({ message })
  });
}

export function getTrace(id: string) { return request<TraceRecord>(`/api/traces/${encodeURIComponent(id)}`); }
