# ch12：部署真实 Agent

目标：部署的是前面构建的 Agent，而不是返回“已收到”的占位接口。

[`cmd/workspace-server`](../cmd/workspace-server/) 包含：

- Bearer token 认证；
- 请求方法和 Content-Type 检查；
- 请求体大小限制和严格 JSON；
- server read/write/idle 超时；
- 每个请求的 context timeout；
- 有界、会清理的限流桶；
- `/livez` 与 `/readyz`，未就绪返回 503；
- 优雅关闭；
- 真实调用 `agent.Runner`。

生产还应在可信反向代理后终止 TLS，并使用集中式身份、密钥、日志和限流服务。

