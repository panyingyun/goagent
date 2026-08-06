import { useMemo, useRef, useState } from 'react'
import { useStore } from '../store'
import { GraphViewer, type GraphViewerHandle } from './GraphViewer'
import { MathTex } from './MathTex'
import type { Equation } from '../types'

type Tab = 'equations' | 'params' | 'analysis'

export function ResultPanel() {
  const session = useStore((s) => s.session)
  const busy = useStore((s) => s.busy)
  const error = useStore((s) => s.error)
  const clearError = useStore((s) => s.clearError)
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

  const parseError = Boolean(error && /解析失败|方程|表达式|syntax|invalid/i.test(error))

  if (!session || !graph) {
    return (
      <section className="flex min-w-[400px] flex-[2.4] flex-col bg-surface-container-low">
        <ViewerHeader onExport={() => undefined} disabled />
        <div className="relative min-h-0 flex-1">
          <EmptyGrid />
          <div className="absolute bottom-4 left-4 right-4 rounded-xl border border-outline-variant bg-surface-container-lowest p-4 shadow-soft">
            <div className="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-on-surface-variant">
              <span className="material-symbols-outlined text-[16px]">visibility</span>
              Equations
            </div>
            <div className="flex items-center gap-2 text-sm text-on-surface-variant">
              <span className="material-symbols-outlined text-[18px]">functions</span>
              无活跃方程
            </div>
          </div>
        </div>
      </section>
    )
  }

  return (
    <section className="flex min-w-0 flex-[2.4] flex-col bg-surface-container-low">
      <ViewerHeader onExport={() => void graphRef.current?.exportPNG()} />

      <div className="relative min-h-0 flex-[7] bg-white p-2">
        <GraphViewer
          ref={graphRef}
          equations={equations}
          viewport={graph.viewport}
          settings={graph.settings}
        />
        {busy && (
          <div className="absolute inset-0 flex flex-col items-center justify-center bg-white/70 backdrop-blur-[1px]">
            <div className="loader mb-3" />
            <p className="text-sm font-medium text-primary">Processing Equation…</p>
          </div>
        )}
      </div>

      {parseError && (
        <div className="flex items-center justify-between gap-3 border-t border-error/20 bg-error-container px-4 py-3 text-sm text-on-error-container">
          <div className="flex min-w-0 items-center gap-2">
            <span className="h-2 w-2 shrink-0 rounded-full bg-error" />
            <span className="truncate font-mono">{error}</span>
          </div>
          <button type="button" onClick={clearError} className="flex shrink-0 items-center gap-1">
            <span className="material-symbols-outlined text-[18px]">error</span>
            关闭
          </button>
        </div>
      )}

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
                <div className="rounded-xl border border-outline-variant bg-surface-container-lowest p-4 text-sm text-on-surface-variant">
                  <div className="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-wider">
                    <span className="material-symbols-outlined text-[16px]">visibility</span>
                    Equations
                  </div>
                  无活跃方程
                </div>
              )}
              {graph.equations.map((e) => {
                const visible = e.visible && !hiddenIds.has(e.id)
                return (
                  <div
                    key={e.id}
                    className="flex items-center justify-between rounded-xl border border-outline-variant bg-surface-container-lowest px-3 py-2.5 shadow-soft"
                  >
                    <div className="flex min-w-0 items-center gap-3">
                      <button
                        type="button"
                        title={visible ? '隐藏' : '显示'}
                        onClick={() => toggleHidden(e.id)}
                        className="text-on-surface-variant hover:text-primary"
                      >
                        <span className="material-symbols-outlined text-[18px]">
                          {visible ? 'visibility' : 'visibility_off'}
                        </span>
                      </button>
                      <span className="h-3.5 w-3.5 rounded-sm" style={{ background: e.color }} />
                      <div className="min-w-0">
                        <div className="truncate">
                          <MathTex tex={e.label || `y=${e.expression}`} />
                        </div>
                        <div className="mt-0.5 font-mono text-[11px] text-on-surface-variant">
                          {e.expression} · 线宽 {e.lineWidth || 2}px
                        </div>
                      </div>
                    </div>
                    <button
                      type="button"
                      className="rounded p-1 text-on-surface-variant hover:text-primary"
                      onClick={() => void navigator.clipboard?.writeText(e.expression)}
                      title="复制"
                    >
                      <span className="material-symbols-outlined text-[18px]">content_copy</span>
                    </button>
                  </div>
                )
              })}
            </div>
          )}

          {tab === 'params' && (
            <dl className="grid grid-cols-2 gap-3 text-sm">
              <ParamCard label="x 范围" value={`[${graph.viewport.xMin}, ${graph.viewport.xMax}]`} />
              <ParamCard label="y 范围" value={`[${graph.viewport.yMin}, ${graph.viewport.yMax}]`} />
              <ParamCard
                label="网格 / 坐标轴 / 图例"
                value={`${graph.settings.showGrid ? '开' : '关'} / ${graph.settings.showAxes ? '开' : '关'} / ${graph.settings.showLegend ? '开' : '关'}`}
              />
              <ParamCard label="采样点" value={String(graph.settings.sampleCount)} />
            </dl>
          )}

          {tab === 'analysis' && (
            <div className="space-y-3 text-sm">
              {!graph.analysis && <p className="text-on-surface-variant">在对话里说「解释图像」可生成分析</p>}
              {graph.analysis?.summary && <p className="leading-6">{graph.analysis.summary}</p>}
              {graph.analysis?.highlights?.length ? (
                <ul className="list-disc space-y-1 pl-5 text-on-surface-variant">
                  {graph.analysis.highlights.map((item) => (
                    <li key={item}>{item}</li>
                  ))}
                </ul>
              ) : null}
              {trace && (
                <div className="rounded-xl border border-outline-variant bg-surface-container-lowest p-3">
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

function ViewerHeader({ onExport, disabled }: { onExport: () => void; disabled?: boolean }) {
  return (
    <div className="flex items-center justify-between border-b border-outline-variant bg-surface-container-lowest px-4 py-3">
      <div className="flex items-center gap-2">
        <span className="text-[11px] font-semibold uppercase tracking-[0.14em] text-on-surface-variant">Viewer</span>
        <span className="text-sm font-semibold text-on-surface">函数图像</span>
      </div>
      <button
        type="button"
        disabled={disabled}
        onClick={onExport}
        className="inline-flex items-center gap-1.5 rounded-lg border border-outline-variant px-3 py-1.5 text-xs font-medium text-on-surface-variant transition hover:border-primary hover:text-primary disabled:opacity-40"
      >
        <span className="material-symbols-outlined text-[16px]">download</span>
        导出 PNG
      </button>
    </div>
  )
}

function ParamCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-outline-variant bg-surface-container-lowest p-3">
      <dt className="text-xs text-on-surface-variant">{label}</dt>
      <dd className="mt-1 font-mono text-on-surface">{value}</dd>
    </div>
  )
}

function EmptyGrid() {
  return (
    <div className="math-grid relative h-full w-full overflow-hidden bg-surface-container-lowest">
      <div className="math-axes" />
    </div>
  )
}
