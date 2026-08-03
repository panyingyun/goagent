// 课时12:向量数据库与高级 RAG
//
// 在 ch08 的基础上做三件事:
// 1. 本地 HashEmbedder(离线可跑,不依赖 OpenAI Key)
// 2. 混合检索:关键词 + 向量融合
// 3. 重排:去重,保证结果多样性
//
// 运行:
//
//	go run . "员工生日当天可以休假吗?"
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

func main() {
	emb := HashEmbedder{}
	ctx := context.Background()

	texts, err := loadDocs("docs")
	if err != nil || len(texts) == 0 {
		fmt.Fprintln(os.Stderr, "docs 目录下没有 .txt 文档")
		os.Exit(1)
	}
	var chunks []string
	for _, t := range texts {
		chunks = append(chunks, chunkText(t, 300)...)
	}
	vecs, err := emb.Embed(ctx, chunks)
	if err != nil {
		fmt.Fprintln(os.Stderr, "向量化失败:", err)
		os.Exit(1)
	}
	hr := NewHybridRetriever(chunks, vecs)
	fmt.Fprintf(os.Stderr, "索引完成: %d 个片段\n", len(chunks))

	question := strings.Join(os.Args[1:], " ")
	if question == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("问> ")
		if scanner.Scan() {
			question = strings.TrimSpace(scanner.Text())
		}
	}
	if question == "" {
		fmt.Fprintln(os.Stderr, "用法: go run . \"问题\"")
		os.Exit(1)
	}

	hits := Rerank(hr.Search(question, 5), 3)
	fmt.Fprintf(os.Stderr, "检索到 %d 个片段(混合检索+重排):\n", len(hits))
	for _, h := range hits {
		fmt.Fprintf(os.Stderr, "  [%.2f] %s\n", h.Score, truncateRune(h.Text, 50))
	}

	system := "你是公司知识助手,只能根据下面资料回答。\n\n=== 资料 ===\n" + joinHits(hits)
	reply, err := chatLLM(ctx, system, question)
	if err != nil {
		fmt.Fprintln(os.Stderr, "失败:", err)
		os.Exit(1)
	}
	fmt.Println(reply)
}

func joinHits(hits []Hit) string {
	parts := make([]string, len(hits))
	for i, h := range hits {
		parts[i] = h.Text
	}
	return strings.Join(parts, "\n---\n")
}

func truncateRune(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}
