# ch20：Harness 与 Graph 系统评估

目标：从“回答看起来不错”升级为对路径、状态、预算、安全和故障恢复的系统级评估。

## 运行

```powershell
go run ./ch20-system-evaluation
go test ./pkg/systemeval ./ch20-system-evaluation -v
```

- [`pkg/systemeval/eval.go`](../pkg/systemeval/eval.go)：案例、Outcome、路径/状态断言、预算指标、基线比较和故障注入；
- [`main.go`](main.go)：评估正常路由、高风险路由和节点故障。

## 评估层次

| 层次 | 主要断言 |
|---|---|
| 回答 | 内容是否满足任务 |
| 工具 | 名称、参数和副作用是否正确 |
| Graph | 路径、终态、循环次数和失败路由 |
| Harness | 能力、成本、步数、超时和人工接管 |
| Learning | 新经验是否改善目标案例且不产生回归 |

`WithFaults` 在指定节点注入确定性失败，使错误路径成为日常回归测试，而不是线上事故后才验证。

候选版本上线门禁至少要求：无新增安全违规、无已通过案例退化、任务成功率不降低；成本或步数增加必须有明确收益。
