import { useEffect, useMemo, useRef, useState } from 'react'
import {
  CheckCircle2,
  Copy,
  Download,
  Eye,
  EyeOff,
  FunctionSquare,
  Home,
  Maximize2,
} from 'lucide-react'
import { useStore } from '../store'
import { GraphViewer, type GraphViewerHandle } from './GraphViewer'
import { MathTex } from './MathTex'
import type { Equation, GraphSettings, PiLabelMode, Viewport } from '../types'

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
  const [viewport, setViewport] = useState<Viewport | null>(null)
  const [baseViewport, setBaseViewport] = useState<Viewport | null>(null)
  const [settings, setSettings] = useState<GraphSettings | null>(null)
  const [piLabelMode, setPiLabelMode] = useState<PiLabelMode>('auto')

  const graph = session?.graphState

  useEffect(() => {
    if (!graph) {
      setViewport(null)
      setBaseViewport(null)
      setSettings(null)
      return
    }
    setViewport({ ...graph.viewport })
    setBaseViewport({ ...graph.viewport })
    setSettings({ ...graph.settings })
  }, [graph])

  const equations = useMemo(() => {
    if (!graph) return [] as Equation[]
    return graph.equations.map((e) => ({ ...e, visible: e.visible && !hiddenIds.has(e.id) }))
  }, [graph, hiddenIds])

  const parseError = Boolean(error && /解析失败|方程|表达式|syntax|invalid/i.test(error))

  if (!session || !graph || !viewport || !settings) {
    return (
      <section className="flex min-w-[400px] flex-[2.4] flex-col bg-surface-low">
        <ViewerHeader onExport={() => undefined} disabled />
        <div className="relative min-h-0 flex-1">
          <EmptyGrid />
          <div className="absolute bottom-4 left-4 right-4 rounded-xl border border-outline bg-white p-4 shadow-soft">
            <div className="mb-2 flex items-center gap-2 text-xs font-semibold tracking-wider text-faint">
              <Eye className="h-3.5 w-3.5" strokeWidth={1.75} />
              Equations
            </div>
            <div className="flex items-center gap-2 text-sm text-muted">
              <FunctionSquare className="h-4 w-4" strokeWidth={1.75} />
              无活跃方程
            </div>
          </div>
        </div>
      </section>
    )
  }

  return (
    <section className="flex min-w-0 flex-[2.4] flex-col bg-surface-low">
      <ViewerHeader onExport={() => void graphRef.current?.exportPNG()} />

      <div className="relative min-h-0 flex-[7] bg-white p-2">
        <GraphViewer ref={graphRef} equations={equations} viewport={viewport} settings={settings} piLabelMode={piLabelMode} />
        <GraphToolbar
          onReset={() => {
            if (baseViewport) setViewport({ ...baseViewport })
            graphRef.current?.resetView()
          }}
          onFullscreen={() => void graphRef.current?.requestFullscreen()}
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
          <button type="button" onClick={clearError} className="shrink-0 text-sm underline">
            关闭
          </button>
        </div>
      )}

      <div className="flex min-h-0 flex-[3] flex-col border-t border-outline bg-white shadow-[0_-4px_12px_rgba(25,27,35,0.04)]">
        <div className="flex gap-1 border-b border-outline px-4 pt-2">
          {(
            [
              ['equations', '方程', String(graph.equations.length)],
              ['params', '参数', null],
              ['analysis', '图像分析', null],
            ] as const
          ).map(([key, label, badge]) => (
            <button
              key={key}
              type="button"
              onClick={() => setTab(key)}
              className={`flex items-center gap-1.5 px-4 py-2 text-[13px] font-medium transition ${
                tab === key
                  ? 'border-b-2 border-primary text-primary'
                  : 'border-b-2 border-transparent text-faint hover:text-ink'
              }`}
            >
              {label}
              {badge != null && (
                <span
                  className={`rounded-full px-1.5 text-[11px] ${
                    tab === key ? 'bg-primary/10 text-primary' : 'bg-surface-mid text-muted'
                  }`}
                >
                  {badge}
                </span>
              )}
            </button>
          ))}
        </div>

        <div className="custom-scrollbar min-h-0 flex-1 overflow-y-auto p-4">
          {tab === 'equations' && (
            <div className="space-y-2">
              {graph.equations.length === 0 && (
                <div className="rounded-xl border border-outline bg-surface-low p-4 text-sm text-muted">
                  无活跃方程
                </div>
              )}
              {graph.equations.map((e) => {
                const visible = e.visible && !hiddenIds.has(e.id)
                return (
                  <div
                    key={e.id}
                    className="flex items-center justify-between rounded-xl border border-outline bg-surface-low px-3 py-2.5 shadow-soft"
                  >
                    <div className="flex min-w-0 items-center gap-3">
                      <button
                        type="button"
                        title={visible ? '隐藏' : '显示'}
                        onClick={() => toggleHidden(e.id)}
                        className="text-faint hover:text-primary"
                      >
                        {visible ? (
                          <Eye className="h-4 w-4" strokeWidth={1.75} />
                        ) : (
                          <EyeOff className="h-4 w-4" strokeWidth={1.75} />
                        )}
                      </button>
                      <span className="h-3.5 w-3.5 rounded-sm" style={{ background: e.color }} />
                      <div className="min-w-0">
                        <div className="truncate">
                          <MathTex tex={e.label || `y=${e.expression}`} />
                        </div>
                        <div className="mt-0.5 font-mono text-[11px] text-faint">
                          {e.expression} · 线宽 {e.lineWidth || 2}px
                        </div>
                      </div>
                    </div>
                    <button
                      type="button"
                      className="rounded p-1 text-faint hover:text-primary"
                      onClick={() => void navigator.clipboard?.writeText(e.expression)}
                      title="复制"
                    >
                      <Copy className="h-4 w-4" strokeWidth={1.75} />
                    </button>
                  </div>
                )
              })}
            </div>
          )}

          {tab === 'params' && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <NumberField
                  label="x 最小值"
                  value={viewport.xMin}
                  onChange={(v) => setViewport({ ...viewport, xMin: v })}
                />
                <NumberField
                  label="x 最大值"
                  value={viewport.xMax}
                  onChange={(v) => setViewport({ ...viewport, xMax: v })}
                />
                <NumberField
                  label="y 最小值"
                  value={viewport.yMin}
                  onChange={(v) => setViewport({ ...viewport, yMin: v })}
                />
                <NumberField
                  label="y 最大值"
                  value={viewport.yMax}
                  onChange={(v) => setViewport({ ...viewport, yMax: v })}
                />
              </div>
              <div className="flex flex-wrap gap-6 pt-1">
                <ToggleRow
                  label="网格"
                  on={settings.showGrid}
                  onToggle={() => setSettings({ ...settings, showGrid: !settings.showGrid })}
                />
                <ToggleRow
                  label="坐标轴"
                  on={settings.showAxes}
                  onToggle={() => setSettings({ ...settings, showAxes: !settings.showAxes })}
                />
                <ToggleRow
                  label="图例"
                  on={settings.showLegend}
                  onToggle={() => setSettings({ ...settings, showLegend: !settings.showLegend })}
                />
                <div className="flex items-center gap-2.5">
                  <span className="text-sm text-ink">横轴刻度</span>
                  <div className="flex overflow-hidden rounded-lg border border-outline">
                    {(
                      [
                        ['auto', '自动'],
                        ['pi', 'π'],
                        ['integer', '整数'],
                      ] as const
                    ).map(([mode, label]) => (
                      <button
                        key={mode}
                        type="button"
                        onClick={() => setPiLabelMode(mode)}
                        className={`px-3 py-1.5 text-sm transition ${
                          piLabelMode === mode
                            ? 'bg-primary text-on-primary'
                            : 'bg-white text-muted hover:text-ink'
                        }`}
                      >
                        {label}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          )}

          {tab === 'analysis' && (
            <div className="space-y-3 text-sm">
              {!graph.analysis && <p className="text-muted">在对话里说「解释图像」可生成分析</p>}
              {graph.analysis?.highlights?.length ? (
                <div className="space-y-2">
                  {graph.analysis.highlights.map((item) => (
                    <div
                      key={item}
                      className="flex items-start gap-2 rounded-xl border border-outline bg-surface-low px-3 py-2.5"
                    >
                      <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-primary" strokeWidth={1.75} />
                      <span className="leading-6 text-ink">{item}</span>
                    </div>
                  ))}
                </div>
              ) : null}
              {graph.analysis?.summary && (
                <p className="rounded-xl border border-outline bg-white px-3 py-3 leading-6 text-muted">
                  {graph.analysis.summary}
                </p>
              )}
              {trace && (
                <div className="rounded-xl border border-outline bg-surface-low p-3">
                  <div className="mb-2 text-xs font-semibold tracking-wider text-faint">执行轨迹</div>
                  <div className="space-y-1 font-mono text-[11px] text-muted">
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

function GraphToolbar({
  onReset,
  onFullscreen,
}: {
  onReset: () => void
  onFullscreen: () => void
}) {
  const items = [
    { label: '重置视图', icon: Home, onClick: onReset },
    { label: '全屏', icon: Maximize2, onClick: onFullscreen },
  ] as const

  return (
    <div className="absolute right-4 top-4 z-10 flex flex-col overflow-hidden rounded-xl border border-outline bg-white/95 shadow-soft backdrop-blur-sm">
      {items.map(({ label, icon: Icon, onClick }, i) => (
        <button
          key={label}
          type="button"
          title={label}
          onClick={onClick}
          className={`flex h-9 w-9 items-center justify-center text-muted transition hover:bg-primary-soft hover:text-primary ${
            i > 0 ? 'border-t border-outline' : ''
          }`}
        >
          <Icon className="h-4 w-4" strokeWidth={1.75} />
        </button>
      ))}
    </div>
  )
}

function ViewerHeader({ onExport, disabled }: { onExport: () => void; disabled?: boolean }) {
  return (
    <div className="flex items-center justify-between border-b border-outline bg-white px-4 py-3">
      <div className="flex flex-col">
        <span className="text-[10px] font-semibold uppercase tracking-[0.16em] text-faint">Viewer</span>
        <span className="text-[15px] font-semibold leading-tight text-ink">函数图像</span>
      </div>
      <button
        type="button"
        disabled={disabled}
        onClick={onExport}
        className="inline-flex items-center gap-1.5 rounded-lg border border-outline px-3 py-1.5 text-xs font-medium text-muted transition hover:border-primary hover:text-primary disabled:opacity-40"
      >
        <Download className="h-3.5 w-3.5" strokeWidth={1.75} />
        导出 PNG
      </button>
    </div>
  )
}

function NumberField({
  label,
  value,
  onChange,
}: {
  label: string
  value: number
  onChange: (v: number) => void
}) {
  return (
    <label className="block">
      <span className="mb-1.5 block text-xs text-faint">{label}</span>
      <input
        type="number"
        value={value}
        onChange={(e) => {
          const n = Number(e.target.value)
          if (!Number.isNaN(n)) onChange(n)
        }}
        className="w-full rounded-xl border border-outline bg-surface-low px-3 py-2 font-mono text-sm text-ink outline-none focus:border-primary focus:ring-1 focus:ring-primary/30"
      />
    </label>
  )
}

function ToggleRow({ label, on, onToggle }: { label: string; on: boolean; onToggle: () => void }) {
  return (
    <div className="flex items-center gap-2.5">
      <button type="button" className="toggle" data-on={on} onClick={onToggle} aria-label={label} />
      <span className="text-sm text-ink">{label}</span>
    </div>
  )
}

function EmptyGrid() {
  return (
    <div className="math-grid relative h-full w-full overflow-hidden bg-white">
      <div className="math-axes" />
    </div>
  )
}
