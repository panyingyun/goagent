import { useEffect } from 'react'
import { useStore } from './store'
import { TopBar } from './components/TopBar'
import { Sidebar } from './components/Sidebar'
import { ChatPanel } from './components/ChatPanel'
import { ResultPanel } from './components/ResultPanel'

export default function App() {
  const loadSessions = useStore((s) => s.loadSessions)
  const current = useStore((s) => s.current)

  useEffect(() => {
    void loadSessions()
  }, [loadSessions])

  return (
    <div className="flex h-full flex-col">
      <TopBar />
      <div className="flex min-h-0 flex-1">
        <Sidebar />
        <ChatPanel />
        {current ? <ResultPanel /> : (
          <div className="flex flex-1 items-center justify-center text-on-surface-variant">
            新建或选择一个会话开始
          </div>
        )}
      </div>
    </div>
  )
}
