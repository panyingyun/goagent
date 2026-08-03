# Go AI Agent 课程

从零用 Go(标准库,零第三方依赖)实现一个调用 DeepSeek 的 AI Agent,
循序渐进掌握 LLM、MCP、SKILL、ReAct 等概念。

## 课时安排(ch01~ch10 已实现)

| 课时 | 目录 | 内容 | 前置 |
|---|---|---|---|
| 1 | [ch01](ch01) | Go 调用 LLM(DeepSeek) | 无(需要 API Key) |
| 2 | [ch02](ch02) | Go 实现 MCP 服务器与客户端 | 无(不需要 API Key) |
| 3 | [ch03](ch03) | Go 实现 SKILL 调用 | 课时1 |
| 4 | [ch04](ch04) | Go 实现 ReAct 循环 | 课时1 |
| 5 | [ch05](ch05) | 综合:完整 AI Agent(ReAct + MCP + SKILL + LLM) | 1–4 |
| 6 | [ch06](ch06) | 多轮对话与记忆(截断/摘要/偏好) | 课时1 |
| 7 | [ch07](ch07) | 流式输出(SSE 解析) | 课时1 |
| 8 | [ch08](ch08) | RAG 入门(路线B:OpenAI embedding) | 课时6 |
| 9 | [ch09](ch09) | 真实工具(HTTP/文件/命令白名单/存储) | 课时4/5 |
| 10 | [ch10](ch10) | 评估、安全与工程化(HTTP 服务) | 课时5 |
| 11 | [ch11](ch11) | 提示词工程(角色/few-shot/结构化输出) | 课时1 |
| 12 | [ch12](ch12) | 向量数据库与高级 RAG(混合检索/重排) | 课时8 |
| 13 | [ch13](ch13) | 工具进阶与并行调用(校验/缓存/并发) | 课时5 |
| 14 | [ch14](ch14) | Agent 工作流与任务规划(状态机/检查点) | 课时5 |
| 15 | [ch15](ch15) | 多 Agent 协作(规划/执行/审查) | 课时5 |
| 16 | [ch16](ch16) | MCP 进阶(HTTP 传输/多服务器路由) | 课时2 |
| 17 | [ch17](ch17) | Agent Web 界面(JSON + SSE 聊天) | 课时7 |
| 18 | [ch18](ch18) | 模型路由与成本优化(flash/pro/缓存) | 课时1 |
| 19 | [ch19](ch19) | 评估进阶:LLM-as-Judge | 课时10 |
| 20 | [ch20](ch20) | 可观测性与日志(span 树/指标) | 课时10 |
| 21 | [ch21](ch21) | 部署与运维(健康检查/限流/优雅关闭/Docker) | 课时10 |
| 22 | [ch22](ch22) | 迷你 LangChain(Chain/AgentExecutor) | 课时4 |
| 23 | [ch23](ch23) | 迷你 LangGraph(节点/边/状态/检查点) | 课时14 |
| 24 | [ch24](ch24) | LangSmith 风格追踪(run 树) | 课时20 |
| 25 | [ch25](ch25) | LangServe 风格服务(/invoke//playground) | 课时17/21 |
| 26 | [ch26](ch26) | 综合案例:AI 数学方程绘图智能体(Go+React+SQLite+DeepSeek) | 1–25 |

> 全部 26 课时完成。课程路线:基础(ch01~05) -> 能力增强(ch06~10) -> 进阶(ch11~21) -> LangChain 生态对照(ch22~25) -> 综合收官(ch26)。

## 怎么开始

```powershell
# 课时1 起需要 API Key(已配置),先体验最快见效的一课
cd ch01
go run . "用一句话介绍你自己"
```

每章目录下都有独立的 `README.md`,包含:概念讲解、运行方法、练习和小结。
建议按顺序学习,但每章都可独立运行。

## 概念心智地图

> LLM 负责"想",Tool 负责"做",Agent 负责把两者用循环接起来;
> Function Calling 让"想调工具"结构化;MCP 让工具即插即用;
> SKILL 把"怎么做"的说明书打包复用。

## 环境要求

- Go 1.25+
- `DEEPSEEK_API_KEY`(DeepSeek 开放平台申请)
- 网络可访问 `https://api.deepseek.com`
