package main

import (
	"sort"
	"strings"
)

// Hit 是一次检索命中的片段及其得分。
type Hit struct {
	Text  string
	Score float64
}

// HybridRetriever 做"关键词 + 向量"混合检索:
// 关键词擅长精确匹配,向量擅长语义相近;两者融合更稳。
type HybridRetriever struct {
	Docs    []string
	Vecs    [][]float64
	KWRatio float64 // 关键词得分权重(0~1)
}

func NewHybridRetriever(docs []string, vecs [][]float64) *HybridRetriever {
	return &HybridRetriever{Docs: docs, Vecs: vecs, KWRatio: 0.4}
}

func (h *HybridRetriever) Search(query string, k int) []Hit {
	qv := hashEmbed(query)
	queryChars := map[rune]bool{}
	for _, r := range query {
		queryChars[r] = true
	}
	hits := make([]Hit, 0, len(h.Docs))
	for i, doc := range h.Docs {
		// 关键词得分:查询字符在文档中出现的比例
		kw := 0.0
		for r := range queryChars {
			if strings.ContainsRune(doc, r) {
				kw++
			}
		}
		if len(queryChars) > 0 {
			kw /= float64(len(queryChars))
		}
		// 向量得分
		vec := cosineSim(h.Vecs[i], qv)
		score := h.KWRatio*kw + (1-h.KWRatio)*vec
		hits = append(hits, Hit{Text: doc, Score: score})
	}
	sort.SliceStable(hits, func(a, b int) bool { return hits[a].Score > hits[b].Score })
	if k > len(hits) {
		k = len(hits)
	}
	return hits[:k]
}

func cosineSim(a, b []float64) float64 {
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (sqrt(na) * sqrt(nb))
}

func sqrt(x float64) float64 {
	// 避免引入 math 以外的复杂度
	return mathSqrt(x)
}

// mathSqrt 委托 math.Sqrt,保持 hybrid.go 自包含可读。
func mathSqrt(x float64) float64 {
	return float64Sqrt(x)
}

var float64Sqrt = func(x float64) float64 { return sqrtImpl(x) }

func sqrtImpl(x float64) float64 {
	if x <= 0 {
		return 0
	}
	// 牛顿迭代
	z := x
	for i := 0; i < 20; i++ {
		z = (z + x/z) / 2
	}
	return z
}
