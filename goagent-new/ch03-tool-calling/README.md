# ch03：原生 Tool Calling

目标：实现最小 Agent 闭环。

```text
用户消息 → LLM → tool_calls? → 执行工具 → tool 结果回填 → LLM → 最终回答
```

实现位于 [`pkg/agent/runner.go`](../pkg/agent/runner.go)。本章只解决循环正确性：

- 限制最大轮数；
- tool call ID 与结果严格对应；
- 未知工具、坏参数和工具错误作为可观察结果返回；
- 所有轮次共享调用方 context；
- 不依赖也不展示模型内部 Thought。

## 验收

使用 scripted fake LLM 完成“调用 calculator 后回答”的离线测试，并覆盖未知工具与超过轮数。

