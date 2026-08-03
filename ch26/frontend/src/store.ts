import { create } from 'zustand'
import { api } from './api'
import type { Session, GraphState, Message } from './types'

interface AppState {
  sessions: Session[]
  current: Session | null
  loading: boolean
  sending: boolean
  error: string | null
  loadSessions: () => Promise<void>
  createSession: () => Promise<void>
  switchSession: (id: string) => Promise<void>
  renameSession: (id: string, title: string) => Promise<void>
  toggleFavorite: (id: string) => Promise<void>
  deleteSession: (id: string) => Promise<void>
  sendMessage: (content: string) => Promise<void>
  updateGraphState: (gs: GraphState) => Promise<void>
  clearError: () => void
}

export const useStore = create<AppState>((set, get) => ({
  sessions: [],
  current: null,
  loading: false,
  sending: false,
  error: null,

  loadSessions: async () => {
    set({ loading: true })
    try {
      const sessions = await api.listSessions()
      set({ sessions, loading: false })
    } catch {
      set({ loading: false })
    }
  },

  createSession: async () => {
    const s = await api.createSession('新建绘图')
    set((st) => ({ sessions: [s, ...st.sessions], current: s }))
  },

  switchSession: async (id) => {
    const s = await api.getSession(id)
    set({ current: s })
  },

  renameSession: async (id, title) => {
    const s = await api.patchSession(id, { title })
    set((st) => ({
      sessions: st.sessions.map((x) => (x.id === id ? s : x)),
      current: st.current?.id === id ? s : st.current,
    }))
  },

  toggleFavorite: async (id) => {
    const cur = get().sessions.find((x) => x.id === id)
    if (!cur) return
    const s = await api.patchSession(id, { isFavorite: !cur.isFavorite })
    set((st) => ({
      sessions: st.sessions.map((x) => (x.id === id ? s : x)),
      current: st.current?.id === id ? s : st.current,
    }))
  },

  deleteSession: async (id) => {
    await api.deleteSession(id)
    set((st) => ({
      sessions: st.sessions.filter((x) => x.id !== id),
      current: st.current?.id === id ? null : st.current,
    }))
  },

  sendMessage: async (content) => {
    const cur = get().current
    if (!cur || !content.trim()) return
    set({ sending: true, error: null })
    const optimistic: Message = {
      id: 'local_' + Date.now(),
      role: 'user',
      content,
      createdAt: new Date().toISOString(),
      status: 'success',
    }
    set({ current: { ...cur, messages: [...cur.messages, optimistic] } })
    try {
      const resp = await api.chat(cur.id, content)
      set((st) => ({
        current: st.current
          ? { ...st.current, messages: [...st.current.messages, resp.message], graphState: resp.graphState }
          : st.current,
        sending: false,
      }))
      void get().loadSessions()
    } catch (e) {
      set({ sending: false, error: e instanceof Error ? e.message : '网络异常' })
    }
  },

  updateGraphState: async (gs) => {
    const cur = get().current
    if (!cur) return
    const s = await api.patchSession(cur.id, { graphState: gs })
    set((st) => ({
      sessions: st.sessions.map((x) => (x.id === s.id ? s : x)),
      current: st.current?.id === s.id ? s : st.current,
    }))
  },

  clearError: () => set({ error: null }),
}))
