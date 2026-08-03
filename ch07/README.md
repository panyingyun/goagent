# 课时7:流式输出

> 目标:让 Agent 的回复"边生成边显示",体验立刻不一样。

## 本课时学什么

1. 为什么流式:真实产品里没人愿意干等十几秒看整段回复;
2. SSE 格式:`stream: true` 之后,响应变成一行行的 `data: {...}`;
3. 增量解析:每个分片里 `choices[0].delta.content` 是一小段文字;
4. 结束标志:`data: [DONE]`。

## SSE 长什么样

```text
data: {"choices":[{"delta":{"content":"你"}}]}

data: {"choices":[{"delta":{"content":"好"}}]}

data: [DONE]
```

注意:消息之间有空行;每个 `data:` 行的内容是一段完整 JSON。
我们的 `StreamChat` 用 `bufio.Scanner` 逐行读,解析增量后回调 `onDelta`。

## 目录结构

```text
ch07/
├── main.go        # 逐字打印 + 耗时统计
├── llm.go         # StreamChat:流式请求与 SSE 解析
├── llm_test.go    # 本地 mock SSE 服务器,离线验证拼接正确
└── README.md
```

## 运行

```powershell
cd E:\codex\project001\goagent\ch07
go run . "用三句话介绍 Go 语言"
go run . "写一首关于夏天的短诗"
```

离线测试:

```powershell
go test ./... -v
```

## 和课时6 怎么配合

流式只是"传输方式",和记忆、工具完全正交。把 ch06 的 `llm.Chat` 换成
`StreamChat`,边接收边攒内容,结束再把完整回复加进历史——这就是完整聊天体验。

## 练习

1. 把 `onDelta` 改成"每收到 10 个字才打印一次",观察有什么不同;
2. 在 `[DONE]` 前突然断开连接,看 `StreamChat` 返回什么错误;
3. 给流式加上"打字机"效果(随机延迟),做点好玩的;
4. 思考:流式下怎么做超时控制?(提示:用 `context.WithTimeout` 整体限时,
   或按"两次分片间隔"限时。)

## 小结

流式 = 请求加一个开关 + 逐行解析 SSE + 回调增量。它是 Agent 产品化的第一步,
课时 16 的 Web 界面、课时 10 的 HTTP 服务都会依赖它。
