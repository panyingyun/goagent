# ch02：结构化输出

目标：让程序消费经过验证的数据，而不是从自然语言中猜字段。

## 运行与代码

```bash
go run ./ch02-structured-output "发布一个 Agent 服务"
go test ./ch02-structured-output ./pkg/structured -v
```

- [`main.go`](main.go)：调用 `structured.CreateTaskPlan`；
- [`main_test.go`](main_test.go)：验证空步骤等业务错误会被拒绝；
- [`pkg/structured/plan.go`](../pkg/structured/plan.go)：未知字段、重复 ID、步骤数量等校验；
- [`pkg/structured/generate.go`](../pkg/structured/generate.go)：JSON mode、`temperature=0`、校验失败有限次修复重试。

## 主线

1. 在支持时优先使用供应商的 `response_format=json_object`（或 JSON Schema）；
2. 无论模型声称什么，应用端仍要解码和业务校验；
3. 区分“JSON 语法错误”“Schema 错误”“业务不变量错误”；
4. 修复重试必须有次数上限（本课默认 3 次），并在最终错误中保留原始输出摘要；
5. 不用正则从任意长文本里盲目捞 JSON。

本课程的 `llm.Message` 和 `agent.ToolCall` 使用明确类型。工具参数最终由工具自身严格验证。

## 练习

定义 `TaskPlan`：标题非空、步骤 1–5 个、每步 ID 唯一。为缺字段、多字段、重复 ID、代码围栏 JSON 分别写测试；再写一个先返回非法文本、第二次返回合法 JSON 的 fake LLM，验证修复重试。
