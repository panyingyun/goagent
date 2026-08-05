# ch16：Context 与 Environment Engineering

目标：在固定预算内装配可信、相关、可追溯的上下文，而不是把所有历史和文件塞给模型。

## 运行

```powershell
go run ./ch16-context-engineering
go test ./pkg/contextkit ./ch16-context-engineering -v
```

- [`pkg/contextkit/context.go`](../pkg/contextkit/context.go)：来源、可信等级、相关性排序、预算和安全 framing；
- [`main.go`](main.go)：从规则、项目文档和旧日志中选择当前任务需要的上下文。

## 与 ch05 的区别

ch05 解决“保存什么长期偏好”；本章解决“本次模型调用应该加载什么”。上下文来源可以是系统规则、用户记忆、项目文件、检索结果、工具输出或历史摘要。

Assembler 的原则：

1. 必需规则先进入，装不下就失败，不静默裁剪；
2. 其余来源按任务相关性和优先级选择；
3. 每个来源保留 ID、类型和可信等级；
4. 外部文档和工具输出默认不可信；
5. 内容使用 JSON 字符串封装，降低伪造边界的歧义；
6. Framing 不是安全沙箱，真正权限仍由工具层执行。

## 练习

增加一个按需 Loader：只有候选来源被选中后才读取完整文件；记录每种来源的估算 token、实际 token 和命中率。
