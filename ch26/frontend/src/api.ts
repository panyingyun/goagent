import type { Session, ChatResponse, GraphState } from './types'

const base = '/api'

async function j<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `HTTP ${res.status}`)
  }
  return res.json() as Promise<T>
}

const jsonHeaders = { 'Content-Type': 'application/json' }

export const api = {
  listSessions: () => fetch(`${base}/sessions`).then((r) => j<Session[]>(r)),
  createSession: (title: string) =>
    fetch(`${base}/sessions`, { method: 'POST', headers: jsonHeaders, body: JSON.stringify({ title }) }).then((r) => j<Session>(r)),
  getSession: (id: string) => fetch(`${base}/sessions/${id}`).then((r) => j<Session>(r)),
  patchSession: (id: string, patch: { title?: string; isFavorite?: boolean; graphState?: GraphState }) =>
    fetch(`${base}/sessions/${id}`, { method: 'PATCH', headers: jsonHeaders, body: JSON.stringify(patch) }).then((r) => j<Session>(r)),
  deleteSession: (id: string) => fetch(`${base}/sessions/${id}`, { method: 'DELETE' }),
  chat: (sessionId: string, message: string) =>
    fetch(`${base}/chat`, {
      method: 'POST',
      headers: jsonHeaders,
      body: JSON.stringify({ sessionId, message }),
    }).then((r) => j<ChatResponse>(r)),
}
