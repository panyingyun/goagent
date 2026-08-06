import { forwardRef, useEffect, useImperativeHandle, useRef } from 'react'
import Plotly from 'plotly.js-dist-min'
import { toTrace } from '../plot'
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

  useEffect(() => {
    if (!divRef.current) return
    const traces = equations.filter((e) => e.visible).map((e) => toTrace(e, viewport, settings))
    const layout = {
      paper_bgcolor: '#ffffff',
      plot_bgcolor: '#ffffff',
      font: { family: 'Inter, LXGWWenKai, sans-serif', color: '#434655', size: 12 },
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
      margin: { l: 48, r: 56, t: 36, b: 48 },
      legend: {
        orientation: 'h',
        x: 0,
        y: 1.08,
        xanchor: 'left',
        bgcolor: 'rgba(255,255,255,0.85)',
        bordercolor: '#e1e2ed',
        borderwidth: 1,
      },
    }
    void Plotly.react(divRef.current, traces, layout, {
      responsive: true,
      displaylogo: false,
      displayModeBar: false,
    })
  }, [equations, viewport, settings])

  return (
    <div ref={wrapRef} className="h-full w-full bg-white">
      <div ref={divRef} className="h-full w-full" />
    </div>
  )
})
