# ch14：MathAI 综合案例

MathAI 是 `goagent-new` 主线之后的第一个完整 Web 案例：用户用自然语言创建和比较函数，Agent 通过受控工具更新图形状态，React 前端实时绘制结果，并且每次 LLM/工具调用都可追踪。

它参考旧课程 `ch26` 的三栏交互与函数可视化目标，但重新使用本课程已经构建的工程架构：

- `pkg/llm`：OpenAI-compatible LLM 客户端；
- `pkg/agent`：原生 tool calls、审批、幂等、超时和观察器；
- `pkg/trace`：LLM 与工具执行轨迹；
- 严格 `GraphPatch`：模型不能直接覆盖数据库对象；
- 乐观并发控制与原子 JSON 持久化；
- 同源 React SPA、受保护 HTTP API 和单容器部署。

## 目录

```text
mathai/
├── backend/              # Go API、Agent、领域模型、存储和测试
├── frontend/             # React + TypeScript + Canvas 前端
├── docs/                 # 架构、API、安全、部署、评估和学习指南
├── testdata/             # 离线评估案例
├── Dockerfile
└── docker-compose.yml
```

## 本地运行

后端需要一个支持原生 tool calling 的 OpenAI-compatible 模型：

```powershell
cd goagent-new/mathai
$env:LLM_API_KEY = "你的密钥"
$env:LLM_BASE_URL = "https://api.deepseek.com"
$env:LLM_MODEL = "供应商当前支持工具调用的模型"
$env:MATHAI_API_TOKEN = "local-dev-token" # 可留空；生产环境不要留空
go run ./backend
```

另开终端启动前端。开发服务器会把 `/api` 代理到 `localhost:8080`：

```powershell
cd goagent-new/mathai/frontend
npm install
npm run dev
```

打开 `http://localhost:5173`。如果配置了 `MATHAI_API_TOKEN`，在左下角输入同一个值。

## 测试与构建

```powershell
cd goagent-new
go test ./mathai/backend -v

cd mathai/frontend
npm run build
```

构建前端后，也可以从 `mathai` 目录运行 `go run ./backend`，由 Go 服务直接托管 `frontend/dist`。

## Docker

```powershell
cd goagent-new/mathai
Copy-Item .env.example .env
# 修改 .env
docker compose up --build
```

访问 `http://localhost:8080`。会话数据保存在命名卷 `mathai-data`；镜像以非 root 用户运行，根文件系统只读。

## 建议学习顺序

1. 阅读 [`docs/course-guide.md`](docs/course-guide.md)，先理解一次聊天请求的完整链路；
2. 阅读 `backend/domain.go`，观察模型输出为什么必须再次经过确定性校验；
3. 阅读 `backend/math_agent.go`，理解读工具、写工具、审批与 trace 如何组合；
4. 运行后端测试，用 scripted LLM 验证 Agent，而不消耗真实 API；
5. 阅读前端 `GraphCanvas.tsx`，理解表达式计算为什么不使用 `eval`；
6. 最后阅读部署与安全文档，完成验收场景。

## 验收场景

- “画出 `y=x` 和 `y=x^2`”产生两条曲线，并保存在刷新后的会话中；
- “隐藏第一条函数”先读取图形状态，再按 `equationId` 修改；
- 非法表达式、未知补丁字段和越界视口被后端拒绝；
- 两个并发请求不会静默覆盖，后提交者得到 `409 Conflict`；
- assistant 消息可以打开 trace，看到 LLM 和工具事件；
- 未配置模型时 `/livez` 正常、`/readyz` 返回 503；
- 容器重启后会话仍然存在。

更详细的接口和设计说明见 [`docs/`](docs/)。
