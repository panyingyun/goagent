# 课时1:Go 调用 LLM(DeepSeek)

> 目标:亲手发一个请求让 DeepSeek"说话",看懂 Chat Completions 的请求/响应结构。

## 本课时学什么

- LLM(大语言模型)是什么:一个只负责"续写文本"的模型,本身不会动手;
- Chat Completions API:把一段**消息列表**发给模型,它返回回复;
- 三个角色:system(系统设定)、user(用户)、assistant(模型);
- 模型名:截至 2026-08 用 `deepseek-v4-flash`(默认,便宜快)或 `deepseek-v4-pro`。

## 目录结构

```text
ch01/
├── main.go   # 整个课时只有这一个文件:读问题 -> 调 API -> 打印回复
└── README.md
```

## 运行

```powershell
cd E:\codex\project001\ch01
go run . "用一句话介绍你自己"
go run . "12 * 8 + 4 等于多少?"
```

也可以从管道读问题:

```powershell
"上海现在几点?" | go run .
```

可选环境变量:

- `DEEPSEEK_MODEL`:模型名,默认 `deepseek-v4-flash`;
- `DEEPSEEK_BASE_URL`:API 地址,默认 `https://api.deepseek.com`。

## 请求到底长什么样

```json
{
  "model": "deepseek-v4-flash",
  "messages": [
    {"role": "system", "content": "你是一个乐于助人的中文助手,回答简洁清晰。"},
    {"role": "user", "content": "12 * 8 + 4 等于多少?"}
  ]
}
```

响应里 `choices[0].message.content` 就是模型说的话。

## 三个关键认知

1. **模型是无状态的**:它不记得上次对话,所以每次都要把完整历史发过去;
2. **system 消息很重要**:角色的行为规范都写在 system 里(课时 3 的 SKILL 就是往这里塞内容);
3. **模型只会"说",不会"做"**:它算不准大数、查不了时间——这就是为什么后面需要
   ReAct(课时 4)和 MCP(课时 2、5)给它"手"。

## 练习

1. 改成**多轮对话**:把上一轮模型回复(role=assistant)追加进 messages,再问一次;
2. 打印出完整的原始响应 JSON,看看除了 content 还有什么字段;
3. 把 `stream` 改成 `true`,观察响应格式的变化(提示:会变成多行数据块)。

## 小结

这一课打通了"Go 和 LLM 之间的通道"。之后所有课时都建立在这条通道上:
SKILL 改的是 system 内容,ReAct 改的是对话循环,MCP 给模型提供工具。
