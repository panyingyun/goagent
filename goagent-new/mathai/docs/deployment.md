# 部署指南

## 配置

| 环境变量 | 必填 | 默认值 | 说明 |
|---|---:|---|---|
| `DEEPSEEK_API_KEY` | 是 | — | DeepSeek API 密钥(兼容旧名 `LLM_API_KEY`) |
| `DEEPSEEK_BASE_URL` | 否 | `https://api.deepseek.com` | API 根地址(兼容 `LLM_BASE_URL`) |
| `DEEPSEEK_MODEL` | 否 | `deepseek-v4-flash` | 模型名(兼容 `LLM_MODEL`；也可用 `deepseek-v4-pro`) |
| `MATHAI_API_TOKEN` | 生产必填 | 空 | `/api` Bearer Token |
| `MATHAI_ALLOW_ORIGIN` | 否 | 空 | 唯一允许的跨域 Origin；同源部署留空 |
| `MATHAI_DATA_FILE` | 否 | `data/sessions.json` | 会话文件路径 |
| `PORT` | 否 | `8080` | HTTP 端口 |

## Docker Compose

从 `goagent-new/mathai` 执行：

```bash
cp .env.example .env
docker compose config
docker compose up --build -d
docker compose ps
curl -s http://localhost:8080/livez
curl -s http://localhost:8080/readyz
```

Compose 分别构建两个镜像，context 都在各自目录：

- 后端 `context: ./backend`。`mathai/backend` 是独立模块 `mathai`，`pkg/agent`、`pkg/llm`、`pkg/trace` 已放在该目录内，构建不依赖仓库根目录。
- 前端 `context: ./frontend`，产出 Nginx 镜像，托管 SPA 并把 `/api`、`/livez`、`/readyz` 反向代理到 `backend:8080`。

浏览器仍访问 `http://localhost:8080`，由前端容器对外暴露 80 端口。后端不对外发布端口。

## 数据与恢复

Compose 把后端的 `/app/data` 挂载到命名卷 `mathai-data`。Store 每次写入先创建同目录临时文件、同步、关闭，再 rename 替换正式文件。

备份前可以短暂停止写请求，然后备份卷中的 `sessions.json`。恢复时停止容器、替换文件并重新启动。损坏的 JSON 会让服务启动失败，而不会静默忽略数据。

## 发布检查

1. 在 `mathai/backend` 下 `go test ./...` 通过；
2. `npm run build` 通过；
3. `docker compose config` 通过；
4. `MATHAI_API_TOKEN` 不是示例值；
5. `/livez` 与 `/readyz` 均为 200；
6. 新建会话并完成一次工具调用；
7. 重启容器后会话仍存在；
8. 错误 Token 返回 401；
9. 反向代理限制请求体，并将超时设为大于后端 80 秒；
10. 为数据卷建立备份和容量告警。

## 横向扩展

当前不要运行多个共享同一 JSON 卷的副本。扩展前需要把 Store 换为数据库，用 `UPDATE ... WHERE version = ?` 保留乐观并发；把限流和 trace 移到共享基础设施；再让负载均衡器分发请求。
