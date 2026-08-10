# A4：框架概念映射

| 本课程概念 | 常见框架名称 |
|---|---|
| `llm.Client` | Chat Model / Runnable |
| `agent.Tool` | Tool / StructuredTool |
| `agent.Runner` | Agent Executor / Tool Loop |
| `workflow.Workflow` | State Graph / Durable Workflow |
| `memory.Store` | Checkpointer / Memory Store |
| `trace.Recorder` | Run Tree / Trace / Span |
| `eval` 数据集 | Dataset / Evaluator |

框架的价值主要是成熟实现、集成和生态，而不是创造新的 Agent 基本原理。先理解左侧，再按项目需求选择右侧。

