# Go AI Agent 课程

从零用 Go(标准库,零第三方依赖)实现一个调用 DeepSeek 的 AI Agent,
循序渐进掌握 LLM、提示词、ReAct、Function Calling、MCP、SKILL 等概念。

> 目录 `ch01`–`ch26` 已按**推荐学习顺序**编号,可从 ch01 起顺序学习。

## 推荐学习路径(先看这个)

```text
必修 A · Agent 内核(~5 课)
  ch01 LLM
  → ch02 提示词工程
  → ch03 ReAct 循环(理解"想-做-看")
  → ch04 完整 Agent(默认关注 -mode tool / Function Calling)
  → ch05 SKILL

必修 B · 能用起来(~5 课)
  ch06 记忆 → ch07 流式
  → ch08 真实工具与安全
  → RAG: ch09 为主(混合检索/重排,零外部 embedding Key)
       ch19 为辅(可选:真实 OpenAI embedding)
  → ch10 工具校验/缓存/并行

必修 C · 工程化(~5–6 课)
  ch11 工作流与检查点
  → ch12 评估/安全/HTTP 服务
  → ch13 成本路由 → ch14 可观测 → ch15 部署
  →(可选) ch16 MCP → ch17 多 Agent → ch18 Web 界面

收官
  ch26 综合案例 —— 完成必修 A+B+C 即可动手,不必先学完选修

选修 / 对照附录(可整段跳过)
  ch19 RAG 入门(OpenAI embedding)  ← 需要真实 embedding 时再学
  ch20 MCP 进阶(HTTP/多服务器)
  ch21 LLM-as-Judge                 ← 做评估体系时再学(前置 ch12)
  ch22 迷你 LangChain  ≈ 给 ch01/03/06 换名字,看对照表即可
  ch23 迷你 LangGraph  ≈ ch11+ch17 的图抽象,与 ch11 合并理解
  ch24 LangSmith 风格  ≈ ch14 span 树换皮,并入 ch14 心智即可
  ch25 LangServe 风格  ≈ ch18+ch15 的接口约定,对接 Python 生态再学
```

**目标工时感**:必修约 12–16 课即可具备「能做出可上线 Agent」的主干;
ch22–25 是生态对照,不是 Agent 新能力。

## 全课时索引

| 课时 | 目录 | 内容 | 学习标记 |
|---|---|---|---|
| 1 | [ch01](ch01) | Go 调用 LLM(DeepSeek) | ✅ 已完成 · 必修 |
| 2 | [ch02](ch02) | 提示词工程 | ✅ 已完成 · 必修 |
| 3 | [ch03](ch03) | ReAct 文本协议循环 | ✅ 已完成 · 必修 |
| 4 | [ch04](ch04) | 完整 Agent(ReAct + MCP + SKILL + FC) | 必修 |
| 5 | [ch05](ch05) | SKILL 调用 | ✅ 已完成 · 必修 |
| 6 | [ch06](ch06) | 多轮对话与记忆 | 必修 |
| 7 | [ch07](ch07) | 流式输出(SSE) | 必修 |
| 8 | [ch08](ch08) | 真实工具与安全边界 | 必修 |
| 9 | [ch09](ch09) | 高级 RAG(哈希向量/混合/重排) | 必修(RAG 主线) |
| 10 | [ch10](ch10) | 工具校验/缓存/并行 | 必修 |
| 11 | [ch11](ch11) | 工作流与检查点 | 必修 |
| 12 | [ch12](ch12) | 评估、安全与工程化 | 必修 |
| 13 | [ch13](ch13) | 模型路由与成本 | 必修 |
| 14 | [ch14](ch14) | 可观测性(span 树) | 必修 |
| 15 | [ch15](ch15) | 部署与运维 | 必修 |
| 16 | [ch16](ch16) | MCP 服务器与客户端(stdio) | 建议·可后置 |
| 17 | [ch17](ch17) | 多 Agent 协作 | 建议 |
| 18 | [ch18](ch18) | Agent Web 界面 | 建议 |
| 19 | [ch19](ch19) | RAG 入门(OpenAI embedding) | 选修·与09合并理解 |
| 20 | [ch20](ch20) | MCP 进阶(HTTP/路由) | 选修 |
| 21 | [ch21](ch21) | LLM-as-Judge | 选修 |
| 22 | [ch22](ch22) | 迷你 LangChain | 选修·对照 |
| 23 | [ch23](ch23) | 迷你 LangGraph | 选修·对照 |
| 24 | [ch24](ch24) | LangSmith 风格追踪 | 选修·对照 |
| 25 | [ch25](ch25) | LangServe 风格服务 | 选修·对照 |
| 26 | [ch26](ch26) | 综合:数学方程绘图智能体 | 收官 |

### 合并理解提示(代码未删,学习时当一章读)

| 主题 | 怎么学 | 可跳过重复练习的部分 |
|---|---|---|
| RAG | 以 **ch09** 建立流水线心智;需要真实 embedding 再跑 **ch19** | 两章都写一遍切块+检索时可只精读一章 |
| 工作流 / 图 | 以 **ch11** 为主;想看「条件边/回边」再扫 **ch23** | ch23 与 ch17 的 plan/execute/review 不要当新概念 |
| 追踪 | 以 **ch14** 为主;导出嵌套 run 树看 **ch24** 即可 | 不必两章都手敲一遍 |
| HTTP+SSE 界面 | 以 **ch18** 为主;**ch25** 仅在要对齐 LangServe 约定时看 | playground 差异可当附录 |

## 怎么开始

```powershell
# 课时1 起需要 API Key,先体验最快见效的一课
cd ch01
go run . "用一句话介绍你自己"
```

推荐起步顺序(与目录编号一致):

```text
ch01 → ch02 → ch03 → ch04 → ch05 → ch06 → ch07 → ch08 → ch09 → ch10
→ ch11 → ch12 → ch13 → ch14 → ch15 →(可选 ch16/ch17/ch18)→ ch26
```

每章目录下都有独立的 `README.md`(概念、运行、练习、小结)。
章节可独立运行;学必修时遇到「每章复制一份 `llm.go`」是刻意设计——
重点看**相对上一章多出来的文件与循环逻辑**,样板客户端可快速扫过。

## 概念心智地图

> LLM 负责"想",Tool 负责"做",Agent 负责把两者用循环接起来;
> Function Calling 让"想调工具"结构化(生产默认);文本 ReAct 利于理解循环;
> MCP 让工具即插即用;SKILL 把"怎么做"的说明书打包复用。

## 环境要求

- Go 1.25+
- `DEEPSEEK_API_KEY`(DeepSeek 开放平台申请)
- 网络可访问 `https://api.deepseek.com`
- 仅 ch19 额外需要 `OPENAI_API_KEY`(可跳过,用 ch09 学 RAG)
