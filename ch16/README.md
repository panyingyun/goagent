# 课时16:MCP 进阶

> 目标:把 ch02 的最小 MCP 升级成"真实生态"的形态——HTTP 传输 + 多服务器路由。

## 本课时学什么

1. **HTTP 传输**:JSON-RPC 走 `POST /mcp`,客户端不再依赖本地子进程;
2. **多服务器路由**:Router 把多个 MCP 服务器合成一个命名空间,
   工具名统一为 `服务器名.工具名`(如 `weather.get_weather`);
3. **传输无关抽象**:`ToolProvider` 接口让 stdio / HTTP / 内存服务器
   对上层完全透明。

## 运行

```powershell
cd E:\codex\project001\goagent\ch16
go run .
```

离线测试:

```powershell
go test ./... -v
```

## 和 ch02 的对比

| | ch02 | ch16 |
|---|---|---|
| 传输 | stdio | stdio + HTTP |
| 服务器数量 | 1 个 | 多个,路由合成 |
| 客户端 | 直连 | Router 统一入口 |

## 练习

1. 给 `callServerTool` 加一个 `multiply` 工具,用 HTTP 客户端调用;
2. 把 `HTTPMCPHandler` 挂到真正的 `http.ListenAndServe` 上,
   用 curl 手动发一个 JSON-RPC 请求;
3. 想一想:生产环境为什么需要"路由 + 鉴权"(提示:多个团队各自维护服务器)。
