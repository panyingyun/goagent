# ch11：可观测性与成本

目标：能回答一次任务为什么慢、贵、失败或越权。

## 运行与代码

```powershell
go run ./ch11-observability
go test ./ch11-observability -v
```

- [`main.go`](main.go)：LLM、工具、失败 span 和 usage 事件；
- [`main_test.go`](main_test.go)：父子关系、token 和错误测试；
- [`pkg/trace/trace.go`](../pkg/trace/trace.go)：基于 context 的并发安全 recorder。

每个 run 至少记录：

- trace/run ID 与父子关系；
- 模型、工具名、状态、错误类型和耗时；
- 输入/输出 token 与估算成本；
- 工具审批结果和幂等键；
- prompt/version、数据集版本和部署版本；
- 脱敏后的参数摘要，而不是完整秘密。

[`pkg/trace`](../pkg/trace/) 使用 context 传播 span，避免全局“活动栈”在并发请求间串线。
