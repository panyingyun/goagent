# 课时8:真实工具——让 Agent 真的"干活"

> 目标:给 Agent 装上联网、文件、命令、存储四只手,同时学会"工具安全"。

## 本课时学什么

1. **http_get**:请求公开 API/网页,Agent 第一次真正"联网";
2. **read_file / write_file**:读写文件,但**限制在 docs 目录内**(防路径穿越);
3. **run_command**:执行系统命令,但**白名单 + 超时**(防命令注入);
4. **kv_get / kv_set**:一个持久化的键值存储,Agent 可以"记账";
5. 安全三原则:**最小权限、白名单、超时**。

## 目录结构

```text
ch08/
├── main.go        # 入口(自动准备 docs/notes.txt 示例文件)
├── llm.go         # DeepSeek 客户端
├── tools.go       # Registry + 6 个真实工具 + 安全校验
├── agent.go       # ReAct 循环(课时3 的内容)
├── tools_test.go  # 离线测试:http_get/路径穿越/白名单/KV
└── README.md
```

## 运行

```powershell
cd E:\codex\project001\goagent\ch08
go run . "访问 https://httpbin.org/get 看看返回什么"
go run . "把 hello agent 写到 docs/note.txt"
go run . "执行 go version 看看环境"
go run . "读取 docs/notes.txt"
```

注意观察 Agent 的每一轮:它会自己决定调用哪个工具,拿到 Observation 后继续。

## 安全设计(本课重点)

```go
// 1. 路径穿越防护:文件名最终必须落在 docs 目录内
func safePath(root, name string) (string, error)

// 2. 命令白名单:只允许明确的命令前缀
func allowedCommand(cmd string) bool // 允许 go version / echo / dir

// 3. 所有工具都有超时(5~10 秒),防止 Agent 把进程卡死
```

为什么必须有这些?因为**模型输出不可信**——用户可以让模型说
"执行 `rm -rf /`",如果工具不做校验,你的电脑就遭殃了。这就是
提示注入的威力,课时12 会从 Agent 层面继续防御。

## 练习

1. 往白名单加一条 `python --version`,试运行;
2. 故意让 Agent "读取 ../README.md",看它被拒绝后怎么跟你解释;
3. 用 `kv_set` 让 Agent 记住"项目代号",再开一个新会话用 `kv_get` 问它;
4. 思考:为什么 `http_get` 也要限制(比如只允许 https)?

## 小结

真实工具让 Agent 从"聊天机器人"变成"能干活的小助手",但能力越大、
风险越大。这一课的安全三原则(最小权限、白名单、超时)是后续所有
真实工具(数据库、支付、邮件)必须遵守的底线。
