// 课时8:RAG——让 Agent 拥有私有知识(路线B:OpenAI embedding)
//
// RAG = Retrieval-Augmented Generation(检索增强生成)。
// 流程:文档切块 -> 向量化 -> 建索引 -> 提问时检索 top-k -> 注入提示词 -> DeepSeek 回答。
//
// 运行(需要 DEEPSEEK_API_KEY 和 OPENAI_API_KEY):
//
//	go run . "员工生日当天可以休假吗?"
//	go run .                      # 交互模式
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

const indexFile = "index.json"

func main() {
	ctx := context.Background()
	emb := NewOpenAIEmbedder()
	llm := NewLLMClient()
	if llm.APIKey == "" {
		fmt.Fprintln(os.Stderr, "缺少 DEEPSEEK_API_KEY 环境变量")
		os.Exit(1)
	}

	// 1. 建索引:有缓存直接加载,没有就用 OPENAI_API_KEY 向量化 docs/
	idx, err := LoadIndex(indexFile)
	if err != nil {
		texts, err := loadDocs("docs")
		if err != nil || len(texts) == 0 {
			fmt.Fprintln(os.Stderr, "docs 目录下没有 .txt 文档:", err)
			os.Exit(1)
		}
		var chunks []string
		for _, t := range texts {
			chunks = append(chunks, chunkText(t, 300)...)
		}
		fmt.Fprintf(os.Stderr, "切块完成,共 %d 个片段,开始向量化...\n", len(chunks))
		idx, err = BuildIndex(ctx, emb, chunks)
		if err != nil {
			fmt.Fprintln(os.Stderr, "向量化失败:", err)
			os.Exit(1)
		}
		if err := idx.Save(indexFile); err != nil {
			fmt.Fprintln(os.Stderr, "保存索引失败:", err)
		} else {
			fmt.Fprintf(os.Stderr, "索引已保存到 %s\n", indexFile)
		}
	} else {
		fmt.Fprintf(os.Stderr, "已加载索引(%d 个片段)\n", len(idx.Chunks))
	}

	// 2. 问答
	question := strings.Join(os.Args[1:], " ")
	if question != "" {
		answer, err := ask(ctx, llm, emb, idx, question)
		if err != nil {
			fmt.Fprintln(os.Stderr, "失败:", err)
			os.Exit(1)
		}
		fmt.Println(answer)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("RAG 问答开始,输入 /exit 退出")
	for {
		fmt.Print("问> ")
		if !scanner.Scan() {
			break
		}
		q := strings.TrimSpace(scanner.Text())
		if q == "" || q == "/exit" {
			break
		}
		answer, err := ask(ctx, llm, emb, idx, q)
		if err != nil {
			fmt.Fprintln(os.Stderr, "失败:", err)
			continue
		}
		fmt.Println("\n答>", answer)
	}
}

func ask(ctx context.Context, llm *LLMClient, emb Embedder, idx *Index, q string) (string, error) {
	hits, err := idx.Search(ctx, emb, q, 3)
	if err != nil {
		return "", err
	}
	fmt.Fprintf(os.Stderr, "检索到 %d 个片段:\n", len(hits))
	for _, h := range hits {
		fmt.Fprintf(os.Stderr, "  - %s\n", truncateRune(h, 60))
	}

	system := "你是公司内部知识助手。只能根据下面资料回答;资料里没有的信息,直接回答\"资料中没有相关信息\"。\n\n=== 资料 ===\n" +
		strings.Join(hits, "\n---\n")
	reply, err := llm.Chat(ctx, []Message{
		{Role: RoleSystem, Content: system},
		{Role: RoleUser, Content: q},
	})
	if err != nil {
		return "", err
	}
	return reply.Content, nil
}

func truncateRune(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}
