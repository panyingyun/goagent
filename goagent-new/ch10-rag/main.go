package main

import (
	"flag"
	"fmt"
	"os"

	"goagent-new/pkg/retrieval"
)

func searchDocuments(query string, k int) []retrieval.Hit {
	documents := []retrieval.Document{
		{ID: "leave-policy", Content: "员工每年享有十天年假，入职不满一年按比例计算。"},
		{ID: "travel-policy", Content: "一线城市住宿报销上限为每晚六百元，报销必须附发票。"},
		{ID: "security-guide", Content: "所有写操作必须审批，提示词不能替代权限边界。"},
	}
	return retrieval.NewKeywordIndex(documents).Search(query, k)
}

func main() {
	query := flag.String("q", "住宿报销上限", "检索问题")
	k := flag.Int("k", 2, "返回结果数")
	flag.Parse()
	hits := searchDocuments(*query, *k)
	if len(hits) == 0 {
		fmt.Fprintln(os.Stderr, "no relevant documents")
		os.Exit(1)
	}
	for _, hit := range hits {
		fmt.Printf("source=%s score=%.3f\n%s\n\n", hit.Document.ID, hit.Score, hit.Document.Content)
	}
	relevant := map[string]bool{"travel-policy": true}
	fmt.Printf("示例标注下 Recall@K=%.2f\n", retrieval.RecallAtK(hits, relevant))
}

