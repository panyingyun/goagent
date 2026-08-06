import { useStore } from '../store'

export function TopBar() {
  const session = useStore((s) => s.session)

  return (
    <header className="z-50 flex h-16 shrink-0 items-center justify-between border-b border-outline-variant bg-surface px-4">
      <div className="flex min-w-0 items-center gap-3">
        <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary text-on-primary shadow-sm">
          <span className="material-symbols-outlined text-[22px]" style={{ fontVariationSettings: "'FILL' 1" }}>
            insights
          </span>
        </div>
        <div className="font-sans text-xl font-bold tracking-tight text-primary">MathGraph AI</div>
      </div>

      <div className="hidden min-w-0 flex-1 px-8 text-center md:block">
        <div className="truncate text-sm font-medium text-on-surface">{session?.title ?? '新的数学探索'}</div>
      </div>

      <div className="flex items-center gap-1">
        <IconButton name="share" label="分享" />
        <IconButton name="download" label="导出" />
        <IconButton name="settings" label="设置" />
        <div className="ml-2 flex h-8 w-8 items-center justify-center overflow-hidden rounded-full border border-outline-variant bg-surface-container text-xs font-semibold text-on-surface-variant">
          MG
        </div>
      </div>
    </header>
  )
}

function IconButton({ name, label }: { name: string; label: string }) {
  return (
    <button
      type="button"
      title={label}
      className="rounded-full p-2 text-on-surface-variant transition-colors hover:bg-surface-container hover:text-on-surface"
    >
      <span className="material-symbols-outlined">{name}</span>
    </button>
  )
}
