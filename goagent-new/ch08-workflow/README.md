# ch08：工作流、恢复与 HITL

目标：把长任务变成可恢复的状态机。

## 运行与代码

```powershell
go run ./ch08-workflow -approve
go run ./ch08-workflow -resume -approve
go test ./ch08-workflow -v
```

- [`main.go`](main.go)：计划、审批、写步骤和恢复入口；
- [`main_test.go`](main_test.go)：未批准暂停、批准后恢复；
- [`pkg/workflow/workflow.go`](../pkg/workflow/workflow.go)：版本化状态与检查点实现。

检查点采用“写临时文件 → flush/close → 同目录替换”的尽力安全写入，并包含版本、当前节点和已完成的幂等键；不同操作系统对文件替换原子性的保证不同，生产系统应使用具备事务或条件写能力的存储。

恢复不等于重新运行所有未看到结果的代码。仅靠 checkpoint 无法保证外部副作用“恰好一次”：进程可能在外部写成功、checkpoint 落盘前崩溃。外部写操作必须让目标系统持久化幂等键并可查询既有结果，或者恢复后进入人工确认状态。

## 练习

构造 `plan → gather → approve → write → done` 工作流。让假的外部写服务按 idempotency key 去重，然后在 write 成功但 checkpoint 写入前模拟崩溃，验证恢复后的第二次请求由外部服务返回原结果而不会重复产生副作用。
