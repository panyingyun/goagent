# 课时23:迷你 LangGraph——从"循环"到"图"

> 目标:理解 LangGraph 的核心抽象——把 Agent 逻辑画成一张图,而不是一个循环。

## 核心概念与对照

| LangGraph | 本课实现 | 说明 |
|---|---|---|
| Node | `AddNode` | 一个处理单元(plan/execute/review) |
| Edge | `AddEdge` / `__next__` | 决定执行流向,`__next__` 支持循环 |
| State | `map[string]any` | 节点间共享的全局状态 |
| Checkpoint | 保存到 JSON 文件 | 执行结果可恢复 |

## 运行

```powershell
cd E:\codex\project001\goagent\ch23
go run . "写一段产品介绍,突出 AI 能力"
```

会看到:plan -> execute -> review,审查不通过自动回到 execute,最多返工 2 次;
状态存到 `checkpoint.json`。

离线测试:

```powershell
go test ./... -v
```

## 为什么"图"比"循环"强

循环是"一条道走到黑",图可以:分支、合并、回退、暂停恢复、可视化。
真实 LangGraph 还有持久化存储、并行分支、人工介入(HITL),
本课是最小但完整的内核。

## 练习

1. 加一个"并行"节点:两个子任务同时执行(参考 ch13 的并发);
2. 给 review 节点加"记录失败原因到 state";
3. 用 `checkpoint.json` 实现"中断后从上次节点恢复"(参考 ch14 的 Resume)。
