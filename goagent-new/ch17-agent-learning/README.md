# ch17：Agent Learning Loop

目标：让 Agent 从经过验证的经验中改进，而不是把一次模型反思直接写进永久记忆。

## 运行

```bash
go run ./ch17-agent-learning
go test ./pkg/learning ./ch17-agent-learning -v
```

- [`pkg/learning/learning.go`](../pkg/learning/learning.go)：Experience、生命周期、持久化、回归门禁和检索；
- [`main.go`](main.go)：把两条 trace 支撑的端口冲突经验升级为可检索策略。

## 学习闭环

```text
task → trace/feedback → candidate → regression verify → verified → promote → retrieve
                                       ↓                         ↓
                                    rejected               human approval
```

只有 `promoted` 经验会进入后续上下文。晋级至少检查：

- 独立证据数量；
- 置信度；
- 固定回归集通过率；
- 是否出现旧任务退化；
- 高风险规则在 promotion 阶段是否经过人工批准。

经验具有 `scope`、`expires_at`、证据和状态。用户或项目范围的经验不会自动污染其他范围；过期和 retired 经验不会被检索。

## 不要这样做

- 保存全部聊天记录并称为学习；
- 让模型根据一次失败直接改 system prompt；
- 检索 candidate 或失败经验；
- 用回答文风代替任务成功率；
- 在没有回归集时自动生成和安装 Skill。
