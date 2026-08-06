type Props = {
  size?: number
  withWordmark?: boolean
  className?: string
}

/** 品牌标识：使用 docs/logo 导出图 */
export function BrandLogo({ size = 36, withWordmark = true, className = '' }: Props) {
  return (
    <div className={`flex items-center gap-2.5 ${className}`}>
      <img
        src="/logo.png"
        alt="MathGraph AI"
        width={size}
        height={size}
        className="shrink-0 object-contain"
        style={{ width: size, height: size }}
      />
      {withWordmark && (
        <span className="font-sans text-xl font-bold tracking-tight text-primary">MathGraph AI</span>
      )}
    </div>
  )
}
