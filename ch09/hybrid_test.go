package main

import (
	"context"
	"strings"
	"testing"
)

func TestHashEmbed(t *testing.T) {
	v1 := hashEmbed("我喜欢喝咖啡")
	v2 := hashEmbed("我喜欢喝咖啡")
	v3 := hashEmbed("今天天气不错")
	if cosineSim(v1, v2) < 0.99 {
		t.Error("相同文本相似度应接近 1")
	}
	if cosineSim(v1, v3) >= 0.99 {
		t.Error("不同文本相似度不应接近 1")
	}
}

// 混合检索:语义相似(共享关键词)应排在前面。
func TestHybridSearch(t *testing.T) {
	docs := []string{"员工生日放假一天", "年假上限 25 天", "总部在上海张江"}
	vecs, _ := (HashEmbedder{}).Embed(context.Background(), docs)
	hr := NewHybridRetriever(docs, vecs)
	hits := hr.Search("生日休假", 1)
	if len(hits) == 0 || !strings.Contains(hits[0].Text, "生日") {
		t.Fatalf("第一个命中应为生日相关,实际: %v", hits)
	}
}

func TestRerankDedup(t *testing.T) {
	hits := []Hit{
		{Text: "生日放假一天", Score: 0.9},
		{Text: "生日放假一天,可带家属", Score: 0.88},
		{Text: "年假规则", Score: 0.8},
	}
	out := Rerank(hits, 2)
	if len(out) != 2 || !strings.Contains(out[1].Text, "年假") {
		t.Fatalf("重排应去重并保留多样性: %v", out)
	}
}
