# ch10：可选 RAG

RAG 不是所有 Agent 的必选组件。本章从可测量的检索问题开始。

## 路径

1. 关键词/BM25 类基线；
2. 定义 `Embedder` 接口并接入真实或本地 embedding；
3. 混合检索、元数据过滤和去重；
4. 文档内容始终标记为不可信引用；
5. 使用 Recall@K、MRR 等指标评估检索，再评估生成答案；
6. 回答提供来源 ID，不让模型伪造引用。

[`pkg/retrieval`](../pkg/retrieval/) 实现一个可离线测试的关键词基线和 Recall@K。它明确不是“语义 embedding 的等价物”。

