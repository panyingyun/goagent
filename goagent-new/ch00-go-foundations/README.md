# ch00：Go 前置能力

本章不讲 Agent 名词，先补齐后续会反复使用的 Go 能力。代码可以直接运行，测试不需要 API Key。

## 运行

```powershell
cd goagent-new
go run ./ch00-go-foundations
go test ./ch00-go-foundations -v
```

## 代码地图

| 文件 | 内容 |
|---|---|
| [`main.go`](main.go) | 严格 JSON 与有限并发演示 |
| [`foundations.go`](foundations.go) | `DecodeJSONStrict`、受 context 控制的 HTTP、worker pool |
| [`foundations_test.go`](foundations_test.go) | 超时、响应体限制、未知字段、并发上限、结果顺序测试 |

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

1. 把 `TestFetchTextUsesContext` 的 context 超时改成 200ms，观察结果变化；
2. 删除 `DisallowUnknownFields`，验证带 `admin` 字段的输入为什么会被静默接受；
3. 给 `RunWorkers` 增加一个返回错误的任务，验证其他 worker 会被取消；
4. 为 `FetchText` 增加非 2xx、客户端自身 Timeout 和调用方 context 同时生效的测试。
