import { useStore } from '../store'

export function TopBar() {
  const current = useStore((s) => s.current)
  const createSession = useStore((s) => s.createSession)

  return (
    <header className="flex h-14 shrink-0 items-center justify-between border-b border-outline bg-surface-container px-4">
      <div className="flex items-center gap-2">
        <span className="text-2xl">📈</span>
        <span className="font-bold">MathGraph AI</span>
      </div>
      <div className="flex items-center gap-4">
        <span className="text-sm text-on-surface-variant">{current?.title ?? '未选择会话'}</span>
        <button
          onClick={() => void createSession()}
          className="rounded-md bg-primary px-3 py-1.5 text-sm text-white hover:opacity-90"
        >
          + 新建绘图
        </button>
      </div>
    </header>
  )
}
