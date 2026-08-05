import { FormEvent, useEffect, useMemo, useState } from "react";
import { createSession, deleteSession, getSession, getToken, getTrace, listSessions, sendChat, setToken } from "./api";
import { GraphCanvas } from "./GraphCanvas";
import type { Session, SessionSummary, TraceRecord } from "./types";

const examples = ["画出 y=x 和 y=x^2，并比较交点", "绘制 sin(x) 与 cos(x)，视口设为 -2*pi 到 2*pi", "画 exp(x) 和 log(x)，解释它们的关系"];

export default function App() {
  const [sessions, setSessions] = useState<SessionSummary[]>([]);
  const [session, setSession] = useState<Session | null>(null);
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [trace, setTrace] = useState<TraceRecord | null>(null);
  const [token, updateToken] = useState(getToken());

  async function refreshSessions(preferredId?: string) {
    const items = await listSessions();
    setSessions(items);
    const target = preferredId ?? session?.id ?? items[0]?.id;
    if (target) setSession(await getSession(target));
    else {
      const created = await createSession();
      setSession(created);
      setSessions(await listSessions());
    }
  }

  useEffect(() => { refreshSessions().catch((cause) => setError(cause.message)); }, []);

  async function submit(event?: FormEvent) {
    event?.preventDefault();
    const text = message.trim();
    if (!session || !text || busy) return;
    setBusy(true); setError(""); setTrace(null);
    try {
      const result = await sendChat(session.id, text);
      setSession(result.session); setMessage("");
      setSessions(await listSessions());
    } catch (cause) { setError(cause instanceof Error ? cause.message : "请求失败"); }
    finally { setBusy(false); }
  }

  async function addSession() {
    try { const created = await createSession(); setSession(created); setSessions(await listSessions()); setTrace(null); }
    catch (cause) { setError(cause instanceof Error ? cause.message : "创建失败"); }
  }

  async function removeSession(id: string) {
    if (!window.confirm("删除这个会话？该操作不可撤销。")) return;
    try {
      await deleteSession(id);
      const items = await listSessions();
      setSessions(items);
      if (session?.id === id) {
        if (items[0]) setSession(await getSession(items[0].id));
        else {
          const created = await createSession();
          setSession(created);
          setSessions(await listSessions());
        }
      }
    }
    catch (cause) { setError(cause instanceof Error ? cause.message : "删除失败"); }
  }

  async function openTrace(id: string) {
    try { setTrace(await getTrace(id)); }
    catch (cause) { setError(cause instanceof Error ? cause.message : "Trace 已过期"); }
  }

  const graph = session?.graphState;
  const eventSummary = useMemo(() => trace?.events.map((item) => `${item.kind}${item.tool ? ` · ${item.tool}` : ""} · ${item.status}`) ?? [], [trace]);

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="brand"><span className="brand-mark">∫</span><div><strong>MathAI</strong><small>函数实验室</small></div></div>
        <div className="status"><span className="status-dot" /> 工具调用模式</div>
      </header>

      <aside className="sidebar">
        <button className="primary-button" onClick={addSession}>新建探索 →</button>
        <p className="section-label">最近会话</p>
        <div className="session-list">
          {sessions.map((item) => <button key={item.id} className={`session-item ${session?.id === item.id ? "active" : ""}`} onClick={() => refreshSessions(item.id)}>
            <span><strong>{item.title}</strong><small>{new Date(item.updatedAt).toLocaleString("zh-CN", { month: "numeric", day: "numeric", hour: "2-digit", minute: "2-digit" })}</small></span>
            <span className="delete" role="button" onClick={(event) => { event.stopPropagation(); removeSession(item.id); }}>×</span>
          </button>)}
        </div>
        <label className="token-field"><span>API Token（可选）</span><input type="password" value={token} placeholder="与服务端配置一致" onChange={(event) => { updateToken(event.target.value); setToken(event.target.value); }} onBlur={() => refreshSessions().catch((cause) => setError(cause.message))} /></label>
      </aside>

      <main className="workspace">
        <section className="chat-panel">
          <div className="panel-heading"><div><span className="eyebrow">对话</span><h1>{session?.title ?? "载入中…"}</h1></div></div>
          <div className="messages">
            {!session?.messages?.length && (
              <div className="welcome">
                <span className="welcome-badge">开源 · Agent 工具模式</span>
                <span className="mark">ƒ(x)</span>
                <h2>告诉我你想探索的 <span className="grad">函数</span></h2>
                <p>MathAI 会通过受控工具更新图形，并解释相关数学特征。</p>
                <div className="example-grid">{examples.map((example) => <button key={example} onClick={() => setMessage(example)}>{example}</button>)}</div>
              </div>
            )}
            {session?.messages?.map((item) => <article key={item.id} className={`message ${item.role}`}><div className="avatar">{item.role === "user" ? "你" : "∫"}</div><div><p>{item.content}</p>{item.traceId && <button className="trace-link" onClick={() => openTrace(item.traceId!)}>查看执行轨迹 →</button>}</div></article>)}
            {busy && <article className="message assistant"><div className="avatar">∫</div><div className="thinking"><i /><i /><i /></div></article>}
          </div>
          {error && <div className="error-banner">{error}<button onClick={() => setError("")}>×</button></div>}
          <form className="composer" onSubmit={submit}><textarea value={message} maxLength={4000} placeholder="例如：画 y=sin(x)，并标出它的周期…" onChange={(event) => setMessage(event.target.value)} onKeyDown={(event) => { if (event.key === "Enter" && !event.shiftKey) { event.preventDefault(); submit(); } }} /><button disabled={!message.trim() || busy} aria-label="发送">↑</button></form>
        </section>

        <section className="graph-panel">
          <div className="panel-heading"><div><span className="eyebrow">画布</span><h2>函数图像</h2></div>{graph && <span className="viewport-badge">x [{graph.viewport.xMin}, {graph.viewport.xMax}] · y [{graph.viewport.yMin}, {graph.viewport.yMax}]</span>}</div>
          <div className="canvas-wrap">{graph ? <GraphCanvas graph={graph} /> : <div className="canvas-placeholder">正在准备画布…</div>}</div>
          {graph?.settings.showLegend && <div className="legend">{(graph.equations ?? []).map((equation) => <span className={equation.visible ? "" : "muted"} key={equation.id}><i style={{ background: equation.color }} />{equation.label}</span>)}</div>}
        </section>

        <aside className="inspector">
          <div className="panel-heading"><div><span className="eyebrow">状态</span><h2>图形状态</h2></div><span className="version">v{session?.version ?? 0}</span></div>
          <section><h3>函数</h3>{graph?.equations?.length ? graph.equations.map((equation) => <div className="equation-card" key={equation.id}><i style={{ background: equation.color }} /><div><strong>{equation.label}</strong><code>{equation.expression}</code></div><span>{equation.visible ? "显示" : "隐藏"}</span></div>) : <p className="empty-copy">尚未添加函数</p>}</section>
          <section><h3>渲染</h3><dl><div><dt>采样点</dt><dd>{graph?.settings.sampleCount ?? "—"}</dd></div><div><dt>网格</dt><dd>{graph?.settings.showGrid ? "开启" : "关闭"}</dd></div><div><dt>坐标轴</dt><dd>{graph?.settings.showAxes ? "开启" : "关闭"}</dd></div></dl></section>
          {graph?.analysis && <section><h3>Agent 分析</h3><p className="analysis-copy">{graph.analysis.summary}</p><ul>{(graph.analysis.highlights ?? []).map((item) => <li key={item}>{item}</li>)}</ul></section>}
          {trace && <section className="trace-card"><h3>执行轨迹</h3><small>{trace.id}</small>{eventSummary.map((item, index) => <p key={`${item}-${index}`}>{item}</p>)}</section>}
        </aside>
      </main>
    </div>
  );
}
