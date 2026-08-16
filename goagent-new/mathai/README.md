# ch14：MathAI 综合案例

目标：用户用自然语言创建和比较函数，Agent 只能通过受控工具改图形，前端实时绘制，每次 LLM/工具调用可追踪。

## 运行与代码

```bash
cd mathai/backend
export DEEPSEEK_API_KEY="你的密钥"
go test ./... -v
go run . -eval
go run .

# 另开终端
cd mathai/frontend
npm install
npm run dev
```

打开 `http://localhost:5173`。可选 `MATHAI_API_TOKEN`；若设置，前端左下角填同一个值。

```bash
cd mathai
cp .env.example .env
docker compose up --build
```

Docker 访问 `http://localhost:8080`。

- [`backend/math_agent.go`](backend/math_agent.go)：读工具、写工具、审批与 trace；
- [`backend/eval.go`](backend/eval.go)：读取 `testdata/eval-cases.json`，用 `pkg/eval` 核对补丁轨迹和 GraphState；
- [`backend/domain.go`](backend/domain.go)：`GraphPatch` 确定性校验，模型不能直接覆盖存储；
- [`frontend/src/components/GraphViewer.tsx`](frontend/src/components/GraphViewer.tsx)：用 Mafs 绘制函数图像，表达式用 mathjs 求值，不使用 `eval`；
- [`docs/course-guide.md`](docs/course-guide.md)：一次聊天请求的完整链路。

后端是独立模块 `mathai`，自带 `pkg/llm`、`pkg/agent`、`pkg/trace` 副本。写图形必须走补丁；并发更新用版本号，冲突返回 `409`。

## 验收场景

- “画出 `y=x` 和 `y=x^2`”产生两条曲线，刷新后仍在；
- “隐藏第一条函数”先读图形状态，再按 `equationId` 修改；
- 非法表达式、未知补丁字段和越界视口被后端拒绝；
- 两个并发请求不会静默覆盖，后提交者得到 `409 Conflict`；
- assistant 消息可打开 trace；
- 未配置模型时 `/livez` 正常、`/readyz` 返回 503；
- 容器重启后会话仍然存在。

更详细的接口、安全和部署见 [`docs/`](docs/)。
