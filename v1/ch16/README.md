# 课时16:MCP 服务器与客户端

> 目标:搞懂 MCP 是什么、消息长什么样、服务器和客户端怎么配合。

## 本课时学什么

- MCP = Model Context Protocol,被比喻为"AI 世界的 USB-C":服务器把工具暴露出来,
  任何支持 MCP 的客户端都能即插即用;
- 传输:最小实现用 **stdio**(标准输入/输出),通信格式是 **JSON-RPC 2.0**,
  每一行一个 JSON 消息;
- 最小闭环只需要 3 个方法:
  - `initialize`:握手,双方确认协议版本和能力;
  - `tools/list`:客户端问"你有什么工具?";
  - `tools/call`:客户端说"帮我调用这个工具",参数传过去,结果传回来。

## 目录结构

```text
ch16/
├── server/main.go   # MCP 服务器:监听 stdin,处理 JSON-RPC,暴露 echo / add 两个工具
├── client/main.go   # MCP 客户端:启动服务器子进程,握手,列出工具并调用
└── README.md
```

## 运行

方式一:直接看协议(最推荐,先理解消息长什么样)。
在 `ch16` 目录下执行:

```powershell
'{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | go run ./server
```

你会看到服务器返回工具列表。再试试调用工具:

```powershell
'{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"add","arguments":{"a":12,"b":8}}}' | go run ./server
```

方式二:跑完整的客户端。

```powershell
go build -o server.exe ./server
go run ./client -server server.exe
```

客户端会依次完成握手、列工具、调用 echo 和 add,把每一步的结果打印出来。

## 消息格式速查

请求(客户端 → 服务器):

```json
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"add","arguments":{"a":12,"b":8}}}
```

响应(服务器 → 客户端):

```json
{"jsonrpc":"2.0","id":1,"result":{"content":[{"type":"text","text":"20"}],"isError":false}}
```

要点:

- `jsonrpc` 固定是 `"2.0"`;
- `id` 用来把请求和响应配对(请求是 1,响应就带 1);
- 通知(notification)没有 `id`,服务器不需要回复;
- 工具结果放在 `result.content` 里,失败时 `isError` 为 `true`。

## 练习

1. 给服务器加一个 `multiply`(乘法)工具,用方式一验证;
2. 把服务器的 stderr 日志去掉,观察会发生什么(提示:日志会污染协议通道);
3. 想一想:如果改成 HTTP 传输,客户端和服务器需要改哪里?

## 小结

MCP 解决的是"**工具怎么被通用地暴露和调用**"的问题。学完这一课,你已经有能力
用 Go 写出一个真正的 MCP 服务器/客户端了——课时4 会把 MCP 接进 Agent。
