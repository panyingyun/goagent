type Props = {
  size?: number
  withWordmark?: boolean
  className?: string
}

/** 品牌标识：蓝色系主色 + 字标 */
export function BrandLogo({ size = 36, withWordmark = true, className = '' }: Props) {
  const r = Math.round(size * 0.22)
  return (
    <div className={`flex items-center gap-2.5 ${className}`}>
      <svg width={size} height={size} viewBox="0 0 40 40" className="shrink-0" aria-hidden>
        <rect width="40" height="40" rx={r} fill="#004ac6" />
        <text
          x="20"
          y="27"
          textAnchor="middle"
          fill="#fff"
          fontFamily="Inter, LXGWWenKai, sans-serif"
          fontSize="20"
          fontWeight="700"
        >
          M
        </text>
      </svg>
      {withWordmark && (
        <span className="font-sans text-xl font-bold tracking-tight text-primary">MathGraph AI</span>
      )}
    </div>
  )
}
