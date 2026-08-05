package main

import "testing"

func TestBuildDemoContextSelectsGoGuide(t *testing.T) {
	bundle, err := buildDemoContext("修复 Go HTTP handler")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range bundle.Items {
		if item.Source.ID == "go-guide" {
			found = true
		}
	}
	if !found {
		t.Fatalf("relevant source was not selected: %+v", bundle.Items)
	}
}
