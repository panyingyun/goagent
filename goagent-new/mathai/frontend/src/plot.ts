import { compile } from 'mathjs'

/**
 * 把白名单校验过的表达式字符串编译成 Mafs `Plot.OfX` 需要的 `(x) => number`。
 *
 * 后端 `domain.go` 的 `NormalizeExpression` 是安全边界；这里只是渲染层的求值，
 * 所以对编译失败、复数、NaN、Infinity 一律返回 NaN，让 Mafs 当作断点处理，
 * 与之前的 `sampleEquation` 行为一致，且不使用 `eval`。
 */
export function compileEquation(expression: string): (x: number) => number {
  let evaluator: ((x: number) => number) | null = null
  try {
    const node = compile(expression)
    evaluator = (x: number) => {
      const value = node.evaluate({ x }) as number
      return Number.isFinite(value) ? value : NaN
    }
  } catch {
    evaluator = () => NaN
  }
  return (x: number): number => {
    try {
      return evaluator!(x)
    } catch {
      return NaN
    }
  }
}
