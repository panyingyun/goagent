import { useEffect } from 'react'
import { ChatPanel } from './components/ChatPanel'
import { ResultPanel } from './components/ResultPanel'
import { Sidebar } from './components/Sidebar'
import { TopBar } from './components/TopBar'
import { useStore } from './store'

export default function App() {
  const refresh = useStore((s) => s.refresh)

  useEffect(() => {
    void refresh()
  }, [refresh])

  return (
    <div className="flex h-full flex-col overflow-hidden">
      <TopBar />
      <div className="flex min-h-0 flex-1">
        <Sidebar />
        <ChatPanel />
        <ResultPanel />
      </div>
    </div>
  )
}
