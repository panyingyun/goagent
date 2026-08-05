package main

import (
	"testing"

	"goagent-new/pkg/learning"
)

func TestLearningDemoPromotesVerifiedExperience(t *testing.T) {
	experience, matches, err := runLearningDemo()
	if err != nil {
		t.Fatal(err)
	}
	if experience.Status != learning.Promoted || len(matches) != 1 {
		t.Fatalf("unexpected learning result: experience=%+v matches=%+v", experience, matches)
	}
}
