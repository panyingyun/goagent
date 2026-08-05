# ch04：工具运行时

目标：工具能力越强，边界越明确。

## 运行与代码

```powershell
go run ./ch04-tool-runtime
go run ./ch04-tool-runtime -approve
go test ./ch04-tool-runtime -v
```

- [`main.go`](main.go)：对比未批准和批准后的写工具；
- [`main_test.go`](main_test.go)：验证审批与进程内幂等去重；
- [`pkg/agent/executor.go`](../pkg/agent/executor.go)：共享工具运行时。

每个 `agent.Tool` 显式声明：

- `SideEffect`：none/read/write/external；
- `Cacheable`：是否允许缓存；
- `Idempotent`：相同请求重复执行是否安全；
- `ParallelSafe`：是否可与其他调用并发；
- `RequiresApproval`：是否必须人工批准；
- `Timeout`：单次调用期限。

运行时位于 [`pkg/agent/executor.go`](../pkg/agent/executor.go)。写操作默认拒绝，除非提供 Approval；有副作用的工具不自动缓存，不与其他写工具并行。

示例运行时只在当前进程内保存幂等结果。跨进程重启时，幂等键必须存进持久化存储，或由支付、邮件、数据库等外部系统原生保证。

## 练习

增加 `send_email` 假工具：要求审批、不可缓存、使用 idempotency key。测试拒绝、批准、重复请求三种情况。
