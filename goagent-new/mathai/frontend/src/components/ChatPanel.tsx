import { useEffect, useRef, useState } from 'react'
import { useStore } from '../store'

const QUICK = [
  { icon: 'show_chart', label: '画一条抛物线', prompt: '画一条抛物线 y = x^2，并解释图像特征' },
  { icon: 'waves', label: '比较正弦与余弦', prompt: '比较 y = sin(x) 和 y = cos(x)' },
  { icon: 'straighten', label: '绘制 y = 2x + 1', prompt: '绘制 y = 2x + 1' },
] as const

export function ChatPanel() {
  const session = useStore((s) => s.session)
  const busy = useStore((s) => s.busy)
  const error = useStore((s) => s.error)
  const chat = useStore((s) => s.chat)
  const clearError = useStore((s) => s.clearError)
  const openTrace = useStore((s) => s.openTrace)
  const [input, setInput] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [session?.messages, busy])

  const send = (text?: string) => {
    const content = (text ?? input).trim()
    if (!content || busy) return
    setInput('')
    void chat(content)
  }

  const empty = !session || session.messages.length === 0

  return (
    <main className="relative flex min-w-0 flex-[2] flex-col border-r border-outline-variant bg-surface-bright">
      <div className="flex h-14 shrink-0 items-center border-b border-outline-variant bg-surface-container-lowest/80 px-6 backdrop-blur-sm">
        <h2 className="text-sm font-medium text-on-surface">{empty ? '新会话' : 'AI 对话'}</h2>
        {!empty && session?.graphState?.equations?.[0] && (
          <span className="ml-2 truncate font-mono text-xs text-on-surface-variant">
            · {session.graphState.equations.map((e) => e.label || e.expression).join(' · ')}
          </span>
        )}
      </div>

      <div className="custom-scrollbar min-h-0 flex-1 overflow-y-auto">
        {empty ? (
          <div className="flex h-full flex-col items-center justify-center px-6 py-10">
            <div className="mb-6 flex h-16 w-16 items-center justify-center overflow-hidden rounded-2xl bg-primary-container shadow-soft">
              <img src="/logo.png" alt="" className="h-12 w-12 object-contain" />
            </div>
            <h1 className="mb-8 max-w-md text-center text-[32px] font-semibold leading-10 tracking-tight text-on-surface">
              你想绘制什么数学图像？
            </h1>
            <div className="flex w-full max-w-sm flex-col gap-3">
              {QUICK.map((item) => (
                <button
                  key={item.label}
                  type="button"
                  onClick={() => send(item.prompt)}
                  className="group flex w-full items-center gap-3 rounded-xl border border-outline-variant bg-surface-container-lowest px-4 py-3 text-left transition hover:border-primary hover:bg-surface-container-low"
                >
                  <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-surface-container text-on-surface-variant transition group-hover:bg-primary-container group-hover:text-on-primary-container">
                    <span className="material-symbols-outlined text-[18px]">{item.icon}</span>
                  </div>
                  <span className="text-sm text-on-surface">{item.label}</span>
                </button>
              ))}
            </div>
          </div>
        ) : (
          <div className="space-y-5 px-5 py-5 pb-4">
            {session.messages.map((m) =>
              m.role === 'user' ? (
                <div key={m.id} className="flex justify-end">
                  <div className="max-w-[85%] rounded-2xl rounded-tr-sm border border-primary/10 bg-user-bubble px-4 py-3 text-sm leading-6 text-on-surface shadow-soft">
                    <p className="whitespace-pre-wrap">{m.content}</p>
                  </div>
                </div>
              ) : (
                <div key={m.id} className="flex flex-col items-start gap-2">
                  <div className="flex items-center gap-2">
                    <div className="flex h-6 w-6 items-center justify-center overflow-hidden rounded-full bg-primary">
                      <img src="/logo.png" alt="" className="h-5 w-5 object-contain brightness-0 invert" />
                    </div>
                    <span className="text-xs font-medium text-on-surface-variant">MathGraph AI</span>
                    <span className="rounded-full bg-secondary-fixed px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-secondary">
                      DeepSeek
                    </span>
                  </div>
                  <div className="max-w-[90%] space-y-3 rounded-2xl rounded-tl-sm border border-outline-variant bg-surface-container-lowest px-5 py-4 text-sm leading-6 shadow-soft">
                    <p className="whitespace-pre-wrap">{m.content}</p>
                    {m.traceId && (
                      <button
                        type="button"
                        onClick={() => void openTrace(m.traceId!)}
                        className="rounded-lg border border-outline-variant px-3 py-1.5 text-xs text-on-surface-variant transition hover:border-primary hover:text-primary"
                      >
                        查看执行轨迹
                      </button>
                    )}
                  </div>
                </div>
              ),
            )}

            {busy && (
              <div className="flex flex-col items-start gap-2">
                <div className="flex items-center gap-2">
                  <div className="flex h-6 w-6 items-center justify-center rounded-full bg-primary text-on-primary">
                    <span className="material-symbols-outlined text-[14px]">psychology</span>
                  </div>
                  <span className="text-xs font-medium text-on-surface-variant">MathGraph AI</span>
                </div>
                <div className="rounded-2xl rounded-tl-sm border border-outline-variant bg-surface-container-low px-5 py-4">
                  <div className="typing-indicator flex items-center gap-1">
                    <span /><span /><span />
                  </div>
                </div>
              </div>
            )}

            {error && !/解析失败|方程/.test(error) && (
              <div className="rounded-xl border border-error/30 bg-error-container px-4 py-3 text-sm text-on-error-container">
                {error}
                <button type="button" className="ml-3 underline" onClick={clearError}>
                  关闭
                </button>
              </div>
            )}
            <div ref={bottomRef} />
          </div>
        )}
      </div>

      <div className="shrink-0 border-t border-outline-variant bg-surface-container-lowest p-4">
        <div className="relative flex items-end gap-2 rounded-2xl border border-outline-variant bg-surface-container-low p-2 transition focus-within:border-primary focus-within:ring-1 focus-within:ring-primary">
          <button type="button" className="p-2 text-on-surface-variant transition hover:text-primary" title="快捷">
            <span className="material-symbols-outlined">add_circle</span>
          </button>
          <textarea
            value={input}
            rows={1}
            maxLength={4000}
            placeholder="输入方程式或描述，例如：画 y = x²"
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                send()
              }
            }}
            className="max-h-32 min-h-[40px] flex-1 resize-none border-0 bg-transparent py-2 text-sm outline-none"
          />
          <button
            type="button"
            disabled={!input.trim() || busy}
            onClick={() => send()}
            className="rounded-xl bg-primary p-2 text-on-primary shadow-soft transition hover:opacity-90 disabled:opacity-40"
            aria-label="发送"
          >
            <span className="material-symbols-outlined" style={{ fontVariationSettings: "'FILL' 1" }}>
              send
            </span>
          </button>
        </div>
        <p className="mt-2 text-center text-[12px] tracking-wide text-outline">
          MathGraph AI 可以辅助数学可视化与分析 · Enter 发送
        </p>
      </div>
    </main>
  )
}
