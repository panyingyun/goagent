# Go AI Agent 新课程

这是一套从零开始、以工程正确性为主线的 Go Agent 课程。课程不把某个框架、模型名或供应商价格当成长期知识；核心目标是学会构造一个可测试、可恢复、可审计、具备明确权限边界的 Agent。

与旧课程不同，本课程采用一个贯穿项目：**Workspace Agent**。它从最小 LLM 调用开始，逐章增加结构化输出、工具调用、权限审批、记忆、流式响应、评估、持久化工作流、安全、检索、追踪和 HTTP 部署。最终项目使用的正是前面逐步构建的共享组件，而不是另写一个无关示例。

> 课程代码默认使用 OpenAI-compatible Chat Completions 协议，但不硬编码具体模型。运行联网示例前，请根据当前供应商官方文档设置 `LLM_MODEL`。

## 学习目标

完成核心课程后，你应该能：

1. 解释 LLM、工具、Agent 循环和工作流之间的边界；
2. 使用原生 tool calls 构建 Agent，而不是依赖脆弱的文本协议；
3. 按副作用对工具分类，并实现超时、取消、审批、幂等和并发控制；
4. 区分会话上下文、摘要、用户偏好和可检索长期记忆；
5. 用测试集、工具轨迹和任务成功率持续评估 Agent；
6. 实现可恢复工作流，并避免恢复时重复执行有副作用的动作；
7. 识别提示注入、间接注入、SSRF、路径逃逸和数据外泄风险；
8. 用 trace、token、延迟、错误率和成本预算观察 Agent；
9. 部署一个带认证、输入限制、超时、限流和优雅关闭的真实 Agent 服务。

## 推荐路径

| 课时 | 主题 | 产出 |
|---|---|---|
| [ch00](ch00-go-foundations/) | Go 前置能力 | context、HTTP、JSON、interface、并发与测试清单 |
| [ch01](ch01-llm-api/) | LLM API | 可配置客户端、错误分类、usage |
| [ch02](ch02-structured-output/) | 结构化输出 | 严格解码、字段校验、失败策略 |
| [ch03](ch03-tool-calling/) | 原生 Tool Calling | 最小但完整的 Agent 循环 |
| [ch04](ch04-tool-runtime/) | 工具运行时 | 权限、审批、幂等、缓存、并行、超时 |
| [ch05](ch05-memory/) | 状态与记忆 | 窗口、摘要、偏好和删除 |
| [ch06](ch06-streaming/) | 流式与取消 | SSE、客户端断开、背压 |
| [ch07](ch07-evaluation/) | 评估驱动开发 | 回归集、工具轨迹、任务成功率 |
| [ch08](ch08-workflow/) | 工作流 | 原子检查点、恢复、HITL |
| [ch09](ch09-security/) | Agent 安全 | 威胁模型和安全工具边界 |
| [ch10](ch10-rag/) | 可选 RAG | 关键词基线、embedding 接口、Recall@K |
| [ch11](ch11-observability/) | 可观测与成本 | trace、usage、延迟和预算 |
| [ch12](ch12-deployment/) | 部署 | 认证、限流、超时、健康检查 |
| [ch13](ch13-capstone/) | 综合项目 | 可运行的 Workspace Agent |
| [ch14](mathai/) | Web 综合案例 | MathAI 前后端、Graph 工具、trace 与 Docker |
| [ch15](ch15-harness-engineering/) | Harness Engineering | 能力清单、任务预算、无进展检测与统一验收 |
| [ch16](ch16-context-engineering/) | Context Engineering | 可信来源、相关性选择与上下文预算 |
| [ch17](ch17-agent-learning/) | Agent Learning Loop | 经验、证据、回归门禁、晋级与检索 |
| [ch18](ch18-graph-engineering/) | Graph Engineering | 节点、边、条件、循环和拓扑验证 |
| [ch19](ch19-durable-graph/) | Durable Graph | Checkpoint、Event Log、恢复和状态迁移 |
| [ch20](ch20-system-evaluation/) | 系统评估 | 路径、状态、预算、故障注入和版本比较 |
| [ch21](ch21-repopilot/) | 高级综合项目 | 可审批、可恢复、可学习的 RepoPilot |

附录位于 [appendices](appendices/)：文本 ReAct、MCP、Skill、多 Agent，以及框架概念映射。附录不是核心主线的前置条件。

## 运行

```powershell
cd goagent-new
Copy-Item .env.example .env
# 将 .env 中的值设置到当前终端，或直接设置环境变量：
$env:LLM_API_KEY = "..."
$env:LLM_BASE_URL = "https://api.deepseek.com"
$env:LLM_MODEL = "供应商当前支持的模型名"

go test ./...
go run ./cmd/workspace-agent
```

启动 HTTP 服务：

```powershell
$env:AGENT_API_TOKEN = "local-dev-token"
go run ./cmd/workspace-server
```

然后请求：

```powershell
Invoke-RestMethod -Method Post -Uri http://localhost:8080/v1/chat `
  -Headers @{ Authorization = "Bearer local-dev-token" } `
  -ContentType application/json `
  -Body '{"message":"列出工作区里的文档"}'
```

## 目录结构

```text
goagent-new/
├── ch00-...ch13/       # 主线课程讲义、验收标准与练习
├── mathai/             # ch14 Web 综合案例
├── ch15-...ch21/       # Harness、学习闭环、执行图与 RepoPilot
├── appendices/         # 非主线生态知识
├── cmd/
│   ├── workspace-agent # 贯穿项目 CLI
│   └── workspace-server# 真实 Agent HTTP 服务
├── pkg/
│   ├── llm             # 可替换 LLM 客户端
│   ├── structured      # 严格结构化输出解码与校验
│   ├── agent           # Agent 循环与工具运行时
│   ├── harness         # 任务能力、预算、Worker 与最终验收
│   ├── contextkit      # 可信来源、相关性和上下文预算装配
│   ├── memory          # 用户偏好与长期记忆
│   ├── learning        # 经验候选、验证、晋级、退役与检索
│   ├── retrieval       # RAG 基线与评估
│   ├── workflow        # 线性可恢复工作流与 HITL
│   ├── graph           # 节点、条件边、循环和执行图引擎
│   ├── graphstore      # Graph Checkpoint、Event Log 与状态迁移
│   ├── stream          # SSE 流式响应、取消与背压
│   ├── eval            # 基础任务与工具轨迹评估
│   ├── systemeval      # Graph/Harness 路径、状态、预算与回归评估
│   ├── trace           # Agent 追踪、事件、耗时与 usage
│   └── workspace       # Workspace Agent 的安全工具和应用组装
└── testdata/           # 课程文档与评估集
```

## 教学原则

- 每个外部输入都不可信，包括用户消息、网页、检索片段、MCP 工具说明和模型输出；
- 默认使用原生 tool calls；文本 ReAct 仅用于理解历史和协议权衡；
- 读工具可以安全并行，写工具默认串行并要求审批；
- 缓存、重试和恢复必须服从副作用与幂等语义；
- 提示词不是安全边界，真正的边界必须在代码、权限和隔离层；
- 没有评估集和 trace 的 Agent，不应声称“已经优化”；
- 示例可以最小，但不能用错误的安全模式换取代码短小。
