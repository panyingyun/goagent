import { useMemo, useState } from 'react'
import { CloudCheck, HelpCircle, History, MoreHorizontal, Plus, Search, Trash2 } from 'lucide-react'
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
  const [menuId, setMenuId] = useState<string | null>(null)

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
    <aside className="hidden w-sidebar shrink-0 flex-col border-r border-outline bg-surface-low md:flex">
      <div className="p-4">
        <button
          type="button"
          onClick={() => void addSession()}
          className="mb-4 flex w-full items-center justify-center gap-2 rounded-full bg-primary py-3 text-sm font-semibold text-on-primary shadow-soft transition hover:bg-primary-container"
        >
          <Plus className="h-4 w-4" strokeWidth={2.25} />
          新建绘图
        </button>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-faint" strokeWidth={1.75} />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="搜索会话或方程"
            className="w-full rounded-xl border border-outline bg-white py-2.5 pl-10 pr-3 text-sm text-ink outline-none transition placeholder:text-faint focus:border-primary focus:ring-1 focus:ring-primary/30"
          />
        </div>
      </div>

      <div className="custom-scrollbar min-h-0 flex-1 overflow-y-auto px-3 pb-3">
        {groups.map(([label, items]) => (
          <div key={label} className="mb-5">
            <h3 className="mb-2 px-2 text-[12px] font-medium tracking-wider text-faint">{label}</h3>
            <ul className="space-y-1">
              {items.map((item) => {
                const active = session?.id === item.id
                return (
                  <li key={item.id} className="relative">
                    <div
                      role="button"
                      tabIndex={0}
                      onClick={() => {
                        setMenuId(null)
                        void openSession(item.id)
                      }}
                      onKeyDown={(e) => e.key === 'Enter' && void openSession(item.id)}
                      className={`group flex cursor-pointer items-center gap-2 rounded-xl px-3 py-2.5 transition ${
                        active
                          ? 'bg-primary-soft text-ink shadow-soft ring-1 ring-primary/15'
                          : 'text-muted hover:bg-white hover:text-ink'
                      }`}
                    >
                      <History className="h-4 w-4 shrink-0 opacity-70" strokeWidth={1.75} />
                      <span className="min-w-0 flex-1 truncate text-sm font-medium">{item.title}</span>
                      <button
                        type="button"
                        title="更多"
                        className={`rounded p-0.5 transition hover:bg-white/80 ${
                          active || menuId === item.id ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'
                        }`}
                        onClick={(e) => {
                          e.stopPropagation()
                          setMenuId((id) => (id === item.id ? null : item.id))
                        }}
                      >
                        <MoreHorizontal className="h-4 w-4" strokeWidth={1.75} />
                      </button>
                    </div>
                    {menuId === item.id && (
                      <div className="absolute right-2 top-full z-20 mt-1 min-w-[120px] rounded-xl border border-outline bg-white py-1 shadow-soft">
                        <button
                          type="button"
                          className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-error hover:bg-error-container"
                          onClick={(e) => {
                            e.stopPropagation()
                            setMenuId(null)
                            if (confirm('删除该会话？')) void removeSession(item.id)
                          }}
                        >
                          <Trash2 className="h-3.5 w-3.5" strokeWidth={1.75} />
                          删除会话
                        </button>
                      </div>
                    )}
                  </li>
                )
              })}
            </ul>
          </div>
        ))}
        {sessions.length === 0 && (
          <p className="px-3 py-6 text-center text-sm text-faint">还没有会话</p>
        )}
      </div>

      <div className="space-y-1 border-t border-outline p-3">
        <label className="mb-1 block px-2 text-[11px] font-medium tracking-wider text-faint">
          API Token（可选）
        </label>
        <input
          type="password"
          value={token}
          placeholder="与服务端配置一致"
          onChange={(e) => setToken(e.target.value)}
          onBlur={() => void refresh()}
          className="mb-2 w-full rounded-xl border border-outline bg-white px-3 py-2 text-sm outline-none focus:border-primary"
        />
        <a className="flex items-center gap-3 rounded-xl px-3 py-2 text-sm text-muted hover:bg-white hover:text-ink">
          <CloudCheck className="h-4 w-4" strokeWidth={1.75} />
          自动保存
        </a>
        <a className="flex items-center gap-3 rounded-xl px-3 py-2 text-sm text-muted hover:bg-white hover:text-ink">
          <HelpCircle className="h-4 w-4" strokeWidth={1.75} />
          使用帮助
        </a>
      </div>
    </aside>
  )
}
