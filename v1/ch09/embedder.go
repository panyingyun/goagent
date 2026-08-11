package main

import (
	"context"
	"math"
)

// Embedder 文本向量化抽象(与 ch19 相同)。
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
}

// HashEmbedder 本地哈希向量化:把每个字符哈希到固定维度的桶里。
// 不需要任何外部 API,离线可跑,适合学习;生产可换成 OpenAI/本地模型。
const hashDim = 256

type HashEmbedder struct{}

func (HashEmbedder) Embed(_ context.Context, texts []string) ([][]float64, error) {
	out := make([][]float64, len(texts))
	for i, t := range texts {
		out[i] = hashEmbed(t)
	}
	return out, nil
}

func hashEmbed(text string) []float64 {
	v := make([]float64, hashDim)
	for _, r := range text {
		if r == ' ' || r == '\t' || r == '\n' {
			continue
		}
		idx := (int(r)*31 + int(r)>>3) % hashDim
		v[idx]++
	}
	var norm float64
	for _, x := range v {
		norm += x * x
	}
	if norm > 0 {
		s := math.Sqrt(norm)
		for i := range v {
			v[i] /= s
		}
	}
	return v
}
