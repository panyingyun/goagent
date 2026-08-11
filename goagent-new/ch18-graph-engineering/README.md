# ch18：Graph Engineering

目标：把 Agent 的控制流表示为可验证、可测试的执行图，而不是隐藏在提示词和 `if` 语句中。

## 运行

```bash
go run ./ch18-graph-engineering
go test ./pkg/graph ./ch18-graph-engineering -v
```

- [`pkg/graph/graph.go`](../pkg/graph/graph.go)：Definition、Node、Edge、条件路由、循环保护和逐节点执行；
- [`main.go`](main.go)：支持请求按风险路由到自动回答或人工升级。

## 图不等于多 Agent

节点可以是确定性代码、工具调用、LLM 调用、人工审批或子图。只有权限或上下文确实需要隔离时，节点才需要成为独立 Agent。

本章运行时具备：

- 构建期检查入口、节点、边、终点和不可达节点；
- 条件边和显式循环；
- 同时命中多条边时拒绝歧义路由；
- 最大节点步数，防止无限循环；
- `Step` API，供外层 Harness 和持久化运行时逐步驱动；
- 节点路径、输出和事件。

并行 Fan-out/Fan-in 不应通过共享可变 map 草率实现。练习是在引入 typed reducer 和确定性 merge 后，再增加并行分支。
