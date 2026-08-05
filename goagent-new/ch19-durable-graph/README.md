# ch19：Durable Graph Runtime

目标：让执行图在崩溃后可以恢复，并清楚区分 checkpoint、事件日志和外部副作用幂等。

## 运行

```powershell
go run ./ch19-durable-graph
go run ./ch19-durable-graph -resume
go test ./pkg/graphstore ./ch19-durable-graph -v
```

- [`pkg/graphstore/store.go`](../pkg/graphstore/store.go)：原子 Checkpoint、JSONL Event Log、读取和逐版本迁移；
- [`main.go`](main.go)：`prepare → publish` 持久化执行图；
- [`pkg/graphstore/store_test.go`](../pkg/graphstore/store_test.go)：模拟“外部写成功但节点返回失败”并恢复。

## 三种不同保证

1. **Checkpoint** 保存控制流状态，用于从当前节点恢复；
2. **Event Log** 记录发生过什么，用于审计、评估和 Replay 分析；
3. **外部幂等** 防止 checkpoint 落盘前崩溃导致重复副作用。

本地 `CompletedKeys` 不能单独提供 exactly-once。恢复测试中的假外部服务自己保存幂等键，所以第二次调用不会重复产生远端效果。

Graph State 带有 `graph_version`。定义升级时必须提供连续迁移函数；不能拿旧状态直接交给新图运行。

## 生产扩展

文件 Store 适合单进程教学。多 Worker 需要数据库条件更新、租约、心跳、事件序号和事务 outbox；不要让多个容器共享一个 JSON 文件。
