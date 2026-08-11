# 课时14:可观测性与日志

> 目标:给 Agent 装上"黑匣子"——每次 LLM 调用、工具调用都有记录和指标。

## 本课时学什么

1. **Span 树**:Start/End 自动形成父子关系,看到"这次任务里 LLM 花了多久、
   工具花了多久";
2. **指标**:计数器(任务次数) + 直方图(平均耗时);
3. 和 LangSmith 的关系:本课的 `Tracer` 就是 LangSmith run 的最小实现,
   课时24 会把它升级成可导出的完整追踪。

## 运行

```powershell
cd E:\codex\project001\goagent\ch14
go run .
```

输出两段:span 树 JSON + 指标快照。

## 练习

1. 给 span 加 `token_used` 属性,统计每次 LLM 调用的成本;
2. 把 `Export` 改成写文件,模拟日志落盘;
3. 想一想:为什么生产日志要"结构化 JSON"而不是一行行文本?
