# 附录示例：完整 Agent（对标旧课 ch04）

> 薄入口：逻辑在 `pkg/`，本目录只负责组装与演示。只用原生 tool calling。

## 框架映射

| 能力 | 实现 |
|---|---|
| Agent 循环 | `pkg/agent.Runner` + `Executor` |
| 工具 | `agent.NewCalculatorTool` / `NewGetTimeTool` |
| LLM | `pkg/llm.HTTPClient` |
| 观测 | `pkg/trace.Recorder` 作为 Observer |

## 运行

在 `goagent-new` 根目录：

```powershell
$env:LLM_API_KEY = "..."
$env:LLM_BASE_URL = "https://api.deepseek.com"
$env:LLM_MODEL = "供应商当前支持的模型名"

go run ./appendices/complete-agent "12 * 8 + 4 等于多少?"
go test ./appendices/complete-agent ./pkg/agent -v
```
