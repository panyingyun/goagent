# 课时25:LangServe 风格服务

> 目标:用 Go 复刻 LangServe 的部署接口——/invoke、/stream、/playground、/docs。

## 本课时学什么

1. **/invoke**:把"输入"变成"输出"的标准接口;
2. **/chat/stream**:SSE 流式(复用课时7/17 的能力);
3. **/playground**:浏览器可视化调试页(LangServe 的招牌功能);
4. **/docs**:接口自描述。

## 运行

```powershell
cd E:\codex\project001\goagent\ch25
go run .
```

浏览器打开 http://localhost:8080/playground,或用命令行:

```powershell
Invoke-RestMethod -Method Post -Uri http://localhost:8080/invoke -ContentType application/json -Body '{"input":"你好"}'
```

## 和真实 LangServe 的关系

LangServe 是 Python(FastAPI)专属,我们的 Go 版实现了同样的接口契约。
如果以后要和 LangServe 服务互操作,只需要对齐这些接口约定即可——
这正是"协议 > 实现"的意义。

## 练习

1. 给 `/invoke` 加 `config` 参数(选模型 flash/pro,参考 ch18);
2. 加一个 `/chain` 接口,把 ch22 的 SequentialChain 部署上去;
3. 用 Docker 打包(ch21 的 Dockerfile 改改就能用)。

## 课程结束语

恭喜,你从"Hello World"走到了一个可以部署的 Go Agent 全栈:
LLM、MCP、SKILL、ReAct、记忆、流式、RAG、工具、评估、可观测性、
以及 LangChain 生态的四个核心概念。剩下的路,就是往真实业务里填了。
