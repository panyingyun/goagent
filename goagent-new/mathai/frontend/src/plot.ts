import { compile } from 'mathjs'
import type { Equation, GraphSettings, Viewport } from './types'

export function sampleEquation(expr: string, xMin: number, xMax: number, count = 400) {
  let fn: ((x: number) => number) | null = null
  try {
    const node = compile(expr)
    fn = (x: number) => node.evaluate({ x }) as number
  } catch {
    return { x: [] as number[], y: [] as Array<number | null> }
  }
  const x: number[] = []
  const y: Array<number | null> = []
  const step = (xMax - xMin) / Math.max(1, count)
  for (let i = 0; i <= count; i++) {
    const xi = xMin + i * step
    try {
      const v = fn(xi)
      x.push(xi)
      y.push(Number.isFinite(v) ? v : null)
    } catch {
      x.push(xi)
      y.push(null)
    }
  }
  return { x, y }
}

export function toTrace(eq: Equation, viewport: Viewport, settings: GraphSettings) {
  const { x, y } = sampleEquation(eq.expression, viewport.xMin, viewport.xMax, settings.sampleCount)
  return {
    x,
    y,
    type: 'scatter' as const,
    mode: 'lines' as const,
    name: eq.label || eq.expression,
    line: { color: eq.color, width: eq.lineWidth || 2 },
  }
}
