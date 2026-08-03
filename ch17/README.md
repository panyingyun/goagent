# 课时17:Agent Web 界面

> 目标:给 Agent 一个浏览器界面,支持普通问答和 SSE 流式打字机效果。

## 本课时学什么

1. `net/http` 静态文件 + JSON API;
2. SSE 服务端推送:回答逐字从浏览器流出来(复用课时7 的流式能力);
3. 前端最小实现:一个 HTML 文件,不引任何框架。

## 运行

```powershell
cd E:\codex\project001\goagent\ch17
go run .
```

浏览器打开 http://localhost:8080,输入问题,对比"发送"和"流式"按钮的体验。

离线测试:

```powershell
go test ./... -v
```

## 练习

1. 给 `/chat` 加历史:前端把之前的对话一起 POST 过来;
2. 把 `/chat/stream` 改成 POST(EventSource 只支持 GET,想想怎么绕,
   提示:fetch + ReadableStream);
3. 给页面加一个"停止生成"按钮。
