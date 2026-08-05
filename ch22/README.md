# 课时22:迷你 LangChain——Go 复刻 LangChain 核心抽象

> 目标:用我们手写过的组件,对照理解 LangChain 到底帮你做了什么。

## 组件对照表

| LangChain 概念 | 本课实现 | 我们之前的代码 |
|---|---|---|
| Chain | `Chain` 接口 | 任意"输入->输出"函数 |
| LLMChain | 系统提示词 + LLM | ch16 的 chat |
| SequentialChain | 链式串联 | 流水线思维 |
| Memory | 消息记忆 | ch06 的 History |
| Tool / AgentExecutor | 工具 + ReAct 循环 | ch03 的 Registry + runReAct |

## 运行

```powershell
cd E:\codex\project001\goagent\ch22
go run . "把这句话翻译成英文: 你好,世界"
```

会看到:翻译链输出 -> 总结链输出 -> AgentExecutor 计算输出。

## 重要认知

- LangChain 的价值是**抽象和生态**(统一接口、海量集成),不是魔法;
- 我们已经亲手实现了这些抽象的最小版本——所以你能看懂 LangChain 在做什么;
- Go 生态里官方的 LangChain 支持较弱,社区版是 `langchaingo`;
  本课证明了用 Go 手写这套抽象完全可行。

## 练习

1. 给 `SequentialChain` 加"任一条失败就回滚"的逻辑;
2. 给 `Memory` 加容量上限(超了丢最旧);
3. 用 `AgentExecutor` 组合两个工具(calculator + 时间),问一个复合问题。
