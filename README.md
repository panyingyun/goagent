# Go AI Agent 课程

从零用 Go(标准库,零第三方依赖)实现一个调用 DeepSeek 的 AI Agent,
循序渐进掌握 LLM、MCP、SKILL、ReAct 等概念。

> 目录 `ch01`–`ch26` 代码全部保留,可独立运行。
> 下表按**学习价值**重排优先级:**不必按编号顺序学完 26 课**。

## 推荐学习路径(先看这个)

```text
必修 A · Agent 内核(~5 课)
  ch01 LLM
  → ch11 提示词工程(提前学,别等到后面)
  → ch04 ReAct 循环(理解"想-做-看")
  → ch05 完整 Agent(默认关注 -mode tool / Function Calling)
  → ch03 SKILL

必修 B · 能用起来(~5 课)
  ch06 记忆 → ch07 流式
  → ch09 真实工具与安全
  → RAG: ch12 为主(混合检索/重排,零外部 embedding Key)
       ch08 为辅(可选:真实 OpenAI embedding)
  → ch13 工具校验/缓存/并行

必修 C · 工程化(~5–6 课)
  ch14 工作流与检查点
  → ch15 多 Agent(可加深,非阻塞收官)
  → ch10 评估/安全/HTTP 服务
  → ch18 成本路由 → ch20 可观测 → ch21 部署
  → ch17 Web 界面(需要浏览器演示再学)

MCP 轨道(穿插即可)
  建议在 ch05 之后再深入:先会"工具循环",再学"工具协议"
  ch02 最小 MCP →(需要时) ch16 HTTP/多服务器

收官
  ch26 综合案例 —— 完成必修 A+B+C 即可动手,不必先学完选修

选修 / 对照附录(可整段跳过)
  ch22 迷你 LangChain  ≈ 给 ch01/04/06 换名字,看对照表即可
  ch23 迷你 LangGraph  ≈ ch14+ch15 的图抽象,与 ch14 合并理解
  ch24 LangSmith 风格  ≈ ch20 span 树换皮,并入 ch20 心智即可
  ch25 LangServe 风格  ≈ ch17+ch21 的接口约定,对接 Python 生态再学
  ch19 LLM-as-Judge    ← 做评估体系时再学(前置 ch10)
```

**目标工时感**:必修约 12–16 课即可具备「能做出可上线 Agent」的主干;
ch22–25 是生态对照,不是 Agent 新能力。

## 全课时索引(代码仍按原目录)

| 课时 | 目录 | 内容 | 学习标记 | 建议前置 |
|---|---|---|---|---|
| 1 | [ch01](ch01) | Go 调用 LLM(DeepSeek) | ✅ 已完成 · 必修 | API Key |
| 2 | [ch02](ch02) | MCP 服务器与客户端(stdio) | 必修·可后置 | 无 Key;建议 ch05 后 |
| 3 | [ch03](ch03) | SKILL 调用 | 必修 | ch01 |
| 4 | [ch04](ch04) | ReAct 文本协议循环 | 必修 | ch01 |
| 5 | [ch05](ch05) | 完整 Agent(ReAct + MCP + SKILL + FC) | 必修 | 1–4;生产默认看 `tool` 模式 |
| 6 | [ch06](ch06) | 多轮对话与记忆 | 必修 | ch01 |
| 7 | [ch07](ch07) | 流式输出(SSE) | 必修 | ch01 |
| 8 | [ch08](ch08) | RAG 入门(OpenAI embedding) | 选修·与12合并理解 | ch06;需 `OPENAI_API_KEY` |
| 9 | [ch09](ch09) | 真实工具与安全边界 | 必修 | ch04/5 |
| 10 | [ch10](ch10) | 评估、安全与工程化 | 必修 | ch05 |
| 11 | [ch11](ch11) | 提示词工程 | ✅ 已完成 · 必修·提前学 | ch01;排在 ch04/5 前或紧后 |
| 12 | [ch12](ch12) | 高级 RAG(哈希向量/混合/重排) | 必修(RAG 主线) | ch06;可不先学 ch08 |
| 13 | [ch13](ch13) | 工具校验/缓存/并行 | 必修 | ch05 |
| 14 | [ch14](ch14) | 工作流与检查点 | 必修 | ch05 |
| 15 | [ch15](ch15) | 多 Agent 协作 | 建议 | ch05 |
| 16 | [ch16](ch16) | MCP 进阶(HTTP/路由) | 选修 | ch02 |
| 17 | [ch17](ch17) | Agent Web 界面 | 建议 | ch07 |
| 18 | [ch18](ch18) | 模型路由与成本 | 必修 | ch01 |
| 19 | [ch19](ch19) | LLM-as-Judge | 选修 | ch10 |
| 20 | [ch20](ch20) | 可观测性(span 树) | 必修 | ch10;(含 ch24 心智) |
| 21 | [ch21](ch21) | 部署与运维 | 必修 | ch10 |
| 22 | [ch22](ch22) | 迷你 LangChain | 选修·对照 | ch04 |
| 23 | [ch23](ch23) | 迷你 LangGraph | 选修·对照 | ch14;(≈14+15) |
| 24 | [ch24](ch24) | LangSmith 风格追踪 | 选修·对照 | ch20;(≈20) |
| 25 | [ch25](ch25) | LangServe 风格服务 | 选修·对照 | ch17/21;(≈17) |
| 26 | [ch26](ch26) | 综合:数学方程绘图智能体 | 收官 | 必修 A+B+C,非 1–25 |

### 合并理解提示(代码未删,学习时当一章读)

| 主题 | 怎么学 | 可跳过重复练习的部分 |
|---|---|---|
| RAG | 以 **ch12** 建立流水线心智;需要真实 embedding 再跑 **ch08** | 两章都写一遍切块+检索时可只精读一章 |
| 工作流 / 图 | 以 **ch14** 为主;想看「条件边/回边」再扫 **ch23** | ch23 与 ch15 的 plan/execute/review 不要当新概念 |
| 追踪 | 以 **ch20** 为主;导出嵌套 run 树看 **ch24** 即可 | 不必两章都手敲一遍 |
| HTTP+SSE 界面 | 以 **ch17** 为主;**ch25** 仅在要对齐 LangServe 约定时看 | playground 差异可当附录 |

## 怎么开始

```powershell
# 课时1 起需要 API Key,先体验最快见效的一课
cd ch01
go run . "用一句话介绍你自己"
```

推荐起步顺序(与编号不完全一致):

```text
ch01 → ch11 → ch04 → ch05 → ch03 → ch06 → ch07 → ch09 → ch12 → ch13
→ ch14 → ch10 → ch18 → ch20 → ch21 →(可选 ch02/ch15/ch17)→ ch26
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
- 仅 ch08 额外需要 `OPENAI_API_KEY`(可跳过,用 ch12 学 RAG)
