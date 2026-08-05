# 课时9:向量数据库与高级 RAG

> 目标:用不依赖外部 embedding Key 的方式做出"好用"的 RAG——混合检索 + 重排。
> (需要真实 OpenAI embedding 时再看选修 ch19。)

## 本课时学什么

1. **本地 HashEmbedder**:把文本哈希成 256 维向量,离线可跑,先讲透原理;
2. **混合检索**:关键词(精确) + 向量(语义)按权重融合;
3. **重排(Rerank)**:去重 + 多样性,避免 top-k 全是同一段话的变体;
4. 生产中的向量数据库(Milvus / Qdrant / Chroma)本质就是
   "能存向量 + 能按相似度检索的数据库",本课的 `HybridRetriever`
   就是它的最小等价物。

## 目录结构

```text
ch09/
├── embedder.go   # HashEmbedder(本地,离线)
├── hybrid.go     # 混合检索 + 余弦相似度
├── rerank.go     # 去重重排
├── rag.go        # 切块 / 文档加载 / DeepSeek 调用
├── main.go       # 索引 + 问答入口
├── docs/         # 示例文档
└── README.md
```

## 运行

```powershell
cd E:\codex\project001\goagent\ch09
go run . "员工生日当天可以休假吗?"
go run . "住宿报销上限是多少?"
```

离线测试:

```powershell
go test ./... -v
```

## 和 ch19 的对比

| | ch19 | ch09 |
|---|---|---|
| 向量化 | OpenAI(需 Key) | 本地 Hash(离线) |
| 检索 | 纯向量 | 关键词 + 向量混合 |
| 重排 | 无 | 去重 + 多样性 |
| 目的 | 跑通链路 | 提高检索质量 |

## 练习

1. 把 `KWRatio` 从 0.4 改成 0 或 1,对比检索结果;
2. 在 `docs/` 加两个内容高度相似的文件,观察重排去重;
3. 想一想:真实向量数据库还有哪些能力本课没做?
   (提示:持久化、分区、过滤、增量更新。)
