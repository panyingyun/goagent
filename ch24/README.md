# 课时24:LangSmith 风格追踪

> 目标:实现 LangSmith 的核心抽象——run 树,并了解官方 Go SDK 的接入方式。

## 本课时学什么

1. **run 树**:每次 LLM/工具调用是一个 run,带父子关系、属性、耗时;
2. **失败记录**:`FailRun` 把错误挂到对应 run 上;
3. **导出**:嵌套 JSON,和 LangSmith 的层级视图同构。

## 运行

```powershell
cd E:\codex\project001\goagent\ch24
go run . "12 * 8 + 4 等于多少?"
```

输出回答 + 一棵 run 树:agent -> llm.think -> tool.calculator -> llm.answer。

## 官方 LangSmith Go SDK(可选接入)

LangSmith 官方提供了 Go SDK(`github.com/langchain-ai/langsmith-go`),
把本课的 run 树上传到平台即可获得可视化、数据集、评估能力:

```go
import "github.com/langchain-ai/langsmith-go"

client, _ := langsmith.NewClient(langsmith.WithAPIKey(os.Getenv("LANGSMITH_API_KEY")))
// 用 client 创建 run,设置 parent,结束上传
```

## 练习

1. 给 run 加 `cost` 属性(结合 ch18 的成本估算);
2. 把 run 树同时写进日志文件(结合 ch10 的结构化日志);
3. 有条件的话,注册 LangSmith 账号,把 run 真正上传上去看可视化。
