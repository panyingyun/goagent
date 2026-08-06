import { create } from 'zustand'
import {
  createSession as apiCreate,
  deleteSession as apiDelete,
  getSession,
  getToken,
  getTrace,
  listSessions,
  sendChat,
  setToken as persistToken,
} from './api'
import type { Session, SessionSummary, TraceRecord } from './types'

type Store = {
  sessions: SessionSummary[]
  session: Session | null
  busy: boolean
  error: string
  token: string
  query: string
  trace: TraceRecord | null
  hiddenIds: Set<string>
  setQuery: (q: string) => void
  setToken: (t: string) => void
  clearError: () => void
  toggleHidden: (id: string) => void
  refresh: (preferredId?: string) => Promise<void>
  addSession: () => Promise<void>
  openSession: (id: string) => Promise<void>
  removeSession: (id: string) => Promise<void>
  chat: (message: string) => Promise<void>
  openTrace: (id: string) => Promise<void>
}

export const useStore = create<Store>((set, get) => ({
  sessions: [],
  session: null,
  busy: false,
  error: '',
  token: getToken(),
  query: '',
  trace: null,
  hiddenIds: new Set(),

  setQuery: (q) => set({ query: q }),
  setToken: (t) => {
    persistToken(t)
    set({ token: t })
  },
  clearError: () => set({ error: '' }),
  toggleHidden: (id) => {
    const next = new Set(get().hiddenIds)
    if (next.has(id)) next.delete(id)
    else next.add(id)
    set({ hiddenIds: next })
  },

  refresh: async (preferredId) => {
    try {
      const items = await listSessions()
      set({ sessions: items })
      const current = get().session
      const target = preferredId ?? current?.id ?? items[0]?.id
      if (target) {
        const detail = await getSession(target)
        set({ session: detail, hiddenIds: new Set(), trace: null })
      } else {
        const created = await apiCreate()
        set({ session: created, sessions: await listSessions(), hiddenIds: new Set(), trace: null })
      }
    } catch (cause) {
      set({ error: cause instanceof Error ? cause.message : '加载失败' })
    }
  },

  addSession: async () => {
    try {
      const created = await apiCreate()
      set({ session: created, sessions: await listSessions(), hiddenIds: new Set(), trace: null, error: '' })
    } catch (cause) {
      set({ error: cause instanceof Error ? cause.message : '创建失败' })
    }
  },

  openSession: async (id) => {
    try {
      const detail = await getSession(id)
      set({ session: detail, hiddenIds: new Set(), trace: null, error: '' })
    } catch (cause) {
      set({ error: cause instanceof Error ? cause.message : '打开失败' })
    }
  },

  removeSession: async (id) => {
    try {
      await apiDelete(id)
      const items = await listSessions()
      set({ sessions: items })
      if (get().session?.id === id) {
        if (items[0]) {
          set({ session: await getSession(items[0].id), hiddenIds: new Set(), trace: null })
        } else {
          const created = await apiCreate()
          set({ session: created, sessions: await listSessions(), hiddenIds: new Set(), trace: null })
        }
      }
    } catch (cause) {
      set({ error: cause instanceof Error ? cause.message : '删除失败' })
    }
  },

  chat: async (message) => {
    const session = get().session
    if (!session || get().busy) return
    set({ busy: true, error: '', trace: null })
    try {
      const result = await sendChat(session.id, message)
      set({ session: result.session, sessions: await listSessions(), hiddenIds: new Set() })
    } catch (cause) {
      set({ error: cause instanceof Error ? cause.message : '请求失败' })
    } finally {
      set({ busy: false })
    }
  },

  openTrace: async (id) => {
    try {
      set({ trace: await getTrace(id) })
    } catch (cause) {
      set({ error: cause instanceof Error ? cause.message : 'Trace 已过期' })
    }
  },
}))
