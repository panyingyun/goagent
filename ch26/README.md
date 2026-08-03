# 课时26:综合案例——AI 数学方程绘图智能体(MathGraph AI)

> 课程收官案例。根据 [docs/sw-design.md](docs/sw-design.md) 的产品设计稿,
> 用 **Go 后端(替代设计稿的 Python)+ React 前端 + SQLite** 完整实现,
> 把 ch01~ch25 的知识点组装成一个能"自然语言画函数图像"的 Web 应用。

## 能做什么

用户用中文描述绘图需求,AI 自动解析并绘制函数图像,支持多轮修改:

```text
画 y = x²            -> 绘制抛物线
再加一条 y = sin(x)   -> 追加曲线,旧曲线保留
把第一条曲线改成红色   -> 修改颜色
把坐标范围改成 -5 到 5 -> 更新视图
解释这个图像          -> AI 分析图像特征
```

## 技术栈

| 层 | 设计稿要求 | 本课实现 |
|---|---|---|
| 后端 | Python + FastAPI | **Go 1.25 + net/http**(纯标准库路由) |
| 数据库 | SQLite + SQLAlchemy | **SQLite + modernc.org/sqlite**(纯 Go,无 CGO) |
| AI | DeepSeek API | DeepSeek `deepseek-v4-flash`(设计稿的 deepseek-chat 已退役) |
| 前端 | React + Vite + TS + Tailwind + Plotly + KaTeX + Zustand | 完全按设计稿实现 |

## 目录结构

```text
ch26/
├── docs/                  # 设计稿(sw-design / ui-design / 截图 / 主题色)
├── backend/               # Go 后端
│   ├── main.go            # 入口(托管 API + 前端 dist)
│   ├── models.go          # 数据模型(与设计稿 §6 对应)
│   ├── store.go           # SQLite 存储(表结构按设计稿 §8)
│   ├── llm.go             # DeepSeek service(设计稿 §10)
│   ├── prompt.go + prompt.txt  # 系统提示词(独立文件,呼应 ch03 SKILL)
│   ├── apply.go           # graphState 更新规则(设计稿 §14)+ 意图兜底
│   ├── api.go             # HTTP API + chat 流程(设计稿 §9)
│   └── api_test.go        # 9 条验收用例的 Go 实现(mock DeepSeek,离线)
├── frontend/              # React 前端
    ├── src/types.ts       # 与设计稿 §6 一致的类型
    ├── src/store.ts       # Zustand 全局状态
    ├── src/plot.ts        # mathjs 安全采样(禁 eval,设计稿 §5.2)
    └── src/components/    # 三栏工作台:Sidebar / ChatPanel / ResultPanel / GraphViewer
└── docker-compose.yml     # 一键启动后端 + 前端
```

## 快速开始

### 1. 后端

```powershell
cd E:\codex\project001\goagent\ch26\backend
$env:DEEPSEEK_API_KEY = "sk-..."
go run .
```

后端默认在 http://localhost:8080,提供 `/api/*` 接口。

### 2. 前端(开发模式)

```powershell
cd E:\codex\project001\goagent\ch26\frontend
npm install
npm run dev
```

浏览器打开 http://localhost:5173(Vite 会把 `/api` 代理到后端)。

### 3. 生产模式(后端托管前端)

```powershell
cd E:\codex\project001\goagent\ch26\frontend
npm run build        # 生成 dist/
cd ..\backend
go run .             # 直接访问 http://localhost:8080 即是完整应用
```

### 4. Docker Compose 一键启动

```powershell
cd E:\codex\project001\goagent\ch26
Copy-Item .env.example .env     # 然后编辑 .env 填入 DEEPSEEK_API_KEY
docker compose up --build
```

- 前端: http://localhost:5173(容器内 `VITE_API_PROXY` 指向 backend 服务)
- 后端 API: http://localhost:8080
- SQLite 数据持久化在 `./data/math_graph_ai.db`(已 gitignore)
- 后端容器带健康检查,前端等后端就绪后才启动

