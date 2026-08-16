import { compile } from 'mathjs'

/**
 * 把白名单校验过的表达式字符串编译成 Mafs `Plot.OfX` 需要的求值函数。
 *
 * `x` 是自变量（前端按视口采样），`t` 是动画参数（前端用动画时钟驱动，默认 0）。
 * 后端 `domain.go` 的 `NormalizeExpression` 是安全边界；这里只是渲染层求值，
 * 所以对编译失败、复数、NaN、Infinity 一律返回 NaN，让 Mafs 当作断点处理，
 * 与之前的 `sampleEquation` 行为一致，且不使用 `eval`。
 */
export function compileEquation(expression: string): (x: number, t?: number) => number {
  let evaluator: ((x: number, t: number) => number) | null = null
  try {
    const node = compile(expression)
    evaluator = (x: number, t: number) => {
      const value = node.evaluate({ x, t }) as number
      return Number.isFinite(value) ? value : NaN
    }
  } catch {
    evaluator = () => NaN
  }
  return (x: number, t = 0): number => {
    try {
      return evaluator!(x, t)
    } catch {
      return NaN
    }
  }
}

/** 表达式是否引用了动画参数 `t`（用于决定是否显示动画控制）。 */
export function hasAnimationParam(expression: string): boolean {
  return /\bt\b/.test(expression)
}

/** 表达式是否含三角函数（sin/cos/tan），用于「自动」模式检测是否用 π 坐标。 */
export function hasTrigFunction(expression: string): boolean {
  return /\b(sin|cos|tan)\b/.test(expression)
}

/**
 * 把 x 轴刻度值格式化成 π 的分数标签：
 * 0 → "0"，π/2 → "π/2"，π → "π"，3π/2 → "3π/2"，2π → "2π"，-π/2 → "-π/2"。
 * 用于切换「π 坐标」时替代 Mafs 默认的整数标签。
 */
export function piLabel(value: number): string {
  const ratio = value / Math.PI
  const [num, den] = approximateFraction(ratio, 12)
  if (num === 0) return '0'
  const sign = num < 0 ? '-' : ''
  const abs = Math.abs(num)
  const pi = abs === 1 && den === 1 ? 'π' : `${abs === 1 ? '' : abs}π`
  return den === 1 ? `${sign}${pi}` : `${sign}${pi}/${den}`
}

/** 求 value 最接近的分数 n/d（|d| ≤ maxDen），用于把 π 的倍数转成漂亮标签。 */
function approximateFraction(value: number, maxDen: number): [number, number] {
  let bestNum = 0
  let bestDen = 1
  let bestErr = Number.POSITIVE_INFINITY
  for (let den = 1; den <= maxDen; den++) {
    const num = Math.round(value * den)
    const err = Math.abs(value - num / den)
    if (err < bestErr) {
      bestErr = err
      bestNum = num
      bestDen = den
    }
  }
  return [bestNum, bestDen]
}

/** 计算"漂亮"的整数刻度间距（1、2、5 × 10^n），使约 targetTicks 个刻度，避免数字重叠。 */
export function integerAxisStep(range: number, targetTicks = 10): number {
  const raw = range / targetTicks
  const exp = Math.floor(Math.log10(raw))
  const base = Math.pow(10, exp)
  const frac = raw / base
  const mult = frac <= 1 ? 1 : frac <= 2 ? 2 : frac <= 5 ? 5 : 10
  return mult * base
}

/** 计算 π 刻度间距（π/2 的整数倍），使约 targetTicks 个刻度。 */
export function piAxisStep(range: number, targetTicks = 8): number {
  const rough = range / targetTicks / (Math.PI / 2)
  return Math.max(1, Math.round(rough)) * (Math.PI / 2)
}
