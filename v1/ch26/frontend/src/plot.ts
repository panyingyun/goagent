// 安全采样:用 mathjs 编译表达式(不用 eval),只允许变量 x。
import { compile } from 'mathjs'
import type { EquationItem, GraphSettings, Viewport } from './types'

export function sampleEquation(
  expr: string,
  xMin: number,
  xMax: number,
  count = 400,
): { x: number[]; y: Array<number | null> } {
  let fn: ((x: number) => number) | null = null
  try {
    const node = compile(expr)
    fn = (x: number) => node.evaluate({ x }) as number
  } catch {
    return { x: [], y: [] }
  }
  const x: number[] = []
  const y: Array<number | null> = []
  const step = (xMax - xMin) / count
  for (let i = 0; i <= count; i++) {
    const xi = xMin + i * step
    let yi: number | null
    try {
      const v = fn(xi)
      yi = Number.isFinite(v) ? v : null
    } catch {
      yi = null
    }
    x.push(xi)
    y.push(yi)
  }
  return { x, y }
}

export function toTrace(eq: EquationItem, viewport: Viewport, settings: GraphSettings) {
  const { x, y } = sampleEquation(eq.normalizedExpression, viewport.xMin, viewport.xMax, settings.sampleCount)
  return {
    x,
    y,
    type: 'scatter',
    mode: 'lines',
    name: eq.label,
    line: { color: eq.color, width: eq.lineWidth },
  }
}
