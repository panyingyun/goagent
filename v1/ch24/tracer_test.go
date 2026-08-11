package main

import (
	"testing"
	"time"
)

func TestRunTree(t *testing.T) {
	tr := NewTracer()
	root := tr.StartRun("agent")
	child := tr.StartRun("llm")
	child.Attrs["tokens"] = 10
	tr.EndRun(child)
	tr.EndRun(root)

	tree := tr.Export()
	if tree["name"] != "agent" {
		t.Fatalf("根 run 应为 agent: %v", tree)
	}
	kids := tree["children"].([]map[string]any)
	if len(kids) != 1 || kids[0]["name"] != "llm" {
		t.Fatalf("子 run 缺失: %v", kids)
	}
}

func TestFailRun(t *testing.T) {
	tr := NewTracer()
	r := tr.StartRun("llm")
	tr.FailRun(r, errTest{})
	tree := tr.Export()
	if tree["error"] == "" {
		t.Fatalf("失败 run 应带 error: %v", tree)
	}
}

type errTest struct{}

func (errTest) Error() string { return "模拟错误" }

func TestStartEndTime(t *testing.T) {
	tr := NewTracer()
	r := tr.StartRun("x")
	time.Sleep(2 * time.Millisecond)
	tr.EndRun(r)
	if !r.End.After(r.Start) {
		t.Fatal("start/end 时间应不同")
	}
}
