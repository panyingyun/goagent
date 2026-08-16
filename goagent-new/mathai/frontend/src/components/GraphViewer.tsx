import { forwardRef, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react'
import { Coordinates, Mafs, Plot } from 'mafs'
import 'mafs/core.css'
import { toPng } from 'html-to-image'
import { compileEquation } from '../plot'
import type { Equation, GraphSettings, Viewport } from '../types'

export type GraphViewerHandle = {
  exportPNG: () => Promise<void>
  requestFullscreen: () => Promise<void>
}

type Props = {
  equations: Equation[]
  viewport: Viewport
  settings: GraphSettings
}

export const GraphViewer = forwardRef<GraphViewerHandle, Props>(function GraphViewer(
  { equations, viewport, settings },
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

  const compiled = useMemo(() => {
    const map = new Map<string, (x: number) => number>()
    for (const equation of equations) {
      map.set(equation.id, compileEquation(equation.expression))
    }
    return map
  }, [equations])

  const visible = useMemo(() => equations.filter((e) => e.visible), [equations])

  useImperativeHandle(ref, () => ({
    exportPNG: async () => {
      const el = wrapRef.current
      if (!el) return
      const dataUrl = await toPng(el, {
        backgroundColor: '#ffffff',
        pixelRatio: 2,
        skipFonts: true,
      })
      const anchor = document.createElement('a')
      anchor.href = dataUrl
      anchor.download = 'math-graph.png'
      anchor.click()
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
            preserveAspectRatio={false}
            pan={false}
            zoom={false}
          >
            <Coordinates.Cartesian
              xAxis={{ axis: settings.showAxes, lines: settings.showGrid ? 1 : false }}
              yAxis={{ axis: settings.showAxes, lines: settings.showGrid ? 1 : false }}
            />
            {visible.map((equation) => (
              <Plot.OfX
                key={equation.id}
                y={compiled.get(equation.id) ?? (() => NaN)}
                color={equation.color}
                weight={equation.lineWidth || 2}
                opacity={1}
              />
            ))}
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
    </div>
  )
})
