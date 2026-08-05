package main

import (
	"testing"

	"goagent-new/pkg/retrieval"
)

func TestSearchDocumentsReturnsSourceAndRecall(t *testing.T) {
	hits := searchDocuments("住宿报销上限", 2)
	if len(hits) == 0 || hits[0].Document.ID != "travel-policy" {
		t.Fatalf("unexpected hits: %+v", hits)
	}
	if recall := retrieval.RecallAtK(hits, map[string]bool{"travel-policy": true}); recall != 1 {
		t.Fatalf("recall = %v", recall)
	}
}

