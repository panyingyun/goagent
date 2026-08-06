import { useEffect, useRef, useState } from 'react'
import { useStore } from '../store'

const EXAMPLES = [
  '画一条抛物线',
  '比较正弦与余弦',
  '绘制 y = 2x + 1',
  '两条曲线求交点',
  '三次函数求极值',
  '解释这个图像',
]

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

  return (
    <main className="relative flex min-w-0 flex-[2] flex-col border-r border-outline-variant bg-surface-container-lowest">
      <div className="flex items-center gap-2 border-b border-outline-variant px-5 py-3">
        <span className="text-sm font-semibold text-on-surface">AI 对话</span>
        {session?.graphState?.equations?.[0] && (
          <span className="truncate font-mono text-xs text-on-surface-variant">
            · {session.graphState.equations.map((e) => e.label || e.expression).join(' · ')}
          </span>
        )}
      </div>

      <div className="custom-scrollbar min-h-0 flex-1 space-y-5 overflow-y-auto px-5 py-5 pb-36">
        {(!session || session.messages.length === 0) && (
          <div className="mx-auto mt-16 max-w-md text-center">
            <h2 className="text-2xl font-semibold tracking-tight text-on-surface">你想绘制什么数学图像？</h2>
            <p className="mt-2 text-sm text-on-surface-variant">用自然语言描述，或直接输入一个关于 x 的方程。</p>
            <div className="mt-6 flex flex-wrap justify-center gap-2">
              {EXAMPLES.map((example) => (
                <button
                  key={example}
                  type="button"
                  onClick={() => send(example)}
                  className="rounded-full border border-outline-variant bg-white px-3 py-1.5 text-xs text-on-surface-variant transition hover:border-primary hover:text-primary"
                >
                  {example}
                </button>
              ))}
            </div>
          </div>
        )}

        {session?.messages.map((m) =>
          m.role === 'user' ? (
            <div key={m.id} className="flex justify-end">
              <div className="max-w-[85%] rounded-2xl rounded-tr-sm bg-user-bubble px-4 py-3 text-sm leading-6 text-on-surface shadow-sm">
                <p className="whitespace-pre-wrap">{m.content}</p>
              </div>
            </div>
          ) : (
            <div key={m.id} className="flex flex-col items-start gap-2">
              <div className="flex items-center gap-2">
                <div className="flex h-6 w-6 items-center justify-center rounded-full bg-primary text-on-primary">
                  <span className="material-symbols-outlined text-[14px]">psychology</span>
                </div>
                <span className="text-xs font-medium text-on-surface-variant">MathGraph AI</span>
                <span className="rounded-full bg-surface-container px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide text-primary">
                  DeepSeek
                </span>
              </div>
              <div className="max-w-[90%] space-y-3 rounded-2xl rounded-tl-sm border border-outline-variant bg-white px-5 py-4 text-sm leading-6 shadow-sm">
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
          <div className="flex items-center gap-2 text-sm text-on-surface-variant">
            <span className="material-symbols-outlined animate-spin text-[18px]">progress_activity</span>
            AI 正在思考…
          </div>
        )}

        {error && (
          <div className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            {error}
            <button type="button" className="ml-3 underline" onClick={clearError}>
              关闭
            </button>
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      <div className="absolute bottom-0 left-0 w-full bg-gradient-to-t from-surface-container-lowest via-surface-container-lowest to-transparent px-5 pb-5 pt-10">
        <div className="relative rounded-2xl border border-outline-variant bg-white shadow-sm focus-within:border-primary focus-within:ring-1 focus-within:ring-primary">
          <textarea
            value={input}
            rows={3}
            maxLength={4000}
            placeholder="输入方程式或描述，例如：画 y = x²"
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                send()
              }
            }}
            className="w-full resize-none border-0 bg-transparent px-4 pb-10 pt-4 text-sm outline-none"
          />
          <button
            type="button"
            disabled={!input.trim() || busy}
            onClick={() => send()}
            className="absolute bottom-3 right-3 flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-on-primary shadow-sm transition hover:bg-primary-container disabled:opacity-40"
            aria-label="发送"
          >
            <span className="material-symbols-outlined text-[18px]">send</span>
          </button>
        </div>
        <p className="mt-2 text-center text-[11px] text-on-surface-variant">
          Enter 发送 · Shift+Enter 换行
        </p>
      </div>
    </main>
  )
}
