# A1：文本 ReAct

文本 ReAct 把模型输出约定为 `Action`、`Action Input` 和 `Observation`。它适合展示 Agent 循环的历史来源，但不作为本课程生产主线，原因包括：

- 格式依赖提示词，容易漂移；
- 自然语言和 JSON 混排，解析脆弱；
- 容易诱导应用展示模型的内部推理文本；
- 工具参数缺少协议级 Schema。

练习：用一个离线 scripted model 对比文本解析和原生 `tool_calls`，记录两者的失败方式。不要把完整内部 Thought 写进生产日志。

