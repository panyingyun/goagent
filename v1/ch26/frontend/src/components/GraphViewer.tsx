import { forwardRef, useEffect, useImperativeHandle, useRef } from 'react'
import Plotly from 'plotly.js-dist-min'
import { toTrace } from '../plot'
import type { EquationItem, GraphSettings, Viewport } from '../types'

export interface GraphViewerHandle {
  exportPNG: () => Promise<void>
}

interface Props {
  equations: EquationItem[]
  viewport: Viewport
  settings: GraphSettings
}

export const GraphViewer = forwardRef<GraphViewerHandle, Props>(({ equations, viewport, settings }, ref) => {
  const divRef = useRef<HTMLDivElement>(null)

  useImperativeHandle(ref, () => ({
    exportPNG: async () => {
      if (!divRef.current) return
      const img = await Plotly.toImage(divRef.current, { format: 'png', width: 1000, height: 700 })
      const a = document.createElement('a')
      a.href = img
      a.download = 'math-graph.png'
      a.click()
    },
  }))

  useEffect(() => {
    if (!divRef.current) return
    const traces = equations.filter((e) => e.visible).map((e) => toTrace(e, viewport, settings))
    const layout = {
      xaxis: { range: [viewport.xMin, viewport.xMax], showgrid: settings.showGrid, zeroline: settings.showAxis },
      yaxis: { range: [viewport.yMin, viewport.yMax], showgrid: settings.showGrid, zeroline: settings.showAxis },
      showlegend: settings.showLegend,
      margin: { l: 48, r: 24, t: 24, b: 48 },
    }
    Plotly.react(divRef.current, traces, layout, { responsive: true, displaylogo: false })
  }, [equations, viewport, settings])

  return <div ref={divRef} className="h-full w-full" />
})
