# ch00：Go 前置能力

本章不讲 Agent 名词，先补齐后续会反复使用的 Go 能力。

## 必会清单

- `context.Context`：超时、取消和请求作用域；
- `net/http`：客户端、服务端、状态码、Header、Body 限制；
- `encoding/json`：严格解码、未知字段、数字类型；
- interface 与依赖注入：生产实现和 fake 实现共用契约；
- goroutine、channel、`sync.WaitGroup`：只并行可以安全并行的工作；
- table-driven test、`httptest` 和 race detector；
- 文件权限、临时文件、原子替换和错误包装。

## 验收

你应能解释为什么所有外部调用都接收 `context.Context`，为什么 HTTP 客户端和服务端都需要超时，以及为什么不能把未经校验的模型输出直接交给 Shell。

## 练习

1. 给一个 `httptest.Server` 设置 200ms 延迟，用 50ms context 验证取消；
2. 用 `json.Decoder.DisallowUnknownFields` 拒绝多余字段；
3. 写一个最多 3 个 worker 的并发函数，并在取消后停止派发任务。

