import { useMemo, useRef, useState } from 'react'
import { useStore } from '../store'
import { GraphViewer, type GraphViewerHandle } from './GraphViewer'
import { MathTex } from './MathTex'
import type { Equation } from '../types'

type Tab = 'equations' | 'params' | 'analysis'

export function ResultPanel() {
  const session = useStore((s) => s.session)
  const hiddenIds = useStore((s) => s.hiddenIds)
  const toggleHidden = useStore((s) => s.toggleHidden)
  const trace = useStore((s) => s.trace)
  const graphRef = useRef<GraphViewerHandle>(null)
  const [tab, setTab] = useState<Tab>('equations')

  const graph = session?.graphState
  const equations = useMemo(() => {
    if (!graph) return [] as Equation[]
    return graph.equations.map((e) => ({ ...e, visible: e.visible && !hiddenIds.has(e.id) }))
  }, [graph, hiddenIds])

  if (!session || !graph) {
    return (
      <section className="flex min-w-[400px] flex-[2.4] items-center justify-center bg-surface-container-low text-sm text-on-surface-variant">
        正在准备图像面板…
      </section>
    )
  }

  return (
    <section className="flex min-w-0 flex-[2.4] flex-col bg-surface-container-low">
      <div className="flex items-center justify-between border-b border-outline-variant bg-surface-container-lowest px-4 py-3">
        <div className="flex items-center gap-2">
          <span className="text-[11px] font-semibold uppercase tracking-[0.12em] text-on-surface-variant">
            Viewer
          </span>
          <span className="text-sm font-semibold text-on-surface">函数图像</span>
        </div>
        <button
          type="button"
          onClick={() => void graphRef.current?.exportPNG()}
          className="rounded-lg border border-outline-variant px-3 py-1.5 text-xs font-medium text-on-surface-variant transition hover:border-primary hover:text-primary"
        >
          导出 PNG
        </button>
      </div>

      <div className="relative min-h-0 flex-[7] bg-white p-2">
        <GraphViewer
          ref={graphRef}
          equations={equations}
          viewport={graph.viewport}
          settings={graph.settings}
        />
      </div>

      <div className="flex min-h-0 flex-[3] flex-col border-t border-outline-variant bg-surface shadow-[0_-4px_12px_rgba(0,0,0,0.04)]">
        <div className="flex gap-1 border-b border-outline-variant px-4 pt-2">
          {(
            [
              ['equations', `方程 (${graph.equations.length})`],
              ['params', '参数'],
              ['analysis', '图像分析'],
            ] as const
          ).map(([key, label]) => (
            <button
              key={key}
              type="button"
              onClick={() => setTab(key)}
              className={`px-4 py-2 text-[12px] font-semibold uppercase tracking-wider transition ${
                tab === key
                  ? 'border-b-2 border-primary text-primary'
                  : 'border-b-2 border-transparent text-on-surface-variant hover:text-on-surface'
              }`}
            >
              {label}
            </button>
          ))}
        </div>

        <div className="custom-scrollbar min-h-0 flex-1 overflow-y-auto p-4">
          {tab === 'equations' && (
            <div className="space-y-2">
              {graph.equations.length === 0 && (
                <p className="text-sm text-on-surface-variant">还没有方程，在对话中描述你想画的函数</p>
              )}
              {graph.equations.map((e) => {
                const visible = e.visible && !hiddenIds.has(e.id)
                return (
                  <div
                    key={e.id}
                    className="flex items-center justify-between rounded-xl border border-outline-variant bg-white px-3 py-2.5 shadow-sm"
                  >
                    <div className="flex min-w-0 items-center gap-3">
                      <input
                        type="checkbox"
                        checked={visible}
                        onChange={() => toggleHidden(e.id)}
                        className="h-4 w-4 rounded border-outline-variant text-primary focus:ring-primary"
                      />
                      <span className="h-3.5 w-3.5 rounded-sm" style={{ background: e.color }} />
                      <span className="truncate">
                        <MathTex tex={e.label || `y=${e.expression}`} />
                      </span>
                    </div>
                    <span className="font-mono text-[11px] text-on-surface-variant">{e.expression}</span>
                  </div>
                )
              })}
            </div>
          )}

          {tab === 'params' && (
            <dl className="grid grid-cols-2 gap-3 text-sm">
              <div className="rounded-xl border border-outline-variant bg-white p-3">
                <dt className="text-xs text-on-surface-variant">x 范围</dt>
                <dd className="mt-1 font-mono">
                  [{graph.viewport.xMin}, {graph.viewport.xMax}]
                </dd>
              </div>
              <div className="rounded-xl border border-outline-variant bg-white p-3">
                <dt className="text-xs text-on-surface-variant">y 范围</dt>
                <dd className="mt-1 font-mono">
                  [{graph.viewport.yMin}, {graph.viewport.yMax}]
                </dd>
              </div>
              <div className="rounded-xl border border-outline-variant bg-white p-3">
                <dt className="text-xs text-on-surface-variant">网格 / 坐标轴 / 图例</dt>
                <dd className="mt-1">
                  {graph.settings.showGrid ? '开' : '关'} / {graph.settings.showAxes ? '开' : '关'} /{' '}
                  {graph.settings.showLegend ? '开' : '关'}
                </dd>
              </div>
              <div className="rounded-xl border border-outline-variant bg-white p-3">
                <dt className="text-xs text-on-surface-variant">采样点</dt>
                <dd className="mt-1 font-mono">{graph.settings.sampleCount}</dd>
              </div>
            </dl>
          )}

          {tab === 'analysis' && (
            <div className="space-y-3 text-sm">
              {!graph.analysis && (
                <p className="text-on-surface-variant">在对话里说「解释图像」可生成分析</p>
              )}
              {graph.analysis?.summary && <p className="leading-6">{graph.analysis.summary}</p>}
              {graph.analysis?.highlights?.length ? (
                <ul className="list-disc space-y-1 pl-5 text-on-surface-variant">
                  {graph.analysis.highlights.map((item) => (
                    <li key={item}>{item}</li>
                  ))}
                </ul>
              ) : null}
              {trace && (
                <div className="rounded-xl border border-outline-variant bg-white p-3">
                  <div className="mb-2 text-xs font-semibold uppercase tracking-wider text-on-surface-variant">
                    执行轨迹
                  </div>
                  <div className="space-y-1 font-mono text-[11px] text-on-surface-variant">
                    {trace.events.map((ev, i) => (
                      <div key={`${ev.kind}-${i}`}>
                        {ev.kind}
                        {ev.tool ? ` · ${ev.tool}` : ''} · {ev.status}
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </section>
  )
}
