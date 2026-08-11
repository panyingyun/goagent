# ch05：状态与记忆

目标：不再把所有历史消息统称为“记忆”。

## 运行与代码

```bash
go run ./ch05-memory set drink coffee
go run ./ch05-memory get drink
go run ./ch05-memory list
go run ./ch05-memory delete drink
go test ./ch05-memory -v
```

- [`main.go`](main.go)：偏好新增、读取、列出和删除命令；
- [`main_test.go`](main_test.go)：持久化生命周期测试；
- [`pkg/memory/store.go`](../pkg/memory/store.go)：带来源和时间的偏好存储。

| 类型 | 用途 | 生命周期 |
|---|---|---|
| 会话窗口 | 当前推理上下文 | 短期 |
| 对话摘要 | 压缩旧对话 | 会话级 |
| 用户偏好 | 稳定、明确同意保存的信息 | 跨会话 |
| 事件记录 | 审计和精确回放 | 按保留策略 |
| 可检索知识 | 文档和事实 | 独立索引 |

偏好存储位于 [`pkg/memory`](../pkg/memory/)，支持列出、覆盖和删除。任何长期记忆都应有来源、时间和删除能力。

## 练习

实现“用户明确说记住时才保存”的策略，并测试“不要记住这句话”和“删除我的全部偏好”。
