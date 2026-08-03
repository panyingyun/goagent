import { useState } from 'react'
import { useStore } from '../store'
import { Math } from './Math'

const QUICK_PROMPTS = ['画 y = x^2', '画 y = sin(x)', '再加一条 y = cos(x)', '把第一条曲线改成红色', '把坐标范围改成 -5 到 5', '解释图像']

export function ChatPanel() {
  const current = useStore((s) => s.current)
  const sending = useStore((s) => s.sending)
  const error = useStore((s) => s.error)
  const sendMessage = useStore((s) => s.sendMessage)
  const clearError = useStore((s) => s.clearError)
  const [input, setInput] = useState('')

  const send = (text?: string) => {
    const content = (text ?? input).trim()
    if (!content || sending) return
    setInput('')
    void sendMessage(content)
  }

  return (
    <main className="flex min-w-0 flex-[2] flex-col border-r border-outline">
      <div className="min-h-0 flex-1 overflow-y-auto p-4">
        {(!current || current.messages.length === 0) && (
          <div className="mt-16 text-center">
            <div className="text-lg font-semibold">你想绘制什么数学图像?</div>
            <div className="mt-2 text-sm text-on-surface-variant">
              例如:画 y = x² · 比较 y = sin(x) 和 y = cos(x)
            </div>
          </div>
        )}
        {current?.messages.map((m) => (
          <div key={m.id} className={`mb-3 flex ${m.role === 'user' ? 'justify-end' : 'justify-start'}`}>
            <div
              className={`max-w-[85%] rounded-lg px-3 py-2 text-sm ${
                m.role === 'user' ? 'bg-primary text-white' : 'bg-surface-container-high'
              }`}
            >
              <div className="whitespace-pre-wrap">{m.content}</div>
              {m.structuredResult && m.structuredResult.equations && m.structuredResult.equations.length > 0 && (
                <div className="mt-2 border-t border-outline/40 pt-2 text-xs">
                  {m.structuredResult.equations.map((e) => (
                    <div key={e.id} style={{ color: e.color }}>
                      <Math tex={e.label} />
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        ))}
        {sending && <div className="text-sm text-on-surface-variant">AI 正在思考...</div>}
        {error && (
          <div className="mb-3 rounded-md border border-red-300 bg-red-50 p-2 text-sm text-red-700">
            {error}
            <button className="ml-2 underline" onClick={clearError}>关闭</button>
          </div>
        )}
      </div>

      <div className="shrink-0 border-t border-outline p-3">
        <div className="mb-2 flex flex-wrap gap-1.5">
          {QUICK_PROMPTS.map((p) => (
            <button
              key={p}
              onClick={() => send(p)}
              className="rounded-full border border-outline px-2.5 py-1 text-xs hover:bg-surface-container"
            >
              {p}
            </button>
          ))}
        </div>
        <div className="flex gap-2">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                send()
              }
            }}
            placeholder="输入自然语言或公式,如:画一条经过原点、斜率为 2 的直线"
            rows={2}
            className="min-w-0 flex-1 resize-none rounded-md border border-outline bg-white p-2 text-sm"
          />
          <button
            onClick={() => send()}
            disabled={sending}
            className="rounded-md bg-primary px-4 text-white disabled:opacity-50"
          >
            发送
          </button>
        </div>
      </div>
    </main>
  )
}
