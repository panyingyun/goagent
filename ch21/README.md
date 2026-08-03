# 课时21:部署与运维

> 目标:让 Agent 服务具备上线的基本素质——健康检查、限流、优雅关闭、密钥安全。

## 本课时学什么

1. **健康检查**:`/health`(存活)+ `/ready`(就绪,检查依赖是否配好);
2. **限流**:每 IP 每秒最多 N 次,防止被打爆;
3. **优雅关闭**:收到 Ctrl+C / SIGTERM 后,等当前请求处理完再退出;
4. **密钥管理**:只从环境变量读,不进代码、不进 git。

## 运行

```powershell
cd E:\codex\project001\goagent\ch21
go run .
```

另一个终端:

```powershell
Invoke-RestMethod -Uri http://localhost:8080/health
Invoke-RestMethod -Uri http://localhost:8080/ready
Invoke-RestMethod -Method Post -Uri http://localhost:8080/chat -ContentType application/json -Body '{"question":"你好"}'
```

连续快速点 `/chat`,第 6 次会看到 429 限流。Ctrl+C 观察优雅关闭日志。

## Docker 部署

```powershell
docker build -t goagent .
docker run -p 8080:8080 -e DEEPSEEK_API_KEY=你的Key goagent
```

## 练习

1. 把限流窗口改成 10 秒,观察行为;
2. 加一个 `/metrics` 接口,输出课时20 的指标;
3. 想想生产还要补什么?(提示:日志聚合、监控告警、HTTPS、密钥轮换。)
