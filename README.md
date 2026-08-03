# Go AI Agent 五课时课程

从零用 Go(标准库,零第三方依赖)实现一个调用 DeepSeek 的 AI Agent,
循序渐进掌握 LLM、MCP、SKILL、ReAct 等概念。

## 课时安排

| 课时 | 目录 | 内容 | 前置 |
|---|---|---|---|
| 1 | [ch01](ch01) | Go 调用 LLM(DeepSeek) | 无(需要 API Key) |
| 2 | [ch02](ch02) | Go 实现 MCP 服务器与客户端 | 无(不需要 API Key) |
| 3 | [ch03](ch03) | Go 实现 SKILL 调用 | 课时1 |
| 4 | [ch04](ch04) | Go 实现 ReAct 循环 | 课时1 |
| 5 | [ch05](ch05) | 综合:完整 AI Agent(ReAct + MCP + SKILL + LLM) | 1–4 |

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