package retrieval

import "testing"

func TestSearchAndRecall(t *testing.T) {
	idx := NewKeywordIndex([]Document{
		{ID: "leave", Content: "员工每年享有十天年假"},
		{ID: "travel", Content: "差旅住宿报销上限为六百元"},
	})
	hits := idx.Search("年假有多少天", 1)
	if len(hits) != 1 || hits[0].Document.ID != "leave" {
		t.Fatalf("unexpected hits: %+v", hits)
	}
	if RecallAtK(hits, map[string]bool{"leave": true}) != 1 {
		t.Fatal("recall should be 1")
	}
}

