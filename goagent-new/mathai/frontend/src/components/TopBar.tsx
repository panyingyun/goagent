import type { ReactNode } from 'react'
import { CircleUserRound, Download, Settings, Share2 } from 'lucide-react'
import { BrandLogo } from './BrandLogo'
import { useStore } from '../store'

export function TopBar() {
  const session = useStore((s) => s.session)
  const title = session?.title?.trim() || '新的数学探索'

  return (
    <header className="relative z-50 flex h-16 shrink-0 items-center justify-between border-b border-outline bg-white px-[18px]">
      <div className="relative z-10 shrink-0">
        <BrandLogo size={36} />
      </div>

      <div className="pointer-events-none absolute inset-x-0 hidden px-48 text-center md:block">
        <div className="truncate text-[15px] font-medium text-ink">{title}</div>
      </div>

      <div className="relative z-10 flex items-center gap-0.5 text-muted">
        <IconButton label="分享">
          <Share2 className="h-[18px] w-[18px]" strokeWidth={1.75} />
        </IconButton>
        <IconButton label="导出">
          <Download className="h-[18px] w-[18px]" strokeWidth={1.75} />
        </IconButton>
        <IconButton label="设置">
          <Settings className="h-[18px] w-[18px]" strokeWidth={1.75} />
        </IconButton>
        <button
          type="button"
          title="账户"
          className="ml-1 rounded-full p-1.5 text-ink transition-colors hover:bg-surface-low"
        >
          <CircleUserRound className="h-7 w-7" strokeWidth={1.5} />
        </button>
      </div>
    </header>
  )
}

function IconButton({ label, children }: { label: string; children: ReactNode }) {
  return (
    <button
      type="button"
      title={label}
      className="rounded-full p-2 transition-colors hover:bg-surface-low hover:text-ink"
    >
      {children}
    </button>
  )
}
