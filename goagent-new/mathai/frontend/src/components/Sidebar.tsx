import { useMemo } from 'react'
import { useStore } from '../store'
import type { SessionSummary } from '../types'

function groupLabel(updatedAt: string) {
  const d = new Date(updatedAt)
  const now = new Date()
  const day = (t: Date) => Math.floor(Date.UTC(t.getFullYear(), t.getMonth(), t.getDate()) / 86400000)
  const diff = day(now) - day(d)
  if (diff === 0) return '今天'
  if (diff === 1) return '昨天'
  if (diff <= 7) return '最近 7 天'
  return '更早'
}

export function Sidebar() {
  const sessions = useStore((s) => s.sessions)
  const session = useStore((s) => s.session)
  const query = useStore((s) => s.query)
  const token = useStore((s) => s.token)
  const setQuery = useStore((s) => s.setQuery)
  const setToken = useStore((s) => s.setToken)
  const addSession = useStore((s) => s.addSession)
  const openSession = useStore((s) => s.openSession)
  const removeSession = useStore((s) => s.removeSession)
  const refresh = useStore((s) => s.refresh)

  const groups = useMemo(() => {
    const filtered = sessions.filter((s) => !query || s.title.toLowerCase().includes(query.toLowerCase()))
    const map = new Map<string, SessionSummary[]>()
    for (const item of filtered) {
      const key = groupLabel(item.updatedAt)
      if (!map.has(key)) map.set(key, [])
      map.get(key)!.push(item)
    }
    const order = ['今天', '昨天', '最近 7 天', '更早']
    return order.filter((k) => map.has(k)).map((k) => [k, map.get(k)!] as const)
  }, [sessions, query])

  return (
    <aside className="hidden w-sidebar shrink-0 flex-col border-r border-outline-variant bg-surface-container-low md:flex">
      <div className="p-4">
        <button
          type="button"
          onClick={() => void addSession()}
          className="mb-4 flex w-full items-center justify-center gap-2 rounded-xl bg-on-surface py-3 text-sm font-semibold text-white shadow-sm transition hover:bg-primary"
        >
          <span className="material-symbols-outlined text-[20px]" style={{ fontVariationSettings: "'FILL' 1" }}>
            add
          </span>
          新建绘图
        </button>
        <div className="relative">
          <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-[20px] text-on-surface-variant">
            search
          </span>
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="搜索会话或方程"
            className="w-full rounded-lg border border-outline-variant bg-surface-container-lowest py-2 pl-10 pr-3 text-sm outline-none transition focus:border-primary focus:ring-1 focus:ring-primary"
          />
        </div>
      </div>

      <div className="custom-scrollbar min-h-0 flex-1 overflow-y-auto px-3 pb-3">
        {groups.map(([label, items]) => (
          <div key={label} className="mb-5">
            <h3 className="mb-2 px-2 text-[12px] font-medium uppercase tracking-wider text-on-surface-variant">
              {label}
            </h3>
            <ul className="space-y-1">
              {items.map((item) => {
                const active = session?.id === item.id
                return (
                  <li key={item.id}>
                    <div
                      role="button"
                      tabIndex={0}
                      onClick={() => void openSession(item.id)}
                      onKeyDown={(e) => e.key === 'Enter' && void openSession(item.id)}
                      className={`group flex cursor-pointer items-center gap-2 rounded-xl px-3 py-2 transition ${
                        active
                          ? 'bg-white text-secondary shadow-sm ring-1 ring-secondary/20'
                          : 'text-on-surface-variant hover:bg-surface-container-highest hover:text-on-surface'
                      }`}
                    >
                      <span className="material-symbols-outlined text-[18px]">history</span>
                      <span className="min-w-0 flex-1 truncate text-sm font-medium">{item.title}</span>
                      <button
                        type="button"
                        title="删除"
                        className="rounded p-0.5 opacity-0 transition group-hover:opacity-100 hover:text-error"
                        onClick={(e) => {
                          e.stopPropagation()
                          if (confirm('删除该会话？')) void removeSession(item.id)
                        }}
                      >
                        <span className="material-symbols-outlined text-[16px]">delete</span>
                      </button>
                    </div>
                  </li>
                )
              })}
            </ul>
          </div>
        ))}
        {sessions.length === 0 && (
          <p className="px-3 py-6 text-center text-sm text-on-surface-variant">还没有会话</p>
        )}
      </div>

      <div className="space-y-2 border-t border-outline-variant p-4">
        <label className="block text-[11px] font-medium uppercase tracking-wider text-on-surface-variant">
          API Token（可选）
        </label>
        <input
          type="password"
          value={token}
          placeholder="与服务端配置一致"
          onChange={(e) => setToken(e.target.value)}
          onBlur={() => void refresh()}
          className="w-full rounded-lg border border-outline-variant bg-surface-container-lowest px-3 py-2 text-sm outline-none focus:border-primary"
        />
        <div className="flex items-center gap-2 pt-1 text-sm text-on-surface-variant">
          <span className="material-symbols-outlined text-[18px]">cloud_done</span>
          自动保存
        </div>
      </div>
    </aside>
  )
}
