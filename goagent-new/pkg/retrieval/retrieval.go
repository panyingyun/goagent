package retrieval

import (
	"sort"
	"strings"
	"unicode"
)

type Document struct {
	ID      string
	Content string
}

type Hit struct {
	Document Document
	Score    float64
}

type KeywordIndex struct {
	docs   []Document
	terms  []map[string]struct{}
}

func NewKeywordIndex(docs []Document) *KeywordIndex {
	idx := &KeywordIndex{docs: append([]Document(nil), docs...), terms: make([]map[string]struct{}, len(docs))}
	for i, doc := range docs {
		idx.terms[i] = tokenize(doc.Content)
	}
	return idx
}

func (i *KeywordIndex) Search(query string, k int) []Hit {
	q := tokenize(query)
	hits := make([]Hit, 0, len(i.docs))
	for n, docTerms := range i.terms {
		intersection := 0
		for term := range q {
			if _, ok := docTerms[term]; ok {
				intersection++
			}
		}
		if intersection > 0 {
			hits = append(hits, Hit{Document: i.docs[n], Score: float64(intersection) / float64(len(q))})
		}
	}
	sort.SliceStable(hits, func(a, b int) bool { return hits[a].Score > hits[b].Score })
	if k < 0 {
		k = 0
	}
	if len(hits) > k {
		hits = hits[:k]
	}
	return hits
}

func RecallAtK(hits []Hit, relevantIDs map[string]bool) float64 {
	if len(relevantIDs) == 0 {
		return 0
	}
	found := 0
	seen := make(map[string]bool)
	for _, hit := range hits {
		if relevantIDs[hit.Document.ID] && !seen[hit.Document.ID] {
			found++
			seen[hit.Document.ID] = true
		}
	}
	return float64(found) / float64(len(relevantIDs))
}

func tokenize(text string) map[string]struct{} {
	text = strings.ToLower(text)
	terms := make(map[string]struct{})
	words := strings.FieldsFunc(text, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) })
	for _, word := range words {
		terms[word] = struct{}{}
		runes := []rune(word)
		for i := 0; i+1 < len(runes); i++ {
			terms[string(runes[i:i+2])] = struct{}{}
		}
	}
	return terms
}

