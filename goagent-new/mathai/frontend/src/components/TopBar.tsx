import { BrandLogo } from './BrandLogo'
import { useStore } from '../store'

export function TopBar() {
  const session = useStore((s) => s.session)

  return (
    <header className="z-50 flex h-16 shrink-0 items-center justify-between border-b border-outline-variant bg-surface px-gutter">
      <BrandLogo size={40} />

      <div className="hidden min-w-0 flex-1 px-8 text-center md:block">
        <div className="truncate text-sm font-medium text-on-surface">
          {session?.title ?? '新的数学探索'}
        </div>
      </div>

      <div className="flex items-center gap-1 text-on-surface-variant">
        <IconButton name="share" label="分享" />
        <IconButton name="csv" label="导出 CSV" />
        <IconButton name="settings" label="设置" />
        <div className="ml-2 flex h-8 w-8 items-center justify-center overflow-hidden rounded-full border border-outline-variant bg-surface-container">
          <img src="/logo.png" alt="" className="h-full w-full object-cover" />
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
      className="rounded-full p-2 transition-colors hover:bg-surface-container hover:text-on-surface"
    >
      <span className="material-symbols-outlined">{name}</span>
    </button>
  )
}
