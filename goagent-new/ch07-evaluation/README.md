# ch07：评估驱动开发

目标：在继续增加能力之前，先建立不会悄悄退化的反馈系统。

## 运行与代码

```bash
go run ./ch07-evaluation
go test ./ch07-evaluation -v
```

- [`main.go`](main.go)：离线数据集、工具轨迹和任务状态评估；
- [`main_test.go`](main_test.go)：把通过率作为回归门槛；
- [`pkg/eval/eval.go`](../pkg/eval/eval.go)：共享确定性评估器。

## 三层评估

1. 单元测试：Schema、工具、权限、状态转换；
2. 轨迹评估：调用了哪个工具、参数是否正确、是否多调用；
3. 任务评估：最终状态是否达到目标，而不只检查回答包含某句话。

[`pkg/eval`](../pkg/eval/) 提供确定性轨迹匹配。LLM-as-Judge 只适用于主观质量，并需要校准、盲测和人工抽检。

## 练习

为 calculator、文档列表和写笔记建立至少 10 条数据集。记录通过率、平均工具调用数和 P95 延迟基线。
