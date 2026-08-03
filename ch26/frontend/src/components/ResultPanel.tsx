import { useEffect, useRef, useState } from 'react'
import { useStore } from '../store'
import { GraphViewer, type GraphViewerHandle } from './GraphViewer'
import { Math } from './Math'
import type { GraphState } from '../types'

type Tab = 'equations' | 'viewport' | 'analysis'

export function ResultPanel() {
  const current = useStore((s) => s.current)
  const updateGraphState = useStore((s) => s.updateGraphState)
  const graphRef = useRef<GraphViewerHandle>(null)
  const [tab, setTab] = useState<Tab>('equations')
  const gs: GraphState | undefined = current?.graphState

  const [viewportDraft, setViewportDraft] = useState<GraphState['viewport'] | undefined>(gs?.viewport)
  useEffect(() => {
    if (gs) setViewportDraft({ ...gs.viewport })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [gs?.viewport.xMin, gs?.viewport.xMax, gs?.viewport.yMin, gs?.viewport.yMax])

  if (!current || !gs) return null

  const applyViewport = () => {
    if (!viewportDraft) return
    void updateGraphState({ ...gs, viewport: viewportDraft })
  }

  const toggleVisible = (id: string) => {
    void updateGraphState({
      ...gs,
      equations: gs.equations.map((e) => (e.id === id ? { ...e, visible: !e.visible } : e)),
    })
  }

  const removeEquation = (id: string) => {
    void updateGraphState({ ...gs, equations: gs.equations.filter((e) => e.id !== id) })
  }

  const setColor = (id: string, color: string) => {
    void updateGraphState({
      ...gs,
      equations: gs.equations.map((e) => (e.id === id ? { ...e, color } : e)),
    })
  }

  return (
    <section className="flex min-w-0 flex-[2.5] flex-col">
      <div className="flex items-center justify-between px-4 py-2">
        <span className="text-sm font-semibold">绘图结果</span>
        <button
          onClick={() => void graphRef.current?.exportPNG()}
          className="rounded-md border border-outline px-3 py-1 text-sm hover:bg-surface-container"
        >
          导出 PNG
        </button>
      </div>

      <div className="min-h-0 flex-[2] border-b border-outline bg-white p-2">
        <GraphViewer ref={graphRef} equations={gs.equations} viewport={gs.viewport} settings={gs.settings} />
      </div>

      <div className="flex min-h-0 flex-1 flex-col">
        <div className="flex gap-1 border-b border-outline px-3">
          {(['equations', 'viewport', 'analysis'] as Tab[]).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`px-3 py-2 text-sm ${tab === t ? 'border-b-2 border-primary font-semibold' : 'text-on-surface-variant'}`}
            >
              {{ equations: '方程列表', viewport: '视图设置', analysis: '图像分析' }[t]}
            </button>
          ))}
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto p-3">
          {tab === 'equations' && (
            <div className="space-y-2">
              {gs.equations.length === 0 && <div className="text-sm text-on-surface-variant">还没有方程</div>}
              {gs.equations.map((e) => (
                <div key={e.id} className="flex items-center gap-2 rounded-md border border-outline bg-white px-3 py-2 text-sm">
                  <input type="checkbox" checked={e.visible} onChange={() => toggleVisible(e.id)} />
                  <span className="min-w-0 flex-1">
                    <Math tex={e.label} />
                  </span>
                  <input type="color" value={e.color} onChange={(ev) => setColor(e.id, ev.target.value)} className="h-6 w-8" />
                  <button onClick={() => removeEquation(e.id)} className="text-on-surface-variant hover:text-red-600">
                    删除
                  </button>
                </div>
              ))}
            </div>
          )}
          {tab === 'viewport' && viewportDraft && (
            <div className="space-y-3">
              <div className="grid grid-cols-2 gap-2">
                {(['xMin', 'xMax', 'yMin', 'yMax'] as const).map((k) => (
                  <label key={k} className="flex items-center gap-2 text-sm">
                    <span className="w-10">{k}</span>
                    <input
                      type="number"
                      value={viewportDraft[k]}
                      onChange={(e) => setViewportDraft({ ...viewportDraft, [k]: Number(e.target.value) })}
                      className="w-full rounded-md border border-outline px-2 py-1"
                    />
                  </label>
                ))}
              </div>
              <div className="flex flex-wrap gap-4 text-sm">
                {(['showGrid', 'showAxis', 'showLegend'] as const).map((k) => (
                  <label key={k} className="flex items-center gap-1">
                    <input
                      type="checkbox"
                      checked={gs.settings[k]}
                      onChange={(e) =>
                        void updateGraphState({ ...gs, settings: { ...gs.settings, [k]: e.target.checked } })
                      }
                    />
                    {k}
                  </label>
                ))}
                <label className="flex items-center gap-1">
                  采样
                  <select
                    value={gs.settings.sampleCount}
                    onChange={(e) =>
                      void updateGraphState({ ...gs, settings: { ...gs.settings, sampleCount: Number(e.target.value) } })
                    }
                    className="rounded-md border border-outline px-1 py-0.5"
                  >
                    <option value={200}>低</option>
                    <option value={600}>中</option>
                    <option value={1200}>高</option>
                  </select>
                </label>
              </div>
              <button
                onClick={applyViewport}
                className="rounded-md bg-primary px-3 py-1.5 text-sm text-white"
              >
                应用坐标范围
              </button>
            </div>
          )}
          {tab === 'analysis' && (
            <div className="space-y-1 text-sm">
              {!gs.analysis && <div className="text-on-surface-variant">在对话里说"解释图像"来生成分析</div>}
              {gs.analysis?.functionType && <div>函数类型:{gs.analysis.functionType}</div>}
              {gs.analysis?.symmetry && <div>对称性:{gs.analysis.symmetry}</div>}
              {gs.analysis?.zeros && (
                <div>零点:{Array.isArray(gs.analysis.zeros) ? gs.analysis.zeros.join(', ') : gs.analysis.zeros}</div>
              )}
              {gs.analysis?.monotonicity &&
                (Array.isArray(gs.analysis.monotonicity)
                  ? gs.analysis.monotonicity.map((m) => <div key={m}>单调性:{m}</div>)
                  : <div>单调性:{gs.analysis.monotonicity}</div>)}
              {gs.analysis?.description && <div className="mt-2">{gs.analysis.description}</div>}
            </div>
          )}
        </div>
      </div>
    </section>
  )
}
