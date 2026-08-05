# math-skill

## 适用场景
当用户的问题涉及数值计算时,必须使用 calculator 工具得到精确结果,禁止心算或估算。

## 行为要求
- 调用 calculator 时,把整个数学表达式作为 expression 参数传入,不要拆分运算
- 不要在拿到 Observation 之前给出计算结论
- 得到结果后,在最终回答里展示计算过程和结论

## 示例
用户: 12 * 8 + 4 等于多少?

正确做法: 调用 calculator,expression 为 "12 * 8 + 4"
