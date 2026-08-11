# ch15：Harness Engineering

目标：把 Agent 周围的运行环境作为一等工程对象，而不是继续堆提示词。

## 运行

```bash
go run ./ch15-harness-engineering
go test ./pkg/harness ./ch15-harness-engineering -v
```

- [`pkg/harness/harness.go`](../pkg/harness/harness.go)：任务、能力清单、预算、Worker、Verifier 与观察事件；
- [`main.go`](main.go)：`inspect → plan → deliver` 的确定性教学案例；
- [`main_test.go`](main_test.go)：验证最终产物和 Harness 状态。

## 核心边界

Harness 不等于模型，也不等于 Agent 循环。它负责：

- 声明任务目标和所需能力；
- 校验环境是否提供这些能力；
- 管理步数、失败、时间和成本预算；
- 检测连续无进展；
- 统一验证最终产物；
- 输出可供 trace 和评估消费的事件。

`pkg/agent.Runner` 仍负责 LLM 与工具之间的循环；`pkg/harness.Runner` 位于更外层，可以包装 LLM Agent、执行图或确定性 Worker。

## 验收

1. 缺少能力时，任务在 Worker 执行前失败；
2. 超过步数、成本或失败预算时明确终止；
3. 连续三步无进展时停止；
4. Worker 声称完成后仍必须通过 Verifier；
5. 取消和超时能够传递到 Worker。
