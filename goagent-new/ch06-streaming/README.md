# ch06：流式与取消

目标：流式不是逐字打印，而是一条从浏览器断开一直传播到模型请求和工具进程的取消链。

## 运行与代码

```powershell
go run ./ch06-streaming
Invoke-WebRequest "http://localhost:8080/stream?q=hello"
go test ./ch06-streaming -v
```

- [`main.go`](main.go)：可取消的事件生产者和 SSE 服务；
- [`main_test.go`](main_test.go)：读取到 done，而不是只读第一个网络分片；
- [`pkg/stream/sse.go`](../pkg/stream/sse.go)：共享 SSE relay。

## 必学

- 上游模型流与下游 SSE/ReadableStream 的转换；
- 客户端断开后取消 LLM 和工具；
- 首 token 超时与分片空闲超时；
- 慢客户端背压；
- 流中途失败的显式 error/done 事件；
- 完整内容接收成功后再提交会话状态。

[`pkg/stream`](../pkg/stream/) 提供可取消、逐帧 flush 的 SSE relay。部署章节的 HTTP 服务保留非流式端点用于稳定测试；接入供应商流式 API 是本章练习。测试必须读到 EOF 或明确的 `done` 事件，禁止用一次 `Read` 假设拿到完整响应。
