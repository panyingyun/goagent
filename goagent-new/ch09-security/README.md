# ch09：Agent 安全

目标：建立威胁模型，而不是背提示注入关键词。

## 信任边界

- 用户输入不可信；
- 网页、邮件和 RAG 文档属于不可信数据，不是指令；
- 模型输出不可信；
- 工具说明、MCP Server 和 Skill 也可能不可信；
- system prompt 不能替代授权和隔离。

## 必做控制

- URL allowlist、解析后 IP 检查、重定向复核，防 SSRF；
- 文件访问使用固定根目录、拒绝符号链接逃逸；
- 不把模型字符串传给 Shell；命令和参数使用固定 argv；
- 写文件、发消息、支付、删除等动作要求审批；
- 输出和 trace 做敏感信息脱敏；
- 为每个主体执行认证、授权、审计和配额。

[`pkg/agent/safe_tools.go`](../pkg/agent/safe_tools.go) 展示不经过 Shell 的安全工具模式。

