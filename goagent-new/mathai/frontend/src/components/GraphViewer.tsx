import { forwardRef, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react'
import { Coordinates, Mafs, Plot } from 'mafs'
import 'mafs/core.css'
import { toPng } from 'html-to-image'
import { Pause, Play, RotateCcw } from 'lucide-react'
import { compileEquation, hasAnimationParam, hasTrigFunction, integerAxisStep, piAxisStep, piLabel } from '../plot'
import type { Equation, GraphSettings, PiLabelMode, Viewport } from '../types'

export type GraphViewerHandle = {
  exportPNG: () => Promise<void>
  requestFullscreen: () => Promise<void>
}

type Props = {
  equations: Equation[]
  viewport: Viewport
  settings: GraphSettings
  piLabelMode: PiLabelMode
}

/** 动画参数 t 的循环周期（弧度），t 在 [0, 2π) 内循环，适合 sin/cos 这类周期函数。 */
const ANIMATION_PERIOD = 2 * Math.PI

/**
 * 导出 PNG 时，Mafs 把坐标轴/网格线的颜色写在 SVG attribute 的 `var(--mafs-*)` 里，
 * html-to-image 不会解析这些 var()，导致坐标轴/网格线消失。这里把变量回退值内联进去。
 */
const MAFS_VAR_FALLBACK: Record<string, string> = {
  '--mafs-bg': '#ffffff',
  '--mafs-fg': '#434655',
  '--mafs-origin-color': '#191b23',
  '--mafs-line-color': '#e1e2ed',
  '--grid-line-subdivision-color': '#eef0f6',
  '--mafs-axis-stroke-width': '1px',
}

function resolveMafsVar(node: Element, value: string): string {
  return value.replace(/var\((--[\w-]+)\)/g, (match, name: string) => {
    const computed = window.getComputedStyle(node).getPropertyValue(name).trim()
    if (computed && computed !== `var(${name})`) return computed
    return MAFS_VAR_FALLBACK[name] ?? match
  })
}

/**
 * 轻量动画时钟：播放时用 requestAnimationFrame 累加相位并循环。
 * 不使用 Mafs 的 useStopwatch，因为它没有真·暂停（stop 会重置到 0），
 * 且其 setTime 内部会乘以 1000。这里暂停时 phase 保持不变，恢复后继续。
 */
function useAnimationClock(playing: boolean) {
  const [phase, setPhase] = useState(0)
  useEffect(() => {
    if (!playing) return
    let raf = 0
    let last = performance.now()
    const tick = (now: number) => {
      const dt = (now - last) / 1000
      last = now
      setPhase((p) => (p + dt) % ANIMATION_PERIOD)
      raf = requestAnimationFrame(tick)
    }
    raf = requestAnimationFrame(tick)
    return () => cancelAnimationFrame(raf)
  }, [playing])
  const reset = () => setPhase(0)
  return { phase, reset }
}

export const GraphViewer = forwardRef<GraphViewerHandle, Props>(function GraphViewer(
  { equations, viewport, settings, piLabelMode },
  ref,
) {
  const wrapRef = useRef<HTMLDivElement>(null)
  const measureRef = useRef<HTMLDivElement>(null)
  const [height, setHeight] = useState(0)

  // Mafs 的 height 只接受固定数字，不响应 CSS 高度，因此用 ResizeObserver 测量容器。
  useEffect(() => {
    const el = measureRef.current
    if (!el) return
    const observer = new ResizeObserver((entries) => {
      const next = entries[0]?.contentRect.height ?? 0
      if (next > 0) setHeight(next)
    })
    observer.observe(el)
    return () => observer.disconnect()
  }, [])

  const animated = useMemo(() => equations.some((e) => hasAnimationParam(e.expression)), [equations])
  const [playing, setPlaying] = useState(true)
  const { phase, reset } = useAnimationClock(playing && animated)

  const compiled = useMemo(() => {
    const map = new Map<string, (x: number, t?: number) => number>()
    for (const equation of equations) {
      map.set(equation.id, compileEquation(equation.expression))
    }
    return map
  }, [equations])

  const visible = useMemo(() => equations.filter((e) => e.visible), [equations])

  const autoPi = useMemo(() => visible.some((e) => hasTrigFunction(e.expression)), [visible])
  const usePiLabels = piLabelMode === 'pi' || (piLabelMode === 'auto' && autoPi)

  const xStep = usePiLabels ? piAxisStep(viewport.xMax - viewport.xMin) : integerAxisStep(viewport.xMax - viewport.xMin)
  const yStep = integerAxisStep(viewport.yMax - viewport.yMin)

  useImperativeHandle(ref, () => ({
    exportPNG: async () => {
      const el = wrapRef.current
      if (!el) return
      const svg = el.querySelector('svg')
      const patches: Array<{ node: Element; name: string; value: string | null }> = []
      if (svg) {
        for (const node of Array.from(svg.querySelectorAll('*'))) {
          for (const name of ['stroke', 'stroke-width', 'fill']) {
            const value = node.getAttribute(name)
            if (value && value.includes('var(')) {
              patches.push({ node, name, value })
              node.setAttribute(name, resolveMafsVar(node, value))
            }
          }
          const style = node.getAttribute('style')
          if (style && style.includes('var(')) {
            patches.push({ node, name: 'style', value: style })
            node.setAttribute('style', resolveMafsVar(node, style))
          }
        }
      }
      try {
        const dataUrl = await toPng(el, {
          backgroundColor: '#ffffff',
          pixelRatio: 2,
          skipFonts: true,
        })
        const anchor = document.createElement('a')
        anchor.href = dataUrl
        anchor.download = 'math-graph.png'
        anchor.click()
      } finally {
        for (const { node, name, value } of patches) {
          if (value == null) node.removeAttribute(name)
          else node.setAttribute(name, value)
        }
      }
    },
    requestFullscreen: async () => {
      const el = wrapRef.current
      if (!el) return
      if (document.fullscreenElement) {
        await document.exitFullscreen()
      } else {
        await el.requestFullscreen()
      }
    },
  }))

  return (
    <div ref={wrapRef} className="relative h-full w-full bg-white">
      <div ref={measureRef} className="absolute inset-0">
        {height > 0 && (
          <Mafs
            height={height}
            viewBox={{ x: [viewport.xMin, viewport.xMax], y: [viewport.yMin, viewport.yMax], padding: 0 }}
            preserveAspectRatio="contain"
            pan={false}
            zoom={false}
          >
            <Coordinates.Cartesian
              xAxis={{
                axis: settings.showAxes,
                lines: settings.showGrid ? xStep : false,
                ...(usePiLabels ? { labels: piLabel } : {}),
              }}
              yAxis={{ axis: settings.showAxes, lines: settings.showGrid ? yStep : false }}
            />
            {visible.map((equation) => {
              const fn = compiled.get(equation.id)
              return (
                <Plot.OfX
                  key={equation.id}
                  y={(x) => (fn ? fn(x, phase) : NaN)}
                  color={equation.color}
                  weight={equation.lineWidth || 2}
                  opacity={1}
                />
              )
            })}
          </Mafs>
        )}
      </div>
      {settings.showLegend && visible.length > 0 && (
        <div className="pointer-events-none absolute left-3 top-3 z-10 flex flex-wrap gap-x-4 gap-y-1.5 rounded-lg border border-outline bg-white/85 px-3 py-2 shadow-soft backdrop-blur-sm">
          {visible.map((equation) => (
            <span key={equation.id} className="flex items-center gap-1.5 text-xs text-muted">
              <span className="h-0.5 w-4 rounded-full" style={{ background: equation.color }} />
              {equation.label || equation.expression}
            </span>
          ))}
        </div>
      )}
      {animated && (
        <div className="absolute bottom-3 left-3 z-10 flex items-center gap-1 rounded-lg border border-outline bg-white/85 px-1.5 py-1 shadow-soft backdrop-blur-sm">
          <button
            type="button"
            title={playing ? '暂停动画' : '播放动画'}
            onClick={() => setPlaying((p) => !p)}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted transition hover:bg-primary-soft hover:text-primary"
          >
            {playing ? <Pause className="h-4 w-4" strokeWidth={1.75} /> : <Play className="h-4 w-4" strokeWidth={1.75} />}
          </button>
          <button
            type="button"
            title="重置动画"
            onClick={reset}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted transition hover:bg-primary-soft hover:text-primary"
          >
            <RotateCcw className="h-4 w-4" strokeWidth={1.75} />
          </button>
        </div>
      )}
    </div>
  )
})
