# ch13：综合项目——Workspace Agent

Workspace Agent 能安全地：

- 列出 `testdata/workspace` 中的文档；
- 读取指定文本文件；
- 搜索课程资料并返回来源；
- 使用 calculator；
- 经用户批准后写入笔记。

它复用课程共享的 LLM 客户端、原生 tool calls、工具策略、审批、trace 和评估结构。CLI 位于 [`cmd/workspace-agent`](../cmd/workspace-agent/)，HTTP 服务位于 [`cmd/workspace-server`](../cmd/workspace-server/)。

## 验收场景

1. “列出文档”只调用只读工具；
2. “计算 12*(8+4)”调用 calculator；
3. “把结果写进 result.txt”必须提示审批；
4. 路径穿越、绝对路径和符号链接被拒绝；
5. 超过最大轮数能明确失败；
6. 客户端取消后，模型和工具都停止；
7. trace 能看到 LLM、工具、审批和耗时；
8. 离线 fake LLM 评估无需 API Key。

## 下一步

在完成基线评估后，再选择性加入 MCP、真实 embedding、流式 Web UI 或多 Agent。不要为了使用名词而增加架构。

