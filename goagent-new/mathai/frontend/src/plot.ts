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
