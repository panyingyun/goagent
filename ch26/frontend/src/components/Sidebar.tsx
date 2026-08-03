import { useMemo, useState } from 'react'
import { useStore } from '../store'
import type { Session } from '../types'

function groupLabel(updatedAt: string): string {
  const d = new Date(updatedAt)
  const now = new Date()
  const day = (t: Date) => Math.floor(t.getTime() / 86400000)
  const diff = day(now) - day(d)
  if (diff === 0) return '今天'
  if (diff === 1) return '昨天'
  if (diff <= 7) return '最近 7 天'
  return '更早'
}

export function Sidebar() {
  const sessions = useStore((s) => s.sessions)
  const current = useStore((s) => s.current)
  const createSession = useStore((s) => s.createSession)
  const switchSession = useStore((s) => s.switchSession)
  const toggleFavorite = useStore((s) => s.toggleFavorite)
  const deleteSession = useStore((s) => s.deleteSession)
  const renameSession = useStore((s) => s.renameSession)
  const [query, setQuery] = useState('')

  const groups = useMemo(() => {
    const filtered = sessions.filter((s) => s.title.includes(query) || query === '')
    const map = new Map<string, Session[]>()
    for (const s of filtered) {
      const key = s.isFavorite ? '收藏' : groupLabel(s.updatedAt)
      if (!map.has(key)) map.set(key, [])
      map.get(key)!.push(s)
    }
    return [...map.entries()]
  }, [sessions, query])

  return (
    <aside className="flex w-60 shrink-0 flex-col border-r border-outline bg-surface-container-low">
      <div className="p-3">
        <button
          onClick={() => void createSession()}
          className="w-full rounded-md bg-primary px-3 py-2 text-sm text-white hover:opacity-90"
        >
          + 新建绘图
        </button>
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="搜索会话..."
          className="mt-2 w-full rounded-md border border-outline bg-white px-2 py-1.5 text-sm"
        />
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto px-2 pb-3">
        {groups.map(([label, items]) => (
          <div key={label} className="mb-3">
            <div className="px-2 py-1 text-xs text-on-surface-variant">{label}</div>
            {items.map((s) => (
              <div
                key={s.id}
                onClick={() => void switchSession(s.id)}
                onDoubleClick={() => {
                  const title = prompt('重命名会话', s.title)
                  if (title) void renameSession(s.id, title)
                }}
                className={`group mb-1 flex cursor-pointer items-center justify-between rounded-md px-2 py-1.5 text-sm hover:bg-surface-container-high ${
                  current?.id === s.id ? 'bg-primary-container/60' : ''
                }`}
              >
                <span className="truncate">{s.title}</span>
                <span className="flex gap-1 opacity-0 group-hover:opacity-100">
                  <button onClick={(e) => { e.stopPropagation(); void toggleFavorite(s.id) }}>
                    {s.isFavorite ? '⭐' : '☆'}
                  </button>
                  <button onClick={(e) => { e.stopPropagation(); if (confirm('删除该会话?')) void deleteSession(s.id) }}>
                    🗑
                  </button>
                </span>
              </div>
            ))}
          </div>
        ))}
        {sessions.length === 0 && <div className="p-3 text-sm text-on-surface-variant">还没有会话</div>}
      </div>
    </aside>
  )
}
