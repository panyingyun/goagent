import { forwardRef, useEffect, useImperativeHandle, useRef } from 'react'
import Plotly from 'plotly.js-dist-min'
import { toTrace } from '../plot'
import type { Equation, GraphSettings, Viewport } from '../types'

export type GraphViewerHandle = { exportPNG: () => Promise<void> }

type Props = {
  equations: Equation[]
  viewport: Viewport
  settings: GraphSettings
}

export const GraphViewer = forwardRef<GraphViewerHandle, Props>(function GraphViewer(
  { equations, viewport, settings },
  ref,
) {
  const divRef = useRef<HTMLDivElement>(null)

  useImperativeHandle(ref, () => ({
    exportPNG: async () => {
      if (!divRef.current) return
      const img = await Plotly.toImage(divRef.current, { format: 'png', width: 1200, height: 800 })
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
      paper_bgcolor: '#ffffff',
      plot_bgcolor: '#ffffff',
      font: { family: 'Inter, sans-serif', color: '#434655', size: 12 },
      xaxis: {
        range: [viewport.xMin, viewport.xMax],
        showgrid: settings.showGrid,
        zeroline: settings.showAxes,
        gridcolor: '#e1e2ed',
        zerolinecolor: '#191b23',
      },
      yaxis: {
        range: [viewport.yMin, viewport.yMax],
        showgrid: settings.showGrid,
        zeroline: settings.showAxes,
        gridcolor: '#e1e2ed',
        zerolinecolor: '#191b23',
      },
      showlegend: settings.showLegend,
      margin: { l: 48, r: 24, t: 28, b: 48 },
      legend: { orientation: 'h', y: 1.08 },
    }
    void Plotly.react(divRef.current, traces, layout, {
      responsive: true,
      displaylogo: false,
      modeBarButtonsToRemove: ['lasso2d', 'select2d'],
    })
  }, [equations, viewport, settings])

  return <div ref={divRef} className="h-full w-full" />
})
