# ch01：LLM API

目标：把模型当成一个不稳定、有限额、会返回非确定结果的远程依赖，而不是魔法函数。

## 学习内容

- system/user/assistant/tool 消息；
- 上下文窗口、输入/输出 token 和截断；
- temperature 等生成参数不等于正确性保证；
- HTTP 4xx、429、5xx、网络超时和取消的区别；
- 从响应中记录 request ID、finish reason 和 usage；
- 模型名、价格和 base URL 必须配置化。

实现位于 [`pkg/llm`](../pkg/llm/)。联网运行前设置 `LLM_API_KEY`、`LLM_BASE_URL` 和 `LLM_MODEL`。

## 验收

客户端必须传播调用方 context、限制响应体、在非 2xx 时返回带状态码的错误，并保留 usage。

## 练习

用 `httptest.Server` 分别模拟 200、429、500、慢响应和损坏 JSON，验证客户端错误分类。

