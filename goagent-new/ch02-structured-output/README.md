# ch02：结构化输出

目标：让程序消费经过验证的数据，而不是从自然语言中猜字段。

## 运行与代码

```powershell
go run ./ch02-structured-output "发布一个 Agent 服务"
go test ./ch02-structured-output -v
```

- [`main.go`](main.go)：请求计划并执行严格解码；
- [`main_test.go`](main_test.go)：验证空步骤等业务错误会被拒绝；
- [`pkg/structured/plan.go`](../pkg/structured/plan.go)：未知字段、重复 ID、步骤数量等校验。

## 主线

1. 在支持时优先使用供应商的 JSON Schema/structured output；
2. 无论模型声称什么，应用端仍要解码和业务校验；
3. 区分“JSON 语法错误”“Schema 错误”“业务不变量错误”；
4. 修复重试必须有次数上限，并保留原始错误用于 trace；
5. 不用正则从任意长文本里盲目捞 JSON。

本课程的 `llm.Message` 和 `agent.ToolCall` 使用明确类型。工具参数最终由工具自身严格验证。

## 练习

定义 `TaskPlan`：标题非空、步骤 1–5 个、每步 ID 唯一。为缺字段、多字段、重复 ID、代码围栏 JSON 分别写测试。
