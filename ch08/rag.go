package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// chunkText 按段落切块,并把小块合并到接近 maxLen 字符,超长段硬切。
func chunkText(text string, maxLen int) []string {
	paras := strings.Split(text, "\n")
	var chunks []string
	var cur strings.Builder
	for _, p := range paras {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		for len(p) > maxLen {
			if cur.Len() > 0 {
				chunks = append(chunks, cur.String())
				cur.Reset()
			}
			chunks = append(chunks, p[:maxLen])
			p = p[maxLen:]
		}
		if cur.Len() > 0 && cur.Len()+len(p) > maxLen {
			chunks = append(chunks, cur.String())
			cur.Reset()
		}
		if cur.Len() > 0 {
			cur.WriteString("\n")
		}
		cur.WriteString(p)
	}
	if cur.Len() > 0 {
		chunks = append(chunks, cur.String())
	}
	return chunks
}

// Index 是"文档片段 -> 向量"的索引,可保存/加载到 JSON。
type Index struct {
	Chunks  []string    `json:"chunks"`
	Vectors [][]float64 `json:"vectors"`
}

func BuildIndex(ctx context.Context, emb Embedder, chunks []string) (*Index, error) {
	vecs, err := emb.Embed(ctx, chunks)
	if err != nil {
		return nil, err
	}
	return &Index{Chunks: chunks, Vectors: vecs}, nil
}

// Search 把问题向量化,按余弦相似度返回最相关的 k 个片段。
func (idx *Index) Search(ctx context.Context, emb Embedder, query string, k int) ([]string, error) {
	qvecs, err := emb.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	type scored struct {
		text  string
		score float64
	}
	items := make([]scored, 0, len(idx.Chunks))
	for i, c := range idx.Chunks {
		items = append(items, scored{c, cosineSimilarity(idx.Vectors[i], qvecs[0])})
	}
	sort.SliceStable(items, func(a, b int) bool { return items[a].score > items[b].score })
	top := make([]string, 0, k)
	for i := 0; i < k && i < len(items); i++ {
		top = append(top, items[i].text)
	}
	return top, nil
}

func (idx *Index) Save(path string) error {
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadIndex(path string) (*Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

// loadDocs 读取目录下所有 .txt 文件。
func loadDocs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var texts []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".txt" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		texts = append(texts, string(data))
	}
	return texts, nil
}
