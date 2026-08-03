# 课时14:Agent 工作流与任务规划

> 目标:把"单循环 Agent"升级成"可拆解、可暂停、可恢复的多步骤工作流"。

## 本课时学什么

1. **任务规划**:让 LLM 把大目标拆成 3~5 个带子问题的步骤(JSON 输出);
2. **状态机**:每个步骤有 pending/done/failed 状态,工作流按序推进;
3. **检查点(Checkpoint)**:每一步都存盘,进程崩溃或中断后 `-resume` 接着跑。

这就是真实工作流引擎(Temporal、LangGraph)的最小形态:
**节点(步骤) + 状态 + 持久化**。

## 运行

```powershell
cd E:\codex\project001\goagent\ch14
go run . -plan "帮我准备一次新员工入职介绍"
go run . -run      # 执行全部步骤
go run . -resume   # 模拟中断后续跑(可手动改 checkpoint.json 里的状态)
```

离线测试:

```powershell
go test ./... -v
```

## 练习

1. 执行到一半,手动把 checkpoint.json 里一个步骤改成 failed,再 `-resume`;
2. 给 `Plan` 加一个"步骤数量"参数(3 步 / 5 步);
3. 想一想:如果某一步失败,重试策略应该是什么?
   (提示:固定次数重试 -> 人工介入,参考 ch10 的 Retry。)

## 小结

工作流把"Agent 一次性对话"变成"可审计的工程过程"。课时 23 的
LangGraph 图编排,本质上就是本课状态机的泛化。