停止:`docker compose down`;清数据:`docker compose down -v`(会删除 data 卷)。

### 测试

```powershell
cd E:\codex\project001\goagent\ch26\backend
go test ./... -v      # 9 条验收用例 + 单元测试,mock DeepSeek,离线可跑
```

## 后端 API(设计稿 §9)

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /api/sessions | 会话列表 |
| POST | /api/sessions | 新建会话 |
| GET | /api/sessions/{id} | 完整会话(消息 + graphState) |
| PATCH | /api/sessions/{id} | 更新标题 / graphState / 收藏 |
| DELETE | /api/sessions/{id} | 删除会话 |
| POST | /api/chat | 核心:自然语言 -> 结构化结果 -> 更新图像 |

## chat 数据流

```mermaid
flowchart LR
    U[用户消息] --> S[保存 user 消息到 SQLite]
    S --> P[组装 system + currentGraphState + 最近消息]
    P --> D[DeepSeek 解析为结构化 JSON]
    D --> V[校验:intent/方程默认值/数字id兼容]
    V --> C[意图兜底:再加->add_equation, unknown+解释->重试]
    C --> A[按规则更新 graphState]
    A --> M[保存 assistant 消息 + graphState]
    M --> F[返回给前端 Plotly 绘制]
```

## 验收用例(设计稿 §22)与本课实现

| # | 用例 | 实现 |
|---|---|---|
| 1 | 绘制一次函数 | `TestPlotOnceFunction` ✓ |
| 2 | 绘制二次函数 | `TestPlotQuadratic` ✓ |
| 3 | 追加函数保留旧曲线 | `TestAddEquationKeepsOld` ✓ |
| 4 | 修改颜色 | `TestUpdateColor` ✓ |
| 5 | 修改坐标范围 | `TestUpdateViewport` ✓ |
| 6 | 非法公式不崩溃 | `TestUnknownIntent` ✓ |
| 7 | 会话保存与隔离 | `TestSessionIsolation` ✓ |
| 8 | 删除方程持久化 | `TestDeleteEquationViaPatch` ✓ |
| 9 | 导出 PNG | 前端 Plotly.toImage(浏览器手测) |

## ch01~ch25 知识点映射

| 用到的地方 | 对应课时 |
|---|---|
| DeepSeek 客户端 / 结构化输出 / 提示词(独立文件) | ch01 / ch11 / ch03 |
| 多轮上下文 + recentMessages | ch06 |
| 解析校验、默认值补齐、意图兜底、重试 | ch10 / ch13 |
| SQLite 持久化、检查点思想 | ch06 / ch14 |
| HTTP 服务、CORS、前端托管 | ch10 / ch25 |
| 结构化日志(可选 DEEPSEEK_DEBUG) | ch20 |
| 工具安全思路(禁 eval、白名单函数) | ch09 |
| 组件化分层(service/apply/handler) | ch22 |
| 状态更新规则(状态机) | ch14 / ch23 |

## 真实联调踩过的三个坑(也是教学点)

1. **模型输出不稳定**:`"id"` 有时是数字 → 自定义 `FlexID` 类型兼容;
2. **意图误判**:"再加一条"被判成 plot(会覆盖旧图)→ 关键词兜底强制转 add_equation;
   "解释这个图像"被误判 unknown → 带方程上下文重试一次;
3. **调试可观测性**:`DEEPSEEK_DEBUG=1` 打印原始响应,没有它很难定位上面的问题。

## 环境变量

```env
DEEPSEEK_API_KEY=你的Key
DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_MODEL=deepseek-v4-flash
DATABASE_PATH=math_graph_ai.db
PORT=8080
```

## 扩展方向

隐函数 / 参数方程 / 求交点 / 导出 SVG/CSV / 深色主题 / 分享链接 /
KaTeX 公式美化 / 自动会话标题(设计稿 P1/P2 待办清单)。
